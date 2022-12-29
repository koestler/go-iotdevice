package victronDevice

import "github.com/koestler/go-iotdevice/dataflow"

type VictronRegisters []VictronRegister
type VictronRegister interface {
	dataflow.Register
	Address() uint16
	Static() bool
}

type VictronRegisterStruct struct {
	dataflow.RegisterStruct
	address uint16
	static  bool
}

type NumberRegisterStruct struct {
	VictronRegisterStruct
	signed bool
	factor int
	offset float64
}

type TextRegisterStruct struct {
	VictronRegisterStruct
}

type EnumRegisterStruct struct {
	VictronRegisterStruct
	enum map[int]string
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
) TextRegisterStruct {
	return TextRegisterStruct{
		VictronRegisterStruct{
			dataflow.CreateRegisterStruct(
				category, name, description,
				dataflow.TextRegister, "",
				sort,
			),
			address,
			static,
		},
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
) NumberRegisterStruct {
	return NumberRegisterStruct{
		VictronRegisterStruct{
			dataflow.CreateRegisterStruct(
				category, name, description,
				dataflow.NumberRegister,
				unit,
				sort,
			),
			address,
			static,
		},
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
) EnumRegisterStruct {
	return EnumRegisterStruct{
		VictronRegisterStruct{
			dataflow.CreateRegisterStruct(
				category, name, description,
				dataflow.EnumRegister,
				"",
				sort,
			),
			address,
			static,
		},
		enum,
	}
}

func (r VictronRegisterStruct) Address() uint16 {
	return r.address
}

func (r VictronRegisterStruct) Static() bool {
	return r.static
}

func (r NumberRegisterStruct) Factor() int {
	return r.factor
}

func (r NumberRegisterStruct) Offset() float64 {
	return r.offset
}

func (r NumberRegisterStruct) Signed() bool {
	return r.signed
}

func (r EnumRegisterStruct) Enum() map[int]string {
	return r.enum
}
