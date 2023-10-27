package mqttForwarders

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"github.com/koestler/go-iotdevice/restarter"
)

func RunMqttForwarders(
	mqttClientConfig config.MqttClientConfig,
	mc mqttClient.Client,
	devicePool *pool.Pool[*restarter.Restarter[device.Device]],
	storage *dataflow.ValueStorage,
) {
	for _, deviceConfig := range mqttClientConfig.AvailabilityDevice().Devices() {
		dev := devicePool.GetByName(deviceConfig.Name())
		runAvailabilityForwarder(mc.GetCtx(), dev.Service(), mc)
	}

	for _, deviceConfig := range mqttClientConfig.Structure().Devices() {
		dev := devicePool.GetByName(deviceConfig.Name())
		runStructureForwarder(mc.GetCtx(), dev.Service(), mc, deviceConfig.RegisterFilter())
	}

	for _, deviceConfig := range mqttClientConfig.Telemetry().Devices() {
		dev := devicePool.GetByName(deviceConfig.Name())
		runTelemetryForwarder(mc.GetCtx(), dev.Service(), mc, storage, deviceConfig.RegisterFilter())
	}

	for _, deviceConfig := range mqttClientConfig.Realtime().Devices() {
		dev := devicePool.GetByName(deviceConfig.Name())
		runRealtimeForwarder(mc.GetCtx(), dev.Service(), mc, storage, deviceConfig.RegisterFilter())
	}

	for _, deviceConfig := range mqttClientConfig.HomeassistantDiscovery().Devices() {
		dev := devicePool.GetByName(deviceConfig.Name())
		runHomeassistantDiscoveryForwarder(mc.GetCtx(), dev.Service(), mc, deviceConfig.RegisterFilter())
	}
}