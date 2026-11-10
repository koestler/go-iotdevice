package mqttForwarders

import (
	"context"
	"encoding/json"
	"log"
	"maps"
	"math"
	"slices"
	"time"

	"github.com/koestler/go-iotdevice/v3/internal/dataflow"
	"github.com/koestler/go-iotdevice/v3/internal/device"
	"github.com/koestler/go-iotdevice/v3/internal/mqttClient"
)

type StructRegister struct {
	Category    string         `json:"Cat" example:"Monitor"`
	Name        string         `json:"Name" example:"PanelPower"`
	Description string         `json:"Desc" example:"Panel power"`
	Type        string         `json:"Type" example:"number"`
	Enum        map[int]string `json:"Enum,omitempty"`
	Unit        string         `json:"Unit,omitempty" example:"W"`
	Sort        int            `json:"Sort" example:"100"`
	Writable    bool           `json:"Cmnd" example:"false"`
}

type StructureMessage struct {
	AvailabilityTopics []string         `json:"Avail,omitempty"`
	TelemetryTopic     string           `json:"Tele,omitempty"`
	RealtimeTopic      string           `json:"Real,omitempty"`
	CommandTopic       string           `json:"Cmnd,omitempty"`
	Registers          []StructRegister `json:"Regs"`
}

func runStructureForwarder(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	filterConf dataflow.RegisterFilterConf,
) {
	// start mqtt forwarder for the register structure messages
	mCfg := cfg.Structure()

	filter := createRegisterValueFilter(filterConf)

	if mCfg.Interval() <= 0 {
		go structureOnUpdateModeRoutine(ctx, cfg, dev, mc, filter)
	} else {
		go structurePeriodicModeRoutine(ctx, cfg, dev, mc, filter)
	}
}

func structureOnUpdateModeRoutine(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	filter dataflow.RegisterFilterFunc,
) {
	devCfg := dev.Config()

	if devCfg.LogDebug() {
		log.Printf(
			"mqttClient[%s]->device[%s]->structure: start on-update mode",
			mc.Name(), devCfg.Name(),
		)
	}

	regSubscription := dev.RegisterDb().Subscribe(ctx, filter)
	structureTopic := cfg.StructureTopic(devCfg.Name())

	ticker := time.NewTicker(math.MaxInt64)
	defer ticker.Stop()

	pending := false
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-regSubscription:
			if !ok {
				return
			}
			// collect updates for a short time and then always send the full list of registers
			if !pending {
				pending = true
				ticker.Reset(100 * time.Millisecond)
			}
		case <-ticker.C:
			ticker.Stop()
			pending = false

			publishStruct(cfg, mc, dev.Name(), structureTopic, dev.RegisterDb().GetFiltered(filter))
		}
	}
}

func structurePeriodicModeRoutine(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	filter dataflow.RegisterFilterFunc,
) {
	structureInterval := cfg.Structure().Interval()

	if cfg.LogDebug() {
		log.Printf(
			"mqttClient[%s]->device[%s]->structure: start periodic mode, send every %s",
			mc.Name(), dev.Name(), structureInterval,
		)
	}

	structureTopic := cfg.StructureTopic(dev.Name())

	ticker := time.NewTicker(structureInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			registers := dev.RegisterDb().GetFiltered(filter)
			publishStruct(cfg, mc, dev.Name(), structureTopic, registers)
		}
	}
}

func getAvailabilityTopics(cfg Config, devName string) (ret []string) {
	ret = make([]string, 0, 2)
	if cfg.AvailabilityClient().Enabled() {
		ret = append(ret, cfg.AvailabilityClientTopic())
	}
	if cfg.AvailabilityDevice().Enabled() && findSectionDevice(cfg.AvailabilityDevice(), devName) != nil {
		ret = append(ret, cfg.AvailabilityDeviceTopic(devName))
	}

	return
}

func publishStruct(cfg Config, mc mqttClient.Client, devName string, topic string, registers []dataflow.RegisterStruct) {
	mCfg := cfg.Structure()

	dataflow.SortRegisterStructs(registers)

	structRegisters, countWritable := applyWritableFiltersToRegisters(cfg, devName, registers)

	msg := StructureMessage{
		AvailabilityTopics: getAvailabilityTopics(cfg, devName),
		TelemetryTopic: func() string {
			if cfg.Telemetry().Enabled() {
				return cfg.TelemetryTopic(devName)
			}
			return ""
		}(),
		RealtimeTopic: func() string {
			if cfg.Realtime().Enabled() {
				return cfg.RealtimeTopic(devName, "%RegisterName%")
			}
			return ""
		}(),
		CommandTopic: func() string {
			if countWritable < 1 {
				return ""
			}
			return cfg.CommandTopic(devName, "%RegisterName%")
		}(),
		Registers: structRegisters,
	}

	if cfg.LogDebug() {
		log.Printf(
			"mqttClient[%s]->device[%s]->structure: send: %v",
			mc.Name(), devName, msg,
		)
	}

	if payload, err := json.Marshal(msg); err != nil {
		log.Printf(
			"mqttClient[%s]->device[%s]->structure: cannot generate message: %s",
			mc.Name(), devName, err,
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

func applyWritableFiltersToRegisters(cfg Config, devName string, inp []dataflow.RegisterStruct) (oup []StructRegister, countWritable int) {
	oup = make([]StructRegister, len(inp))

	filter := getCommandFilter(cfg, devName)

	for i, r := range inp {
		sr := newStructRegister(r)
		sr.Writable = filter(r)
		if sr.Writable {
			countWritable++
		}
		oup[i] = sr
	}

	return
}

func newStructRegister(reg dataflow.Register) StructRegister {
	return StructRegister{
		Category:    reg.Category(),
		Name:        reg.Name(),
		Description: reg.Description(),
		Type:        reg.RegisterType().String(),
		Enum:        reg.Enum(),
		Unit:        reg.Unit(),
		Sort:        reg.Sort(),
	}
}

func (s StructRegister) Equal(other StructRegister) bool {
	return s.Category == other.Category &&
		s.Name == other.Name &&
		s.Description == other.Description &&
		s.Type == other.Type &&
		s.Unit == other.Unit &&
		s.Sort == other.Sort &&
		s.Writable == other.Writable &&
		maps.Equal(s.Enum, other.Enum)
}

func (sm StructureMessage) Equal(other StructureMessage) bool {
	return sm.TelemetryTopic == other.TelemetryTopic &&
		sm.RealtimeTopic == other.RealtimeTopic &&
		sm.CommandTopic == other.CommandTopic &&
		slices.Equal(sm.AvailabilityTopics, other.AvailabilityTopics) &&
		slices.EqualFunc(sm.Registers, other.Registers, func(a, b StructRegister) bool {
			return a.Equal(b)
		})
}
