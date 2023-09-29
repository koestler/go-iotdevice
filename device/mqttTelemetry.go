package device

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"log"
	"strings"
	"time"
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

func runTelemetryForwarders(
	ctx context.Context,
	dev Device,
	mqttClientPool *pool.Pool[mqttClient.Client],
	storage *dataflow.ValueStorage,
	deviceFilter func(v dataflow.Value) bool,
) {
	// start mqtt forwarders for telemetry messages
	for _, mc := range mqttClientPool.GetByNames(dev.Config().TelemetryViaMqttClients()) {
		mc := mc

		if telemetryInterval := mc.Config().TelemetryInterval(); telemetryInterval > 0 {
			go func() {
				ticker := time.NewTicker(telemetryInterval)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						if dev.Config().LogDebug() {
							log.Printf(
								"device[%s]->mqttClient[%s]->telemetry: exit",
								dev.Config().Name(), mc.Config().Name(),
							)
						}
						return
					case <-ticker.C:
						if dev.Config().LogDebug() {
							log.Printf(
								"device[%s]->mqttClient[%s]->telemetry: tick",
								dev.Config().Name(), mc.Config().Name(),
							)
						}

						if !dev.IsAvailable() {
							// do not send telemetry when device is disconnected
							continue
						}

						values := storage.GetStateFiltered(deviceFilter)

						now := time.Now()
						telemetryMessage := TelemetryMessage{
							Time:                   timeToString(now),
							NextTelemetry:          timeToString(now.Add(telemetryInterval)),
							Model:                  dev.Model(),
							SecondsSinceLastUpdate: now.Sub(dev.LastUpdated()).Seconds(),
							NumericValues:          convertValuesToNumericTelemetryValues(values),
							TextValues:             convertValuesToTextTelemetryValues(values),
							EnumValues:             convertValuesToEnumTelemetryValues(values),
						}

						if payload, err := json.Marshal(telemetryMessage); err != nil {
							log.Printf(
								"device[%s]->mqttClient[%s]->telemetry: cannot generate message: %s",
								dev.Config().Name(), mc.Config().Name(), err,
							)
						} else {
							mc.Publish(
								getTelemetryTopic(mc.Config().TelemetryTopic(), dev),
								payload,
								mc.Config().Qos(),
								mc.Config().TelemetryRetain(),
							)
						}
					}
				}
			}()

			log.Printf(
				"device[%s]->mqttClient[%s]->telemetry: start sending messages every %s",
				dev.Config().Name(), mc.Config().Name(), telemetryInterval.String(),
			)
		}
	}
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
