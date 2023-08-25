package device

import (
	"context"
	"encoding/json"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"log"
	"time"
)

func RunMqttForwarders(
	ctx context.Context,
	d Device,
	mqttClientPool *pool.Pool[mqttClient.Client],
	storage *dataflow.ValueStorage,
) {
	deviceName := d.Config().Name()
	deviceFilter := func(v dataflow.Value) bool {
		return v.DeviceName() == deviceName
	}

	// start mqtt forwarders for realtime messages (send data as soon as it arrives) output
	for _, mc := range mqttClientPool.GetByNames(d.Config().RealtimeViaMqttClients()) {
		if !mc.Config().RealtimeEnable() {
			continue
		}

		subscription := storage.Subscribe(ctx, deviceFilter)

		// transmitRealtime values from data store and publish to mqtt broker
		go func() {
			for {
				select {
				case <-ctx.Done():
					if d.Config().LogDebug() {
						log.Printf(
							"device[%s]->mqttClient[%s]: context canceled, exit transmit realtime",
							d.Config().Name(), mc.Config().Name(),
						)
					}
					return
				case value := <-subscription.Drain():
					if d.Config().LogDebug() {
						log.Printf(
							"device[%s]->mqttClient[%s]: send realtime : %s",
							d.Config().Name(), mc.Config().Name(), value,
						)
					}

					if payload, err := json.Marshal(convertValueToRealtimeMessage(value)); err != nil {
						log.Printf(
							"device[%s]->mqttClient[%s]: cannot generate realtime message: %s",
							d.Config().Name(), mc.Config().Name(), err,
						)
					} else {
						mc.Publish(
							GetRealtimeTopic(mc.Config().RealtimeTopic(), d.Config().Name(), value.Register()),
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
			d.Config().Name(), mc.Config().Name(),
		)
	}

	// start mqtt forwarders for telemetry messages
	for _, mc := range mqttClientPool.GetByNames(d.Config().TelemetryViaMqttClients()) {
		if telemetryInterval := mc.Config().TelemetryInterval(); telemetryInterval > 0 {
			go func() {
				ticker := time.NewTicker(telemetryInterval)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						if d.Config().LogDebug() {
							log.Printf(
								"device[%s]->mqttClient[%s]: context canceled, exit transmit telemetry",
								d.Config().Name(), mc.Config().Name(),
							)
						}
						return
					case <-ticker.C:
						if d.Config().LogDebug() {
							log.Printf(
								"device[%s]->mqttClient[%s]: telemetry tick",
								d.Config().Name(), mc.Config().Name(),
							)
						}

						if !d.IsAvailable() {
							// do not send telemetry when device is disconnected
							continue
						}

						values := storage.GetStateFiltered(deviceFilter)

						now := time.Now()
						telemetryMessage := TelemetryMessage{
							Time:                   timeToString(now),
							NextTelemetry:          timeToString(now.Add(telemetryInterval)),
							Model:                  d.Model(),
							SecondsSinceLastUpdate: now.Sub(d.LastUpdated()).Seconds(),
							NumericValues:          convertValuesToNumericTelemetryValues(values),
							TextValues:             convertValuesToTextTelemetryValues(values),
						}

						if payload, err := json.Marshal(telemetryMessage); err != nil {
							log.Printf(
								"device[%s]->mqttClient[%s]: cannot generate telemetry message: %s",
								d.Config().Name(), mc.Config().Name(), err,
							)
						} else {
							mc.Publish(
								getTelemetryTopic(mc.Config().TelemetryTopic(), d),
								payload,
								mc.Config().Qos(),
								mc.Config().TelemetryRetain(),
							)
						}
					}
				}
			}()

			log.Printf(
				"device[%s]->mqttClient[%s]: start sending telemetry messages every %s",
				d.Config().Name(), mc.Config().Name(), telemetryInterval.String(),
			)
		}
	}
}

func timeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}
