package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/httpDevice"
	"github.com/koestler/go-iotdevice/modbus"
	"github.com/koestler/go-iotdevice/modbusDevice"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/mqttDevice"
	"github.com/koestler/go-iotdevice/pool"
	"github.com/koestler/go-iotdevice/restarter"
	"github.com/koestler/go-iotdevice/victronDevice"
	"log"
)

func runDevices(
	cfg *config.Config,
	mqttClientPool *pool.Pool[mqttClient.Client],
	modbusPoolInstance *pool.Pool[*modbus.ModbusStruct],
	stateStorage *dataflow.ValueStorageInstance,
	commandStorage *dataflow.ValueStorageInstance,
) (devicePoolInstance *pool.Pool[*restarter.Restarter[device.Device]]) {
	devicePoolInstance = pool.RunPool[*restarter.Restarter[device.Device]]()

	for _, deviceConfig := range cfg.VictronDevices() {
		if cfg.LogWorkerStart() {
			log.Printf("device[%s]: start victron type", deviceConfig.Name())
		}

		dev := victronDevice.NewDevice(deviceConfig, deviceConfig, stateStorage)
		watchedDev := restarter.RunRestarter[device.Device](deviceConfig, dev)
		device.RunMqttForwarders(watchedDev.GetCtx(), dev, mqttClientPool, stateStorage)
		devicePoolInstance.Add(watchedDev)
	}

	for _, deviceConfig := range cfg.ModbusDevices() {
		if cfg.LogWorkerStart() {
			log.Printf("device[%s]: start modbus type", deviceConfig.Name())
		}

		modbusInstance := modbusPoolInstance.GetByName(deviceConfig.Bus())
		if modbusInstance == nil {
			log.Printf("device[%s]: start failed: bus=%s unavailable", deviceConfig.Name(), deviceConfig.Bus())
			continue
		}

		dev := modbusDevice.NewDevice(deviceConfig, deviceConfig, modbusInstance, stateStorage, commandStorage)
		watchedDev := restarter.RunRestarter[device.Device](deviceConfig, dev)
		device.RunMqttForwarders(watchedDev.GetCtx(), dev, mqttClientPool, stateStorage)
		devicePoolInstance.Add(watchedDev)
	}

	for _, deviceConfig := range cfg.MqttDevices() {
		if cfg.LogWorkerStart() {
			log.Printf("device[%s]: start mqtt type", deviceConfig.Name())
		}

		dev := mqttDevice.NewDevice(deviceConfig, deviceConfig, stateStorage, mqttClientPool)
		watchedDev := restarter.RunRestarter[device.Device](deviceConfig, dev)
		device.RunMqttForwarders(watchedDev.GetCtx(), dev, mqttClientPool, stateStorage)
		devicePoolInstance.Add(watchedDev)
	}

	for _, deviceConfig := range cfg.HttpDevices() {
		if cfg.LogWorkerStart() {
			log.Printf("device[%s]: start tearacom type", deviceConfig.Name())
		}

		dev := httpDevice.NewDevice(deviceConfig, deviceConfig, stateStorage, commandStorage)
		watchedDev := restarter.RunRestarter[device.Device](deviceConfig, dev)
		device.RunMqttForwarders(watchedDev.GetCtx(), dev, mqttClientPool, stateStorage)
		devicePoolInstance.Add(watchedDev)
	}

	return
}
