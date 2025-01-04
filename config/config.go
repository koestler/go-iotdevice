package config

import (
	"bytes"
	"cmp"
	"fmt"
	"github.com/google/uuid"
	"github.com/koestler/go-iotdevice/v3/types"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
	"log"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"
)

const NameRegexp = "^[a-zA-Z0-9\\-]{1,32}$"

var nameMatcher = regexp.MustCompile(NameRegexp)

const EncryptionKeyRegexp = "^[0-9a-f]{32}$"

var encryptionKeyMatcher = regexp.MustCompile(EncryptionKeyRegexp)

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

	ret.victronBleDevices, e = TransformAndValidateMapToList(
		c.VictronBleDevices,
		func(inp victronBleDeviceConfigRead, name string) (VictronBleDeviceConfig, []error) {
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

	ret.gpioDevices, e = TransformAndValidateMapToList(
		c.GpioDevices,
		func(inp gpioDeviceConfigRead, name string) (GpioDeviceConfig, []error) {
			return inp.TransformAndValidate(name)
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

	ret.mqttDevices, e = TransformAndValidateMapToList(
		c.MqttDevices,
		func(inp mqttDeviceConfigRead, name string) (MqttDeviceConfig, []error) {
			return inp.TransformAndValidate(name)
		},
	)
	err = append(err, e...)

	ret.devices = make([]DeviceConfig, 0,
		len(ret.victronDevices)+
			len(ret.victronBleDevices)+
			len(ret.modbusDevices)+
			len(ret.gpioDevices)+
			len(ret.httpDevices)+
			len(ret.mqttDevices)+
			len(c.GensetDevices),
	)
	for _, d := range ret.victronDevices {
		ret.devices = append(ret.devices, d.DeviceConfig)
	}
	for _, d := range ret.victronBleDevices {
		ret.devices = append(ret.devices, d.DeviceConfig)
	}
	for _, d := range ret.modbusDevices {
		ret.devices = append(ret.devices, d.DeviceConfig)
	}
	for _, d := range ret.gpioDevices {
		ret.devices = append(ret.devices, d.DeviceConfig)
	}
	for _, d := range ret.httpDevices {
		ret.devices = append(ret.devices, d.DeviceConfig)
	}
	for _, d := range ret.mqttDevices {
		ret.devices = append(ret.devices, d.DeviceConfig)
	}

	ret.gensetDevices, e = TransformAndValidateMapToList(
		c.GensetDevices,
		func(inp gensetDeviceConfigRead, name string) (GensetDeviceConfig, []error) {
			return inp.TransformAndValidate(name, ret.devices)
		},
	)
	err = append(err, e...)

	for _, d := range ret.gensetDevices {
		ret.devices = append(ret.devices, d.DeviceConfig)
	}

	ret.mqttClients, e = TransformAndValidateMapToList(
		c.MqttClients,
		func(inp mqttClientConfigRead, name string) (MqttClientConfig, []error) {
			return inp.TransformAndValidate(name, ret.devices, ret.mqttDevices)
		},
	)
	err = append(err, e...)

	{
		var viewsErr []error
		ret.views, viewsErr = TransformAndValidateListUnique(
			c.Views,
			func(inp viewConfigRead) (ViewConfig, []error) {
				return inp.TransformAndValidate(ret.devices)
			},
			func(needle ViewConfig, haystack []ViewConfig) (err []error) {
				if existsByName(needle.Name(), haystack) {
					err = append(err, fmt.Errorf("duplicate name='%s'", needle.Name()))
				}
				return err
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

func (c mqttClientConfigRead) TransformAndValidate(
	name string,
	devices []DeviceConfig,
	mqttDevices []MqttDeviceConfig,
) (ret MqttClientConfig, err []error) {
	ret = MqttClientConfig{
		name:     name,
		user:     c.User,
		password: c.Password,
	}

	errPrefix := fmt.Sprintf("MqttClients->%s", ret.name)

	if !nameMatcher.MatchString(ret.name) {
		err = append(err, fmt.Errorf("%s name does not match %s", errPrefix, NameRegexp))
	}

	if len(c.Broker) < 1 {
		err = append(err, fmt.Errorf("%s->Broker must not be empty", errPrefix))
	} else {
		if broker, e := url.ParseRequestURI(c.Broker); e != nil {
			err = append(err, fmt.Errorf("%s->Broker invalid url: %s", errPrefix, e))
		} else if broker == nil {
			err = append(err, fmt.Errorf("%s->Broker cannot parse broker", errPrefix))
		} else {
			ret.broker = broker
		}
	}

	if c.ProtocolVersion == nil {
		ret.protocolVersion = 5
	} else if *c.ProtocolVersion == 5 {
		ret.protocolVersion = *c.ProtocolVersion
	} else {
		err = append(err, fmt.Errorf("%s->Protocol=%d but must be 5 (3 is not supported anymore)", errPrefix, *c.ProtocolVersion))
	}

	if c.ClientId == nil {
		ret.clientId = "go-iotdevice-" + uuid.New().String()
	} else {
		ret.clientId = *c.ClientId
	}

	if len(c.KeepAlive) < 1 {
		ret.keepAlive = time.Minute
	} else if keepAlive, e := time.ParseDuration(c.KeepAlive); e != nil {
		err = append(err, fmt.Errorf("%s->KeepAlive='%s' parse error: %s",
			errPrefix, c.KeepAlive, e,
		))
	} else if keepAlive < time.Second {
		err = append(err, fmt.Errorf("%s->KeepAlive='%s' must be >=1s",
			errPrefix, c.KeepAlive,
		))
	} else if keepAlive%time.Second != 0 {
		err = append(err, fmt.Errorf("%s->KeepAlive='%s' must be a multiple of a second",
			errPrefix, c.KeepAlive,
		))
	} else {
		ret.keepAlive = keepAlive
	}

	if len(c.ConnectRetryDelay) < 1 {
		ret.connectRetryDelay = 10 * time.Second
	} else if connectRetryDelay, e := time.ParseDuration(c.ConnectRetryDelay); e != nil {
		err = append(err, fmt.Errorf("%s->ConnectRetryDelay='%s' parse error: %s",
			errPrefix, c.ConnectRetryDelay, e,
		))
	} else if connectRetryDelay < 100*time.Millisecond {
		err = append(err, fmt.Errorf("%s->ConnectRetryDelay='%s' must be >=100ms",
			errPrefix, c.ConnectRetryDelay,
		))
	} else {
		ret.connectRetryDelay = connectRetryDelay
	}

	if len(c.ConnectTimeout) < 1 {
		ret.connectTimeout = 5 * time.Second
	} else if connectTimeout, e := time.ParseDuration(c.ConnectTimeout); e != nil {
		err = append(err, fmt.Errorf("%s->ConnectTimeout='%s' parse error: %s",
			errPrefix, c.ConnectTimeout, e,
		))
	} else if connectTimeout < 100*time.Millisecond {
		err = append(err, fmt.Errorf("%s->ConnectTimeout='%s' must be >=100ms",
			errPrefix, c.ConnectTimeout,
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
	ret.mqttDevices, e = TransformAndValidateMapToList(
		c.MqttDevices,
		func(inp mqttClientDeviceConfigRead, name string) (MqttClientDeviceConfig, []error) {
			return inp.TransformAndValidate(
				name, fmt.Sprintf("%s->MqttDevices->", errPrefix),
				mqttDevices,
			)
		},
	)
	err = append(err, e...)

	nonLoopMqttDevices := make([]DeviceConfig, 0, len(devices)-len(ret.mqttDevices))
	for _, d := range devices {
		// do not allow to send mqtt messages for any de
		if existsByName(d.Name(), ret.mqttDevices) {
			continue
		}
		nonLoopMqttDevices = append(nonLoopMqttDevices, d)
	}

	ret.availabilityClient, e = c.AvailabilityClient.TransformAndValidate(
		fmt.Sprintf("%s->AvailabilityClient->", errPrefix),
		nonLoopMqttDevices,
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
		fmt.Sprintf("%s->AvailabilityDevice->", errPrefix),
		nonLoopMqttDevices,
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
		fmt.Sprintf("%s->Structure->", errPrefix),
		nonLoopMqttDevices,
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
		fmt.Sprintf("%s->Telemetry->", errPrefix),
		nonLoopMqttDevices,
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
		fmt.Sprintf("%s->Realtime->", errPrefix),
		nonLoopMqttDevices,
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
		fmt.Sprintf("%s->HomeassistantDiscovery->", errPrefix),
		nonLoopMqttDevices,
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
		fmt.Sprintf("%s->Command->", errPrefix),
		nonLoopMqttDevices,
		false, // command is not affected by read only
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

func (c mqttClientDeviceConfigRead) TransformAndValidate(
	name string,
	logPrefix string,
	mqttDevices []MqttDeviceConfig,
) (ret MqttClientDeviceConfig, err []error) {
	ret = MqttClientDeviceConfig{
		name:       name,
		mqttTopics: c.MqttTopics,
	}

	if !existsByName(name, mqttDevices) {
		err = append(err, fmt.Errorf("%s: MqttDevice='%s' is not defined", logPrefix, name))
	}

	if len(ret.mqttTopics) < 1 {
		err = append(err, fmt.Errorf("%s%s->MqttTopics must not be empty", logPrefix, name))
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

	if c.IoLog != nil {
		ret.ioLog = *c.IoLog
	}

	if len(c.PollInterval) < 1 {
		// use default 100ms
		ret.pollInterval = 500 * time.Millisecond
	} else if pollInterval, e := time.ParseDuration(c.PollInterval); e != nil {
		err = append(err, fmt.Errorf("VictronDevices->%s->PollInterval='%s' parse error: %s",
			name, c.PollInterval, e,
		))
	} else if pollInterval < time.Millisecond {
		err = append(err, fmt.Errorf("VictronDevices->%s->PollInterval='%s' must be >=1ms",
			name, c.PollInterval,
		))
	} else {
		ret.pollInterval = pollInterval
	}

	return
}

func (c victronBleDeviceConfigRead) TransformAndValidate(name string) (ret VictronBleDeviceConfig, err []error) {
	var e []error
	ret.DeviceConfig, e = c.deviceConfigRead.TransformAndValidate(name)
	err = append(err, e...)

	if len(c.AnnouncedName) < 1 {
		// use default name
		ret.announcedName = name
	} else {
		ret.announcedName = c.AnnouncedName
	}

	if !encryptionKeyMatcher.MatchString(c.EncryptionKey) {
		err = append(err, fmt.Errorf("VictronBleDevices->%s->EncryptionKey='%s' does not match %s",
			name, c.EncryptionKey, EncryptionKeyRegexp,
		))
	} else {
		ret.encryptionKey = c.EncryptionKey
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

	if strings.Contains(c.Address, "0x") {
		if n, e := fmt.Sscanf(c.Address, "0x%x", &ret.address); n != 1 || e != nil {
			err = append(err, fmt.Errorf("ModbusDevices->%s: hex Adress=%s is invalid: %s", name, c.Address, e))
		}
	} else {
		if n, e := fmt.Sscanf(c.Address, "%d", &ret.address); n != 1 || e != nil {
			err = append(err, fmt.Errorf("ModbusDevices->%s: decimal Adress=%s is invalid: %s", name, c.Address, e))
		}
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

func (c gpioDeviceConfigRead) TransformAndValidate(deviceName string) (ret GpioDeviceConfig, err []error) {
	var e []error
	ret.DeviceConfig, e = c.deviceConfigRead.TransformAndValidate(deviceName)
	err = append(err, e...)

	if c.Chip == nil {
		ret.chip = "gpiochip0"
	} else if len(*c.Chip) < 1 {
		err = append(err, fmt.Errorf("GpioDevices->%s->Chip must not be empty", deviceName))
	} else {
		ret.chip = *c.Chip
	}

	if len(c.InputDebounce) < 1 {
		// use default 100ms
		ret.inputDebounce = 100 * time.Millisecond
	} else if pollInterval, e := time.ParseDuration(c.InputDebounce); e != nil {
		err = append(err, fmt.Errorf("GpioDevices->%s->InputDebounce='%s' parse error: %s",
			deviceName, c.InputDebounce, e,
		))
	} else {
		ret.inputDebounce = pollInterval
	}

	ret.inputOptions = make([]string, 0)
	for _, opt := range c.InputOptions {
		switch opt {
		case "WithBiasDisabled", "WithPullDown", "WithPullUp":
			ret.inputOptions = append(ret.inputOptions, opt)
		default:
			err = append(err, fmt.Errorf("GpioDevices->%s->InputOptions='%s' is invalid", deviceName, opt))
		}
	}
	if len(ret.inputOptions) > 1 {
		err = append(err, fmt.Errorf("GpioDevices->%s->InputOptions must not contain more than one option", deviceName))
	}

	ret.outputOptions = make([]string, 0)
	for _, opt := range c.OutputOptions {
		switch opt {
		case "AsOpenDrain", "AsOpenSource", "AsPushPull":
			ret.outputOptions = append(ret.outputOptions, opt)
		default:
			err = append(err, fmt.Errorf("GpioDevices->%s->OutputOptions='%s' is invalid", deviceName, opt))
		}
	}
	if len(ret.outputOptions) > 1 {
		err = append(err, fmt.Errorf("GpioDevices->%s->OutputOptions must not contain more than one option", deviceName))
	}

	ret.inputs, e = TransformAndValidateMapToList(
		c.Inputs,
		func(inp pinConfigRead, name string) (PinConfig, []error) {
			return inp.TransformAndValidate(name, fmt.Sprintf("GpioDevices->%s->Inputs->%s", deviceName, name))
		},
	)
	err = append(err, e...)

	ret.outputs, e = TransformAndValidateMapToList(
		c.Outputs,
		func(inp pinConfigRead, name string) (PinConfig, []error) {
			return inp.TransformAndValidate(name, fmt.Sprintf("GpioDevices->%s->Outputs->%s", deviceName, name))
		},
	)
	err = append(err, e...)

	return
}

func (c pinConfigRead) TransformAndValidate(name, errPrefix string) (ret PinConfig, err []error) {
	ret = PinConfig{
		pin:         c.Pin,
		name:        name,
		description: name,
		lowLabel:    "low",
		highLabel:   "high",
	}

	if len(c.Pin) < 1 {
		err = append(err, fmt.Errorf("%s->Pin must not be empty", errPrefix))
	}

	if !nameMatcher.MatchString(ret.name) {
		err = append(err, fmt.Errorf("%s name '%s' does not match %s", errPrefix, ret.name, NameRegexp))
	}

	if c.Description != nil {
		ret.description = *c.Description
	}

	if c.LowLabel != nil {
		ret.lowLabel = *c.LowLabel
	}

	if c.HighLabel != nil {
		ret.highLabel = *c.HighLabel
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

func (c mqttDeviceConfigRead) TransformAndValidate(name string) (ret MqttDeviceConfig, err []error) {
	ret = MqttDeviceConfig{
		kind: types.MqttDeviceKindFromString(c.Kind),
	}

	if ret.kind == types.MqttDeviceUndefinedKind {
		err = append(err, fmt.Errorf("MqttDevices->%s->Kind='%s' is invalid", name, c.Kind))
	}

	var e []error
	ret.DeviceConfig, e = c.deviceConfigRead.TransformAndValidate(name)
	err = append(err, e...)

	return
}

func (c gensetDeviceConfigRead) TransformAndValidate(name string, devices []DeviceConfig) (ret GensetDeviceConfig, err []error) {
	var e []error
	ret.DeviceConfig, e = c.deviceConfigRead.TransformAndValidate(name)
	err = append(err, e...)

	ret.inputBindings, e = c.InputBindings.TransformAndValidate(devices)
	err = append(err, e...)

	ret.outputBindings, e = c.OutputBindings.TransformAndValidate(devices)
	err = append(err, e...)

	if len(c.PrimingTimeout) < 1 {
		// use default 10s
		ret.primingTimeout = 10 * time.Second
	} else if primingTimeout, e := time.ParseDuration(c.PrimingTimeout); e != nil {
		err = append(err, fmt.Errorf("GensetDevices->%s->PrimingTimeout='%s' parse error: %s",
			name, c.PrimingTimeout, e,
		))
	} else {
		ret.primingTimeout = primingTimeout
	}

	if len(c.CrankingTimeout) < 1 {
		// use default 10s
		ret.crankingTimeout = 10 * time.Second
	} else if crankingTimeout, e := time.ParseDuration(c.CrankingTimeout); e != nil {
		err = append(err, fmt.Errorf("GensetDevices->%s->CrankingTimeout='%s' parse error: %s",
			name, c.CrankingTimeout, e,
		))
	} else {
		ret.crankingTimeout = crankingTimeout
	}

	if len(c.WarmUpTimeout) < 1 {
		// use default 10m
		ret.warmUpTimeout = 10 * time.Minute
	} else if warmUpTimeout, e := time.ParseDuration(c.WarmUpTimeout); e != nil {
		err = append(err, fmt.Errorf("GensetDevices->%s->WarmUpTimeout='%s' parse error: %s",
			name, c.WarmUpTimeout, e,
		))
	} else {
		ret.warmUpTimeout = warmUpTimeout
	}

	if len(c.WarmUpMinTime) < 1 {
		// use default 2m
		ret.warmUpMinTime = 2 * time.Minute
	} else if warmUpMinTime, e := time.ParseDuration(c.WarmUpMinTime); e != nil {
		err = append(err, fmt.Errorf("GensetDevices->%s->WarmUpMinTime='%s' parse error: %s",
			name, c.WarmUpMinTime, e,
		))
	} else {
		ret.warmUpMinTime = warmUpMinTime
	}

	if ret.warmUpMinTime > ret.warmUpTimeout {
		err = append(err, fmt.Errorf("GensetDevices->%s->WarmUpMinTime='%s' must be less or equal than WarmUpTimeout='%s'",
			name, c.WarmUpMinTime, c.WarmUpTimeout,
		))
	}

	if c.WarmUpTemp == nil {
		ret.warmUpTemp = 50 // default 50°C
	} else {
		ret.warmUpTemp = *c.WarmUpTemp
	}

	if len(c.EngineCoolDownTimeout) < 1 {
		// use default 5m
		ret.engineCoolDownTimeout = 5 * time.Minute
	} else if engineCoolDownTimeout, e := time.ParseDuration(c.EngineCoolDownTimeout); e != nil {
		err = append(err, fmt.Errorf("GensetDevices->%s->EngineCoolDownTimeout='%s' parse error: %s",
			name, c.EngineCoolDownTimeout, e,
		))
	} else {
		ret.engineCoolDownTimeout = engineCoolDownTimeout
	}

	if len(c.EngineCoolDownMinTime) < 1 {
		// use default 2m
		ret.engineCoolDownMinTime = 2 * time.Minute
	} else if engineCoolDownMinTime, e := time.ParseDuration(c.EngineCoolDownMinTime); e != nil {
		err = append(err, fmt.Errorf("GensetDevices->%s->EngineCoolDownMinTime='%s' parse error: %s",
			name, c.EngineCoolDownMinTime, e,
		))
	} else {
		ret.engineCoolDownMinTime = engineCoolDownMinTime
	}

	if ret.engineCoolDownMinTime > ret.engineCoolDownTimeout {
		err = append(err, fmt.Errorf("GensetDevices->%s->EngineCoolDownMinTime='%s' must be less or equal than EngineCoolDownTimeout='%s'",
			name, c.EngineCoolDownMinTime, c.EngineCoolDownTimeout,
		))
	}

	if c.EngineCoolDownTemp == nil {
		ret.engineCoolDownTemp = 70 // default 70°C
	} else {
		ret.engineCoolDownTemp = *c.EngineCoolDownTemp
	}

	if len(c.EnclosureCoolDownTimeout) < 1 {
		// use default 10m
		ret.enclosureCoolDownTimeout = 10 * time.Minute
	} else if enclosureCoolDownTimeout, e := time.ParseDuration(c.EnclosureCoolDownTimeout); e != nil {
		err = append(err, fmt.Errorf("GensetDevices->%s->EnclosureCoolDownTimeout='%s' parse error: %s",
			name, c.EnclosureCoolDownTimeout, e,
		))
	} else {
		ret.enclosureCoolDownTimeout = enclosureCoolDownTimeout
	}

	if len(c.EnclosureCoolDownMinTime) < 1 {
		// use default 2m
		ret.enclosureCoolDownMinTime = 2 * time.Minute
	} else if enclosureCoolDownMinTime, e := time.ParseDuration(c.EnclosureCoolDownMinTime); e != nil {
		err = append(err, fmt.Errorf("GensetDevices->%s->EnclosureCoolDownMinTime='%s' parse error: %s",
			name, c.EnclosureCoolDownMinTime, e,
		))
	} else {
		ret.enclosureCoolDownMinTime = enclosureCoolDownMinTime
	}

	if ret.enclosureCoolDownMinTime > ret.enclosureCoolDownTimeout {
		err = append(err, fmt.Errorf("GensetDevices->%s->EnclosureCoolDownMinTime='%s' must be less or equal than EnclosureCoolDownTimeout='%s'",
			name, c.EnclosureCoolDownMinTime, c.EnclosureCoolDownTimeout,
		))
	}

	if c.EnclosureCoolDownTemp == nil {
		ret.enclosureCoolDownTemp = 30 // default 30°C
	} else {
		ret.enclosureCoolDownTemp = *c.EnclosureCoolDownTemp
	}

	if c.EngineTempMin == nil {
		ret.engineTempMin = -20 // default -20°C
	} else {
		ret.engineTempMin = *c.EngineTempMin
	}

	if c.EngineTempMax == nil {
		ret.engineTempMax = 90 // default 90°C
	} else {
		ret.engineTempMax = *c.EngineTempMax
	}

	if c.AuxTemp0Min == nil {
		ret.auxTemp0Min = -20 // default -20°C
	} else {
		ret.auxTemp0Min = *c.AuxTemp0Min
	}

	if c.AuxTemp0Max == nil {
		ret.auxTemp0Max = 120 // default 120°C
	} else {
		ret.auxTemp0Max = *c.AuxTemp0Max
	}

	if c.AuxTemp1Min == nil {
		ret.auxTemp1Min = -20 // default -20°C
	} else {
		ret.auxTemp1Min = *c.AuxTemp1Min
	}

	if c.AuxTemp1Max == nil {
		ret.auxTemp1Max = 120 // default 120°C
	} else {
		ret.auxTemp1Max = *c.AuxTemp1Max
	}

	if c.SinglePhase != nil {
		ret.singlePhase = *c.SinglePhase
	}

	if c.UMin == nil {
		ret.uMin = 220 // default 220V
	} else if *c.UMin < 0 {
		err = append(err, fmt.Errorf("GensetDevices->%s->UMin='%f' must be >=0", name, *c.UMin))
	} else {
		ret.uMin = *c.UMin
	}

	if c.UMax == nil {
		ret.uMax = 240 // default 240V
	} else if *c.UMax < 0 {
		err = append(err, fmt.Errorf("GensetDevices->%s->UMax='%f' must be >=0", name, *c.UMax))
	} else {
		ret.uMax = *c.UMax
	}

	if c.FMin == nil {
		ret.fMin = 45 // default 45Hz
	} else if *c.FMin < 0 {
		err = append(err, fmt.Errorf("GensetDevices->%s->FMin='%f' must be >=0", name, *c.FMin))
	} else {
		ret.fMin = *c.FMin
	}

	if c.FMax == nil {
		ret.fMax = 55 // default 55Hz
	} else if *c.FMax < 0 {
		err = append(err, fmt.Errorf("GensetDevices->%s->FMax='%f' must be >=0", name, *c.FMax))
	} else {
		ret.fMax = *c.FMax
	}

	if c.PMax == nil {
		ret.pMax = 1000000 // default 1MW
	} else if *c.PMax < 0 {
		err = append(err, fmt.Errorf("GensetDevices->%s->PMax='%f' must be >=0", name, *c.PMax))
	} else {
		ret.pMax = *c.PMax
	}

	if c.PTotMax == nil {
		ret.pTotMax = 1000000 // default 1MW
	} else if *c.PTotMax < 0 {
		err = append(err, fmt.Errorf("GensetDevices->%s->PTotMax='%f' must be >=0", name, *c.PTotMax))
	} else {
		ret.pTotMax = *c.PTotMax
	}

	return
}

func (c gensetDeviceBindingConfigRead) TransformAndValidate(devices []DeviceConfig) (ret []GensetDeviceBindingConfig, err []error) {
	for deviceName, m := range c {
		if !existsByName(deviceName, devices) {
			err = append(err, fmt.Errorf("device='%s' is not defined", deviceName))
		}

		for registerName, name := range m {
			ret = append(ret, GensetDeviceBindingConfig{
				deviceName:   deviceName,
				registerName: registerName,
				name:         name,
			})
		}
	}
	slices.SortFunc(ret, func(i, j GensetDeviceBindingConfig) int {
		return cmp.Or(
			cmp.Compare(i.deviceName, j.deviceName),
			cmp.Compare(i.registerName, j.registerName),
			cmp.Compare(i.name, j.name),
		)
	})

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
			func(needle ViewDeviceConfig, haystack []ViewDeviceConfig) (err []error) {
				if existsByName(needle.Name(), haystack) {
					err = append(err, fmt.Errorf("duplicate name='%s'", needle.Name()))
				}
				return err
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
	slices.Sort(keys)

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
	uniqueErr func(needle O, haystack []O) []error,
) (ret []O, err []error) {
	ret = make([]O, 0, len(inp))
	for _, cr := range inp {
		r, e := transformer(cr)

		if e := uniqueErr(r, ret); e != nil {
			err = append(err, e...)
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
