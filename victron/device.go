package victron

import (
	"fmt"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/mqttClient"
	"github.com/koestler/go-iotdevice/vedirect"
	"log"
	"sync"
	"time"
)

type Config interface {
	Name() string
	Kind() config.VictronDeviceKind
	Device() string
	SkipFields() []string
	SkipCategories() []string
	TelemetryViaMqttClients() []string
	RealtimeViaMqttClients() []string
	LogDebug() bool
	LogComDebug() bool
}

type DeviceStruct struct {
	cfg Config

	source *dataflow.Source

	deviceId         vedirect.VeProduct
	registers        dataflow.Registers
	lastUpdated      time.Time
	lastUpdatedMutex sync.RWMutex
	model            string

	shutdown chan struct{}
	closed   chan struct{}
}

func RunDevice(cfg Config,
	mqttClientPool *mqttClient.ClientPool,
	storage *dataflow.ValueStorageInstance,
) (device device.Device, err error) {
	// setup output chain
	output := make(chan dataflow.Value, 128)
	source := dataflow.CreateSource(output)
	// pipe all data to next stage
	source.Append(storage)

	c := &DeviceStruct{
		cfg:      cfg,
		source:   source,
		shutdown: make(chan struct{}),
		closed:   make(chan struct{}),
	}

	if cfg.Kind() == config.VedirectKind {
		err = startVedirect(c, output)
	} else if cfg.Kind() == config.RandomBmvKind {
		err = startRandom(c, output, RegisterListBmv712)
	} else if cfg.Kind() == config.RandomSolarKind {
		err = startRandom(c, output, RegisterListSolar)
	} else {
		return nil, fmt.Errorf("unknown device kind: %s", cfg.Kind().String())
	}

	if err != nil {
		return nil, err
	}

	deviceFilter := dataflow.Filter{IncludeDevices: map[string]bool{cfg.Name(): true}}

	// start mqtt forwarders for realtime messages (send data as soon as it arrives) output
	for _, mc := range mqttClientPool.GetClientsByNames(c.cfg.RealtimeViaMqttClients()) {
		// transmitRealtime values from data store and publish to mqtt broker
		go func() {
			// setup empty filter (everything)
			subscription := storage.Subscribe(deviceFilter)
			defer subscription.Shutdown()

			for {
				select {
				case <-c.shutdown:
					return
				case value := <-subscription.GetOutput():
					if c.cfg.LogDebug() {
						log.Printf(
							"device[%s]->mqttClient[%s]: send realtime : %s",
							c.cfg.Name(), mc.Name(), value,
						)
					}

					if err := mc.PublishRealtimeMessage(value); err != nil {
						log.Printf(
							"device[%s]->mqttClient[%s]: cannot publish realtime: %s",
							c.cfg.Name(), mc.Name(), err,
						)
					}
				}
			}
		}()

		log.Printf(
			"device[%s]->mqttClient[%s]: start sending realtime stat messages",
			c.cfg.Name(), mc.Name(),
		)
	}

	// start mqtt forwarders for telemetry messages
	for _, mc := range mqttClientPool.GetClientsByNames(c.cfg.TelemetryViaMqttClients()) {
		if interval := mc.TelemetryInterval(); interval > 0 {
			go func() {
				ticker := time.NewTicker(interval)
				for {
					select {
					case <-c.shutdown:
						return
					case <-ticker.C:
						if c.cfg.LogDebug() {
							log.Printf(
								"device[%s]->mqttClient[%s]: telemetry tick",
								c.cfg.Name(), mc.Name(), err,
							)
						}

						values := storage.GetSlice(deviceFilter)
						if err := mc.PublishTelemetryMessage(cfg.Name(), c.model, c.LastUpdated(), values); err != nil {
							log.Printf(
								"device[%s]->mqttClient[%s]: cannot publish telemetry: %s",
								c.cfg.Name(), mc.Name(), err,
							)
						}
					}
				}
			}()

			log.Printf(
				"device[%s]->mqttClient[%s]: start sending telemetry messages every %s",
				c.cfg.Name(), mc.Name(), interval.String(),
			)
		}
	}

	return c, nil
}

func (c *DeviceStruct) Name() string {
	return c.cfg.Name()
}

func (c *DeviceStruct) Registers() dataflow.Registers {
	return c.registers
}

func (c *DeviceStruct) SetLastUpdatedNow() {
	c.lastUpdatedMutex.Lock()
	defer c.lastUpdatedMutex.Unlock()
	c.lastUpdated = time.Now()
}

func (c *DeviceStruct) LastUpdated() time.Time {
	c.lastUpdatedMutex.RLock()
	defer c.lastUpdatedMutex.RUnlock()
	return c.lastUpdated
}

func (c *DeviceStruct) Model() string {
	return c.model
}

func (c *DeviceStruct) Shutdown() {
	close(c.shutdown)
	<-c.closed
	log.Printf("device[%s]: shutdown completed", c.cfg.Name())
}
