package mqttClient

import (
	"context"
	"github.com/koestler/go-iotdevice/config"
	"net/url"
	"time"
)

type Config interface {
	Name() string
	Broker() *url.URL

	User() string
	Password() string
	ClientId() string

	KeepAlive() time.Duration
	ConnectRetryDelay() time.Duration
	ConnectTimeout() time.Duration
	TopicPrefix() string
	MaxBacklogSize() int

	AvailabilityClient() config.MqttSectionConfig
	AvailabilityClientTopic() string

	AvailabilityDevice() config.MqttSectionConfig
	AvailabilityDeviceTopic(deviceName string) string

	Structure() config.MqttSectionConfig
	StructureTopic(deviceName string) string

	Telemetry() config.MqttSectionConfig
	TelemetryTopic(deviceName string) string

	Realtime() config.MqttSectionConfig
	RealtimeTopic(deviceName, registerName string) string

	LogDebug() bool
	LogMessages() bool
}

type Client interface {
	Name() string
	Config() Config
	GetCtx() context.Context
	Run()
	Shutdown()
	Publish(topic string, payload []byte, qos byte, retain bool)
	AddRoute(subscribeTopic string, messageHandler MessageHandler)
}

type MessageHandler func(Message)

type Message struct {
	topic   string
	payload []byte
}

func (m Message) Topic() string {
	return m.topic
}

func (m Message) Payload() []byte {
	return m.payload
}
