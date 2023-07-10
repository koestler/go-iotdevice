package victronDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
)

type Config interface {
	Device() string
	Kind() config.VictronDeviceKind
}

type DeviceStruct struct {
	device.State
	victronConfig Config

	registers VictronRegisters
	model     string
}

func CreateDevice(
	deviceConfig device.Config,
	victronConfig Config,
	stateStorage *dataflow.ValueStorageInstance,
) *DeviceStruct {
	return &DeviceStruct{
		State: device.CreateState(
			deviceConfig,
			stateStorage,
		),
		victronConfig: victronConfig,
	}
}

func (c *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	switch c.victronConfig.Kind() {
	case config.VictronVedirectKind:
		return runVedirect(ctx, c, c.StateStorage())
	case config.VictronRandomBmvKind:
		return runRandom(ctx, c, c.StateStorage(), RegisterListBmv712)
	case config.VictronRandomSolarKind:
		return runRandom(ctx, c, c.StateStorage(), RegisterListSolar)
	default:
		return fmt.Errorf("unknown device kind: %s", c.victronConfig.Kind().String()), true
	}
}

func (c *DeviceStruct) Registers() dataflow.Registers {
	ret := make(dataflow.Registers, len(c.registers)+1)
	for i, r := range c.registers {
		ret[i] = r.(dataflow.Register)
	}
	ret[len(c.registers)] = device.GetAvailabilityRegister()
	return ret
}

func (c *DeviceStruct) GetRegister(registerName string) dataflow.Register {
	for _, r := range c.registers {
		if r.Name() == registerName {
			return r
		}
	}
	return nil
}

func (c *DeviceStruct) Model() string {
	return c.model
}
