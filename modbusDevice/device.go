package modbusDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/types"
	"time"
)

type Config interface {
	Bus() string
	Kind() types.ModbusDeviceKind
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

	commandStorage *dataflow.ValueStorage
	modbus         Modbus
}

func NewDevice(
	deviceConfig device.Config,
	modbusConfig Config,
	modbus Modbus,
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) *DeviceStruct {
	return &DeviceStruct{
		State: device.NewState(
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
	case types.ModbusWaveshareRtuRelay8Kind:
		return runWaveshareRtuRelay8(ctx, c)
	case types.ModbusFinder7M38Kind:
		return runFinder7M38(ctx, c)
	default:
		return fmt.Errorf("unknown device kind: %s", c.modbusConfig.Kind().String()), true
	}
}

func (c *DeviceStruct) Model() string {
	return c.modbusConfig.Kind().String()
}
