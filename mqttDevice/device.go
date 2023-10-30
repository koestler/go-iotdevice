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
	"sync"
	"sync/atomic"
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

	registerFilter    dataflow.RegisterFilterFunc
	subscriptionSetup atomic.Bool

	avail     map[string]bool
	availLock sync.Mutex
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
		registerFilter: dataflow.RegisterFilter(deviceConfig.RegisterFilter()),
		avail:          make(map[string]bool),
	}
}

func (c *DeviceStruct) Model() string {
	return "mqtt-" + c.mqttConfig.Kind().String()
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
				// parse struct message
				structMessage, err := parseStructPayload(m.Payload())
				if err != nil {
					log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse realtime payload: %s", c.Name(), mc.Name(), err)
					return
				}

				if c.Config().LogDebug() {
					log.Printf("mqttDevice[%s]->mqttClient[%s]: new struct received: %v", c.Name(), mc.Name(), structMessage)
				}

				// do not block the current go routine of the router and continue in a separate go routine
				go func() {
					if !c.subscriptionSetup.Load() {
						c.subscriptionSetup.Store(true)

						c.setupAvailabilitySubscription(mc, structMessage.AvailabilityTopics)
						if topic := structMessage.TelemetryTopic; len(topic) > 0 {
							c.setupTelemetrySubscription(mc, topic)
						}
						if topicTemplate := structMessage.RealtimeTopic; len(topicTemplate) > 0 {
							c.setupRealtimeSubscription(mc, topicTemplate)
						}
					}

					c.updateRegisters(structMessage.Registers)
				}()
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

func parseTelemetryPayload(payload []byte) (msg mqttForwarders.TelemetryMessage, err error) {
	err = json.Unmarshal(payload, &msg)
	return
}

func parseRealtimePayload(payload []byte) (msg mqttForwarders.RealtimeMessage, err error) {
	err = json.Unmarshal(payload, &msg)
	return
}

func (c *DeviceStruct) updateRegisters(registers []mqttForwarders.StructRegister) {
	structRegs := make([]dataflow.Register, 0, len(registers))
	for _, reg := range registers {
		r := StructRegister{reg}
		if c.registerFilter(r) {
			structRegs = append(structRegs, r)
		}
	}
	c.RegisterDb().Add(structRegs...)
}

func (c *DeviceStruct) setupAvailabilitySubscription(mc mqttClient.Client, availabilityTopics []string) {
	numbAvail := len(availabilityTopics)

	for _, topic := range availabilityTopics {
		if c.Config().LogDebug() {
			log.Printf("mqttDevice[%s]->mqttClient[%s]: subscribe to topic=%s", c.Name(), mc.Name(), topic)
		}

		mc.AddRoute(topic, func(m mqttClient.Message) {
			if c.Config().LogDebug() {
				log.Printf("mqttDevice[%s]->mqttClient[%s]: received availability topic=%s, msg=%s",
					c.Name(), mc.Name(), m.Topic(), m.Payload(),
				)
			}

			c.availLock.Lock()
			defer c.availLock.Unlock()

			c.avail[m.Topic()] = func(s string) bool {
				return s == "online"
			}(string(m.Payload()))

			c.SetAvailable(countTrue(c.avail) == numbAvail)
		})
	}
}

func countTrue(m map[string]bool) int {
	ret := 0
	for _, b := range m {
		if b {
			ret += 1
		}
	}
	return ret
}

func (c *DeviceStruct) setupTelemetrySubscription(mc mqttClient.Client, topic string) {
	if c.Config().LogDebug() {
		log.Printf("mqttDevice[%s]->mqttClient[%s]: subscribe to topic=%s", c.Name(), mc.Name(), topic)
	}

	mc.AddRoute(topic, func(m mqttClient.Message) {
		telemetryMessage, err := parseTelemetryPayload(m.Payload())
		if err != nil {
			log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse telemetry payload: %s", c.Name(), mc.Name(), err)
			return
		}

		// get register
		for _, register := range c.RegisterDb().GetAll() {
			switch register.RegisterType() {

			case dataflow.NumberRegister:
				if v, ok := telemetryMessage.NumericValues[register.Name()]; ok {
					c.StateStorage().Fill(dataflow.NewNumericRegisterValue(c.Name(), register, v.Value))
				}
			case dataflow.TextRegister:
				if v, ok := telemetryMessage.TextValues[register.Name()]; ok {
					c.StateStorage().Fill(dataflow.NewTextRegisterValue(c.Name(), register, v.Value))
				}
			case dataflow.EnumRegister:
				if v, ok := telemetryMessage.EnumValues[register.Name()]; ok {
					c.StateStorage().Fill(dataflow.NewEnumRegisterValue(c.Name(), register, v.EnumIdx))
				}
			default:
				if c.Config().LogDebug() {
					log.Printf("mqttDevice[%s]->mqttClient[%s]: register not found in telemetry message registerName=%v",
						c.Name(), mc.Name(), register.Name(),
					)
				}
			}
		}
	})
}

func (c *DeviceStruct) setupRealtimeSubscription(mc mqttClient.Client, topicTemplate string) {
	topic := strings.Replace(topicTemplate, "%RegisterName%", "+", 1)

	if c.Config().LogDebug() {
		log.Printf("mqttDevice[%s]->mqttClient[%s]: subscribe to topic=%s", c.Name(), mc.Name(), topic)
	}

	topicMatcher, err := createRealtimeTopicMatcher(topicTemplate)
	if err != nil {
		log.Printf("mqttDevice[%s]->mqttClient[%s]: invalid realtime topic: %s", c.Name(), mc.Name(), err)
		return
	}

	mc.AddRoute(topic, func(m mqttClient.Message) {
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

		// get register
		register, ok := c.RegisterDb().GetByName(registerName)

		if !ok {
			if c.Config().LogDebug() {
				log.Printf("mqttDevice[%s]->mqttClient[%s]: unknown register, registerName=%v, ignore",
					c.Name(), mc.Name(), registerName,
				)

			}
			return
		}

		switch register.RegisterType() {
		case dataflow.NumberRegister:
			if v := realtimeMessage.NumericValue; v != nil {
				c.StateStorage().Fill(dataflow.NewNumericRegisterValue(c.Name(), register, *v))
			}
		case dataflow.TextRegister:
			if v := realtimeMessage.TextValue; v != nil {
				c.StateStorage().Fill(dataflow.NewTextRegisterValue(c.Name(), register, *v))
			}
		case dataflow.EnumRegister:
			if v := realtimeMessage.EnumIdx; v != nil {
				c.StateStorage().Fill(dataflow.NewEnumRegisterValue(c.Name(), register, *v))
			}
		}
	})
}
