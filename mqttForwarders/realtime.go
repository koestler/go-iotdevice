package mqttForwarders

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"log"
	"time"
)

type RealtimeMessage struct {
	NumericValue *float64 `json:"NumVal,omitempty"`
	TextValue    *string  `json:"TextVal,omitempty"`
	EnumIdx      *int     `json:"EnumIdx,omitempty"`
}

func runRealtimeForwarder(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	storage *dataflow.ValueStorage,
	registerFilter dataflow.RegisterFilterConf,
) {
	// start mqtt forwarder for realtime messages (send data as soon as it arrives) output
	mCfg := cfg.Realtime()

	filter := createDeviceAndRegisterValueFilter(dev, registerFilter)

	// immediate mode: Interval is set to zero
	// -> send values immediately when they change
	// delayed update mode: Interval > 0 and Repeat false
	// -> have a timer, whenever it ticks, send the newest version of the changed values
	// periodic full mode: Interval > 0 and Repeat true
	// -> have a timer, whenever it ticks, send all values
	if mCfg.Interval() <= 0 {
		go realtimeImmediateModeRoutine(ctx, cfg, dev, mc, storage, filter)
	} else {
		go realtimeDelayedUpdateModeRoutine(ctx, cfg, dev, mc, storage, filter)
	}
}

func realtimeImmediateModeRoutine(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	storage *dataflow.ValueStorage,
	filter func(v dataflow.Value) bool,
) {
	if cfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->realtime: start immediate mode",
			dev.Name(), mc.Name(),
		)
		defer log.Printf(
			"device[%s]->mqttClient[%s]->realtime: exit immediate mode",
			dev.Name(), mc.Name(),
		)
	}

	subscription := storage.SubscribeSendInitial(ctx, filter)
	// for loop ends when subscription is canceled and closes its output chan
	for value := range subscription.Drain() {
		publishRealtimeMessage(cfg, mc, dev.Name(), value)
	}
}

func realtimeDelayedUpdateModeRoutine(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	storage *dataflow.ValueStorage,
	filter func(v dataflow.Value) bool,
) {
	realtimeInterval := cfg.Realtime().Interval()

	if cfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->realtime: start delayed update mode, send every %s",
			dev.Name(), mc.Name(), realtimeInterval,
		)
		defer log.Printf(
			"device[%s]->mqttClient[%s]->realtime: exit delayed update mode",
			dev.Name(), mc.Name(),
		)
	}

	ticker := time.NewTicker(realtimeInterval)
	defer ticker.Stop()

	updates := make(map[string]dataflow.Value)

	subscription := storage.SubscribeSendInitial(ctx, filter)
	for {
		select {
		case <-ctx.Done():
			return
		case value := <-subscription.Drain():
			// new value received, save newest version per register name
			updates[value.Register().Name()] = value
		case <-ticker.C:
			if cfg.LogDebug() {
				log.Printf(
					"device[%s]->mqttClient[%s]->realtime: tick: send updates",
					dev.Name(), mc.Name(),
				)
			}
			for _, v := range updates {
				publishRealtimeMessage(cfg, mc, dev.Name(), v)
			}
			clear(updates)
		}
	}
}

func publishRealtimeMessage(cfg Config, mc mqttClient.Client, devName string, value dataflow.Value) {
	mCfg := cfg.Realtime()

	if cfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->realtime: send: %s",
			devName, mc.Name(), value,
		)
	}

	if payload, err := json.Marshal(convertValueToRealtimeMessage(value)); err != nil {
		log.Printf(
			"device[%s]->mqttClient[%s]->realtime: cannot generate message: %s",
			devName, mc.Name(), err,
		)
	} else {
		mc.Publish(
			cfg.RealtimeTopic(devName, value.Register().Name()),
			payload,
			mCfg.Qos(),
			mCfg.Retain(),
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
