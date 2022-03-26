package dataflow

type State map[string]ValueMap

type ValueStorageInstance struct {
	// this represents the state of the storage instance and must only be access by the main go routine

	// state: 1. dimension: eevice.Name, 2. dimension: register.Name
	state         State
	subscriptions map[*Subscription]struct{}

	// communication channels to/from the main go routine
	inputChannel            chan Value
	subscriptionChannel     chan *Subscription
	readStateRequestChannel chan *readStateRequest
}

type Filter struct {
	Devices       map[string]bool
	RegisterNames map[string]bool
}

type readStateRequest struct {
	filter   Filter
	response chan State
}

func (instance *ValueStorageInstance) mainStorageRoutine() {
	for {
		select {
		case newValue := <-instance.inputChannel:
			instance.handleNewValue(newValue)
		case newSubscription := <-instance.subscriptionChannel:
			instance.subscriptions[newSubscription] = struct{}{}
		case newReadStateRequest := <-instance.readStateRequestChannel:
			instance.handleNewReadStateRequest(newReadStateRequest)
		}
	}
}

func (instance *ValueStorageInstance) handleNewValue(newValue Value) {
	// check if the newValue is not present or has been changed
	if _, ok := instance.state[newValue.DeviceName()]; !ok {
		instance.state[newValue.DeviceName()] = make(ValueMap)
	}
	if currentValue, ok := instance.state[newValue.DeviceName()][newValue.Register().Name()]; !ok || !currentValue.Equals(newValue) {
		// copy the input value to all subscribed output channels
		for subscription, _ := range instance.subscriptions {
			// check if Subscription was shut down
			select {
			case <-subscription.shutdownChannel:
				delete(instance.subscriptions, subscription)
			default:
				// Subscription was not shut down -> forward new value
				subscription.forward(newValue)
			}
		}

		// and save the new state
		instance.state[newValue.DeviceName()][newValue.Register().Name()] = newValue
	}
}

func (instance *ValueStorageInstance) handleNewReadStateRequest(newReadStateRequest *readStateRequest) {
	filter := &newReadStateRequest.filter

	response := make(State)

	for deviceName, deviceState := range instance.state {
		if !filterByDevice(filter, deviceName) {
			continue
		}

		response[deviceName] = make(ValueMap)

		for registerName, value := range deviceState {
			if !filterByRegisterName(filter, registerName) {
				continue
			}

			response[deviceName][registerName] = value
		}
	}

	newReadStateRequest.response <- response
}

func ValueStorageCreate() (valueStorageInstance *ValueStorageInstance) {
	valueStorageInstance = &ValueStorageInstance{
		state:                   make(State),
		subscriptions:           make(map[*Subscription]struct{}),
		inputChannel:            make(chan Value, 32), // input channel is buffered
		subscriptionChannel:     make(chan *Subscription),
		readStateRequestChannel: make(chan *readStateRequest),
	}

	// start main go routine
	go valueStorageInstance.mainStorageRoutine()

	return
}

func (instance *ValueStorageInstance) Shutdown() {
	for subscription, _ := range instance.subscriptions {
		subscription.Shutdown()
	}
}

func (instance *ValueStorageInstance) GetState(filter Filter) State {
	response := make(chan State)

	request := readStateRequest{
		filter:   filter,
		response: response,
	}

	instance.readStateRequestChannel <- &request

	return <-request.response
}

func (instance *ValueStorageInstance) GetMap(filter Filter) (result ValueMap) {
	result = make(ValueMap)

	state := instance.GetState(filter)

	for _, deviceState := range state {
		for registerName, value := range deviceState {
			result[registerName] = value
		}
	}

	return
}

// this is a simple fan-in routine which copies all inputs to the same NewValue channel
func (instance *ValueStorageInstance) Fill(input <-chan Value) {
	go func() {
		for value := range input {
			instance.inputChannel <- value
		}
	}()
}

func (instance *ValueStorageInstance) Drain() Subscription {
	return instance.Subscribe(Filter{})
}

func (instance *ValueStorageInstance) Subscribe(filter Filter) Subscription {
	s := Subscription{
		shutdownChannel: make(chan struct{}),
		outputChannel:   make(chan Value, 32),
		filter:          filter,
	}

	instance.subscriptionChannel <- &s

	return s
}
