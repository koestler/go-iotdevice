package dataflow

type Registers []Register

type RegisterType int

const (
	StringRegister RegisterType = iota
	NumberRegister
	EnumRegister
)

type Register interface {
	Category() string
	Name() string
	Description() string
	Address() uint16
	Static() bool
	Type() RegisterType
	Unit() *string
}

type RegisterStruct struct {
	category    string
	name        string
	description string
	address     uint16
	static      bool
}

type TextRegisterStruct struct {
	RegisterStruct
}

type NumberRegisterStruct struct {
	RegisterStruct
	signed bool
	factor int
	unit   *string
}

type EnumRegisterStruct struct {
	RegisterStruct
	enum map[int]string
}

func CreateTextRegisterStruct(category, name, description string, address uint16, static bool) TextRegisterStruct {
	return TextRegisterStruct{
		RegisterStruct{
			category:    category,
			name:        name,
			description: description,
			address:     address,
			static:      static,
		},
	}
}

func CreateNumberRegisterStruct(
	category, name, description string,
	address uint16,
	static bool,
	signed bool,
	factor int,
	unit string,
) NumberRegisterStruct {
	var u *string = nil
	if len(unit) > 0 {
		u = &unit
	}

	return NumberRegisterStruct{
		RegisterStruct: RegisterStruct{
			category:    category,
			name:        name,
			description: description,
			address:     address,
			static:      static,
		},
		signed: signed,
		factor: factor,
		unit:   u,
	}
}

func CreateEnumRegisterStruct(category, name, description string, address uint16, static bool, enum map[int]string) EnumRegisterStruct {
	return EnumRegisterStruct{
		RegisterStruct: RegisterStruct{
			category:    category,
			name:        name,
			description: description,
			address:     address,
			static:      static,
		},
		enum: enum,
	}
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

func (r RegisterStruct) Static() bool {
	return r.static
}

func (r TextRegisterStruct) Unit() *string {
	return nil
}

func (r EnumRegisterStruct) Unit() *string {
	return nil
}

func (r NumberRegisterStruct) Factor() int {
	return r.factor
}

func (r NumberRegisterStruct) Unit() *string {
	return r.unit
}

func (r NumberRegisterStruct) Signed() bool {
	return r.signed
}

func (r EnumRegisterStruct) Enum() map[int]string {
	return r.enum
}

func (r TextRegisterStruct) Type() RegisterType {
	return StringRegister
}

func (r NumberRegisterStruct) Type() RegisterType {
	return NumberRegister
}

func (r EnumRegisterStruct) Type() RegisterType {
	return EnumRegister
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
