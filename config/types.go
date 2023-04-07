package config

import (
	"net/url"
	"time"
)

type Config struct {
	version         int                    // must be 1
	projectTitle    string                 // optional: default go-iotdevice
	logConfig       bool                   // optional: default False
	logWorkerStart  bool                   // optional: default False
	logStorageDebug bool                   // optional: default False
	httpServer      HttpServerConfig       // optional: default Disabled
	authentication  AuthenticationConfig   // optional: default Disabled
	mqttClients     []*MqttClientConfig    // mandatory: at least 1 must be defined
	devices         []*DeviceConfig        // aggregated over all types
	victronDevices  []*VictronDeviceConfig // optional: default empty
	modbusDevices   []*ModbusDeviceConfig  // optional: default empty
	httpDevices     []*HttpDeviceConfig    // optional: default empty
	mqttDevices     []*MqttDeviceConfig    // optional: default empty
	views           []*ViewConfig          // optional: default empty
}

type HttpServerConfig struct {
	enabled         bool          // defined automatically if HttpServer section exists
	bind            string        // optional: defaults to ::1 (ipv6 loopback)
	port            int           // optional: defaults to 8000
	logRequests     bool          // optional: default False
	frontendProxy   *url.URL      // optional: default deactivated; otherwise an address of the frontend dev-server
	frontendPath    string        // optional: default "frontend-build"; otherwise set to a path where the frontend build is located
	frontendExpires time.Duration // optional: default 5min; what cache-control header to send for static frontend files
	configExpires   time.Duration // optional: default 1min; what cache-control header to send for config endpoint
	logDebug        bool          // optional: default false
}

type AuthenticationConfig struct {
	enabled           bool          // defined automatically if Authentication section exists
	jwtSecret         []byte        `yaml:"JwtSecret"`         // optional: default new random string on startup
	jwtValidityPeriod time.Duration `yaml:"JwtValidityPeriod"` // optional: default 1h
	htaccessFile      string        `yaml:"HtaccessFile"`      // optional: default no valid users
}

type MqttClientConfig struct {
	name              string        // defined automatically by map key
	broker            *url.URL      // mandatory
	protocolVersion   int           // optional: default 5
	user              string        // optional: default empty
	password          string        // optional: default empty
	clientId          string        // optional: default go-iotdevice-UUID
	qos               byte          // optional: default 1, must be 0, 1, 2
	keepAlive         time.Duration // optional: default 60s
	connectRetryDelay time.Duration // optional: default 10s
	connectTimeout    time.Duration // optional: default 5s
	availabilityTopic string        // optional: default %Prefix%tele/%ClientId%/status
	telemetryInterval time.Duration // optional: "10s"
	telemetryTopic    string        // optional: "%Prefix%tele/go-iotdevice/%DeviceName%/state"
	telemetryRetain   bool          // optional: default false
	realtimeEnable    bool          // default: false
	realtimeTopic     string        // optional: default "%Prefix%stat/go-iotdevice/%DeviceName%/%ValueName%"
	realtimeRetain    bool          // optional: default true
	topicPrefix       string        // optional: default empty
	logDebug          bool          // optional: default False
	logMessages       bool          // optional: default False
}

type DeviceConfig struct {
	name                    string   // defined automatically by map key
	skipFields              []string // optional: a list of fields that shall be ignored (Eg. Temperature when no sensor is connected)
	skipCategories          []string // optional: a list of categories that shall be ignored (Eg. Historic)
	telemetryViaMqttClients []string // optional: default empty
	realtimeViaMqttClients  []string // optional: default empty
	logDebug                bool     // optional: default False
	logComDebug             bool     // optional: default False
}

type VictronDeviceConfig struct {
	DeviceConfig
	device string            // mandatory: the serial device path eg. /dev/ttyUSB0
	kind   VictronDeviceKind // mandatory: what connection protocol is used
}

type ModbusDeviceConfig struct {
	DeviceConfig
	device  string           // mandatory: the serial device path eg. /dev/ttyUSB0
	kind    ModbusDeviceKind // mandatory: what connection protocol is used
	address uint             // mandatory: the modbus address of the device; format: 0x0A
}

type HttpDeviceConfig struct {
	DeviceConfig
	url                    *url.URL       // mandatory: how to connect to the device. eg. http://device0.local/
	kind                   HttpDeviceKind // mandatory: what connection protocol is used
	username               string         // optional: username used to login
	password               string         // optional: password used to login
	pollInterval           time.Duration  // optional: default 1s
	pollIntervalMaxBackoff time.Duration  // optional: default 10s
}

type MqttDeviceConfig struct {
	DeviceConfig
	mqttTopics  []string // mandatory: at least 1 must be defined
	mqttClients []string
}

type ViewDeviceConfig struct {
	name  string // mandatory: a technical name
	title string // mandatory: a nice title for the frontend
}

type ViewConfig struct {
	name           string              // mandatory: A technical name used in the URLs
	title          string              // mandatory: a nice title for the frontend
	devices        []*ViewDeviceConfig // mandatory: a list of deviceClient names
	autoplay       bool                // optional: default false
	allowedUsers   map[string]struct{} // optional: if empty: view is public; otherwise only allowed to listed users
	hidden         bool                // optional: if true, view is not shown in menu unless logged in
	skipFields     []string            // optional: a list of fields that are not shown
	skipCategories []string            // optional: a list of categories that are not shown
}

// victron device kind
type VictronDeviceKind int

const (
	VictronUndefinedKind VictronDeviceKind = iota
	VictronRandomBmvKind
	VictronRandomSolarKind
	VictronVedirectKind
)

func (dk VictronDeviceKind) String() string {
	switch dk {
	case VictronRandomBmvKind:
		return "RandomBmv"
	case VictronRandomSolarKind:
		return "RandomSolar"
	case VictronVedirectKind:
		return "Vedirect"
	default:
		return "Undefined"
	}
}

func VictronDeviceKindFromString(s string) VictronDeviceKind {
	if s == "RandomBmv" {
		return VictronRandomBmvKind
	}
	if s == "RandomSolar" {
		return VictronRandomSolarKind
	}
	if s == "Vedirect" {
		return VictronVedirectKind
	}
	return VictronUndefinedKind
}

// Modbus device kind
type ModbusDeviceKind int

const (
	ModbusUndefinedKind ModbusDeviceKind = iota
	ModbusWaveshareRtuRelay8Kind
)

func (dk ModbusDeviceKind) String() string {
	switch dk {
	case ModbusWaveshareRtuRelay8Kind:
		return "WaveshareRtuRelay8"
	default:
		return "Undefined"
	}
}

func ModbusDeviceKindFromString(s string) ModbusDeviceKind {
	if s == "WaveshareRtuRelay8" {
		return ModbusWaveshareRtuRelay8Kind
	}

	return ModbusUndefinedKind
}

// http device kind
type HttpDeviceKind int

const (
	HttpUndefinedKind HttpDeviceKind = iota
	HttpTeracomKind
	HttpShellyEm3Kind
)

func (dk HttpDeviceKind) String() string {
	switch dk {
	case HttpTeracomKind:
		return "Teracom"
	case HttpShellyEm3Kind:
		return "Shelly3m"
	default:
		return "Undefined"
	}
}

func HttpDeviceKindFromString(s string) HttpDeviceKind {
	if s == "Teracom" {
		return HttpTeracomKind
	}
	if s == "ShellyEm3" {
		return HttpShellyEm3Kind
	}

	return HttpUndefinedKind
}
