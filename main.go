package main

import (
	"errors"
	"github.com/koestler/go-ve-sensor/vedirect"
	"log"
	"time"
)

type veRegister struct {
	Name    string
	Address uint16
	Factor  float64
	Unit    string
}

type veValue struct {
	Name  string
	Unit  string
	Value int
}

var veRegisterList = []veRegister{
	veRegister{
		Name:    "MainVoltage",
		Address: 0xED8D,
		Factor:  0.01,
		Unit:    "V",
	},
	veRegister{
		Name:    "MainCurrent",
		Address: 0xED8F,
		Factor:  0.1,
		Unit:    "A",
	},
}

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

		vd.RecvSyncHex()

		ans := make([]byte, 7)
		vd.Recv(ans)

	}
}

func veGet(register veRegister) (value veValue, err error) {

	err = errors.New("No implemented yet")

	return
}
