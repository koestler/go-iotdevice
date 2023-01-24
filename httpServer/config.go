package httpServer

import (
	"github.com/gin-gonic/gin"
	"github.com/koestler/go-iotdevice/config"
	"log"
)

type configResponse struct {
	ProjectTitle   string         `json:"projectTitle" example:"go-iotdevice"`
	BackendVersion string         `json:"backendVersion" example:"v1.2.3"`
	Views          []viewResponse `json:"views"`
}

type viewResponse struct {
	Name     string               `json:"name" example:"public"`
	Title    string               `json:"title" example:"Outlook"`
	Devices  []deviceViewResponse `json:"devices"`
	Autoplay bool                 `json:"autoplay" example:"True"`
	IsPublic bool                 `json:"isPublic" example:"False"`
	Hidden   bool                 `json:"hidden" example:"False"`
}

type deviceViewResponse struct {
	Name  string `json:"name" example:"0-cam-east"`
	Title string `json:"title" example:"East"`
}

// setupConfig godoc
// @Summary Frontend configuration structure
// @Description Includes a project title,
// @Description a list of possible views (collection of devices / authentication)
// @Description and for every view names of the devices.
// @ID config
// @Produce json
// @Success 200 {object} configResponse
// @Failure 500 {object} ErrorResponse
// @Router /config [get]
func setupConfig(r *gin.RouterGroup, env *Environment) {
	r.GET("config", func(c *gin.Context) {
		response := configResponse{
			ProjectTitle:   env.ProjectTitle,
			BackendVersion: env.Config.BuildVersion(),
			Views:          make([]viewResponse, 0),
		}

		for _, v := range env.Views {
			response.Views = append(response.Views, viewResponse{
				Name:  v.Name(),
				Title: v.Title(),
				Devices: func(devices []*config.ViewDeviceConfig) (ret []deviceViewResponse) {
					ret = make([]deviceViewResponse, len(devices))
					for i, c := range devices {
						ret[i] = deviceViewResponse{
							Name:  c.Name(),
							Title: c.Title(),
						}
					}
					return
				}(v.Devices()),
				Autoplay: v.Autoplay(),
				IsPublic: v.IsPublic(),
				Hidden:   v.Hidden(),
			})
		}

		setCacheControlPublic(c, env.Config.ConfigExpires())
		jsonGetResponse(c, response)
	})
	if env.Config.LogConfig() {
		log.Printf("httpServer: GET %sconfig -> serve config", r.BasePath())
	}
}
