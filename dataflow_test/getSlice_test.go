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
		got := getAsStrings(storage.GetState())
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
		got := getAsStrings(storage.GetState())
		if !equalIgnoreOrder(expected, got) {
			t.Errorf("expected %#v but got %#v", expected, got)
		}
	})

	fillSetC(storage)
	storage.Wait()

	t.Run("setABCfilterDevice", func(t *testing.T) {
		expected := []string{
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
		}
		got := getAsStrings(storage.GetStateFiltered(dataflow.DeviceFilter("device-0")))
		if !equalIgnoreOrder(expected, got) {
			t.Errorf("expected %#v but got %#v", expected, got)
		}
	})

	t.Run("setABCfilterRegister", func(t *testing.T) {
		expected := []string{
			"device-0:register-a=1.000000",
		}
		got := getAsStrings(storage.GetStateFiltered(dataflow.RegisterFilter(
			[]string{"register-b"},
			[]string{"set-b", "set-c"},
		)))

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
		storage.GetState()
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
