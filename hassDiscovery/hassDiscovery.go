package hassDiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"github.com/koestler/go-iotdevice/restarter"
	"log"
	"regexp"
	"strings"
	"sync"
)

type ConfigItem interface {
	TopicPrefix() string
	ViaMqttClients() []string
	Devices() []string
	CategoriesMatcher() []*regexp.Regexp
	RegistersMatcher() []*regexp.Regexp
}

type HassDiscovery struct {
	configItems    []ConfigItem
	devicePool     *pool.Pool[*restarter.Restarter[device.Device]]
	mqttClientPool *pool.Pool[mqttClient.Client]

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func Create[CI ConfigItem](
	configItems []CI,
	devicePool *pool.Pool[*restarter.Restarter[device.Device]],
	mqttClientPool *pool.Pool[mqttClient.Client],
) *HassDiscovery {
	ctx, cancel := context.WithCancel(context.Background())

	configItemsInterface := make([]ConfigItem, len(configItems))
	for i, ci := range configItems {
		configItemsInterface[i] = ci
	}

	hd := HassDiscovery{
		configItems:    configItemsInterface,
		devicePool:     devicePool,
		mqttClientPool: mqttClientPool,
		ctx:            ctx,
		cancel:         cancel,
	}

	return &hd
}

func (hd *HassDiscovery) Run() {
	for deviceName, configItems := range hd.getDevices() {
		hd.wg.Add(1)
		go func(deviceName string, configItems []ConfigItem) {
			defer hd.wg.Done()
			hd.handleRegisters(deviceName, configItems)
		}(deviceName, configItems)
	}
}

func (hd *HassDiscovery) handleRegisters(deviceName string, configItems []ConfigItem) {
	devRestarter := hd.devicePool.GetByName(deviceName)
	if devRestarter == nil {
		log.Printf("hassDiscovery: device '%s' not found", deviceName)
		return
	}
	dev := devRestarter.Service()

	registerSubscription := dev.RegisterDb().Subscribe(hd.ctx)

	type key struct {
		mqttClientName string
		topicPrefix    string
		registerName   string
	}
	alreadySent := make(map[key]struct{})

	// the subscription closes the chan when the hd.ctx expires
	for reg := range registerSubscription {
		for _, ci := range configItems {
			// check if config item matches this register
			if !regMatchesConfigItem(reg, ci) {
				continue
			}

			for _, mqttClientName := range ci.ViaMqttClients() {
				// only send discovery messages on mqtt clients were we also send realtime messages
				if !stringContains(mqttClientName, dev.Config().ViaMqttClients()) {
					continue
				}

				// only publish once per device, mqttClient, Topic, and register Name
				k := key{
					mqttClientName: mqttClientName,
					topicPrefix:    ci.TopicPrefix(),
					registerName:   reg.Name(),
				}

				if _, exists := alreadySent[k]; exists {
					continue
				}
				alreadySent[k] = struct{}{}

				mc := hd.mqttClientPool.GetByName(k.mqttClientName)
				if mc == nil {
					log.Printf("hassDiscovery: mqttClient '%s' not found", k.mqttClientName)
					continue
				}

				hd.publishDiscoveryMessage(
					k.topicPrefix,
					mc,
					deviceName,
					reg,
				)
			}
		}
	}
}

func (hd *HassDiscovery) Shutdown() {
	hd.cancel()
	hd.wg.Wait()
}

// getDevices create a map of deviceName -> list of config items for this device
func (hd *HassDiscovery) getDevices() (ret map[string][]ConfigItem) {
	ret = make(map[string][]ConfigItem)

	for _, ci := range hd.configItems {
		for _, deviceName := range ci.Devices() {
			if _, ok := ret[deviceName]; !ok {
				ret[deviceName] = []ConfigItem{ci}
			} else {
				ret[deviceName] = append(ret[deviceName], ci)
			}
		}
	}
	return
}

func regMatchesConfigItem(reg dataflow.Register, ci ConfigItem) bool {
	return matchesRegexList(reg.Category(), ci.CategoriesMatcher()) &&
		matchesRegexList(reg.Name(), ci.RegistersMatcher())
}

func matchesRegexList(needle string, haystack []*regexp.Regexp) bool {
	if len(haystack) < 1 {
		// default (= empty list) is match anything
		return true
	}

	for _, t := range haystack {
		if t.MatchString(needle) {
			return true
		}
	}
	return false
}

func stringContains(needle string, haystack []string) bool {
	for _, t := range haystack {
		if t == needle {
			return true
		}
	}
	return false
}

func (hd *HassDiscovery) publishDiscoveryMessage(
	discoveryPrefix string,
	mc mqttClient.Client,
	deviceName string,
	register dataflow.Register,
) {
	var topic string
	var msg discoveryMessage

	switch register.RegisterType() {
	case dataflow.NumberRegister:
		topic, msg = getSensorMessage(
			discoveryPrefix,
			mc.Config(),
			deviceName,
			register,
			"{{ value_json.NumVal }}",
		)
	case dataflow.TextRegister:
		topic, msg = getSensorMessage(
			discoveryPrefix,
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

		topic, msg = getSensorMessage(
			discoveryPrefix,
			mc.Config(),
			deviceName,
			register,
			valueTemplate.String(),
		)
	default:
		return
	}

	log.Printf("hassDiscovery[%s]: send %s %#v", mc.Name(), topic, msg)

	if payload, err := json.Marshal(msg); err != nil {
		log.Printf("hassDiscovery: cannot generate discovery message: %s", err)
	} else {
		mc.Publish(
			topic,
			payload,
			mc.Config().Qos(),
			mc.Config().RealtimeRetain(),
		)
	}
}
