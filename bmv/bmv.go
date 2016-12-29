package bmv

import (
	//"github.com/koestler/go-ve-sensor/bmv"
	"github.com/koestler/go-ve-sensor/vedirect"
	"log"
)

type bmvRegister struct {
	Name    string
	Address uint16
	Factor  float64
	Unit    string
}

type bmvIntValue struct {
	Name  string
	Unit  string
	Value int64
}

var BmvRegisterList = []bmvRegister{
	bmvRegister{
		Name:    "BatteryCapacity",
		Address: 0x1000,
		Factor:  1,
		Unit:    "Ah",
	},
	/*
		bmvRegister{
			Name:    "MainVoltage",
			Address: 0xED8D,
			Factor:  0.01,
			Unit:    "V",
		},
		bmvRegister{
			Name:    "MainCurrent",
			Address: 0xED8F,
			Factor:  0.1,
			Unit:    "A",
		},
	*/
}

func (reg bmvRegister) RecvInt(vd *vedirect.Vedirect) (value bmvIntValue, err error) {
	log.Printf("bmv.BmvGetResgiter begin\n")

	_, err = vd.VeCommandGet(reg.Address)

	if err != nil {
		log.Printf("bmv.BmvGetResgite failed: %v", err)
		return
	}

	value = bmvIntValue{
		Name:  reg.Name,
		Unit:  reg.Unit,
		Value: 42,
	}

	log.Printf("bmv.BmvGetResgiter end, value=%v\n", value)
	return
}
