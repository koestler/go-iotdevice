package config

import (
	"fmt"
	"errors"
)

type HttpServerConfig struct {
	Bind string
	Port int
}

func GetHttpServerConfig() (httpdConfig *HttpServerConfig, err error) {
	httpdConfig = &HttpServerConfig{
		Bind: "127.0.0.1",
		Port: 0,
	}

	err = config.Section("Httpd").MapTo(httpdConfig)

	if err != nil {
		return nil, fmt.Errorf("cannot read httpd configuration: %v", err)
	}

	if httpdConfig.Port == 0 {
		return nil, errors.New("Httpd:Port missing")
	}

	return
}
