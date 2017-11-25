package config

import (
	"strings"
	"log"
	"errors"
	"fmt"
)

type FtpServerConfig struct {
	Bind     string
	Port     int
	DebugLog bool
}

type FtpCameraConfig struct {
	Name     string
	Password string
}

func GetFtpServerConfig() (ftpConfig *FtpServerConfig, err error) {
	ftpConfig = &FtpServerConfig{
		Bind:     "127.0.0.1",
		Port:     21,
		DebugLog: false,
	}

	// check if ftpServer sections exists
	_, err = config.GetSection("FtpServer")
	if err != nil {
		return nil, errors.New("no ftpServer configuration found")
	}

	err = config.Section("FtpServer").MapTo(ftpConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot read ftpServer configuration: %v", err)
	}

	return
}

const ftpCameraPrefix = "FtpCamera."

func GetFtpCameraConfig(sectionName string) (cameraConfig *FtpCameraConfig) {
	cameraConfig = &FtpCameraConfig{
		Name:     sectionName[len(ftpCameraPrefix):],
		Password: "empty",
	}

	err := config.Section(sectionName).MapTo(&cameraConfig)

	if err != nil {
		log.Fatal("config: cannot read ftpServer configuration: %v", err)
	}

	return
}

func GetFtpCameraConfigs() (cameraConfigs []*FtpCameraConfig) {
	sections := config.SectionStrings()
	for _, sectionName := range sections {
		if !strings.HasPrefix(sectionName, ftpCameraPrefix) {
			continue
		}
		cameraConfigs = append(cameraConfigs, GetFtpCameraConfig(sectionName))
	}

	return
}
