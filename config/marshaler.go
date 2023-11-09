package config

import (
	"fmt"
	"golang.org/x/exp/maps"
)

func (c Config) MarshalYAML() (interface{}, error) {
	return configRead{
		Version:                &c.version,
		ProjectTitle:           c.projectTitle,
		LogConfig:              &c.logConfig,
		LogWorkerStart:         &c.logWorkerStart,
		LogStateStorageDebug:   &c.logStateStorageDebug,
		LogCommandStorageDebug: &c.logCommandStorageDebug,
		HttpServer:             convertEnableableToRead[HttpServerConfig, httpServerConfigRead](c.httpServer),
		Authentication:         convertEnableableToRead[AuthenticationConfig, authenticationConfigRead](c.authentication),
		MqttClients:            convertMapToRead[MqttClientConfig, mqttClientConfigRead](c.mqttClients),
		Modbus:                 convertMapToRead[ModbusConfig, modbusConfigRead](c.modbus),
		VictronDevices:         convertMapToRead[VictronDeviceConfig, victronDeviceConfigRead](c.victronDevices),
		ModbusDevices:          convertMapToRead[ModbusDeviceConfig, modbusDeviceConfigRead](c.modbusDevices),
		HttpDevices:            convertMapToRead[HttpDeviceConfig, httpDeviceConfigRead](c.httpDevices),
		MqttDevices:            convertMapToRead[MqttDeviceConfig, mqttDeviceConfigRead](c.mqttDevices),
		Views:                  convertListToRead[ViewConfig, viewConfigRead](c.views),
	}, nil
}

type convertable[O any] interface {
	convertToRead() O
}

type enableable[O any] interface {
	Enabled() bool
	convertable[O]
}

func convertEnableableToRead[I enableable[O], O any](inp I) *O {
	if !inp.Enabled() {
		return nil
	}
	r := inp.convertToRead()
	return &r
}

type mappable[O any] interface {
	Nameable
	convertable[O]
}

func convertMapToRead[I mappable[O], O any](inp []I) (oup map[string]O) {
	oup = make(map[string]O, len(inp))
	for _, c := range inp {
		oup[c.Name()] = c.convertToRead()
	}
	return
}

func convertListToRead[I convertable[O], O any](inp []I) (oup []O) {
	oup = make([]O, len(inp))
	i := 0
	for _, c := range inp {
		oup[i] = c.convertToRead()
		i++
	}
	return
}

