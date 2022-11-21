package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"log"
)

func runMqttClient(
	cfg *config.Config,
	devicePoolInstance *device.DevicePool,
	storage *dataflow.ValueStorageInstance,
) (mqttClientPoolInstance *mqttClient.ClientPool) {
	// run pool
	mqttClientPoolInstance = mqttClient.RunPool()

	for _, mqttClientConfig := range cfg.MqttClients() {
		if cfg.LogWorkerStart() {
			log.Printf(
				"mqttClient[%s]: start: Broker='%s', ClientId='%s'",
				mqttClientConfig.Name(), mqttClientConfig.Broker(), mqttClientConfig.ClientId(),
			)
		}

		var client mqttClient.Client
		client = mqttClient.CreateV5(mqttClientConfig, devicePoolInstance, storage)

		mqttClientPoolInstance.AddClient(client)
	}

	return
}
