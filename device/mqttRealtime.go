package device

import (
	"context"
	"encoding/json"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"log"
	"strings"
)

type RealtimeMessage struct {
	Category     string   `json:"Cat"`
	Description  string   `json:"Desc"`
	NumericValue *float64 `json:"NumVal,omitempty"`
	TextValue    *string  `json:"TextVal,omitempty"`
	EnumIdx      *int     `json:"EnumIdx,omitempty"`
	Unit         string   `json:"Unit,omitempty"`
	Sort         int      `json:"Sort"`
}

func runRealtimeForwarders(
	ctx context.Context,
	dev Device,
	mqttClientPool *pool.Pool[mqttClient.Client],
	storage *dataflow.ValueStorage,
	deviceFilter func(v dataflow.Value) bool,
) {
	// start mqtt forwarders for realtime messages (send data as soon as it arrives) output
	for _, mc := range mqttClientPool.GetByNames(dev.Config().RealtimeViaMqttClients()) {
		mc := mc

		if !mc.Config().RealtimeEnabled() {
			continue
		}

		subscription := storage.Subscribe(ctx, deviceFilter)

		// transmitRealtime values from data store and publish to mqtt broker
		go func() {
			for {
				select {
				case <-ctx.Done():
					if dev.Config().LogDebug() {
						log.Printf(
							"device[%s]->mqttClient[%s]: context canceled, exit transmit realtime",
							dev.Config().Name(), mc.Config().Name(),
						)
					}
					return
				case value := <-subscription.Drain():
					if dev.Config().LogDebug() {
						log.Printf(
							"device[%s]->mqttClient[%s]: send realtime : %s",
							dev.Config().Name(), mc.Config().Name(), value,
						)
					}

					if payload, err := json.Marshal(convertValueToRealtimeMessage(value)); err != nil {
						log.Printf(
							"device[%s]->mqttClient[%s]: cannot generate realtime message: %s",
							dev.Config().Name(), mc.Config().Name(), err,
						)
					} else {
						mc.Publish(
							GetRealtimeTopic(mc.Config().RealtimeTopic(), dev.Config().Name(), value.Register()),
							payload,
							mc.Config().Qos(),
							mc.Config().RealtimeRetain(),
						)
					}
				}
			}
		}()

		log.Printf(
			"device[%s]->mqttClient[%s]: start sending realtime stat messages",
			dev.Config().Name(), mc.Config().Name(),
		)
	}
}

func convertValueToRealtimeMessage(value dataflow.Value) interface{} {
	ret := RealtimeMessage{
		Category:    value.Register().Category(),
		Description: value.Register().Description(),
		Unit:        value.Register().Unit(),
		Sort:        value.Register().Sort(),
	}

	if numeric, ok := value.(dataflow.NumericRegisterValue); ok {
		v := numeric.Value()
		ret.NumericValue = &v
	} else if text, ok := value.(dataflow.TextRegisterValue); ok {
		v := text.Value()
		ret.TextValue = &v
	} else if enum, ok := value.(dataflow.EnumRegisterValue); ok {
		v := enum.EnumIdx()
		ret.EnumIdx = &v
	}

	return ret
}

func GetRealtimeTopic(
	topic string,
	deviceName string,
	register dataflow.Register,
) string {
	topic = strings.Replace(topic, "%DeviceName%", deviceName, 1)
	topic = strings.Replace(topic, "%ValueName%", register.Name(), 1)
	if valueUnit := register.Unit(); valueUnit != "" {
		topic = strings.Replace(topic, "%ValueUnit%", valueUnit, 1)
	}

	return topic
}
