package config

import "fmt"

func (c Config) MarshalYAML() (interface{}, error) {

	return configRead{
		Version:      &c.version,
		ProjectTitle: c.projectTitle,
		Authentication: func() *authenticationConfigRead {
			if !c.authentication.enabled {
				return nil
			}
			r := c.authentication.convertToRead()
			return &r
		}(),
		MqttClients: func() mqttClientConfigReadMap {
			mqttClients := make(mqttClientConfigReadMap, len(c.mqttClients))
			for _, c := range c.mqttClients {
				mqttClients[c.name] = c.convertToRead()
			}
			return mqttClients
		}(),
		VictronDevices: func() victronDeviceConfigReadMap {
			devices := make(victronDeviceConfigReadMap, len(c.devices))
			for _, c := range c.victronDevices {
				devices[c.name] = c.convertToRead()
			}
			return devices
		}(),
		ModbusDevices: func() modbusDeviceConfigReadMap {
			devices := make(modbusDeviceConfigReadMap, len(c.devices))
			for _, c := range c.modbusDevices {
				devices[c.name] = c.convertToRead()
			}
			return devices
		}(),
		HttpDevices: func() httpDeviceConfigReadMap {
			devices := make(httpDeviceConfigReadMap, len(c.devices))
			for _, c := range c.httpDevices {
				devices[c.name] = c.convertToRead()
			}
			return devices
		}(),
		MqttDevices: func() mqttDeviceConfigReadMap {
			devices := make(mqttDeviceConfigReadMap, len(c.devices))
			for _, c := range c.mqttDevices {
				devices[c.name] = c.convertToRead()
			}
			return devices
		}(),
		Views: func() viewConfigReadList {
			views := make(viewConfigReadList, len(c.views))
			i := 0
			for _, c := range c.views {
				views[i] = c.convertToRead()
				i++
			}
			return views
		}(),
		HttpServer: func() *httpServerConfigRead {
			if !c.httpServer.enabled {
				return nil
			}
			r := c.httpServer.convertToRead()
			return &r
		}(),
		LogConfig:       &c.logConfig,
		LogWorkerStart:  &c.logWorkerStart,
		LogStorageDebug: &c.logStorageDebug,
	}, nil
}

func (c AuthenticationConfig) convertToRead() authenticationConfigRead {
	jwtSecret := string(c.jwtSecret)
	return authenticationConfigRead{
		JwtSecret:         &jwtSecret,
		JwtValidityPeriod: c.jwtValidityPeriod.String(),
		HtaccessFile:      &c.htaccessFile,
	}
}

func (c MqttClientConfig) convertToRead() mqttClientConfigRead {
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

func (c DeviceConfig) convertToRead() deviceConfigRead {
	return deviceConfigRead{
		SkipFields:              c.skipFields,
		SkipCategories:          c.skipCategories,
		TelemetryViaMqttClients: c.telemetryViaMqttClients,
		RealtimeViaMqttClients:  c.realtimeViaMqttClients,
		LogDebug:                &c.logDebug,
		LogComDebug:             &c.logComDebug,
	}
}

func (c VictronDeviceConfig) convertToRead() victronDeviceConfigRead {
	return victronDeviceConfigRead{
		General: c.DeviceConfig.convertToRead(),
		Device:  c.device,
		Kind:    c.kind.String(),
	}
}

func (c ModbusDeviceConfig) convertToRead() modbusDeviceConfigRead {
	return modbusDeviceConfigRead{
		General: c.DeviceConfig.convertToRead(),
		Device:  c.device,
		Kind:    c.kind.String(),
		Address: fmt.Sprintf("0x%02x", c.address),
	}
}

func (c HttpDeviceConfig) convertToRead() httpDeviceConfigRead {
	return httpDeviceConfigRead{
		General:                c.DeviceConfig.convertToRead(),
		Url:                    c.url.String(),
		Kind:                   c.kind.String(),
		Username:               c.username,
		Password:               c.password,
		PollInterval:           c.pollInterval.String(),
		PollIntervalMaxBackoff: c.pollIntervalMaxBackoff.String(),
	}
}

func (c MqttDeviceConfig) convertToRead() mqttDeviceConfigRead {
	return mqttDeviceConfigRead{
		General:     c.DeviceConfig.convertToRead(),
		MqttTopics:  c.mqttTopics,
		MqttClients: c.mqttClients,
	}
}

func (c ViewDeviceConfig) convertToRead() viewDeviceConfigRead {
	return viewDeviceConfigRead{
		Name:  c.name,
		Title: c.title,
	}
}

func (c ViewConfig) convertToRead() viewConfigRead {
	return viewConfigRead{
		Name:  c.name,
		Title: c.title,
		Devices: func() viewDeviceConfigReadList {
			views := make(viewDeviceConfigReadList, len(c.devices))
			for i, c := range c.devices {
				views[i] = c.convertToRead()
			}
			return views
		}(),
		Autoplay:       &c.autoplay,
		AllowedUsers:   mapKeys(c.allowedUsers),
		Hidden:         &c.hidden,
		SkipFields:     c.skipFields,
		SkipCategories: c.skipCategories,
	}
}

func mapKeys(m map[string]struct{}) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

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