//lint:ignore U1000 linter does not catch that this is used generic code
func (c HttpServerConfig) convertToRead() httpServerConfigRead {
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

//lint:ignore U1000 linter does not catch that this is used generic code
func (c AuthenticationConfig) convertToRead() authenticationConfigRead {
	jwtSecret := string(c.jwtSecret)
	return authenticationConfigRead{
		JwtSecret:         &jwtSecret,
		JwtValidityPeriod: c.jwtValidityPeriod.String(),
		HtaccessFile:      &c.htaccessFile,
	}
}

//lint:ignore U1000 linter does not catch that this is used generic code
func (c MqttClientConfig) convertToRead() mqttClientConfigRead {
	return mqttClientConfigRead{
		Broker:          c.broker.String(),
		ProtocolVersion: &c.protocolVersion,

		User:     c.user,
		Password: c.password,
		ClientId: &c.clientId,

		KeepAlive:         c.keepAlive.String(),
		ConnectRetryDelay: c.connectRetryDelay.String(),
		ConnectTimeout:    c.connectTimeout.String(),
		TopicPrefix:       &c.topicPrefix,
		ReadOnly:          &c.readOnly,
		MaxBacklogSize:    &c.maxBacklogSize,

		AvailabilityClient:     c.availabilityClient.convertToRead(),
		AvailabilityDevice:     c.availabilityDevice.convertToRead(),
		Structure:              c.structure.convertToRead(),
		Telemetry:              c.telemetry.convertToRead(),
		Realtime:               c.realtime.convertToRead(),
		HomeassistantDiscovery: c.homeassistantDiscovery.convertToRead(),
		Command:                c.command.convertToRead(),

		LogDebug:    &c.logDebug,
		LogMessages: &c.logMessages,
	}
}

//lint:ignore U1000 linter does not catch that this is used generic code
func (c MqttSectionConfig) convertToRead() mqttSectionConfigRead {
	return mqttSectionConfigRead{
		Enabled:       &c.enabled,
		TopicTemplate: &c.topicTemplate,
		Interval:      c.interval.String(),
		Retain:        &c.retain,
		Qos:           &c.qos,
		Devices:       convertMapToRead[MqttDeviceSectionConfig, mqttDeviceSectionConfigRead](c.devices),
	}
}

//lint:ignore U1000 linter does not catch that this is used generic code
func (c MqttDeviceSectionConfig) convertToRead() mqttDeviceSectionConfigRead {
	rf := c.filter.convertToRead()
	return mqttDeviceSectionConfigRead{
		Filter: &rf,
	}
}

//lint:ignore U1000 linter does not catch that this is used generic code
func (c ModbusConfig) convertToRead() modbusConfigRead {
	return modbusConfigRead{
		Device:      c.device,
		BaudRate:    c.baudRate,
		ReadTimeout: c.readTimeout.String(),
		LogDebug:    &c.logDebug,
	}
}

//lint:ignore U1000 linter does not catch that this is used generic code
func (c DeviceConfig) convertToRead() deviceConfigRead {
	return deviceConfigRead{
		Filter:                    c.filter.convertToRead(),
		RestartInterval:           c.restartInterval.String(),
		RestartIntervalMaxBackoff: c.restartIntervalMaxBackoff.String(),
		LogDebug:                  &c.logDebug,
		LogComDebug:               &c.logComDebug,
	}
}

//lint:ignore U1000 linter does not catch that this is used generic code
func (c VictronDeviceConfig) convertToRead() victronDeviceConfigRead {
	return victronDeviceConfigRead{
		deviceConfigRead: c.DeviceConfig.convertToRead(),
		Device:           c.device,
		Kind:             c.kind.String(),
	}
}

//lint:ignore U1000 linter does not catch that this is used generic code
func (c ModbusDeviceConfig) convertToRead() modbusDeviceConfigRead {
	return modbusDeviceConfigRead{
		deviceConfigRead: c.DeviceConfig.convertToRead(),
		Bus:              c.bus,
		Kind:             c.kind.String(),
		Address:          fmt.Sprintf("0x%02x", c.address),
		Relays: func(inp map[string]RelayConfig) (oup map[string]relayConfigRead) {
			oup = make(map[string]relayConfigRead, len(inp))
			for k, v := range inp {
				oup[k] = v.convertToRead()
			}
			return oup
		}(c.relays),
		PollInterval: c.pollInterval.String(),
	}
}

//lint:ignore U1000 linter does not catch that this is used generic code
func (c RelayConfig) convertToRead() relayConfigRead {
	return relayConfigRead{
		Description: &c.description,
		OpenLabel:   &c.openLabel,
		ClosedLabel: &c.closedLabel,
	}
}

//lint:ignore U1000 linter does not catch that this is used generic code
func (c HttpDeviceConfig) convertToRead() httpDeviceConfigRead {
	return httpDeviceConfigRead{
		deviceConfigRead: c.DeviceConfig.convertToRead(),
		Url:              c.url.String(),
		Kind:             c.kind.String(),
		Username:         c.username,
		Password:         c.password,
		PollInterval:     c.pollInterval.String(),
	}
}

//lint:ignore U1000 linter does not catch that this is used generic code
func (c MqttDeviceConfig) convertToRead() mqttDeviceConfigRead {
	return mqttDeviceConfigRead{
		deviceConfigRead: c.DeviceConfig.convertToRead(),
		Kind:             c.kind.String(),
		MqttClients:      c.mqttClients,
		MqttTopics:       c.mqttTopics,
	}
}

//lint:ignore U1000 linter does not catch that this is used generic code
func (c ViewConfig) convertToRead() viewConfigRead {
	return viewConfigRead{
		Name:         c.name,
		Title:        c.title,
		Devices:      convertListToRead[ViewDeviceConfig, viewDeviceConfigRead](c.devices),
		Autoplay:     &c.autoplay,
		AllowedUsers: maps.Keys(c.allowedUsers),
		Hidden:       &c.hidden,
	}
}

//lint:ignore U1000 linter does not catch that this is used generic code
func (c ViewDeviceConfig) convertToRead() viewDeviceConfigRead {
	return viewDeviceConfigRead{
		Name:   c.name,
		Title:  c.title,
		Filter: c.filter.convertToRead(),
	}
}

//lint:ignore U1000 linter does not catch that this is used generic code
func (c FilterConfig) convertToRead() filterConfigRead {
	return filterConfigRead{
		IncludeRegisters:  c.includeRegisters,
		SkipRegisters:     c.skipRegisters,
		IncludeCategories: c.includeCategories,
		SkipCategories:    c.skipCategories,
		DefaultInclude:    &c.defaultInclude,
	}
}
