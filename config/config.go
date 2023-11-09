package config

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/koestler/go-iotdevice/types"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
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

	d := yaml.NewDecoder(bytes.NewReader(yamlStr))
	d.KnownFields(true)

	if e := d.Decode(&configRead); e != nil {
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
		err = append(err, fmt.Errorf("Version must be defined. Use Version=2"))
	} else {
		ret.version = *c.Version
		if ret.version != 2 {
			err = append(err, fmt.Errorf("version=%d is not supported, only version=2 is supported", ret.version))
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

	if c.LogStateStorageDebug != nil && *c.LogStateStorageDebug {
		ret.logStateStorageDebug = true
	}

	if c.LogCommandStorageDebug != nil && *c.LogCommandStorageDebug {
		ret.logCommandStorageDebug = true
	}

	ret.httpServer, e = c.HttpServer.TransformAndValidate()
	err = append(err, e...)

	ret.authentication, e = c.Authentication.TransformAndValidate(bypassFileCheck)
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
			return inp.TransformAndValidate(name)
		},
	)
	err = append(err, e...)

	ret.modbusDevices, e = TransformAndValidateMapToList(
		c.ModbusDevices,
		func(inp modbusDeviceConfigRead, name string) (ModbusDeviceConfig, []error) {
			return inp.TransformAndValidate(name, ret.modbus)
		},
	)
	err = append(err, e...)

	ret.httpDevices, e = TransformAndValidateMapToList(
		c.HttpDevices,
		func(inp httpDeviceConfigRead, name string) (HttpDeviceConfig, []error) {
			return inp.TransformAndValidate(name)
		},
	)
	err = append(err, e...)

	nonMqttDevices := make(
		[]DeviceConfig, 0,
		len(ret.victronDevices)+len(ret.modbusDevices)+len(ret.httpDevices),
	)

	for _, d := range ret.victronDevices {
		nonMqttDevices = append(nonMqttDevices, d.DeviceConfig)
	}
	for _, d := range ret.modbusDevices {
		nonMqttDevices = append(nonMqttDevices, d.DeviceConfig)
	}
	for _, d := range ret.httpDevices {
		nonMqttDevices = append(nonMqttDevices, d.DeviceConfig)
	}

	ret.mqttClients, e = TransformAndValidateMapToList(
		c.MqttClients,
		func(inp mqttClientConfigRead, name string) (MqttClientConfig, []error) {
			return inp.TransformAndValidate(name, nonMqttDevices)
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

	// create devices array
	ret.devices = make([]DeviceConfig, 0, len(nonMqttDevices)+len(ret.mqttDevices))
	ret.devices = append(ret.devices, nonMqttDevices...)
	for _, d := range ret.mqttDevices {
		ret.devices = append(ret.devices, d.DeviceConfig)
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

func (c mqttClientConfigRead) TransformAndValidate(name string, devices []DeviceConfig) (ret MqttClientConfig, err []error) {
	ret = MqttClientConfig{
		name:     name,
		user:     c.User,
		password: c.Password,
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
	} else if *c.ProtocolVersion == 5 {
		ret.protocolVersion = *c.ProtocolVersion
	} else {
		err = append(err, fmt.Errorf("MqttClientConfig->%s->Protocol=%d but must be 5 (3 is not supported anymore)", name, *c.ProtocolVersion))
	}

	if c.ClientId == nil {
		ret.clientId = "go-iotdevice-" + uuid.New().String()
	} else {
		ret.clientId = *c.ClientId
	}

	if len(c.KeepAlive) < 1 {
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

	if c.TopicPrefix == nil {
		ret.topicPrefix = "go-iotdevice/"
	} else {
		ret.topicPrefix = *c.TopicPrefix
	}

	if c.ReadOnly == nil {
		ret.readOnly = false
	} else {
		ret.readOnly = *c.ReadOnly
	}

	if ret.readOnly {
		ret.maxBacklogSize = 0
	} else if c.MaxBacklogSize == nil {
		ret.maxBacklogSize = 256
	} else {
		ret.maxBacklogSize = *c.MaxBacklogSize
	}

	var e []error
	ret.availabilityClient, e = c.AvailabilityClient.TransformAndValidate(
		fmt.Sprintf("MqttClientConfig->%s->AvailabilityClient->", name),
		devices,
		ret.readOnly,
		true,
		"%Prefix%avail/%ClientId%",
		0,
		false,
		true,
		true,
		true,
		false,
		false,
	)
	err = append(err, e...)

	ret.availabilityDevice, e = c.AvailabilityDevice.TransformAndValidate(
		fmt.Sprintf("MqttClientConfig->%s->AvailabilityDevice->", name),
		devices,
		ret.readOnly,
		true,
		"%Prefix%avail/%DeviceName%",
		0,
		false,
		true,
		true,
		true,
		true,
		false,
	)
	err = append(err, e...)

	ret.structure, e = c.Structure.TransformAndValidate(
		fmt.Sprintf("MqttClientConfig->%s->Structure->", name),
		devices,
		ret.readOnly,
		false,
		"%Prefix%struct/%DeviceName%",
		0,
		true,
		true,
		true,
		true,
		true,
		true,
	)
	err = append(err, e...)

	ret.telemetry, e = c.Telemetry.TransformAndValidate(
		fmt.Sprintf("MqttClientConfig->%s->Telemetry->", name),
		devices,
		ret.readOnly,
		false,
		"%Prefix%tele/%DeviceName%",
		time.Second,
		true,
		false,
		true,
		false,
		true,
		true,
	)
	err = append(err, e...)

	ret.realtime, e = c.Realtime.TransformAndValidate(
		fmt.Sprintf("MqttClientConfig->%s->Realtime->", name),
		devices,
		ret.readOnly,
		false,
		"%Prefix%real/%DeviceName%/%RegisterName%",
		0,
		true,
		true,
		true,
		false,
		true,
		true,
	)
	err = append(err, e...)

	ret.homeassistantDiscovery, e = c.HomeassistantDiscovery.TransformAndValidate(
		fmt.Sprintf("MqttClientConfig->%s->HomeassistantDiscovery->", name),
		devices,
		ret.readOnly,
		false,
		"homeassistant/%Component%/%NodeId%/%ObjectId%/config",
		0,
		true,
		true,
		true,
		false,
		true,
		true,
	)
	err = append(err, e...)

	ret.command, e = c.Command.TransformAndValidate(
		fmt.Sprintf("MqttClientConfig->%s->Command->", name),
		devices,
		ret.readOnly,
		false,
		"%Prefix%cmnd/%DeviceName%/%RegisterName%",
		0,
		false,
		true,
		false,
		false,
		true,
		true,
	)
	err = append(err, e...)

	if c.LogDebug != nil && *c.LogDebug {
		ret.logDebug = true
	}

	if c.LogMessages != nil && *c.LogMessages {
		ret.logMessages = true
	}

	return
}

func (c mqttSectionConfigRead) TransformAndValidate(
	logPrefix string,
	devices []DeviceConfig,
	readOnly bool,
	defaultEnabled bool,
	defaultTopicTemplate string,
	defaultInterval time.Duration,
	allowInterval bool,
	allowZeroInterval bool,
	allowRetain bool,
	defaultRetain bool,
	allowDevices bool,
	allowFilter bool,
) (ret MqttSectionConfig, err []error) {
	if readOnly {
		ret.enabled = false
	} else if c.Enabled == nil {
		ret.enabled = defaultEnabled
	} else {
		ret.enabled = *c.Enabled
	}

	if c.TopicTemplate == nil {
		ret.topicTemplate = defaultTopicTemplate
	} else if len(*c.TopicTemplate) < 1 {
		err = append(err, fmt.Errorf("%sTopicTemplate must no be empty", logPrefix))
	} else {
		ret.topicTemplate = *c.TopicTemplate
	}

	if !allowInterval {
		if len(c.Interval) > 0 {
			err = append(err, fmt.Errorf("%sInterval not supported", logPrefix))
		}
	} else if len(c.Interval) < 1 {
		ret.interval = defaultInterval
	} else if interval, e := time.ParseDuration(c.Interval); e != nil {
		err = append(err, fmt.Errorf("%sInterval='%s' parse error: %s", logPrefix, c.Interval, e))
	} else if allowZeroInterval && interval < 0 {
		err = append(err, fmt.Errorf("%sInterval='%s' must be >= 0", logPrefix, c.Interval))
	} else if !allowZeroInterval && interval <= 0 {
		err = append(err, fmt.Errorf("%sInterval='%s' must be > 0", logPrefix, c.Interval))
	} else {
		ret.interval = interval
	}

	if !allowRetain {
		if c.Retain != nil {
			err = append(err, fmt.Errorf("%sRetain not supported", logPrefix))
		}
	} else if c.Retain == nil {
		ret.retain = defaultRetain
	} else {
		ret.retain = *c.Retain
	}

	if c.Qos == nil {
		ret.qos = 1
	} else if *c.Qos == 0 || *c.Qos == 1 || *c.Qos == 2 {
		ret.qos = *c.Qos
	} else {
		err = append(err, fmt.Errorf("%sQos=%d but must be 0, 1 or 2", logPrefix, *c.Qos))
	}

	if !allowDevices {
		if len(c.Devices) > 0 {
			err = append(err, fmt.Errorf("%sDevices must not be set", logPrefix))
		}
	} else {
		if len(c.Devices) == 0 {
			// no devices given, default to all
			c.Devices = make(map[string]mqttDeviceSectionConfigRead, len(devices))
			for _, dev := range devices {
				c.Devices[dev.Name()] = mqttDeviceSectionConfigRead{}
			}
		}

		var e []error
		ret.devices, e = TransformAndValidateMapToList(
			c.Devices,
			func(inp mqttDeviceSectionConfigRead, name string) (MqttDeviceSectionConfig, []error) {
				return inp.TransformAndValidate(
					name,
					devices,
					fmt.Sprintf("%s%s->", logPrefix, name),
					allowFilter,
				)
			},
		)
		err = append(err, e...)
	}

	return
}

func (c mqttDeviceSectionConfigRead) TransformAndValidate(
	name string,
	devices []DeviceConfig,
	logPrefix string,
	allowFilter bool,
) (ret MqttDeviceSectionConfig, err []error) {
	ret = MqttDeviceSectionConfig{
		name: name,
	}

	if !existsByName(name, devices) {
		err = append(err, fmt.Errorf("%sDevices: device='%s' is not defined or is an MqttDevice", logPrefix, name))
	}

	if !allowFilter && c.Filter != nil {
		err = append(err, fmt.Errorf("%sFilter must not be set", logPrefix))
	}

	if c.Filter == nil {
		c.Filter = &filterConfigRead{}
	}
	var e []error
	ret.filter, e = c.Filter.TransformAndValidate()
	err = append(err, e...)

	return
}

func (c deviceConfigRead) TransformAndValidate(name string) (ret DeviceConfig, err []error) {
	ret = DeviceConfig{
		name: name,
	}

	if !nameMatcher.MatchString(ret.name) {
		err = append(err, fmt.Errorf("Devices->Name='%s' does not match %s", ret.name, NameRegexp))
	}

	var e []error
	ret.filter, e = c.Filter.TransformAndValidate()
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

func (c victronDeviceConfigRead) TransformAndValidate(name string) (ret VictronDeviceConfig, err []error) {
	ret = VictronDeviceConfig{
		kind:   types.VictronDeviceKindFromString(c.Kind),
		device: c.Device,
	}

	var e []error
	ret.DeviceConfig, e = c.deviceConfigRead.TransformAndValidate(name)
	err = append(err, e...)

	if ret.kind == types.VictronUndefinedKind {
		err = append(err, fmt.Errorf("VictronDevices->%s->Kind='%s' is invalid", name, c.Kind))
	}

	if ret.kind == types.VictronVedirectKind && len(c.Device) < 1 {
		err = append(err, fmt.Errorf("VictronDevices->%s->Device must not be empty", name))
	}

	return
}

func (c modbusDeviceConfigRead) TransformAndValidate(
	name string, modbus []ModbusConfig,
) (ret ModbusDeviceConfig, err []error) {
	ret = ModbusDeviceConfig{
		kind: types.ModbusDeviceKindFromString(c.Kind),
		bus:  c.Bus,
	}

	var e []error
	ret.DeviceConfig, e = c.deviceConfigRead.TransformAndValidate(name)
	err = append(err, e...)

	if ret.kind == types.ModbusUndefinedKind {
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

func (c httpDeviceConfigRead) TransformAndValidate(name string) (ret HttpDeviceConfig, err []error) {
	ret = HttpDeviceConfig{
		kind:     types.HttpDeviceKindFromString(c.Kind),
		username: c.Username,
		password: c.Password,
	}

	var e []error
	ret.DeviceConfig, e = c.deviceConfigRead.TransformAndValidate(name)
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

	if ret.kind == types.HttpUndefinedKind {
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
		kind:        types.MqttDeviceKindFromString(c.Kind),
		mqttClients: make([]string, 0, len(c.MqttClients)),
		mqttTopics:  c.MqttTopics,
	}

	if ret.kind == types.MqttDeviceUndefinedKind {
		err = append(err, fmt.Errorf("MqttDevices->%s->Kind='%s' is invalid", name, c.Kind))
	}

	if len(c.MqttClients) < 1 {
		err = append(err, fmt.Errorf("MqttDevices->%s->MqttClients: must not be empty", name))
	} else {
		for _, clientName := range c.MqttClients {
			if !existsByName(clientName, mqttClients) {
				err = append(err, fmt.Errorf("MqttDevices->%s->MqttClients: client='%s' is not defined", name, clientName))
			} else {
				ret.mqttClients = append(ret.mqttClients, clientName)
			}
		}
	}

	if len(c.MqttTopics) < 1 {
		err = append(err, fmt.Errorf("MqttDevices->%s->MqttTopics: must not be empty", name))
	}

	var e []error
	ret.DeviceConfig, e = c.deviceConfigRead.TransformAndValidate(name)
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

	if !existsByName(c.Name, devices) {
		err = append(err, fmt.Errorf("device='%s' is not defined", c.Name))
	}

	var e []error
	ret.filter, e = c.Filter.TransformAndValidate()
	err = append(err, e...)

	return
}

func (c filterConfigRead) TransformAndValidate() (ret FilterConfig, err []error) {
	ret = FilterConfig{
		includeRegisters:  c.IncludeRegisters,
		skipRegisters:     c.SkipRegisters,
		includeCategories: c.IncludeCategories,
		skipCategories:    c.SkipCategories,
	}

	if c.DefaultInclude == nil {
		ret.defaultInclude = true
	} else {
		ret.defaultInclude = *c.DefaultInclude
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
