package dataflow_test

import "github.com/koestler/go-iotdevice/v3/internal/dataflow"

// MockRegisterFilterConf implements dataflow.RegisterFilterConf by returning its fields.
type MockRegisterFilterConf struct {
	IncludeRegistersV  []string
	SkipRegistersV     []string
	IncludeCategoriesV []string
	SkipCategoriesV    []string
	DefaultIncludeV    bool
}

func (m *MockRegisterFilterConf) IncludeRegisters() []string  { return m.IncludeRegistersV }
func (m *MockRegisterFilterConf) SkipRegisters() []string     { return m.SkipRegistersV }
func (m *MockRegisterFilterConf) IncludeCategories() []string { return m.IncludeCategoriesV }
func (m *MockRegisterFilterConf) SkipCategories() []string    { return m.SkipCategoriesV }
func (m *MockRegisterFilterConf) DefaultInclude() bool        { return m.DefaultIncludeV }

// MockRegister implements dataflow.Register by returning its fields.
type MockRegister struct {
	CategoryV     string
	NameV         string
	DescriptionV  string
	RegisterTypeV dataflow.RegisterType
	EnumV         map[int]string
	UnitV         string
	SortV         int
	WritableV     bool
}

func (m *MockRegister) Category() string                    { return m.CategoryV }
func (m *MockRegister) Name() string                        { return m.NameV }
func (m *MockRegister) Description() string                 { return m.DescriptionV }
func (m *MockRegister) RegisterType() dataflow.RegisterType { return m.RegisterTypeV }
func (m *MockRegister) Enum() map[int]string                { return m.EnumV }
func (m *MockRegister) Unit() string                        { return m.UnitV }
func (m *MockRegister) Sort() int                           { return m.SortV }
func (m *MockRegister) Writable() bool                      { return m.WritableV }
