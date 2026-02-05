package httpServer

import (
	"net/http"
)

// @title go-iotdevice http API v2
// @version 2.0
// @description Reads parameters from Victron Energy Battery Monitor, Solar Chargers and other devices and exposed them via a REST api.
// @externalDocs.url https://github.com/koestler/go-iotdevice/

// @license.name MIT
// @license.url https://github.com/koestler/go-iotdevice/blob/main/LICENSE

// @BasePath /api/v2
// @schemes http

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func setupDocs(mux *http.ServeMux, env *Environment) {
	config := env.Config

	serveStatic(mux, config, "/api/v2/docs", "docs/swagger.html")
	serveStatic(mux, config, "/api/v2/docs/swagger.json", "docs/swagger.json")
	serveStatic(mux, config, "/api/v2/docs/swagger.yaml", "docs/swagger.yaml")
}
