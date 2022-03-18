package httpServer

import (
	"github.com/gin-gonic/gin"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/pkg/errors"
	"log"
	"net/http"
)

type valueResponse interface{}

// setupValues godoc
// @Summary Outputs the latest values of all the fields of a device.
// @ID values
// @Param viewName path string true "View name as provided by the config endpoint"
// @Param deviceName path string true "Device name as provided in devices array of the config endpoint"
// @Produce json
// @success 200 {array} valueResponse
// @Failure 404 {object} ErrorResponse
// @Router /values/{viewName}/{deviceName}.json [get]
// @Security ApiKeyAuth
func setupValues(r *gin.RouterGroup, env *Environment) {
	// add dynamic routes
	for _, v := range env.Views {
		view := v
		for _, deviceName := range view.DeviceNames() {

			device := env.DevicePoolInstance.GetDevice(deviceName)
			if device == nil {
				continue
			}

			relativePath := "values/" + view.Name() + "/" + deviceName + ".json"

			// the follow line uses a loop variable; it must be outside the closure
			deviceFilter := dataflow.Filter{Devices: map[string]bool{deviceName: true}}
			r.GET(relativePath, func(c *gin.Context) {
				// check authorization
				if !isViewAuthenticated(view, c) {
					jsonErrorResponse(c, http.StatusForbidden, errors.New("User is not allowed here"))
					return
				}

				values := env.Storage.GetMap(deviceFilter)
				response := make(map[string]valueResponse, len(values))
				for registerName, value := range values {
					if numeric, ok := value.(dataflow.NumericRegisterValue); ok {
						response[registerName] = numeric.Value()
					} else if text, ok := value.(dataflow.TextRegisterValue); ok {
						response[registerName] = text.Value()
					}

				}

				jsonGetResponse(c, response)
			})
			if env.Config.LogConfig() {
				log.Printf("httpServer: %s%s -> serve fields", r.BasePath(), relativePath)
			}
		}
	}
}
