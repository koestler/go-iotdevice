package dataflow

import (
	"reflect"
	"testing"
)

func TestNewNumericRegisterValue(t *testing.T) {
	testReg := getTestNumberRegister()

	nrv := NewNumericRegisterValue(
		"device-name",
		testReg,
		3.14,
	)

	if expected, got := "device-name", nrv.DeviceName(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := testReg, nrv.Register(); !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %#v but got %#v", expected, got)
	}
	if expected, got := "test-number-register-name=3.140000test-number-register-unit", nrv.String(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := 3.14, nrv.Value(); expected != got {
		t.Errorf("expected %f but got %f", expected, got)
	}
	if expected, got := 3.14, nrv.GenericValue(); !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %#v but got %#v", expected, got)
	}

	{
		matchingNrv := NewNumericRegisterValue(
			"device-name",
			getTestNumberRegister(),
			3.14,
		)
		if !nrv.Equals(matchingNrv) {
			t.Errorf("expected %#v to be equal to %#v", matchingNrv, nrv)
		}
	}

	{
		differentNrv := NewNumericRegisterValue(
			"device-name",
			getTestNumberRegister(),
			3.15,
		)
		if nrv.Equals(differentNrv) {
			t.Errorf("expected %#v to NOT be equal to %#v", differentNrv, nrv)
		}
	}

	{
		differentNrv := NewNumericRegisterValue(
			"device-name",
			getTestTextRegister(),
			3.14,
		)
		if nrv.Equals(differentNrv) {
			t.Errorf("expected %#v to NOT be equal to %#v", differentNrv, nrv)
		}
	}
}

func TestNewTextRegisterValue(t *testing.T) {
	testReg := getTestNumberRegister()

	nrv := NewTextRegisterValue(
		"device-name",
		testReg,
		"foobar",
	)

	if expected, got := "device-name", nrv.DeviceName(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := testReg, nrv.Register(); !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %#v but got %#v", expected, got)
	}
	if expected, got := "test-number-register-name=foobar", nrv.String(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := "foobar", nrv.Value(); expected != got {
		t.Errorf("expected %s but got %s", expected, got)
	}
	if expected, got := "foobar", nrv.GenericValue(); !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %#v but got %#v", expected, got)
	}
}

func TestNewEnumRegisterValue(t *testing.T) {
	testReg := getTestEnumRegister()

	nrv := NewEnumRegisterValue(
		"device-name",
		testReg,
		1,
	)

	if expected, got := "device-name", nrv.DeviceName(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := testReg, nrv.Register(); !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %#v but got %#v", expected, got)
	}
	if expected, got := "test-enum-register-name=1:B", nrv.String(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := 1, nrv.EnumIdx(); expected != got {
		t.Errorf("expected %d but got %d", expected, got)
	}
	if expected, got := 1, nrv.GenericValue(); !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %#v but got %#v", expected, got)
	}
}

func TestNewNullRegisterValue(t *testing.T) {
	testReg := getTestEnumRegister()

	nrv := NewNullRegisterValue(
		"device-name",
		testReg,
	)

	if expected, got := "device-name", nrv.DeviceName(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if expected, got := testReg, nrv.Register(); !reflect.DeepEqual(expected, got) {
		t.Errorf("expected %#v but got %#v", expected, got)
	}
	if expected, got := "NULL", nrv.String(); expected != got {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
	if got := nrv.GenericValue(); got != nil {
		t.Errorf("expected nil but got %#v", got)
	}
}
