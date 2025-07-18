package main

import (
	"github.com/koestler/go-iotdevice/v3/config"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/device"
	"github.com/koestler/go-iotdevice/v3/httpServer"
	"github.com/koestler/go-iotdevice/v3/pool"
	"github.com/koestler/go-iotdevice/v3/restarter"
	"log"
)

func runHttpServer(
	cfg *config.Config,
	devicePool *pool.Pool[*restarter.Restarter[device.Device]],
	stateStorage *dataflow.ValueStorage,
	commandStorage *dataflow.ValueStorage,
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
				cfg.LogConfig(),
			},
			ProjectTitle: cfg.ProjectTitle(),
			Views: func(inp []config.ViewConfig) (oup []httpServer.ViewConfig) {
				oup = make([]httpServer.ViewConfig, len(inp))
				for i, r := range inp {
					oup[i] = viewConfig{r}
				}
				return oup
			}(cfg.Views()),
			Authentication: cfg.Authentication(),
			RegisterDbOfDevice: func(deviceName string) *dataflow.RegisterDb {
				return devicePool.GetByName(deviceName).Service().RegisterDb()
			},
			StateStorage:   stateStorage,
			CommandStorage: commandStorage,
		},
	)
}

type httpServerConfig struct {
	config.HttpServerConfig
	logConfig bool
}

func (c httpServerConfig) LogConfig() bool {
	return c.logConfig
}

func (c httpServerConfig) BuildVersion() string {
	return buildVersion
}

type viewConfig struct {
	config.ViewConfig
}

func (c viewConfig) Devices() []httpServer.ViewDeviceConfig {
	devices := c.ViewConfig.Devices()
	ret := make([]httpServer.ViewDeviceConfig, len(devices))
	for i, d := range devices {
		ret[i] = viewDeviceConfig{d}
	}
	return ret
}

type viewDeviceConfig struct {
	config.ViewDeviceConfig
}

func (c viewDeviceConfig) Filter() dataflow.RegisterFilterConf {
	return c.ViewDeviceConfig.Filter()
}
