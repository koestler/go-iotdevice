package config

import (
	"net/url"
	"strings"
	"time"

	"github.com/koestler/go-iotdevice/v3/types"
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

func (c Config) LogStateStorageDebug() bool {
	return c.logStateStorageDebug
}

func (c Config) LogCommandStorageDebug() bool {
	return c.logCommandStorageDebug
}

func (c Config) HttpServer() HttpServerConfig {
	return c.httpServer
}

func (c Config) Authentication() AuthenticationConfig {
	return c.authentication
}

func (c Config) MqttClients() []MqttClientConfig {
	return c.mqttClients
}

func (c Config) Modbus() []ModbusConfig {
	return c.modbus
}

func (c Config) Devices() []DeviceConfig {
	return c.devices
}

func (c Config) VictronDevices() []VictronDeviceConfig {
	return c.victronDevices
}

func (c Config) ModbusDevices() []ModbusDeviceConfig {
	return c.modbusDevices
}

func (c Config) GpioDevices() []GpioDeviceConfig {
	return c.gpioDevices
}

func (c Config) HttpDevices() []HttpDeviceConfig {
	return c.httpDevices
}

func (c Config) MqttDevices() []MqttDeviceConfig {
	return c.mqttDevices
}

func (c Config) GensetDevices() []GensetDeviceConfig {
	return c.gensetDevices
}

func (c Config) Views() []ViewConfig {
	return c.views
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

func (c MqttClientConfig) getTopicTemplateOldNewPairs(oldnew ...string) []string {
	return append(oldnew, "%Prefix%", c.TopicPrefix(), "%ClientId%", c.ClientId())
}

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

func (c MqttClientConfig) KeepAlive() time.Duration {
	return c.keepAlive
}

func (c MqttClientConfig) ConnectRetryDelay() time.Duration {
	return c.connectRetryDelay
}

func (c MqttClientConfig) ConnectTimeout() time.Duration {
	return c.connectTimeout
}

func (c MqttClientConfig) TopicPrefix() string {
	return c.topicPrefix
}

func (c MqttClientConfig) ReadOnly() bool {
	return c.readOnly
}

func (c MqttClientConfig) MaxBacklogSize() int {
	return c.maxBacklogSize
}

func (c MqttClientConfig) MqttDevices() []MqttClientDeviceConfig {
	return c.mqttDevices
}

func (c MqttClientConfig) AvailabilityClient() MqttSectionConfig {
	return c.availabilityClient
}

func (c MqttClientConfig) AvailabilityClientTopic() string {
	r := strings.NewReplacer(c.getTopicTemplateOldNewPairs()...)
	return r.Replace(c.availabilityClient.topicTemplate)
}

func (c MqttClientConfig) AvailabilityDevice() MqttSectionConfig {
	return c.availabilityDevice
}

func (c MqttClientConfig) AvailabilityDeviceTopic(deviceName string) string {
	r := strings.NewReplacer(c.getTopicTemplateOldNewPairs("%DeviceName%", deviceName)...)
	return r.Replace(c.availabilityDevice.topicTemplate)
}

func (c MqttClientConfig) Structure() MqttSectionConfig {
	return c.structure
}

func (c MqttClientConfig) StructureTopic(deviceName string) string {
	r := strings.NewReplacer(c.getTopicTemplateOldNewPairs("%DeviceName%", deviceName)...)
	return r.Replace(c.structure.topicTemplate)
}

func (c MqttClientConfig) Telemetry() MqttSectionConfig {
	return c.telemetry
}

func (c MqttClientConfig) TelemetryTopic(deviceName string) string {
	r := strings.NewReplacer(c.getTopicTemplateOldNewPairs("%DeviceName%", deviceName)...)
	return r.Replace(c.telemetry.topicTemplate)
}

func (c MqttClientConfig) Realtime() MqttSectionConfig {
	return c.realtime
}

func (c MqttClientConfig) RealtimeTopic(deviceName, registerName string) string {
	r := strings.NewReplacer(c.getTopicTemplateOldNewPairs(
		"%DeviceName%", deviceName,
		"%RegisterName%", registerName,
	)...)
	return r.Replace(c.realtime.topicTemplate)
}

func (c MqttClientConfig) HomeassistantDiscovery() MqttSectionConfig {
	return c.homeassistantDiscovery
}

func (c MqttClientConfig) HomeassistantDiscoveryTopic(component, nodeId, objectId string) string {
	r := strings.NewReplacer(c.getTopicTemplateOldNewPairs(
		"%Component%", component,
		"%NodeId%", nodeId,
		"%ObjectId%", objectId,
	)...)
	return r.Replace(c.homeassistantDiscovery.topicTemplate)
}

func (c MqttClientConfig) Command() MqttSectionConfig {
	return c.command
}

func (c MqttClientConfig) CommandTopic(deviceName, registerName string) string {
	r := strings.NewReplacer(c.getTopicTemplateOldNewPairs(
		"%DeviceName%", deviceName,
		"%RegisterName%", registerName,
	)...)
	return r.Replace(c.command.topicTemplate)
}

func (c MqttClientConfig) LogDebug() bool {
	return c.logDebug
}

func (c MqttClientConfig) LogMessages() bool {
	return c.logMessages
}

// Getters for MqttClientDeviceConfig

func (c MqttClientDeviceConfig) Name() string {
	return c.name
}

func (c MqttClientDeviceConfig) MqttTopics() []string {
	return c.mqttTopics
}

// Getters for MqttSection struct

func (c MqttSectionConfig) Enabled() bool {
	return c.enabled
}

func (c MqttSectionConfig) TopicTemplate() string {
	return c.topicTemplate
}

func (c MqttSectionConfig) Interval() time.Duration {
	return c.interval
}

func (c MqttSectionConfig) Retain() bool {
	return c.retain
}

func (c MqttSectionConfig) Qos() byte {
	return c.qos
}

func (c MqttSectionConfig) Devices() []MqttDeviceSectionConfig {
	return c.devices
}

// Getters for MqttDeviceSectionConfig struct

func (c MqttDeviceSectionConfig) Name() string {
	return c.name
}

func (c MqttDeviceSectionConfig) Filter() FilterConfig {
	return c.filter
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

func (c DeviceConfig) Filter() FilterConfig {
	return c.filter
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

func (c VictronDeviceConfig) Kind() types.VictronDeviceKind {
	return c.kind
}

func (c VictronDeviceConfig) PollInterval() time.Duration {
	return c.pollInterval
}

func (c VictronDeviceConfig) IoLog() string {
	return c.ioLog
}

// Getters for ModbusDeviceConfig struct

func (c ModbusDeviceConfig) Bus() string {
	return c.bus
}

func (c ModbusDeviceConfig) Kind() types.ModbusDeviceKind {
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

// Getters for GpioDeviceConfig struct
func (c GpioDeviceConfig) Chip() string {
	return c.chip
}

func (c GpioDeviceConfig) InputDebounce() time.Duration {
	return c.inputDebounce
}

func (c GpioDeviceConfig) InputOptions() []string {
	return c.inputOptions
}

func (c GpioDeviceConfig) OutputOptions() []string {
	return c.outputOptions
}

func (c GpioDeviceConfig) Inputs() []PinConfig {
	return c.inputs
}

func (c GpioDeviceConfig) Outputs() []PinConfig {
	return c.outputs
}

// Getters for PinConfig struct

func (c PinConfig) Pin() string {
	return c.pin
}

func (c PinConfig) Name() string {
	return c.name
}

func (c PinConfig) Description() string {
	return c.description
}

func (c PinConfig) LowLabel() string {
	return c.lowLabel
}

func (c PinConfig) HighLabel() string {
	return c.highLabel
}

// Getters for HttpDeviceConfig struct

func (c HttpDeviceConfig) Url() *url.URL {
	return c.url
}

func (c HttpDeviceConfig) Kind() types.HttpDeviceKind {
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

// Getters for MqttDeviceConfig struct

func (c MqttDeviceConfig) Kind() types.MqttDeviceKind {
	return c.kind
}

// Getter for GensetDeviceConfig struct

func (c GensetDeviceConfig) InputBindings() []GensetDeviceBindingConfig {
	return c.inputBindings
}

func (c GensetDeviceConfig) OutputBindings() []GensetDeviceBindingConfig {
	return c.outputBindings
}

func (c GensetDeviceConfig) PrimingTimeout() time.Duration {
	return c.primingTimeout
}

func (c GensetDeviceConfig) CrankingTimeout() time.Duration {
	return c.crankingTimeout
}

func (c GensetDeviceConfig) StabilizingTimeout() time.Duration {
	return c.stabilizingTimeout
}

func (c GensetDeviceConfig) WarmUpTimeout() time.Duration {
	return c.warmUpTimeout
}

func (c GensetDeviceConfig) WarmUpMinTime() time.Duration {
	return c.warmUpMinTime
}

func (c GensetDeviceConfig) WarmUpTemp() float64 {
	return c.warmUpTemp
}

func (c GensetDeviceConfig) EngineCoolDownTimeout() time.Duration {
	return c.engineCoolDownTimeout
}

func (c GensetDeviceConfig) EngineCoolDownMinTime() time.Duration {
	return c.engineCoolDownMinTime
}

func (c GensetDeviceConfig) EngineCoolDownTemp() float64 {
	return c.engineCoolDownTemp
}

func (c GensetDeviceConfig) EnclosureCoolDownTimeout() time.Duration {
	return c.enclosureCoolDownTimeout
}

func (c GensetDeviceConfig) EnclosureCoolDownMinTime() time.Duration {
	return c.enclosureCoolDownMinTime
}

func (c GensetDeviceConfig) EnclosureCoolDownTemp() float64 {
	return c.enclosureCoolDownTemp
}

func (c GensetDeviceConfig) EngineTempMin() float64 {
	return c.engineTempMin
}

func (c GensetDeviceConfig) EngineTempMax() float64 {
	return c.engineTempMax
}

func (c GensetDeviceConfig) AuxTemp0Min() float64 {
	return c.auxTemp0Min
}

func (c GensetDeviceConfig) AuxTemp0Max() float64 {
	return c.auxTemp0Max
}

func (c GensetDeviceConfig) AuxTemp1Min() float64 {
	return c.auxTemp1Min
}

func (c GensetDeviceConfig) AuxTemp1Max() float64 {
	return c.auxTemp1Max
}

func (c GensetDeviceConfig) SinglePhase() bool {
	return c.singlePhase
}

func (c GensetDeviceConfig) UMin() float64 {
	return c.uMin
}

func (c GensetDeviceConfig) UMax() float64 {
	return c.uMax
}

func (c GensetDeviceConfig) UAvgWindow() int {
	return c.uAvgWindow
}

func (c GensetDeviceConfig) FMin() float64 {
	return c.fMin
}

func (c GensetDeviceConfig) FMax() float64 {
	return c.fMax
}

func (c GensetDeviceConfig) FAvgWindow() int {
	return c.fAvgWindow
}

func (c GensetDeviceConfig) PMax() float64 {
	return c.pMax
}

func (c GensetDeviceConfig) PTotMax() float64 {
	return c.pTotMax
}

// Getter for GensetDeviceBindingConfig struct

func (c GensetDeviceBindingConfig) Name() string {
	return c.name
}

func (c GensetDeviceBindingConfig) DeviceName() string {
	return c.deviceName
}

func (c GensetDeviceBindingConfig) RegisterName() string {
	return c.registerName
}

// Getters for ViewConfig struct

func (c ViewConfig) Name() string {
	return c.name
}

func (c ViewConfig) Title() string {
	return c.title
}

func (c ViewConfig) Devices() []ViewDeviceConfig {
	return c.devices
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

func (c ViewDeviceConfig) Filter() FilterConfig {
	return c.filter
}

// Getters for FilterConfig struct

func (c FilterConfig) IncludeRegisters() []string {
	return c.includeRegisters
}

func (c FilterConfig) SkipRegisters() []string {
	return c.skipRegisters
}

func (c FilterConfig) IncludeCategories() []string {
	return c.includeCategories
}

func (c FilterConfig) SkipCategories() []string {
	return c.skipCategories
}

func (c FilterConfig) DefaultInclude() bool {
	return c.defaultInclude
}
