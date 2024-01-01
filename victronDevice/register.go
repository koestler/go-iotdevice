package victronDevice

import (
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-victron/veconst"
	"github.com/koestler/go-victron/veregister"
)

// Register is used as a wrapper for veregister.Register to implement dataflow.Register
type Register struct {
	veregister.Register
}

func (r Register) RegisterType() dataflow.RegisterType {
	switch r.Type() {
	case veregister.Number:
		return dataflow.NumberRegister
	case veregister.Text:
		return dataflow.TextRegister
	case veregister.Enum:
		return dataflow.EnumRegister
	default:
		return dataflow.UndefinedRegister
	}
}

type enumFactory interface {
	Factory() veconst.EnumFactory
}

func (r Register) Enum() map[int]string {
	if ef, ok := r.Register.(enumFactory); ok {
		return ef.Factory().IntToStringMap()
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

func addToRegisterDb(rdb *dataflow.RegisterDb, rl veregister.RegisterList) {
	registers := rl.GetRegisters()
	dataflowRegisters := make([]dataflow.RegisterStruct, len(registers))
	for i, r := range registers {
		dataflowRegisters[i] = dataflow.NewRegisterStructByInterface(Register{r})
	}
	rdb.AddStruct(dataflowRegisters...)
}
