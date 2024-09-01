package generatorDevice

import (
	"context"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
)

type Config interface {
}

type DeviceStruct struct {
	device.State

	stateStorage   *dataflow.ValueStorage
	commandStorage *dataflow.ValueStorage
}

func NewDevice(
	deviceConfig device.Config,
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) *DeviceStruct {
	return &DeviceStruct{
		State: device.NewState(
			deviceConfig,
			stateStorage,
		),
		stateStorage:   stateStorage,
		commandStorage: commandStorage,
	}
}

func (c *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	return nil, false
}

func (c *DeviceStruct) Model() string {
	return "Generator Controller"
}
