package gpioDevice

import (
	"errors"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/warthog618/go-gpiocdev"
)

var ErrRegisterNotFound = errors.New("register not found")

type GpioRegister struct {
	dataflow.RegisterStruct
	pin    string
	offset int
}

func (r GpioRegister) String() string {
	return fmt.Sprintf("name=%s, pin=%s, offset=%d", r.RegisterStruct.Name(), r.pin, r.offset)
}

func pinToRegisterMap(chip *gpiocdev.Chip, bindings []Pin, category string, sort int, writable bool) (map[string]GpioRegister, error) {
	regs := make(map[string]GpioRegister, len(bindings))
	for i, b := range bindings {
		r, err := pinToRegister(chip, b, category, sort+i, writable)
		if err != nil {
			return nil, err
		}
		regs[b.Name()] = r
	}
	return regs, nil
}

func pinToRegister(chip *gpiocdev.Chip, b Pin, category string, sort int, writable bool) (r GpioRegister, err error) {
	offset, err := chip.FindLine(b.Pin())
	if err != nil {
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
		pin:    b.Pin(),
		offset: offset,
	}

	return r, nil
}

func isValidValue(value int) bool {
	return value == 0 || value == 1
}

func addToRegisterDb(rdb *dataflow.RegisterDb, registers map[string]GpioRegister) {
	dataflowRegisters := make([]dataflow.RegisterStruct, 0, len(registers))
	for _, r := range registers {
		dataflowRegisters = append(dataflowRegisters, r.RegisterStruct)
	}
	rdb.AddStruct(dataflowRegisters...)
}
