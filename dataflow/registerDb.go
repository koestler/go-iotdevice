package dataflow

import (
	"context"
	"github.com/koestler/go-iotdevice/list"
	"golang.org/x/exp/maps"
	"sync"
)

type RegisterSubscription struct {
	ctx           context.Context
	outputChannel chan RegisterStruct
}

type RegisterDb struct {
	registers     map[string]RegisterStruct // key: register name
	subscriptions *list.List[RegisterSubscription]
	lock          sync.RWMutex
}

func NewRegisterDb() (registerDb *RegisterDb) {
	return &RegisterDb{
		registers:     make(map[string]RegisterStruct),
		subscriptions: list.New[RegisterSubscription](),
	}
}

func (rdb *RegisterDb) Add(registers ...Register) {
	// convert interface type to structs
	registerStructs := make([]RegisterStruct, len(registers))
	for i, r := range registers {
		registerStructs[i] = NewRegisterStructByInterface(r)
	}

	rdb.lock.Lock()
	defer rdb.lock.Unlock()

	// save to map
	for _, reg := range registerStructs {
		rdb.registers[reg.Name()] = reg
	}

	// forward to subscriptions
	for e := rdb.subscriptions.Front(); e != nil; e = e.Next() {
		for _, reg := range registerStructs {
			e.Value.outputChannel <- reg
		}
	}
}

func (rdb *RegisterDb) GetAll() []RegisterStruct {
	rdb.lock.RLock()
	defer rdb.lock.RUnlock()

	return maps.Values(rdb.registers)
}

func (rdb *RegisterDb) GetByName(registerName string) (reg RegisterStruct, ok bool) {
	rdb.lock.RLock()
	defer rdb.lock.RUnlock()

	reg, ok = rdb.registers[registerName]
	return
}

func (rdb *RegisterDb) Subscribe(ctx context.Context) RegisterSubscription {
	s := RegisterSubscription{
		ctx:           ctx,
		outputChannel: make(chan RegisterStruct, 16),
	}

	rdb.lock.Lock()
	defer rdb.lock.Unlock()

	// add subscription
	elem := rdb.subscriptions.PushBack(s)

	// create routine to send initial values and shut down the subscription once the context is canceled
	go func(initialRegisters []RegisterStruct) {
		// sending initial set of registers to the output chan
		for _, reg := range initialRegisters {
			s.outputChannel <- reg
		}

		<-s.ctx.Done()

		// remove from subscriptions list
		rdb.lock.Lock()
		rdb.subscriptions.Remove(elem)
		rdb.lock.Unlock()

		// close output channel
		close(s.outputChannel)
	}(maps.Values(rdb.registers))

	return s
}
