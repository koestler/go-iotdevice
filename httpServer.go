package main

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
	"github.com/koestler/go-iotdevice/httpServer"
	"github.com/koestler/go-iotdevice/pool"
	"github.com/koestler/go-iotdevice/restarter"
	"log"
)

func runHttpServer(
	cfg *config.Config,
	devicePoolInstance *pool.Pool[*restarter.Restarter[device.Device]],
	stateStorage *dataflow.ValueStorageInstance,
	commandStorage *dataflow.ValueStorageInstance,
) *httpServer.HttpServer {
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
			},
			ProjectTitle:       cfg.ProjectTitle(),
			Views:              cfg.Views(),
			Authentication:     cfg.Authentication(),
			DevicePoolInstance: devicePoolInstance,
			StateStorage:       stateStorage,
			CommandStorage:     commandStorage,
		},
	)
}

type httpServerConfig struct {
	config.HttpServerConfig
	viewNames []string
	logConfig bool
}

func (c httpServerConfig) GetViewNames() []string {
	return c.viewNames
}

func (c httpServerConfig) LogConfig() bool {
	return c.logConfig
}

func (c httpServerConfig) BuildVersion() string {
	return buildVersion
}
