package config

import (
	"strings"
	"log"
)

type VedeviceConfigRead struct {
	Name               string
	Model              string
	Device             string
	FrontendConfigPath string
}

type VedeviceConfig struct {
	Name           string
	Model          string
	Device         string
	FrontendConfig interface{}
}

const vedevicePrefix = "Vedevice."

func GetVedeviceConfig(sectionName string) (bmvConfig *VedeviceConfig) {
	bmvConfigRead := &VedeviceConfigRead{
		Name:               sectionName[len(vedevicePrefix):],
		Model:              "unset",
		FrontendConfigPath: "",
		Device:             "unset",
	}

	err := config.Section(sectionName).MapTo(bmvConfigRead)
	if err != nil {
		log.Fatalf("config: cannot read vedevices configuration: %v", err)
	}

	bmvConfig = &VedeviceConfig{
		Name:       bmvConfigRead.Name,
		Model:      bmvConfigRead.Model,
		Device:     bmvConfigRead.Device,
	}

	bmvConfig.FrontendConfig = readJsonConfig(bmvConfigRead.FrontendConfigPath)

	return
}

func GetVedeviceConfigs() (bmvConfigs []*VedeviceConfig) {
	sections := config.SectionStrings()
	for _, sectionName := range sections {
		if !strings.HasPrefix(sectionName, vedevicePrefix) {
			continue
		}
		bmvConfigs = append(bmvConfigs, GetVedeviceConfig(sectionName))
	}

	return
}
