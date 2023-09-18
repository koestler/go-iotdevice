package device

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"strings"
)

type TelemetryMessage struct {
	Time                   string                           `json:"Time"`
	NextTelemetry          string                           `json:"NextTelemetry"`
	Model                  string                           `json:"Model"`
	SecondsSinceLastUpdate float64                          `json:"SecondsSinceLastUpdate"`
	NumericValues          map[string]NumericTelemetryValue `json:"NumericValues,omitempty"`
	TextValues             map[string]TextTelemetryValue    `json:"TextValues,omitempty"`
	EnumValues             map[string]EnumTelemetryValue    `json:"EnumValues,omitempty"`
}

type NumericTelemetryValue struct {
	Category    string  `json:"Cat"`
	Description string  `json:"Desc"`
	Value       float64 `json:"Val"`
	Unit        string  `json:"Unit,omitempty"`
}

type TextTelemetryValue struct {
	Category    string `json:"Cat"`
	Description string `json:"Desc"`
	Value       string `json:"Val"`
}

type EnumTelemetryValue struct {
	Category    string `json:"Cat"`
	Description string `json:"Desc"`
	EnumIdx     int    `json:"Idx"`
	Value       string `json:"Val"`
}

func convertValuesToNumericTelemetryValues(values []dataflow.Value) (ret map[string]NumericTelemetryValue) {
	ret = make(map[string]NumericTelemetryValue)

	for _, value := range values {
		if numeric, ok := value.(dataflow.NumericRegisterValue); ok {
			ret[value.Register().Name()] = NumericTelemetryValue{
				Category:    numeric.Register().Category(),
				Description: numeric.Register().Description(),
				Value:       numeric.Value(),
				Unit:        numeric.Register().Unit(),
			}
		}
	}

	return
}

func convertValuesToTextTelemetryValues(values []dataflow.Value) (ret map[string]TextTelemetryValue) {
	ret = make(map[string]TextTelemetryValue)

	for _, value := range values {
		if text, ok := value.(dataflow.TextRegisterValue); ok {
			ret[value.Register().Name()] = TextTelemetryValue{
				Category:    text.Register().Category(),
				Description: text.Register().Description(),
				Value:       text.Value(),
			}
		}
	}

	return
}

func convertValuesToEnumTelemetryValues(values []dataflow.Value) (ret map[string]EnumTelemetryValue) {
	ret = make(map[string]EnumTelemetryValue)

	for _, value := range values {
		if enum, ok := value.(dataflow.EnumRegisterValue); ok {
			ret[value.Register().Name()] = EnumTelemetryValue{
				Category:    enum.Register().Category(),
				Description: enum.Register().Description(),
				EnumIdx:     enum.EnumIdx(),
				Value:       enum.Value(),
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
