package device

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"log"
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
	// start mqtt forwarders for realtime messages (send data as soon as it arrives) output
	for _, mc := range mqttClientPool.GetByNames(dev.Config().RealtimeViaMqttClients()) {
		mcCfg := mc.Config()
		if !mcCfg.RealtimeEnabled() {
			continue
		}

		// immediate mode: Interval is set to zero
		// -> send values immediately when they change
		// delayed update mode: Interval > 0 and Repeat false
		// -> have a timer, whenever it ticks, send the newest version of the changed values
		// periodic full mode: Interval > 0 and Repeat true
		// -> have a timer, whenever it ticks, send all values
		if mcCfg.RealtimeInterval() <= 0 {
			go immediateModeRoutine(ctx, dev, mc, storage, deviceFilter)
		} else if !mcCfg.RealtimeRepeat() {
			go delayedUpdateModeRoutine(ctx, dev, mc, storage, deviceFilter)
		} else {
			go periodicFullModeRoutine(ctx, dev, mc, storage, deviceFilter)
		}
	}
}

func immediateModeRoutine(
	ctx context.Context,
	dev Device,
	mc mqttClient.Client,
	storage *dataflow.ValueStorage,
	deviceFilter func(v dataflow.Value) bool,
) {
	mcCfg := mc.Config()
	devCfg := dev.Config()

	if devCfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->realtime: start immediate mode",
			devCfg.Name(), mcCfg.Name(),
		)
		defer func() {
			log.Printf(
				"device[%s]->mqttClient[%s]->realtime: exit immediate mode",
				devCfg.Name(), mcCfg.Name(),
			)
		}()
	}

	subscription := storage.SubscribeSendInitial(ctx, deviceFilter)
	// for loop ends when subscription is canceled and closes its output chan
	for value := range subscription.Drain() {
		publishRealtimeMessage(mc, devCfg, value)
	}
}

func delayedUpdateModeRoutine(
	ctx context.Context,
	dev Device,
	mc mqttClient.Client,
	storage *dataflow.ValueStorage,
	deviceFilter func(v dataflow.Value) bool,
) {
	mcCfg := mc.Config()
	devCfg := dev.Config()
	realtimeInterval := mcCfg.RealtimeInterval()

	if devCfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->realtime: start delayed update mode, send every %s",
			devCfg.Name(), mcCfg.Name(), realtimeInterval,
		)
		defer func() {
			log.Printf(
				"device[%s]->mqttClient[%s]->realtime: exit delayed update mode",
				devCfg.Name(), mcCfg.Name(),
			)
		}()
	}

	ticker := time.NewTicker(realtimeInterval)
	defer ticker.Stop()

	updates := make(map[string]dataflow.Value)

	subscription := storage.SubscribeSendInitial(ctx, deviceFilter)
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
			updates = make(map[string]dataflow.Value)
		}
	}
}

func periodicFullModeRoutine(
	ctx context.Context,
	dev Device,
	mc mqttClient.Client,
	storage *dataflow.ValueStorage,
	deviceFilter func(v dataflow.Value) bool,
) {
	mcCfg := mc.Config()
	devCfg := dev.Config()
	realtimeInterval := mcCfg.RealtimeInterval()

	if devCfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->realtime: start periodic full mode, send every %s",
			devCfg.Name(), mcCfg.Name(), realtimeInterval,
		)
		defer func() {
			log.Printf(
				"device[%s]->mqttClient[%s]->realtime: exit periodic full mode",
				devCfg.Name(), mcCfg.Name(),
			)
		}()
	}

	ticker := time.NewTicker(realtimeInterval)
	defer ticker.Stop()

	avail, availChan := dev.SubscribeAvailable(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case avail = <-availChan:
		case <-ticker.C:
			if !avail {
				// do not send messages when device is disconnected
				continue
			}

			if devCfg.LogDebug() {
				log.Printf(
					"device[%s]->mqttClient[%s]->realtime: tick: send everything",
					devCfg.Name(), mcCfg.Name(),
				)
			}

			values := storage.GetStateFiltered(deviceFilter)
			for _, v := range values {
				publishRealtimeMessage(mc, devCfg, v)
			}
		}
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
			mcCfg.RealtimeTopic(devConfig.Name(), value.Register().Name()),
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
