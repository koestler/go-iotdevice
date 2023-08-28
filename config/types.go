package config

import (
	"net/url"
	"regexp"
	"time"
)

type Nameable interface {
	Name() string
}

type Config struct {
	version         int                   // must be 1
	projectTitle    string                // optional: default go-iotdevice
	logConfig       bool                  // optional: default True
	logWorkerStart  bool                  // optional: default True
	logStorageDebug bool                  // optional: default False
	httpServer      HttpServerConfig      // optional: default Disabled
	authentication  AuthenticationConfig  // optional: default Disabled
	mqttClients     []MqttClientConfig    // mandatory: at least 1 must be defined
	modbus          []ModbusConfig        // optional: default empty
	devices         []DeviceConfig        // aggregated over all types
	victronDevices  []VictronDeviceConfig // optional: default empty
	modbusDevices   []ModbusDeviceConfig  // optional: default empty
	httpDevices     []HttpDeviceConfig    // optional: default empty
	mqttDevices     []MqttDeviceConfig    // optional: default empty
	views           []ViewConfig          // optional: default empty
	hassDiscovery   []HassDiscovery       // optional: default empty
}

type HttpServerConfig struct {
	enabled         bool          // defined automatically if HttpServer section exists
	bind            string        // optional: defaults to ::1 (ipv6 loopback)
	port            int           // optional: defaults to 8000
	logRequests     bool          // optional: default True
	frontendProxy   *url.URL      // optional: default deactivated; otherwise an address of the frontend dev-server
	frontendPath    string        // optional: default "frontend-build"; otherwise set to a path where the frontend build is located
	frontendExpires time.Duration // optional: default 5min; what cache-control header to send for static frontend files
	configExpires   time.Duration // optional: default 1min; what cache-control header to send for config endpoint
	logDebug        bool          // optional: default false
}

type AuthenticationConfig struct {
	enabled           bool          // defined automatically if Authentication section exists
	jwtSecret         []byte        // optional: default new random string on startup
	jwtValidityPeriod time.Duration // optional: default 1h
	htaccessFile      string        // optional: default no valid users
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

type ModbusConfig struct {
	name        string        // defined automatically by map key
	device      string        // mandatory: the serial device path eg. /dev/ttyUSB0
	baudRate    int           // mandatory: eg. 9600
	readTimeout time.Duration // optional: default 100ms
	logDebug    bool          // optional: default False
}

type DeviceConfig struct {
	name                      string        // defined automatically by map key
	skipFields                []string      // optional: a list of fields that shall be ignored (Eg. Temperature when no sensor is connected)
	skipCategories            []string      // optional: a list of categories that shall be ignored (Eg. Historic)
	telemetryViaMqttClients   []string      // optional: default empty
	realtimeViaMqttClients    []string      // optional: default empty
	restartInterval           time.Duration // optional: default 200ms
	restartIntervalMaxBackoff time.Duration // optional: default 1m
	logDebug                  bool          // optional: default False
	logComDebug               bool          // optional: default False
}

type VictronDeviceConfig struct {
	DeviceConfig
	device string            // mandatory: the serial device path eg. /dev/ttyUSB0
	kind   VictronDeviceKind // mandatory: what connection protocol is used
}

type ModbusDeviceConfig struct {
	DeviceConfig
	bus          string                 // mandatory: id of the modbus
	kind         ModbusDeviceKind       // mandatory: what connection protocol is used
	address      byte                   // mandatory: the modbus address of the device; format: 0x0A
	relays       map[string]RelayConfig // optional: custom labels for the relays
	pollInterval time.Duration          // optional: default 1s
}

type RelayConfig struct {
	description string // optional: default channel name
	openLabel   string // optional: default:"open"
	closedLabel string // optional: default "closed"
}

type HttpDeviceConfig struct {
	DeviceConfig
	url          *url.URL       // mandatory: how to connect to the device. eg. http://device0.local/
	kind         HttpDeviceKind // mandatory: what connection protocol is used
	username     string         // optional: username used to login
	password     string         // optional: password used to login
	pollInterval time.Duration  // optional: default 1s
}

type MqttDeviceConfig struct {
	DeviceConfig
	mqttTopics  []string // mandatory: at least 1 must be defined
	mqttClients []string
}

type ViewConfig struct {
	name         string              // mandatory: A technical name used in the URLs
	title        string              // mandatory: a nice title for the frontend
	devices      []ViewDeviceConfig  // mandatory: a list of deviceClient names
	autoplay     bool                // optional: default false
	allowedUsers map[string]struct{} // optional: if empty: view is public; otherwise only allowed to listed users
	hidden       bool                // optional: if true, view is not shown in menu unless logged in
}

type ViewDeviceConfig struct {
	name           string   // mandatory: a technical name
	title          string   // mandatory: a nice title for the frontend
	skipFields     []string // optional: a list of fields that are not shown
	skipCategories []string // optional: a list of categories that are not shown
}

type HassDiscovery struct {
	topicPrefix       string   // optional: default "homeassistant"
	viaMqttClients    []string // optional: default empty
	devices           []string
	categories        []string
	categoriesMatcher []*regexp.Regexp
	registers         []string
	registersMatcher  []*regexp.Regexp
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
