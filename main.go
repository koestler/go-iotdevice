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
		log.Printf("bmvConfig=%v", bmvConfig)
		log.Printf("device=%v", device)
		log.Printf("registers=%v", registers)

		s := dataflow.SourceBmvStartDummy(device, registers)
		dataflow.SinkLog(s)

		/*
		// register values in valueDb
		for registerName, register := range registers {
			dataflow.ValueCreate(registerName, register.Unit)
		}
		*/

		//dataflow.DevicePrintToLog()
		//dataflow.ValuePrintToLog()

		/*
		go func(deviceId dataflow.DeviceId) {
			for _ = range time.Tick(time.Second) {

			}
		}(deviceId)
		*/

	}

	log.Print("main: start completed")
	select {}

}
