package dataflow

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

type Register interface {
	Category() string
	Name() string
	Description() string
	Address() uint16
	Type() RegisterType
}

type RegisterStruct struct {
	category    string
	name        string
	description string
	address     uint16
}

func (r RegisterStruct) Category() string {
	return r.category
}

func (r RegisterStruct) Name() string {
	return r.name
}

func (r RegisterStruct) Description() string {
	return r.description
}

func (r RegisterStruct) Address() uint16 {
	return r.address
}

type StringRegisterStruct struct {
	RegisterStruct
}

func (r StringRegisterStruct) Type() RegisterType {
	return StringRegister
}

type NumberRegisterStruct struct {
	RegisterStruct
	factor float64
	unit   string
}

type SignedNumberRegisterStruct struct {
	NumberRegisterStruct
}

func (r SignedNumberRegisterStruct) Type() RegisterType {
	return SignedNumberRegister
}

type UnsignedNumberRegisterStruct struct {
	NumberRegisterStruct
}

func (r UnsignedNumberRegisterStruct) Type() RegisterType {
	return UnsignedNumberRegister
}

func MergeRegisters(maps ...Registers) (output Registers) {
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

func FilterRegisters(input Registers, exclude []string) (output Registers) {
	output = make(Registers, 0, len(input))
	for _, r := range input {
		if registerExcluded(exclude, r) {
			continue
		}
		output = append(output, r)
	}
	return
}

func registerExcluded(exclude []string, r Register) bool {
	for _, e := range exclude {
		if e == r.Name() {
			return true
		}
	}
	return false
}
