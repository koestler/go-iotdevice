package config

type configRead struct {
	Version                *int                               `yaml:"Version"`
	ProjectTitle           string                             `yaml:"ProjectTitle"`
	LogConfig              *bool                              `yaml:"LogConfig"`
	LogWorkerStart         *bool                              `yaml:"LogWorkerStart"`
	LogStateStorageDebug   *bool                              `yaml:"LogStateStorageDebug"`
	LogCommandStorageDebug *bool                              `yaml:"LogCommandStorageDebug"`
	HttpServer             *httpServerConfigRead              `yaml:"HttpServer"`
	Authentication         *authenticationConfigRead          `yaml:"Authentication"`
	MqttClients            map[string]mqttClientConfigRead    `yaml:"MqttClients"`
	Modbus                 map[string]modbusConfigRead        `yaml:"Modbus"`
	VictronDevices         map[string]victronDeviceConfigRead `yaml:"VictronDevices"`
	ModbusDevices          map[string]modbusDeviceConfigRead  `yaml:"ModbusDevices"`
	HttpDevices            map[string]httpDeviceConfigRead    `yaml:"HttpDevices"`
	MqttDevices            map[string]mqttDeviceConfigRead    `yaml:"MqttDevices"`
	Views                  []viewConfigRead                   `yaml:"Views"`
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
	Broker          string `yaml:"Broker"`
	ProtocolVersion *int   `yaml:"ProtocolVersion"`

	User     string  `yaml:"User"`
	Password string  `yaml:"Password"`
	ClientId *string `yaml:"ClientId"`

	KeepAlive         string  `yaml:"KeepAlive"`
	ConnectRetryDelay string  `yaml:"ConnectRetryDelay"`
	ConnectTimeout    string  `yaml:"ConnectTimeout"`
	TopicPrefix       *string `yaml:"TopicPrefix"`
	ReadOnly          *bool   `yaml:"ReadOnly"`
	MaxBacklogSize    *int    `yaml:"MaxBacklogSize"`

	MqttDevices map[string]mqttClientDeviceConfigRead `yaml:"MqttDevices"`

	AvailabilityClient     mqttSectionConfigRead `yaml:"AvailabilityClient"`
	AvailabilityDevice     mqttSectionConfigRead `yaml:"AvailabilityDevice"`
	Structure              mqttSectionConfigRead `yaml:"Structure"`
	Telemetry              mqttSectionConfigRead `yaml:"Telemetry"`
	Realtime               mqttSectionConfigRead `yaml:"Realtime"`
	HomeassistantDiscovery mqttSectionConfigRead `yaml:"HomeassistantDiscovery"`
	Command                mqttSectionConfigRead `yaml:"Command"`

	LogDebug    *bool `yaml:"LogDebug"`
	LogMessages *bool `yaml:"LogMessages"`
}

type mqttClientDeviceConfigRead struct {
	MqttTopics []string `yaml:"MqttTopics"`
}

type mqttSectionConfigRead struct {
	Enabled       *bool                                  `yaml:"Enabled"`
	TopicTemplate *string                                `yaml:"TopicTemplate"`
	Interval      string                                 `yaml:"Interval"`
	Retain        *bool                                  `yaml:"Retain"`
	Qos           *byte                                  `yaml:"Qos"`
	Devices       map[string]mqttDeviceSectionConfigRead `yaml:"Devices"`
}

type mqttDeviceSectionConfigRead struct {
	Filter *filterConfigRead `yaml:"Filter"`
}

type modbusConfigRead struct {
	Device      string `yaml:"Device"`
	BaudRate    int    `yaml:"BaudRate"`
	ReadTimeout string `yaml:"ReadTimeout"`
	LogDebug    *bool  `yaml:"LogDebug"`
}

type deviceConfigRead struct {
	Filter                    filterConfigRead `yaml:"Filter"`
	RestartInterval           string           `yaml:"RestartInterval"`
	RestartIntervalMaxBackoff string           `yaml:"RestartIntervalMaxBackoff"`
	LogDebug                  *bool            `yaml:"LogDebug"`
	LogComDebug               *bool            `yaml:"LogComDebug"`
}

type victronDeviceConfigRead struct {
	deviceConfigRead `yaml:",inline"`
	Device           string `yaml:"Device"`
	Kind             string `yaml:"Kind"`
}

type modbusDeviceConfigRead struct {
	deviceConfigRead `yaml:",inline"`
	Bus              string                     `yaml:"Bus"`
	Kind             string                     `yaml:"Kind"`
	Address          string                     `yaml:"Address"`
	Relays           map[string]relayConfigRead `yaml:"Relays"`
	PollInterval     string                     `yaml:"PollInterval"`
}

type relayConfigRead struct {
	Description *string `yaml:"Description"`
	OpenLabel   *string `yaml:"OpenLabel"`
	ClosedLabel *string `yaml:"ClosedLabel"`
}

type httpDeviceConfigRead struct {
	deviceConfigRead `yaml:",inline"`
	Url              string `yaml:"Url"`
	Kind             string `yaml:"Kind"`
	Username         string `yaml:"Username"`
	Password         string `yaml:"Password"`
	PollInterval     string `yaml:"PollInterval"`
}

type mqttDeviceConfigRead struct {
	deviceConfigRead `yaml:",inline"`
	Kind             string `yaml:"Kind"`
}

type viewConfigRead struct {
	Name         string                 `yaml:"Name"`
	Title        string                 `yaml:"Title"`
	Devices      []viewDeviceConfigRead `yaml:"Devices"`
	Autoplay     *bool                  `yaml:"Autoplay"`
	AllowedUsers []string               `yaml:"AllowedUsers"`
	Hidden       *bool                  `yaml:"Hidden"`
}

type viewDeviceConfigRead struct {
	Name   string           `yaml:"Name"`
	Title  string           `yaml:"Title"`
	Filter filterConfigRead `yaml:"Filter"`
}

type filterConfigRead struct {
	IncludeRegisters  []string `yaml:"IncludeRegisters"`
	SkipRegisters     []string `yaml:"SkipRegisters"`
	IncludeCategories []string `yaml:"IncludeCategories"`
	SkipCategories    []string `yaml:"SkipCategories"`
	DefaultInclude    *bool    `yaml:"DefaultInclude"`
}
