package config

import (
	"fmt"
	"golang.org/x/exp/maps"
)

func (c Config) MarshalYAML() (interface{}, error) {
	return configRead{
		Version:         &c.version,
		ProjectTitle:    c.projectTitle,
		LogConfig:       &c.logConfig,
		LogWorkerStart:  &c.logWorkerStart,
		LogStorageDebug: &c.logStorageDebug,
		HttpServer:      ConvertEnableableToRead[HttpServerConfig, httpServerConfigRead](c.httpServer),
		Authentication:  ConvertEnableableToRead[AuthenticationConfig, authenticationConfigRead](c.authentication),
		MqttClients:     ConvertMapToRead[MqttClientConfig, mqttClientConfigRead](c.mqttClients),
		Modbus:          ConvertMapToRead[ModbusConfig, modbusConfigRead](c.modbus),
		VictronDevices:  ConvertMapToRead[VictronDeviceConfig, victronDeviceConfigRead](c.victronDevices),
		ModbusDevices:   ConvertMapToRead[ModbusDeviceConfig, modbusDeviceConfigRead](c.modbusDevices),
		HttpDevices:     ConvertMapToRead[HttpDeviceConfig, httpDeviceConfigRead](c.httpDevices),
		MqttDevices:     ConvertMapToRead[MqttDeviceConfig, mqttDeviceConfigRead](c.mqttDevices),
		Views:           ConvertListToRead[ViewConfig, viewConfigRead](c.views),
	}, nil
}

type convertable[O any] interface {
	ConvertToRead() O
}

type enableable[O any] interface {
	Enabled() bool
	convertable[O]
}

func ConvertEnableableToRead[I enableable[O], O any](inp I) *O {
	if !inp.Enabled() {
		return nil
	}
	r := inp.ConvertToRead()
	return &r
}

type mappable[O any] interface {
	Nameable
	convertable[O]
}

func ConvertMapToRead[I mappable[O], O any](inp []*I) (oup map[string]O) {
	oup = make(map[string]O, len(inp))
	for _, c := range inp {
		oup[(*c).Name()] = (*c).ConvertToRead()
	}
	return
}

func ConvertListToRead[I convertable[O], O any](inp []*I) (oup []O) {
	oup = make([]O, len(inp))
	i := 0
	for _, c := range inp {
		oup[i] = (*c).ConvertToRead()
		i++
	}
	return
}

func (c HttpServerConfig) ConvertToRead() httpServerConfigRead {
	frontendProxy := ""
	if c.frontendProxy != nil {
		frontendProxy = c.frontendProxy.String()
	}

	return httpServerConfigRead{
		Bind:            c.bind,
		Port:            &c.port,
		LogRequests:     &c.logRequests,
		FrontendProxy:   frontendProxy,
		FrontendPath:    c.frontendPath,
		FrontendExpires: c.frontendExpires.String(),
		ConfigExpires:   c.configExpires.String(),
		LogDebug:        &c.logDebug,
	}
}

func (c AuthenticationConfig) ConvertToRead() authenticationConfigRead {
	jwtSecret := string(c.jwtSecret)
	return authenticationConfigRead{
		JwtSecret:         &jwtSecret,
		JwtValidityPeriod: c.jwtValidityPeriod.String(),
		HtaccessFile:      &c.htaccessFile,
	}
}

func (c MqttClientConfig) ConvertToRead() mqttClientConfigRead {
	return mqttClientConfigRead{
		Broker:            c.broker.String(),
		ProtocolVersion:   &c.protocolVersion,
		User:              c.user,
		Password:          c.password,
		ClientId:          &c.clientId,
		Qos:               &c.qos,
		KeepAlive:         c.keepAlive.String(),
		ConnectRetryDelay: c.connectRetryDelay.String(),
		ConnectTimeout:    c.connectTimeout.String(),
		AvailabilityTopic: &c.availabilityTopic,
		TelemetryInterval: c.telemetryInterval.String(),
		TelemetryTopic:    &c.telemetryTopic,
		TelemetryRetain:   &c.telemetryRetain,
		RealtimeEnable:    &c.realtimeEnable,
		RealtimeTopic:     &c.realtimeTopic,
		RealtimeRetain:    &c.realtimeRetain,
		TopicPrefix:       c.topicPrefix,
		LogDebug:          &c.logDebug,
		LogMessages:       &c.logMessages,
	}
}

func (c ModbusConfig) ConvertToRead() modbusConfigRead {
	return modbusConfigRead{
		Device:      c.device,
		BaudRate:    c.baudRate,
		ReadTimeout: c.readTimeout.String(),
	}
}

func (c DeviceConfig) ConvertToRead() deviceConfigRead {
	return deviceConfigRead{
		SkipFields:              c.skipFields,
		SkipCategories:          c.skipCategories,
		TelemetryViaMqttClients: c.telemetryViaMqttClients,
		RealtimeViaMqttClients:  c.realtimeViaMqttClients,
		LogDebug:                &c.logDebug,
		LogComDebug:             &c.logComDebug,
	}
}

func (c VictronDeviceConfig) ConvertToRead() victronDeviceConfigRead {
	return victronDeviceConfigRead{
		General: c.DeviceConfig.ConvertToRead(),
		Device:  c.device,
		Kind:    c.kind.String(),
	}
}

func (c ModbusDeviceConfig) ConvertToRead() modbusDeviceConfigRead {
	return modbusDeviceConfigRead{
		General:      c.DeviceConfig.ConvertToRead(),
		Bus:          c.bus,
		Kind:         c.kind.String(),
		Address:      fmt.Sprintf("0x%02x", c.address),
		Relays:       c.relays,
		PollInterval: c.pollInterval.String(),
	}
}

func (c HttpDeviceConfig) ConvertToRead() httpDeviceConfigRead {
	return httpDeviceConfigRead{
		General:                c.DeviceConfig.ConvertToRead(),
		Url:                    c.url.String(),
		Kind:                   c.kind.String(),
		Username:               c.username,
		Password:               c.password,
		PollInterval:           c.pollInterval.String(),
		PollIntervalMaxBackoff: c.pollIntervalMaxBackoff.String(),
	}
}

func (c MqttDeviceConfig) ConvertToRead() mqttDeviceConfigRead {
	return mqttDeviceConfigRead{
		General:     c.DeviceConfig.ConvertToRead(),
		MqttTopics:  c.mqttTopics,
		MqttClients: c.mqttClients,
	}
}

func (c ViewDeviceConfig) ConvertToRead() viewDeviceConfigRead {
	return viewDeviceConfigRead{
		Name:  c.name,
		Title: c.title,
	}
}

func (c ViewConfig) ConvertToRead() viewConfigRead {
	return viewConfigRead{
		Name:           c.name,
		Title:          c.title,
		Devices:        ConvertListToRead[ViewDeviceConfig, viewDeviceConfigRead](c.devices),
		Autoplay:       &c.autoplay,
		AllowedUsers:   maps.Keys(c.allowedUsers),
		Hidden:         &c.hidden,
		SkipFields:     c.skipFields,
		SkipCategories: c.skipCategories,
	}
}
