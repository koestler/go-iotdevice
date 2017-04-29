package main

import (
	"github.com/koestler/go-ve-sensor/config"
	"log"
	"github.com/koestler/go-ve-sensor/dataflow"
	"github.com/koestler/go-ve-sensor/bmv"
)

func main() {
	log.Print("main: start go-ve-sensor...")

	// get devices from database and create them
	for _, bmvConfig := range config.GetBmvConfigs() {
		// register device in deviceDb
		device := dataflow.DeviceCreate(bmvConfig.Name, bmvConfig.Model);
		registers := bmv.BmvRegisterFactory(bmvConfig.Model);

		// setup the datasource
		source := dataflow.SourceBmvStartDummy(device, registers)

		// store raw values
		rawStorage := dataflow.StorageCreate()
		rawStorage.Receive(source)

		// store rounded values
		roundedStorage := dataflow.StorageCreate()
		rounded := dataflow.Rounder(rawStorage.Subscribe())
		roundedStorage.Receive(rounded)

		// sink everything
		dataflow.SinkLog("raw    ", rawStorage.Subscribe())
		dataflow.SinkLog("rounded", roundedStorage.Subscribe())
	}

	log.Print("main: start completed")
	select {}

}
