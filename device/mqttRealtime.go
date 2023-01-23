package device

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"strings"
)

type RealtimeMessage struct {
	Category     string   `json:"Cat"`
	Description  string   `json:"Desc"`
	NumericValue *float64 `json:"NumVal,omitempty"`
	TextValue    *string  `json:"TextVal,omitempty"`
	Unit         string   `json:"Unit,omitempty"`
	Sort         int      `json:"Sort"`
}

func convertValueToRealtimeMessage(value dataflow.Value) interface{} {
	ret := RealtimeMessage{
		Category:    value.Register().Category(),
		Description: value.Register().Description(),
		Unit:        value.Register().Unit(),
		Sort:        value.Register().Sort(),
	}

	if numeric, ok := value.(dataflow.NumericRegisterValue); ok {
		v := numeric.Value()
		ret.NumericValue = &v
	} else if text, ok := value.(dataflow.TextRegisterValue); ok {
		v := text.Value()
		ret.TextValue = &v
	} else if enum, ok := value.(dataflow.EnumRegisterValue); ok {
		v := enum.Value()
		ret.TextValue = &v
	}

	return ret
}

func getRealtimeTopic(
	topic string,
	device Device,
	register dataflow.Register,
) string {
	topic = strings.Replace(topic, "%DeviceName%", device.Config().Name(), 1)
	topic = strings.Replace(topic, "%ValueName%", register.Name(), 1)
	if valueUnit := register.Unit(); valueUnit != "" {
		topic = strings.Replace(topic, "%ValueUnit%", valueUnit, 1)
	}

	return topic
}
