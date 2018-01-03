package config

import (
	"log"
)

var VedirectConfig = VedirectConfigStruct{
	DebugPrint: false,
}

type VedirectConfigStruct struct {
	DebugPrint bool
}

func setupVedirect() {
	err := config.Section("Vedirect").MapTo(&VedirectConfig)
	if err != nil {
		log.Printf("cannot read Vedirect configuration: %v", err)
	}
}
