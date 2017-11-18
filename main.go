package main

import (
	"log"
	"github.com/koestler/go-ve-sensor/cam"
	"github.com/koestler/go-ve-sensor/dataflow"
	"github.com/koestler/go-ve-sensor/config"
	"github.com/koestler/go-ve-sensor/webserver"
)

var rawStorage, roundedStorage *dataflow.ValueStorageInstance

func main() {
	log.Print("main: start go-ve-sensor...")

	setupStorageAndDataFlow()
	setupBmvSources()
	setupTestSinks()
	setupHttpServer()
	setupFtpServer()

	log.Print("main: start completed; run until kill signal is received")

	select {}
}

func setupStorageAndDataFlow() {
	log.Printf("main: setup storage and data flow")

	// setup dataflow pipeline
	// 1. sources:
	// those are appended by separate routines

	// 2. storage for raw values
	rawStorage = dataflow.ValueStorageCreate()

	// 3. rounder
	rounder := dataflow.RounderCreate()

	// 4. storage for rounded values
	roundedStorage = dataflow.ValueStorageCreate()

	// chain those
	rawStorage.Append(rounder)
	rounder.Append(roundedStorage)

}

func setupBmvSources() {
	log.Printf("main: setup Bmv sources")

	sources := BmvDevicesSetupAndRun()
	for _, source := range sources {
		source.Append(rawStorage)
	}
}

func setupTestSinks() {
	// setup some test sinks
	devices := dataflow.DevicesGet()
	f0 := dataflow.Filter{
		Devices:    map[*dataflow.Device]bool{devices[0]: true},
		ValueNames: map[string]bool{"Power": true},
	};
	dataflow.SinkLog("raw    ", rawStorage.Subscribe(f0))
	dataflow.SinkLog("rounded", roundedStorage.Drain())
}

func setupHttpServer() {
	httpdConfig, err := config.GetHttpdConfig()
	if err == nil {
		log.Print("main: start webserver server, config=%v", httpdConfig)

		env := &webserver.Environment{
			RoundedStorage: roundedStorage,
			Devices:        dataflow.DevicesGet(),
		}

		webserver.Run(httpdConfig.Bind, httpdConfig.Port, env)
	} else {
		log.Printf("main: skip webserver server, err=%v", err)
	}
}

func setupFtpServer() {
	// todo: get ftp config
	cam.Run()
}