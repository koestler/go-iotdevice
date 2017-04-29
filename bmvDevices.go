package main

import (
	"github.com/koestler/go-ve-sensor/dataflow"
	"github.com/koestler/go-ve-sensor/config"
	"github.com/koestler/go-ve-sensor/bmv"
)

func BmvDevicesSetupAndRun() (sources []dataflow.Drainable) {

	configs := config.GetBmvConfigs()

	sources = make([]dataflow.Drainable, len(configs))

	// get devices from database and create them
	for i, bmvConfig := range configs {
		// register device in deviceDb
		device := dataflow.DeviceCreate(bmvConfig.Name, bmvConfig.Model);

		// get relevant registers
		registers := bmv.BmvRegisterFactory(bmvConfig.Model);

		// setup the datasource
		sources[i] = dataflow.SourceCreateBmvStartDummy(device, registers)
	}

	return
}
