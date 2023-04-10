package modbusDevice

import "github.com/koestler/go-iotdevice/dataflow"

type ModbusRegisters []ModbusRegister
type ModbusRegister interface {
	dataflow.Register
	Address() uint16
}

type ModbusRegisterStruct struct {
	dataflow.RegisterStruct
	address uint16
}

func CreateEnumRegisterStruct(
	category, name, description string,
	address uint16,
	enum map[int]string,
	sort int,
) ModbusRegisterStruct {
	return ModbusRegisterStruct{
		dataflow.CreateRegisterStruct(
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

func (r ModbusRegisterStruct) Address() uint16 {
	return r.address
}
