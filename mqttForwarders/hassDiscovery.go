package mqttForwarders

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"log"
	"strings"
	"time"
)

type homeassistantDiscoveryAvailabilityStruct struct {
	Topic string `json:"t"`
}

type homeassistantDiscoveryBaseMessage struct {
	UniqueId         string                                     `json:"uniq_id"`
	Name             string                                     `json:"name"`
	Availability     []homeassistantDiscoveryAvailabilityStruct `json:"avty"`
	AvailabilityMode string                                     `json:"avty_mode"`
}

type homeassistantDiscoverySensorMessage struct {
	homeassistantDiscoveryBaseMessage
	StateTopic        string `json:"stat_t"`
	ValueTemplate     string `json:"val_tpl"`
	UnitOfMeasurement string `json:"unit_of_meas,omitempty"`
}

type homeassistantDiscoverySwitchMessage struct {
	homeassistantDiscoveryBaseMessage
	CommandTopic  string `json:"cmd_t"`
	StateTopic    string `json:"stat_t"`
	ValueTemplate string `json:"val_tpl"`
	PayloadOff    string `json:"pl_off"`
	PayloadOn     string `json:"pl_on"`
}

func runHomeassistantDiscoveryForwarder(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	filterConf dataflow.RegisterFilterConf,
) {

	// check if realtime messages are activated
	realtimeCfg := getRealtimeCfg(cfg, dev.Name())

	if realtimeCfg == nil {
		log.Printf(
			"mqttClient[%s]->device[%s]->homeassistantDiscovery: "+
				"realtime messages are not enabled for this device; do not send any discovery messages",
			mc.Name(), dev.Name(),
		)

		return
	}

	hassFilter := createRegisterValueFilter(filterConf)
	realtimeFilter := createRegisterValueFilter(realtimeCfg.Filter())
	var filter dataflow.RegisterFilterFunc = func(r dataflow.Register) bool {
		return hassFilter(r) && realtimeFilter(r)
	}
	commandFilter := getCommandFilter(cfg, dev.Name())

	if cfg.HomeassistantDiscovery().Interval() <= 0 {
		go homeassistantDiscoveryOnUpdateModeRoutine(ctx, cfg, dev, mc, filter, commandFilter)
	} else {
		go homeassistantDiscoveryPeriodicModeRoutine(ctx, cfg, dev, mc, filter, commandFilter)
	}
}

func getRealtimeCfg(cfg Config, deviceName string) MqttDeviceSectionConfig {
	if cfg.Realtime().Enabled() {
		for _, d := range cfg.Realtime().Devices() {
			if d.Name() == deviceName {
				return d
			}
		}
	}
	return nil
}

func homeassistantDiscoveryOnUpdateModeRoutine(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	filter dataflow.RegisterFilterFunc,
	commandFilter dataflow.RegisterFilterFunc,
) {
	if cfg.LogDebug() {
		log.Printf(
			"mqttClient[%s]->device[%s]->homeassistantDiscovery: start on-update mode",
			mc.Name(), dev.Name(),
		)
	}

	regSubscription := dev.RegisterDb().Subscribe(ctx, filter)

	for {
		select {
		case <-ctx.Done():
			return
		case reg := <-regSubscription:
			publishHomeassistantDiscoveryMessage(cfg, mc, dev.Name(), reg, commandFilter)
		}
	}
}

func homeassistantDiscoveryPeriodicModeRoutine(
	ctx context.Context,
	cfg Config,
	dev device.Device,
	mc mqttClient.Client,
	filter dataflow.RegisterFilterFunc,
	commandFilter dataflow.RegisterFilterFunc,
) {
	interval := cfg.HomeassistantDiscovery().Interval()

	if cfg.LogDebug() {
		log.Printf(
			"mqttClient[%s]->device[%s]->homeassistantDiscovery: start periodic mode, send every %s",
			mc.Name(), dev.Name(), interval,
		)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, reg := range dev.RegisterDb().GetFiltered(filter) {
				publishHomeassistantDiscoveryMessage(cfg, mc, dev.Name(), reg, commandFilter)
			}
		}
	}
}

