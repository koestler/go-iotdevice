package modbusDevice

import (
	"fmt"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/modbus"
	"log"
	"sync"
	"time"
)

type Config interface {
	Bus() string
	Kind() config.ModbusDeviceKind
	Address() byte
	PollInterval() time.Duration
}

type DeviceStruct struct {
	deviceConfig   device.Config
	modbusConfig   Config
	stateStorage   *dataflow.ValueStorageInstance
	commandStorage *dataflow.ValueStorageInstance

	modbus modbus.Modbus

	registers        ModbusRegisters
	lastUpdated      time.Time
	lastUpdatedMutex sync.RWMutex

	shutdown chan struct{}
	closed   chan struct{}
}

func RunDevice(
	deviceConfig device.Config,
	modbusConfig Config,
	modbus modbus.Modbus,
	stateStorage *dataflow.ValueStorageInstance,
	commandStorage *dataflow.ValueStorageInstance,
) (device device.Device, err error) {
	c := &DeviceStruct{
		deviceConfig:   deviceConfig,
		modbusConfig:   modbusConfig,
		stateStorage:   stateStorage,
		commandStorage: commandStorage,
		modbus:         modbus,
		shutdown:       make(chan struct{}),
		closed:         make(chan struct{}),
	}

	if modbusConfig.Kind() == config.ModbusWaveshareRtuRelay8Kind {
		err = startWaveshareRtuRelay8(c)
	} else {
		return nil, fmt.Errorf("unknown device kind: %s", modbusConfig.Kind().String())
	}

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *DeviceStruct) Name() string {
	return c.deviceConfig.Name()
}

func (c *DeviceStruct) Config() device.Config {
	return c.deviceConfig
}

func (c *DeviceStruct) ShutdownChan() chan struct{} {
	return c.shutdown
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
	return c.modbusConfig.Kind().String()
}

func (c *DeviceStruct) Shutdown() {
	close(c.shutdown)
	<-c.closed
	log.Printf("device[%s]: shutdown completed", c.deviceConfig.Name())
}
