package config

import (
	"errors"
	"fmt"
	"github.com/go-ini/ini"
	"log"
	"strings"
)

var config *ini.File

func init() {
	log.Printf("config: load configuration...")
	var err error
	config, err = ini.Load("config.ini")
	if err != nil {
		log.Fatal("config: cannot load configuration: %v", err)
	}
}

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

type MongoConfig struct {
	MongoHost          string
	DatabaseName       string
	RawValuesIntervall int
}

func GetMongoConfig() (mongoConfig *MongoConfig, err error) {
	mongoConfig = &MongoConfig{
		MongoHost:          "127.0.0.1",
		DatabaseName:       "go-ve-sensor",
		RawValuesIntervall: 2000,
	}

	// check if mongo sections exists
	_, err = config.GetSection("Mongo")
	if err != nil {
		// Section Mongo does not exist
		return nil, errors.New("no mongo configuration found")
	}

	err = config.Section("Mongo").MapTo(mongoConfig)

	if err != nil {
		return nil, fmt.Errorf("cannot read mongo configuration: %v", err)
	}

	return
}

type BmvConfig struct {
	Name       string
	Model      string
	Device     string
	Aux        string
	DebugPrint bool
}

func GetBmvConfig(sectionName string) (bmvConfig BmvConfig) {
	bmvConfig = BmvConfig{
		Name:       sectionName[4:],
		Model:      "unset",
		Device:     "unset",
		Aux:        "none",
		DebugPrint: false,
	}

	err := config.Section(sectionName).MapTo(&bmvConfig)

	if err != nil {
		log.Fatal("config: cannot read bmv configuration: %v", err)
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

type CamConfig struct {
	Name          string
	Type          string
	DirectoryName string
	DebugPrint    bool
}

func GetCamConfig(sectionName string) (camConfig CamConfig) {
	camConfig = CamConfig{
		Name:          sectionName[4:],
		Type:          "unset",
		DirectoryName: "unset",
		DebugPrint:    false,
	}

	err := config.Section(sectionName).MapTo(&camConfig)

	if err != nil {
		log.Fatal("config: cannot read cam configuration: %v", err)
	}

	return
}

func GetCamConfigs() (camConfigs []CamConfig) {

	sections := config.SectionStrings()
	for _, sectionName := range sections {
		if !strings.HasPrefix(sectionName, "Cam.") {
			continue
		}
		camConfigs = append(camConfigs, GetCamConfig(sectionName))
	}

	return
}
