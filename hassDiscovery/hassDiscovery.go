package hassDiscovery

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/pool"
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

		for {
			select {
			case <-hd.ctx.Done():
				return
			}
		}
	}()
}

func (hd *HassDiscovery) Shutdown() {
	hd.cancel()
	hd.wg.Wait()
}
