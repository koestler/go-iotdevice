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
	name               string        // defined automatically by map key
	broker             string        // mandatory
	user               string        // optional: default empty
	password           string        // optional: default empty
	clientId           string        // optional: default go-iotdevice-UUID
	qos                byte          // optional: default 1, must be 0, 1, 2
	topicPrefix        string        // optional: ""
	availabilityEnable bool          // optional: default true
	availabilityTopic  string        // optional: default %Prefix%tele/%ClientId%/status
	telemetryInterval  time.Duration // optional: "10s"
	telemetryTopic     string        // optional: "%Prefix%tele/go-iotdevice/%DeviceName%/state"
	telemetryRetain    bool          // optional: default false
	realtimeEnable     bool          // optional: default false
	realtimeTopic      string        // optional: default "%Prefix%stat/go-iotdevice/%DeviceName%/%ValueName%"
	realtimeRetain     bool          // optional: default true
	logDebug           bool          // optional: default false
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

// Read structs are given to yaml for decoding and are slightly less exact in types
type configRead struct {
	Version        *int                    `yaml:"Version"`
	ProjectTitle   string                  `yaml:"ProjectTitle"`
	Auth           *authConfigRead         `yaml:"Auth"`
	MqttClients    mqttClientConfigReadMap `yaml:"MqttClients"`
	Devices        deviceConfigReadMap     `yaml:"Devices"`
	Views          viewConfigReadList      `yaml:"Views"`
	HttpServer     *httpServerConfigRead   `yaml:"HttpServer"`
	LogConfig      *bool                   `yaml:"LogConfig"`
	LogWorkerStart *bool                   `yaml:"LogWorkerStart"`
	LogDebug       *bool                   `yaml:"LogDebug"`
}

type authConfigRead struct {
	JwtSecret         *string `yaml:"JwtSecret"`
	JwtValidityPeriod string  `yaml:"JwtValidityPeriod"`
	HtaccessFile      *string `yaml:"HtaccessFile"`
}

type mqttClientConfigRead struct {
	Broker             string  `yaml:"Broker"`
	User               string  `yaml:"User"`
	Password           string  `yaml:"Password"`
	ClientId           string  `yaml:"ClientId"`
	Qos                *byte   `yaml:"Qos"`
	TopicPrefix        *string `yaml:"TopicPrefix"`
	AvailabilityEnable *bool   `yaml:"AvailabilityEnable"`
	AvailabilityTopic  *string `yaml:"AvailabilityTopic"`
	TelemetryInterval  string  `yaml:"TelemetryInterval"`
	TelemetryTopic     *string `yaml:"TelemetryTopic"`
	TelemetryRetain    *bool   `yaml:"TelemetryRetain"`
	RealtimeEnable     *bool   `yaml:"RealtimeEnable"`
	RealtimeTopic      *string `yaml:"RealtimeTopic"`
	RealtimeRetain     *bool   `yaml:"RealtimeRetain"`
	LogDebug           *bool   `yaml:"LogDebug"`
}

type mqttClientConfigReadMap map[string]mqttClientConfigRead

type deviceConfigRead struct {
	Kind           string   `yaml:"Kind"`
	Device         string   `yaml:"Device"`
	SkipFields     []string `yaml:"SkipFields"`
	SkipCategories []string `yaml:"SkipCategories"`
	LogDebug       *bool    `yaml:"LogDebug"`
	LogComDebug    *bool    `yaml:"LogComDebug"`
}

type deviceConfigReadMap map[string]deviceConfigRead

type viewDeviceConfigRead struct {
	Name  string `yaml:"Name"`
	Title string `yaml:"Title"`
}

type viewDeviceConfigReadList []viewDeviceConfigRead

type viewConfigRead struct {
	Name           string                   `yaml:"Name"`
	Title          string                   `yaml:"Title"`
	Devices        viewDeviceConfigReadList `yaml:"Devices"`
	Autoplay       *bool                    `yaml:"Autoplay"`
	AllowedUsers   []string                 `yaml:"AllowedUsers"`
	Hidden         *bool                    `yaml:"Hidden"`
	SkipFields     []string                 `yaml:"SkipFields"`
	SkipCategories []string                 `yaml:"SkipCategories"`
}

type viewConfigReadList []viewConfigRead

type httpServerConfigRead struct {
	Bind            string `yaml:"Bind"`
	Port            *int   `yaml:"Port"`
	LogRequests     *bool  `yaml:"LogRequests"`
	EnableDocs      *bool  `yaml:"EnableDocs"`
	FrontendProxy   string `yaml:"FrontendProxy"`
	FrontendPath    string `yaml:"FrontendPath"`
	FrontendExpires string `yaml:"FrontendExpires"`
	ConfigExpires   string `yaml:"ConfigExpires"`
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
