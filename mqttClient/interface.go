package mqttClient

import (
	"github.com/koestler/go-iotdevice/dataflow"
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
	RealtimeTopic() string
	RealtimeRetain() bool
	TopicPrefix() string
	LogDebug() bool
	LogMessages() bool
}

type Client interface {
	Name() string
	Run()
	Shutdown()
	ReplaceTemplate(template string) string
	AddRoute(subscribeTopic string, messageHandler MessageHandler)
	PublishTelemetryMessage(deviceName string, model string, lastUpdated time.Time, values []dataflow.Value) error
	PublishRealtimeMessage(value dataflow.Value) error
	TelemetryInterval() time.Duration
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
