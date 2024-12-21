package gpioDevice

import (
	"errors"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

var ErrRegisterNotFound = errors.New("register not found")

type GpioRegister struct {
	dataflow.RegisterStruct
	pin gpio.PinIO
}

func pinToRegisterMap(bindings []Pin, category string, sort int, writable bool) (map[string]GpioRegister, error) {
	regs := make(map[string]GpioRegister, len(bindings))
	for i, b := range bindings {
		r, err := pinToRegister(b, category, sort+i, writable)
		if err != nil {
			return nil, err
		}
		regs[b.Name()] = r
	}
	return regs, nil
}

func pinToRegister(b Pin, category string, sort int, writable bool) (r GpioRegister, err error) {
	pin := gpioreg.ByName(b.Pin())
	if pin == nil {
		return r, fmt.Errorf("%w: pinName=%s", ErrRegisterNotFound, b.Pin())
	}

	r = GpioRegister{
		RegisterStruct: dataflow.NewRegisterStruct(
			category, b.Name(), b.Description(),
			dataflow.EnumRegister,
			map[int]string{
				0: b.LowLabel(),
				1: b.HighLabel(),
			},
			"", sort, writable,
		),
		pin: pin,
	}

	return r, nil
}

func addToRegisterDb(rdb *dataflow.RegisterDb, registers map[string]GpioRegister) {
	dataflowRegisters := make([]dataflow.RegisterStruct, 0, len(registers))
	for _, r := range registers {
		dataflowRegisters = append(dataflowRegisters, r.RegisterStruct)
	}
	rdb.AddStruct(dataflowRegisters...)
}
