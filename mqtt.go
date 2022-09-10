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
) (clientPoolInstance *mqttClient.ClientPool) {
	clientPoolInstance = mqttClient.RunPool()

	for _, cfgClient := range cfg.MqttClients() {
		if cfg.LogWorkerStart() {
			log.Printf(
				"mqttClient[%s]: broker='%s'",
				cfgClient.Name(),
				cfgClient.Broker(),
			)
		}

		if client, err := mqttClient.RunClient(cfgClient, devicePoolInstance, storage); err != nil {
			log.Printf("mqttClient[%s]: start failed: %s", cfgClient.Name(), err)
		} else {
			clientPoolInstance.AddClient(client)
			if cfg.LogWorkerStart() {
				log.Printf(
					"mqttClient[%s]: started",
					client.Config().Name(),
				)
			}
		}
	}

	return
}
