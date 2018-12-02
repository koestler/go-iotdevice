package mqttClient

import (
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-ve-sensor/config"
	"github.com/koestler/go-ve-sensor/dataflow"
	"log"
	"os"
	"strings"
	"time"
)

const timeFormat string = "2006-01-02T15:04:05"

type MqttClient struct {
	config *config.MqttClientConfig
	client mqtt.Client
}

func Run(config *config.MqttClientConfig, storage *dataflow.ValueStorageInstance) (mqttClient *MqttClient) {
	// configure client and start connection
	opts := mqtt.NewClientOptions().AddBroker(config.Broker).SetClientID(config.ClientId)
	if len(config.User) > 0 {
		opts.SetUsername(config.User)
	}
	if len(config.Password) > 0 {
		opts.SetPassword(config.Password)
	}

	availableTopic := replaceTemplate(config.AvailableTopic, config)

	if (config.AvailableEnable) {
		opts.SetWill(availableTopic, "Offline", config.Qos, true)
	}

	mqtt.ERROR = log.New(os.Stdout, "", 0)
	if config.DebugLog {
		mqtt.DEBUG = log.New(os.Stdout, "", 0)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("mqttClient connect failed", token.Error())
	}
	log.Printf("mqttClient: connected to %v", config.Broker)

	mqttClient = &MqttClient{
		config: config,
		client: client,
	}

	// send Online
	if (config.AvailableEnable) {
		client.Publish(availableTopic, config.Qos, true, "Online")
	}

	// setup empty filter (everything)
	storageFilter := dataflow.Filter{}

	// setup Realtime (send data as soon as it arrives) output
	if config.RealtimeEnable {
		// transmitRealtime values from data store and publish to mqtt broker
		dataChan := storage.Subscribe(storageFilter)
		log.Print("mqtttClient: start sending realtime stat messages")
		transmitRealtime(dataChan, mqttClient)
	}

	// setup Telemetry support
	if interval, err := time.ParseDuration(config.TelemetryInterval); err == nil && interval > 0 {
		log.Printf("mqtttClient: start sending telemetry messages every %s", interval.String())
		transmitTelemetry(storage, storageFilter, interval, mqttClient)
	}

	return
}

type RealtimeMessage struct {
	Time  string
	Value float64
	Unit  string
}

func convertValueToRealtimeMessage(value dataflow.Value) (RealtimeMessage) {
	return RealtimeMessage{
		Time:  timeToString(time.Now()),
		Value: value.Value,
		Unit:  value.Unit,
	}
}

func transmitRealtime(input <-chan dataflow.Value, mqttClient *MqttClient) {
	go func() {
		cfg := mqttClient.config

		topic := replaceTemplate(cfg.RealtimeTopic, cfg)

		for value := range input {
			if !mqttClient.client.IsConnected() {
				continue
			}

			// replace Device/Value specific palceholders
			topic := strings.Replace(topic, "%DeviceName%", value.Device.Name, 1)
			topic = strings.Replace(topic, "%DeviceModel%", value.Device.Model, 1)
			topic = strings.Replace(topic, "%ValueName%", value.Name, 1)
			topic = strings.Replace(topic, "%ValueUnit%", value.Unit, 1)

			if b, err := json.Marshal(convertValueToRealtimeMessage(value)); err == nil {
				mqttClient.client.Publish(topic, cfg.Qos, cfg.RealtimeRetain, b)
			}
		}
	}()
}

type TelemetryMessage struct {
	Time     string
	NextTele string
	TimeZone string
	Values   map[string]float64
	Units    map[string]string
}

func transmitTelemetry(
	storage *dataflow.ValueStorageInstance,
	filter dataflow.Filter,
	interval time.Duration,
	mqttClient *MqttClient,
) {
	go func() {
		cfg := mqttClient.config

		for now := range time.Tick(interval) {
			for device, deviceState := range storage.GetState(filter) {

				topic := replaceTemplate(cfg.TelemetryTopic, cfg)
				topic = strings.Replace(topic, "%DeviceName%", device.Name, 1)

				values := make(map[string]float64, len(deviceState))
				units := make(map[string]string, len(deviceState))

				for valueName, value := range deviceState.ConvertToEssential() {
					values[valueName] = value.Value
					if len(value.Unit) > 0 {
						units[valueName] = value.Unit
					}
				}

				payload := TelemetryMessage{
					Time:     timeToString(now.UTC()),
					NextTele: timeToString(now.Add(interval)),
					Values:   values,
					Units:    units,
				}

				if b, err := json.Marshal(payload); err == nil {
					mqttClient.client.Publish(topic, cfg.Qos, cfg.TelemetryRetain, b)
				}
			}
		}
	}()
}

func replaceTemplate(template string, config *config.MqttClientConfig) (r string) {
	r = strings.Replace(template, "%Prefix%", config.TopicPrefix, 1)
	r = strings.Replace(r, "%ClientId%", config.ClientId, 1)
	return
}

func timeToString(t time.Time) (string) {
	return t.UTC().Format(timeFormat)
}
