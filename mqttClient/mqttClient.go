package mqttClient

import (
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-ve-sensor/config"
	"github.com/koestler/go-ve-sensor/dataflow"
	"log"
	"os"
	"strings"
)

var client mqtt.Client

func Run(config *config.MqttClientConfig, storage *dataflow.ValueStorageInstance) {
	// configure client and start connection
	opts := mqtt.NewClientOptions().AddBroker(config.Broker).SetClientID(config.ClientId)
	if len(config.User) > 0 {
		opts.SetUsername(config.User)
	}
	if len(config.Password) > 0 {
		opts.SetPassword(config.Password)
	}

	opts.SetWill(config.AvailableTopic, "Offline", config.Qos, true)

	mqtt.ERROR = log.New(os.Stdout, "", 0)
	if config.DebugLog {
		mqtt.DEBUG = log.New(os.Stdout, "", 0)
	}

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("mqttClient connect failed", token.Error())
	}
	log.Printf("mqttClient: connected to %v", config.Broker)

	client.Publish(config.AvailableTopic, config.Qos, true, "Online")

	// sink values from data store and publish to mqtt broker
	dataChan := storage.Subscribe(dataflow.Filter{})
	sink(dataChan, config.Qos, config.ValueTopic)
}

type Message struct {
	Value float64
	Unit  string
}

func convertValueToMessage(value dataflow.Value) (Message) {
	return Message{
		Value: value.Value,
		Unit:  value.Unit,
	}
}

func sink(input <-chan dataflow.Value, qos byte, topicTemplate string) {
	go func() {
		for value := range input {
			if !client.IsConnected() {
				continue
			}

			topic := strings.Replace(topicTemplate, "%DeviceName%", value.Device.Name, 1)
			topic = strings.Replace(topic, "%DeviceModel%", value.Device.Model, 1)
			topic = strings.Replace(topic, "%ValueName%", value.Name, 1)
			topic = strings.Replace(topic, "%ValueUnit%", value.Unit, 1)

			if b, err := json.Marshal(convertValueToMessage(value)); err == nil {
				client.Publish(topic, qos, false, b)
			}
		}
	}()
}
