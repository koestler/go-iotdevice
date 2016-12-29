package bmv

import (
	"github.com/koestler/go-ve-sensor/vedirect"
	"log"
)

type Register struct {
	Name    string
	Address uint16
	Factor  float64
	Unit    string
	Signed  bool
}

type NumericValue struct {
	Name  string
	Value float64
	Unit  string
}

var RegisterList = []Register{
	Register{
		Name:    "BatteryCapacity",
		Address: 0x1000,
		Factor:  1,
		Unit:    "Ah",
		Signed:  false,
	},
	Register{
		Name:    "MainVoltage",
		Address: 0xED8D,
		Factor:  0.01,
		Unit:    "V",
		Signed:  false,
	},
	Register{
		Name:    "MainCurrent",
		Address: 0xED8F,
		Factor:  0.1,
		Unit:    "A",
		Signed:  true,
	},
}

func (reg Register) RecvNumeric(vd *vedirect.Vedirect) (result NumericValue, err error) {
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

	result = NumericValue{
		Name:  reg.Name,
		Value: value * reg.Factor,
		Unit:  reg.Unit,
	}

	log.Printf("bmv.BmvGetResgiter end, result=%v\n", result)
	return
}
