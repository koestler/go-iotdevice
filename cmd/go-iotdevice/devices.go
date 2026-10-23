package main

import (
	"github.com/koestler/go-iotdevice/v3/config"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"github.com/koestler/go-iotdevice/v3/gensetDevice"
	"github.com/koestler/go-iotdevice/v3/gpioDevice"
	"github.com/koestler/go-iotdevice/v3/httpDevice"
	"github.com/koestler/go-iotdevice/v3/modbus"
	"github.com/koestler/go-iotdevice/v3/modbusDevice"
	"github.com/koestler/go-iotdevice/v3/mqttClient"
	"github.com/koestler/go-iotdevice/v3/mqttDevice"
	"github.com/koestler/go-iotdevice/v3/pool"
	"github.com/koestler/go-iotdevice/v3/restarter"
	"github.com/koestler/go-iotdevice/v3/victronDevice"
	"log"
	"time"
)

// gensetRunDelay is the delay before starting genset devices.
// This is used so that input/output devices have a change to start up and fetch their register list before
// the genset device sets up its bindings.
const gensetRunDelay = 2 * time.Second

func runDevicePool() *pool.Pool[*restarter.Restarter[device.Device]] {
	return pool.RunPool[*restarter.Restarter[device.Device]]()
}

func runNonMqttGensetDevices(
	cfg *config.Config,
	devicePool *pool.Pool[*restarter.Restarter[device.Device]],
	modbusPool *pool.Pool[*modbus.ModbusStruct],
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) {
	for _, deviceConfig := range cfg.VictronDevices() {
		if cfg.LogWorkerStart() {
			log.Printf("device[%s]: start victron type", deviceConfig.Name())
		}

		deviceConfig := victronDeviceConfig{deviceConfig}
		dev := victronDevice.NewDevice(deviceConfig, deviceConfig, stateStorage)
		watchedDev := restarter.CreateRestarter[device.Device](deviceConfig, dev)
		watchedDev.Run()
		devicePool.Add(watchedDev)
	}

	for _, deviceConfig := range cfg.ModbusDevices() {
		if cfg.LogWorkerStart() {
			log.Printf("device[%s]: start modbus type", deviceConfig.Name())
		}

		deviceConfig := modbusDeviceConfig{deviceConfig}
		modbusInstance := modbusPool.GetByName(deviceConfig.Bus())
		if modbusInstance == nil {
			log.Printf("device[%s]: start failed: bus=%s unavailable", deviceConfig.Name(), deviceConfig.Bus())
			continue
		}

		dev := modbusDevice.NewDevice(deviceConfig, deviceConfig, modbusInstance, stateStorage, commandStorage)
		watchedDev := restarter.CreateRestarter[device.Device](deviceConfig, dev)
		watchedDev.Run()
		devicePool.Add(watchedDev)
	}

	for _, deviceConfig := range cfg.GpioDevices() {
		if cfg.LogWorkerStart() {
			log.Printf("device[%s]: start gpio type", deviceConfig.Name())
		}

		deviceConfig := gpioDeviceConfig{deviceConfig}

		dev, err := gpioDevice.NewDevice(deviceConfig, deviceConfig, stateStorage, commandStorage)
		if err != nil {
			log.Printf("device[%s]: start failed: %s", deviceConfig.Name(), err)
			continue
		}
		watchedDev := restarter.CreateRestarter[device.Device](deviceConfig, dev)
		watchedDev.Run()
		devicePool.Add(watchedDev)
	}

	for _, deviceConfig := range cfg.HttpDevices() {
		if cfg.LogWorkerStart() {
			log.Printf("device[%s]: start tearacom type", deviceConfig.Name())
		}

		deviceConfig := httpDeviceConfig{deviceConfig}
		dev := httpDevice.NewDevice(deviceConfig, deviceConfig, stateStorage, commandStorage)
		watchedDev := restarter.CreateRestarter[device.Device](deviceConfig, dev)
		watchedDev.Run()
		devicePool.Add(watchedDev)
	}
}

func runMqttDevices(
	cfg *config.Config,
	devicePool *pool.Pool[*restarter.Restarter[device.Device]],
	mqttClientPool *pool.Pool[mqttClient.Client],
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) {
	for _, deviceConfig := range cfg.MqttDevices() {
		if cfg.LogWorkerStart() {
			log.Printf("device[%s]: start mqtt type", deviceConfig.Name())
		}

		deviceConfig := mqttDeviceConfig{deviceConfig, cfg.MqttClients()}
		dev := mqttDevice.NewDevice(deviceConfig, deviceConfig, stateStorage, commandStorage, mqttClientPool)
		watchedDev := restarter.CreateRestarter[device.Device](deviceConfig, dev)
		watchedDev.Run()
		devicePool.Add(watchedDev)
	}
}

