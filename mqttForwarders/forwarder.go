package mqttForwarders

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"github.com/koestler/go-iotdevice/restarter"
	"time"
)

type Config interface {
	ClientId() string

	AvailabilityClient() MqttSectionConfig
	AvailabilityClientTopic() string

	AvailabilityDevice() MqttSectionConfig
	AvailabilityDeviceTopic(deviceName string) string

	Structure() MqttSectionConfig
	StructureTopic(deviceName string) string

	Telemetry() MqttSectionConfig
	TelemetryTopic(deviceName string) string

	Realtime() MqttSectionConfig
	RealtimeTopic(deviceName, registerName string) string

	HomeassistantDiscovery() MqttSectionConfig
	HomeassistantDiscoveryTopic(component, nodeId, objectId string) string

	LogDebug() bool
}

type MqttSectionConfig interface {
	Enabled() bool
	Interval() time.Duration
	Retain() bool
	Qos() byte
	Devices() []MqttDeviceSectionConfig
}

type MqttDeviceSectionConfig interface {
	Name() string
	RegisterFilter() dataflow.RegisterFilterConf
}

func RunMqttForwarders(
	cfg Config,
	mc mqttClient.Client,
	devicePool *pool.Pool[*restarter.Restarter[device.Device]],
	storage *dataflow.ValueStorage,
) {
	if sCfg := cfg.AvailabilityDevice(); sCfg.Enabled() {
		for _, deviceConfig := range sCfg.Devices() {
			dev := devicePool.GetByName(deviceConfig.Name())
			runAvailabilityForwarder(mc.GetCtx(), cfg, dev.Service(), mc)
		}
	}

	if sCfg := cfg.Structure(); sCfg.Enabled() {
		for _, deviceConfig := range sCfg.Devices() {
			dev := devicePool.GetByName(deviceConfig.Name())
			runStructureForwarder(mc.GetCtx(), cfg, dev.Service(), mc, deviceConfig.RegisterFilter())
		}
	}

	if sCfg := cfg.Telemetry(); sCfg.Enabled() {
		for _, deviceConfig := range sCfg.Devices() {
			dev := devicePool.GetByName(deviceConfig.Name())
			runTelemetryForwarder(mc.GetCtx(), cfg, dev.Service(), mc, storage, deviceConfig.RegisterFilter())
		}
	}

	if sCfg := cfg.Realtime(); sCfg.Enabled() {
		for _, deviceConfig := range sCfg.Devices() {
			dev := devicePool.GetByName(deviceConfig.Name())
			runRealtimeForwarder(mc.GetCtx(), cfg, dev.Service(), mc, storage, deviceConfig.RegisterFilter())
		}
	}

	if sCfg := cfg.HomeassistantDiscovery(); sCfg.Enabled() {
		for _, deviceConfig := range cfg.HomeassistantDiscovery().Devices() {
			dev := devicePool.GetByName(deviceConfig.Name())
			runHomeassistantDiscoveryForwarder(mc.GetCtx(), cfg, dev.Service(), mc, deviceConfig.RegisterFilter())
		}
	}
}
