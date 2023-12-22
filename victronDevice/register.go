package victronDevice

import (
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-victron/veregisters"
)

// Register is used as a wrapper for veregisters.Register to implement dataflow.Register
type Register struct {
	veregisters.Register
}

func (r Register) RegisterType() dataflow.RegisterType {
	if _, ok := r.Register.(veregisters.NumberRegister); ok {
		return dataflow.NumberRegister
	}
	if _, ok := r.Register.(veregisters.TextRegister); ok {
		return dataflow.TextRegister
	}
	if _, ok := r.Register.(veregisters.EnumRegister); ok {
		return dataflow.EnumRegister
	}
	return dataflow.UndefinedRegister
}

func (r Register) Enum() map[int]string {
	if er, ok := r.Register.(veregisters.EnumRegister); ok {
		return er.Enum()
	}
	return nil
}

func (r Register) Unit() string {
	if nr, ok := r.Register.(veregisters.NumberRegister); ok {
		return nr.Unit()
	}
	return ""
}

func addToRegisterDb(rdb *dataflow.RegisterDb, rl veregisters.RegisterList) {
	registers := rl.GetRegisters()
	dataflowRegisters := make([]dataflow.RegisterStruct, len(registers))
	for i, r := range registers {
		dataflowRegisters[i] = dataflow.NewRegisterStructByInterface(Register{r})
	}
	rdb.AddStruct(dataflowRegisters...)
}
