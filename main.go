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

	// send restart
	vd.SendVeCommand(vedirect.VeCommandRestart, []byte{})

	// ping every 100ms
	go func(vd *vedirect.Vedirect) {
		for {
			time.Sleep(100 * time.Millisecond)
		}
	}(vd)

	// read for a while...
	for {
		//log.Println("Sleep 500ms")
		//time.Sleep(500 * time.Millisecond)

		/*
			if err := vd.VeCommandPing(); err != nil {
				log.Printf("main: VeCommandPing failed: %v", err)
			}
		*/

		for _, reg := range bmv.BmvRegisterList {
			if _, err := reg.RecvInt(vd); err != nil {
				log.Printf("main: bmv.BmvGetRegister failed: %v", err)
			}
		}

		log.Printf("\n\n\n")

	}
}
