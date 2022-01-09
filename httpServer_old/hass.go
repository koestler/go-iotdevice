package httpServer

import (
	"github.com/koestler/go-victron-to-mqtt/config"
	"github.com/koestler/go-victron-to-mqtt/mqttClient"
	"github.com/koestler/go-victron-to-mqtt/vedevices"
	"net/http"
	"strings"
)

// Example YMAL
// - platform: mqtt
//   name:                  "ve_24v_bmv_current"
//   state_topic:           "piegn/stat/ve/24v-bmv/Current"
//   availability_topic:    "piegn/tele/software/srv1-go-victron-to-mqtt/LWT"
//   value_template:        "{{ value_json.Value }}"
//   unit_of_measurement:   "W"
//   payload_available:     "Online"
//   payload_not_available: "Offline"
type hassSensor struct {
	Platform            string `yaml:"platform"`
	Name                string `yaml:"name"`
	StateTopic          string `yaml:"state_topic"`
	AvailabilityTopic   string `yaml:"availability_topic"`
	ValueTemplate       string `yaml:"value_template"`
	UnitOfMeasurement   string `yaml:"unit_of_measurement"`
	PayloadAvailable    string `yaml:"payload_available"`
	PayloadNotAvailable string `yaml:"payload_not_available"`
}

func HandleHassMqttSensorsYaml(env *Environment, w http.ResponseWriter, r *http.Request) Error {
	configs := make([]hassSensor, 0)
	for _, device := range env.Devices {
		registers := vedevices.RegisterFactoryByProduct(device.DeviceId)
		if registers == nil {
			continue
		}

		for valueName, register := range registers {
			configs = append(configs,
				registerToHassSensor(
					env.MqttClientConfig,
					device.Name,
					device.Model,
					valueName,
					register.Unit,
				),
			)
		}
	}

	writeYamlHeaders(w)
	b, err := yaml.Marshal(configs)
	if err != nil {
		return StatusError{500, err}
	}
	w.Write(b)
	return nil
}

func registerToHassSensor(
	mqttClientConfig *config.MqttClientConfig,
	deviceName string,
	deviceModel string,
	valueName string,
	unit string,
) hassSensor {
	return hassSensor{
		Platform: "mqtt",
		Name:     cleanupHassName(deviceName) + "_" + cleanupHassName(valueName),
		StateTopic: mqttClient.GetRealtimeTopic(
			mqttClientConfig,
			deviceName,
			deviceModel,
			valueName,
			unit,
		),
		AvailabilityTopic:   mqttClient.GetAvailableTopic(mqttClientConfig),
		ValueTemplate:       "{{ value_json.Value }}",
		UnitOfMeasurement:   unit,
		PayloadAvailable:    "Online",
		PayloadNotAvailable: "Offline",
	}
}

var hassNameReplace = strings.NewReplacer("-", "_")

func cleanupHassName(i string) string {
	return hassNameReplace.Replace(i)
}
