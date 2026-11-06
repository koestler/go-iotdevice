package mqttDevice

import (
	"context"
	"log"

	"github.com/koestler/go-iotdevice/v3/internal/dataflow"
	"github.com/koestler/go-iotdevice/v3/internal/device"
	"github.com/koestler/go-iotdevice/v3/internal/mqttClient"
	"github.com/koestler/go-iotdevice/v3/internal/mqttForwarders"
	"github.com/koestler/go-iotdevice/v3/internal/pool"
	"github.com/koestler/go-iotdevice/v3/internal/types"
)

type Config interface {
	Kind() types.MqttDeviceKind
	MqttClientTopics() map[string][]string
}

type DeviceStruct struct {
	device.State
	mqttConfig Config

	mqttClientPool *pool.Pool[mqttClient.Client]
	commandStorage *dataflow.ValueStorage

	registerFilter dataflow.RegisterFilterFunc
}

func NewDevice(
	deviceConfig device.Config,
	mqttConfig Config,
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
	mqttClientPool *pool.Pool[mqttClient.Client],
) *DeviceStruct {
	return &DeviceStruct{
		State: device.NewState(
			deviceConfig,
			stateStorage,
		),
		mqttConfig:     mqttConfig,
		mqttClientPool: mqttClientPool,
		commandStorage: commandStorage,
		registerFilter: dataflow.RegisterFilter(deviceConfig.Filter()),
	}
}

func (c *DeviceStruct) Model() string {
	return "mqtt-" + c.mqttConfig.Kind().String()
}

func (c *DeviceStruct) Run(ctx context.Context) (err error, immediateError bool) {
	mCfg := c.mqttConfig

	if mCfg.Kind() != types.MqttGoIotdeviceV3Kind {
		log.Printf("mqttDevice[%s]: unsupported type: %s", c.Name(), mCfg.Kind().String())
		return
	}

	for mqttClientName, topics := range mCfg.MqttClientTopics() {
		mc, ok := c.mqttClientPool.GetByName(mqttClientName)
		if !ok {
			continue
		}

		runForwarder(ctx, c, c.Config().LogDebug(), mc, topics)
	}

	<-ctx.Done()
	return nil, false
}

type forwarder struct {
	Device *DeviceStruct

	LogDebug bool

	Client       mqttClient.Client
	StructTopics []string

	structMessage chan mqttForwarders.StructureMessage
	availMessage  chan availabilityMessage
}

func runForwarder(
	ctx context.Context,
	c *DeviceStruct,
	logDebug bool,
	client mqttClient.Client,
	structTopics []string,
) {
	f := &forwarder{
		Device: c,

		LogDebug:     logDebug,
		Client:       client,
		StructTopics: structTopics,

		structMessage: make(chan mqttForwarders.StructureMessage),
		availMessage:  make(chan availabilityMessage),
	}

	// subscribe to structure updates
	for _, topic := range f.StructTopics {
		if f.LogDebug {
			log.Printf("mqttDevice[%s]->mqttClient[%s]: subscribe to topic=%s", f.Device.Name(), f.Client.Name(), topic)
		}

		// subscribe to the struct topics
		sc := f.structMessage
		f.Client.AddRoute(topic, func(m mqttClient.Message) {
			s, err := parseStructPayload(m.Payload())
			if err != nil {
				if logDebug {
					log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse structure payload: %s", f.Device.Name(), f.Client.Name(), err)
				}
				return
			}
			sc <- s
		})
	}

	go f.run(ctx)
}

type forwarderState struct {
	sm mqttForwarders.StructureMessage

	avail map[string]bool // topic -> availability

	remove func()
}

func (f *forwarder) run(ctx context.Context) {
	s := forwarderState{
		avail:  make(map[string]bool),
		remove: func() {},
	}

	for {
		select {
		case <-ctx.Done():
			s.remove()
			return
		case m := <-f.structMessage:
			if equalStructure(s.sm, m) {
				continue
			}

			if f.LogDebug {
				log.Printf("mqttDevice[%s]->mqttClient[%s]: new struct received: %v", f.Device.Name(), f.Client.Name(), m)
			}
			s.sm = m

			removeAvail := f.setupAvailabilitySubscription(m.AvailabilityTopics)
			removeTelemetry := f.setupTelemetrySubscription(m.TelemetryTopic)
			removeRealtime := f.setupRealtimeSubscription(m.RealtimeTopic)
			removeCommand := f.setupCommandForwarder(m.CommandTopic, m.Registers)
			f.Device.updateRegisters(m.Registers)

			s.remove() // remove old subscriptions
			s.remove = func() {
				removeAvail()
				removeTelemetry()
				removeRealtime()
				removeCommand()
			}
		case m := <-f.availMessage:
			s.avail[m.Topic] = m.Available
			f.Device.SetAvailable(countTrue(s.avail) == len(s.sm.AvailabilityTopics))
		}
	}
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
func countTrue(m map[string]bool) int {
	ret := 0
	for _, b := range m {
		if b {
			ret++
		}
	}
	return ret
}
