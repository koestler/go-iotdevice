package main

import (
	"github.com/koestler/go-ve-sensor/bmv"
	"github.com/koestler/go-ve-sensor/vedata"
	"github.com/koestler/go-ve-sensor/vedirect"
	"github.com/koestler/go-ve-sensor/vehttp"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	vd, err := vedirect.Open("/dev/ttyUSB0")
	if err != nil {
		log.Fatalf("main:cannot create vedirect")
		return
	}
	defer vd.Close()

	// setup vedata (the database)
	vedata.Run()
	bmvDeviceId := vedata.CreateDevice("test0")

	// start bmv reader
	go func() {
		numericValues := make(bmv.NumericValues)

		for _ = range time.Tick(500 * time.Millisecond) {
			if err := vd.VeCommandPing(); err != nil {
				log.Printf("main: VeCommandPing failed: %v", err)
			}

			for regName, reg := range bmv.RegisterList700 {
				if numericValue, err := reg.RecvNumeric(vd); err != nil {
					log.Printf("main: bmv.RecvNumeric failed: %v", err)
				} else {
					numericValues[regName] = numericValue
				}
			}

			bmvDeviceId.Write(numericValues)
		}
	}()

	// start http server
	go func() {
		bind := os.Getenv("BIND")
		if len(bind) < 1 {
			bind = "127.0.0.1"
		}

		port, err := strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			log.Fatal("while parsing the PORT env variable:")
			log.Fatal(err.Error())
			return
		}

		vehttp.Run(bind, port, HttpRoutes)
	}()

	select {}
}
