package dataflow

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"testing"
)

func getSimpleTestRegister(category, name string) RegisterStruct {
	return NewRegisterStruct(
		category,
		name,
		"",
		NumberRegister,
		map[int]string{},
		"",
		40,
		false,
	)
}
func fillSetA(storage *ValueStorageInstance) {
	storage.Fill(NewNumericRegisterValue(
		"device-0",
		getSimpleTestRegister("set-a", "register-a"),
		0,
	))

	storage.Fill(NewNumericRegisterValue(
		"device-0",
		getSimpleTestRegister("set-a", "register-a"),
		1,
	))

	// filling the same register multiple times must not make a difference
	for i := 0; i < 10; i += 1 {
		storage.Fill(NewNumericRegisterValue(
			"device-0",
			getSimpleTestRegister("set-a", "register-b"),
			10,
		))

		storage.Fill(NewNumericRegisterValue(
			"device-1",
			getSimpleTestRegister("set-a", "register-a"),
			100,
		))
	}
}

func fillSetB(storage *ValueStorageInstance) {
	storage.Fill(NewNumericRegisterValue(
		"device-1",
		getSimpleTestRegister("set-b", "register-a"),
		101,
	))

	storage.Fill(NewNumericRegisterValue(
		"device-2",
		getSimpleTestRegister("set-b", "register-a"),
		200,
	))
}

func fillSetC(storage *ValueStorageInstance) {
	for i := 0; i < 1000; i += 1 {
		storage.Fill(NewNumericRegisterValue(
			"device-3",
			getSimpleTestRegister("set-c", fmt.Sprintf("register-%d", i)),
			float64(i),
		))
	}
}

func TestGetState(t *testing.T) {
	storage := NewValueStorage()

	fillSetA(storage)
	storage.Wait()

	t.Run("setA", func(t *testing.T) {
		expected := []string{
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
			"device-1:register-a=100.000000",
		}
		got := getAsStrings(storage.GetState(Filter{}))
		if !equalIgnoreOrder(expected, got) {
			t.Errorf("expected %#v but got %#v", expected, got)
		}
	})

	fillSetB(storage)
	storage.Wait()

	t.Run("setAB", func(t *testing.T) {
		expected := []string{
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
			"device-1:register-a=101.000000",
			"device-2:register-a=200.000000",
		}
		got := getAsStrings(storage.GetState(Filter{}))
		if !equalIgnoreOrder(expected, got) {
			t.Errorf("expected %#v but got %#v", expected, got)
		}
	})

	fillSetC(storage)
	storage.Wait()

	t.Run("setABCfilterIncludeDevices", func(t *testing.T) {
		expected := []string{
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
		}
		got := getAsStrings(storage.GetState(Filter{
			IncludeDevices: map[string]bool{"device-0": true},
		}))
		if !equalIgnoreOrder(expected, got) {
			t.Errorf("expected %#v but got %#v", expected, got)
		}
	})

	t.Run("setABCfilterSkipRegisterNames", func(t *testing.T) {
		expected := []string{
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
		}
		got := getAsStrings(storage.GetState(Filter{
			IncludeDevices: map[string]bool{"device-0": true, "device-1": true, "device-2": true},
			SkipRegisterNames: map[SkipRegisterNameStruct]bool{
				SkipRegisterNameStruct{
					Device:   "device-1",
					Register: "register-a",
				}: true,
				SkipRegisterNameStruct{
					Device:   "device-2",
					Register: "register-a",
				}: true,
			},
		}))
		if !equalIgnoreOrder(expected, got) {
			t.Errorf("expected %#v but got %#v", expected, got)
		}
	})

	t.Run("setABCfilterSkipRegisterCategories", func(t *testing.T) {
		expected := []string{
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
		}
		got := getAsStrings(storage.GetState(Filter{
			IncludeDevices: map[string]bool{"device-0": true, "device-3": true},
			SkipRegisterCategories: map[SkipRegisterCategoryStruct]bool{SkipRegisterCategoryStruct{
				Device:   "device-3",
				Category: "set-c",
			}: true},
		}))
		if !equalIgnoreOrder(expected, got) {
			t.Errorf("expected %#v but got %#v", expected, got)
		}
	})
}

func BenchmarkFill(b *testing.B) {
	storage := NewValueStorage()

	for i := 0; i < b.N; i++ {
		storage.Fill(NewNumericRegisterValue(
			"device-0",
			getSimpleTestRegister("categoryName", "registerName"),
			float64(i),
		))
	}
}

func BenchmarkGetState(b *testing.B) {
	storage := NewValueStorage()
	fillSetA(storage)
	fillSetB(storage)
	fillSetC(storage)
	storage.Wait()

	for i := 0; i < b.N; i++ {
		storage.GetState(Filter{})
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
