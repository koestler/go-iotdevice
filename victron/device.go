package victron

import (
	"fmt"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/vedirect"
	"log"
	"sync"
	"time"
)

type Config interface {
	Name() string
	Kind() config.VictronDeviceKind
	Device() string
	SkipFields() []string
	SkipCategories() []string
	TelemetryViaMqttClients() []string
	RealtimeViaMqttClients() []string
	LogDebug() bool
	LogComDebug() bool
}

type DeviceStruct struct {
	cfg Config

	source *dataflow.Source

	deviceId         vedirect.VeProduct
	registers        dataflow.Registers
	lastUpdated      time.Time
	lastUpdatedMutex sync.RWMutex
	model            string

	shutdown chan struct{}
	closed   chan struct{}
}

func RunDevice(cfg Config, target dataflow.Fillable) (device device.Device, err error) {
	// setup output chain
	output := make(chan dataflow.Value, 128)
	source := dataflow.CreateSource(output)
	// pipe all data to next stage
	source.Append(target)

	c := &DeviceStruct{
		cfg:      cfg,
		source:   source,
		shutdown: make(chan struct{}),
		closed:   make(chan struct{}),
	}

	if cfg.Kind() == config.VedirectKind {
		err = startVedirect(c, output)
	} else if cfg.Kind() == config.RandomBmvKind {
		err = startRandom(c, output, RegisterListBmv712)
	} else if cfg.Kind() == config.RandomSolarKind {
		err = startRandom(c, output, RegisterListSolar)
	} else {
		return nil, fmt.Errorf("unknown device kind: %s", cfg.Kind().String())
	}

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *DeviceStruct) Name() string {
	return c.cfg.Name()
}

func (c *DeviceStruct) Registers() dataflow.Registers {
	return c.registers
}

func (c *DeviceStruct) SetLastUpdatedNow() {
	c.lastUpdatedMutex.Lock()
	defer c.lastUpdatedMutex.Unlock()
	c.lastUpdated = time.Now()
}

func (c *DeviceStruct) LastUpdated() time.Time {
	c.lastUpdatedMutex.RLock()
	defer c.lastUpdatedMutex.RUnlock()
	return c.lastUpdated
}

func (c *DeviceStruct) Model() string {
	return c.model
}

func (c *DeviceStruct) Shutdown() {
	close(c.shutdown)
	<-c.closed
	log.Printf("device[%s]: shutdown completed", c.cfg.Name())
}
