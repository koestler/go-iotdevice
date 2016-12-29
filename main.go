package main

import (
	"github.com/koestler/go-ve-sensor/bmv"
	"github.com/koestler/go-ve-sensor/vedirect"
	"log"
	"net/http"
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

	// start bmv reader
	go func() {
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
	}()

	// start http server
	go func() {
		router := NewHttpRouter()

		port, err := strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			log.Fatal("while parsing the PORT env variable:")
			log.Fatal(err.Error())
			return
		}

		bind := os.Getenv("BIND")
		if len(bind) < 1 {
			bind = "127.0.0.1"
		}

		log.Printf("[go-ve-sensor] listening on port %v", port)
		log.Fatal(http.ListenAndServe(bind+":"+strconv.Itoa(port), router))
	}()

	select {}
}
