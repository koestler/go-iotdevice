package httpServer

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func addApiV2Routes(r *gin.Engine, env *Environment) {
	v2 := r.Group("/api/v2/")
	v2.Use(gzip.Gzip(gzip.BestCompression))
	setupConfig(v2, env)
	setupLogin(v2, env)
	setupRegisters(v2, env)
	setupValuesGetJson(v2, env)
	setupValuesPatch(v2, env)
	setupDocs(v2, env)

	v2Ws := r.Group("/api/v2/")
	setupValuesWs(v2Ws, env)
}
