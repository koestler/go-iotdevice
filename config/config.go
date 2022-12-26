package config

import (
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
	"log"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const NameRegexp = "^[a-zA-Z0-9\\-]{1,32}$"

var nameMatcher = regexp.MustCompile(NameRegexp)

func ReadConfigFile(exe, source string) (config Config, err []error) {
	yamlStr, e := os.ReadFile(source)
	if e != nil {
		return config, []error{fmt.Errorf("cannot read configuration: %v; use see `%s --help`", err, exe)}
	}

	return ReadConfig(yamlStr)
}

func ReadConfig(yamlStr []byte) (config Config, err []error) {
	var configRead configRead

	yamlStr = []byte(os.ExpandEnv(string(yamlStr)))
	e := yaml.Unmarshal(yamlStr, &configRead)
	if e != nil {
		return config, []error{fmt.Errorf("cannot parse yaml: %s", e)}
	}

	return configRead.TransformAndValidate()
}

func (c Config) PrintConfig() (err error) {
	newYamlStr, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("cannot encode yaml again: %s", err)
	}

	log.Print("config: use the following config:")
	for _, line := range strings.Split(string(newYamlStr), "\n") {
		log.Print("config: ", line)
	}
	return nil
}

func (c configRead) TransformAndValidate() (ret Config, err []error) {
	var e []error
	ret.auth, e = c.Auth.TransformAndValidate()
	err = append(err, e...)

	ret.mqttClients, e = c.MqttClients.TransformAndValidate()
	err = append(err, e...)

	ret.victronDevices, e = c.VictronDevices.TransformAndValidate(ret.mqttClients)
	err = append(err, e...)

	ret.httpDevices, e = c.HttpDevices.TransformAndValidate(ret.mqttClients)
	err = append(err, e...)

	ret.mqttDevices, e = c.MqttDevices.TransformAndValidate(ret.mqttClients)
	err = append(err, e...)

	{
		ret.devices = make([]*DeviceConfig, len(ret.victronDevices)+len(ret.httpDevices)+len(ret.mqttDevices))

		i := 0
		for _, d := range ret.victronDevices {
			ret.devices[i] = &d.DeviceConfig
			i += 1
		}
		for _, d := range ret.httpDevices {
			ret.devices[i] = &d.DeviceConfig
			i += 1
		}
		for _, d := range ret.mqttDevices {
			ret.devices[i] = &d.DeviceConfig
			i += 1
		}
	}

	ret.views, e = c.Views.TransformAndValidate(ret.devices)
	err = append(err, e...)

	ret.httpServer, e = c.HttpServer.TransformAndValidate()
	err = append(err, e...)

	if c.Version == nil {
		err = append(err, fmt.Errorf("Version must be defined. Use Version=1."))
	} else {
		ret.version = *c.Version
		if ret.version != 1 {
			err = append(err, fmt.Errorf("Version=%d is not supported.", ret.version))
		}
	}

	if len(c.ProjectTitle) > 0 {
		ret.projectTitle = c.ProjectTitle
	} else {
		ret.projectTitle = "go-iotdevice"
	}

	if c.LogConfig != nil && *c.LogConfig {
		ret.logConfig = true
	}

	if c.LogWorkerStart != nil && *c.LogWorkerStart {
		ret.logWorkerStart = true
	}

	if c.LogDebug != nil && *c.LogDebug {
		ret.logDebug = true
	}

	return
}

