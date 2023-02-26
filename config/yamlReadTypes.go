package config

type configRead struct {
	Version         *int                       `yaml:"Version"`
	ProjectTitle    string                     `yaml:"ProjectTitle"`
	Authentication  *authenticationConfigRead  `yaml:"Authentication"`
	MqttClients     mqttClientConfigReadMap    `yaml:"MqttClients"`
	VictronDevices  victronDeviceConfigReadMap `yaml:"VictronDevices"`
	HttpDevices     httpDeviceConfigReadMap    `yaml:"HttpDevices"`
	MqttDevices     mqttDeviceConfigReadMap    `yaml:"MqttDevices"`
	Views           viewConfigReadList         `yaml:"Views"`
	HttpServer      *httpServerConfigRead      `yaml:"HttpServer"`
	LogConfig       *bool                      `yaml:"LogConfig"`
	LogWorkerStart  *bool                      `yaml:"LogWorkerStart"`
	LogStorageDebug *bool                      `yaml:"LogStorageDebug"`
}

type authenticationConfigRead struct {
	JwtSecret         *string `yaml:"JwtSecret"`
	JwtValidityPeriod string  `yaml:"JwtValidityPeriod"`
	HtaccessFile      *string `yaml:"HtaccessFile"`
}

type mqttClientConfigRead struct {
	Broker            string  `yaml:"Broker"`
	ProtocolVersion   *int    `yaml:"ProtocolVersion"`
	User              string  `yaml:"User"`
	Password          string  `yaml:"Password"`
	ClientId          *string `yaml:"ClientId"`
	Qos               *byte   `yaml:"Qos"`
	KeepAlive         string  `yaml:"KeepAlive"`
	ConnectRetryDelay string  `yaml:"ConnectRetryDelay"`
	ConnectTimeout    string  `yaml:"ConnectTimeout"`
	AvailabilityTopic *string `yaml:"AvailabilityTopic"`
	TelemetryInterval string  `yaml:"TelemetryInterval"`
	TelemetryTopic    *string `yaml:"TelemetryTopic"`
	TelemetryRetain   *bool   `yaml:"TelemetryRetain"`
	RealtimeEnable    *bool   `yaml:"RealtimeEnable"`
	RealtimeTopic     *string `yaml:"RealtimeTopic"`
	RealtimeRetain    *bool   `yaml:"RealtimeRetain"`
	TopicPrefix       string  `yaml:"TopicPrefix"`
	LogDebug          *bool   `yaml:"LogDebug"`
	LogMessages       *bool   `yaml:"LogMessages"`
}

type mqttClientConfigReadMap map[string]mqttClientConfigRead

type deviceConfigRead struct {
	SkipFields              []string `yaml:"SkipFields"`
	SkipCategories          []string `yaml:"SkipCategories"`
	TelemetryViaMqttClients []string `yaml:"TelemetryViaMqttClients"`
	RealtimeViaMqttClients  []string `yaml:"RealtimeViaMqttClients"`
	LogDebug                *bool    `yaml:"LogDebug"`
	LogComDebug             *bool    `yaml:"LogComDebug"`
}

type victronDeviceConfigRead struct {
	General deviceConfigRead `yaml:"General"`
	Device  string           `yaml:"Device"`
	Kind    string           `yaml:"Kind"`
}

type victronDeviceConfigReadMap map[string]victronDeviceConfigRead

type httpDeviceConfigRead struct {
	General                deviceConfigRead `yaml:"General"`
	Url                    string           `yaml:"Url"`
	Kind                   string           `yaml:"Kind"`
	Username               string           `yaml:"Username"`
	Password               string           `yaml:"Password"`
	PollInterval           string           `yaml:"PollInterval"`
	PollIntervalMaxBackoff string           `yaml:"PollIntervalMaxBackoff"`
}

type httpDeviceConfigReadMap map[string]httpDeviceConfigRead

type mqttDeviceConfigRead struct {
	General     deviceConfigRead `yaml:"General"`
	MqttTopics  []string         `yaml:"MqttTopics"`
	MqttClients []string         `yaml:"MqttClients"`
}

type mqttDeviceConfigReadMap map[string]mqttDeviceConfigRead

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
	FrontendProxy   string `yaml:"FrontendProxy"`
	FrontendPath    string `yaml:"FrontendPath"`
	FrontendExpires string `yaml:"FrontendExpires"`
	ConfigExpires   string `yaml:"ConfigExpires"`
	LogDebug        *bool  `yaml:"LogDebug"`
}
