package dataflow

import (
	"context"
	"github.com/koestler/go-iotdevice/list"
	"sync"
)

type State map[string]ValueMap

type ValueStorage struct {
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

func (vs *ValueStorage) mainStorageRoutine() {
	for {
		select {
		case <-vs.ctx.Done():
			return
		case newValue := <-vs.inputChannel:
			vs.handleNewValue(newValue)
			vs.inputWaitGroup.Done()
		}
	}
}

func (vs *ValueStorage) handleNewValue(newValue Value) {
	// make sure device exists
	if _, ok := vs.state[newValue.DeviceName()]; !ok {
		vs.state[newValue.DeviceName()] = make(ValueMap)
	}

	// check if the newValue is not present or has been changed
	if currentValue, ok := vs.state[newValue.DeviceName()][newValue.Register().Name()]; !ok || !currentValue.Equals(newValue) {
		// update state
		vs.stateMutex.Lock()
		if _, ok := newValue.(NullRegisterValue); ok {
			delete(vs.state[newValue.DeviceName()], newValue.Register().Name())
		} else {
			// and save the new state
			vs.state[newValue.DeviceName()][newValue.Register().Name()] = newValue
		}
		vs.stateMutex.Unlock()

		// copy the input value to all subscribed output channels
		vs.subscriptionsMutex.RLock()
		for e := vs.subscriptions.Front(); e != nil; e = e.Next() {
			e.Value.forward(newValue)
		}
		vs.subscriptionsMutex.RUnlock()
	}
}

func NewValueStorage() (valueStorageInstance *ValueStorage) {
	ctx, cancel := context.WithCancel(context.Background())

	valueStorageInstance = &ValueStorage{
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

func (vs *ValueStorage) Shutdown() {
	vs.ctxCancel()
}

func (vs *ValueStorage) GetState(filter Filter) State {
	vs.stateMutex.RLock()
	defer vs.stateMutex.RUnlock()

	response := make(State)

	for deviceName, deviceState := range vs.state {
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

func (vs *ValueStorage) GetSlice(filter Filter) (result []Value) {
	state := vs.GetState(filter)

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

func (vs *ValueStorage) Fill(value Value) {
	vs.inputWaitGroup.Add(1)
	vs.inputChannel <- value
}

// Wait until all inputs are processed (useful for testing)
func (vs *ValueStorage) Wait() {
	vs.inputWaitGroup.Wait()
}

func (vs *ValueStorage) Subscribe(ctx context.Context, filter Filter) Subscription {
	s := Subscription{
		ctx:           ctx,
		outputChannel: make(chan Value, 128),
		filter:        filter,
	}

	vs.subscriptionsMutex.Lock()
	elem := vs.subscriptions.PushBack(s)
	vs.subscriptionsMutex.Unlock()

	go func() {
		// wait for the cancellation of the subscription context
		<-s.ctx.Done()

		// remove from subscriptions list
		vs.subscriptionsMutex.Lock()
		vs.subscriptions.Remove(elem)
		vs.subscriptionsMutex.Unlock()

		// close output channel
		close(s.outputChannel)
	}()

	return s
}
