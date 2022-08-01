package device

import (
	"fmt"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
)

type Config interface {
	Name() string
	Kind() config.DeviceKind
	Device() string
	LogDebug() bool
	LogComDebug() bool
}

type Device interface {
	Config() Config
	Registers() dataflow.Registers
	SetRegisters(registers dataflow.Registers)
	Shutdown()
}

type Creator func(deviceStruct DeviceStruct, output chan dataflow.Value) (device Device, err error)

var creators = make(map[config.DeviceKind]Creator)

func RegisterCreator(kind config.DeviceKind, creator Creator) {
	creators[kind] = creator
}

type DeviceStruct struct {
	// configuration
	cfg Config

	source    *dataflow.Source
	registers dataflow.Registers

	shutdown chan struct{}
	closed   chan struct{}
}

func RunDevice(cfg Config, target dataflow.Fillable) (device Device, err error) {
	// setup output chain
	output := make(chan dataflow.Value, 128)
	source := dataflow.CreateSource(output)
	// pipe all data to next stage
	source.Append(target)

	if c, ok := creators[cfg.Kind()]; ok {
		return c(
			DeviceStruct{
				cfg:      cfg,
				source:   source,
				shutdown: make(chan struct{}),
				closed:   make(chan struct{}),
			},
			output,
		)
	}

	return nil, fmt.Errorf("unknown kind: %s", cfg.Kind().String())
}

func (c DeviceStruct) Config() Config {
	return c.cfg
}

func (c DeviceStruct) Registers() dataflow.Registers {
	return c.registers
}

func (c *DeviceStruct) SetRegisters(registers dataflow.Registers) {
	c.registers = registers
}

func (c DeviceStruct) Shutdown() {
	close(c.shutdown)
	<-c.closed
	log.Printf("device[%s]: shutdown completed", c.cfg.Name())
}

func (c DeviceStruct) GetShutdownChan() chan struct{} {
	return c.shutdown
}

func (c DeviceStruct) GetClosedChan() chan struct{} {
	return c.closed
}
