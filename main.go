package main

import (
	"github.com/koestler/go-ve-sensor/bmv"
	"github.com/koestler/go-ve-sensor/http"
	"github.com/koestler/go-ve-sensor/vedata"
	"github.com/koestler/go-ve-sensor/vedirect"
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

	// setup routes
	var routes = http.Routes{
		http.Route{
			"Index",
			"GET",
			"/",
			Index,
		},
	}

	// setup vedata (the database)
	vedata.Run()
	bmvDeviceId := vedata.CreateDevice("test0")

	// start bmv reader
	routes = append(routes,
		http.Route{
			"bmv",
			"GET",
			"/bmv/",
			HttpHandleBmv,
		},
	)

	go func() {
		numericValues := make(bmv.NumericValues)

		for _ = range time.Tick(200 * time.Millisecond) {
			if err := vd.VeCommandPing(); err != nil {
				log.Printf("main: VeCommandPing failed: %v", err)
			}

			for regName, reg := range bmv.RegisterList700 {
				log.Printf("main: bmv.RecvNumeric regName=%v", regName)

				if numericValue, err := reg.RecvNumeric(vd); err != nil {
					log.Printf("main: bmv.RecvNumeric failed: %v", err)
				} else {
					numericValues[regName] = numericValue
					bmvDeviceId.WriteNumericValue(regName, numericValue)
				}
			}

			log.Printf("numericValues=%v", numericValues)
		}
	}()

	// start http server
	go func(routes http.Routes) {
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

		http.Run(bind, port, routes)
	}(routes)

	select {}
}
