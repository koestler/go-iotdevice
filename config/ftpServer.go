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

type FtpCameraConfigRead struct {
	Name               string
	User               string
	Password           string
	FrontendConfigPath string
}

type FtpCameraConfig struct {
	Name           string
	User           string
	Password       string
	FrontendConfig interface{}
}

func GetFtpServerConfig() (ftpConfig *FtpServerConfig, err error) {
	ftpConfig = &FtpServerConfig{
		Bind:     "",
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
	cameraConfigRead := &FtpCameraConfigRead{
		Name:               sectionName[len(ftpCameraPrefix):],
		User:               "",
		Password:           "",
		FrontendConfigPath: "",
	}

	err := config.Section(sectionName).MapTo(&cameraConfigRead)
	if err != nil {
		log.Fatal("config: cannot read ftpServer configuration: %v", err)
	}

	cameraConfig = &FtpCameraConfig{
		Name:     cameraConfigRead.Name,
		User:     cameraConfigRead.User,
		Password: cameraConfigRead.Password,
	}

	cameraConfig.FrontendConfig = readJsonConfig(cameraConfigRead.FrontendConfigPath)

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
