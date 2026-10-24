package run

import (
	"log"

	"github.com/koestler/go-iotdevice/v3/internal/config"
	"github.com/koestler/go-iotdevice/v3/internal/dataflow"
	"github.com/koestler/go-iotdevice/v3/internal/device"
	"github.com/koestler/go-iotdevice/v3/internal/mqttClient"
	"github.com/koestler/go-iotdevice/v3/internal/mqttForwarders"
	"github.com/koestler/go-iotdevice/v3/internal/pool"
	"github.com/koestler/go-iotdevice/v3/internal/restarter"
)

func MqttForwarders(
	cfg *config.Config,
	devicePool *pool.Pool[*restarter.Restarter[device.Device]],
	mqttClientPool *pool.Pool[mqttClient.Client],
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) {
	for _, c := range cfg.MqttClients() {
		forwarderCfg := forwarderConfig{c}
		client, ok := mqttClientPool.GetByName(c.Name())
		if !ok {
			log.Printf("mqttForwarders: mqttClient=%s not found", c.Name())
			continue
		}
		go mqttForwarders.RunMqttForwarders(forwarderCfg, client, devicePool, stateStorage, commandStorage)
	}
}
