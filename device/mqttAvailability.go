package device

import (
	"context"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
)

func runAvailabilityForwarders(
	ctx context.Context,
	dev Device,
	mqttClientPool *pool.Pool[mqttClient.Client],
) {
	devCfg := dev.Config()

	for _, mc := range mqttClientPool.GetByNames(devCfg.ViaMqttClients()) {
		mcCfg := mc.Config()

		if !mcCfg.AvailabilityDeviceEnabled() {
			continue
		}

		topic := mcCfg.AvailabilityDeviceTopic(devCfg.Name())

		go func(mc mqttClient.Client) {
			availChan := dev.SubscribeAvailableSendInitial(ctx)
			for {
				select {
				case <-ctx.Done():
					return
				case avail := <-availChan:
					payload := availabilityOfflineValue
					if avail {
						payload = availabilityOnlineValue
					}
					mc.Publish(
						topic,
						[]byte(payload),
						mcCfg.Qos(),
						mcCfg.AvailabilityDeviceRetain(),
					)
				}
			}
		}(mc)
	}
}
