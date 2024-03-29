package victronDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"github.com/koestler/go-iotdevice/v3/types"
	"github.com/koestler/go-victron/veregister"
	"time"
)

type Config interface {
	Device() string
	Kind() types.VictronDeviceKind
	IoLog() string
	PollInterval() time.Duration
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
		rl := veregister.NewRegisterList()
		veregister.AppendBmv(&rl)
		return runRandom(ctx, c, c.StateStorage(), rl)
	case types.VictronRandomSolarKind:
		rl := veregister.NewRegisterList()
		veregister.AppendSolar(&rl)
		return runRandom(ctx, c, c.StateStorage(), rl)
	default:
		return fmt.Errorf("unknown device kind: %s", c.victronConfig.Kind().String()), true
	}
}

func (c *DeviceStruct) Model() string {
	return c.model
}