func (c *authConfigRead) TransformAndValidate() (ret AuthConfig, err []error) {
	ret.enabled = false
	ret.jwtValidityPeriod = time.Hour

	if randString, e := randomString(64); err == nil {
		ret.jwtSecret = []byte(randString)
	} else {
		err = append(err, fmt.Errorf("Auth->JwtSecret: error while generating random secret: %s", e))
	}

	if c == nil {
		return
	}

	ret.enabled = true

	if c.JwtSecret != nil {
		if len(*c.JwtSecret) < 32 {
			err = append(err, fmt.Errorf("Auth->JwtSecret must be empty ot >= 32 chars"))
		} else {
			ret.jwtSecret = []byte(*c.JwtSecret)
		}
	}

	if len(c.JwtValidityPeriod) < 1 {
		// use default
	} else if authJwtValidityPeriod, e := time.ParseDuration(c.JwtValidityPeriod); e != nil {
		err = append(err, fmt.Errorf("Auth->JwtValidityPeriod='%s' parse error: %s",
			c.JwtValidityPeriod, e,
		))
	} else if authJwtValidityPeriod < 0 {
		err = append(err, fmt.Errorf("Auth->JwtValidityPeriod='%s' must be positive",
			c.JwtValidityPeriod,
		))
	} else {
		ret.jwtValidityPeriod = authJwtValidityPeriod
	}

	if c.HtaccessFile != nil && len(*c.HtaccessFile) > 0 {
		if info, e := os.Stat(*c.HtaccessFile); e != nil {
			err = append(err, fmt.Errorf("Auth->HtaccessFile='%s' cannot open file. error: %s",
				*c.HtaccessFile, e,
			))
		} else if info.IsDir() {
			err = append(err, fmt.Errorf("Auth->HtaccessFile='%s' must be a file, not a directory",
				*c.HtaccessFile,
			))
		}

		ret.htaccessFile = *c.HtaccessFile
	}

	return
}

func (c *httpServerConfigRead) TransformAndValidate() (ret HttpServerConfig, err []error) {
	ret.enabled = false
	ret.bind = "[::1]"
	ret.port = 8000

	if c == nil {
		return
	}

	ret.enabled = true

	if len(c.Bind) > 0 {
		ret.bind = c.Bind
	}

	if c.Port != nil {
		ret.port = *c.Port
	}

	if c.LogRequests != nil && *c.LogRequests {
		ret.logRequests = true
	}

	if len(c.FrontendProxy) > 0 {
		u, parseError := url.Parse(c.FrontendProxy)
		if parseError == nil {
			ret.frontendProxy = u
		} else {
			err = append(err, fmt.Errorf("HttpServerConfig->FrontendProxy must not be empty (=disabled) or a valid URL, err: %s", parseError))
		}
	}

	if len(c.FrontendPath) > 0 {
		ret.frontendPath = c.FrontendPath
	} else {
		ret.frontendPath = "frontend-build"
	}

	if len(c.FrontendExpires) < 1 {
		// use default 5min
		ret.frontendExpires = 5 * time.Minute
	} else if frontendExpires, e := time.ParseDuration(c.FrontendExpires); e != nil {
		err = append(err, fmt.Errorf("HttpServerConfig->FrontendExpires='%s' parse error: %s", c.FrontendExpires, e))
	} else if frontendExpires < 0 {
		err = append(err, fmt.Errorf("HttpServerConfig->FrontendExpires='%s' must be positive", c.FrontendExpires))
	} else {
		ret.frontendExpires = frontendExpires
	}

	if len(c.ConfigExpires) < 1 {
		// use default 1min
		ret.configExpires = 1 * time.Minute
	} else if configExpires, e := time.ParseDuration(c.ConfigExpires); e != nil {
		err = append(err, fmt.Errorf("HttpServerConfig->ConfigExpires='%s' parse error: %s", c.ConfigExpires, e))
	} else if configExpires < 0 {
		err = append(err, fmt.Errorf("HttpServerConfig->ConfigExpires='%s' must be positive", c.ConfigExpires))
	} else {
		ret.configExpires = configExpires
	}

	return
}

func (c mqttClientConfigReadMap) getOrderedKeys() (ret []string) {
	ret = make([]string, len(c))
	i := 0
	for k := range c {
		ret[i] = k
		i++
	}
	sort.Strings(ret)
	return
}

