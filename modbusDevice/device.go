package modbusDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"time"
)

type Config interface {
	Bus() string
	Kind() config.ModbusDeviceKind
	Address() byte
	RelayDescription(name string) string
	RelayOpenLabel(name string) string
	RelayClosedLabel(name string) string
	PollInterval() time.Duration
}

type Modbus interface {
	Name() string
	Shutdown()
	WriteRead(request []byte, responseBuf []byte) error
}

type DeviceStruct struct {
	device.State
	modbusConfig Config

	commandStorage *dataflow.ValueStorageInstance

	modbus    Modbus
	registers ModbusRegisters
}

func CreateDevice(
	deviceConfig device.Config,
	modbusConfig Config,
	modbus Modbus,
	stateStorage *dataflow.ValueStorageInstance,
	commandStorage *dataflow.ValueStorageInstance,
) *DeviceStruct {
	return &DeviceStruct{
		State: device.CreateState(
			deviceConfig,
			stateStorage,
		),
		modbusConfig:   modbusConfig,
		commandStorage: commandStorage,
		modbus:         modbus,
	}
}

func (c *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	switch c.modbusConfig.Kind() {
	case config.ModbusWaveshareRtuRelay8Kind:
		return runWaveshareRtuRelay8(ctx, c)
	default:
		return fmt.Errorf("unknown device kind: %s", c.modbusConfig.Kind().String()), true
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
	return c.modbusConfig.Kind().String()
}
