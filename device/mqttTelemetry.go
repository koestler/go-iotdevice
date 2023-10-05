package device

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"log"
	"time"
)

type TelemetryMessage struct {
	Time          string                           `json:"Time"`
	NextTelemetry string                           `json:"NextTelemetry"`
	Model         string                           `json:"Model"`
	NumericValues map[string]NumericTelemetryValue `json:"NumericValues,omitempty"`
	TextValues    map[string]TextTelemetryValue    `json:"TextValues,omitempty"`
	EnumValues    map[string]EnumTelemetryValue    `json:"EnumValues,omitempty"`
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
	devCfg := dev.Config()

	// start mqtt forwarders for telemetry messages
	for _, mc := range mqttClientPool.GetByNames(devCfg.ViaMqttClients()) {
		mcCfg := mc.Config()

		telemetryTopic := mcCfg.TelemetryTopic(devCfg.Name())
		if telemetryInterval := mcCfg.TelemetryInterval(); telemetryInterval > 0 {
			go func(mc mqttClient.Client) {
				ticker := time.NewTicker(telemetryInterval)
				defer ticker.Stop()

				avail, availChan := dev.SubscribeAvailable(ctx)
				for {
					select {
					case <-ctx.Done():
						if devCfg.LogDebug() {
							log.Printf(
								"device[%s]->mqttClient[%s]->telemetry: exit",
								devCfg.Name(), mcCfg.Name(),
							)
						}
						return
					case avail = <-availChan:
						if devCfg.LogDebug() {
							s := "stopped"
							if avail {
								s = "started"
							}

							log.Printf(
								"device[%s]->mqttClient[%s]->telemetry: %s sending due to availability",
								devCfg.Name(), mcCfg.Name(), s,
							)
						}
					case <-ticker.C:
						if devCfg.LogDebug() {
							log.Printf("device[%s]->mqttClient[%s]->telemetry: tick", devCfg.Name(), mcCfg.Name())
						}

						if !avail {
							// do not send telemetry when device is disconnected
							continue
						}

						values := storage.GetStateFiltered(deviceFilter)

						now := time.Now()
						telemetryMessage := TelemetryMessage{
							Time:          timeToString(now),
							NextTelemetry: timeToString(now.Add(telemetryInterval)),
							Model:         dev.Model(),
							NumericValues: convertValuesToNumericTelemetryValues(values),
							TextValues:    convertValuesToTextTelemetryValues(values),
							EnumValues:    convertValuesToEnumTelemetryValues(values),
						}

						if payload, err := json.Marshal(telemetryMessage); err != nil {
							log.Printf(
								"device[%s]->mqttClient[%s]->telemetry: cannot generate message: %s",
								devCfg.Name(), mcCfg.Name(), err,
							)
						} else {
							mc.Publish(
								telemetryTopic,
								payload,
								mcCfg.Qos(),
								mcCfg.TelemetryRetain(),
							)
						}
					}
				}
			}(mc)

			log.Printf(
				"device[%s]->mqttClient[%s]->telemetry: start sending messages every %s",
				devCfg.Name(), mcCfg.Name(), telemetryInterval.String(),
			)
		}
	}
}

func convertValuesToNumericTelemetryValues(values []dataflow.Value) (ret map[string]NumericTelemetryValue) {
	ret = make(map[string]NumericTelemetryValue)

	for _, value := range values {
		if numeric, ok := value.(dataflow.NumericRegisterValue); ok {
			reg := value.Register()
			ret[reg.Name()] = NumericTelemetryValue{
				Category:    reg.Category(),
				Description: reg.Description(),
				Value:       numeric.Value(),
				Unit:        reg.Unit(),
			}
		}
	}

	return
}

func convertValuesToTextTelemetryValues(values []dataflow.Value) (ret map[string]TextTelemetryValue) {
	ret = make(map[string]TextTelemetryValue)

	for _, value := range values {
		if text, ok := value.(dataflow.TextRegisterValue); ok {
			reg := value.Register()
			ret[reg.Name()] = TextTelemetryValue{
				Category:    reg.Category(),
				Description: reg.Description(),
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
			reg := value.Register()
			ret[reg.Name()] = EnumTelemetryValue{
				Category:    reg.Category(),
				Description: reg.Description(),
				EnumIdx:     enum.EnumIdx(),
				Value:       enum.Value(),
			}
		}
	}

	return
}
