package config

import (
	"net/url"
	"regexp"
	"time"
)

// Getters for Config struct

func (c Config) Version() int {
	return c.version
}

func (c Config) ProjectTitle() string {
	return c.projectTitle
}

func (c Config) LogConfig() bool {
	return c.logConfig
}

func (c Config) LogWorkerStart() bool {
	return c.logWorkerStart
}

func (c Config) LogStorageDebug() bool {
	return c.logStorageDebug
}

func (c Config) HttpServer() HttpServerConfig {
	return c.httpServer
}

func (c Config) Authentication() AuthenticationConfig {
	return c.authentication
}

func (c Config) MqttClients() []*MqttClientConfig {
	return c.mqttClients
}

func (c Config) Modbus() []*ModbusConfig {
	return c.modbus
}

func (c Config) Devices() []*DeviceConfig {
	return c.devices
}

func (c Config) VictronDevices() []*VictronDeviceConfig {
	return c.victronDevices
}

func (c Config) ModbusDevices() []*ModbusDeviceConfig {
	return c.modbusDevices
}

func (c Config) HttpDevices() []*HttpDeviceConfig {
	return c.httpDevices
}

func (c Config) MqttDevices() []*MqttDeviceConfig {
	return c.mqttDevices
}

func (c Config) Views() []*ViewConfig {
	return c.views
}

func (c Config) HassDiscovery() []*HassDiscovery {
	return c.hassDiscovery
}

func (c Config) GetViewNames() (ret []string) {
	ret = []string{}
	for _, v := range c.Views() {
		ret = append(ret, v.Name())
	}
	return
}

// Getters for HttpServerConfig struct

func (c HttpServerConfig) Enabled() bool {
	return c.enabled
}

func (c HttpServerConfig) Bind() string {
	return c.bind
}

func (c HttpServerConfig) Port() int {
	return c.port
}

func (c HttpServerConfig) LogRequests() bool {
	return c.logRequests
}

func (c HttpServerConfig) FrontendProxy() *url.URL {
	return c.frontendProxy
}

func (c HttpServerConfig) FrontendPath() string {
	return c.frontendPath
}

func (c HttpServerConfig) FrontendExpires() time.Duration {
	return c.frontendExpires
}

func (c HttpServerConfig) ConfigExpires() time.Duration {
	return c.configExpires
}

func (c HttpServerConfig) LogDebug() bool {
	return c.logDebug
}

// Getters for Authentication struct

func (c AuthenticationConfig) Enabled() bool {
	return c.enabled
}

func (c AuthenticationConfig) JwtSecret() []byte {
	return c.jwtSecret
}

func (c AuthenticationConfig) JwtValidityPeriod() time.Duration {
	return c.jwtValidityPeriod
}

func (c AuthenticationConfig) HtaccessFile() string {
	return c.htaccessFile
}

// Getters for MqttClientConfig struct

func (c MqttClientConfig) Name() string {
	return c.name
}

// Broker always returns a non-nil pointer.
func (c MqttClientConfig) Broker() *url.URL {
	return c.broker
}

func (c MqttClientConfig) ProtocolVersion() int {
	return c.protocolVersion
}

func (c MqttClientConfig) User() string {
	return c.user
}

func (c MqttClientConfig) Password() string {
	return c.password
}

func (c MqttClientConfig) ClientId() string {
	return c.clientId
}

func (c MqttClientConfig) Qos() byte {
	return c.qos
}

func (c MqttClientConfig) KeepAlive() time.Duration {
	return c.keepAlive
}

func (c MqttClientConfig) ConnectRetryDelay() time.Duration {
	return c.connectRetryDelay
}

func (c MqttClientConfig) ConnectTimeout() time.Duration {
	return c.connectTimeout
}

func (c MqttClientConfig) AvailabilityTopic() string {
	return c.availabilityTopic
}

func (c MqttClientConfig) TelemetryInterval() time.Duration {
	return c.telemetryInterval
}

func (c MqttClientConfig) TelemetryTopic() string {
	return c.telemetryTopic
}

func (c MqttClientConfig) TelemetryRetain() bool {
	return c.telemetryRetain
}

func (c MqttClientConfig) RealtimeEnable() bool {
	return c.realtimeEnable
}

func (c MqttClientConfig) RealtimeTopic() string {
	return c.realtimeTopic
}

func (c MqttClientConfig) RealtimeRetain() bool {
	return c.realtimeRetain
}

func (c MqttClientConfig) TopicPrefix() string {
	return c.topicPrefix
}

func (c MqttClientConfig) LogDebug() bool {
	return c.logDebug
}

func (c MqttClientConfig) LogMessages() bool {
	return c.logMessages
}

// Getters for ModbusConfig struct

func (c ModbusConfig) Name() string {
	return c.name
}

func (c ModbusConfig) Device() string {
	return c.device
}

