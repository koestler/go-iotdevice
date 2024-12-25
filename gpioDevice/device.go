//go:build !linux

package gpioDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
)

func NewDevice(
	deviceConfig device.Config,
	gpioConfig Config,
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) *DeviceStruct {
	return &DeviceStruct{
		State: device.NewState(
			deviceConfig,
			stateStorage,
		),
		gpioConfig:     gpioConfig,
		commandStorage: commandStorage,
	}
}

func (d *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	return fmt.Errorf("gpioDevice[%s]: not supported on this platform", d.Config().Name()), true
}
