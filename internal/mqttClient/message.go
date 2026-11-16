package mqttClient

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
