package mqttForwarders

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
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

func runTelemetryForwarder(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	storage *dataflow.ValueStorage,
	registerFilter dataflow.RegisterFilterConf,
) {
	mCfg := cfg.Telemetry()

	telemetryInterval := mCfg.Interval()
	telemetryTopic := cfg.TelemetryTopic(dev.Name())

	filter := createDeviceAndRegisterValueFilter(dev, registerFilter)

	go func(mc mqttClient.Client) {
		if cfg.LogDebug() {
			log.Printf(
				"mqttClient[%s]->device[%s]->telemetry: start sending messages every %s",
				mc.Name(), dev.Name(), telemetryInterval.String(),
			)
		}

		ticker := time.NewTicker(telemetryInterval)
		defer ticker.Stop()

		var avail bool
		availChan := dev.SubscribeAvailableSendInitial(ctx)
		for {
			select {
			case <-ctx.Done():

				return
			case avail = <-availChan:
				if cfg.LogDebug() {
					s := "stopped"
					if avail {
						s = "started"
					}

					log.Printf(
						"mqttClient[%s]->device[%s]->telemetry: %s sending due to availability",
						mc.Name(), dev.Name(), s,
					)
				}
			case <-ticker.C:
				if cfg.LogDebug() {
					log.Printf("mqttClient[%s]->device[%s]->telemetry: tick", mc.Name(), dev.Name())
				}

				if !avail {
					// do not send telemetry when device is disconnected
					continue
				}

				values := storage.GetStateFiltered(filter)

				now := time.Now()
				telemetryMessage := TelemetryMessage{
					Time:          now.Format(time.RFC3339),
					NextTelemetry: now.Add(telemetryInterval).Format(time.RFC3339),
					Model:         dev.Model(),
					NumericValues: convertValuesToNumericTelemetryValues(values),
					TextValues:    convertValuesToTextTelemetryValues(values),
					EnumValues:    convertValuesToEnumTelemetryValues(values),
				}

				if payload, err := json.Marshal(telemetryMessage); err != nil {
					log.Printf(
						"mqttClient[%s]->device[%s]->telemetry: cannot generate message: %s",
						mc.Name(), dev.Name(), err,
					)
				} else {
					mc.Publish(
						telemetryTopic,
						payload,
						mCfg.Qos(),
						mCfg.Retain(),
					)
				}
			}
		}
	}(mc)
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
