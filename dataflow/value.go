package dataflow

import "fmt"

type Value interface {
	DeviceName() string
	Register() Register
	String() string
}

type ValueMap map[string]Value

type RegisterValue struct {
	deviceName string
	register   Register
}

func (v RegisterValue) DeviceName() string {
	return v.deviceName
}

func (v RegisterValue) Register() Register {
	return v.register
}

type NumericRegisterValue struct {
	RegisterValue
	value float64
}

func (v NumericRegisterValue) String() string {
	return fmt.Sprintf("%f", v.value)
}

func NewNumericRegisterValue(deviceName string, register Register, value float64) NumericRegisterValue {
	return NumericRegisterValue{
		RegisterValue: RegisterValue{
			deviceName: deviceName,
			register:   register,
		},
		value: value,
	}
}

type StringRegisterValue struct {
	RegisterValue
	value string
}

func (v StringRegisterValue) String() string {
	return v.value
}
