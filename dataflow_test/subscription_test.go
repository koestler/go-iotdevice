package dataflow_test

import (
	"context"
	"github.com/koestler/go-iotdevice/dataflow"
	"sync"
	"testing"
)

func TestValueStorageSubscribe(t *testing.T) {
	storage := dataflow.NewValueStorage()
	ctx, cancel := context.WithCancel(context.Background())

	numberOfSubscriptions := 42

	counts := make(chan int, numberOfSubscriptions)
	wg := sync.WaitGroup{}
	wg.Add(numberOfSubscriptions)
	for i := 0; i < numberOfSubscriptions; i += 1 {
		subscription := storage.Subscribe(ctx, dataflow.Filter{})
		go func() {
			counter := 0
			defer wg.Done()
			defer func() {
				counts <- counter
			}()
			for range subscription.Drain() {
				counter += 1
			}
		}()
	}

	fillSetA(storage)
	fillSetB(storage)
	fillSetC(storage)
	storage.Wait()
	cancel()
	wg.Wait()
	close(counts)

	{
		i := 0
		expected := fillSetALength + fillSetBLength + fillSetCLength
		for got := range counts {
			if expected != got {
				t.Errorf("expected count=%d but got %d", expected, got)
			}
			i += 1
		}
		if numberOfSubscriptions != i {
			t.Errorf("expected to receive %d counts but got %d", numberOfSubscriptions, i)
		}
	}
}

func TestValueStorageSubscribeWithFilter(t *testing.T) {
	run := func(filter dataflow.Filter) (values []dataflow.Value) {
		storage := dataflow.NewValueStorage()
		ctx, cancel := context.WithCancel(context.Background())
		subscription := storage.Subscribe(ctx, filter)

		wg := sync.WaitGroup{}
		wg.Add(1)
		values = make([]dataflow.Value, 0)
		go func() {
			defer wg.Done()
			for v := range subscription.Drain() {
				values = append(values, v)
			}
		}()

		// send values to storage
		fillSetA(storage)
		fillSetB(storage)
		fillSetC(storage)
		storage.Wait()
		cancel()
		wg.Wait()

		return
	}

	t.Run("filterDevice", func(t *testing.T) {
		values := run(dataflow.Filter{IncludeDevices: map[string]bool{"device-0": true}})

		// check values
		expected := []string{
			"device-0:register-a=0.000000",
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
		}
		got := getAsStrings(values)

		if !equalIgnoreOrder(expected, got) {
			t.Errorf("expected %#v but got %#v", expected, got)
		}
	})

	t.Run("filterSkipRegisterCategories", func(t *testing.T) {
		values := run(dataflow.Filter{
			IncludeDevices: map[string]bool{"device-0": true, "device-3": true},
			SkipRegisterCategories: map[dataflow.SkipRegisterCategoryStruct]bool{dataflow.SkipRegisterCategoryStruct{
				Device:   "device-3",
				Category: "set-c",
			}: true},
		})

		// check values
		expected := []string{
			"device-0:register-a=0.000000",
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
		}
		got := getAsStrings(values)

		if !equalIgnoreOrder(expected, got) {
			t.Errorf("expected %#v but got %#v", expected, got)
		}
	})
}
