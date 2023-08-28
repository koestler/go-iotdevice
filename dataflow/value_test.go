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

	if expect, got := "device-name", nrv.DeviceName(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := testReg, nrv.Register(); !reflect.DeepEqual(expect, got) {
		t.Errorf("expect %#v but got %#v", expect, got)
	}
	if expect, got := "test-number-register-name=3.140000test-number-register-unit", nrv.String(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := 3.14, nrv.Value(); expect != got {
		t.Errorf("expect %f but got %f", expect, got)
	}
	if expect, got := 3.14, nrv.GenericValue(); !reflect.DeepEqual(expect, got) {
		t.Errorf("expect %#v but got %#v", expect, got)
	}

	{
		matchingNrv := NewNumericRegisterValue(
			"device-name",
			getTestNumberRegister(),
			3.14,
		)
		if !nrv.Equals(matchingNrv) {
			t.Errorf("expect %#v to be equal to %#v", matchingNrv, nrv)
		}
	}

	{
		differentNrv := NewNumericRegisterValue(
			"device-name",
			getTestNumberRegister(),
			3.15,
		)
		if nrv.Equals(differentNrv) {
			t.Errorf("expect %#v to NOT be equal to %#v", differentNrv, nrv)
		}
	}

	{
		differentNrv := NewNumericRegisterValue(
			"device-name",
			getTestTextRegister(),
			3.14,
		)
		if nrv.Equals(differentNrv) {
			t.Errorf("expect %#v to NOT be equal to %#v", differentNrv, nrv)
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

	if expect, got := "device-name", nrv.DeviceName(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := testReg, nrv.Register(); !reflect.DeepEqual(expect, got) {
		t.Errorf("expect %#v but got %#v", expect, got)
	}
	if expect, got := "test-number-register-name=foobar", nrv.String(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := "foobar", nrv.Value(); expect != got {
		t.Errorf("expect %s but got %s", expect, got)
	}
	if expect, got := "foobar", nrv.GenericValue(); !reflect.DeepEqual(expect, got) {
		t.Errorf("expect %#v but got %#v", expect, got)
	}
}

func TestNewEnumRegisterValue(t *testing.T) {
	testReg := getTestEnumRegister()

	nrv := NewEnumRegisterValue(
		"device-name",
		testReg,
		1,
	)

	if expect, got := "device-name", nrv.DeviceName(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := testReg, nrv.Register(); !reflect.DeepEqual(expect, got) {
		t.Errorf("expect %#v but got %#v", expect, got)
	}
	if expect, got := "test-enum-register-name=1:B", nrv.String(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := 1, nrv.EnumIdx(); expect != got {
		t.Errorf("expect %d but got %d", expect, got)
	}
	if expect, got := 1, nrv.GenericValue(); !reflect.DeepEqual(expect, got) {
		t.Errorf("expect %#v but got %#v", expect, got)
	}
}

func TestNewNullRegisterValue(t *testing.T) {
	testReg := getTestEnumRegister()

	nrv := NewNullRegisterValue(
		"device-name",
		testReg,
	)

	if expect, got := "device-name", nrv.DeviceName(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if expect, got := testReg, nrv.Register(); !reflect.DeepEqual(expect, got) {
		t.Errorf("expect %#v but got %#v", expect, got)
	}
	if expect, got := "NULL", nrv.String(); expect != got {
		t.Errorf("expect '%s' but got '%s'", expect, got)
	}
	if got := nrv.GenericValue(); got != nil {
		t.Errorf("expect nil but got %#v", got)
	}
}
