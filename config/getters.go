package config

import (
	"net/url"
	"time"
)

func (c Config) Version() int {
	return c.version
}

func (c Config) ProjectTitle() string {
	return c.projectTitle
}

func (c Config) Auth() AuthConfig {
	return c.auth
}

func (c Config) MqttClients() []*MqttClientConfig {
	return c.mqttClients
}

func (c Config) Devices() []*DeviceConfig {
	return c.devices
}

func (c Config) VictronDevices() []*VictronDeviceConfig {
	return c.victronDevices
}

func (c Config) TeracomDevices() []*TeracomDeviceConfig {
	return c.teracomDevices
}

func (c Config) MqttDevices() []*MqttDeviceConfig {
	return c.mqttDevices
}

func (c Config) Views() []*ViewConfig {
	return c.views
}

func (c Config) HttpServer() HttpServerConfig {
	return c.httpServer
}

func (c Config) LogConfig() bool {
	return c.logConfig
}

func (c Config) LogWorkerStart() bool {
	return c.logWorkerStart
}

func (c Config) LogDebug() bool {
	return c.logDebug
}

func (c AuthConfig) Enabled() bool {
	return c.enabled
}

func (c AuthConfig) JwtSecret() []byte {
	return c.jwtSecret
}

func (c AuthConfig) JwtValidityPeriod() time.Duration {
	return c.jwtValidityPeriod
}

func (c AuthConfig) HtaccessFile() string {
	return c.htaccessFile
}

// getters for MqttClientConfig struct

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

func (c DeviceConfig) LogDebug() bool {
	return c.logDebug
}

func (c DeviceConfig) LogComDebug() bool {
	return c.logComDebug
}

func (c VictronDeviceConfig) Device() string {
	return c.device
}

func (c VictronDeviceConfig) Kind() VictronDeviceKind {
	return c.kind
}

func (c TeracomDeviceConfig) Url() *url.URL {
	return c.url
}

func (c TeracomDeviceConfig) Username() string {
	return c.username
}

func (c TeracomDeviceConfig) Password() string {
	return c.password
}

func (c MqttDeviceConfig) MqttTopics() []string {
	return c.mqttTopics
}

func (c MqttDeviceConfig) MqttClients() []string {
	return c.mqttClients
}

func (c ViewDeviceConfig) Name() string {
	return c.name
}

func (c ViewDeviceConfig) Title() string {
	return c.title
}

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

func (c ViewConfig) SkipFields() []string {
	return c.skipFields
}

func (c ViewConfig) SkipCategories() []string {
	return c.skipCategories
}

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

func (c HttpServerConfig) EnableDocs() bool {
	return c.enableDocs
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

func (c Config) GetViewNames() (ret []string) {
	ret = []string{}
	for _, v := range c.Views() {
		ret = append(ret, v.Name())
	}
	return
}
