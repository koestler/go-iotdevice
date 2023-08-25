package dataflow

import (
	"context"
	"github.com/koestler/go-iotdevice/list"
	"sync"
)

type State map[string]ValueMap

type ValueStorageInstance struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	// this represents the state of the storage instance and must only be access by the main go routine

	// state: 1. dimension: device.Name, 2. dimension: register.Name
	state      State
	stateMutex sync.RWMutex

	subscriptions      *list.List[Subscription]
	subscriptionsMutex sync.RWMutex

	// communication channels to/from the main go routine
	inputChannel   chan Value
	inputWaitGroup sync.WaitGroup
}

type SkipRegisterNameStruct struct {
	Device   string
	Register string
}

type SkipRegisterCategoryStruct struct {
	Device   string
	Category string
}

type Filter struct {
	IncludeDevices         map[string]bool
	SkipRegisterNames      map[SkipRegisterNameStruct]bool
	SkipRegisterCategories map[SkipRegisterCategoryStruct]bool
	SkipNull               bool
}

type readStateRequest struct {
	filter   Filter
	response chan State
}

func (instance *ValueStorageInstance) mainStorageRoutine() {
	for {
		select {
		case <-instance.ctx.Done():
			return
		case newValue := <-instance.inputChannel:
			instance.handleNewValue(newValue)
			instance.inputWaitGroup.Done()
		}
	}
}

func (instance *ValueStorageInstance) handleNewValue(newValue Value) {
	// make sure device exists
	if _, ok := instance.state[newValue.DeviceName()]; !ok {
		instance.state[newValue.DeviceName()] = make(ValueMap)
	}

	// check if the newValue is not present or has been changed
	if currentValue, ok := instance.state[newValue.DeviceName()][newValue.Register().Name()]; !ok || !currentValue.Equals(newValue) {
		// update state
		instance.stateMutex.Lock()
		if _, ok := newValue.(NullRegisterValue); ok {
			delete(instance.state[newValue.DeviceName()], newValue.Register().Name())
		} else {
			// and save the new state
			instance.state[newValue.DeviceName()][newValue.Register().Name()] = newValue
		}
		instance.stateMutex.Unlock()

		// copy the input value to all subscribed output channels
		instance.subscriptionsMutex.RLock()
		for e := instance.subscriptions.Front(); e != nil; e = e.Next() {
			e.Value.forward(newValue)
		}
		instance.subscriptionsMutex.RUnlock()
	}
}

func NewValueStorage() (valueStorageInstance *ValueStorageInstance) {
	ctx, cancel := context.WithCancel(context.Background())

	valueStorageInstance = &ValueStorageInstance{
		ctx:           ctx,
		ctxCancel:     cancel,
		state:         make(State),
		subscriptions: list.New[Subscription](),
		inputChannel:  make(chan Value, 1024),
	}

	// start main go routine
	go valueStorageInstance.mainStorageRoutine()

	return
}

func (instance *ValueStorageInstance) Shutdown() {
	instance.ctxCancel()
}

func (instance *ValueStorageInstance) GetState(filter Filter) State {
	instance.stateMutex.RLock()
	defer instance.stateMutex.RUnlock()

	response := make(State)

	for deviceName, deviceState := range instance.state {
		if !filterByDevice(&filter, deviceName) {
			continue
		}

		response[deviceName] = make(ValueMap)

		for registerName, value := range deviceState {
			if !filterByRegister(&filter, deviceName, value.Register()) {
				continue
			}

			response[deviceName][registerName] = value
		}
	}

	return response
}

func (instance *ValueStorageInstance) GetSlice(filter Filter) (result []Value) {
	state := instance.GetState(filter)

	// create result slice of correct capacity
	capacity := 0
	for _, deviceState := range state {
		capacity += len(deviceState)
	}
	result = make([]Value, 0, capacity)

	for _, deviceState := range state {
		for _, value := range deviceState {
			result = append(result, value)
		}
	}
	return
}

func (instance *ValueStorageInstance) Fill(value Value) {
	instance.inputWaitGroup.Add(1)
	instance.inputChannel <- value
}

// Wait until all inputs are processed (useful for testing)
func (instance *ValueStorageInstance) Wait() {
	instance.inputWaitGroup.Wait()
}

func (instance *ValueStorageInstance) Subscribe(ctx context.Context, filter Filter) Subscription {
	s := Subscription{
		ctx:           ctx,
		outputChannel: make(chan Value, 128),
		filter:        filter,
	}

	instance.subscriptionsMutex.Lock()
	elem := instance.subscriptions.PushBack(s)
	instance.subscriptionsMutex.Unlock()

	go func() {
		// wait for the cancellation of the subscription context
		<-s.ctx.Done()

		// remove from subscriptions list
		instance.subscriptionsMutex.Lock()
		instance.subscriptions.Remove(elem)
		instance.subscriptionsMutex.Unlock()

		// close output channel
		close(s.outputChannel)
	}()

	return s
}
