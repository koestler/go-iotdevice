package modbusDevice

import "github.com/koestler/go-iotdevice/dataflow"

type ModbusRegister struct {
	dataflow.RegisterStruct
	address uint16
}

func NewModbusRegister(
	category, name, description string,
	address uint16,
	enum map[int]string,
	sort int,
) ModbusRegister {
	return ModbusRegister{
		dataflow.NewRegisterStruct(
			category, name, description,
			dataflow.EnumRegister,
			enum,
			"",
			sort,
			true,
		),
		address,
	}
}

func addToRegisterDb(rdb *dataflow.RegisterDb, registers []ModbusRegister) {
	dataflowRegisters := make([]dataflow.RegisterStruct, len(registers))
	for i, r := range registers {
		dataflowRegisters[i] = r.RegisterStruct
	}
	rdb.AddStruct(dataflowRegisters...)
}
