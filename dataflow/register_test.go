package dataflow_test

import (
	"github.com/koestler/go-iotdevice/dataflow"
	mock_dataflow "github.com/koestler/go-iotdevice/dataflow/mock"
	"go.uber.org/mock/gomock"
	"reflect"
	"testing"
)

func getTestTextRegisterWithName(name string) dataflow.RegisterStruct {
	return dataflow.NewRegisterStruct(
		"test-text-register-category",
		name,
		"test-text-register-description",
		dataflow.TextRegister,
		map[int]string{},
		"test-text-register-unit",
		40,
		false,
	)
}

func getTestTextRegister() dataflow.RegisterStruct {
	return getTestTextRegisterWithName("test-text-register-name")
}

func getTestNumberRegister() dataflow.RegisterStruct {
	return dataflow.NewRegisterStruct(
		"test-number-register-category",
		"test-number-register-name",
		"test-number-register-description",
		dataflow.NumberRegister,
		map[int]string{},
		"test-number-register-unit",
		41,
		false,
	)
}

func getTestEnumRegister() dataflow.RegisterStruct {
	return dataflow.NewRegisterStruct(
		"test-enum-register-category",
		"test-enum-register-name",
		"test-enum-register-description",
		dataflow.EnumRegister,
		map[int]string{0: "A", 1: "B"},
		"test-enum-register-unit",
		42,
		false,
	)
}

