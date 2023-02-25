package modbusDevice

import (
	"fmt"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"log"
	"sync"
	"time"
)

type Config interface {
	Device() string
	Kind() config.ModbusDeviceKind
}

type DeviceStruct struct {
	deviceConfig device.Config
	modbusConfig Config
	storage      *dataflow.ValueStorageInstance

	output chan dataflow.Value

	registers        map[string]dataflow.Register
	registersMutex   sync.RWMutex
	lastUpdated      time.Time
	lastUpdatedMutex sync.RWMutex

	shutdown chan struct{}
}

func RunDevice(
	deviceConfig device.Config,
	modbusConfig Config,
	storage *dataflow.ValueStorageInstance,
) (device device.Device, err error) {
	// setup output chain
	output := make(chan dataflow.Value, 128)
	source := dataflow.CreateSource(output)
	// pipe all data to next stage
	source.Append(storage)

	ds := &DeviceStruct{
		deviceConfig: deviceConfig,
		modbusConfig: modbusConfig,
		storage:      storage,
		output:       output,
		registers:    make(map[string]dataflow.Register),
		shutdown:     make(chan struct{}),
	}

	c := &DeviceStruct{
		deviceConfig: deviceConfig,
		modbusConfig: modbusConfig,
		source:       source,
		shutdown:     make(chan struct{}),
		closed:       make(chan struct{}),
	}

	if modbusConfig.Kind() == config.ModbusVedirectKind {
		err = startVedirect(c, output)
	} else if modbusConfig.Kind() == config.ModbusRandomBmvKind {
		err = startRandom(c, output, RegisterListBmv712)
	} else if modbusConfig.Kind() == config.ModbusRandomSolarKind {
		err = startRandom(c, output, RegisterListSolar)
	} else {
		return nil, fmt.Errorf("unknown device kind: %s", modbusConfig.Kind().String())
	}

	if err != nil {
		return nil, err
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
	ret := make(dataflow.Registers, len(c.registers))
	for i, r := range c.registers {
		ret[i] = r.(dataflow.Register)
	}
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
	log.Printf("device[%s]: shutdown completed", c.deviceConfig.Name())
}
