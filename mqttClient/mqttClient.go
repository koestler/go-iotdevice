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

type Client struct {
	cfg        Config
	mqttClient mqtt.Client
	shutdown   chan struct{}
}

type Config interface {
	Name() string
	Broker() string
	User() string
	Password() string
	ClientId() string
	Qos() byte
	TopicPrefix() string
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

func RunClient(
	cfg Config,
	devicePoolInstance *device.DevicePool,
	storage *dataflow.ValueStorageInstance,
) (*Client, error) {
	// configure client and start connection
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.Broker()).
		SetClientID(cfg.ClientId()).
		SetOrderMatters(false).
		SetCleanSession(true) // use clean, non-persistent session since we only publish

	if user := cfg.User(); len(user) > 0 {
		opts.SetUsername(user)
	}
	if password := cfg.Password(); len(password) > 0 {
		opts.SetPassword(password)
	}

	// setup availability topic using will
	if availabilityTopic := getAvailabilityTopic(cfg); len(availabilityTopic) > 0 {
		opts.SetWill(availabilityTopic, "offline", cfg.Qos(), true)

		// publish availability after each connect
		opts.SetOnConnectHandler(func(client mqtt.Client) {
			client.Publish(availabilityTopic, cfg.Qos(), true, "online")
		})
	}

	mqtt.ERROR = log.New(os.Stdout, "", 0)
	if cfg.LogDebug() {
		mqtt.DEBUG = log.New(os.Stdout, "", 0)
	}

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("connect failed: %s", token.Error())
	}

	clientStruct := Client{
		cfg:        cfg,
		mqttClient: mqttClient,
		shutdown:   make(chan struct{}),
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

func (c *Client) Config() Config {
	return c.cfg
}

func (c *Client) Shutdown() {
	close(c.shutdown)

	// publish availability offline
	if availabilityTopic := getAvailabilityTopic(c.cfg); len(availabilityTopic) > 0 {
		c.mqttClient.Publish(availabilityTopic, c.cfg.Qos(), true, "offline")
	}

	c.mqttClient.Disconnect(1000)
	log.Printf("mqttClient[%s]: shutdown completed", c.cfg.Name())
}

func getAvailabilityTopic(cfg Config) string {
	return replaceTemplate(cfg.AvailabilityTopic(), cfg)
}

func replaceTemplate(template string, cfg Config) (r string) {
	r = strings.Replace(template, "%Prefix%", cfg.TopicPrefix(), 1)
	r = strings.Replace(r, "%ClientId%", cfg.ClientId(), 1)
	return
}
