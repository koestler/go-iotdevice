package dataflow

import (
	"context"
	"github.com/koestler/go-list"
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

type ValueSubscription struct {
	ctx           context.Context
	outputChannel chan Value
	filter        ValueFilterFunc
	sendMutex     *sync.Mutex // prevents sending on the channel while it is being closed
}

func (s *ValueSubscription) send(v Value) {
	s.sendMutex.Lock()
	defer s.sendMutex.Unlock()
	if s.ctx.Err() != nil {
		return
	}
	select {
	case s.outputChannel <- v:
	case <-s.ctx.Done():
	}
}

func (s *ValueSubscription) Drain() <-chan Value {
	return s.outputChannel
}

func NewValueStorage() (valueStorage *ValueStorage) {
	ctx, cancel := context.WithCancel(context.Background())

	valueStorage = &ValueStorage{
		ctx:           ctx,
		ctxCancel:     cancel,
		state:         make(map[StateKey]Value),
		subscriptions: list.New[ValueSubscription](),
		inputChannel:  make(chan Value, 32),
	}

	// start main goroutine
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
			var subscriptions []ValueSubscription
			if vs.updateState(newValue) {
				subscriptions = vs.snapshotSubscriptions()
			}
			vs.mutex.Unlock()

			// send outside the lock so a slow subscriber cannot block subscription cleanup
			forwardToSubscriptions(subscriptions, newValue)
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
		// a null value means -> remove from state
		delete(vs.state, k)
	} else {
		// update value
		vs.state[k] = newValue
	}

	return true
}

// must be called with vs.mutex held
func (vs *ValueStorage) snapshotSubscriptions() (result []ValueSubscription) {
	result = make([]ValueSubscription, 0, vs.subscriptions.Len())
	for e := vs.subscriptions.Front(); e != nil; e = e.Next() {
		result = append(result, e.Value)
	}
	return
}

// copy the input value to all subscribed output channels
func forwardToSubscriptions(subscriptions []ValueSubscription, newValue Value) {
	for _, s := range subscriptions {
		if s.filter(newValue) {
			s.send(newValue)
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
		sendMutex:     &sync.Mutex{},
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

		// close output channel; wait for in-flight sends (they abort as the context is done)
		subscription.sendMutex.Lock()
		close(subscription.outputChannel)
		subscription.sendMutex.Unlock()
	}()

	// the channel is buffered; this blocks if the initial values do not fit and the consumer is slow, until the context is done
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
