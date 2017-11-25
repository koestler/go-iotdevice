package config

import (
	"github.com/go-ini/ini"
	"log"
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
