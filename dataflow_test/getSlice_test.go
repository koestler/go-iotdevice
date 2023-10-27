package dataflow_test

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/koestler/go-iotdevice/dataflow"
	mock_dataflow "github.com/koestler/go-iotdevice/dataflow/mock"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestValueStorageGetSlice(t *testing.T) {
	ctrl := gomock.NewController(t)

	storage := dataflow.NewValueStorage()

	fillSetA(storage)
	storage.Wait()

	t.Run("setA", func(t *testing.T) {
		expect := []string{
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
			"device-1:register-a=100.000000",
		}
		got := getAsStrings(storage.GetState())
		if !equalIgnoreOrder(expect, got) {
			t.Errorf("expect %#v but got %#v", expect, got)
		}
	})

	fillSetB(storage)
	storage.Wait()

	t.Run("setAB", func(t *testing.T) {
		expect := []string{
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
			"device-1:register-a=101.000000",
			"device-2:register-a=200.000000",
		}
		got := getAsStrings(storage.GetState())
		if !equalIgnoreOrder(expect, got) {
			t.Errorf("expect %#v but got %#v", expect, got)
		}
	})

	fillSetC(storage)
	storage.Wait()

	t.Run("setABCfilterDevice", func(t *testing.T) {
		expect := []string{
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
		}
		got := getAsStrings(storage.GetStateFiltered(dataflow.DeviceNameValueFilter("device-0")))
		if !equalIgnoreOrder(expect, got) {
			t.Errorf("expect %#v but got %#v", expect, got)
		}
	})

	t.Run("setABCfilterRegister", func(t *testing.T) {
		expect := []string{
			"device-0:register-a=1.000000",
		}

		fc := mock_dataflow.NewMockRegisterFilterConf(ctrl)
		fc.EXPECT().SkipRegisters().Return([]string{"register-b"}).AnyTimes()
		fc.EXPECT().IncludeRegisters().Return([]string{}).AnyTimes()
		fc.EXPECT().SkipCategories().Return([]string{"set-b", "set-c"}).AnyTimes()
		fc.EXPECT().IncludeCategories().Return([]string{}).AnyTimes()
		fc.EXPECT().DefaultInclude().Return(true).AnyTimes()

		got := getAsStrings(storage.GetStateFiltered(dataflow.RegisterValueFilter(fc)))

		if !equalIgnoreOrder(expect, got) {
			t.Errorf("expect %#v but got %#v", expect, got)
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

func BenchmarkValueStorageGetState(b *testing.B) {
	storage := dataflow.NewValueStorage()
	fillSetA(storage)
	fillSetB(storage)
	fillSetC(storage)
	storage.Wait()

	for i := 0; i < b.N; i++ {
		storage.GetState()
	}
}

func BenchmarkValueStorageGetStateFiltered(b *testing.B) {
	storage := dataflow.NewValueStorage()
	fillSetA(storage)
	fillSetB(storage)
	fillSetC(storage)
	storage.Wait()

	deviceFilter := func(value dataflow.Value) bool {
		return value.DeviceName() == "device-0"
	}

	for i := 0; i < b.N; i++ {
		storage.GetStateFiltered(deviceFilter)
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
