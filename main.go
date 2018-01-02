package main

import (
	"log"
	"github.com/koestler/go-ve-sensor/ftpServer"
	"github.com/koestler/go-ve-sensor/dataflow"
	"github.com/koestler/go-ve-sensor/config"
	"github.com/koestler/go-ve-sensor/httpServer"
	"github.com/koestler/go-ve-sensor/storage"
	"github.com/koestler/go-ve-sensor/vedevices"
	"github.com/jessevdk/go-flags"
	"os"
)

type CmdOptions struct {
	Config flags.Filename `short:"c" long:"config" description:"Config File in ini format" default:"./config.ini"`
}

var cmdOptions CmdOptions

var rawStorage, roundedStorage *dataflow.ValueStorageInstance

func main() {
	log.Print("main: start go-ve-sensor...")

	setupConfig()
	setupStorageAndDataFlow()
	setupBmvDevices()
	setupCameraDevices()
	//setupTestSinks()
	setupHttpServer()
	setupFtpServer()

	log.Print("main: start completed; run until kill signal is received")

	select {}
}

func setupConfig() {
	log.Printf("main: setup config")

	// parse command line options
	parser := flags.NewParser(&cmdOptions, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
	// initialize config library
	config.Setup(string(cmdOptions.Config))
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

func setupBmvDevices() {
	log.Printf("main: setup Bmv Devices")

	configs := config.GetVedeviceConfigs()

	sources := make([]dataflow.Drainable, 0, len(configs))

	// get devices from database and create them
	for _, c := range configs {
		log.Printf(
			"bmvDevices: setup name=%v model=%v device=%v",
			c.Name, c.Model, c.Device,
		)

		// register device in storage
		device := storage.DeviceCreate(c.Name, c.Model, c.FrontendConfig);

		// setup the datasource
		if "dummy" == c.Device {
			sources = append(sources, vedevices.CreateDummySource(device, c))
		} else {
			if err, source := vedevices.CreateSource(device, c); err == nil {
				sources = append(sources, source)
			} else {
				log.Printf("bmvDevices: error while CreateSoruce: %v", err)
			}
		}
	}

	// append them as sources to the raw storage
	for _, source := range sources {
		source.Append(rawStorage)
	}
}

func setupCameraDevices() {
	log.Printf("main: setup Camera Devices")

	cameras := config.GetFtpCameraConfigs()

	for _, camera := range cameras {
		storage.DeviceCreate(camera.Name, "ftpCamera", camera.FrontendConfig);
	}
}

func setupTestSinks() {
	// setup some test sinks
	/*
	devices := dataflow.DevicesGet()
	f0 := dataflow.Filter{
		Devices:    map[*dataflow.Device]bool{devices[0]: true},
		ValueNames: map[string]bool{"Power": true},
	};
	dataflow.SinkLog("raw    ", rawStorage.Subscribe(f0))
	*/
	dataflow.SinkLog("rounded", roundedStorage.Drain())
}

func setupHttpServer() {
	httpServerConfig, err := config.GetHttpServerConfig()
	if err == nil {
		log.Printf("main: start httpServer, Bind=%v, Port=%v", httpServerConfig.Bind, httpServerConfig.Port)

		env := &httpServer.Environment{
			RoundedStorage: roundedStorage,
			Devices:        storage.GetAll(),
		}

		httpServer.Run(httpServerConfig.Bind, httpServerConfig.Port, env)
	} else {
		log.Printf("main: skip httpServer, err=%v", err)
	}
}

func setupFtpServer() {
	ftpServerConfig, err := config.GetFtpServerConfig()
	if err == nil {
		log.Printf(
			"main: start ftpServer server, Bind=%v, Port=%v",
			ftpServerConfig.Bind, ftpServerConfig.Port,
		)
		ftpServer.Run(ftpServerConfig, config.GetFtpCameraConfigs())
	} else {
		log.Printf("main: skip ftpServer server, err=%v", err)
	}
}