func (c ModbusConfig) BaudRate() int {
	return c.baudRate
}

func (c ModbusConfig) ReadTimeout() time.Duration {
	return c.readTimeout
}

func (c ModbusConfig) LogDebug() bool {
	return c.logDebug
}

// Getters for DeviceConfig struct

func (c DeviceConfig) Name() string {
	return c.name
}

func (c DeviceConfig) SkipFields() []string {
	return c.skipFields
}

func (c DeviceConfig) SkipCategories() []string {
	return c.skipCategories
}

func (c DeviceConfig) TelemetryViaMqttClients() []string {
	return c.telemetryViaMqttClients
}

func (c DeviceConfig) RealtimeViaMqttClients() []string {
	return c.realtimeViaMqttClients
}

func (c DeviceConfig) RestartInterval() time.Duration {
	return c.restartInterval
}

func (c DeviceConfig) RestartIntervalMaxBackoff() time.Duration {
	return c.restartIntervalMaxBackoff
}

func (c DeviceConfig) LogDebug() bool {
	return c.logDebug
}

func (c DeviceConfig) LogComDebug() bool {
	return c.logComDebug
}

// Getters for VictronDeviceConfig struct

func (c VictronDeviceConfig) Device() string {
	return c.device
}

func (c VictronDeviceConfig) Kind() VictronDeviceKind {
	return c.kind
}

// Getters for ModbusDeviceConfig struct

func (c ModbusDeviceConfig) Bus() string {
	return c.bus
}

func (c ModbusDeviceConfig) Kind() ModbusDeviceKind {
	return c.kind
}

func (c ModbusDeviceConfig) Address() byte {
	return c.address
}

func (c ModbusDeviceConfig) RelayDescription(name string) string {
	if v, ok := c.relays[name]; ok {
		return v.description
	}
	return name
}

func (c ModbusDeviceConfig) RelayOpenLabel(name string) string {
	if v, ok := c.relays[name]; ok {
		return v.openLabel
	}
	return "open"
}

func (c ModbusDeviceConfig) RelayClosedLabel(name string) string {
	if v, ok := c.relays[name]; ok {
		return v.closedLabel
	}
	return "closed"
}

func (c ModbusDeviceConfig) PollInterval() time.Duration {
	return c.pollInterval
}

// Getters for HttpDeviceConfig struct

func (c HttpDeviceConfig) Url() *url.URL {
	return c.url
}

func (c HttpDeviceConfig) Kind() HttpDeviceKind {
	return c.kind
}

func (c HttpDeviceConfig) Username() string {
	return c.username
}

func (c HttpDeviceConfig) Password() string {
	return c.password
}

func (c HttpDeviceConfig) PollInterval() time.Duration {
	return c.pollInterval
}

func (c HttpDeviceConfig) LogDebug() bool {
	return c.logDebug
}

func (c MqttDeviceConfig) MqttTopics() []string {
	return c.mqttTopics
}

// Getters for MqttDeviceConfig struct

func (c MqttDeviceConfig) MqttClients() []string {
	return c.mqttClients
}

// Getters for ViewConfig struct

func (c ViewConfig) Name() string {
	return c.name
}

func (c ViewConfig) Title() string {
	return c.title
}

func (c ViewConfig) Devices() []*ViewDeviceConfig {
	return c.devices
}

func (c ViewConfig) DeviceNames() []string {
	names := make([]string, len(c.devices))
	for i, device := range c.devices {
		names[i] = device.Name()
	}
	return names
}

func (c ViewConfig) Autoplay() bool {
	return c.autoplay
}

func (c ViewConfig) IsAllowed(user string) bool {
	_, ok := c.allowedUsers[user]
	return ok
}

func (c ViewConfig) IsPublic() bool {
	return len(c.allowedUsers) == 0
}

func (c ViewConfig) Hidden() bool {
	return c.hidden
}

// Getters for ViewDeviceConfig struct

func (c ViewDeviceConfig) Name() string {
	return c.name
}

func (c ViewDeviceConfig) Title() string {
	return c.title
}

func (c ViewDeviceConfig) SkipFields() []string {
	return c.skipFields
}

func (c ViewDeviceConfig) SkipCategories() []string {
	return c.skipCategories
}

// Gettters for HassDiscovery struct

func (c HassDiscovery) TopicPrefix() string {
	return c.topicPrefix
}

func (c HassDiscovery) ViaMqttClients() []string {
	return c.viaMqttClients
}

func (c HassDiscovery) Devices() []string {
	return c.devices
}

func (c HassDiscovery) Categories() []string {
	return c.categories
}

func (c HassDiscovery) CategoriesMatcher() []*regexp.Regexp {
	return c.categoriesMatcher
}

func (c HassDiscovery) Registers() []string {
	return c.registers
}

func (c HassDiscovery) RegistersMatcher() []*regexp.Regexp {
	return c.registersMatcher
}
