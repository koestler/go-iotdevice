package dataflow

import "sort"

//go:generate mockgen -source register.go -destination mock/register_mock.go

type Register interface {
	Category() string
	Name() string
	Description() string
	RegisterType() RegisterType
	Enum() map[int]string
	Unit() string
	Sort() int
	Controllable() bool
}

type RegisterStruct struct {
	category     string
	name         string
	description  string
	registerType RegisterType
	enum         map[int]string
	unit         string
	sort         int
	controllable bool
}

func NewRegisterStruct(
	category, name, description string,
	registerType RegisterType,
	enum map[int]string,
	unit string,
	sort int,
	controllable bool,
) RegisterStruct {
	return RegisterStruct{
		category:     category,
		name:         name,
		description:  description,
		registerType: registerType,
		enum:         enum,
		unit:         unit,
		sort:         sort,
		controllable: controllable,
	}
}

func NewRegisterStructByInterface(reg Register) RegisterStruct {
	return RegisterStruct{
		category:     reg.Category(),
		name:         reg.Name(),
		description:  reg.Description(),
		registerType: reg.RegisterType(),
		enum:         reg.Enum(),
		unit:         reg.Unit(),
		sort:         reg.Sort(),
		controllable: reg.Controllable(),
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

func (r RegisterStruct) RegisterType() RegisterType {
	return r.registerType
}

func (r RegisterStruct) Enum() map[int]string {
	return r.enum
}

func (r RegisterStruct) Unit() string {
	return r.unit
}

func (r RegisterStruct) Sort() int {
	return r.sort
}

func (r RegisterStruct) Controllable() bool {
	return r.controllable
}

func FilterRegisters[R Register](input []R, registerFilter RegisterFilterConf) (output []R) {
	output = make([]R, 0, len(input))
	f := RegisterFilter(registerFilter)

	for _, r := range input {
		if f(r) {
			output = append(output, r)
		}
	}
	return
}

func SortRegisters(input []Register) []Register {
	sort.SliceStable(input, func(i, j int) bool { return input[i].Sort() < input[j].Sort() })
	return input
}
