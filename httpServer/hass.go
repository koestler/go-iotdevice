package httpServer

import (
	"github.com/gin-gonic/gin"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"strings"
)

type hassSensor struct {
	Platform            string  `yaml:"platform"`
	Name                string  `yaml:"name"`
	StateTopic          string  `yaml:"state_topic"`
	AvailabilityTopic   string  `yaml:"availability_topic"`
	ValueTemplate       string  `yaml:"value_template"`
	UnitOfMeasurement   *string `yaml:"unit_of_measurement"`
	PayloadAvailable    string  `yaml:"payload_available"`
	PayloadNotAvailable string  `yaml:"payload_not_available"`
}

// setupHassYaml godoc
// @Summary Outputs a homeassistant configuration file setting fetching of the available values via mqtt.
// @ID hassYaml
// @Param viewName path string true "View name as provided by the config endpoint"
// @Produce json
// @success 200 {array} valueResponse
// @Failure 404 {object} ErrorResponse
// @Router /hass/{viewName}.yaml [get]
// @Security ApiKeyAuth
// Example YMAL
// - platform: mqtt
//   name:                  "ve_24v_bmv_current"
//   state_topic:           "piegn/stat/ve/24v-bmv/Current"
//   availability_topic:    "piegn/tele/software/srv1-go-iotdevice/LWT"
//   value_template:        "{{ value_json.Value }}"
//   unit_of_measurement:   "W"
//   payload_available:     "Online"
//   payload_not_available: "Offline"
func setupHassYaml(r *gin.RouterGroup, env *Environment) {
	for _, v := range env.Views {
		view := v

		relativePath := "hass/" + view.Name() + ".yaml"
		r.GET(relativePath, func(c *gin.Context) {
			// check authorization
			if !isViewAuthenticated(view, c) {
				jsonErrorResponse(c, http.StatusForbidden, errors.New("User is not allowed here"))
				return
			}

			// compile response
			sensors := make([]hassSensor, 0, 128)
			for _, deviceName := range view.DeviceNames() {
				device := env.DevicePoolInstance.GetDevice(deviceName)
				for _, register := range device.GetRegisters() {
					sensors = append(sensors, registerToHassSensor(deviceName, register))
				}

			}
			yamlGetResponse(c, sensors)
		})
		if env.Config.LogConfig() {
			log.Printf("httpServer: %s%s -> serve hass configuration as yaml", r.BasePath(), relativePath)
		}

	}
}

func registerToHassSensor(
	deviceName string,
	register dataflow.Register,
) hassSensor {
	return hassSensor{
		Platform:            "mqtt",
		Name:                cleanupHassName(deviceName) + "-" + cleanupHassName(register.Name()),
		StateTopic:          "TODO-state-topic",
		AvailabilityTopic:   "TODO-availability-topic",
		ValueTemplate:       "{{ value_json.Value }}",
		UnitOfMeasurement:   register.Unit(),
		PayloadAvailable:    "Online",
		PayloadNotAvailable: "Offline",
	}
}

var hassNameReplace = strings.NewReplacer("-", "_")

func cleanupHassName(i string) string {
	return hassNameReplace.Replace(i)
}
