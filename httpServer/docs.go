package httpServer

import (
	"github.com/gin-gonic/gin"
)

func setupDocs(engine *gin.Engine, env *Environment) {
	config := env.Config

	serveStatic(engine, config, "/docs", "docs/swagger.html")
	serveStatic(engine, config, "/docs/swagger.yaml", "docs/swagger.yaml")
}
