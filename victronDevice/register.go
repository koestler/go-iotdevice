package victronDevice

import "github.com/koestler/go-iotdevice/dataflow"

type VictronRegisters []VictronRegister
type VictronRegister interface {
	dataflow.Register
	Address() uint16
	Static() bool
	Signed() bool
	Factor() int
	Offset() float64
}

type VictronRegisterStruct struct {
	dataflow.RegisterStruct
	address uint16
	static  bool

	// only relevant for number registers
	signed bool
	factor int
	offset float64
}

func MergeRegisters(maps ...VictronRegisters) (output VictronRegisters) {
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

	output = make([]VictronRegister, numb)
	i := 0
	for _, m := range maps {
		for _, v := range m {
			output[i] = v
			i += 1
		}
	}
	return output
}

func FilterRegisters(input VictronRegisters, excludeFields []string, excludeCategories []string) (output VictronRegisters) {
	output = make(VictronRegisters, 0, len(input))
	for _, r := range input {
		if dataflow.RegisterNameExcluded(excludeFields, r) {
			continue
		}
		if dataflow.RegisterCategoryExcluded(excludeCategories, r) {
			continue
		}
		output = append(output, r)
	}
	return
}

func FilterRegistersByName(input VictronRegisters, names ...string) (output VictronRegisters) {
	output = make(VictronRegisters, 0, len(input))
	for _, r := range input {
		if dataflow.RegisterNameExcluded(names, r) {
			continue
		}
		output = append(output, r)
	}
	return
}

func CreateTextRegisterStruct(
	category, name, description string,
	address uint16,
	static bool,
	sort int,
) VictronRegisterStruct {
	return VictronRegisterStruct{
		dataflow.CreateRegisterStruct(
			category, name, description,
			dataflow.TextRegister,
			nil,
			"",
			sort,
			false,
		),
		address,
		static,
		false, // unused
		1,     // unused
		0,     // unused
	}
}

func CreateNumberRegisterStruct(
	category, name, description string,
	address uint16,
	static bool,
	signed bool,
	factor int,
	offset float64,
	unit string,
	sort int,
) VictronRegisterStruct {
	return VictronRegisterStruct{
		dataflow.CreateRegisterStruct(
			category, name, description,
			dataflow.NumberRegister,
			nil,
			unit,
			sort,
			false,
		),
		address,
		static,
		signed,
		factor,
		offset,
	}
}

func CreateEnumRegisterStruct(
	category, name, description string,
	address uint16,
	static bool,
	enum map[int]string,
	sort int,
) VictronRegisterStruct {
	return VictronRegisterStruct{
		dataflow.CreateRegisterStruct(
			category, name, description,
			dataflow.EnumRegister,
			enum,
			"",
			sort,
			false,
		),
		address,
		static,
		false, // unused
		1,     // unused
		0,     // unused
	}
}

func (r VictronRegisterStruct) Address() uint16 {
	return r.address
}

func (r VictronRegisterStruct) Static() bool {
	return r.static
}

func (r VictronRegisterStruct) Factor() int {
	return r.factor
}

func (r VictronRegisterStruct) Offset() float64 {
	return r.offset
}

func (r VictronRegisterStruct) Signed() bool {
	return r.signed
}
