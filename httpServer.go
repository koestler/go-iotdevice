package main

import (
	"github.com/koestler/go-victron-to-mqtt/config"
	"github.com/koestler/go-victron-to-mqtt/httpServer"
	"github.com/koestler/go-victron-to-mqtt/vedevices"
	"log"
)

//go:generate swag init -g httpServer/swagger.go

func runHttpServer(cfg *config.Config, devicePoolInstance *vedevices.DevicePool) *httpServer.HttpServer {
	httpServerCfg := cfg.HttpServer()
	if !httpServerCfg.Enabled() {
		return nil
	}

	if cfg.LogWorkerStart() {
		log.Printf("httpServer: start: bind=%s, port=%d", httpServerCfg.Bind(), httpServerCfg.Port())
	}

	return httpServer.Run(
		&httpServer.Environment{
			Config: httpServerConfig{
				cfg.HttpServer(),
				cfg.GetViewNames(),
				cfg.LogConfig(),
				cfg.LogDebug(),
			},
			ProjectTitle:       cfg.ProjectTitle(),
			Views:              cfg.Views(),
			Auth:               cfg.Auth(),
			DevicePoolInstance: devicePoolInstance,
		},
	)
}

type httpServerConfig struct {
	config.HttpServerConfig
	viewNames []string
	logConfig bool
	logDebug  bool
}

func (c httpServerConfig) GetViewNames() []string {
	return c.viewNames
}

func (c httpServerConfig) LogConfig() bool {
	return c.logConfig
}

func (c httpServerConfig) LogDebug() bool {
	return c.logDebug
}

func (c httpServerConfig) BuildVersion() string {
	return buildVersion
}
