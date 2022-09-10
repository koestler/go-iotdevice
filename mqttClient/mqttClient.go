package mqttClient

import (
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"log"
	"os"
	"strings"
	"time"
)

type Config interface {
	Name() string
	Broker() string
	User() string
	Password() string
	ClientId() string
	Qos() byte
	TopicPrefix() string
	AvailabilityEnable() bool
	AvailabilityTopic() string
	TelemetryInterval() time.Duration
	TelemetryTopic() string
	TelemetryRetain() bool
	RealtimeEnable() bool
	RealtimeTopic() string
	RealtimeQos() byte
	RealtimeRetain() bool
	LogDebug() bool
}

type Client interface {
	Config() Config
	Shutdown()
}

type ClientStruct struct {
	cfg        Config
	mqttClient mqtt.Client
	shutdown   chan struct{}
}

func RunClient(
	cfg Config,
	devicePoolInstance *device.DevicePool,
	storage *dataflow.ValueStorageInstance,
) (client Client, err error) {
	// configure client and start connection
	opts := mqtt.NewClientOptions().AddBroker(cfg.Broker()).SetClientID(cfg.ClientId())
	if len(cfg.User()) > 0 {
		opts.SetUsername(cfg.User())
	}
	if len(cfg.Password()) > 0 {
		opts.SetPassword(cfg.Password())
	}

	if cfg.AvailabilityEnable() {
		// public availability offline as will
		opts.SetWill(GetAvailabilityTopic(cfg), "offline", cfg.Qos(), true)
	}

	mqtt.ERROR = log.New(os.Stdout, "", 0)
	if cfg.LogDebug() {
		mqtt.DEBUG = log.New(os.Stdout, "", 0)
	}

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("connect failed: %s", token.Error())
	}

	clientStruct := ClientStruct{
		cfg:        cfg,
		mqttClient: mqttClient,
		shutdown:   make(chan struct{}),
	}

	// send Online
	if cfg.AvailabilityEnable() {
		mqttClient.Publish(GetAvailabilityTopic(cfg), cfg.Qos(), true, "online")
	}

	// setup Realtime (send data as soon as it arrives) output
	if cfg.RealtimeEnable() {
		// transmitRealtime values from data store and publish to mqtt broker
		go func() {
			// setup empty filter (everything)
			subscription := storage.Subscribe(dataflow.Filter{})

			defer subscription.Shutdown()
			for {
				select {
				case <-clientStruct.shutdown:
					return
				case value := <-subscription.GetOutput():
					if !mqttClient.IsConnected() {
						continue
					}

					if b, err := convertValueToRealtimeMessage(value); err == nil {
						mqttClient.Publish(
							getRealtimeTopic(
								cfg,
								value.DeviceName(),
								value.Register().Name(),
								value.Register().Unit(),
							),
							cfg.RealtimeQos(),
							cfg.RealtimeRetain(),
							b,
						)
					}
				}
			}
		}()
		log.Printf("mqttClient[%s]: start sending realtime stat messages", cfg.Name())
	}

	// setup Telemetry support
	if interval := cfg.TelemetryInterval(); interval > 0 {
		go func() {
			ticker := time.NewTicker(interval)
			for {
				select {
				case <-clientStruct.shutdown:
					return
				case <-ticker.C:
					for deviceName, dev := range devicePoolInstance.GetDevices() {
						deviceFilter := dataflow.Filter{IncludeDevices: map[string]bool{deviceName: true}}
						values := storage.GetSlice(deviceFilter)

						now := time.Now()
						payload := TelemetryMessage{
							Time:                   timeToString(now),
							NextTelemetry:          timeToString(now.Add(interval)),
							Model:                  dev.GetModel(),
							SecondsSinceLastUpdate: now.Sub(dev.GetLastUpdated()).Seconds(),
							NumericValues:          convertValuesToNumericTelemetryValues(values),
							TextValues:             convertValuesToTextTelemetryValues(values),
						}

						if b, err := json.Marshal(payload); err == nil {
							mqttClient.Publish(
								getTelemetryTopic(cfg, deviceName),
								cfg.Qos(),
								cfg.TelemetryRetain(),
								b,
							)
						}
					}
				}
			}
		}()

		log.Printf("mqttClient[%s]: start sending telemetry messages every %s", cfg.Name(), interval.String())
	}

	return &clientStruct, nil
}

func (c ClientStruct) Config() Config {
	return c.cfg
}

func (c ClientStruct) Shutdown() {
	close(c.shutdown)

	// publish availability offline
	if c.cfg.AvailabilityEnable() {
		c.mqttClient.Publish(GetAvailabilityTopic(c.cfg), c.cfg.Qos(), true, "offline")
	}

	c.mqttClient.Disconnect(1000)
	log.Printf("mqttClient[%s]: shutdown completed", c.cfg.Name())
}

func GetAvailabilityTopic(cfg Config) string {
	return replaceTemplate(cfg.AvailabilityTopic(), cfg)
}

func replaceTemplate(template string, cfg Config) (r string) {
	r = strings.Replace(template, "%Prefix%", cfg.TopicPrefix(), 1)
	r = strings.Replace(r, "%ClientId%", cfg.ClientId(), 1)
	return
}
