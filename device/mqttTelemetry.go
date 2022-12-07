package device

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"strings"
)

type TelemetryMessage struct {
	Time                   string
	NextTelemetry          string
	Model                  string
	SecondsSinceLastUpdate float64
	NumericValues          map[string]NumericTelemetryValue
	TextValues             map[string]TextTelemetryValue
}

type NumericTelemetryValue struct {
	Value float64
	Unit  string
}

type TextTelemetryValue struct {
	Value string
}

func convertValuesToNumericTelemetryValues(values []dataflow.Value) (ret map[string]NumericTelemetryValue) {
	ret = make(map[string]NumericTelemetryValue, len(values))

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
		}
	}

	return
}

func convertValuesToTextTelemetryValues(values []dataflow.Value) (ret map[string]TextTelemetryValue) {
	ret = make(map[string]TextTelemetryValue, len(values))

	for _, value := range values {
		if text, ok := value.(dataflow.TextRegisterValue); ok {
			ret[value.Register().Name()] = TextTelemetryValue{
				Value: text.Value(),
			}
		}
	}

	return
}

func getTelemetryTopic(topic string, device Device) string {
	// replace Device/Value specific placeholders
	topic = strings.Replace(topic, "%DeviceName%", device.Config().Name(), 1)
	return topic
}
