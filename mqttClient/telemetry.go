package mqttClient

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"strings"
)

type TelemetryMessage struct {
	Time                   string
	NextTelemetry          string
	Model                  string
	SecondsSinceLastUpdate float64
	Values                 map[string]interface{}
}

type NumericTelemetryValue struct {
	Value float64
	Unit  string
}

type TextTelemetryValue struct {
	Value string
}

func convertValuesToTelemetryValues(values []dataflow.Value) (ret map[string]interface{}) {
	ret = make(map[string]interface{}, len(values))

	for _, value := range values {
		if numeric, ok := value.(dataflow.NumericRegisterValue); ok {
			ret[value.Register().Name()] = NumericTelemetryValue{
				Value: numeric.Value(),
				Unit: func() string {
					if u := numeric.Register().Unit(); u != nil {
						return *u
					}
					return ""
				}(),
			}
		} else if text, ok := value.(dataflow.TextRegisterValue); ok {
			ret[value.Register().Name()] = TextTelemetryValue{
				Value: text.Value(),
			}
		}
	}

	return
}

func getTelemetryTopic(
	cfg Config,
	deviceName string,
) string {
	topic := replaceTemplate(cfg.TelemetryTopic(), cfg)
	// replace Device/Value specific placeholders
	topic = strings.Replace(topic, "%DeviceName%", deviceName, 1)
	return topic
}
