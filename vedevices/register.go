package vedevices

import (
	"github.com/koestler/go-iotdevice/vedirect"
	"log"
)

type NumericValues map[string]NumericValue

type NumericValue struct {
	Value float64
	Unit  string
}

type Registers []Register

type RegisterType int

const (
	StringRegister RegisterType = iota
	SignedNumberRegister
	UnsignedNumberRegister
)

type Register struct {
	Category    string
	Name        string
	Description string
	Address     uint16
	Type        RegisterType
	Factor      float64
	Unit        string
}

func (reg Register) RecvNumeric(vd *vedirect.Vedirect) (result NumericValue, err error) {
	var value float64

	switch reg.Type {
	case SignedNumberRegister:
		var intValue int64
		intValue, err = vd.VeCommandGetInt(reg.Address)
		value = float64(intValue)
	case UnsignedNumberRegister:
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

	numb := 0
	for _, m := range maps {
		numb += len(m)
	}

	output = make(Registers, numb)
	i := 0
	for _, m := range maps {
		for _, v := range m {
			output[i] = v
			i += 1
		}
	}
	return output
}

func registerExcluded(exclude []string, r Register) bool {
	for _, e := range exclude {
		if e == r.Name {
			return true
		}
	}
	return false
}

func filterRegisters(input Registers, exclude []string) (output Registers) {
	output = make(Registers, 0, len(input))
	for _, r := range input {
		if registerExcluded(exclude, r) {
			continue
		}
		output = append(output, r)
	}
	return
}
