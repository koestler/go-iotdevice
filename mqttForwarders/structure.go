package mqttForwarders

import (
	"context"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"github.com/koestler/go-iotdevice/v3/mqttClient"
	"golang.org/x/exp/maps"
	"log"
	"math"
	"time"
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
	// start mqtt forwarder for realtime messages (send data as soon as it arrives) output
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

	// when a new register arrives, wait until no new register is received for 100ms
	// and then send all updates together
	ticker := time.NewTicker(math.MaxInt64)
	defer ticker.Stop()

	registers := make(map[string]dataflow.RegisterStruct)
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

			registers[regName] = reg
		case <-ticker.C:
			ticker.Stop()

			publishStruct(cfg, mc, dev.Name(), structureTopic, maps.Values(registers))
			clear(registers)
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
	if cfg.AvailabilityDevice().Enabled() && existsByName(devName, cfg.AvailabilityDevice().Devices()) {
		ret = append(ret, cfg.AvailabilityDeviceTopic(devName))
	}

	return
}

type Nameable interface {
	Name() string
}

func existsByName[N Nameable](needle string, haystack []N) bool {
	for _, t := range haystack {
		if needle == t.Name() {
			return true
		}
	}
	return false
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
		sr := NewStructRegister(r)
		sr.Writable = filter(r)
		if sr.Writable {
			countWritable++
		}
		oup[i] = sr
	}

	return
}

func NewStructRegister(reg dataflow.Register) StructRegister {
	return StructRegister{
		Category:    reg.Category(),
		Name:        reg.Name(),
		Description: reg.Description(),
		Type:        reg.RegisterType().String(),
		Enum:        reg.Enum(),
		Unit:        reg.Unit(),
		Sort:        reg.Sort(),
		Writable:    reg.Writable(),
	}
}
