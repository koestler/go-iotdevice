package main

import (
	"github.com/koestler/go-iotdevice/v3/config"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"github.com/koestler/go-iotdevice/v3/mqttClient"
	"github.com/koestler/go-iotdevice/v3/mqttForwarders"
	"github.com/koestler/go-iotdevice/v3/pool"
	"github.com/koestler/go-iotdevice/v3/restarter"
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
