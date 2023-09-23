package httpServer

import (
	"context"
	"errors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/pool"
	"github.com/koestler/go-iotdevice/restarter"
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
	Config         Config
	ProjectTitle   string
	Views          []config.ViewConfig
	Authentication config.AuthenticationConfig
	DevicePool     *pool.Pool[*restarter.Restarter[device.Device]]
	StateStorage   *dataflow.ValueStorage
	CommandStorage *dataflow.ValueStorage
}

type Config interface {
	BuildVersion() string
	Bind() string
	Port() int
	LogRequests() bool
	LogDebug() bool
	LogConfig() bool
	FrontendProxy() *url.URL
	FrontendPath() string
	GetViewNames() []string
	FrontendExpires() time.Duration
	ConfigExpires() time.Duration
}

func Run(env *Environment) (httpServer *HttpServer) {
	cfg := env.Config

	gin.SetMode("release")
	engine := gin.New()
	if cfg.LogRequests() {
		engine.Use(gin.Logger())
	}
	engine.Use(gin.Recovery())
	engine.Use(authJwtMiddleware(env))

	addApiV2Routes(engine, env)
	setupFrontend(engine, cfg)

	server := &http.Server{
		Addr:    cfg.Bind() + ":" + strconv.Itoa(cfg.Port()),
		Handler: engine,
	}

	go func() {
		if cfg.LogDebug() {
			log.Printf("httpServer: listening on %v", server.Addr)
		}
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Printf("httpServer: stopped due to error: %s", err)
		}
	}()

	return &HttpServer{
		config: cfg,
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

func addApiV2Routes(r *gin.Engine, env *Environment) {
	v2 := r.Group("/api/v2/")
	v2.Use(gzip.Gzip(gzip.BestCompression))
	setupConfig(v2, env)
	setupLogin(v2, env)
	setupRegisters(v2, env)
	setupValuesGetJson(v2, env)
	setupValuesPatch(v2, env)

	v2Ws := r.Group("/api/v2/")
	setupValuesWs(v2Ws, env)
}
