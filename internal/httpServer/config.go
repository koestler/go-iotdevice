package httpServer

import (
	"log"
	"net/http"
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
// @Summary Frontend configuration
// @Description Return the configuration needed to render a frontend.
// @Description Includes a project title,
// @Description a list of possible views (collection of devices / authentication)
// @Description and for every view names of the devices.
// @Produce json
// @Success 200 {object} configResponse
// @Failure 500 {object} ErrorResponse
// @Router /config/frontend [get]
func setupConfig(mux *http.ServeMux, env *Environment) {
	mux.HandleFunc("GET /api/v2/config/frontend", gzipMiddleware(configHandler(env)))
	if env.Config.LogConfig() {
		log.Printf("httpServer: GET /api/v2/config/frontend -> serve config")
	}
}

func configHandler(env *Environment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := configResponse{
			ProjectTitle:   env.ProjectTitle,
			BackendVersion: env.Config.BuildVersion(),
			Views:          make([]viewResponse, 0),
		}

		for _, v := range env.Views {
			response.Views = append(response.Views, viewResponse{
				Name:  v.Name(),
				Title: v.Title(),
				Devices: func(devices []ViewDeviceConfig) (ret []deviceViewResponse) {
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

		setCacheControlPublic(w, env.Config.ConfigExpires())
		jsonGetResponse(w, r, response)
	}
}
