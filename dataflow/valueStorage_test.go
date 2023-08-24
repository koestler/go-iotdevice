package dataflow

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"testing"
)

func getSimpleTestRegister(name string) RegisterStruct {
	return NewRegisterStruct(
		"",
		name,
		"",
		NumberRegister,
		map[int]string{},
		"",
		40,
		false,
	)
}

func TestGetState(t *testing.T) {
	storage := NewValueStorage()

	storage.Fill(NewNumericRegisterValue(
		"device-0",
		getSimpleTestRegister("register-a"),
		0,
	))

	storage.Fill(NewNumericRegisterValue(
		"device-0",
		getSimpleTestRegister("register-a"),
		1,
	))

	storage.Fill(NewNumericRegisterValue(
		"device-0",
		getSimpleTestRegister("register-b"),
		10,
	))

	storage.Fill(NewNumericRegisterValue(
		"device-1",
		getSimpleTestRegister("register-a"),
		100,
	))

	{
		expected := []string{
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
			"device-1:register-a=100.000000",
		}
		storage.Wait()
		got := getAsStrings(storage.GetState(Filter{}))
		if !equalIgnoreOrder(expected, got) {
			t.Errorf("expected %#v but got %#v", expected, got)
		}
	}

	storage.Fill(NewNumericRegisterValue(
		"device-1",
		getSimpleTestRegister("register-a"),
		101,
	))

	storage.Fill(NewNumericRegisterValue(
		"device-2",
		getSimpleTestRegister("register-a"),
		200,
	))

	{
		expected := []string{
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
			"device-1:register-a=101.000000",
			"device-2:register-a=200.000000",
		}
		storage.Wait()
		got := getAsStrings(storage.GetState(Filter{}))
		if !equalIgnoreOrder(expected, got) {
			t.Errorf("expected %#v but got %#v", expected, got)
		}
	}

}

func equalIgnoreOrder(a, b []string) bool {
	less := func(a, b string) bool { return a < b }
	return cmp.Diff(a, b, cmpopts.SortSlices(less)) == ""
}

func getAsStrings(state State) (lines []string) {
	for deviceName, values := range state {
		for _, v := range values {
			lines = append(lines, fmt.Sprintf("%s:%s", deviceName, v.String()))
		}
	}
	return
}
