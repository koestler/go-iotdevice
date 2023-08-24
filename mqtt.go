package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"log"
)

func runMqttClient(
	cfg *config.Config,
) (mqttClientPoolInstance *pool.Pool[mqttClient.Client]) {
	// run pool
	mqttClientPoolInstance = pool.RunPool[mqttClient.Client]()

	for _, mqttClientConfig := range cfg.MqttClients() {
		if cfg.LogWorkerStart() {
			log.Printf(
				"mqttClient[%s]: start: Broker='%s', ClientId='%s'",
				mqttClientConfig.Name(), mqttClientConfig.Broker(), mqttClientConfig.ClientId(),
			)
		}

		client := mqttClient.NewV5(mqttClientConfig)
		client.Run()
		mqttClientPoolInstance.Add(client)
	}

	return
}
