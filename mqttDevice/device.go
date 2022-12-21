package mqttDevice

import (
	"encoding/json"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"log"
	"strings"
	"sync"
	"time"
)

type Config interface {
	MqttTopics() []string
	MqttClients() []string
}

type DeviceStruct struct {
	deviceConfig device.Config
	mqttConfig   Config

	source *dataflow.Source

	registers        map[string]dataflow.Register
	registersMutex   sync.RWMutex
	lastUpdated      time.Time
	lastUpdatedMutex sync.RWMutex

	shutdown chan struct{}
}

func RunDevice(
	deviceConfig device.Config,
	mqttConfig Config,
	storage *dataflow.ValueStorageInstance,
	mqttClientPool *mqttClient.ClientPool,
) (device device.Device, err error) {
	// setup output chain
	output := make(chan dataflow.Value, 128)
	source := dataflow.CreateSource(output)
	// pipe all data to next stage
	source.Append(storage)

	c := &DeviceStruct{
		deviceConfig: deviceConfig,
		mqttConfig:   mqttConfig,
		source:       source,
		registers:    make(map[string]dataflow.Register),
		shutdown:     make(chan struct{}),
	}

	// setup mqtt listeners
	counter := 0
	for _, mc := range mqttClientPool.GetClientsByNames(mqttConfig.MqttClients()) {
		for _, topic := range mqttConfig.MqttTopics() {
			log.Printf("mqttDevice[%s] subscribe to mqttClient=%s topic=%s", deviceConfig.Name(), mc.Config().Name(), topic)
			mc.AddRoute(topic, func(m mqttClient.Message) {
				registerName, err := parseTopic(m.Topic())
				if err != nil {
					log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse topic: %s", deviceConfig.Name(), mc.Config().Name(), err)
					return
				}
				realtimeMessage, err := parsePayload(m.Payload())
				if err != nil {
					log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse payload: %s", deviceConfig.Name(), mc.Config().Name(), err)
					return
				}

				register := c.addIgnoreRegister(registerName, realtimeMessage)
				if v := realtimeMessage.NumericValue; v != nil {
					output <- dataflow.NewNumericRegisterValue(deviceConfig.Name(), register, *v)
				} else if v := realtimeMessage.TextValue; v != nil {
					output <- dataflow.NewTextRegisterValue(deviceConfig.Name(), register, *v)
				}
				c.SetLastUpdatedNow()

			})
			counter += 1
		}
	}

	if counter < 1 {
		return nil, fmt.Errorf("no listener was starrted")
	}

	return c, nil
}

func parseTopic(topic string) (registerName string, err error) {
	registerName = topic[strings.LastIndex(topic, "/")+1:]
	if len(registerName) < 1 {
		err = fmt.Errorf("cannot extract registerName from topic='%s'", topic)
	}

	return
}

func parsePayload(payload []byte) (msg device.RealtimeMessage, err error) {
	err = json.Unmarshal(payload, &msg)
	return
}

func (c *DeviceStruct) Config() device.Config {
	return c.deviceConfig
}

func (c *DeviceStruct) ShutdownChan() chan struct{} {
	return c.shutdown
}

func (c *DeviceStruct) Registers() dataflow.Registers {
	c.registersMutex.RLock()
	defer c.registersMutex.RUnlock()

	ret := make(dataflow.Registers, len(c.registers))
	i := 0
	for _, r := range c.registers {
		ret[i] = r
		i += 1
	}
	return ret
}

func (c *DeviceStruct) addIgnoreRegister(registerName string, msg device.RealtimeMessage) dataflow.Register {
	// check if this register exists already and the properties are still the same
	c.registersMutex.RLock()
	if r, ok := c.registers[registerName]; ok {
		if r.Category() == msg.Category &&
			r.Description() == msg.Description &&
			r.Unit() == msg.Unit &&
			r.Sort() == msg.Sort {
			c.registersMutex.RUnlock()
			return r
		}
	}
	c.registersMutex.RUnlock()

	// add the register into the list
	c.registersMutex.Lock()
	defer c.registersMutex.Unlock()

	var r dataflow.Register

	if msg.NumericValue != nil {
		unit := ""
		if msg.Unit != nil {
			unit = *msg.Unit
		}

		r = dataflow.CreateNumberRegisterStruct(
			msg.Category,
			registerName,
			msg.Description,
			0,
			false,
			true,
			1,
			unit,
			msg.Sort,
		)
	} else {
		r = dataflow.CreateTextRegisterStruct(
			msg.Category,
			registerName,
			msg.Description,
			0,
			false,
			msg.Sort,
		)
	}

	c.registers[registerName] = r
	return r
}

func (c *DeviceStruct) SetLastUpdatedNow() {
	c.lastUpdatedMutex.Lock()
	defer c.lastUpdatedMutex.Unlock()
	c.lastUpdated = time.Now()
}

func (c *DeviceStruct) LastUpdated() time.Time {
	c.lastUpdatedMutex.RLock()
	defer c.lastUpdatedMutex.RUnlock()
	return c.lastUpdated
}

func (c *DeviceStruct) Model() string {
	return "mqtt"
}

func (c *DeviceStruct) Shutdown() {
	close(c.shutdown)
	log.Printf("mqttDevice[%s]: shutdown completed", c.deviceConfig.Name())
}