func publishHomeassistantDiscoveryMessage(
	cfg Config,
	mc mqttClient.Client,
	deviceName string,
	register dataflow.Register,
	commandFilter dataflow.RegisterFilterFunc,
) {
	mCfg := cfg.HomeassistantDiscovery()

	var topic string
	var msg interface{}

	switch register.RegisterType() {
	case dataflow.NumberRegister:
		topic, msg = getHomeassistantDiscoverySensorMessage(
			cfg,
			deviceName,
			register,
			"{{ value_json.NumVal }}",
		)
	case dataflow.TextRegister:
		topic, msg = getHomeassistantDiscoverySensorMessage(
			cfg,
			deviceName,
			register,
			"{{ value_json.TextVal }}",
		)
	case dataflow.EnumRegister:
		// generate Jinja2 template to translate enumIdx to string
		enum := register.Enum()
		var valueTemplate strings.Builder
		op := "if"
		for idx, value := range enum {
			fmt.Fprintf(&valueTemplate, "{%% %s value_json.EnumIdx == %d %%}%s", op, idx, value)
			op = "elif"
		}
		valueTemplate.WriteString("{% endif %}")

		if commandFilter(register) {
			offIdx := 0
			onIdx := 1

			topic, msg = getHomeassistantDiscoverySwitchMessage(
				cfg,
				deviceName,
				register,
				valueTemplate.String(),
				getCommandPayload(CommandMessage{EnumIdx: &offIdx}, mc.Name(), deviceName),
				getCommandPayload(CommandMessage{EnumIdx: &onIdx}, mc.Name(), deviceName),
			)
		} else {
			topic, msg = getHomeassistantDiscoverySensorMessage(
				cfg,
				deviceName,
				register,
				valueTemplate.String(),
			)
		}
	default:
		return
	}

	if payload, err := json.Marshal(msg); err != nil {
		log.Printf("mqttClient[%s]->device[%s]->homeassistantDiscovery: cannot generate discovery message: %s",
			mc.Name(), deviceName, err,
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

func getHomeassistantDiscoverySensorMessage(
	cfg Config,
	deviceName string,
	register dataflow.Register,
	valueTemplate string,
) (topic string, msg homeassistantDiscoverySensorMessage) {
	uniqueId, base := getHomeassistantDiscoveryBaseMessage(cfg, deviceName, register)

	topic = cfg.HomeassistantDiscoveryTopic("sensor", cfg.ClientId(), uniqueId)

	msg = homeassistantDiscoverySensorMessage{
		homeassistantDiscoveryBaseMessage: base,
		StateTopic:                        cfg.RealtimeTopic(deviceName, register.Name()),
		ValueTemplate:                     valueTemplate,
		UnitOfMeasurement:                 register.Unit(),
	}

	return
}

func getHomeassistantDiscoverySwitchMessage(
	cfg Config,
	deviceName string,
	register dataflow.Register,
	valueTemplate,
	payloadOff,
	payloadOn string,
) (topic string, msg homeassistantDiscoverySwitchMessage) {
	uniqueId, base := getHomeassistantDiscoveryBaseMessage(cfg, deviceName, register)

	topic = cfg.HomeassistantDiscoveryTopic("switch", cfg.ClientId(), uniqueId)

	msg = homeassistantDiscoverySwitchMessage{
		homeassistantDiscoveryBaseMessage: base,
		CommandTopic:                      cfg.CommandTopic(deviceName, register.Name()),
		StateTopic:                        cfg.RealtimeTopic(deviceName, register.Name()),
		ValueTemplate:                     valueTemplate,
		PayloadOff:                        payloadOff,
		PayloadOn:                         payloadOn,
	}

	return
}

func getHomeassistantDiscoveryBaseMessage(
	cfg Config,
	deviceName string,
	register dataflow.Register,
) (uniqueId string, base homeassistantDiscoveryBaseMessage) {
	uniqueId = fmt.Sprintf("%s-%s", deviceName, CamelToSnakeCase(register.Name()))
	name := fmt.Sprintf("%s %s", deviceName, register.Description())

	base = homeassistantDiscoveryBaseMessage{
		UniqueId:         uniqueId,
		Name:             name,
		Availability:     getHomeassistantDiscoveryAvailabilityTopics(cfg, deviceName),
		AvailabilityMode: "all",
	}

	return
}

func getHomeassistantDiscoveryAvailabilityTopics(cfg Config, deviceName string) (ret []homeassistantDiscoveryAvailabilityStruct) {
	if cfg.AvailabilityClient().Enabled() {
		ret = append(ret, homeassistantDiscoveryAvailabilityStruct{cfg.AvailabilityClientTopic()})
	}
	if cfg.AvailabilityDevice().Enabled() {
		ret = append(ret, homeassistantDiscoveryAvailabilityStruct{cfg.AvailabilityDeviceTopic(deviceName)})
	}

	return
}

func getCommandPayload(msg CommandMessage, mcName, deviceName string) (payload string) {
	if payload, err := json.Marshal(msg); err != nil {
		log.Printf(
			"mqttClient[%s]->device[%s]->homeassistantDiscovery: cannot generate command message: %s",
			mcName, deviceName, err,
		)
		return "{}"
	} else {
		return string(payload)
	}
}
