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

	Command() MqttSectionConfig
	CommandTopic(deviceName, registerName string) string

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
	Filter() dataflow.RegisterFilterConf
}

func RunMqttForwarders(
	cfg Config,
	mc mqttClient.Client,
	devicePool *pool.Pool[*restarter.Restarter[device.Device]],
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
) {
	if sCfg := cfg.HomeassistantDiscovery(); sCfg.Enabled() {
		for _, deviceConfig := range cfg.HomeassistantDiscovery().Devices() {
			if dev := devicePool.GetByName(deviceConfig.Name()); dev != nil {
				runHomeassistantDiscoveryForwarder(mc.GetCtx(), cfg, dev.Service(), mc, deviceConfig.Filter())
			}
		}
	}

	// delay first avail / struct / realtime message to give hass time to first process the discovery message and then
	// get the initial state from the realtime message
	time.Sleep(time.Second)

	if sCfg := cfg.AvailabilityDevice(); sCfg.Enabled() {
		for _, deviceConfig := range sCfg.Devices() {
			if dev := devicePool.GetByName(deviceConfig.Name()); dev != nil {
				runAvailabilityForwarder(mc.GetCtx(), cfg, dev.Service(), mc)
			}
		}
	}

	if sCfg := cfg.Structure(); sCfg.Enabled() {
		for _, deviceConfig := range sCfg.Devices() {
			if dev := devicePool.GetByName(deviceConfig.Name()); dev != nil {
				runStructureForwarder(mc.GetCtx(), cfg, dev.Service(), mc, deviceConfig.Filter())
			}
		}
	}

	if sCfg := cfg.Telemetry(); sCfg.Enabled() {
		for _, deviceConfig := range sCfg.Devices() {
			if dev := devicePool.GetByName(deviceConfig.Name()); dev != nil {
				runTelemetryForwarder(mc.GetCtx(), cfg, dev.Service(), mc, stateStorage, deviceConfig.Filter())
			}
		}
	}

	if sCfg := cfg.Realtime(); sCfg.Enabled() {
		for _, deviceConfig := range sCfg.Devices() {
			if dev := devicePool.GetByName(deviceConfig.Name()); dev != nil {
				runRealtimeForwarder(mc.GetCtx(), cfg, dev.Service(), mc, stateStorage, deviceConfig.Filter())
			}
		}
	}

	if sCfg := cfg.Command(); sCfg.Enabled() {
		for _, deviceConfig := range cfg.Command().Devices() {
			if dev := devicePool.GetByName(deviceConfig.Name()); dev != nil {
				runCommandForwarder(mc.GetCtx(), cfg, dev.Service(), mc, commandStorage, deviceConfig.Filter())
			}
		}
	}
}
