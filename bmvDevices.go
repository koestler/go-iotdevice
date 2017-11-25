package main

import (
	"github.com/koestler/go-ve-sensor/dataflow"
	"github.com/koestler/go-ve-sensor/config"
	"github.com/koestler/go-ve-sensor/vedevices"
	"log"
)

func BmvDevicesSetupAndRun() (sources []dataflow.Drainable) {

	configs := config.GetVedeviceConfigs()

	sources = make([]dataflow.Drainable, 0, len(configs))

	// get devices from database and create them
	for _, c := range configs {
		log.Printf(
			"bmvDevices: setup name=%v model=%v device=%v",
			c.Name, c.Model, c.Device,
		)

		// register device in deviceDb
		device := dataflow.DeviceCreate(c.Name, c.Model);

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

	return
}
