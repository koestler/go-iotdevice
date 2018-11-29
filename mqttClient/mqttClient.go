package mqttClient

import (
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/koestler/go-ve-sensor/config"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var client mqtt.Client

func Run(config *config.MqttClientConfig)  {

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

	go signalHandler()
}

func Stop() {
	client.Disconnect(250)
}

func signalHandler() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM)
	for {
		switch <-ch {
		case syscall.SIGTERM:
			Stop()
			break
		}
	}
}
