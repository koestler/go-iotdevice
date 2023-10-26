package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/mqttForwarders"
	"github.com/koestler/go-iotdevice/pool"
	"github.com/koestler/go-iotdevice/restarter"
	"log"
)

func runMqttClient(
	cfg *config.Config,
	devicePool *pool.Pool[*restarter.Restarter[device.Device]],
	storage *dataflow.ValueStorage,
) (mqttClientPool *pool.Pool[mqttClient.Client]) {
	// run pool
	mqttClientPool = pool.RunPool[mqttClient.Client]()

	for _, mqttClientConfig := range cfg.MqttClients() {
		if cfg.LogWorkerStart() {
			log.Printf(
				"mqttClient[%s]: start: Broker='%s', ClientId='%s'",
				mqttClientConfig.Name(), mqttClientConfig.Broker(), mqttClientConfig.ClientId(),
			)
		}

		client := mqttClient.NewV5(mqttClientConfig)
		client.Run()
		mqttClientPool.Add(client)

		mqttForwarders.RunMqttForwarders(mqttClientConfig, client, devicePool, storage)
	}

	return
}
