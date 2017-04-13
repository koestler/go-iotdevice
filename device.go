package main

import (
	"github.com/koestler/go-ve-sensor/bmv"
	"github.com/koestler/go-ve-sensor/vedata"
	"github.com/koestler/go-ve-sensor/vedirect"
	"log"
	"time"
)

func BmvStart(config BmvConfig) {

	// create new db device connection
	bmvDeviceId := vedata.CreateDevice(config.DeviceName)

	// create
	vd, err := vedirect.Open(config.Device)
	if err != nil {
		log.Fatalf("main:cannot create vedirect, device=%v", config.Device)
		return
	}
	//defer vd.Close()

	// start bmv reader
	go func() {
		numericValues := make(bmv.NumericValues)

		for _ = range time.Tick(200 * time.Millisecond) {
			if err := vd.VeCommandPing(); err != nil {
				log.Printf("main: VeCommandPing failed: %v", err)
			}

			var registers bmv.Registers

			switch config.Type {
			case "700":
				registers = bmv.RegisterList700
				break
			case "702":
				registers = bmv.RegisterList702
				break
			default:
				log.Fatalf("unknown Bmv.Type: %v", config.Type)
			}

			for regName, reg := range registers {
				if numericValue, err := reg.RecvNumeric(vd); err != nil {
					log.Printf("main: bmv.RecvNumeric failed: %v", err)
				} else {
					numericValues[regName] = numericValue
					log.Printf("%v : %v = %v", config.DeviceName, regName, numericValue)
				}
			}

			bmvDeviceId.Write(numericValues)
		}
	}()

}
