package config

type configRead struct {
	Version         *int                               `yaml:"Version"`
	ProjectTitle    string                             `yaml:"ProjectTitle"`
	LogConfig       *bool                              `yaml:"LogConfig"`
	LogWorkerStart  *bool                              `yaml:"LogWorkerStart"`
	LogStorageDebug *bool                              `yaml:"LogStorageDebug"`
	HttpServer      *httpServerConfigRead              `yaml:"HttpServer"`
	Authentication  *authenticationConfigRead          `yaml:"Authentication"`
	MqttClients     map[string]mqttClientConfigRead    `yaml:"MqttClients"`
	Modbus          map[string]modbusConfigRead        `yaml:"Modbus"`
	VictronDevices  map[string]victronDeviceConfigRead `yaml:"VictronDevices"`
	ModbusDevices   map[string]modbusDeviceConfigRead  `yaml:"ModbusDevices"`
	HttpDevices     map[string]httpDeviceConfigRead    `yaml:"HttpDevices"`
	MqttDevices     map[string]mqttDeviceConfigRead    `yaml:"MqttDevices"`
	Views           []viewConfigRead                   `yaml:"Views"`
}

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

type modbusConfigRead struct {
	Device      string `yaml:"Device"`
	BaudRate    int    `yaml:"BaudRate"`
	ReadTimeout string `yaml:"ReadTimeout"`
	LogDebug    *bool  `yaml:"LogDebug"`
}

type deviceConfigRead struct {
	SkipFields                []string `yaml:"SkipFields"`
	SkipCategories            []string `yaml:"SkipCategories"`
	TelemetryViaMqttClients   []string `yaml:"TelemetryViaMqttClients"`
	RealtimeViaMqttClients    []string `yaml:"RealtimeViaMqttClients"`
	RestartInterval           string   `yaml:"RestartInterval"`
	RestartIntervalMaxBackoff string   `yaml:"RestartIntervalMaxBackoff"`
	LogDebug                  *bool    `yaml:"LogDebug"`
	LogComDebug               *bool    `yaml:"LogComDebug"`
}

type victronDeviceConfigRead struct {
	General deviceConfigRead `yaml:"General"`
	Device  string           `yaml:"Device"`
	Kind    string           `yaml:"Kind"`
}

type modbusDeviceConfigRead struct {
	General      deviceConfigRead           `yaml:"General"`
	Bus          string                     `yaml:"Bus"`
	Kind         string                     `yaml:"Kind"`
	Address      string                     `yaml:"Address"`
	Relays       map[string]relayConfigRead `yaml:"Relays"`
	PollInterval string                     `yaml:"PollInterval"`
}

type relayConfigRead struct {
	Description *string `yaml:"Description"`
	OpenLabel   *string `yaml:"OpenLabel"`
	ClosedLabel *string `yaml:"ClosedLabel"`
}

type httpDeviceConfigRead struct {
	General      deviceConfigRead `yaml:"General"`
	Url          string           `yaml:"Url"`
	Kind         string           `yaml:"Kind"`
	Username     string           `yaml:"Username"`
	Password     string           `yaml:"Password"`
	PollInterval string           `yaml:"PollInterval"`
}

type mqttDeviceConfigRead struct {
	General     deviceConfigRead `yaml:"General"`
	MqttTopics  []string         `yaml:"MqttTopics"`
	MqttClients []string         `yaml:"MqttClients"`
}

type viewDeviceConfigRead struct {
	Name           string   `yaml:"Name"`
	Title          string   `yaml:"Title"`
	SkipFields     []string `yaml:"SkipFields"`
	SkipCategories []string `yaml:"SkipCategories"`
}

type viewConfigRead struct {
	Name         string                 `yaml:"Name"`
	Title        string                 `yaml:"Title"`
	Devices      []viewDeviceConfigRead `yaml:"Devices"`
	Autoplay     *bool                  `yaml:"Autoplay"`
	AllowedUsers []string               `yaml:"AllowedUsers"`
	Hidden       *bool                  `yaml:"Hidden"`
}
