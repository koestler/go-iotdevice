package config

import (
	"strings"
	"log"
)

type VedeviceConfigRead struct {
	Name               string
	Model              string
	Device             string
	DebugPrint         bool
	FrontendConfigPath string
}

type VedeviceConfig struct {
	Name           string
	Model          string
	Device         string
	DebugPrint     bool
	FrontendConfig interface{}
}

const vedevicePrefix = "Vedevice."

func GetVedeviceConfig(sectionName string) (bmvConfig *VedeviceConfig) {
	bmvConfigRead := &VedeviceConfigRead{
		Name:               sectionName[len(vedevicePrefix):],
		Model:              "unset",
		FrontendConfigPath: "",
		Device:             "unset",
		DebugPrint:         false,
	}

	err := config.Section(sectionName).MapTo(bmvConfigRead)
	if err != nil {
		log.Fatalf("config: cannot read vedevices configuration: %v", err)
	}

	bmvConfig = &VedeviceConfig{
		Name:       bmvConfigRead.Name,
		Model:      bmvConfigRead.Model,
		Device:     bmvConfigRead.Device,
		DebugPrint: bmvConfigRead.DebugPrint,
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
