package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/modbus"
	"github.com/koestler/go-iotdevice/mqttClient"
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

		dev := victronDevice.CreateDevice(deviceConfig, deviceConfig, stateStorage)
		watchedDev := restarter.RunRestarter[device.Device](dev)
		device.RunMqttForwarders(dev, mqttClientPool, stateStorage)
		devicePoolInstance.Add(watchedDev)
	}

	/*
		for _, deviceConfig := range cfg.ModbusDevices() {
			if cfg.LogWorkerStart() {
				log.Printf("device[%s]: start modbus type", deviceConfig.Name())
			}

			modbus := modbusPoolInstance.GetByName(deviceConfig.Bus())
			if modbus == nil {
				log.Printf("device[%s]: start failed: bus=%s unavailable", deviceConfig.Name(), deviceConfig.Bus())
				continue
			}
			if dev, err := modbusDevice.RunDevice(deviceConfig, deviceConfig, modbus, stateStorage, commandStorage); err != nil {
				log.Printf("device[%s]: start failed: %s", deviceConfig.Name(), err)
			} else {
				device.RunMqttForwarders(dev, mqttClientPool, stateStorage)
				devicePoolInstance.Add(dev)
			}
		}

		for _, deviceConfig := range cfg.MqttDevices() {
			if cfg.LogWorkerStart() {
				log.Printf("device[%s]: start mqtt type", deviceConfig.Name())
			}

			if dev, err := mqttDevice.RunDevice(deviceConfig, deviceConfig, stateStorage, mqttClientPool); err != nil {
				log.Printf("device[%s]: start failed: %s", deviceConfig.Name(), err)
			} else {
				device.RunMqttForwarders(dev, mqttClientPool, stateStorage)
				devicePoolInstance.Add(dev)
			}
		}

		for _, deviceConfig := range cfg.HttpDevices() {
			if cfg.LogWorkerStart() {
				log.Printf("device[%s]: start tearacom type", deviceConfig.Name())
			}

			if dev, err := httpDevice.RunDevice(deviceConfig, deviceConfig, stateStorage, commandStorage); err != nil {
				log.Printf("device[%s]: start failed: %s", deviceConfig.Name(), err)
			} else {
				device.RunMqttForwarders(dev, mqttClientPool, stateStorage)
				devicePoolInstance.Add(dev)
			}
		}
	*/

	return
}
