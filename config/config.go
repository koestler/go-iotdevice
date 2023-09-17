package config

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
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

func ReadConfigFile(exe, source string, bypassFileCheck bool) (config Config, err []error) {
	yamlStr, e := os.ReadFile(source)
	if e != nil {
		return config, []error{fmt.Errorf("cannot read configuration: %v; use see `%s --help`", err, exe)}
	}

	return ReadConfig(yamlStr, bypassFileCheck)
}

func ReadConfig(yamlStr []byte, bypassFileCheck bool) (config Config, err []error) {
	var configRead configRead

	yamlStr = []byte(os.ExpandEnv(string(yamlStr)))
	e := yaml.Unmarshal(yamlStr, &configRead)
	if e != nil {
		return config, []error{fmt.Errorf("cannot parse yaml: %s", e)}
	}

	return configRead.TransformAndValidate(bypassFileCheck)
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

func (c configRead) TransformAndValidate(bypassFileCheck bool) (ret Config, err []error) {
	ret = Config{
		logConfig:      true,
		logWorkerStart: true,
	}

	var e []error

	if c.Version == nil {
		err = append(err, fmt.Errorf("version must be defined. Use Version=1"))
	} else {
		ret.version = *c.Version
		if ret.version != 1 {
			err = append(err, fmt.Errorf("version=%d is not supported", ret.version))
		}
	}

	if len(c.ProjectTitle) > 0 {
		ret.projectTitle = c.ProjectTitle
	} else {
		ret.projectTitle = "go-iotdevice"
	}

	if c.LogConfig != nil && !*c.LogConfig {
		ret.logConfig = false
	}

	if c.LogWorkerStart != nil && !*c.LogWorkerStart {
		ret.logWorkerStart = false
	}

	if c.LogStorageDebug != nil && *c.LogStorageDebug {
		ret.logStorageDebug = true
	}

	ret.httpServer, e = c.HttpServer.TransformAndValidate()
	err = append(err, e...)

	ret.authentication, e = c.Authentication.TransformAndValidate(bypassFileCheck)
	err = append(err, e...)

	ret.mqttClients, e = TransformAndValidateMapToList(
		c.MqttClients,
		func(inp mqttClientConfigRead, name string) (MqttClientConfig, []error) {
			return inp.TransformAndValidate(name)
		},
	)
	err = append(err, e...)

	ret.modbus, e = TransformAndValidateMapToList(
		c.Modbus,
		func(inp modbusConfigRead, name string) (ModbusConfig, []error) {
			return inp.TransformAndValidate(name)
		},
	)
	err = append(err, e...)

	ret.victronDevices, e = TransformAndValidateMapToList(
		c.VictronDevices,
		func(inp victronDeviceConfigRead, name string) (VictronDeviceConfig, []error) {
			return inp.TransformAndValidate(name, ret.mqttClients)
		},
	)
	err = append(err, e...)

	ret.modbusDevices, e = TransformAndValidateMapToList(
		c.ModbusDevices,
		func(inp modbusDeviceConfigRead, name string) (ModbusDeviceConfig, []error) {
			return inp.TransformAndValidate(name, ret.mqttClients, ret.modbus)
		},
	)
	err = append(err, e...)

	ret.httpDevices, e = TransformAndValidateMapToList(
		c.HttpDevices,
		func(inp httpDeviceConfigRead, name string) (HttpDeviceConfig, []error) {
			return inp.TransformAndValidate(name, ret.mqttClients)
		},
	)
	err = append(err, e...)

	ret.mqttDevices, e = TransformAndValidateMapToList(
		c.MqttDevices,
		func(inp mqttDeviceConfigRead, name string) (MqttDeviceConfig, []error) {
			return inp.TransformAndValidate(name, ret.mqttClients)
		},
	)
	err = append(err, e...)

	{
		ret.devices = make(
			[]DeviceConfig,
			len(ret.victronDevices)+len(ret.modbusDevices)+len(ret.httpDevices)+len(ret.mqttDevices),
		)

		i := 0
		for _, d := range ret.victronDevices {
			ret.devices[i] = d.DeviceConfig
			i += 1
		}
		for _, d := range ret.modbusDevices {
			ret.devices[i] = d.DeviceConfig
			i += 1
		}
		for _, d := range ret.httpDevices {
			ret.devices[i] = d.DeviceConfig
			i += 1
		}
		for _, d := range ret.mqttDevices {
			ret.devices[i] = d.DeviceConfig
			i += 1
		}
	}

	{
		var viewsErr []error
		ret.views, viewsErr = TransformAndValidateListUnique(
			c.Views,
			func(inp viewConfigRead) (ViewConfig, []error) {
				return inp.TransformAndValidate(ret.devices)
			},
		)
		for _, ve := range viewsErr {
			err = append(err, fmt.Errorf("section Views: %s", ve))
		}
	}

	ret.hassDiscovery, e = TransformAndValidateList(
		c.HassDiscovery,
		func(idx int, inp hassDiscoveryRead) (HassDiscovery, []error) {
			return inp.TransformAndValidate(idx, ret.mqttClients, ret.devices)
		},
	)
	err = append(err, e...)

	return
}

func (c *httpServerConfigRead) TransformAndValidate() (ret HttpServerConfig, err []error) {
	ret.enabled = false
	ret.port = 8000
	ret.logRequests = true

	if c == nil {
		return
	}

	ret.enabled = true

	if len(c.Bind) > 0 {
		ret.bind = c.Bind
	} else {
		err = append(err, errors.New("HttpServer->Bind must be either set or the whole section must be missing"))
	}

	if c.Port != nil {
		ret.port = *c.Port
	}

	if c.LogRequests != nil && !*c.LogRequests {
		ret.logRequests = false
	}

	if len(c.FrontendProxy) > 0 {
		u, parseError := url.Parse(c.FrontendProxy)
		if parseError == nil {
			ret.frontendProxy = u
		} else {
			err = append(err, fmt.Errorf("HttpServer->FrontendProxy must not be empty (=disabled) or a valid URL, err: %s", parseError))
		}
	}

	if len(c.FrontendPath) > 0 {
		ret.frontendPath = c.FrontendPath
	} else {
		ret.frontendPath = "./frontend-build/"
	}

	if len(c.FrontendExpires) < 1 {
		// use default 5min
		ret.frontendExpires = 5 * time.Minute
	} else if frontendExpires, e := time.ParseDuration(c.FrontendExpires); e != nil {
		err = append(err, fmt.Errorf("HttpServer->FrontendExpires='%s' parse error: %s", c.FrontendExpires, e))
	} else if frontendExpires < 0 {
		err = append(err, fmt.Errorf("HttpServer->FrontendExpires='%s' must be positive", c.FrontendExpires))
	} else {
		ret.frontendExpires = frontendExpires
	}

	if len(c.ConfigExpires) < 1 {
		// use default 1min
		ret.configExpires = 1 * time.Minute
	} else if configExpires, e := time.ParseDuration(c.ConfigExpires); e != nil {
		err = append(err, fmt.Errorf("HttpServer->ConfigExpires='%s' parse error: %s", c.ConfigExpires, e))
	} else if configExpires < 0 {
		err = append(err, fmt.Errorf("HttpServer->ConfigExpires='%s' must be positive", c.ConfigExpires))
	} else {
		ret.configExpires = configExpires
	}

	if c.LogDebug != nil && *c.LogDebug {
		ret.logDebug = true
	}

	return
}

func (c *authenticationConfigRead) TransformAndValidate(bypassFileCheck bool) (ret AuthenticationConfig, err []error) {
	ret.enabled = false
	ret.jwtValidityPeriod = time.Hour

	if randString, e := randomString(64); err == nil {
		ret.jwtSecret = []byte(randString)
	} else {
		err = append(err, fmt.Errorf("Authentication->JwtSecret: error while generating random secret: %s", e))
	}

	if c == nil {
		return
	}

	ret.enabled = true

	if c.JwtSecret != nil {
		if len(*c.JwtSecret) < 32 {
			err = append(err, fmt.Errorf("Authentication->JwtSecret must be empty or >= 32 chars"))
		} else {
			ret.jwtSecret = []byte(*c.JwtSecret)
		}
	}

	if len(c.JwtValidityPeriod) < 1 {
		// use default
	} else if authJwtValidityPeriod, e := time.ParseDuration(c.JwtValidityPeriod); e != nil {
		err = append(err, fmt.Errorf("Authentication->JwtValidityPeriod='%s' parse error: %s",
			c.JwtValidityPeriod, e,
		))
	} else if authJwtValidityPeriod < 0 {
		err = append(err, fmt.Errorf("Authentication->JwtValidityPeriod='%s' must be positive",
			c.JwtValidityPeriod,
		))
	} else {
		ret.jwtValidityPeriod = authJwtValidityPeriod
	}

	if c.HtaccessFile != nil && len(*c.HtaccessFile) > 0 {
		if !bypassFileCheck {
			if info, e := os.Stat(*c.HtaccessFile); e != nil {
				err = append(err, fmt.Errorf("Authentication->HtaccessFile='%s' cannot open file. error: %s",
					*c.HtaccessFile, e,
				))
			} else if info.IsDir() {
				err = append(err, fmt.Errorf("Authentication->HtaccessFile='%s' must be a file, not a directory",
					*c.HtaccessFile,
				))
			}
		}

		ret.htaccessFile = *c.HtaccessFile
	} else {
		err = append(err, errors.New("Authentication->HtaccessFile must not be empty"))
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

	if c.ReadOnly == nil {
		ret.readOnly = false
	} else {
		ret.readOnly = *c.ReadOnly
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

	if ret.readOnly {
		ret.availabilityTopic = ""
	} else if c.AvailabilityTopic == nil {
		// use default
		ret.availabilityTopic = "%Prefix%tele/%ClientId%/status"
	} else {
		ret.availabilityTopic = *c.AvailabilityTopic
	}

	if ret.readOnly {
		ret.telemetryInterval = 0
	} else if len(c.TelemetryInterval) < 1 {
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

	if c.RealtimeEnable == nil || ret.readOnly {
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

	if ret.readOnly {
		ret.maxBacklogSize = 0
	} else if c.MaxBacklogSize == nil {
		ret.maxBacklogSize = 256
	} else {
		ret.maxBacklogSize = *c.MaxBacklogSize
	}

	if c.LogDebug != nil && *c.LogDebug {
		ret.logDebug = true
	}

	if c.LogMessages != nil && *c.LogMessages {
		ret.logMessages = true
	}

	return
}

func (c hassDiscoveryRead) TransformAndValidate(idx int, mqttClients []MqttClientConfig, devices []DeviceConfig) (ret HassDiscovery, err []error) {
	ret = HassDiscovery{
		devices: c.Devices,
	}

	if c.Categories != nil {
		ret.categories = c.Categories
	} else {
		ret.categories = []string{}
	}

	if c.Registers != nil {
		ret.registers = c.Registers
	} else {
		ret.registers = []string{}
	}

	if c.TopicPrefix == nil {
		ret.topicPrefix = "homeassistant"
	} else {
		ret.topicPrefix = *c.TopicPrefix
	}

	var e []error

	ret.viaMqttClients, e = allOrCheckedMqttClients(
		c.ViaMqttClients, mqttClients,
		func(clientName string) error {
			return fmt.Errorf("HassDisovery->%d->MqttClients: client='%s' is not defined", idx, clientName)
		},
	)
	err = append(err, e...)

	if len(ret.devices) < 1 {
		err = append(err, fmt.Errorf("HassDiscovery->%d->Devices: must not be empty", idx))
	} else {
		for _, deviceName := range ret.devices {
			if !existsByName(deviceName, devices) {
				err = append(err, fmt.Errorf("HassDiscovery->%d->Devices: device='%s' is not defined", idx, deviceName))
			}
		}
	}

	ret.categoriesMatcher, e = stringToRegexp(c.Categories)
	err = append(err, e...)

	ret.registersMatcher, e = stringToRegexp(c.Registers)
	err = append(err, e...)

	return
}

func stringToRegexp(inp []string) (ret []*regexp.Regexp, err []error) {
	ret = make([]*regexp.Regexp, 0, len(inp))
	for _, v := range inp {
		if r, e := regexp.Compile(v); e != nil {
			err = append(err, fmt.Errorf("invalid regexp: %s", e))
		} else {
			ret = append(ret, r)
		}

	}

	return
}

func (c deviceConfigRead) TransformAndValidate(name string, mqttClients []MqttClientConfig) (ret DeviceConfig, err []error) {
	ret = DeviceConfig{
		name: name,
	}

	if c.SkipFields != nil {
		ret.skipFields = c.SkipFields
	} else {
		ret.skipFields = []string{}
	}

	if c.SkipCategories != nil {
		ret.skipCategories = c.SkipCategories
	} else {
		ret.skipCategories = []string{}
	}

	if !nameMatcher.MatchString(ret.name) {
		err = append(err, fmt.Errorf("Devices->Name='%s' does not match %s", ret.name, NameRegexp))
	}

	var e []error
	ret.telemetryViaMqttClients, e = allOrCheckedMqttClients(
		c.TelemetryViaMqttClients, mqttClients,
		func(clientName string) error {
			return fmt.Errorf("Devices->%s->TelemetryViaMqttClients: client='%s' is not defined", name, clientName)
		},
	)
	err = append(err, e...)

	ret.realtimeViaMqttClients, e = allOrCheckedMqttClients(
		c.RealtimeViaMqttClients, mqttClients,
		func(clientName string) error {
			return fmt.Errorf("Devices->%s->RealtimeViaMqttClients: client='%s' is not defined", name, clientName)
		},
	)
	err = append(err, e...)

	if len(c.RestartInterval) < 1 {
		// use default 200ms
		ret.restartInterval = 200 * time.Millisecond
	} else if restartInterval, e := time.ParseDuration(c.RestartInterval); e != nil {
		err = append(err, fmt.Errorf("Devices->%s->RestartInterval='%s' parse error: %s",
			name, c.RestartInterval, e,
		))
	} else if restartInterval < 10*time.Millisecond {
		err = append(err, fmt.Errorf("Devices->%s->RestartInterval='%s' must be >=10ms",
			name, c.RestartInterval,
		))
	} else {
		ret.restartInterval = restartInterval
	}

	if len(c.RestartIntervalMaxBackoff) < 1 {
		// use default 1min
		ret.restartIntervalMaxBackoff = time.Minute
	} else if restartIntervalMaxBackoff, e := time.ParseDuration(c.RestartIntervalMaxBackoff); e != nil {
		err = append(err, fmt.Errorf("Devices->%s->RestartIntervalMaxBackoff='%s' parse error: %s",
			name, c.RestartIntervalMaxBackoff, e,
		))
	} else if restartIntervalMaxBackoff < 10*time.Millisecond {
		err = append(err, fmt.Errorf("Devices->%s->RestartIntervalMaxBackoff='%s' must be >=10ms",
			name, c.RestartIntervalMaxBackoff,
		))
	} else {
		ret.restartIntervalMaxBackoff = restartIntervalMaxBackoff
	}

	if c.LogDebug != nil && *c.LogDebug {
		ret.logDebug = true
	}

	if c.LogComDebug != nil && *c.LogComDebug {
		ret.logComDebug = true
	}

	return
}

func (c victronDeviceConfigRead) TransformAndValidate(name string, mqttClients []MqttClientConfig) (ret VictronDeviceConfig, err []error) {
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

func (c modbusDeviceConfigRead) TransformAndValidate(
	name string, mqttClients []MqttClientConfig, modbus []ModbusConfig,
) (ret ModbusDeviceConfig, err []error) {
	ret = ModbusDeviceConfig{
		kind: ModbusDeviceKindFromString(c.Kind),
		bus:  c.Bus,
	}

	var e []error
	ret.DeviceConfig, e = c.General.TransformAndValidate(name, mqttClients)
	err = append(err, e...)

	if ret.kind == ModbusUndefinedKind {
		err = append(err, fmt.Errorf("ModbusDevices->%s->Kind='%s' is invalid", name, c.Kind))
	}

	if !existsByName(c.Bus, modbus) {
		err = append(err, fmt.Errorf("ModbusDevices->%s: Bus='%s' is not defidnedd", name, c.Bus))
	}

	if n, e := fmt.Sscanf(c.Address, "0x%x", &ret.address); n != 1 || e != nil {
		err = append(err, fmt.Errorf("ModbusDevices->%s: Adress=%s is invalid: %s", name, c.Address, e))
	}

	ret.relays, e = TransformAndValidateMap(
		c.Relays,
		func(inp relayConfigRead, name string) (RelayConfig, []error) {
			return inp.TransformAndValidate()
		},
	)
	err = append(err, e...)

	if len(c.PollInterval) < 1 {
		// use default 1s
		ret.pollInterval = time.Second
	} else if pollInterval, e := time.ParseDuration(c.PollInterval); e != nil {
		err = append(err, fmt.Errorf("HttpDevices->%s->PollInterval='%s' parse error: %s",
			name, c.PollInterval, e,
		))
	} else if pollInterval < time.Millisecond {
		err = append(err, fmt.Errorf("HttpDevices->%s->PollInterval='%s' must be >=1ms",
			name, c.PollInterval,
		))
	} else {
		ret.pollInterval = pollInterval
	}

	return
}

func (c relayConfigRead) TransformAndValidate() (ret RelayConfig, err []error) {
	ret = RelayConfig{
		description: "",
		openLabel:   "",
		closedLabel: "",
	}

	if c.Description != nil {
		ret.description = *c.Description
	}

	if c.OpenLabel != nil {
		ret.openLabel = *c.OpenLabel
	}

	if c.ClosedLabel != nil {
		ret.closedLabel = *c.ClosedLabel
	}

	return
}

func (c httpDeviceConfigRead) TransformAndValidate(name string, mqttClients []MqttClientConfig) (ret HttpDeviceConfig, err []error) {
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

	return
}

func (c mqttDeviceConfigRead) TransformAndValidate(name string, mqttClients []MqttClientConfig) (ret MqttDeviceConfig, err []error) {
	ret = MqttDeviceConfig{
		mqttTopics: c.MqttTopics,
	}

	var e []error
	ret.mqttClients, e = allOrCheckedMqttClients(
		c.MqttClients, mqttClients,
		func(clientName string) error {
			return fmt.Errorf("MqttDevices->%s->mqttClients: client='%s' is not defined", name, clientName)
		},
	)
	err = append(err, e...)

	// Do not allow for {Telemetry, Realtime}ViaMqttClients to contain MqttClients
	// This can possibly result in an infinite loop where we listen to messages published by our self
	filteredMqttClients := make([]MqttClientConfig, 0, len(mqttClients))
	for _, mc := range mqttClients {
		found := false
		for _, n := range ret.mqttClients {
			if n == mc.Name() {
				found = true
				break
			}
		}
		if !found {
			filteredMqttClients = append(filteredMqttClients, mc)
		}
	}

	ret.DeviceConfig, e = c.General.TransformAndValidate(name, filteredMqttClients)
	err = append(err, e...)

	return
}

func (c modbusConfigRead) TransformAndValidate(name string) (ret ModbusConfig, err []error) {
	ret = ModbusConfig{
		name:     name,
		device:   c.Device,
		baudRate: c.BaudRate,
	}

	if !nameMatcher.MatchString(ret.name) {
		err = append(err, fmt.Errorf("Modbus->Name='%s' does not match %s", ret.name, NameRegexp))
	}

	if len(c.Device) < 1 {
		err = append(err, fmt.Errorf("ModbusDevices->%s->Device must not be empty", name))
	}

	if c.BaudRate < 1 {
		err = append(err, fmt.Errorf("ModbusDevices->%s->BaudRate must be positiv", name))
	}

	if len(c.ReadTimeout) < 1 {
		// use default 100ms
		ret.readTimeout = 100 * time.Millisecond
	} else if readTimeout, e := time.ParseDuration(c.ReadTimeout); e != nil {
		err = append(err, fmt.Errorf("ModbusDevices->%s->ReadTimeout='%s' parse error: %s",
			name, c.ReadTimeout, e,
		))
	} else if readTimeout < time.Millisecond {
		err = append(err, fmt.Errorf("ModbusDevices->%s->ReadTimeout='%s' must be >=1ms",
			name, c.ReadTimeout,
		))
	} else {
		ret.readTimeout = readTimeout
	}

	if c.LogDebug != nil && *c.LogDebug {
		ret.logDebug = true
	}

	return
}

func (c viewConfigRead) TransformAndValidate(devices []DeviceConfig) (ret ViewConfig, err []error) {
	ret = ViewConfig{
		name:         c.Name,
		title:        c.Title,
		allowedUsers: make(map[string]struct{}),
		hidden:       false,
		autoplay:     true,
	}

	if !nameMatcher.MatchString(ret.name) {
		err = append(err, fmt.Errorf("Views->Name='%s' does not match %s", ret.name, NameRegexp))
	}

	if len(c.Title) < 1 {
		err = append(err, fmt.Errorf("Views->%s->Title must not be empty", c.Name))
	}

	{
		var devicesErr []error
		ret.devices, devicesErr = TransformAndValidateListUnique(
			c.Devices,
			func(inp viewDeviceConfigRead) (ViewDeviceConfig, []error) {
				return inp.TransformAndValidate(devices)
			},
		)

		for _, ce := range devicesErr {
			err = append(err, fmt.Errorf("section Views->%s: %s", c.Name, ce))
		}
	}

	if c.Autoplay != nil && !*c.Autoplay {
		ret.autoplay = false
	}

	for _, user := range c.AllowedUsers {
		ret.allowedUsers[user] = struct{}{}
	}

	if c.Hidden != nil && *c.Hidden {
		ret.hidden = true
	}

	return
}

func (c viewDeviceConfigRead) TransformAndValidate(
	devices []DeviceConfig,
) (ret ViewDeviceConfig, err []error) {
	ret = ViewDeviceConfig{
		name:  c.Name,
		title: c.Title,
	}

	if c.SkipFields != nil {
		ret.skipFields = c.SkipFields
	} else {
		ret.skipFields = []string{}
	}

	if c.SkipCategories != nil {
		ret.skipCategories = c.SkipCategories
	} else {
		ret.skipCategories = []string{}
	}

	if !existsByName(c.Name, devices) {
		err = append(err, fmt.Errorf("device='%s' is not defined", c.Name))
	}

	return
}

func TransformAndValidateMapToList[I any, O any](
	inp map[string]I,
	transformer func(inp I, name string) (ret O, err []error),
) (ret []O, err []error) {
	keys := maps.Keys(inp)
	sort.Strings(keys)

	ret = make([]O, len(inp))
	for i, name := range keys {
		var e []error
		ret[i], e = transformer(inp[name], name)
		err = append(err, e...)
	}
	return
}

func TransformAndValidateMap[I any, O any](
	inp map[string]I,
	transformer func(inp I, name string) (ret O, err []error),
) (ret map[string]O, err []error) {
	ret = make(map[string]O, len(inp))
	for k, v := range inp {
		var e []error
		ret[k], e = transformer(v, k)
		err = append(err, e...)
	}
	return
}

func TransformAndValidateList[I any, O any](
	inp []I,
	transformer func(idx int, inp I) (ret O, err []error),
) (ret []O, err []error) {
	ret = make([]O, len(inp))
	for i, cr := range inp {
		var e []error
		ret[i], e = transformer(i, cr)
		err = append(err, e...)
	}

	return
}

func TransformAndValidateListUnique[I any, O Nameable](
	inp []I,
	transformer func(inp I) (ret O, err []error),
) (ret []O, err []error) {
	ret = make([]O, 0, len(inp))
	for _, cr := range inp {
		r, e := transformer(cr)

		if existsByName(r.Name(), ret) {
			err = append(err, fmt.Errorf("duplicate name='%s'", r.Name()))
		}

		ret = append(ret, r)
		err = append(err, e...)
	}

	return
}

func allOrCheckedMqttClients(inp []string, mqttClients []MqttClientConfig, errorFunc func(clientName string) error) (oup []string, err []error) {
	if len(inp) < 1 {
		return getNames(mqttClients), nil
	}

	oup = make([]string, 0, len(inp))

	for _, clientName := range inp {
		if !existsByName(clientName, mqttClients) {
			err = append(err, errorFunc(clientName))
		} else {
			oup = append(oup, clientName)
		}
	}

	return
}

func existsByName[N Nameable](needle string, haystack []N) bool {
	for _, t := range haystack {
		if needle == t.Name() {
			return true
		}
	}
	return false
}

func getNames[N Nameable](list []N) (ret []string) {
	ret = make([]string, len(list))
	for i, t := range list {
		ret[i] = t.Name()
	}
	return
}
