package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/httpDevice"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/mqttDevice"
	"github.com/koestler/go-iotdevice/victronDevice"
	"log"
)

func runDevices(
	cfg *config.Config,
	mqttClientPool *mqttClient.ClientPool,
	stateStorage *dataflow.ValueStorageInstance,
	commandStorage *dataflow.ValueStorageInstance,
) (devicePoolInstance *device.DevicePool) {
	devicePoolInstance = device.RunPool()

	for _, deviceConfig := range cfg.VictronDevices() {
		if cfg.LogWorkerStart() {
			log.Printf("device[%s]: start victron type", deviceConfig.Name())
		}

		if dev, err := victronDevice.RunDevice(deviceConfig, deviceConfig, stateStorage); err != nil {
			log.Printf("device[%s]: start failed: %s", deviceConfig.Name(), err)
		} else {
			device.RunMqttForwarders(dev, mqttClientPool, stateStorage)
			devicePoolInstance.AddDevice(dev)
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
			devicePoolInstance.AddDevice(dev)
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
			devicePoolInstance.AddDevice(dev)
		}
	}

	return
}
