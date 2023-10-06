package device

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"time"
)

func RunMqttForwarders(
	ctx context.Context,
	dev Device,
	mqttClientPool *pool.Pool[mqttClient.Client],
	storage *dataflow.ValueStorage,
) {
	deviceName := dev.Config().Name()
	deviceFilter := func(v dataflow.Value) bool {
		return v.DeviceName() == deviceName
	}

	runAvailabilityForwarders(ctx, dev, mqttClientPool)
	runStructureForwarders(ctx, dev, mqttClientPool)
	runTelemetryForwarders(ctx, dev, mqttClientPool, storage, deviceFilter)
	runRealtimeForwarders(ctx, dev, mqttClientPool, storage, deviceFilter)
}

func timeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}
