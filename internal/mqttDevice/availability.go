package mqttDevice

import (
	"log"

	"github.com/koestler/go-iotdevice/v3/internal/device"
	"github.com/koestler/go-iotdevice/v3/internal/mqttClient"
)

type availabilityMessage struct {
	Topic     string
	Available bool
}

func (f *forwarder) setupAvailabilitySubscription(availabilityTopics []string) (remove func()) {
	removes := make([]func(), 0, len(availabilityTopics))
	for _, topic := range availabilityTopics {
		if f.Device.Config().LogDebug() {
			log.Printf("mqttDevice[%s]->mqttClient[%s]: subscribe to topic=%s", f.Device.Name(), f.Client.Name(), topic)
		}

		r := f.Client.AddRoute(
			topic,
			availabilityHandler(
				f.Device.Config().LogDebug(),
				f.Device.Name(), f.Client.Name(),
				f.availMessage,
			),
		)
		removes = append(removes, r)
	}

	return func() {
		for _, remove := range removes {
			remove()
		}
	}
}

func availabilityHandler(logDebug bool, deviceName, clientName string, c chan<- availabilityMessage) func(message mqttClient.Message) {
	return func(m mqttClient.Message) {
		if logDebug {
			log.Printf("mqttDevice[%s]->mqttClient[%s]: received availability topic=%s, msg=%s",
				deviceName, clientName, m.Topic(), m.Payload(),
			)
		}

		// go routine of paho router is not blocked because c is buffered
		c <- parseAvailabilityPayload(m)
	}
}

func parseAvailabilityPayload(m mqttClient.Message) availabilityMessage {
	return availabilityMessage{
		Topic:     m.Topic(),
		Available: string(m.Payload()) == device.AvailabilityOnlineValue,
	}
}
