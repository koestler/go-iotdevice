package config

import (
	"fmt"
	"errors"
)

type HttpServerConfigRead struct {
	Bind               string
	Port               int
	FrontendConfigPath string
}
type HttpServerConfig struct {
	Bind           string
	Port           int
	FrontendConfig interface{}
}

func GetHttpServerConfig() (httpServerConfig *HttpServerConfig, err error) {
	httpServerConfigRead := &HttpServerConfigRead{
		Bind:               "127.0.0.1",
		Port:               0,
		FrontendConfigPath: "",
	}

	err = config.Section("HttpServer").MapTo(httpServerConfigRead)
	if err != nil {
		return nil, fmt.Errorf("cannot read httpServer configuration: %v", err)
	}

	if httpServerConfigRead.Port == 0 {
		return nil, errors.New("HttpServer: Port missing")
	}

	httpServerConfig = &HttpServerConfig{
		Bind: httpServerConfigRead.Bind,
		Port: httpServerConfigRead.Port,
	}

	httpServerConfig.FrontendConfig = readJsonConfig(httpServerConfigRead.FrontendConfigPath)

	return
}
