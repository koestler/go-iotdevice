package mqttClient

import (
	"context"
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
	ReadOnly() bool
	MaxBacklogSize() int

	AvailabilityClient() MqttSectionConfig
	AvailabilityClientTopic() string

	LogDebug() bool
	LogMessages() bool
}

type MqttSectionConfig interface {
	Enabled() bool
	Retain() bool
	Qos() byte
}

type Client interface {
	Name() string
	GetCtx() context.Context
	Shutdown()
	Publish(topic string, payload []byte, qos byte, retain bool)
	AddRoute(ctx context.Context, subscribeTopic string, messageHandler MessageHandler) (remove func())
}

// MessageHandler is called for every received message.
// Handlers are executed synchronously in the client's single run routine.
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
