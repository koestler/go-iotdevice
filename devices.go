package main

import (
	"github.com/koestler/go-iotdevice/v3/config"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"github.com/koestler/go-iotdevice/v3/httpDevice"
	"github.com/koestler/go-iotdevice/v3/modbus"
	"github.com/koestler/go-iotdevice/v3/modbusDevice"
	"github.com/koestler/go-iotdevice/v3/mqttClient"
	"github.com/koestler/go-iotdevice/v3/mqttDevice"
	"github.com/koestler/go-iotdevice/v3/pool"
	"github.com/koestler/go-iotdevice/v3/restarter"
	"github.com/koestler/go-iotdevice/v3/victronDevice"
	"log"
)

func runNonMqttDevices(
	cfg *config.Config,
	modbusPool *pool.Pool[*modbus.ModbusStruct],
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) (devicePool *pool.Pool[*restarter.Restarter[device.Device]]) {
	devicePool = pool.RunPool[*restarter.Restarter[device.Device]]()

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

	return
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
