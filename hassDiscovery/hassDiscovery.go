package hassDiscovery

import (
	"context"
	"encoding/json"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
	"log"
	"regexp"
	"sync"
)

type ConfigItem interface {
	TopicPrefix() string
	ViaMqttClients() []string
	DevicesMatcher() []*regexp.Regexp
	CategoriesMatcher() []*regexp.Regexp
	RegistersMatcher() []*regexp.Regexp
}

type HassDiscovery struct {
	configItems []ConfigItem

	stateStorage   *dataflow.ValueStorageInstance
	mqttClientPool *pool.Pool[mqttClient.Client]

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func Create[CI ConfigItem](
	configItems []CI,
	stateStorage *dataflow.ValueStorageInstance,
	mqttClientPool *pool.Pool[mqttClient.Client],
) *HassDiscovery {
	ctx, cancel := context.WithCancel(context.Background())

	hd := HassDiscovery{
		stateStorage:   stateStorage,
		mqttClientPool: mqttClientPool,
		ctx:            ctx,
		cancel:         cancel,
	}

	// cast struct slice to interface slice
	hd.configItems = make([]ConfigItem, len(configItems))
	for i, c := range configItems {
		hd.configItems[i] = c
	}

	return &hd
}

func (hd *HassDiscovery) Run() {
	hd.wg.Add(1)

	go func() {
		defer hd.wg.Done()

		filter := dataflow.Filter{OnlyOnce: true}
		subscription := hd.stateStorage.Subscribe(filter)
		defer subscription.Shutdown()

		log.Printf("hassDiscovery: filter: %+v", filter)

		configItem := hd.configItems[0]

		log.Printf("hassDiscovery: run main routine")

		for {
			select {
			case <-hd.ctx.Done():
				return
			case value := <-subscription.GetOutput():
				log.Printf("hassDiscovery: value received: %v", value)

				hd.handleRegister(
					configItem.TopicPrefix(),
					hd.mqttClientPool.GetByName(configItem.ViaMqttClients()[0]),
					value.DeviceName(),
					value.Register(),
				)
			}
		}
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
	} else if err := mc.Publish(
		topic,
		payload,
		mc.Config().Qos(),
		mc.Config().RealtimeRetain(),
	); err != nil {
		log.Printf("hassDiscovery: cannot publish realtime: %s", err)
	}

}
