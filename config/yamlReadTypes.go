package config

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
