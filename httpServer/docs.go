package httpServer

import (
	"net/http"
	"os"
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

	docsFS := os.DirFS("docs")

	serveStatic(mux, config, docsFS, "swagger.html", "/api/v2/docs/swagger.html")
	serveStatic(mux, config, docsFS, "swagger.json", "/api/v2/docs/swagger.json")
	serveStatic(mux, config, docsFS, "swagger.yaml", "/api/v2/docs/swagger.yaml")
}
