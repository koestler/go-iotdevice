package mqttClient

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"strings"
)

type TelemetryMessage struct {
	Time     string
	NextTele string
	TimeZone string
	Model    string
	Values   []interface{}
}

type NumericTelemetryValue struct {
	Value float64
	Unit  string
}

type TextTelemetryValue struct {
	Value string
}

func convertValuesToTelemetryValues(values []dataflow.Value) (ret []interface{}) {
	ret = make([]interface{}, len(values))

	i := 0
	for _, value := range values {
		if numeric, ok := value.(dataflow.NumericRegisterValue); ok {
			ret[i] = NumericTelemetryValue{
				Value: numeric.Value(),
				Unit: func() string {
					if u := numeric.Register().Unit(); u != nil {
						return *u
					}
					return ""
				}(),
			}
			i += 1
		} else if text, ok := value.(dataflow.TextRegisterValue); ok {
			ret[i] = TextTelemetryValue{
				Value: text.Value(),
			}
			i += 1
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
