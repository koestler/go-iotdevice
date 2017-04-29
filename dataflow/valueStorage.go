package dataflow

type StorageKey struct {
	Name   string
	Device *Device
}

type subscription struct {
	outputChannel chan Value
	filters       SubscriptionFilter
}

type SubscriptionFilter struct {
	Devices    map[*Device]bool
	ValueNames map[string]bool
}

type ValueStorageInstance struct {
	// this represents the state of the storage instance and must only be access by the main go routine
	state         map[StorageKey]Value
	subscriptions []subscription

	// communication channels to/from the main go routine
	inputChannel        chan Value
	subscriptionChannel chan subscription
}

func mainStorageRoutine(valueStorageInstance *ValueStorageInstance) {
	for {
		select {
		case newValue := <-valueStorageInstance.inputChannel:
			// compute key
			key := StorageKey{
				Name:   newValue.Name,
				Device: newValue.Device,
			}

			// check if the newValue is not present or has been changed
			if currentValue, ok := valueStorageInstance.state[key]; !ok || currentValue != newValue {
				// copy the input value to all subscribed output channels
				for _, subscription := range valueStorageInstance.subscriptions {
					subscription.forward(newValue)
				}

				// and save the new state
				valueStorageInstance.state[key] = newValue
			}
		case newSubscription := <-valueStorageInstance.subscriptionChannel:
			valueStorageInstance.subscriptions = append(valueStorageInstance.subscriptions, newSubscription)
		}
	}
}

func ValueStorageCreate() (valueStorageInstance *ValueStorageInstance) {
	valueStorageInstance = &ValueStorageInstance{
		state:               make(map[StorageKey]Value),
		inputChannel:        make(chan Value, 4),
		subscriptionChannel: make(chan subscription),
	}

	// main go routine
	go mainStorageRoutine(valueStorageInstance)

	return
}

// this is a simple fan-in routine which copies all inputs to the same NewValue channel
func (valueStorageInstance *ValueStorageInstance) Fill(input <-chan Value) {
	go func() {
		for value := range input {
			valueStorageInstance.inputChannel <- value
		}
	}()
}

func (valueStorageInstance *ValueStorageInstance) Drain() <-chan Value {
	return valueStorageInstance.Subscribe(SubscriptionFilter{});
}

func (valueStorageInstance *ValueStorageInstance) Subscribe(filter SubscriptionFilter) <-chan Value {
	output := make(chan Value)

	valueStorageInstance.subscriptionChannel <- subscription{
		outputChannel: output,
		filters:       filter,
	}

	return output
}

func (valueStorageInstance *ValueStorageInstance) Append(fillable Fillable) Fillable {
	fillable.Fill(valueStorageInstance.Drain())
	return fillable
}

func (subscription subscription) forward(newValue Value) {
	filters := subscription.filters

	// implement filters
	if _, ok := filters.Devices[newValue.Device]; len(filters.Devices) > 0 && !ok {
		// device list is not empty and the device is not on the list -> do not forward
		return;
	}

	if _, ok := filters.ValueNames[newValue.Name]; len(filters.ValueNames) > 0 && !ok {
		// value names list is not empty and the value name is not on the list -> do not forward
		return;
	}

	// forward value
	subscription.outputChannel <- newValue
}
