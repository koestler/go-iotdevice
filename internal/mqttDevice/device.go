package mqttDevice

import (
	"cmp"
	"context"
	"encoding/json"
	"log"
	"slices"

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
		log.Printf("mqttDevice[%s]: unsupported kind: %s", c.Name(), mCfg.Kind().String())
		return
	}

	for mqttClientName, topics := range mCfg.MqttClientTopics() {
		mc, ok := c.mqttClientPool.GetByName(mqttClientName)
		if !ok {
			continue
		}

		runForwarder(ctx, c, mc, topics)
	}

	<-ctx.Done()
	return nil, false
}

type forwarder struct {
	Device *DeviceStruct

	Client       mqttClient.Client
	StructTopics []string

	structMessage chan mqttForwarders.StructureMessage
	availMessage  chan availabilityMessage
}

func runForwarder(
	ctx context.Context,
	c *DeviceStruct,
	client mqttClient.Client,
	structTopics []string,
) {
	f := &forwarder{
		Device: c,

		Client:       client,
		StructTopics: structTopics,

		structMessage: make(chan mqttForwarders.StructureMessage),
		availMessage:  make(chan availabilityMessage),
	}

	logDebug := c.Config().LogDebug()

	// subscribe to structure updates
	removeStructs := make([]func(), 0, len(f.StructTopics))
	for _, structTopic := range f.StructTopics {
		if logDebug {
			log.Printf("mqttDevice[%s]->mqttClient[%s]: subscribe to topic=%s", f.Device.Name(), f.Client.Name(), structTopic)
		}

		// subscribe to the struct topics
		sc := f.structMessage
		rm := f.Client.AddRoute(ctx, structTopic, func(m mqttClient.Message) {
			s, err := parseStructPayload(m.Payload())
			if err != nil {
				log.Printf("mqttDevice[%s]->mqttClient[%s]: cannot parse structure payload: %s", f.Device.Name(), f.Client.Name(), err)
				return
			}

			if logDebug {
				log.Printf("mqttDevice[%s]->mqttClient[%s]: structure message parsed: %v", f.Device.Name(), f.Client.Name(), s)
			}

			select {
			case <-ctx.Done():
			case sc <- s:
			}
		})
		removeStructs = append(removeStructs, rm)
	}

	go f.run(ctx, removeStructs)
}

func parseStructPayload(payload []byte) (msg mqttForwarders.StructureMessage, err error) {
	err = json.Unmarshal(payload, &msg)

	// we ignore the order received and sort the slices to have a deterministic order for comparison
	slices.Sort(msg.AvailabilityTopics)
	slices.SortFunc(msg.Registers, func(a, b mqttForwarders.StructRegister) int {
		return cmp.Or(cmp.Compare(a.Sort, b.Sort), cmp.Compare(a.Name, b.Name))
	})

	return
}

type forwarderState struct {
	structure mqttForwarders.StructureMessage
	avail     map[string]bool // topic -> availability
	remove    func()
}

func (f *forwarder) run(ctx context.Context, removeStructs []func()) {
	s := forwarderState{
		avail:  make(map[string]bool),
		remove: func() {},
	}

	for {
		select {
		case <-ctx.Done():
			for _, rm := range removeStructs {
				rm()
			}
			s.remove()
			return
		case m := <-f.structMessage:
			if s.structure.Equal(m) {
				if f.Device.Config().LogDebug() {
					log.Printf("mqttDevice[%s]->mqttClient[%s]: new structure ignored (no change): %v", f.Device.Name(), f.Client.Name(), m)
				}
				continue
			}

			if f.Device.Config().LogDebug() {
				log.Printf("mqttDevice[%s]->mqttClient[%s]: new structure received: %v", f.Device.Name(), f.Client.Name(), m)
			}
			s.structure = m

			// setup new subscriptions based on the new structure before removing the old routes
			// to avoid unsubscribing / subscribing in case the topic did not change
			removeAvail := f.setupAvailabilitySubscription(ctx, m.AvailabilityTopics)
			removeTelemetry := f.setupTelemetrySubscription(ctx, m.TelemetryTopic)
			removeRealtime := f.setupRealtimeSubscription(ctx, m.RealtimeTopic)

			// remove command subscription before setting up the new command forwarder to avoid sending the commands twice
			s.remove()

			removeCommand := f.setupCommandForwarder(ctx, m.CommandTopic, m.Registers)
			f.Device.updateRegisters(m.Registers)

			s.remove = func() {
				removeAvail()
				removeTelemetry()
				removeRealtime()
				removeCommand()
			}

			// update availability based on the new structure
			for k := range s.avail {
				if !slices.Contains(m.AvailabilityTopics, k) {
					delete(s.avail, k)
				}
			}
			f.setAvailable(s)

		case m := <-f.availMessage:
			s.avail[m.Topic] = m.Available
			f.setAvailable(s)
		}
	}
}

func (f *forwarder) setAvailable(s forwarderState) {
	f.Device.SetAvailable(countTrue(s.avail) == len(s.structure.AvailabilityTopics))
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
