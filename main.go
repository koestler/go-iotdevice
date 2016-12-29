package main

import (
	"github.com/koestler/go-ve-sensor/bmv"
	"github.com/koestler/go-ve-sensor/vedirect"
	"log"
	"time"
)

func main() {
	vd, err := vedirect.Open("/dev/ttyUSB0")
	if err != nil {
		log.Fatalf("main:cannot create vedirect")
		return
	}
	defer vd.Close()

	// start bmv reader
	numericValues := make([]bmv.NumericValue, len(bmv.RegisterList))
	for {
		if err := vd.VeCommandPing(); err != nil {
			log.Printf("main: VeCommandPing failed: %v", err)
		}

		for i, reg := range bmv.RegisterList {
			if numericValue, err := reg.RecvNumeric(vd); err != nil {
				log.Printf("main: bmv.RecvNumeric failed: %v", err)
			} else {
				numericValues[i] = numericValue
			}
		}

		time.Sleep(200 * time.Millisecond)
	}
}
