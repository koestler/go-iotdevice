package victronDevice

import "github.com/koestler/go-iotdevice/dataflow"

type VictronRegister struct {
	dataflow.RegisterStruct
	address uint16
	static  bool

	// only relevant for number registers
	signed bool
	factor int
	offset float64
}

func MergeRegisters(maps ...[]VictronRegister) (output []VictronRegister) {
	if len(maps) == 0 {
		return
	}
	output = maps[0]
	for i := 1; i < len(maps); i++ {
		output = append(output, maps[i]...)
	}

	return
}

func FilterRegisters(input []VictronRegister, excludeFields []string, excludeCategories []string) (output []VictronRegister) {
	output = make([]VictronRegister, 0, len(input))
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

func FilterRegistersByName(input []VictronRegister, names ...string) (output []VictronRegister) {
	output = make([]VictronRegister, 0, len(input))
	for _, r := range input {
		if dataflow.RegisterNameExcluded(names, r) {
			continue
		}
		output = append(output, r)
	}
	return
}

func NewTextRegisterStruct(
	category, name, description string,
	address uint16,
	static bool,
	sort int,
) VictronRegister {
	return VictronRegister{
		dataflow.NewRegisterStruct(
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

func NewNumberRegisterStruct(
	category, name, description string,
	address uint16,
	static bool,
	signed bool,
	factor int,
	offset float64,
	unit string,
	sort int,
) VictronRegister {
	return VictronRegister{
		dataflow.NewRegisterStruct(
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

func NewEnumRegisterStruct(
	category, name, description string,
	address uint16,
	static bool,
	enum map[int]string,
	sort int,
) VictronRegister {
	return VictronRegister{
		dataflow.NewRegisterStruct(
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
