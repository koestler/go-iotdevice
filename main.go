package main

import (
	"log"
	"github.com/koestler/go-ve-sensor/dataflow"
)

func main() {
	log.Print("main: start go-ve-sensor...")

	// setup dataflow pipeline
	// 1. sources:
	sources := BmvDevicesSetupAndRun()

	// 2. storage for raw values
	rawStorage := dataflow.ValueStorageCreate()

	// 3. rounder
	rounder := dataflow.RounderCreate()

	// 4. storage for rounded values
	roundedStorage := dataflow.ValueStorageCreate()

	// chain those
	for _, source:= range sources {
		source.Append(rawStorage)
	}
	rawStorage.Append(rounder)
	rounder.Append(roundedStorage)

	// setup some test sinks
	devices := dataflow.DevicesGet()
	f0 := dataflow.SubscriptionFilter{
		Devices: map[*dataflow.Device]bool{devices[0]: true},
		ValueNames: map[string]bool{"Power": true},
	};
	dataflow.SinkLog("raw    ", rawStorage.Subscribe(f0))
	dataflow.SinkLog("rounded", roundedStorage.Drain())

	log.Print("main: start completed")
	select {}

}
