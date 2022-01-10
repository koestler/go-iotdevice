package httpServer

import (
	"github.com/gin-gonic/gin"
	"log"
)

// setupFields godoc
// @Summary Outputs information about all the available fields.
// @Description Depending on the device model (bmv, bluesolar) a different set of variables are available.
// @Description This endpoint outputs a list of fields (variables) including a name, a unit and a datatype.
// @ID fields
// @Produce json
// @Failure 404 {object} ErrorResponse
// @Router /fields/{viewName}/{deviceName}.json [get]
// @Security ApiKeyAuth
func setupFields(r *gin.RouterGroup, env *Environment) {
	// add dynamic routes
	for _, v := range env.Views {
		view := v
		for _, deviceName := range view.DeviceNames() {

			//client := env.CameraClientPoolInstance.GetClient(camera)
			//if client == nil {
			//continue
			//}

			relativePath := "fields/" + view.Name() + "/" + deviceName + ".json"
			r.GET(relativePath, func(c *gin.Context) {
				c.JSON(200, struct{}{})
			})
			if env.Config.LogConfig() {
				log.Printf("httpServer: %s%s -> serve fields", r.BasePath(), relativePath)
			}
		}
	}
}
