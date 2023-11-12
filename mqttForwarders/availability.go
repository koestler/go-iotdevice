package mqttForwarders

import (
	"context"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
)

func runAvailabilityForwarder(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
) {
	mCfg := cfg.AvailabilityDevice()
	topic := cfg.AvailabilityDeviceTopic(dev.Name())

	go func(mc mqttClient.Client) {
		availChan := dev.SubscribeAvailableSendInitial(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case avail := <-availChan:
				payload := device.AvailabilityOfflineValue
				if avail {
					payload = device.AvailabilityOnlineValue
				}
				mc.Publish(
					topic,
					[]byte(payload),
					mCfg.Qos(),
					mCfg.Retain(),
				)
			}
		}
	}(mc)
}
