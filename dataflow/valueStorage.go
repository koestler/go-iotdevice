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

	state         map[StateKey]Value
	subscriptions *list.List[ValueSubscription]
	mutex         sync.RWMutex

	inputChannel   chan Value
	inputWaitGroup sync.WaitGroup
}

func NewValueStorage() (valueStorage *ValueStorage) {
	ctx, cancel := context.WithCancel(context.Background())

	valueStorage = &ValueStorage{
		ctx:           ctx,
		ctxCancel:     cancel,
		state:         make(map[StateKey]Value, 64),
		subscriptions: list.New[ValueSubscription](),
		inputChannel:  make(chan Value, 1024),
	}

	// start main go routine
	go valueStorage.mainStorageRoutine()

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
			vs.mutex.Lock()
			if vs.updateState(newValue) {
				vs.forwardToSubscriptions(newValue)
			}
			vs.mutex.Unlock()
			vs.inputWaitGroup.Done()
		}
	}
}

func (vs *ValueStorage) updateState(newValue Value) (updated bool) {
	k := StateKey{
		deviceName:   newValue.DeviceName(),
		registerName: newValue.Register().Name(),
	}

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
	for e := vs.subscriptions.Front(); e != nil; e = e.Next() {
		s := e.Value
		if s.filter(newValue) {
			s.outputChannel <- newValue
		}
	}
}

func (vs *ValueStorage) GetState() (result []Value) {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	result = make([]Value, 0, len(vs.state))
	for _, value := range vs.state {
		result = append(result, value)
	}

	return
}

func (vs *ValueStorage) GetStateFiltered(filter ValueFilterFunc) (result []Value) {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	result = vs.getStateFilteredUnlocked(filter)
	return
}

func (vs *ValueStorage) getStateFilteredUnlocked(filter ValueFilterFunc) (result []Value) {
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

const subscriptionDefaultCap = 128

func (vs *ValueStorage) newSubscription(ctx context.Context, filter ValueFilterFunc) (
	initial []Value, subscription ValueSubscription, elem *list.Element[ValueSubscription],
) {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	initial = vs.getStateFilteredUnlocked(filter)

	subscription = ValueSubscription{
		ctx:           ctx,
		outputChannel: make(chan Value, subscriptionDefaultCap),
		filter:        filter,
	}
	elem = vs.subscriptions.PushBack(subscription)

	return
}

func (vs *ValueStorage) SubscribeReturnInitial(ctx context.Context, filter ValueFilterFunc) (initial []Value, subscription ValueSubscription) {
	initial, subscription, elem := vs.newSubscription(ctx, filter)
	go vs.sendInitialAndCleanupValueSubscription([]Value{}, subscription, elem)
	return
}

func (vs *ValueStorage) SubscribeSendInitial(ctx context.Context, filter ValueFilterFunc) (subscription ValueSubscription) {
	initial, subscription, elem := vs.newSubscription(ctx, filter)
	go vs.sendInitialAndCleanupValueSubscription(initial, subscription, elem)
	return
}

func (vs *ValueStorage) sendInitialAndCleanupValueSubscription(
	initial []Value,
	subscription ValueSubscription,
	elem *list.Element[ValueSubscription],
) {
	// cleanup subscription
	defer func() {
		// remove from subscriptions list
		vs.mutex.Lock()
		vs.subscriptions.Remove(elem)
		vs.mutex.Unlock()

		// close output channel
		close(subscription.outputChannel)
	}()

	// this never blocks since channel is always buffered, big enough and main routine is blocked
	for _, initialV := range initial {
		select {
		case subscription.outputChannel <- initialV:
		case <-subscription.ctx.Done():
			return
		}
	}

	// wait for the cancellation of the subscription context
	<-subscription.ctx.Done()
}
