package mqttForwarders

import (
	"context"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
)

func runAvailabilityForwarder(
	ctx context.Context,
	dev device.Device,
	mc mqttClient.Client,
) {
	devCfg := dev.Config()

	mCfg := mc.Config().AvailabilityDevice()
	topic := mc.Config().AvailabilityDeviceTopic(devCfg.Name())

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
