package httpServer

import (
	"github.com/gin-gonic/gin"
	"log"
)

type registerResponse struct {
	Name        string `json:"name" example:"PanelPower"`
	Description string `json:"description" example:"Panel power"`
	Unit        string `json:"unit" example:"W"`
}

// setupRegisters godoc
// @Summary Outputs information about all the available fields.
// @Description Depending on the device model (bmv, bluesolar) a different set of variables are available.
// @Description This endpoint outputs a list of fields (variables) including a name, a unit and a datatype.
// @ID fields
// @Param viewName path string true "View name as provided by the config endpoint"
// @Param deviceName path string true "Device name as provided in devices array of the config endpoint"
// @Produce json
// @success 200 {array} registerResponse
// @Failure 404 {object} ErrorResponse
// @Router /registers/{viewName}/{deviceName}.json [get]
// @Security ApiKeyAuth
func setupRegisters(r *gin.RouterGroup, env *Environment) {
	// add dynamic routes
	for _, v := range env.Views {
		view := v
		for _, deviceName := range view.DeviceNames() {

			device := env.DevicePoolInstance.GetDevice(deviceName)
			if device == nil {
				continue
			}

			registers := device.Registers()
			response := make([]registerResponse, len(registers))
			for i, v := range registers {
				response[i] = registerResponse{
					Name:        v.Name,
					Description: v.Description,
					Unit:        v.Unit,
				}
			}

			relativePath := "registers/" + view.Name() + "/" + deviceName + ".json"
			r.GET(relativePath, func(c *gin.Context) {
				c.JSON(200, response)
			})
			if env.Config.LogConfig() {
				log.Printf("httpServer: %s%s -> serve fields", r.BasePath(), relativePath)
			}
		}
	}
}
