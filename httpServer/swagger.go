package httpServer

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/koestler/go-iotdevice/docs"
	"github.com/swaggo/files"       // swagger embed files
	"github.com/swaggo/gin-swagger" // gin-swagger middleware
	"log"
	"net/http"
)

// @title go-iotdevice API v1
// @version 1.0
// @description This API exposes all the values of the connected devices, allows for authentication
// and also serves a frontend

// @license.name MIT
// @license.url https://github.com/koestler/go-iotdevice/blob/main/LICENSE

// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func setupSwaggerDocs(r *gin.Engine, config Config) {
	docs.SwaggerInfo.Host = fmt.Sprintf("127.0.0.1:%d", config.Port())
	r.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if config.LogConfig() {
		log.Print("httpServer: /swagger/* -> serve using ginSwagger wrapper")
	}

	r.GET("swagger", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})

	if config.LogConfig() {
		log.Print("httpServer: /swagger -> redirect to /swagger/index.html")
	}
}
