package config

import (
	"net/url"
	"time"
)

type Nameable interface {
	Name() string
}

type Config struct {
	version         int
	projectTitle    string
	logConfig       bool
	logWorkerStart  bool
	logStorageDebug bool
	httpServer      HttpServerConfig
	authentication  AuthenticationConfig
	mqttClients     []MqttClientConfig
	modbus          []ModbusConfig
	devices         []DeviceConfig
	victronDevices  []VictronDeviceConfig
	modbusDevices   []ModbusDeviceConfig
	httpDevices     []HttpDeviceConfig
	mqttDevices     []MqttDeviceConfig
	views           []ViewConfig
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

	availabilityClient     MqttSectionConfig
	availabilityDevice     MqttSectionConfig
	structure              MqttSectionConfig
	telemetry              MqttSectionConfig
	realtime               MqttSectionConfig
	homeassistantDiscovery MqttSectionConfig

	logDebug    bool
	logMessages bool
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
	name           string
	registerFilter RegisterFilterConfig
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
	registerFilter            RegisterFilterConfig
	restartInterval           time.Duration
	restartIntervalMaxBackoff time.Duration
	logDebug                  bool
	logComDebug               bool
}

type VictronDeviceConfig struct {
	DeviceConfig
	device string
	kind   VictronDeviceKind
}

type ModbusDeviceConfig struct {
	DeviceConfig
	bus          string
	kind         ModbusDeviceKind
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
	kind         HttpDeviceKind
	username     string
	password     string
	pollInterval time.Duration
}

type MqttDeviceConfig struct {
	DeviceConfig
	mqttTopics  []string
	mqttClients []string
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
	name           string
	title          string
	registerFilter RegisterFilterConfig
}

type RegisterFilterConfig struct {
	includeRegisters  []string
	skipRegisters     []string
	includeCategories []string
	skipCategories    []string
	defaultInclude    bool
}

type VictronDeviceKind int

const (
	VictronUndefinedKind VictronDeviceKind = iota
	VictronRandomBmvKind
	VictronRandomSolarKind
	VictronVedirectKind
)

type ModbusDeviceKind int

const (
	ModbusUndefinedKind ModbusDeviceKind = iota
	ModbusWaveshareRtuRelay8Kind
)

type HttpDeviceKind int

const (
	HttpUndefinedKind HttpDeviceKind = iota
	HttpTeracomKind
	HttpShellyEm3Kind
)