func TestTextRegisterCreatorAndGetters(t *testing.T) {
	register := getTestTextRegister()

	if expect, got := "test-text-register-category", register.Category(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := "test-text-register-name", register.Name(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := "test-text-register-description", register.Description(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := dataflow.TextRegister, register.RegisterType(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := map[int]string{}, register.Enum(); !reflect.DeepEqual(expect, got) {
		t.Errorf("expect %#v but got %#v", expect, got)
	}
	if expect, got := "test-text-register-unit", register.Unit(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := 40, register.Sort(); expect != got {
		t.Errorf("expect %d but got %d", expect, got)
	}
	if got := register.Controllable(); got {
		t.Errorf("expect controllable to be false")
	}
}

func TestNumberRegisterCreatorAndGetters(t *testing.T) {
	register := getTestNumberRegister()

	if expect, got := "test-number-register-category", register.Category(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := "test-number-register-name", register.Name(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := "test-number-register-description", register.Description(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := dataflow.NumberRegister, register.RegisterType(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := map[int]string{}, register.Enum(); !reflect.DeepEqual(expect, got) {
		t.Errorf("expect %#v but got %#v", expect, got)
	}
	if expect, got := "test-number-register-unit", register.Unit(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := 41, register.Sort(); expect != got {
		t.Errorf("expect %d but got %d", expect, got)
	}
	if got := register.Controllable(); got {
		t.Errorf("expect controllable to be false")
	}
}

func TestEnumRegisterCreatorAndGetters(t *testing.T) {
	register := getTestEnumRegister()

	if expect, got := "test-enum-register-category", register.Category(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := "test-enum-register-name", register.Name(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := "test-enum-register-description", register.Description(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := dataflow.EnumRegister, register.RegisterType(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := map[int]string{0: "A", 1: "B"}, register.Enum(); !reflect.DeepEqual(expect, got) {
		t.Errorf("expect %#v but got %#v", expect, got)
	}
	if expect, got := "test-enum-register-unit", register.Unit(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := 42, register.Sort(); expect != got {
		t.Errorf("expect %d but got %d", expect, got)
	}
	if got := register.Controllable(); got {
		t.Errorf("expect controllable to be false")
	}
}

func TestFilterRegisters(t *testing.T) {
	stimuliRegisters := []dataflow.Register{
		getTestTextRegisterWithName("a"),
		getTestTextRegisterWithName("b"),
		getTestNumberRegister(),
		getTestEnumRegister(),
	}

	ctrl := gomock.NewController(t)

	t.Run("nothing", func(t *testing.T) {
		fc := mock_dataflow.NewMockRegisterFilterConf(ctrl)
		fc.EXPECT().SkipRegisters().Return([]string{}).AnyTimes()
		fc.EXPECT().IncludeRegisters().Return([]string{}).AnyTimes()
		fc.EXPECT().SkipCategories().Return([]string{}).AnyTimes()
		fc.EXPECT().IncludeCategories().Return([]string{}).AnyTimes()
		fc.EXPECT().DefaultInclude().Return(true).AnyTimes()

		got := dataflow.FilterRegisters(stimuliRegisters, fc)

		expect := []dataflow.Register{
			getTestTextRegisterWithName("a"),
			getTestTextRegisterWithName("b"),
			getTestNumberRegister(),
			getTestEnumRegister(),
		}

		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expect %#v but got %#v", expect, got)
		}
	})

	t.Run("onlyIncludeRegisters", func(t *testing.T) {
		fc := mock_dataflow.NewMockRegisterFilterConf(ctrl)
		fc.EXPECT().SkipRegisters().Return([]string{}).AnyTimes()
		fc.EXPECT().IncludeRegisters().Return([]string{"a", "b"}).AnyTimes()
		fc.EXPECT().SkipCategories().Return([]string{}).AnyTimes()
		fc.EXPECT().IncludeCategories().Return([]string{}).AnyTimes()
		fc.EXPECT().DefaultInclude().Return(false).AnyTimes()

		got := dataflow.FilterRegisters(stimuliRegisters, fc)

		expect := []dataflow.Register{
			getTestTextRegisterWithName("a"),
			getTestTextRegisterWithName("b"),
		}

		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expect %#v but got %#v", expect, got)
		}
	})

	t.Run("onlySkipRegisters", func(t *testing.T) {
		fc := mock_dataflow.NewMockRegisterFilterConf(ctrl)
		fc.EXPECT().SkipRegisters().Return([]string{"a"}).AnyTimes()
		fc.EXPECT().IncludeRegisters().Return([]string{}).AnyTimes()
		fc.EXPECT().SkipCategories().Return([]string{}).AnyTimes()
		fc.EXPECT().IncludeCategories().Return([]string{}).AnyTimes()
		fc.EXPECT().DefaultInclude().Return(true).AnyTimes()

		got := dataflow.FilterRegisters(stimuliRegisters, fc)

		expect := []dataflow.Register{
			getTestTextRegisterWithName("b"),
			getTestNumberRegister(),
			getTestEnumRegister(),
		}

		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expect %#v but got %#v", expect, got)
		}
	})

	t.Run("onlyIncludeCategories", func(t *testing.T) {

		fc := mock_dataflow.NewMockRegisterFilterConf(ctrl)
		fc.EXPECT().SkipRegisters().Return([]string{}).AnyTimes()
		fc.EXPECT().IncludeRegisters().Return([]string{}).AnyTimes()
		fc.EXPECT().SkipCategories().Return([]string{}).AnyTimes()
		fc.EXPECT().IncludeCategories().Return([]string{"test-number-register-category"}).AnyTimes()
		fc.EXPECT().DefaultInclude().Return(false).AnyTimes()

		got := dataflow.FilterRegisters(stimuliRegisters, fc)

		expect := []dataflow.Register{
			getTestNumberRegister(),
		}

		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expect %#v but got %#v", expect, got)
		}
	})

	t.Run("onlySkipCategories", func(t *testing.T) {

		fc := mock_dataflow.NewMockRegisterFilterConf(ctrl)
		fc.EXPECT().SkipRegisters().Return([]string{}).AnyTimes()
		fc.EXPECT().IncludeRegisters().Return([]string{}).AnyTimes()
		fc.EXPECT().SkipCategories().Return([]string{"test-number-register-category"}).AnyTimes()
		fc.EXPECT().IncludeCategories().Return([]string{}).AnyTimes()
		fc.EXPECT().DefaultInclude().Return(true).AnyTimes()

		got := dataflow.FilterRegisters(stimuliRegisters, fc)

		expect := []dataflow.Register{
			getTestTextRegisterWithName("a"),
			getTestTextRegisterWithName("b"),
			getTestEnumRegister(),
		}

		if !reflect.DeepEqual(expect, got) {
			t.Errorf("expect %#v but got %#v", expect, got)
		}
	})
}

func TestSortRegisters(t *testing.T) {
	stimuliRegisters := []dataflow.Register{
		getTestNumberRegister(),
		getTestTextRegisterWithName("a"),
		getTestEnumRegister(),
		getTestTextRegisterWithName("b"),
	}

	got := dataflow.SortRegisters(stimuliRegisters)

	expect := []dataflow.Register{
		getTestTextRegisterWithName("a"),
		getTestTextRegisterWithName("b"),
		getTestNumberRegister(),
		getTestEnumRegister(),
	}

	if !reflect.DeepEqual(expect, got) {
		t.Errorf("expect %#v but got %#v", expect, got)
	}
}
