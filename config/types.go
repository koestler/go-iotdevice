package config

import (
	"github.com/koestler/go-iotdevice/v3/types"
	"net/url"
	"time"
)

type Nameable interface {
	Name() string
}

type Config struct {
	version                int
	projectTitle           string
	logConfig              bool
	logWorkerStart         bool
	logStateStorageDebug   bool
	logCommandStorageDebug bool
	httpServer             HttpServerConfig
	authentication         AuthenticationConfig
	mqttClients            []MqttClientConfig
	modbus                 []ModbusConfig
	devices                []DeviceConfig
	victronDevices         []VictronDeviceConfig
	modbusDevices          []ModbusDeviceConfig
	httpDevices            []HttpDeviceConfig
	mqttDevices            []MqttDeviceConfig
	gensetDevices          []GensetDeviceConfig
	views                  []ViewConfig
}

type HttpServerConfig struct {
	enabled         bool
	bind            string
	port            int
	logRequests     bool
	frontendProxy   *url.URL
	frontendPath    string
	frontendExpires time.Duration
	configExpires   time.Duration
	logDebug        bool
}

type AuthenticationConfig struct {
	enabled           bool
	jwtSecret         []byte
	jwtValidityPeriod time.Duration
	htaccessFile      string
}

type MqttClientConfig struct {
	name            string
	broker          *url.URL
	protocolVersion int

	user     string
	password string
	clientId string

	keepAlive         time.Duration
	connectRetryDelay time.Duration
	connectTimeout    time.Duration
	topicPrefix       string
	readOnly          bool
	maxBacklogSize    int

	mqttDevices []MqttClientDeviceConfig

	availabilityClient     MqttSectionConfig
	availabilityDevice     MqttSectionConfig
	structure              MqttSectionConfig
	telemetry              MqttSectionConfig
	realtime               MqttSectionConfig
	homeassistantDiscovery MqttSectionConfig
	command                MqttSectionConfig

	logDebug    bool
	logMessages bool
}

type MqttClientDeviceConfig struct {
	name       string
	mqttTopics []string
}

type MqttSectionConfig struct {
	enabled       bool
	topicTemplate string
	interval      time.Duration
	retain        bool
	qos           byte
	devices       []MqttDeviceSectionConfig
}

type MqttDeviceSectionConfig struct {
	name   string
	filter FilterConfig
}

type ModbusConfig struct {
	name        string
	device      string
	baudRate    int
	readTimeout time.Duration
	logDebug    bool
}

type DeviceConfig struct {
	name                      string
	filter                    FilterConfig
	restartInterval           time.Duration
	restartIntervalMaxBackoff time.Duration
	logDebug                  bool
	logComDebug               bool
}

type VictronDeviceConfig struct {
	DeviceConfig
	device       string
	kind         types.VictronDeviceKind
	pollInterval time.Duration
	ioLog        string
}

type ModbusDeviceConfig struct {
	DeviceConfig
	bus          string
	kind         types.ModbusDeviceKind
	address      byte
	relays       map[string]RelayConfig
	pollInterval time.Duration
}

type RelayConfig struct {
	description string
	openLabel   string
	closedLabel string
}

type HttpDeviceConfig struct {
	DeviceConfig
	url          *url.URL
	kind         types.HttpDeviceKind
	username     string
	password     string
	pollInterval time.Duration
}

type MqttDeviceConfig struct {
	DeviceConfig
	kind types.MqttDeviceKind
}

type GensetDeviceConfig struct {
	DeviceConfig

	inputBindings  GensetDeviceBindingsConfig
	outputBindings GensetDeviceBindingsConfig

	primingTimeout           time.Duration
	crankingTimeout          time.Duration
	warmUpTimeout            time.Duration
	warmUpMinTime            time.Duration
	warmUpTemp               float32
	engineCoolDownTimeout    time.Duration
	engineCoolDownMinTime    time.Duration
	engineCoolDownTemp       float32
	enclosureCoolDownTimeout time.Duration
	enclosureCoolDownMinTime time.Duration
	enclosureCoolDownTemp    float32

	engineTempMin float32
	engineTempMax float32
	auxTemp0Min   float32
	auxTemp0Max   float32
	auxTemp1Min   float32
	auxTemp1Max   float32

	singlePhase bool
	uMin        float32
	uMax        float32
	fMin        float32
	fMax        float32
	pMax        float32
	pTotMax     float32
}

type GensetDeviceBindingsConfig []GensetDeviceBindingConfig

type GensetDeviceBindingConfig struct {
	name         string
	deviceName   string
	registerName string
}

type ViewConfig struct {
	name         string
	title        string
	devices      []ViewDeviceConfig
	autoplay     bool
	allowedUsers map[string]struct{}
	hidden       bool
}

type ViewDeviceConfig struct {
	name   string
	title  string
	filter FilterConfig
}

type FilterConfig struct {
	includeRegisters  []string
	skipRegisters     []string
	includeCategories []string
	skipCategories    []string
	defaultInclude    bool
}
