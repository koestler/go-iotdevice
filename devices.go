package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/victron"
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
			log.Printf("device[%s]: start", deviceConfig.Name())
		}

		if dev, err := victron.RunDevice(deviceConfig, mqttClientPool, storage); err != nil {
			log.Printf("device[%s]: start failed: %s", deviceConfig.Name(), err)
		} else {
			devicePoolInstance.AddDevice(dev)
			countStarted += 1
		}
	}

	return
}
