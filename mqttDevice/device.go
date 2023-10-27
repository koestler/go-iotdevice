package mqttDevice

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/mqttForwarders"
	"github.com/koestler/go-iotdevice/pool"
	"log"
	"strings"
)

type Config interface {
	MqttTopics() []string
	MqttClients() []string
}

type DeviceStruct struct {
	device.State
	mqttConfig Config

	mqttClientPool *pool.Pool[mqttClient.Client]
}

func NewDevice(
	deviceConfig device.Config,
	mqttConfig Config,
	stateStorage *dataflow.ValueStorage,
	mqttClientPool *pool.Pool[mqttClient.Client],
) *DeviceStruct {
	return &DeviceStruct{
		State: device.NewState(
			deviceConfig,
			stateStorage,
		),
		mqttConfig:     mqttConfig,
		mqttClientPool: mqttClientPool,
	}
}

func (c *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	// setup mqtt listeners
	counter := 0
	for _, mc := range c.mqttClientPool.GetByNames(c.mqttConfig.MqttClients()) {
		for _, topic := range c.mqttConfig.MqttTopics() {
			log.Printf("mqttDevice[%s] subscribe to mqttClient=%s topic=%s", c.Name(), mc.Name(), topic)
			mc.AddRoute(topic, func(m mqttClient.Message) {
				registerName, err := parseTopic(m.Topic())
				if err != nil {
					log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse topic: %s", c.Name(), mc.Name(), err)
					return
				}
				realtimeMessage, err := parsePayload(m.Payload())
				if err != nil {
					log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse payload: %s", c.Name(), mc.Name(), err)
					return
				}

				log.Printf("mqttDevice[%s]->mqttClient[%s]: received registerName=%v, msg=%v", c.Name(), mc.Name(), registerName, realtimeMessage)

				/*
					register := c.addIgnoreRegister(registerName, realtimeMessage)
					if register != nil {
						if v := realtimeMessage.NumericValue; v != nil {
							c.StateStorage().Fill(dataflow.NewNumericRegisterValue(c.Name(), register, *v))
						} else if v := realtimeMessage.TextValue; v != nil {
							c.StateStorage().Fill(dataflow.NewTextRegisterValue(c.Name(), register, *v))
						}
					}
				*/
			})
			counter += 1
		}
	}

	if counter < 1 {
		log.Printf("mqttDevice[%s]: no listener was starrted", c.Name())
	}

	<-ctx.Done()
	return nil, false
}

func parseTopic(topic string) (registerName string, err error) {
	registerName = topic[strings.LastIndex(topic, "/")+1:]
	if len(registerName) < 1 {
		err = fmt.Errorf("cannot extract registerName from topic='%s'", topic)
	}

	return
}

func parsePayload(payload []byte) (msg mqttForwarders.RealtimeMessage, err error) {
	err = json.Unmarshal(payload, &msg)
	return
}

/*
func (c *DeviceStruct) addIgnoreRegister(registerName string, msg device.RealtimeMessage) dataflow.Register {
	// check if this register exists already and the properties are still the same
	if r := c.RegisterDb().GetByName(registerName); r != nil {
		if r.Category() == msg.Category &&
			r.Description() == msg.Description &&
			r.Unit() == msg.Unit &&
			r.Sort() == msg.Sort {
			return r
		}
	}

	// check if register is on ignore list
	if device.IsExcluded(registerName, msg.Category, c.Config()) {
		return nil
	}

	// create new register
	var r dataflow.Register
	var registerType = dataflow.TextRegister

	if msg.NumericValue != nil {
		registerType = dataflow.NumberRegister
	}

	r = dataflow.NewRegisterStruct(
		msg.Category,
		registerName,
		msg.Description,
		registerType,
		nil,
		msg.Unit,
		msg.Sort,
		false,
	)

	// add the register into the list
	c.RegisterDb().Add(r)

	return r
}
*/

func (c *DeviceStruct) Model() string {
	return "mqtt"
}
