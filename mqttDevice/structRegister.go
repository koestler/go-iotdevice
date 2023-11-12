package mqttDevice

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mqttForwarders"
)

type StructRegister struct {
	mqttForwarders.StructRegister
}

func (s StructRegister) Category() string {
	return s.StructRegister.Category
}

func (s StructRegister) Name() string {
	return s.StructRegister.Name
}

func (s StructRegister) Description() string {
	return s.StructRegister.Description
}

func (s StructRegister) RegisterType() dataflow.RegisterType {
	return dataflow.RegisterTypeFromString(s.StructRegister.Type)
}

func (s StructRegister) Enum() map[int]string {
	return s.StructRegister.Enum
}

func (s StructRegister) Unit() string {
	return s.StructRegister.Unit
}

func (s StructRegister) Sort() int {
	return s.StructRegister.Sort
}

func (s StructRegister) Commandable() bool {
	return s.StructRegister.Commandable
}
