package config

import (
	"fmt"
	"errors"
)

type HttpdConfig struct {
	Bind string
	Port int
}

func GetHttpdConfig() (httpdConfig *HttpdConfig, err error) {
	httpdConfig = &HttpdConfig{
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
