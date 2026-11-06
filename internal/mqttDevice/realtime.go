package mqttDevice

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/koestler/go-iotdevice/v3/internal/dataflow"
	"github.com/koestler/go-iotdevice/v3/internal/mqttClient"
	"github.com/koestler/go-iotdevice/v3/internal/mqttForwarders"
	"github.com/koestler/go-iotdevice/v3/internal/topicMatcher"
)

func (f *forwarder) setupRealtimeSubscription(topicTemplate string) (remove func()) {
	topic := strings.Replace(topicTemplate, "%RegisterName%", "+", 1)

	tm, err := topicMatcher.CreateMatcherSingleVariable(topicTemplate, "%RegisterName%")
	if err != nil {
		log.Printf("mqttDevice[%s]->mqttClient[%s]: invalid realtime topic: %s", f.Device.Name(), f.Client.Name(), err)
		return
	}

	return f.Client.AddRoute(
		topic,
		realtimeHandler(
			f.LogDebug,
			f.Device.Name(), f.Client.Name(),
			tm,
			f.Device.RegisterDb().GetByName,
			f.Device.StateStorage(),
		),
	)
}

func realtimeHandler(
	logDebug bool,
	deviceName, clientName string,
	tm topicMatcher.TopicMatcher,
	getRegisterByName func(registerName string) (reg dataflow.RegisterStruct, ok bool),
	fillable dataflow.Fillable,
) func(message mqttClient.Message) {
	return func(m mqttClient.Message) {
		if len(m.Payload()) < 1 {
			// ignore empty messages; those are used to remove retained messages
			return
		}

		registerName, err := tm.ParseTopic(m.Topic())
		if err != nil {
			log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse realtime topic: %s", deviceName, clientName, err)
			return
		}

		realtimeMessage, err := parseRealtimePayload(m.Payload())
		if err != nil {
			log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse realtime payload: %s", deviceName, clientName, err)
			return
		}

		// get register
		register, ok := getRegisterByName(registerName)

		if !ok {
			if logDebug {
				log.Printf("mqttDevice[%s]->mqttClient[%s]: unknown register, registerName=%v, ignore",
					deviceName, clientName, registerName,
				)

			}
			return
		}

		switch register.RegisterType() {
		case dataflow.NumberRegister:
			if v := realtimeMessage.NumericValue; v != nil {
				fillable.Fill(dataflow.NewNumericRegisterValue(deviceName, register, *v))
			}
		case dataflow.TextRegister:
			if v := realtimeMessage.TextValue; v != nil {
				fillable.Fill(dataflow.NewTextRegisterValue(deviceName, register, *v))
			}
		case dataflow.EnumRegister:
			if v := realtimeMessage.EnumIdx; v != nil {
				fillable.Fill(dataflow.NewEnumRegisterValue(deviceName, register, *v))
			}
		default:
			if logDebug {
				log.Printf("mqttDevice[%s]->mqttClient[%s]: unsupported register type in realtime message registerName=%v",
					deviceName, clientName, register.Name(),
				)
			}
		}
	}
}

func parseRealtimePayload(payload []byte) (msg mqttForwarders.RealtimeMessage, err error) {
	err = json.Unmarshal(payload, &msg)
	return
}
