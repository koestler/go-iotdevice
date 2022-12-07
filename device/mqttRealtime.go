package device

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"strings"
	"time"
)

type NumericRealtimeMessage struct {
	Time         string
	NumericValue float64
	Unit         string
}

type TextRealtimeMessage struct {
	Time      string
	TextValue string
}

func convertValueToRealtimeMessage(value dataflow.Value) interface{} {
	now := timeToString(time.Now())
	if numeric, ok := value.(dataflow.NumericRegisterValue); ok {
		return NumericRealtimeMessage{
			Time:         now,
			NumericValue: numeric.Value(),
			Unit: func() string {
				if u := numeric.Register().Unit(); u != nil {
					return *u
				}
				return ""
			}(),
		}
	} else if text, ok := value.(dataflow.TextRegisterValue); ok {
		return TextRealtimeMessage{
			Time:      now,
			TextValue: text.Value(),
		}
	}

	return nil
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
