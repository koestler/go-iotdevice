package mqttClient

import (
	"net/url"
	"time"
)

type Config interface {
	Name() string
	Broker() *url.URL
	User() string
	Password() string
	ClientId() string
	Qos() byte
	KeepAlive() time.Duration
	ConnectRetryDelay() time.Duration
	ConnectTimeout() time.Duration
	AvailabilityTopic() string
	TelemetryInterval() time.Duration
	TelemetryTopic() string
	TelemetryRetain() bool
	RealtimeEnable() bool
	RealtimeTopic() string
	RealtimeRetain() bool
	TopicPrefix() string
	LogDebug() bool
	LogMessages() bool
}

type Client interface {
	Config() Config
	Run()
	Shutdown()
	Publish(topic string, payload []byte, qos byte, retain bool) error
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
