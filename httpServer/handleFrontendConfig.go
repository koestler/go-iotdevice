package httpServer

import (
	"encoding/json"
	"github.com/koestler/go-victron-to-mqtt/config"
	"log"
	"net/http"
)

var frontendConfigSet bool = false
var frontendConfig interface{}

func HandleFrontendConfig(env *Environment, w http.ResponseWriter, r *http.Request) Error {
	if !frontendConfigSet {
		httpServerConfig, err := config.GetHttpServerConfig()
		log.Printf("httpServerConfig=%v", httpServerConfig)
		if err == nil {
			frontendConfig = httpServerConfig.FrontendConfig
			log.Printf("frontendConfig=%v", frontendConfig)
		}
		frontendConfigSet = false
	}

	// cache for 5 minutes
	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJsonHeaders(w)

	b, err := json.MarshalIndent(frontendConfig, "", "    ")
	if err != nil {
		return StatusError{500, err}
	}
	w.Write(b)

	return nil
}
