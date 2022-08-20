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
		AvailabilityTopic:  &c.availabilityTopic,
		TelemetryInterval:  c.telemetryInterval.String(),
		TelemetryTopic:     &c.telemetryTopic,
		TelemetryRetain:    &c.telemetryRetain,
		RealtimeEnable:     &c.realtimeEnable,
		RealtimeTopic:      &c.realtimeTopic,
		RealtimeRetain:     &c.realtimeRetain,
		LogDebug:           &c.logDebug,
	}
}

func (c DeviceConfig) convertToRead() deviceConfigRead {
	return deviceConfigRead{
		Kind:           c.kind.String(),
		Device:         c.device,
		SkipFields:     c.skipFields,
		SkipCategories: c.skipCategories,
		LogDebug:       &c.logDebug,
		LogComDebug:    &c.logComDebug,
	}
}

func (c ViewDeviceConfig) convertToRead() viewDeviceConfigRead {
	return viewDeviceConfigRead{
		Name:           c.name,
		Title:          c.title,
		SkipFields:     c.skipFields,
		SkipCategories: c.skipCategories,
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
		Autoplay:     &c.autoplay,
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
