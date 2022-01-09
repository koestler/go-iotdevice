package mqttClient

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-victron-to-mqtt/config"
	"github.com/koestler/go-victron-to-mqtt/dataflow"
	"log"
	"os"
	"strings"
	"time"
)

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

	availableTopic := GetAvailableTopic(config)

	if config.AvailableEnable {
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
	if config.AvailableEnable {
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

func GetAvailableTopic(cfg *config.MqttClientConfig) string {
	return replaceTemplate(cfg.AvailableTopic, cfg)
}

func replaceTemplate(template string, config *config.MqttClientConfig) (r string) {
	r = strings.Replace(template, "%Prefix%", config.TopicPrefix, 1)
	r = strings.Replace(r, "%ClientId%", config.ClientId, 1)
	return
}
