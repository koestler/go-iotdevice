package config

import (
	"github.com/go-ini/ini"
	"log"
	"path/filepath"
	"io/ioutil"
	"encoding/json"
)

var config *ini.File
var configDir string

func Setup(source string) {
	configDir = filepath.Dir(source) + "/"

	log.Printf("config: load configuration source=%v, configDir=%v", source, configDir)

	var err error
	config, err = ini.Load(source)
	if err != nil {
		log.Fatalf("config: cannot load configuration: %v", err)
	}
}

func readJsonConfig(frontendConfigPath string) (frontendConfig interface{}) {
	if len(frontendConfigPath) > 0 {
		b, err := ioutil.ReadFile(configDir + frontendConfigPath)
		if err != nil {
			log.Fatalf("config: cannot read frontendConfig file: %v", frontendConfigPath)
		}
		var data interface{}
		err = json.Unmarshal(b, &data)
		if err != nil {
			log.Fatalf("config: cannot decode frontendConfig: %s", b)
		}

		frontendConfig = data
	} else {
		// add empty dictionary
		frontendConfig = make(map[string]string)
	}

	return
}
