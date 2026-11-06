package mqttDevice

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/koestler/go-iotdevice/v3/internal/dataflow"
	"github.com/koestler/go-iotdevice/v3/internal/mqttForwarders"
)

func (f *forwarder) setupCommandForwarder(
	topicTemplate string,
	registers []mqttForwarders.StructRegister,
) (remove func()) {
	deviceName := f.Device.Name()
	commandStorage := f.Device.commandStorage
	commandRegister := filterCommandRegisters(registers)

	ctx, cancel := context.WithCancel(context.Background())

	subscription := commandStorage.SubscribeSendInitial(ctx, func(curr, prev dataflow.Value) bool {
		if prev.Equals(curr) || curr.DeviceName() != deviceName {
			return false
		}

		_, ok := commandRegister[curr.Register().Name()]
		return ok
	})

	go commandForwarder(
		f.Device.Config().LogDebug(),
		deviceName, f.Client.Name(),
		subscription,
		topicTemplate,
		f.Client.Publish,
		commandStorage,
	)

	return cancel
}

func commandForwarder(
	logDebug bool,
	deviceName, clientName string,
	subscription dataflow.ValueSubscription,
	topicTemplate string,
	publish func(topic string, payload []byte, qos byte, retain bool),
	commandStorage dataflow.Fillable,
) {
	for command := range subscription.Drain() {
		topic := strings.Replace(topicTemplate, "%RegisterName%", command.Register().Name(), 1)

		var msg mqttForwarders.CommandMessage

		if numeric, ok := command.(dataflow.NumericRegisterValue); ok {
			v := numeric.Value()
			msg.NumericValue = &v
		} else if text, ok := command.(dataflow.TextRegisterValue); ok {
			v := text.Value()
			msg.TextValue = &v
		} else if enum, ok := command.(dataflow.EnumRegisterValue); ok {
			v := enum.EnumIdx()
			msg.EnumIdx = &v
		} else {
			continue
		}

		payload, err := json.Marshal(msg)
		if err != nil {
			log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot generate command message: %s",
				deviceName, clientName, err,
			)
			continue
		}

		publish(topic, payload, 1, false)
		if logDebug {
			log.Printf("mqttDevice[%s]->mqttClient[%s]: sent command message to topic=%s payload=%s",
				deviceName, clientName, topic, string(payload),
			)
		}

		// reset the command; this allows the same command (e.g. toggle) to be sent again
		commandStorage.Fill(dataflow.NewNullRegisterValue(deviceName, command.Register()))
	}
}

func filterCommandRegisters(inp []mqttForwarders.StructRegister) (oup map[string]struct{}) {
	oup = make(map[string]struct{})
	for _, r := range inp {
		if r.Writable {
			oup[r.Name] = struct{}{}
		}
	}
	return
}
