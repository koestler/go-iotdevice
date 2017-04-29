package dataflow

type State map[*Device]ValueMap

type ValueStorageInstance struct {
	// this represents the state of the storage instance and must only be access by the main go routine

	// state: 1. dimension: Device, 2. dimension: value.Name
	state         State
	subscriptions []subscription

	// communication channels to/from the main go routine
	inputChannel            chan Value
	subscriptionChannel     chan *subscription
	readStateRequestChannel chan *readStateRequest
}

type Filter struct {
	Devices    map[*Device]bool
	ValueNames map[string]bool
}

type subscription struct {
	outputChannel chan Value
	filter        Filter
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
			instance.subscriptions = append(instance.subscriptions, *newSubscription)
		case newReadStateRequest := <-instance.readStateRequestChannel:
			instance.handleNewReadStateRequest(newReadStateRequest)
		}
	}
}

func (instance *ValueStorageInstance) handleNewValue(newValue Value) {
	// check if the newValue is not present or has been changed
	if _, ok := instance.state[newValue.Device]; !ok {
		instance.state[newValue.Device] = make(ValueMap)
	}
	if currentValue, ok := instance.state[newValue.Device][newValue.Name]; !ok || currentValue != newValue {
		// copy the input value to all subscribed output channels
		for _, subscription := range instance.subscriptions {
			subscription.forward(newValue)
		}

		// and save the new state
		instance.state[newValue.Device][newValue.Name] = newValue
	}
}

func (instance *ValueStorageInstance) handleNewReadStateRequest(newReadStateRequest *readStateRequest) {
	filter := &newReadStateRequest.filter

	response := make(State)

	for device, deviceState := range instance.state {
		if !filterByDevice(filter, device) {
			continue
		}

		response[device] = make(ValueMap)

		for valueName, value := range deviceState {
			if !filterByValueName(filter, valueName) {
				continue
			}

			response[device][valueName] = value
		}
	}

	newReadStateRequest.response <- response
}

func ValueStorageCreate() (valueStorageInstance *ValueStorageInstance) {
	valueStorageInstance = &ValueStorageInstance{
		state:                   make(State),
		inputChannel:            make(chan Value, 4), // input channel is buffered
		subscriptionChannel:     make(chan *subscription),
		readStateRequestChannel: make(chan *readStateRequest),
	}

	// main go routine
	go valueStorageInstance.mainStorageRoutine()

	return
}

func (instance *ValueStorageInstance) GetState(filter Filter) (State) {
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
		for valueName, value := range deviceState {
			result[valueName] = value
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

func (instance *ValueStorageInstance) Drain() <-chan Value {
	return instance.Subscribe(Filter{});
}

func (instance *ValueStorageInstance) Subscribe(filter Filter) <-chan Value {
	output := make(chan Value)

	instance.subscriptionChannel <- &subscription{
		outputChannel: output,
		filter:        filter,
	}

	return output
}

func (instance *ValueStorageInstance) Append(fillable Fillable) Fillable {
	fillable.Fill(instance.Drain())
	return fillable
}

func filterByDevice(filter *Filter, device *Device) bool {
	// list is empty -> every device is ok
	if len(filter.Devices) < 1 {
		return true
	}

	// only ok if present and true
	_, ok := filter.Devices[device];
	return ok && filter.Devices[device];
}

func filterByValueName(filter *Filter, valueName string) bool {
	// list is empty -> every device is ok
	if len(filter.ValueNames) < 1 {
		return true
	}

	// only ok if present and true
	_, ok := filter.ValueNames[valueName];
	return ok && filter.ValueNames[valueName];
}

func filterValue(filter *Filter, value *Value) bool {
	return filterByDevice(filter, value.Device) && filterByValueName(filter, value.Name)
}

func (subscription subscription) forward(newValue Value) {
	filter := subscription.filter

	if filterValue(&filter, &newValue) {
		// forward value
		subscription.outputChannel <- newValue
	}
}
