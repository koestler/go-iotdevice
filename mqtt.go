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
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) (mqttClientPool *pool.Pool[mqttClient.Client]) {
	// run pool
	mqttClientPool = pool.RunPool[mqttClient.Client]()

	for _, c := range cfg.MqttClients() {
		if cfg.LogWorkerStart() {
			log.Printf(
				"mqttClient[%s]: start: Broker='%s', ClientId='%s'",
				c.Name(), c.Broker(), c.ClientId(),
			)
		}

		mcCfg := mqttClientConfig{c}
		client := mqttClient.NewV5(mcCfg)
		client.Run()
		mqttClientPool.Add(client)

		forwarderCfg := forwarderConfig{c}
		mqttForwarders.RunMqttForwarders(forwarderCfg, client, devicePool, stateStorage, commandStorage)
	}

	return
}

type mqttClientConfig struct {
	config.MqttClientConfig
}

func (c mqttClientConfig) AvailabilityClient() mqttClient.MqttSectionConfig {
	return c.MqttClientConfig.AvailabilityClient()
}

type forwarderConfig struct {
	config.MqttClientConfig
}

func (c forwarderConfig) AvailabilityClient() mqttForwarders.MqttSectionConfig {
	return forwarderMqttSectionConfig{c.MqttClientConfig.AvailabilityClient()}
}

func (c forwarderConfig) AvailabilityDevice() mqttForwarders.MqttSectionConfig {
	return forwarderMqttSectionConfig{c.MqttClientConfig.AvailabilityDevice()}
}

func (c forwarderConfig) Structure() mqttForwarders.MqttSectionConfig {
	return forwarderMqttSectionConfig{c.MqttClientConfig.Structure()}
}

func (c forwarderConfig) Telemetry() mqttForwarders.MqttSectionConfig {
	return forwarderMqttSectionConfig{c.MqttClientConfig.Telemetry()}
}

func (c forwarderConfig) Realtime() mqttForwarders.MqttSectionConfig {
	return forwarderMqttSectionConfig{c.MqttClientConfig.Realtime()}
}

func (c forwarderConfig) HomeassistantDiscovery() mqttForwarders.MqttSectionConfig {
	return forwarderMqttSectionConfig{c.MqttClientConfig.HomeassistantDiscovery()}
}

func (c forwarderConfig) Command() mqttForwarders.MqttSectionConfig {
	return forwarderMqttSectionConfig{c.MqttClientConfig.Command()}
}

type forwarderMqttSectionConfig struct {
	config.MqttSectionConfig
}

func (c forwarderMqttSectionConfig) Devices() []mqttForwarders.MqttDeviceSectionConfig {
	devices := c.MqttSectionConfig.Devices()
	ret := make([]mqttForwarders.MqttDeviceSectionConfig, len(devices))
	for i, d := range devices {
		ret[i] = forwarderMqttDeviceSectionConfig{d}
	}
	return ret
}

type forwarderMqttDeviceSectionConfig struct {
	config.MqttDeviceSectionConfig
}

func (c forwarderMqttDeviceSectionConfig) Filter() dataflow.RegisterFilterConf {
	return c.MqttDeviceSectionConfig.Filter()
}
