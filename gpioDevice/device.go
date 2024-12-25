//go:build !linux

package gpioDevice

import (
	"context"
	"errors"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
)

func NewDevice(
	deviceConfig device.Config,
	gpioConfig Config,
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) (*DeviceStruct, error) {
	return nil, errors.New("not supported on this platform")
}

func (d *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	return
}
