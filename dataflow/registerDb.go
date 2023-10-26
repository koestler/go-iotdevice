package dataflow

import (
	"context"
	"github.com/koestler/go-iotdevice/list"
	"sync"
)

type RegisterSubscription struct {
	ctx           context.Context
	outputChannel chan Register
	filter        RegisterFilterFunc
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
	rdb.AddStruct(registerStructs...)
}

func (rdb *RegisterDb) AddStruct(registerStructs ...RegisterStruct) {
	rdb.lock.Lock()
	defer rdb.lock.Unlock()

	// save to map
	for _, reg := range registerStructs {
		rdb.registers[reg.Name()] = reg
	}

	// forward to subscriptions
	for e := rdb.subscriptions.Front(); e != nil; e = e.Next() {
		for _, reg := range registerStructs {
			s := e.Value
			if s.filter(reg) {
				s.outputChannel <- reg
			}
		}
	}
}

func (rdb *RegisterDb) GetAll() []Register {
	rdb.lock.RLock()
	defer rdb.lock.RUnlock()

	ret := make([]Register, 0, len(rdb.registers))
	for _, r := range rdb.registers {
		ret = append(ret, r)
	}
	return ret
}

func (rdb *RegisterDb) GetFiltered(filter RegisterFilterFunc) []Register {
	rdb.lock.RLock()
	defer rdb.lock.RUnlock()

	ret := rdb.getFilteredUnlocked(filter)
	return ret
}

func (rdb *RegisterDb) getFilteredUnlocked(filter RegisterFilterFunc) (ret []Register) {
	ret = make([]Register, 0)
	for _, r := range rdb.registers {
		if filter(r) {
			ret = append(ret, r)
		}
	}

	return ret
}

func (rdb *RegisterDb) GetByName(registerName string) Register {
	rdb.lock.RLock()
	defer rdb.lock.RUnlock()

	if reg, ok := rdb.registers[registerName]; ok {
		return reg
	}
	return nil
}

func (rdb *RegisterDb) Subscribe(ctx context.Context, filter RegisterFilterFunc) <-chan Register {
	s := RegisterSubscription{
		ctx:           ctx,
		outputChannel: make(chan Register, 16),
		filter:        filter,
	}

	rdb.lock.Lock()
	defer rdb.lock.Unlock()

	// add subscription
	elem := rdb.subscriptions.PushBack(s)

	// create routine to send initial values and shut down the subscription once the context is canceled
	go func(initialRegisters []Register) {
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
	}(rdb.getFilteredUnlocked(filter))

	return s.outputChannel
}
