package mqttClient

import (
	"encoding/json"
	"github.com/eclipse/paho.golang/paho"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"strings"
	"time"
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

func (c *ClientStruct) getTelemetryPublishMessage(deviceName string, dev device.Device, values []dataflow.Value) (*paho.Publish, error) {
	now := time.Now()
	payload := TelemetryMessage{
		Time:                   timeToString(now),
		NextTelemetry:          timeToString(now.Add(c.cfg.TelemetryInterval())),
		Model:                  dev.Model(),
		SecondsSinceLastUpdate: now.Sub(dev.LastUpdated()).Seconds(),
		NumericValues:          convertValuesToNumericTelemetryValues(values),
		TextValues:             convertValuesToTextTelemetryValues(values),
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &paho.Publish{
		QoS:     c.cfg.Qos(),
		Topic:   c.getTelemetryTopic(deviceName),
		Payload: b,
		Retain:  c.cfg.TelemetryRetain(),
	}, nil
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

func (c *ClientStruct) getTelemetryTopic(deviceName string) string {
	topic := replaceTemplate(c.cfg.TelemetryTopic(), c.cfg)
	// replace Device/Value specific placeholders
	topic = strings.Replace(topic, "%DeviceName%", deviceName, 1)
	return topic
}
