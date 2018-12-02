package config

import (
	"errors"
	"fmt"
)

type HttpServerConfigRead struct {
	Bind               string
	Port               int
	FrontendConfigPath string

	// empty string: logging disabled
	// -           : log to stdoud (default)
	// else        : used as file path for a log file
	LogFile string
}
type HttpServerConfig struct {
	Bind           string
	Port           int
	FrontendConfig interface{}
	LogFile        string
}

func GetHttpServerConfig() (httpServerConfig *HttpServerConfig, err error) {
	httpServerConfigRead := &HttpServerConfigRead{
		Bind:               "127.0.0.1",
		Port:               0,
		FrontendConfigPath: "",
		LogFile:            "-",
	}

	err = config.Section("HttpServer").MapTo(httpServerConfigRead)
	if err != nil {
		return nil, fmt.Errorf("cannot read httpServer configuration: %v", err)
	}

	if httpServerConfigRead.Port == 0 {
		return nil, errors.New("HttpServer: Port missing")
	}

	httpServerConfig = &HttpServerConfig{
		Bind:    httpServerConfigRead.Bind,
		Port:    httpServerConfigRead.Port,
		LogFile: httpServerConfigRead.LogFile,
	}

	httpServerConfig.FrontendConfig = readJsonConfig(httpServerConfigRead.FrontendConfigPath)

	return
}
