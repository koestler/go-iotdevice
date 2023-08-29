package hassDiscovery

import (
	"context"
	"encoding/json"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"github.com/koestler/go-iotdevice/restarter"
	"log"
	"regexp"
	"sync"
)

type ConfigItem interface {
	TopicPrefix() string
	ViaMqttClients() []string
	Devices() []string
	CategoriesMatcher() []*regexp.Regexp
	RegistersMatcher() []*regexp.Regexp
}

// only publish once per mqttClient, topic, device and register name even if multiple configItems match
type key struct {
	mqttClientName string
	topicPrefix    string
	deviceName     string
	registerName   string
}

type HassDiscovery struct {
	mqttClientPool *pool.Pool[mqttClient.Client]

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	discoverables map[key]dataflow.Register
}

func Create[CI ConfigItem](
	configItems []CI,
	devicePool *pool.Pool[*restarter.Restarter[device.Device]],
	mqttClientPool *pool.Pool[mqttClient.Client],
) *HassDiscovery {
	ctx, cancel := context.WithCancel(context.Background())

	hd := HassDiscovery{
		mqttClientPool: mqttClientPool,
		ctx:            ctx,
		cancel:         cancel,
		discoverables:  getDiscoverables(configItems, devicePool),
	}

	return &hd
}

func getDiscoverables[CI ConfigItem](
	configItems []CI,
	devicePool *pool.Pool[*restarter.Restarter[device.Device]],
) (discoverables map[key]dataflow.Register) {
	discoverables = make(map[key]dataflow.Register)

	// create a map of all discovery messages to send; only send one message per mqttClient, Topic, Device, Register
	for _, ci := range configItems {
		devices := devicePool.GetByNames(ci.Devices())
		for _, devRestarter := range devices {
			dev := devRestarter.Service()
			for _, reg := range dev.Registers() {
				if !regMatchesConfigItem(reg, ci) {
					continue
				}
				for _, mqttClientName := range ci.ViaMqttClients() {
					if !stringContains(mqttClientName, dev.Config().RealtimeViaMqttClients()) {
						// only send discovery messages on mqtt clients were we also send realtime messages
						continue
					}
					k := key{
						mqttClientName: mqttClientName,
						topicPrefix:    ci.TopicPrefix(),
						deviceName:     dev.Name(),
					}
					discoverables[k] = reg
				}
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

func (hd *HassDiscovery) Run() {
	hd.wg.Add(1)

	log.Printf("HassDiscovery: discoverables: %+v", hd.discoverables)

	go func() {
		defer hd.wg.Done()

		/*
			subscription := hd.stateStorage.Subscribe(hd.ctx, dataflow.EmptyFilter)

			configItem := hd.configItems[0]

			log.Printf("hassDiscovery: run main routine")

			for {
				select {
				case <-hd.ctx.Done():
					return
				case value := <-subscription.Drain():
					log.Printf("hassDiscovery: value received: %v", value)

					hd.handleRegister(
						configItem.TopicPrefix(),
						hd.mqttClientPool.GetByName(configItem.ViaMqttClients()[0]),
						value.DeviceName(),
						value.Register(),
					)
				}
			}
		*/
	}()
}

func (hd *HassDiscovery) Shutdown() {
	hd.cancel()
	hd.wg.Wait()
}

func (hd *HassDiscovery) handleRegister(
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
		)
	default:
		return
	}

	log.Printf("hassDiscovery: send %s %#v", topic, msg)

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
