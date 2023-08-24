package dataflow

import (
	"reflect"
	"testing"
)

func getTestTextRegister() RegisterStruct {
	return NewRegisterStruct(
		"test-text-register-category",
		"test-text-register-name",
		"test-text-register-description",
		TextRegister,
		map[int]string{},
		"test-text-register-unit",
		40,
		false,
	)
}

func getTestNumberRegister() RegisterStruct {
	return NewRegisterStruct(
		"test-number-register-category",
		"test-number-register-name",
		"test-number-register-description",
		NumberRegister,
		map[int]string{},
		"test-number-register-unit",
		41,
		false,
	)
}

func getTestEnumRegister() RegisterStruct {
	return NewRegisterStruct(
		"test-enum-register-category",
		"test-enum-register-name",
		"test-enum-register-description",
		EnumRegister,
		map[int]string{0: "A", 1: "B"},
		"test-enum-register-unit",
		42,
		false,
	)
}

func TestTextRegisterCreatorAndGetters(t *testing.T) {
	register := getTestTextRegister()

	if expected, got := "test-text-register-category", register.Category(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := "test-text-register-name", register.Name(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := "test-text-register-description", register.Description(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := TextRegister, register.RegisterType(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := map[int]string{}, register.Enum(); !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %#v but got %#v", expected, got)
	}
	if expected, got := "test-text-register-unit", register.Unit(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := 40, register.Sort(); expected != got {
		t.Errorf("expected %d but got %d", expected, got)
	}
	if got := register.Controllable(); got {
		t.Errorf("expected controllable to be false")
	}
}

func TestNumberRegisterCreatorAndGetters(t *testing.T) {
	register := getTestNumberRegister()

	if expected, got := "test-number-register-category", register.Category(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := "test-number-register-name", register.Name(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := "test-number-register-description", register.Description(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := NumberRegister, register.RegisterType(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := map[int]string{}, register.Enum(); !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %#v but got %#v", expected, got)
	}
	if expected, got := "test-number-register-unit", register.Unit(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := 41, register.Sort(); expected != got {
		t.Errorf("expected %d but got %d", expected, got)
	}
	if got := register.Controllable(); got {
		t.Errorf("expected controllable to be false")
	}
}

func TestEnumRegisterCreatorAndGetters(t *testing.T) {
	register := getTestEnumRegister()

	if expected, got := "test-enum-register-category", register.Category(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := "test-enum-register-name", register.Name(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := "test-enum-register-description", register.Description(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := EnumRegister, register.RegisterType(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := map[int]string{0: "A", 1: "B"}, register.Enum(); !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %#v but got %#v", expected, got)
	}
	if expected, got := "test-enum-register-unit", register.Unit(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := 42, register.Sort(); expected != got {
		t.Errorf("expected %d but got %d", expected, got)
	}
	if got := register.Controllable(); got {
		t.Errorf("expected controllable to be false")
	}
}

func TestFilterRegisters(t *testing.T) {

}
