package mqttForwarders

import (
	"context"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"github.com/koestler/go-iotdevice/v3/mqttClient"
	"log"
)

type CommandMessage struct {
	NumericValue *float64 `json:"NumVal,omitempty"`
	TextValue    *string  `json:"TextVal,omitempty"`
	EnumIdx      *int     `json:"EnumIdx,omitempty"`
}

func runCommandForwarder(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	commandStorage *dataflow.ValueStorage,
	filterConf dataflow.RegisterFilterConf,
) {
	filter := createCommandableAndRegisterValueFilter(filterConf)
	go commandRoutine(ctx, cfg, dev, mc, commandStorage, filter)
}

func commandRoutine(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	commandStorage *dataflow.ValueStorage,
	filter dataflow.RegisterFilterFunc,
) {
	regSubscription := dev.RegisterDb().Subscribe(ctx, filter)

	for {
		select {
		case <-ctx.Done():
			return
		case reg := <-regSubscription:
			setupCommandSubscription(cfg, dev, mc, commandStorage, reg)
		}
	}
}

func setupCommandSubscription(
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	commandStorage *dataflow.ValueStorage,
	register dataflow.Register,
) {
	topic := cfg.CommandTopic(dev.Name(), register.Name())
	logDebug := cfg.LogDebug()

	if logDebug {
		log.Printf("mqttDevice[%s]->mqttClient[%s]->command: subscribe to topic=%s", mc.Name(), dev.Name(), topic)
	}

	registerDb := dev.RegisterDb()
	deviceName := dev.Name()

	register, ok := registerDb.GetByName(register.Name())
	if !ok {
		log.Printf("mqttDevice[%s]->mqttClient[%s]->command: unknown register, registerName=%s", mc.Name(), deviceName, register.Name())
		return
	}

	mc.AddRoute(topic, func(m mqttClient.Message) {
		msg, err := parseCommandMessagePayload(m.Payload())
		if err != nil {
			log.Printf("mqttDevice[%s]->mqttClient[%s]->command: cannod parse message: %s", mc.Name(), dev.Name(), err)
			return
		}

		switch register.RegisterType() {
		case dataflow.NumberRegister:
			if v := msg.NumericValue; v != nil {
				rv := dataflow.NewNumericRegisterValue(deviceName, register, *v)
				commandStorage.Fill(rv)
				if logDebug {
					log.Printf("mqttDevice[%s]->mqttClient[%s]->command: send NumericValue deviceName=%s: %s", mc.Name(), dev.Name(), deviceName, rv.String())
				}
				return
			}
		case dataflow.TextRegister:
			if v := msg.TextValue; v != nil {
				rv := dataflow.NewTextRegisterValue(deviceName, register, *v)
				commandStorage.Fill(rv)
				if logDebug {
					log.Printf("mqttDevice[%s]->mqttClient[%s]->command: send TextValue deviceName=%s: %s", mc.Name(), dev.Name(), deviceName, rv.String())
				}
				return
			}
		case dataflow.EnumRegister:
			if v := msg.EnumIdx; v != nil {
				rv := dataflow.NewEnumRegisterValue(deviceName, register, *v)
				commandStorage.Fill(rv)
				if logDebug {
					log.Printf("mqttDevice[%s]->mqttClient[%s]->command: send EnumValue deviceName=%s: %s", mc.Name(), dev.Name(), deviceName, rv.String())
				}
				return
			}
		}

		log.Printf("mqttDevice[%s]->mqttClient[%s]->command: invalid command message: %#v", mc.Name(), dev.Name(), msg)
	})
}

func parseCommandMessagePayload(payload []byte) (msg CommandMessage, err error) {
	err = json.Unmarshal(payload, &msg)
	return
}
