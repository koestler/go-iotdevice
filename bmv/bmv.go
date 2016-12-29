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
	Signed  bool
}

type bmvNumericValue struct {
	Name  string
	Value float64
	Unit  string
}

var BmvRegisterList = []bmvRegister{
	bmvRegister{
		Name:    "BatteryCapacity",
		Address: 0x1000,
		Factor:  1,
		Unit:    "Ah",
		Signed:  false,
	},
	bmvRegister{
		Name:    "MainVoltage",
		Address: 0xED8D,
		Factor:  0.01,
		Unit:    "V",
		Signed:  false,
	},
	/*
		bmvRegister{
			Name:    "MainCurrent",
			Address: 0xED8F,
			Factor:  0.1,
			Unit:    "A",
		},
	*/
}

func (reg bmvRegister) RecvNumeric(vd *vedirect.Vedirect) (result bmvNumericValue, err error) {
	log.Printf("bmv.BmvGetResgiter begin\n")

	var value float64

	if reg.Signed {
		var intValue int64
		intValue, err = vd.VeCommandGetInt(reg.Address)
		value = float64(intValue)
	} else {
		var intValue uint64
		intValue, err = vd.VeCommandGetUint(reg.Address)
		value = float64(intValue)
	}

	if err != nil {
		log.Printf("bmv.BmvGetResgite failed: %v", err)
		return
	}

	result = bmvNumericValue{
		Name:  reg.Name,
		Value: value * reg.Factor,
		Unit:  reg.Unit,
	}

	log.Printf("bmv.BmvGetResgiter end, result=%v\n", result)
	return
}
