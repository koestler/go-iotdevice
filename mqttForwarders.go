package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/mqttForwarders"
	"github.com/koestler/go-iotdevice/pool"
	"github.com/koestler/go-iotdevice/restarter"
)

func runMqttForwarders(
	cfg *config.Config,
	devicePool *pool.Pool[*restarter.Restarter[device.Device]],
	mqttClientPool *pool.Pool[mqttClient.Client],
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) {
	for _, c := range cfg.MqttClients() {
		forwarderCfg := forwarderConfig{c}
		client := mqttClientPool.GetByName(c.Name())
		go mqttForwarders.RunMqttForwarders(forwarderCfg, client, devicePool, stateStorage, commandStorage)
	}
}
