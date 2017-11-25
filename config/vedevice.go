package config

import (
	"strings"
	"log"
)

type VedeviceConfig struct {
	Name       string
	Model      string
	Device     string
	DebugPrint bool
}

const vedevicePrefix = "Vedevice."

func GetVedeviceConfig(sectionName string) (bmvConfig VedeviceConfig) {
	bmvConfig = VedeviceConfig{
		Name:       sectionName[len(vedevicePrefix):],
		Model:      "unset",
		Device:     "unset",
		DebugPrint: false,
	}

	err := config.Section(sectionName).MapTo(&bmvConfig)

	if err != nil {
		log.Fatal("config: cannot read vedevices configuration: %v", err)
	}

	return
}

func GetVedeviceConfigs() (bmvConfigs []VedeviceConfig) {
	sections := config.SectionStrings()
	for _, sectionName := range sections {
		if !strings.HasPrefix(sectionName, vedevicePrefix) {
			continue
		}
		bmvConfigs = append(bmvConfigs, GetVedeviceConfig(sectionName))
	}

	return
}
