package config

func (c Config) MarshalYAML() (interface{}, error) {

	return configRead{
		Version:      &c.version,
		ProjectTitle: c.projectTitle,
		Auth: func() *authConfigRead {
			if !c.auth.enabled {
				return nil
			}
			r := c.auth.convertToRead()
			return &r
		}(),
		MqttClients: func() mqttClientConfigReadMap {
			mqttClients := make(mqttClientConfigReadMap, len(c.mqttClients))
			for _, c := range c.mqttClients {
				mqttClients[c.name] = c.convertToRead()
			}
			return mqttClients
		}(),
		Devices: func() deviceConfigReadMap {
			devices := make(deviceConfigReadMap, len(c.devices))
			for _, c := range c.devices {
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
		LogConfig:      &c.logConfig,
		LogWorkerStart: &c.logWorkerStart,
		LogDebug:       &c.logDebug,
	}, nil
}

func (c AuthConfig) convertToRead() authConfigRead {
	jwtSecret := string(c.jwtSecret)
	return authConfigRead{
		JwtSecret:         &jwtSecret,
		JwtValidityPeriod: c.jwtValidityPeriod.String(),
		HtaccessFile:      &c.htaccessFile,
	}
}

func (c MqttClientConfig) convertToRead() mqttClientConfigRead {
	return mqttClientConfigRead{
		Broker:             c.broker,
		User:               c.user,
		Password:           c.password,
		ClientId:           c.clientId,
		Qos:                &c.qos,
		TopicPrefix:        &c.topicPrefix,
		AvailabilityEnable: &c.availabilityEnable,
		TelemetryInterval:  c.telemetryInterval.String(),
		TelemetryTopic:     &c.telemetryTopic,
		TelemetryRetain:    &c.telemetryRetain,
		RealtimeEnable:     &c.realtimeEnable,
		RealtimeTopic:      &c.realtimeTopic,
		RealtimeRetain:     &c.realtimeRetain,
		LogMessages:        &c.logMessages,
	}
}

func (c DeviceConfig) convertToRead() deviceConfigRead {
	return deviceConfigRead{
		Device: c.device,
		Kind:   c.kind.String(),
	}
}

func (c ViewDeviceConfig) convertToRead() viewDeviceConfigRead {
	return viewDeviceConfigRead{
		Title:  c.title,
		Fields: c.fields,
	}
}

func (c ViewConfig) convertToRead() viewConfigRead {
	return viewConfigRead{
		Name:  c.name,
		Title: c.title,
		Devices: func() viewDeviceConfigReadMap {
			views := make(viewDeviceConfigReadMap, len(c.devices))
			for _, c := range c.devices {
				views[c.name] = c.convertToRead()
			}
			return views
		}(),
		AllowedUsers: mapKeys(c.allowedUsers),
		Hidden:       &c.hidden,
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
		EnableDocs:      &c.enableDocs,
		FrontendProxy:   frontendProxy,
		FrontendPath:    c.frontendPath,
		FrontendExpires: c.frontendExpires.String(),
		ConfigExpires:   c.configExpires.String(),
	}
}
