package mqttClient

import (
	"encoding/json"
	"github.com/koestler/go-ve-sensor/dataflow"
	"strings"
	"time"
)

type RealtimeMessage struct {
	Time  string
	Value float64
	Unit  string
}

func convertValueToRealtimeMessage(value dataflow.Value) (RealtimeMessage) {
	return RealtimeMessage{
		Time:  timeToString(time.Now()),
		Value: value.Value,
		Unit:  value.Unit,
	}
}

func transmitRealtime(input <-chan dataflow.Value, mqttClient *MqttClient) {
	go func() {
		cfg := mqttClient.config

		topic := replaceTemplate(cfg.RealtimeTopic, cfg)

		for value := range input {
			if !mqttClient.client.IsConnected() {
				continue
			}

			// replace Device/Value specific palceholders
			topic := strings.Replace(topic, "%DeviceName%", value.Device.Name, 1)
			topic = strings.Replace(topic, "%DeviceModel%", value.Device.Model, 1)
			topic = strings.Replace(topic, "%ValueName%", value.Name, 1)
			topic = strings.Replace(topic, "%ValueUnit%", value.Unit, 1)

			if b, err := json.Marshal(convertValueToRealtimeMessage(value)); err == nil {
				mqttClient.client.Publish(topic, cfg.Qos, cfg.RealtimeRetain, b)
			}
		}
	}()
}