func (c mqttClientConfigReadMap) TransformAndValidate() (ret []*MqttClientConfig, err []error) {
	ret = make([]*MqttClientConfig, len(c))
	j := 0
	for _, name := range c.getOrderedKeys() {
		r, e := c[name].TransformAndValidate(name)
		ret[j] = &r
		err = append(err, e...)
		j++
	}
	return
}

func (c mqttClientConfigRead) TransformAndValidate(name string) (ret MqttClientConfig, err []error) {
	ret = MqttClientConfig{
		name:        name,
		user:        c.User,
		password:    c.Password,
		topicPrefix: c.TopicPrefix,
	}

	if !nameMatcher.MatchString(ret.name) {
		err = append(err, fmt.Errorf("MqttClientConfig->Name='%s' does not match %s", ret.name, NameRegexp))
	}

	if len(c.Broker) < 1 {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->Broker must not be empty", name))
	} else {
		if broker, e := url.ParseRequestURI(c.Broker); e != nil {
			err = append(err, fmt.Errorf("MqttClientConfig->%s->Broker invalid url: %s", name, e))
		} else if broker == nil {
			err = append(err, fmt.Errorf("MqttClientConfig->%s->Broker cannot parse broker", name))
		} else {
			ret.broker = broker
		}
	}

	if c.ProtocolVersion == nil {
		ret.protocolVersion = 5
	} else if *c.ProtocolVersion == 3 || *c.ProtocolVersion == 5 {
		ret.protocolVersion = *c.ProtocolVersion
	} else {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->Protocol=%d but must be 3 or 5", name, *c.ProtocolVersion))
	}

	if c.ClientId == nil {
		ret.clientId = "go-iotdevice-" + uuid.New().String()
	} else {
		ret.clientId = *c.ClientId
	}

	if c.Qos == nil {
		ret.qos = 1 // default qos is 1
	} else if *c.Qos == 0 || *c.Qos == 1 || *c.Qos == 2 {
		ret.qos = *c.Qos
	} else {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->Qos=%d but must be 0, 1 or 2", name, *c.Qos))
	}

	if len(c.KeepAlive) < 1 {
		// use default 60s
		ret.keepAlive = time.Minute
	} else if keepAlive, e := time.ParseDuration(c.KeepAlive); e != nil {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->KeepAlive='%s' parse error: %s",
			name, c.KeepAlive, e,
		))
	} else if keepAlive < time.Second {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->KeepAlive='%s' must be >=1s",
			name, c.KeepAlive,
		))
	} else if keepAlive%time.Second != 0 {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->KeepAlive='%s' must be a multiple of a second",
			name, c.KeepAlive,
		))
	} else {
		ret.keepAlive = keepAlive
	}

	if len(c.ConnectRetryDelay) < 1 {
		// use default 10s
		ret.connectRetryDelay = 10 * time.Second
	} else if connectRetryDelay, e := time.ParseDuration(c.ConnectRetryDelay); e != nil {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->ConnectRetryDelay='%s' parse error: %s",
			name, c.ConnectRetryDelay, e,
		))
	} else if connectRetryDelay < 100*time.Millisecond {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->ConnectRetryDelay='%s' must be >=100ms",
			name, c.ConnectRetryDelay,
		))
	} else {
		ret.connectRetryDelay = connectRetryDelay
	}

	if len(c.ConnectTimeout) < 1 {
		// use default 5s
		ret.connectTimeout = 5 * time.Second
	} else if connectTimeout, e := time.ParseDuration(c.ConnectTimeout); e != nil {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->ConnectTimeout='%s' parse error: %s",
			name, c.ConnectTimeout, e,
		))
	} else if connectTimeout < 100*time.Millisecond {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->ConnectTimeout='%s' must be >=100ms",
			name, c.ConnectTimeout,
		))
	} else {
		ret.connectTimeout = connectTimeout
	}

	if c.AvailabilityTopic == nil {
		// use default
		ret.availabilityTopic = "%Prefix%tele/%ClientId%/status"
	} else {
		ret.availabilityTopic = *c.AvailabilityTopic
	}

	if len(c.TelemetryInterval) < 1 {
		// use default 10s
		ret.telemetryInterval = 10 * time.Second
	} else if telemetryInterval, e := time.ParseDuration(c.TelemetryInterval); e != nil {
		err = append(err, fmt.Errorf("HttpServerConfig->TelemetryInterval='%s' parse error: %s", c.TelemetryInterval, e))
	} else if telemetryInterval < 0 {
		err = append(err, fmt.Errorf("HttpServerConfig->TelemetryInterval='%s' must be positive", c.TelemetryInterval))
	} else {
		ret.telemetryInterval = telemetryInterval
	}

	if c.TelemetryTopic == nil {
		ret.telemetryTopic = "%Prefix%tele/go-iotdevice/%DeviceName%/state"
	} else {
		ret.telemetryTopic = *c.TelemetryTopic
	}

	if c.TelemetryRetain == nil {
		ret.realtimeRetain = false
	} else {
		ret.telemetryRetain = *c.TelemetryRetain
	}

	if c.RealtimeEnable == nil {
		ret.realtimeEnable = false
	} else {
		ret.realtimeEnable = *c.RealtimeEnable
	}

	if c.RealtimeTopic == nil {
		ret.realtimeTopic = "%Prefix%stat/go-iotdevice/%DeviceName%/%ValueName%"
	} else {
		ret.realtimeTopic = *c.RealtimeTopic
	}

	if c.RealtimeRetain == nil {
		ret.realtimeRetain = true
	} else {
		ret.realtimeRetain = *c.RealtimeRetain
	}

	if c.LogDebug != nil && *c.LogDebug {
		ret.logDebug = true
	}

	if c.LogMessages != nil && *c.LogMessages {
		ret.logMessages = true
	}

	return
}

