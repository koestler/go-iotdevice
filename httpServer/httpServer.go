package httpServer

import (
	"context"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type HttpServer struct {
	config Config
	server *http.Server
}

type Environment struct {
	Config             Config
	ProjectTitle       string
	Views              []*config.ViewConfig
	Auth               config.AuthConfig
	DevicePoolInstance *device.DevicePool
	Storage            *dataflow.ValueStorageInstance
}

type Config interface {
	BuildVersion() string
	Bind() string
	Port() int
	LogRequests() bool
	LogDebug() bool
	LogConfig() bool
	EnableDocs() bool
	FrontendProxy() *url.URL
	FrontendPath() string
	GetViewNames() []string
	FrontendExpires() time.Duration
	ConfigExpires() time.Duration
}

func Run(env *Environment) (httpServer *HttpServer) {
	config := env.Config

	gin.SetMode("release")
	engine := gin.New()
	if config.LogRequests() {
		engine.Use(gin.Logger())
	}
	engine.Use(gin.Recovery())
	engine.Use(gzip.Gzip(gzip.BestCompression))
	engine.Use(authJwtMiddleware(env))

	if config.EnableDocs() {
		setupSwaggerDocs(engine, config)
	}
	addApiV1Routes(engine, config, env)
	setupFrontend(engine, config)

	server := &http.Server{
		Addr:    config.Bind() + ":" + strconv.Itoa(config.Port()),
		Handler: engine,
	}

	go func() {
		if config.LogDebug() {
			log.Printf("httpServer: listening on %v", server.Addr)
		}
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("httpServer: stopped due to error: %s", err)
		}
	}()

	return &HttpServer{
		config: config,
		server: server,
	}
}

func (s *HttpServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err != nil {
		log.Printf("httpServer: graceful shutdown failed: %s", err)
	}
}

func addApiV1Routes(r *gin.Engine, config Config, env *Environment) {
	v1 := r.Group("/api/v1/")
	setupConfig(v1, env)
	setupLogin(v1, env)
	setupRegisters(v1, env)
	setupValuesJson(v1, env)
	setupValuesWs(v1, env)
	setupHassYaml(v1, env)
}
