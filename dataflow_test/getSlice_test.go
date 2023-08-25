package dataflow_test

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/koestler/go-iotdevice/dataflow"
	"testing"
)

func TestValueStorageGetSlice(t *testing.T) {
	storage := dataflow.NewValueStorage()

	fillSetA(storage)
	storage.Wait()

	t.Run("setA", func(t *testing.T) {
		expected := []string{
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
			"device-1:register-a=100.000000",
		}
		got := getAsStrings(storage.GetSlice(dataflow.Filter{}))
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
		got := getAsStrings(storage.GetSlice(dataflow.Filter{}))
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
		got := getAsStrings(storage.GetSlice(dataflow.Filter{
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
		got := getAsStrings(storage.GetSlice(dataflow.Filter{
			IncludeDevices: map[string]bool{"device-0": true, "device-1": true, "device-2": true},
			SkipRegisterNames: map[dataflow.SkipRegisterNameStruct]bool{
				dataflow.SkipRegisterNameStruct{
					Device:   "device-1",
					Register: "register-a",
				}: true,
				dataflow.SkipRegisterNameStruct{
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
		got := getAsStrings(storage.GetSlice(dataflow.Filter{
			IncludeDevices: map[string]bool{"device-0": true, "device-3": true},
			SkipRegisterCategories: map[dataflow.SkipRegisterCategoryStruct]bool{dataflow.SkipRegisterCategoryStruct{
				Device:   "device-3",
				Category: "set-c",
			}: true},
		}))
		if !equalIgnoreOrder(expected, got) {
			t.Errorf("expected %#v but got %#v", expected, got)
		}
	})
}

func BenchmarkValueStorageFill(b *testing.B) {
	storage := dataflow.NewValueStorage()

	for i := 0; i < b.N; i++ {
		storage.Fill(dataflow.NewNumericRegisterValue(
			"device-0",
			getSimpleTestRegister("categoryName", "registerName"),
			float64(i),
		))
	}
}

func BenchmarkValueStorageGetSlice(b *testing.B) {
	storage := dataflow.NewValueStorage()
	fillSetA(storage)
	fillSetB(storage)
	fillSetC(storage)
	storage.Wait()

	for i := 0; i < b.N; i++ {
		storage.GetSlice(dataflow.Filter{})
	}
}

func equalIgnoreOrder(a, b []string) bool {
	less := func(a, b string) bool { return a < b }
	return cmp.Diff(a, b, cmpopts.SortSlices(less)) == ""
}

func getAsStrings(values []dataflow.Value) (lines []string) {
	for _, v := range values {
		lines = append(lines, fmt.Sprintf("%s:%s", v.DeviceName(), v.String()))
	}
	return
}
