package mqttDevice

import (
	"encoding/json"
	"log"

	"github.com/koestler/go-iotdevice/v3/internal/dataflow"
	"github.com/koestler/go-iotdevice/v3/internal/mqttClient"
	"github.com/koestler/go-iotdevice/v3/internal/mqttForwarders"
)

func parseTelemetryPayload(payload []byte) (msg mqttForwarders.TelemetryMessage, err error) {
	err = json.Unmarshal(payload, &msg)
	return
}

func (f *forwarder) setupTelemetrySubscription(topic string) (remove func()) {
	return f.Client.AddRoute(
		topic,
		telemetryHandler(
			f.Device.Config().LogDebug(),
			f.Device.Name(), f.Client.Name(),
			f.Device.RegisterDb().GetAll,
			f.Device.StateStorage(),
		),
	)
}

func telemetryHandler(
	logDebug bool,
	deviceName, clientName string,
	allRegisters func() []dataflow.RegisterStruct,
	fillable dataflow.Fillable,
) func(message mqttClient.Message) {
	return func(m mqttClient.Message) {
		telemetryMessage, err := parseTelemetryPayload(m.Payload())
		if err != nil {
			log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse telemetry payload: %s", deviceName, clientName, err)
			return
		}

		// loop over all registers
		for _, register := range allRegisters() {
			switch register.RegisterType() {

			case dataflow.NumberRegister:
				if v, ok := telemetryMessage.NumericValues[register.Name()]; ok {
					fillable.Fill(dataflow.NewNumericRegisterValue(deviceName, register, v.Value))
				}
			case dataflow.TextRegister:
				if v, ok := telemetryMessage.TextValues[register.Name()]; ok {
					fillable.Fill(dataflow.NewTextRegisterValue(deviceName, register, v.Value))
				}
			case dataflow.EnumRegister:
				if v, ok := telemetryMessage.EnumValues[register.Name()]; ok {
					fillable.Fill(dataflow.NewEnumRegisterValue(deviceName, register, v.EnumIdx))
				}
			}
		}
	}
}