func (c victronDeviceConfigReadMap) TransformAndValidate(mqttClients []*MqttClientConfig) (ret []*VictronDeviceConfig, err []error) {
	// order map keys by name
	keys := make([]string, len(c))
	i := 0
	for k := range c {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	ret = make([]*VictronDeviceConfig, len(c))
	j := 0
	for _, name := range keys {
		r, e := c[name].TransformAndValidate(name, mqttClients)
		ret[j] = &r
		err = append(err, e...)
		j++
	}
	return
}

func (c httpDeviceConfigReadMap) TransformAndValidate(mqttClients []*MqttClientConfig) (ret []*HttpDeviceConfig, err []error) {
	// order map keys by name
	keys := make([]string, len(c))
	i := 0
	for k := range c {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	ret = make([]*HttpDeviceConfig, len(c))
	j := 0
	for _, name := range keys {
		r, e := c[name].TransformAndValidate(name, mqttClients)
		ret[j] = &r
		err = append(err, e...)
		j++
	}
	return
}

func (c mqttDeviceConfigReadMap) TransformAndValidate(mqttClients []*MqttClientConfig) (ret []*MqttDeviceConfig, err []error) {
	// order map keys by name
	keys := make([]string, len(c))
	i := 0
	for k := range c {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	ret = make([]*MqttDeviceConfig, len(c))
	j := 0
	for _, name := range keys {
		r, e := c[name].TransformAndValidate(name, mqttClients)
		ret[j] = &r
		err = append(err, e...)
		j++
	}
	return
}

func (c deviceConfigRead) TransformAndValidate(name string, mqttClients []*MqttClientConfig) (ret DeviceConfig, err []error) {
	ret = DeviceConfig{
		name:                    name,
		telemetryViaMqttClients: c.TelemetryViaMqttClients,
		realtimeViaMqttClients:  c.RealtimeViaMqttClients,
		skipFields:              c.SkipFields,
		skipCategories:          c.SkipCategories,
	}

	if !nameMatcher.MatchString(ret.name) {
		err = append(err, fmt.Errorf("Devices->Name='%s' does not match %s", ret.name, NameRegexp))
	}

	for _, clientName := range ret.telemetryViaMqttClients {
		if !mqttClientExists(clientName, mqttClients) {
			err = append(err, fmt.Errorf("Devices->%s->TelemetryViaMqttClients client='%s' is not defined", name, clientName))
		}
	}
	for _, clientName := range ret.realtimeViaMqttClients {
		if !mqttClientExists(clientName, mqttClients) {
			err = append(err, fmt.Errorf("Devices->%s->RealtimeViaMqttClients client='%s' is not defined", name, clientName))
		}
	}

	if c.LogDebug != nil && *c.LogDebug {
		ret.logDebug = true
	}

	if c.LogComDebug != nil && *c.LogComDebug {
		ret.logComDebug = true
	}

	return
}

func (c victronDeviceConfigRead) TransformAndValidate(name string, mqttClients []*MqttClientConfig) (ret VictronDeviceConfig, err []error) {
	ret = VictronDeviceConfig{
		kind:   VictronDeviceKindFromString(c.Kind),
		device: c.Device,
	}

	var e []error
	ret.DeviceConfig, e = c.General.TransformAndValidate(name, mqttClients)
	err = append(err, e...)

	if ret.kind == VictronUndefinedKind {
		err = append(err, fmt.Errorf("VictronDevices->%s->Kind='%s' is invalid", name, c.Kind))
	}

	if ret.kind == VictronVedirectKind && len(c.Device) < 1 {
		err = append(err, fmt.Errorf("VictronDevices->%s->Device must not be empty", name))
	}

	return
}

func (c httpDeviceConfigRead) TransformAndValidate(name string, mqttClients []*MqttClientConfig) (ret HttpDeviceConfig, err []error) {
	ret = HttpDeviceConfig{
		kind:     HttpDeviceKindFromString(c.Kind),
		username: c.Username,
		password: c.Password,
	}

	var e []error
	ret.DeviceConfig, e = c.General.TransformAndValidate(name, mqttClients)
	err = append(err, e...)

	if len(c.Url) < 1 {
		err = append(err, fmt.Errorf("HttpDevices->%s->Url must not be empty", name))
	} else {
		if u, e := url.ParseRequestURI(c.Url); e != nil {
			err = append(err, fmt.Errorf("HttpDevices->%s->Url invalid url: %s", name, e))
		} else if u == nil {
			err = append(err, fmt.Errorf("HttpDevices->%s->Url cannot parse url", name))
		} else {
			ret.url = u
		}
	}

	if ret.kind == HttpUndefinedKind {
		err = append(err, fmt.Errorf("HttpDevices->%s->Kind='%s' is invalid", name, c.Kind))
	}

	if len(c.PollInterval) < 1 {
		// use default 1s
		ret.pollInterval = time.Second
	} else if pollInterval, e := time.ParseDuration(c.PollInterval); e != nil {
		err = append(err, fmt.Errorf("HttpDevices->%s->PollInterval='%s' parse error: %s",
			name, c.PollInterval, e,
		))
	} else if pollInterval < 100*time.Millisecond {
		err = append(err, fmt.Errorf("HttpDevices->%s->PollInterval='%s' must be >=100ms",
			name, c.PollInterval,
		))
	} else {
		ret.pollInterval = pollInterval
	}

	if len(c.PollIntervalMaxBackoff) < 1 {
		// use default 10s
		ret.pollIntervalMaxBackoff = 10 * time.Second
	} else if pollIntervalMaxBackoff, e := time.ParseDuration(c.PollIntervalMaxBackoff); e != nil {
		err = append(err, fmt.Errorf("HttpDevices->%s->PollIntervalMaxBackoff='%s' parse error: %s",
			name, c.PollIntervalMaxBackoff, e,
		))
	} else if pollIntervalMaxBackoff < 100*time.Millisecond {
		err = append(err, fmt.Errorf("HttpDevices->%s->PollIntervalMaxBackoff='%s' must be >=100ms",
			name, c.PollIntervalMaxBackoff,
		))
	} else {
		ret.pollIntervalMaxBackoff = pollIntervalMaxBackoff
	}

	return
}

func (c mqttDeviceConfigRead) TransformAndValidate(name string, mqttClients []*MqttClientConfig) (ret MqttDeviceConfig, err []error) {
	ret = MqttDeviceConfig{
		mqttTopics:  c.MqttTopics,
		mqttClients: c.MqttClients,
	}

	var e []error
	ret.DeviceConfig, e = c.General.TransformAndValidate(name, mqttClients)
	err = append(err, e...)

	for _, clientName := range ret.mqttClients {
		if !mqttClientExists(clientName, mqttClients) {
			err = append(err, fmt.Errorf("MqttDevices->%s->mqttClients client='%s' is not defined", name, clientName))
		}
	}

	return
}

func mqttClientExists(clientName string, mqttClients []*MqttClientConfig) bool {
	for _, client := range mqttClients {
		if clientName == client.name {
			return true
		}
	}
	return false
}

func (c viewConfigReadList) TransformAndValidate(devices []*DeviceConfig) (ret []*ViewConfig, err []error) {
	if len(c) < 1 {
		return ret, []error{fmt.Errorf("Views section must no be empty.")}
	}

	ret = make([]*ViewConfig, len(c))
	j := 0
	for _, cr := range c {
		r, e := cr.TransformAndValidate(devices)

		// check for duplicate name
		for i := 0; i < j; i++ {
			if r.Name() == ret[i].Name() {
				err = append(err, fmt.Errorf("Views->Name='%s': name must be unique", r.Name()))
			}
		}

		ret[j] = &r
		err = append(err, e...)
		j++
	}

	return
}

func (c viewConfigRead) TransformAndValidate(devices []*DeviceConfig) (ret ViewConfig, err []error) {
	ret = ViewConfig{
		name:           c.Name,
		title:          c.Title,
		allowedUsers:   make(map[string]struct{}),
		hidden:         false,
		skipFields:     c.SkipFields,
		skipCategories: c.SkipCategories,
	}

	if !nameMatcher.MatchString(ret.name) {
		err = append(err, fmt.Errorf("Views->Name='%s' does not match %s", ret.name, NameRegexp))
	}

	if len(c.Title) < 1 {
		err = append(err, fmt.Errorf("Views->%s->Title must not be empty", c.Name))
	}

	{
		var devicesErr []error
		ret.devices, devicesErr = c.Devices.TransformAndValidate(devices)
		for _, ce := range devicesErr {
			err = append(err, fmt.Errorf("Views->%s: %s", c.Name, ce))
		}
	}

	if c.Autoplay != nil && *c.Autoplay {
		ret.autoplay = true
	}

	for _, user := range c.AllowedUsers {
		ret.allowedUsers[user] = struct{}{}
	}

	if c.Hidden != nil && *c.Hidden {
		ret.hidden = true
	}

	return
}

func (c viewDeviceConfigReadList) TransformAndValidate(devices []*DeviceConfig) (ret []*ViewDeviceConfig, err []error) {
	if len(c) < 1 {
		return ret, []error{fmt.Errorf("Clients section must no be empty.")}
	}

	ret = make([]*ViewDeviceConfig, len(c))
	for i, device := range c {
		r, e := device.TransformAndValidate(devices)
		ret[i] = &r
		err = append(err, e...)
	}
	return

}

func (c viewDeviceConfigRead) TransformAndValidate(
	devices []*DeviceConfig,
) (ret ViewDeviceConfig, err []error) {
	if !deviceExists(c.Name, devices) {
		err = append(err, fmt.Errorf("Device='%s' is not defined", c.Name))
	}

	ret = ViewDeviceConfig{
		name:  c.Name,
		title: c.Title,
	}

	return
}

func deviceExists(deviceName string, devices []*DeviceConfig) bool {
	for _, client := range devices {
		if deviceName == client.name {
			return true
		}
	}
	return false
}
