package victronDevice

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
	Kind() config.VictronDeviceKind
}

type DeviceStruct struct {
	deviceConfig  device.Config
	victronConfig Config

	registers        VictronRegisters
	lastUpdated      time.Time
	lastUpdatedMutex sync.RWMutex
	model            string

	shutdown chan struct{}
	closed   chan struct{}
}

func RunDevice(
	deviceConfig device.Config,
	victronConfig Config,
	storage *dataflow.ValueStorageInstance,
) (device device.Device, err error) {
	c := &DeviceStruct{
		deviceConfig:  deviceConfig,
		victronConfig: victronConfig,
		shutdown:      make(chan struct{}),
		closed:        make(chan struct{}),
	}

	if victronConfig.Kind() == config.VictronVedirectKind {
		err = startVedirect(c, storage)
	} else if victronConfig.Kind() == config.VictronRandomBmvKind {
		err = startRandom(c, storage, RegisterListBmv712)
	} else if victronConfig.Kind() == config.VictronRandomSolarKind {
		err = startRandom(c, storage, RegisterListSolar)
	} else {
		return nil, fmt.Errorf("unknown device kind: %s", victronConfig.Kind().String())
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
