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
	switch r.Type() {
	case veregisters.Number:
		return dataflow.NumberRegister
	case veregisters.Text:
		return dataflow.TextRegister
	case veregisters.Enum:
		return dataflow.EnumRegister
	default:
		return dataflow.UndefinedRegister
	}
}

type enumer interface {
	Enum() map[int]string
}

func (r Register) Enum() map[int]string {
	if er, ok := r.Register.(enumer); ok {
		return er.Enum()
	}
	return nil
}

type uniter interface {
	Unit() string
}

func (r Register) Unit() string {
	if nr, ok := r.Register.(uniter); ok {
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
