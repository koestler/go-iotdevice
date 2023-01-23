package dataflow

import "fmt"

type Value interface {
	DeviceName() string
	Register() Register
	String() string
	GenericValue() interface{}
	Equals(comp Value) bool
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
	return fmt.Sprintf("%s=%f%s", v.Register().Name(), v.value, v.Register().Unit())
}

func (v NumericRegisterValue) Value() float64 {
	return v.value
}

func (v NumericRegisterValue) GenericValue() interface{} {
	return v.value
}

func (v NumericRegisterValue) Equals(comp Value) bool {
	b, ok := comp.(NumericRegisterValue)
	if !ok {
		return false
	}
	return b.Register().Name() == b.Register().Name() && v.value == b.value
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

type TextRegisterValue struct {
	RegisterValue
	value string
}

func (v TextRegisterValue) String() string {
	return fmt.Sprintf("%s=%s", v.Register().Name(), v.value)
}

func (v TextRegisterValue) Value() string {
	return v.value
}

func (v TextRegisterValue) GenericValue() interface{} {
	return v.value
}

func (v TextRegisterValue) Equals(comp Value) bool {
	b, ok := comp.(TextRegisterValue)
	if !ok {
		return false
	}
	return b.Register().Name() == b.Register().Name() && v.value == b.value
}

func NewTextRegisterValue(deviceName string, register Register, value string) TextRegisterValue {
	return TextRegisterValue{
		RegisterValue: RegisterValue{
			deviceName: deviceName,
			register:   register,
		},
		value: value,
	}
}

type EnumRegisterValue struct {
	RegisterValue
	value int
}

func (v EnumRegisterValue) String() string {
	return fmt.Sprintf("%s=%f%s", v.Register().Name(), v.value, v.Register().Unit())
}

func (v EnumRegisterValue) Value() string {
	if v, ok := v.register.Enum()[v.value]; ok {
		return v
	}
	return ""
}

func (v EnumRegisterValue) GenericValue() interface{} {
	return v.value
}

func (v EnumRegisterValue) Equals(comp Value) bool {
	b, ok := comp.(EnumRegisterValue)
	if !ok {
		return false
	}
	return b.Register().Name() == b.Register().Name() && v.value == b.value
}

func NewEnumRegisterValue(deviceName string, register Register, value int) EnumRegisterValue {
	return EnumRegisterValue{
		RegisterValue: RegisterValue{
			deviceName: deviceName,
			register:   register,
		},
		value: value,
	}
}
