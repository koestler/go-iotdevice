package mqttForwarders

import (
	"context"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"golang.org/x/exp/maps"
	"log"
	"math"
	"time"
)

type StructRegister struct {
	Category     string         `json:"Cat" example:"Monitor"`
	Name         string         `json:"Name" example:"PanelPower"`
	Description  string         `json:"Desc" example:"Panel power"`
	Type         string         `json:"Type" example:"number"`
	Enum         map[int]string `json:"Enum,omitempty"`
	Unit         string         `json:"Unit,omitempty" example:"W"`
	Sort         int            `json:"Sort" example:"100"`
	Controllable bool           `json:"Cont" example:"false"`
}

type StructureMessage struct {
	AvailabilityTopics []string         `json:"Avail,omitempty"`
	RealtimeTopic      string           `json:"Realtime,omitempty"`
	Registers          []StructRegister `json:"Registers"`
}

func runStructureForwarder(
	ctx context.Context,
	dev device.Device,
	mc mqttClient.Client,
	registerFilter config.RegisterFilterConfig,
) {
	// start mqtt forwarder for realtime messages (send data as soon as it arrives) output
	mCfg := mc.Config().Structure()

	filter := createRegisterValueFilter(registerFilter)

	if mCfg.Interval() <= 0 {
		go structureOnUpdateModeRoutine(ctx, dev, mc, filter)
	} else {
		go structurePeriodicModeRoutine(ctx, dev, mc, filter)
	}
}

func structureOnUpdateModeRoutine(
	ctx context.Context,
	dev device.Device,
	mc mqttClient.Client,
	filter dataflow.RegisterFilterFunc,
) {
	devCfg := dev.Config()
	mcCfg := mc.Config()

	if devCfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->structure: start on-update mode",
			devCfg.Name(), mcCfg.Name(),
		)

		defer log.Printf(
			"device[%s]->mqttClient[%s]->structure: exit",
			devCfg.Name(), mcCfg.Name(),
		)
	}

	regSubscription := dev.RegisterDb().Subscribe(ctx, filter)
	structureTopic := mcCfg.StructureTopic(devCfg.Name())

	// when a new register arrives, wait until no new register is received for 100ms
	// and then send all updates together
	ticker := time.NewTicker(math.MaxInt64)
	defer ticker.Stop()

	registers := make(map[string]StructRegister)
	for {
		select {
		case <-ctx.Done():
			return
		case reg := <-regSubscription:
			if len(registers) < 1 {
				ticker.Reset(100 * time.Millisecond)
			}
			regName := reg.Name()
			if regName == device.AvailabilityRegisterName {
				// do not use Availability as a register in mqtt; availability is handled separately
				continue
			}

			registers[regName] = NewStructRegister(reg)
		case <-ticker.C:
			ticker.Stop()

			publishStruct(mc, devCfg, structureTopic, maps.Values(registers))
			clear(registers)
		}
	}
}

func structurePeriodicModeRoutine(
	ctx context.Context,
	dev device.Device,
	mc mqttClient.Client,
	filter dataflow.RegisterFilterFunc,
) {
	devCfg := dev.Config()
	mcCfg := mc.Config()
	structureInterval := mcCfg.Structure().Interval()

	if devCfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->structure: start periodic mode, send every %s",
			devCfg.Name(), mcCfg.Name(), structureInterval,
		)

		defer log.Printf(
			"device[%s]->mqttClient[%s]->structure: exit",
			devCfg.Name(), mcCfg.Name(),
		)
	}

	structureTopic := mcCfg.StructureTopic(devCfg.Name())

	ticker := time.NewTicker(structureInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			registers := dev.RegisterDb().GetFiltered(filter)
			publishStruct(mc, devCfg, structureTopic, NewStructRegisters(registers...))
		}
	}
}

func getAvailabilityTopics(devCfg device.Config, mcCfg mqttClient.Config) (ret []string) {
	ret = make([]string, 0, 2)
	if mcCfg.AvailabilityClient().Enabled() {
		ret = append(ret, mcCfg.AvailabilityClientTopic())
	}
	if mcCfg.AvailabilityDevice().Enabled() && existsByName(devCfg.Name(), mcCfg.AvailabilityDevice().Devices()) {
		ret = append(ret, mcCfg.AvailabilityDeviceTopic(devCfg.Name()))
	}

	return
}

func existsByName[N config.Nameable](needle string, haystack []N) bool {
	for _, t := range haystack {
		if needle == t.Name() {
			return true
		}
	}
	return false
}

func getRealtimeTopic(devCfg device.Config, mcCfg mqttClient.Config) string {
	if mcCfg.Realtime().Enabled() {
		return mcCfg.RealtimeTopic(devCfg.Name(), "%RegisterName%")
	}
	return ""
}

func publishStruct(mc mqttClient.Client, devCfg device.Config, topic string, registers []StructRegister) {
	mcCfg := mc.Config()
	mCfg := mcCfg.Structure()

	msg := StructureMessage{
		AvailabilityTopics: getAvailabilityTopics(devCfg, mcCfg),
		RealtimeTopic:      getRealtimeTopic(devCfg, mcCfg),
		Registers:          registers,
	}

	if devCfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->structure: send: %v",
			devCfg.Name(), mcCfg.Name(), msg,
		)
	}

	if payload, err := json.Marshal(msg); err != nil {
		log.Printf(
			"device[%s]->mqttClient[%s]->structure: cannot generate message: %s",
			devCfg.Name(), mcCfg.Name(), err,
		)
	} else {
		mc.Publish(
			topic,
			payload,
			mCfg.Qos(),
			mCfg.Retain(),
		)
	}
}

func NewStructRegisters(regs ...dataflow.RegisterStruct) (ret []StructRegister) {
	ret = make([]StructRegister, len(regs))
	for i, reg := range regs {
		ret[i] = NewStructRegister(reg)
	}
	return ret
}

func NewStructRegister(reg dataflow.Register) StructRegister {
	return StructRegister{
		Category:     reg.Category(),
		Name:         reg.Name(),
		Description:  reg.Description(),
		Type:         reg.RegisterType().String(),
		Enum:         reg.Enum(),
		Unit:         reg.Unit(),
		Sort:         reg.Sort(),
		Controllable: reg.Controllable(),
	}
}
