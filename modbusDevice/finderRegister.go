package modbusDevice

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
)

type FinderRegister struct {
	dataflow.RegisterStruct
	registerType FinderRegisterType
	addressBegin uint16
	addressEnd   uint16
}

type FinderRegisterType int

const (
	FinderT1 FinderRegisterType = iota
	FinderTStr2
	FinderTStr8
	FinderTStr16
	FinderTFloat
)

func NewFinderRegister(
	category, name, description string,
	registerType FinderRegisterType,
	addressBegin, addressEnd uint16,
	enum map[int]string,
	unit string,
	sort int,
) FinderRegister {
	var rt dataflow.RegisterType

	expectF := func(expectBytes uint16) {
		gotBytes := (addressEnd - addressBegin + 1) * 2
		if gotBytes != expectBytes {
			log.Fatalf("FinderDevice: registerName=%s: expect %d bytes but got %d", name, expectBytes, gotBytes)
		}
	}

	switch registerType {
	case FinderT1:
		if enum != nil {
			rt = dataflow.EnumRegister
		} else {
			rt = dataflow.NumberRegister
		}
		expectF(2)
	case FinderTStr2:
		rt = dataflow.TextRegister
		expectF(2)
	case FinderTStr8:
		rt = dataflow.TextRegister
		expectF(8)
	case FinderTStr16:
		rt = dataflow.TextRegister
		expectF(16)
	case FinderTFloat:
		rt = dataflow.NumberRegister
		expectF(4)
	}

	return FinderRegister{
		dataflow.NewRegisterStruct(
			category, name, description,
			rt,
			enum,
			unit,
			sort,
			false,
		),
		registerType,
		addressBegin,
		addressEnd,
	}
}

func addToRegisterDb(rdb *dataflow.RegisterDb, registers []FinderRegister) {
	dataflowRegisters := make([]dataflow.RegisterStruct, len(registers))
	for i, r := range registers {
		dataflowRegisters[i] = r.RegisterStruct
	}
	rdb.AddStruct(dataflowRegisters...)
}

func (r FinderRegister) CountRegisters() int {
	return int(r.addressEnd) - int(r.addressBegin) + 1
}

func (r FinderRegister) CountBytes() int {
	// finder registers are 16 bit wide
	return r.CountRegisters() * 2
}
