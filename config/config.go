package config

import (
	"github.com/go-ini/ini"
	"log"
)

var config *ini.File

func Setup(source string) {
	log.Printf("config: load configuration source=%v", source)
	var err error
	config, err = ini.Load(source)
	if err != nil {
		log.Fatalf("config: cannot load configuration: %v", err)
	}
}
