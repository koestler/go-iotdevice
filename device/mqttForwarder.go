package device

import (
	"encoding/json"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mqttClient"
	"log"
	"time"
)

func RunMqttForwarders(d Device, mqttClientPool *mqttClient.ClientPool, storage *dataflow.ValueStorageInstance) {
	deviceFilter := dataflow.Filter{IncludeDevices: map[string]bool{d.Config().Name(): true}}

	// start mqtt forwarders for realtime messages (send data as soon as it arrives) output
	for _, mc := range mqttClientPool.GetClientsByNames(d.Config().RealtimeViaMqttClients()) {
		if !mc.Config().RealtimeEnable() {
			continue
		}

		// transmitRealtime values from data store and publish to mqtt broker
		go func() {
			// setup empty filter (everything)
			subscription := storage.Subscribe(deviceFilter)
			defer subscription.Shutdown()

			for {
				select {
				case <-d.ShutdownChan():
					return
				case value := <-subscription.GetOutput():
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
					} else if err := mc.Publish(
						getRealtimeTopic(mc.Config().RealtimeTopic(), d, value.Register()),
						payload,
						mc.Config().Qos(),
						mc.Config().RealtimeRetain(),
					); err != nil {
						log.Printf(
							"device[%s]->mqttClient[%s]: cannot publish realtime: %s",
							d.Config().Name(), mc.Config().Name(), err,
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
	for _, mc := range mqttClientPool.GetClientsByNames(d.Config().TelemetryViaMqttClients()) {
		if telemetryInterval := mc.Config().TelemetryInterval(); telemetryInterval > 0 {
			go func() {
				ticker := time.NewTicker(telemetryInterval)
				defer ticker.Stop()
				for {
					select {
					case <-d.ShutdownChan():
						return
					case <-ticker.C:
						if d.Config().LogDebug() {
							log.Printf(
								"device[%s]->mqttClient[%s]: telemetry tick",
								d.Config().Name(), mc.Config().Name(),
							)
						}

						values := storage.GetSlice(deviceFilter)

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
						} else if err := mc.Publish(
							getTelemetryTopic(mc.Config().TelemetryTopic(), d),
							payload,
							mc.Config().Qos(),
							mc.Config().TelemetryRetain(),
						); err != nil {
							log.Printf(
								"device[%s]->mqttClient[%s]: cannot publish telemetry: %s",
								d.Config().Name(), mc.Config().Name(), err,
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
