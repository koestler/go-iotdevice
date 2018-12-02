package config

import (
	"errors"
	"fmt"
)

type MqttClientConfig struct {
	Broker            string
	User              string
	Password          string
	ClientId          string
	Qos               byte
	DebugLog          bool
	TopicPrefix       string
	AvailableEnable   bool
	AvailableTopic    string
	TelemetryInterval string
	TelemetryTopic    string
	TelemetryRetain   bool
	RealtimeEnable    bool
	RealtimeTopic     string
	RealtimeRetain    bool
}

func GetMqttClientConfig() (mqttClientConfig *MqttClientConfig, err error) {
	mqttClientConfig = &MqttClientConfig{
		Broker:            "",
		User:              "",
		Password:          "",
		ClientId:          "go-ve-sensor",
		Qos:               1,
		DebugLog:          false,
		TopicPrefix:       "",
		AvailableEnable:   true,
		AvailableTopic:    "%Prefix%%ClientId%/LWT",
		TelemetryInterval: "10s",
		TelemetryTopic:    "%Prefix%tele/%ClientId%/%DeviceName%",
		TelemetryRetain:   false,
		RealtimeEnable:    false,
		RealtimeTopic:     "%Prefix%stat/%ClientId%/%DeviceName%/%ValueName%",
		RealtimeRetain:    true,
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
