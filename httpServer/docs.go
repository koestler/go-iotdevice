package httpServer

import (
	"github.com/gin-gonic/gin"
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

func setupDocs(r *gin.RouterGroup, env *Environment) {
	config := env.Config

	serveStatic(r, config, "docs", "docs/swagger.html")
	serveStatic(r, config, "docs/swagger.json", "docs/swagger.json")
	serveStatic(r, config, "docs/swagger.yaml", "docs/swagger.yaml")
}
