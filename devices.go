package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/mqttDevice"
	"github.com/koestler/go-iotdevice/victronDevice"
	"log"
)

func runDevices(
	cfg *config.Config,
	mqttClientPool *mqttClient.ClientPool,
	storage *dataflow.ValueStorageInstance,
) (devicePoolInstance *device.DevicePool) {
	devicePoolInstance = device.RunPool()

	countStarted := 0
	for _, deviceConfig := range cfg.VictronDevices() {
		if cfg.LogWorkerStart() {
			log.Printf("device[%s]: start victron type", deviceConfig.Name())
		}

		if dev, err := victronDevice.RunDevice(deviceConfig, deviceConfig, storage); err != nil {
			log.Printf("device[%s]: start failed: %s", deviceConfig.Name(), err)
		} else {
			device.RunMqttForwarders(dev, mqttClientPool, storage)
			devicePoolInstance.AddDevice(dev)
			countStarted += 1
		}
	}

	for _, deviceConfig := range cfg.MqttDevices() {
		if cfg.LogWorkerStart() {
			log.Printf("device[%s]: start mqtt type", deviceConfig.Name())
		}

		if dev, err := mqttDevice.RunDevice(deviceConfig, deviceConfig, storage, mqttClientPool); err != nil {
			log.Printf("device[%s]: start failed: %s", deviceConfig.Name(), err)
		} else {
			device.RunMqttForwarders(dev, mqttClientPool, storage)
			devicePoolInstance.AddDevice(dev)
			countStarted += 1
		}
	}

	return
}
