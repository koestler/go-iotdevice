package config

import (
	"strings"
	"log"
)

type CamConfig struct {
	Name          string
	Type          string
	DirectoryName string
	DebugPrint    bool
}

const camPrefix ="Cam."

func GetCamConfig(sectionName string) (camConfig CamConfig) {
	camConfig = CamConfig{
		Name:          sectionName[len(camPrefix):],
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
		if !strings.HasPrefix(sectionName, camPrefix) {
			continue
		}
		camConfigs = append(camConfigs, GetCamConfig(sectionName))
	}

	return
}
