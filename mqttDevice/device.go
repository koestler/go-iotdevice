package mqttDevice

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/mqttForwarders"
	"github.com/koestler/go-iotdevice/pool"
	"github.com/koestler/go-iotdevice/types"
	"log"
	"regexp"
	"strings"
)

type Config interface {
	Kind() types.MqttDeviceKind
	MqttClients() []string
	MqttTopics() []string
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
	mCfg := c.mqttConfig

	if mCfg.Kind() != types.MqttDeviceGoIotdeviceKind {
		log.Printf("mqttDevice[%s]: unsuported type: %s", c.Name(), mCfg.Kind().String())
		return
	}

	for _, mc := range c.mqttClientPool.GetByNames(mCfg.MqttClients()) {
		for _, topic := range mCfg.MqttTopics() {
			if c.Config().LogDebug() {
				log.Printf("mqttDevice[%s]->mqttClient[%s]: subscribe to topic=%s", c.Name(), mc.Name(), topic)
			}

			mc.AddRoute(topic, func(m mqttClient.Message) {
				log.Printf("mqttDevice[%s]->mqttClient[%s]: received in struct handler topic=%s, payload=%s",
					c.Name(), mc.Name(), m.Topic(), m.Payload(),
				)

				// parse struct message
				structMessage, err := parseStructPayload(m.Payload())
				if err != nil {
					log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse realtime payload: %s", c.Name(), mc.Name(), err)
					return
				}

				if c.Config().LogDebug() {
					log.Printf("mqttDevice[%s]->mqttClient[%s]: new struct received: %v", c.Name(), mc.Name(), structMessage)
				}

				if topicTemplate := structMessage.RealtimeTopic; len(topicTemplate) > 0 {
					topic := strings.Replace(topicTemplate, "%RegisterName%", "+", 1)

					if c.Config().LogDebug() {
						log.Printf("mqttDevice[%s]->mqttClient[%s]: subscribe to topic=%s", c.Name(), mc.Name(), topic)
					}

					topicMatcher, err := createRealtimeTopicMatcher(topicTemplate)
					if err != nil {
						log.Printf("mqttDevice[%s]->mqttClient[%s]: invalid realtime topic: %s", c.Name(), mc.Name(), err)
						return
					}
					go func() {
						mc.AddRoute(topic, func(m mqttClient.Message) {
							log.Printf("mqttDevice[%s]->mqttClient[%s]: received in realtime handler"+
								" topic=%s, payload=%s",
								c.Name(), mc.Name(), m.Topic(), m.Payload(),
							)

							registerName, err := parseRealtimeTopic(topicMatcher, m.Topic())
							if err != nil {
								log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse realtime topic: %s", c.Name(), mc.Name(), err)
								return
							}
							realtimeMessage, err := parseRealtimePayload(m.Payload())
							if err != nil {
								log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse realtime payload: %s", c.Name(), mc.Name(), err)
								return
							}

							if c.Config().LogDebug() {
								log.Printf("mqttDevice[%s]->mqttClient[%s]: received registerName=%v, msg=%v",
									c.Name(), mc.Name(), registerName, realtimeMessage,
								)
							}

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

					}()

				}
			})
		}
	}

	<-ctx.Done()
	return nil, false
}

func parseStructPayload(payload []byte) (msg mqttForwarders.StructureMessage, err error) {
	err = json.Unmarshal(payload, &msg)
	return
}

func parseRealtimeTopic(matcher *regexp.Regexp, topic string) (registerName string, err error) {
	matches := matcher.FindStringSubmatch(topic)
	if matches == nil {
		err = fmt.Errorf("topic='%s' does not match", topic)
	} else {
		registerName = matches[1]
	}

	return
}

func createRealtimeTopicMatcher(topicTemplate string) (matcher *regexp.Regexp, err error) {
	// create regexp to match against
	regNameExpr := "([^\\/]+)"

	// must not have anything before / after
	expr := "^" + strings.Replace(regexp.QuoteMeta(topicTemplate), "%RegisterName%", regNameExpr, 1) + "$"

	matcher, err = regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("cannot create topic machter: invalid regexp: %s", err)
	}

	return
}

func parseRealtimePayload(payload []byte) (msg mqttForwarders.RealtimeMessage, err error) {
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
	return "mqtt-" + c.mqttConfig.Kind().String()
}
