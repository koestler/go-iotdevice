package mqttForwarders

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/config"
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

type homeassistantDiscoveryMessage struct {
	UniqueId string `json:"uniq_id"`
	Name     string `json:"name"`

	StateTopic        string                                     `json:"stat_t"`
	Availability      []homeassistantDiscoveryAvailabilityStruct `json:"avty"`
	AvailabilityMode  string                                     `json:"avty_mode"`
	ValueTemplate     string                                     `json:"val_tpl"`
	UnitOfMeasurement string                                     `json:"unit_of_meas,omitempty"`
}

func runHomeassistantDiscoveryForwarder(
	ctx context.Context,
	dev device.Device,
	mc mqttClient.Client,
	registerFilter config.RegisterFilterConfig,
) {
	filter := createRegisterValueFilter(registerFilter)

	if mc.Config().HomeassistantDiscovery().Interval() <= 0 {
		go homeassistantDiscoveryOnUpdateModeRoutine(ctx, dev, mc, filter)
	} else {
		go homeassistantDiscoveryPeriodicModeRoutine(ctx, dev, mc, filter)
	}
}

func homeassistantDiscoveryOnUpdateModeRoutine(
	ctx context.Context,
	dev device.Device,
	mc mqttClient.Client,
	filter dataflow.RegisterFilterFunc,
) {
	devCfg := dev.Config()
	if devCfg.LogDebug() {
		mcName := mc.Config().Name()

		log.Printf(
			"device[%s]->mqttClient[%s]->homeassistantDiscovery: start on-update mode",
			devCfg.Name(), mcName,
		)

		defer log.Printf(
			"device[%s]->mqttClient[%s]->homeassistantDiscovery: exit",
			devCfg.Name(), mcName,
		)
	}

	regSubscription := dev.RegisterDb().Subscribe(ctx, filter)

	for {
		select {
		case <-ctx.Done():
			return
		case reg := <-regSubscription:
			publishHomeassistantDiscoveryMessage(mc, dev.Name(), reg)
		}
	}
}

func homeassistantDiscoveryPeriodicModeRoutine(
	ctx context.Context,
	dev device.Device,
	mc mqttClient.Client,
	filter dataflow.RegisterFilterFunc,
) {
	devCfg := dev.Config()
	mcCfg := mc.Config()
	interval := mcCfg.HomeassistantDiscovery().Interval()

	if devCfg.LogDebug() {
		log.Printf(
			"device[%s]->mqttClient[%s]->homeassistantDiscovery: start periodic mode, send every %s",
			devCfg.Name(), mcCfg.Name(), interval,
		)

		defer log.Printf(
			"device[%s]->mqttClient[%s]->homeassistantDiscovery: exit",
			devCfg.Name(), mcCfg.Name(),
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
				publishHomeassistantDiscoveryMessage(mc, dev.Name(), reg)
			}
		}
	}
}

func publishHomeassistantDiscoveryMessage(
	mc mqttClient.Client,
	deviceName string,
	register dataflow.Register,
) {
	mCfg := mc.Config().HomeassistantDiscovery()

	var topic string
	var msg homeassistantDiscoveryMessage

	switch register.RegisterType() {
	case dataflow.NumberRegister:
		topic, msg = getHomeassistantDiscoverySensorMessage(
			mc.Config(),
			deviceName,
			register,
			"{{ value_json.NumVal }}",
		)
	case dataflow.TextRegister:
		topic, msg = getHomeassistantDiscoverySensorMessage(
			mc.Config(),
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

		topic, msg = getHomeassistantDiscoverySensorMessage(
			mc.Config(),
			deviceName,
			register,
			valueTemplate.String(),
		)
	default:
		return
	}

	log.Printf("homeassistantDiscovery[%s]: send %s %#v", mc.Name(), topic, msg)

	if payload, err := json.Marshal(msg); err != nil {
		log.Printf("homeassistantDiscovery: cannot generate discovery message: %s", err)
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
	mcCfg mqttClient.Config,
	deviceName string,
	register dataflow.Register,
	valueTemplate string,
) (topic string, msg homeassistantDiscoveryMessage) {
	uniqueId := fmt.Sprintf("%s-%s", deviceName, CamelToSnakeCase(register.Name()))
	name := fmt.Sprintf("%s %s", deviceName, register.Description())

	topic = mcCfg.HomeassistantDiscoveryTopic("sensor", mcCfg.ClientId(), uniqueId)

	msg = homeassistantDiscoveryMessage{
		UniqueId:          uniqueId,
		Name:              name,
		StateTopic:        mcCfg.RealtimeTopic(deviceName, register.Name()),
		Availability:      getHomeassistantDiscoveryAvailabilityTopics(deviceName, mcCfg),
		AvailabilityMode:  "all",
		ValueTemplate:     valueTemplate,
		UnitOfMeasurement: register.Unit(),
	}

	return
}

func getHomeassistantDiscoveryAvailabilityTopics(deviceName string, mcCfg mqttClient.Config) (ret []homeassistantDiscoveryAvailabilityStruct) {
	if mcCfg.AvailabilityClient().Enabled() {
		ret = append(ret, homeassistantDiscoveryAvailabilityStruct{mcCfg.AvailabilityClientTopic()})
	}
	if mcCfg.AvailabilityDevice().Enabled() {
		ret = append(ret, homeassistantDiscoveryAvailabilityStruct{mcCfg.AvailabilityDeviceTopic(deviceName)})
	}

	return
}
