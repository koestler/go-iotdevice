package mqttClient

import (
	"encoding/json"
	"github.com/koestler/go-victron-to-mqtt/config"
	"github.com/koestler/go-victron-to-mqtt/dataflow"
	"strings"
	"time"
)

type RealtimeMessage struct {
	Time  string
	Value float64
	Unit  string
}

func convertValueToRealtimeMessage(value dataflow.Value) RealtimeMessage {
	return RealtimeMessage{
		Time:  timeToString(time.Now()),
		Value: value.Value,
		Unit:  value.Unit,
	}
}

func GetRealtimeTopic(
	cfg *config.MqttClientConfig,
	deviceName string,
	deviceModel string,
	valueName string,
	valueUnit string,
) string {
	topic := replaceTemplate(cfg.RealtimeTopic, cfg)
	// replace Device/Value specific palceholders

	topic = strings.Replace(topic, "%DeviceName%", deviceName, 1)
	topic = strings.Replace(topic, "%DeviceModel%", deviceModel, 1)
	topic = strings.Replace(topic, "%ValueName%", valueName, 1)
	topic = strings.Replace(topic, "%ValueUnit%", valueUnit, 1)

	return topic
}

func transmitRealtime(input <-chan dataflow.Value, mqttClient *MqttClient) {
	go func() {
		cfg := mqttClient.config

		for value := range input {
			if !mqttClient.client.IsConnected() {
				continue
			}

			if b, err := json.Marshal(convertValueToRealtimeMessage(value)); err == nil {
				mqttClient.client.Publish(
					GetRealtimeTopic(
						cfg,
						value.Device.Name,
						value.Device.Model,
						value.Name,
						value.Unit,
					),
					cfg.Qos,
					cfg.RealtimeRetain,
					b,
				)
			}
		}
	}()
}
