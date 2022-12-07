package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/mqttClient"
	"log"
)

func runMqttClient(
	cfg *config.Config,
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
		client = mqttClient.CreateV5(mqttClientConfig)
		mqttClientPoolInstance.AddClient(client)
	}

	return
}
