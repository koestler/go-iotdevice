package mqttForwarders

import (
	"context"
	"github.com/koestler/go-iotdevice/config"
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
	dev device.Device,
	mc mqttClient.Client,
	storage *dataflow.ValueStorage,
	registerFilter config.RegisterFilterConfig,
) {
	// start mqtt forwarder for realtime messages (send data as soon as it arrives) output
	mCfg := mc.Config().Realtime()

	filter := createDeviceAndRegisterValueFilter(dev, registerFilter)

	// immediate mode: Interval is set to zero
	// -> send values immediately when they change
	// delayed update mode: Interval > 0 and Repeat false
	// -> have a timer, whenever it ticks, send the newest version of the changed values
	// periodic full mode: Interval > 0 and Repeat true
	// -> have a timer, whenever it ticks, send all values
	if mCfg.Interval() <= 0 {
		go realtimeImmediateModeRoutine(ctx, dev, mc, storage, filter)
	} else {
		go realtimeDelayedUpdateModeRoutine(ctx, dev, mc, storage, filter)
	}
}

func realtimeImmediateModeRoutine(
	ctx context.Context,
	dev device.Device,
	mc mqttClient.Client,
	storage *dataflow.ValueStorage,
	filter func(v dataflow.Value) bool,
) {
	mcCfg := mc.Config()
	devCfg := dev.Config()

	if devCfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->realtime: start immediate mode",
			devCfg.Name(), mcCfg.Name(),
		)
		defer log.Printf(
			"device[%s]->mqttClient[%s]->realtime: exit immediate mode",
			devCfg.Name(), mcCfg.Name(),
		)
	}

	subscription := storage.SubscribeSendInitial(ctx, filter)
	// for loop ends when subscription is canceled and closes its output chan
	for value := range subscription.Drain() {
		publishRealtimeMessage(mc, devCfg, value)
	}
}

func realtimeDelayedUpdateModeRoutine(
	ctx context.Context,
	dev device.Device,
	mc mqttClient.Client,
	storage *dataflow.ValueStorage,
	filter func(v dataflow.Value) bool,
) {
	mcCfg := mc.Config()
	devCfg := dev.Config()
	realtimeInterval := mcCfg.Realtime().Interval()

	if devCfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->realtime: start delayed update mode, send every %s",
			devCfg.Name(), mcCfg.Name(), realtimeInterval,
		)
		defer log.Printf(
			"device[%s]->mqttClient[%s]->realtime: exit delayed update mode",
			devCfg.Name(), mcCfg.Name(),
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
			if devCfg.LogDebug() {
				log.Printf(
					"device[%s]->mqttClient[%s]->realtime: tick: send updates",
					devCfg.Name(), mcCfg.Name(),
				)
			}
			for _, v := range updates {
				publishRealtimeMessage(mc, devCfg, v)
			}
			clear(updates)
		}
	}
}

func publishRealtimeMessage(mc mqttClient.Client, devConfig device.Config, value dataflow.Value) {
	mcCfg := mc.Config()
	mCfg := mcCfg.Realtime()

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
			mc.Config().RealtimeTopic(devConfig.Name(), value.Register().Name()),
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