func runGensetDevices(
	cfg *config.Config,
	devicePool *pool.Pool[*restarter.Restarter[device.Device]],
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) {
	for _, deviceConfig := range cfg.GensetDevices() {
		if cfg.LogWorkerStart() {
			log.Printf("device[%s]: start genset type", deviceConfig.Name())
		}

		deviceConfig := gensetDeviceConfig{deviceConfig}
		dev := gensetDevice.NewDevice(
			deviceConfig,
			deviceConfig,
			stateStorage,
			commandStorage,
			func(deviceName string) *dataflow.RegisterDb {
				return devicePool.GetByName(deviceName).Service().RegisterDb()
			},
		)
		watchedDev := restarter.CreateRestarter[device.Device](deviceConfig, dev)
		go func() {
			time.Sleep(gensetRunDelay)
			watchedDev.Run()
		}()
		devicePool.Add(watchedDev)
	}
}

// the following structs / methods are used to cast config.FilterConfig into dataflow.RegisterFilterConf

type victronDeviceConfig struct {
	config.VictronDeviceConfig
}

func (c victronDeviceConfig) Filter() dataflow.RegisterFilterConf {
	return c.VictronDeviceConfig.Filter()
}

type modbusDeviceConfig struct {
	config.ModbusDeviceConfig
}

func (c modbusDeviceConfig) Filter() dataflow.RegisterFilterConf {
	return c.ModbusDeviceConfig.Filter()
}

type gpioDeviceConfig struct {
	config.GpioDeviceConfig
}

func (c gpioDeviceConfig) Filter() dataflow.RegisterFilterConf {
	return c.GpioDeviceConfig.Filter()
}

func (c gpioDeviceConfig) Inputs() []gpioDevice.Pin {
	inp := c.GpioDeviceConfig.Inputs()
	oup := make([]gpioDevice.Pin, len(inp))
	for i, b := range inp {
		oup[i] = gpioDevice.Pin(b)
	}
	return oup
}

func (c gpioDeviceConfig) Outputs() []gpioDevice.Pin {
	inp := c.GpioDeviceConfig.Outputs()
	oup := make([]gpioDevice.Pin, len(inp))
	for i, b := range inp {
		oup[i] = gpioDevice.Pin(b)
	}
	return oup
}

type httpDeviceConfig struct {
	config.HttpDeviceConfig
}

func (c httpDeviceConfig) Filter() dataflow.RegisterFilterConf {
	return c.HttpDeviceConfig.Filter()
}

type mqttDeviceConfig struct {
	config.MqttDeviceConfig
	mqttClients []config.MqttClientConfig
}

func (c mqttDeviceConfig) Filter() dataflow.RegisterFilterConf {
	return c.MqttDeviceConfig.Filter()
}

type gensetDeviceConfig struct {
	config.GensetDeviceConfig
}

func (c gensetDeviceConfig) Filter() dataflow.RegisterFilterConf {
	return c.GensetDeviceConfig.Filter()
}

func (c gensetDeviceConfig) InputBindings() []gensetDevice.Binding {
	inp := c.GensetDeviceConfig.InputBindings()
	oup := make([]gensetDevice.Binding, len(inp))
	for i, b := range inp {
		oup[i] = gensetDevice.Binding(b)
	}
	return oup
}

func (c gensetDeviceConfig) OutputBindings() []gensetDevice.Binding {
	inp := c.GensetDeviceConfig.OutputBindings()
	oup := make([]gensetDevice.Binding, len(inp))
	for i, b := range inp {
		oup[i] = gensetDevice.Binding(b)
	}
	return oup
}

func (c mqttDeviceConfig) MqttClientTopics() map[string][]string {
	ret := make(map[string][]string)

	for _, mc := range c.mqttClients {
		for _, d := range mc.MqttDevices() {
			if d.Name() != c.Name() {
				continue
			}

			if _, ok := ret[mc.Name()]; !ok {
				ret[mc.Name()] = make([]string, 0)
			}

			ret[mc.Name()] = append(ret[mc.Name()], d.MqttTopics()...)
		}
	}

	return ret
}
