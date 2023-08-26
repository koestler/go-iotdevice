package dataflow

import (
	"context"
	"github.com/koestler/go-iotdevice/list"
	"sync"
)

type StateKey struct {
	deviceName   string
	registerName string
}

type ValueStorage struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	state      map[StateKey]Value
	stateMutex sync.RWMutex

	inputChannel   chan Value
	inputWaitGroup sync.WaitGroup

	subscriptions      *list.List[Subscription]
	subscriptionsMutex sync.RWMutex
}

func NewValueStorage() (valueStorageInstance *ValueStorage) {
	ctx, cancel := context.WithCancel(context.Background())

	valueStorageInstance = &ValueStorage{
		ctx:           ctx,
		ctxCancel:     cancel,
		state:         make(map[StateKey]Value, 64),
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

func (vs *ValueStorage) mainStorageRoutine() {
	for {
		select {
		case <-vs.ctx.Done():
			return
		case newValue := <-vs.inputChannel:
			if vs.updateState(newValue) {
				vs.forwardToSubscriptions(newValue)
			}
			vs.inputWaitGroup.Done()
		}
	}
}

func (vs *ValueStorage) updateState(newValue Value) (updated bool) {
	k := StateKey{
		deviceName:   newValue.DeviceName(),
		registerName: newValue.Register().Name(),
	}

	vs.stateMutex.Lock()
	defer vs.stateMutex.Unlock()

	currentValue, ok := vs.state[k]

	if ok && currentValue.Equals(newValue) {
		return false
	}

	if _, ok := newValue.(NullRegisterValue); ok {
		// null values means -> remove from state
		delete(vs.state, k)
	} else {
		// update value
		vs.state[k] = newValue
	}

	return true
}

// copy the input value to all subscribed output channels
func (vs *ValueStorage) forwardToSubscriptions(newValue Value) {
	vs.subscriptionsMutex.RLock()
	defer vs.subscriptionsMutex.RUnlock()

	for e := vs.subscriptions.Front(); e != nil; e = e.Next() {
		s := e.Value
		if s.filter(newValue) {
			s.outputChannel <- newValue
		}
	}
}

func (vs *ValueStorage) GetState() (result []Value) {
	vs.stateMutex.RLock()
	defer vs.stateMutex.RUnlock()

	result = make([]Value, 0, len(vs.state))
	for _, value := range vs.state {
		result = append(result, value)
	}

	return
}

func (vs *ValueStorage) GetStateFiltered(filter FilterFunc) (result []Value) {
	vs.stateMutex.RLock()
	defer vs.stateMutex.RUnlock()

	result = make([]Value, 0)
	for _, value := range vs.state {
		if filter(value) {
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

func (vs *ValueStorage) Subscribe(ctx context.Context, filter FilterFunc) Subscription {
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
