package main

import (
	"github.com/go-ini/ini"
	"log"
	"strings"
)

var config *ini.File

func init() {
	log.Printf("load configuration...")
	var err error
	config, err = ini.Load("config.ini")
	if err != nil {
		log.Fatal("cannot load configuration: %v", err)
	}
}

type HttpdConfig struct {
	Bind string
	Port int
}

func GetHttpdConfig() (httpdConfig *HttpdConfig) {
	httpdConfig = &HttpdConfig{
		Bind: "127.0.0.1",
		Port: 8000,
	}

	err := config.Section("Httpd").MapTo(httpdConfig)

	if err != nil {
		log.Fatal("cannot read httpd configuration: %v", err)
	}

	return
}

type BmvConfig struct {
	Type   string
	Device string
	Aux    string
}

func GetBmvConfig(sectionName string) (bmvConfig BmvConfig) {
	bmvConfig = BmvConfig{
		Type:   "unset",
		Device: "unset",
		Aux:    "none",
	}

	err := config.Section(sectionName).MapTo(&bmvConfig)

	if err != nil {
		log.Fatal("cannot read bmv configuration: %v", err)
	}

	return
}

func GetBmvConfigs() (bmvConfigs []BmvConfig) {

	sections := config.SectionStrings()
	for _, sectionName := range sections {
		if !strings.HasPrefix(sectionName, "Bmv.") {
			continue
		}
		bmvConfigs = append(bmvConfigs, GetBmvConfig(sectionName))
	}

	return
}
