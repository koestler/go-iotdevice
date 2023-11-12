package httpServer

import (
	"github.com/gin-gonic/gin"
)

func setupDocs(r *gin.RouterGroup, env *Environment) {
	config := env.Config

	serveStatic(r, config, "docs", "docs/swagger.html")
	serveStatic(r, config, "docs/swagger.json", "docs/swagger.json")
	serveStatic(r, config, "docs/swagger.yaml", "docs/swagger.yaml")
}
