package dataflow

type StorageKey struct {
	Name   string
	Device *Device
}

type subscription struct {
	outputChannel chan Value
}

type StorageInstance struct {
	// this represents the state of the storage instance and must only be access by the main go routine
	State         map[StorageKey]Value
	subscriptions []subscription

	// communication channels to/from the main go routine
	inputChannel        chan Value
	subscriptionChannel chan subscription
}

func mainStorageRoutine(storageInstance *StorageInstance) {
	for {
		select {
		case newValue := <-storageInstance.inputChannel:
			// compute key
			key := StorageKey{
				Name:   newValue.Name,
				Device: newValue.Device,
			}

			// check if the newValue is not present or has been changed
			if currentValue, ok := storageInstance.State[key]; !ok || currentValue != newValue {
				// copy the input value to all subscribed output channels
				for _, subscription := range storageInstance.subscriptions {
					subscription.outputChannel <- newValue
				}

				// and save the new state
				storageInstance.State[key] = newValue
			}
		case newSubscription := <-storageInstance.subscriptionChannel:
			storageInstance.subscriptions = append(storageInstance.subscriptions, newSubscription)
		}
	}
}

func StorageCreate() (storageInstance *StorageInstance) {
	storageInstance = &StorageInstance{
		State:               make(map[StorageKey]Value),
		inputChannel:        make(chan Value, 4),
		subscriptionChannel: make(chan subscription),
	}

	// main go routine
	go mainStorageRoutine(storageInstance)

	return
}

// this is a simple fan-in routine which copies all inputs to the same NewValue channel
func (storageInstance *StorageInstance) Receive(input <-chan Value) {
	go func() {
		for value := range input {
			storageInstance.inputChannel <- value

		}
	}()
}

func (storageInstance *StorageInstance) Subscribe() <-chan Value {
	output := make(chan Value)

	storageInstance.subscriptionChannel <- subscription{
		outputChannel: output,
	}

	return output
}
