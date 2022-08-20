package device

import (
	"fmt"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
	"sync"
	"time"
)

type Config interface {
	Name() string
	Kind() config.DeviceKind
	Device() string
	SkipFields() []string
	SkipCategories() []string
	LogDebug() bool
	LogComDebug() bool
}

type Device interface {
	Config() Config
	GetRegisters() dataflow.Registers
	SetRegisters(registers dataflow.Registers)
	SetLastUpdatedNow()
	GetLastUpdated() time.Time
	SetModel(model string)
	GetModel() string
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

	source *dataflow.Source

	registers        dataflow.Registers
	registersMutex   sync.RWMutex
	lastUpdated      time.Time
	lastUpdatedMutex sync.RWMutex
	model            string
	modelMutex       sync.RWMutex

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

func (c *DeviceStruct) Config() Config {
	return c.cfg
}

func (c *DeviceStruct) SetRegisters(registers dataflow.Registers) {
	c.registersMutex.Lock()
	defer c.registersMutex.Unlock()

	c.registers = registers
}

func (c *DeviceStruct) GetRegisters() dataflow.Registers {
	c.registersMutex.RLock()
	defer c.registersMutex.RUnlock()
	return c.registers
}

func (c *DeviceStruct) SetLastUpdatedNow() {
	c.lastUpdatedMutex.Lock()
	defer c.lastUpdatedMutex.Unlock()
	c.lastUpdated = time.Now()
}

func (c *DeviceStruct) GetLastUpdated() time.Time {
	c.lastUpdatedMutex.RLock()
	defer c.lastUpdatedMutex.RUnlock()
	return c.lastUpdated
}

func (c *DeviceStruct) SetModel(model string) {
	c.modelMutex.Lock()
	defer c.modelMutex.Unlock()
	c.model = model
}

func (c *DeviceStruct) GetModel() string {
	c.modelMutex.RLock()
	defer c.modelMutex.RUnlock()
	return c.model
}

func (c *DeviceStruct) Shutdown() {
	close(c.shutdown)
	<-c.closed
	log.Printf("device[%s]: shutdown completed", c.cfg.Name())
}

func (c *DeviceStruct) GetShutdownChan() chan struct{} {
	return c.shutdown
}

func (c *DeviceStruct) GetClosedChan() chan struct{} {
	return c.closed
}
