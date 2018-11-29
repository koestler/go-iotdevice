package config

import (
	"errors"
	"fmt"
)

type MqttClientConfig struct {
	Broker string
	User string
	Password string
	ClientId string
	Qos byte
	DebugLog bool
	AvailableTopic string
}

func GetMqttClientConfig() (mqttClientConfig *MqttClientConfig, err error) {
	mqttClientConfig = &MqttClientConfig{
		Broker: "",
		User: "",
		Password: "",
		ClientId: "go-ve-sensor",
		Qos: 1,
		DebugLog: false,
		AvailableTopic: "go-ve-sensor/LWT",
	}

	// check if mqttClient sections exists
	_, err = config.GetSection("MqttClient")
	if err != nil {
		return nil, errors.New("no mqttClient configuration found")
	}

	err = config.Section("MqttClient").MapTo(mqttClientConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot read mqttClient configuration: %v", err)
	}

	if len(mqttClientConfig.Broker) < 1 {
		return nil, errors.New("mqttClient: Broker not specified")
	}

	if len(mqttClientConfig.ClientId) < 1 {
		return nil, errors.New("mqttClient: ClientId not specified")
	}

	return
}
