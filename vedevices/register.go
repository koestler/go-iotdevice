package vedevices

import (
	"github.com/koestler/go-victron-to-mqtt/vedirect"
	"log"
)

type NumericValues map[string]NumericValue

type NumericValue struct {
	Value float64
	Unit  string
}

type Registers map[string]Register

type Register struct {
	Address       uint16
	Factor        float64
	Unit          string
	Signed        bool
	RoundDecimals int
}

func (reg Register) RecvNumeric(vd *vedirect.Vedirect) (result NumericValue, err error) {
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
		log.Printf("vedevices.RecvNumeric failed: %v", err)
		return
	}

	result = NumericValue{
		Value: value * reg.Factor,
		Unit:  reg.Unit,
	}

	return
}

func mergeRegisters(maps ...Registers) (output Registers) {
	size := len(maps)
	if size == 0 {
		return output
	}
	if size == 1 {
		return maps[0]
	}
	output = make(Registers)
	for _, m := range maps {
		for k, v := range m {
			output[k] = v
		}
	}
	return output
}
