package httpServer

import (
	"context"
	"errors"
	"github.com/koestler/go-iotdevice/v3/dataflow"
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

type RegisterDbOfDeviceFunc func(deviceName string) *dataflow.RegisterDb

type Environment struct {
	Config             Config
	ProjectTitle       string
	Views              []ViewConfig
	Authentication     AuthenticationConfig
	RegisterDbOfDevice RegisterDbOfDeviceFunc
	StateStorage       *dataflow.ValueStorage
	CommandStorage     *dataflow.ValueStorage
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
	FrontendExpires() time.Duration
	ConfigExpires() time.Duration
}

type ViewConfig interface {
	Name() string
	Title() string
	Devices() []ViewDeviceConfig
	Autoplay() bool
	IsAllowed(user string) bool
	IsPublic() bool
	Hidden() bool
}

type ViewDeviceConfig interface {
	Name() string
	Title() string
	Filter() dataflow.RegisterFilterConf
}

type AuthenticationConfig interface {
	Enabled() bool
	JwtSecret() []byte
	JwtValidityPeriod() time.Duration
	HtaccessFile() string
}

func Run(env *Environment) (httpServer *HttpServer) {
	cfg := env.Config

	mux := http.NewServeMux()

	// Create handler chain with middleware
	handler := loggingMiddleware(cfg)(
		recoveryMiddleware(
			authJwtMiddleware(env)(mux),
		),
	)

	addApiV2Routes(mux, env)
	setupFrontend(mux, env)

	server := &http.Server{
		Addr:    cfg.Bind() + ":" + strconv.Itoa(cfg.Port()),
		Handler: handler,
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
