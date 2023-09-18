package device

import (
	"context"
	"encoding/json"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"log"
	"strings"
	"time"
)

type RealtimeMessage struct {
	NumericValue *float64 `json:"NumVal,omitempty"`
	TextValue    *string  `json:"TextVal,omitempty"`
	EnumIdx      *int     `json:"EnumIdx,omitempty"`
}

func runRealtimeForwarders(
	ctx context.Context,
	dev Device,
	mqttClientPool *pool.Pool[mqttClient.Client],
	storage *dataflow.ValueStorage,
	deviceFilter func(v dataflow.Value) bool,
) {
	devCfg := dev.Config()

	// start mqtt forwarders for realtime messages (send data as soon as it arrives) output
	for _, mc := range mqttClientPool.GetByNames(dev.Config().RealtimeViaMqttClients()) {
		if !mc.Config().RealtimeEnabled() {
			continue
		}

		mcCfg := mc.Config()

		go func(mc mqttClient.Client) {
			if dev.Config().LogDebug() {
				defer func() {
					log.Printf(
						"device[%s]->mqttClient[%s]->realtime: exit",
						dev.Config().Name(), mcCfg.Name(),
					)
				}()
			}

			// mode 0: Interval is set to zero
			// -> send values immediately when they change
			// mode 1: Interval > 0 and Repeat false
			// -> have a timer, whenever it ticks, send the newest version of the changed values
			// mode 2: Interval > 0 and Repeat true
			// -> have a timer, whenever it ticks, send all values

			if realtimeInterval := mcCfg.RealtimeInterval(); realtimeInterval <= 0 {
				// mode 0
				// for loop ends when subscription is canceled and closes its output chan
				subscription := storage.Subscribe(ctx, deviceFilter)
				for value := range subscription.Drain() {
					publishRealtimeMessage(mc, devCfg, value)
				}
			} else {
				log.Printf(
					"device[%s]->mqttClient[%s]->realtime: start sending messages every %s",
					devCfg.Name(), mcCfg.Name(), realtimeInterval.String(),
				)

				ticker := time.NewTicker(realtimeInterval)
				defer ticker.Stop()
				if !mcCfg.RealtimeRepeat() {
					updates := make(map[string]dataflow.Value)

					// mode 1
					subscription := storage.Subscribe(ctx, deviceFilter)
					for {
						select {
						case <-ctx.Done():
							return
						case value := <-subscription.Drain():
							// new value received, save newest version per register name
							updates[value.Register().Name()] = value
						case <-ticker.C:
							if dev.Config().LogDebug() {
								log.Printf(
									"device[%s]->mqttClient[%s]->realtime: tick: send updates",
									dev.Config().Name(), mcCfg.Name(),
								)
							}
							for _, v := range updates {
								publishRealtimeMessage(mc, devCfg, v)
							}
							updates = make(map[string]dataflow.Value)
						}
					}
				} else {
					// mode 2
					for {
						select {
						case <-ctx.Done():
							return
						case <-ticker.C:
							if !dev.IsAvailable() {
								// do not send messages when device is disconnected
								continue
							}

							if dev.Config().LogDebug() {
								log.Printf(
									"device[%s]->mqttClient[%s]->realtime: tick: send everything",
									dev.Config().Name(), mcCfg.Name(),
								)
							}

							values := storage.GetStateFiltered(deviceFilter)
							for _, v := range values {
								publishRealtimeMessage(mc, devCfg, v)
							}
						}
					}
				}
			}
		}(mc)

		log.Printf(
			"device[%s]->mqttClient[%s]->realtime: start sending messages",
			devCfg.Name(), mcCfg.Name(),
		)
	}
}

func publishRealtimeMessage(mc mqttClient.Client, devConfig Config, value dataflow.Value) {
	mcCfg := mc.Config()

	if devConfig.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->realtime: send: %s",
			devConfig.Name(), mcCfg.Name(), value,
		)
	}

	if payload, err := json.Marshal(convertValueToRealtimeMessage(value)); err != nil {
		log.Printf(
			"device[%s]->mqttClient[%s]->realtime: cannot generate message: %s",
			devConfig.Name(), mcCfg.Name(), err,
		)
	} else {
		mc.Publish(
			GetRealtimeTopic(mcCfg.RealtimeTopic(), devConfig.Name(), value.Register()),
			payload,
			mcCfg.Qos(),
			mcCfg.RealtimeRetain(),
		)
	}
}

func convertValueToRealtimeMessage(value dataflow.Value) interface{} {
	ret := RealtimeMessage{}

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
