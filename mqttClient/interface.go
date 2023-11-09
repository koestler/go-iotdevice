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
	TopicPrefix() string
	ReadOnly() bool
	MaxBacklogSize() int

	AvailabilityClient() MqttSectionConfig
	AvailabilityClientTopic() string

	LogDebug() bool
	LogMessages() bool
}

type MqttSectionConfig interface {
	Enabled() bool
	Interval() time.Duration
	Retain() bool
	Qos() byte
}

type Client interface {
	Name() string
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
