package mqttDevice

import (
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/mqttForwarders"
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
	return dataflow.RegisterTypeFromString(s.StructRegister.Type) //nolint:staticcheck
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

func (s StructRegister) Writable() bool {
	return s.StructRegister.Writable
}
