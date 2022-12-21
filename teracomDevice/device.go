package teracomDevice

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"log"
	"net/url"
	"sync"
	"time"
)

type Config interface {
	Url() *url.URL
	Username() string
	Password() string
	PollInterval() time.Duration
}

type DeviceStruct struct {
	deviceConfig  device.Config
	teracomConfig Config

	source *dataflow.Source

	registers        map[string]dataflow.Register
	registersMutex   sync.RWMutex
	lastUpdated      time.Time
	lastUpdatedMutex sync.RWMutex

	shutdown chan struct{}
}

func RunDevice(
	deviceConfig device.Config,
	teracomConfig Config,
	storage *dataflow.ValueStorageInstance,
	mqttClientPool *mqttClient.ClientPool,
) (device device.Device, err error) {
	// setup output chain
	output := make(chan dataflow.Value, 128)
	source := dataflow.CreateSource(output)
	// pipe all data to next stage
	source.Append(storage)

	c := &DeviceStruct{
		deviceConfig:  deviceConfig,
		teracomConfig: teracomConfig,
		source:        source,
		registers:     make(map[string]dataflow.Register),
		shutdown:      make(chan struct{}),
	}

	return c, nil
}

func (c *DeviceStruct) Config() device.Config {
	return c.deviceConfig
}

func (c *DeviceStruct) ShutdownChan() chan struct{} {
	return c.shutdown
}

func (c *DeviceStruct) Registers() dataflow.Registers {
	c.registersMutex.RLock()
	defer c.registersMutex.RUnlock()

	ret := make(dataflow.Registers, len(c.registers))
	i := 0
	for _, r := range c.registers {
		ret[i] = r
		i += 1
	}
	return ret
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
	return "teracom"
}

func (c *DeviceStruct) Shutdown() {
	close(c.shutdown)
	log.Printf("teracomDevice[%s]: shutdown completed", c.deviceConfig.Name())
}
