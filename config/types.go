package config

import (
	"fmt"
	"net/url"
	"time"
)

type Config struct {
	version        int                 `yaml:"Version"`        // must be 0
	projectTitle   string              `yaml:"ProjectTitle"`   // optional: default go-iotdevice
	auth           AuthConfig          `yaml:"Auth"`           // optional: default Disabled
	mqttClients    []*MqttClientConfig `yaml:"MqttClients"`    // mandatory: at least 1 must be defined
	devices        []*DeviceConfig     `yaml:"Devices"`        // mandatory: at least 1 must be defined
	views          []*ViewConfig       `yaml:"Views"`          // mandatory: at least 1 must be defined
	httpServer     HttpServerConfig    `yaml:"HttpServer"`     // optional: default Disabled
	logConfig      bool                `yaml:"LogConfig"`      // optional: default False
	logWorkerStart bool                `yaml:"LogWorkerStart"` // optional: default False
	logDebug       bool                `yaml:"LogDebug"`       // optional: default False
}

type AuthConfig struct {
	enabled           bool          // defined automatically if Auth section exists
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
	clientId          string        // optional: default go-mqtt-to-influx-UUID
	qos               byte          // optional: default 1, must be 0, 1, 2
	keepAlive         time.Duration // optional: default 60s
	connectRetryDelay time.Duration // optional: default 10s
	connectTimeout    time.Duration // optional: default 5s
	availabilityTopic string        // optional: default %Prefix%tele/%ClientId%/status
	telemetryInterval time.Duration // optional: "10s"
	telemetryTopic    string        // optional: "%Prefix%tele/go-iotdevice/%DeviceName%/state"
	telemetryRetain   bool          // optional: default false
	realtimeEnable    bool          // optional: default false
	realtimeTopic     string        // optional: default "%Prefix%stat/go-iotdevice/%DeviceName%/%ValueName%"
	realtimeRetain    bool          // optional: default true
	topicPrefix       string        // optional: default empty
	logDebug          bool          // optional: default False
	logMessages       bool          // optional: default False
}

type DeviceConfig struct {
	name           string     // defined automatically by map key
	kind           DeviceKind // mandatory: what connection protocol is used
	device         string     // optional: the serial device eg. /dev/ttyVE0
	skipFields     []string   // optional: a list of fields that shall be ignored (Eg. Temperature when no sensor is connected)
	skipCategories []string   // optional: a list of categories that shall be ignored (Eg. Historic)
	logDebug       bool       // optional: default False
	logComDebug    bool       // optional: default False
}

type ViewDeviceConfig struct {
	name  string // defined automatically by map key
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

type HttpServerConfig struct {
	enabled         bool          // defined automatically if HttpServer section exists
	bind            string        // optional: defaults to ::1 (ipv6 loopback)
	port            int           // optional: defaults to 8000
	logRequests     bool          // optional: default False
	enableDocs      bool          // optional: default True
	frontendProxy   *url.URL      // optional: default deactivated; otherwise an address of the frontend dev-server
	frontendPath    string        // optional: default "frontend-build"; otherwise set to a path where the frontend build is located
	frontendExpires time.Duration // optional: default 5min; what cache-control header to sent for static frontend files
	configExpires   time.Duration // optional: default 1min; what cache-control header to sent for static frontend files
}

// device kind
type DeviceKind int

const (
	UndefinedKind DeviceKind = iota
	RandomBmvKind
	RandomSolarKind
	VedirectKind
)

func (dk DeviceKind) String() string {
	switch dk {
	case UndefinedKind:
		return "Undefined"
	case RandomBmvKind:
		return "RandomBmv"
	case RandomSolarKind:
		return "RandomSolar"
	case VedirectKind:
		return "Vedirect"
	default:
		return fmt.Sprintf("Kind%d", int(dk))
	}
}

func DeviceKindFromString(s string) DeviceKind {
	if s == "RandomBmv" {
		return RandomBmvKind
	}
	if s == "RandomSolar" {
		return RandomSolarKind
	}
	if s == "Vedirect" {
		return VedirectKind
	}
	return UndefinedKind
}
