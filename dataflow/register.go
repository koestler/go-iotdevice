package dataflow

import (
	"cmp"
	"slices"
)

//go:generate mockgen -source register.go -destination mock/register_mock.go

type Register interface {
	Category() string
	Name() string
	Description() string
	RegisterType() RegisterType
	Enum() map[int]string
	Unit() string
	Sort() int
	Writable() bool
}

type RegisterStruct struct {
	category     string
	name         string
	description  string
	registerType RegisterType
	enum         map[int]string
	unit         string
	sort         int
	writable     bool
}

func NewRegisterStruct(
	category, name, description string,
	registerType RegisterType,
	enum map[int]string,
	unit string,
	sort int,
	writable bool,
) RegisterStruct {
	return RegisterStruct{
		category:     category,
		name:         name,
		description:  description,
		registerType: registerType,
		enum:         enum,
		unit:         unit,
		sort:         sort,
		writable:     writable,
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
		writable:     reg.Writable(),
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

func (r RegisterStruct) Writable() bool {
	return r.writable
}

func FilterRegisters[R Register](input []R, filterConf RegisterFilterConf) (output []R) {
	output = make([]R, 0, len(input))
	f := RegisterFilter(filterConf)

	for _, r := range input {
		if f(r) {
			output = append(output, r)
		}
	}
	return
}

func SortRegisterStructs(input []RegisterStruct) {
	slices.SortFunc(input, func(a, b RegisterStruct) int { return cmp.Compare(a.Sort(), b.Sort()) })
}

func (r RegisterStruct) Equals(b RegisterStruct) bool {
	return r.category == b.category &&
		r.name == b.name &&
		r.description == b.description &&
		r.registerType == b.registerType &&
		r.unit == b.unit &&
		r.sort == b.sort &&
		r.writable == b.writable &&
		mapEquals(r.enum, b.enum)
}

func mapEquals(a, b map[int]string) bool {
	if a == nil && b == nil {
		return true
	}

	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
