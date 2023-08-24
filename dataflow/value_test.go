package dataflow

import "testing"

func TestNewNumericRegisterValue(t *testing.T) {
	register := NewRegisterStruct(
		"category",
		"register-name",
		"register-description",
		NumberRegister,
		map[int]string{},
		"unit",
		42,
		false,
	)

	nrv := NewNumericRegisterValue(
		"device-name",
		register,
		3.14,
	)

	if expected, got := "device-name", nrv.DeviceName(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}

	{
		gotReg := nrv.Register()
		if expected, got := "category", gotReg.Category(); expected != got {
			t.Errorf("expected '%s' but got '%s'", expected, got)
		}
		if expected, got := "category", gotReg.Category(); expected != got {
			t.Errorf("expected '%s' but got '%s'", expected, got)
		}

	}
}
