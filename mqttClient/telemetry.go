package mqttClient

import (
	"encoding/json"
	"github.com/koestler/go-victron-to-mqtt/dataflow"
	"strings"
	"time"
)

type TelemetryMessage struct {
	Time     string
	NextTele string
	TimeZone string
	Model    string
	Values   dataflow.ValueEssentialMap
}

func transmitTelemetry(
	storage *dataflow.ValueStorageInstance,
	filter dataflow.Filter,
	interval time.Duration,
	mqttClient *MqttClient,
) {
	go func() {
		cfg := mqttClient.config

		for now := range time.Tick(interval) {
			for device, deviceState := range storage.GetState(filter) {
				topic := replaceTemplate(cfg.TelemetryTopic, cfg)
				topic = strings.Replace(topic, "%DeviceName%", device.Name, 1)

				payload := TelemetryMessage{
					Time:     timeToString(now.UTC()),
					NextTele: timeToString(now.Add(interval)),
					TimeZone: "UTC",
					Model:    device.Model,
					Values:   deviceState.ConvertToEssential(),
				}

				if b, err := json.Marshal(payload); err == nil {
					mqttClient.client.Publish(topic, cfg.Qos, cfg.TelemetryRetain, b)
				}
			}
		}
	}()
}
