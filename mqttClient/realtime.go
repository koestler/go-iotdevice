package mqttClient

import (
	"encoding/json"
	"github.com/koestler/go-iotdevice/dataflow"
	"strings"
	"time"
)

type NumericRealtimeMessage struct {
	Time  string
	Value float64
	Unit  string
}

type TextRealtimeMessage struct {
	Time  string
	Value string
}

func convertValueToRealtimeMessage(value dataflow.Value) ([]byte, error) {
	now := timeToString(time.Now())
	var v interface{}
	if numeric, ok := value.(dataflow.NumericRegisterValue); ok {
		v = NumericRealtimeMessage{
			Time:  now,
			Value: numeric.Value(),
			Unit: func() string {
				if u := numeric.Register().Unit(); u != nil {
					return *u
				}
				return ""
			}(),
		}
	} else if text, ok := value.(dataflow.TextRegisterValue); ok {
		v = TextRealtimeMessage{
			Time:  now,
			Value: text.Value(),
		}
	}

	return json.Marshal(v)
}

func getRealtimeTopic(
	cfg Config,
	deviceName string,
	valueName string,
	valueUnit *string,
) string {
	topic := replaceTemplate(cfg.RealtimeTopic(), cfg)
	// replace Device/Value specific placeholders

	topic = strings.Replace(topic, "%DeviceName%", deviceName, 1)
	topic = strings.Replace(topic, "%ValueName%", valueName, 1)
	if valueUnit != nil {
		topic = strings.Replace(topic, "%ValueUnit%", *valueUnit, 1)
	}

	return topic
}
