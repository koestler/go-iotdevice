package mqttClient

import (
	"encoding/json"
	"github.com/eclipse/paho.golang/paho"
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

func (c *ClientStruct) getRealtimePublishMessage(value dataflow.Value) (*paho.Publish, error) {
	payload := convertValueToRealtimeMessage(value)

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &paho.Publish{
		QoS:     c.cfg.Qos(),
		Topic:   c.getRealtimeTopic(value.DeviceName(), value.Register()),
		Payload: b,
		Retain:  c.cfg.TelemetryRetain(),
	}, nil
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

func (c *ClientStruct) getRealtimeTopic(
	deviceName string,
	register dataflow.Register,
) string {
	topic := replaceTemplate(c.cfg.RealtimeTopic(), c.cfg)
	// replace Device/Value specific placeholders

	topic = strings.Replace(topic, "%DeviceName%", deviceName, 1)
	topic = strings.Replace(topic, "%ValueName%", register.Name(), 1)
	if valueUnit := register.Unit(); valueUnit != nil {
		topic = strings.Replace(topic, "%ValueUnit%", *valueUnit, 1)
	}

	return topic
}
