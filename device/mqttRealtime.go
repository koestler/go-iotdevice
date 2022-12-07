package device

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"strings"
)

type RealtimeMessage struct {
	Category     string
	Description  string
	NumericValue *float64
	TextValue    *string
	Unit         *string
}

func convertValueToRealtimeMessage(value dataflow.Value) interface{} {
	ret := RealtimeMessage{
		Category:    value.Register().Category(),
		Description: value.Register().Description(),
		Unit:        value.Register().Unit(),
	}

	if numeric, ok := value.(dataflow.NumericRegisterValue); ok {
		v := numeric.Value()
		ret.NumericValue = &v
	} else if text, ok := value.(dataflow.TextRegisterValue); ok {
		v := text.Value()
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
	if valueUnit := register.Unit(); valueUnit != nil {
		topic = strings.Replace(topic, "%ValueUnit%", *valueUnit, 1)
	}

	return topic
}
