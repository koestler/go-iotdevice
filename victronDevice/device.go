package victronDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/types"
)

type Config interface {
	Device() string
	Kind() types.VictronDeviceKind
}

type DeviceStruct struct {
	device.State
	victronConfig Config

	model string
}

func NewDevice(
	deviceConfig device.Config,
	victronConfig Config,
	stateStorage *dataflow.ValueStorage,
) *DeviceStruct {
	return &DeviceStruct{
		State: device.NewState(
			deviceConfig,
			stateStorage,
		),
		victronConfig: victronConfig,
	}
}

func (c *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	switch c.victronConfig.Kind() {
	case types.VictronVedirectKind:
		return runVedirect(ctx, c, c.StateStorage())
	case types.VictronRandomBmvKind:
		return runRandom(ctx, c, c.StateStorage(), RegisterListBmv712)
	case types.VictronRandomSolarKind:
		return runRandom(ctx, c, c.StateStorage(), RegisterListSolar)
	default:
		return fmt.Errorf("unknown device kind: %s", c.victronConfig.Kind().String()), true
	}
}

func (c *DeviceStruct) Model() string {
	return c.model
}
