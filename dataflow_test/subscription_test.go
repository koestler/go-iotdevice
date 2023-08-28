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
		subscription := storage.Subscribe(ctx, dataflow.EmptyFilter)
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
		expect := fillSetALength + fillSetBLength + fillSetCLength
		for got := range counts {
			if expect != got {
				t.Errorf("expect count=%d but got %d", expect, got)
			}
			i += 1
		}
		if numberOfSubscriptions != i {
			t.Errorf("expect to receive %d counts but got %d", numberOfSubscriptions, i)
		}
	}
}

func TestValueStorageSubscribeWithFilter(t *testing.T) {
	run := func(filter dataflow.FilterFunc) (values []dataflow.Value) {
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
		values := run(dataflow.DeviceFilter("device-0"))

		// check values
		expect := []string{
			"device-0:register-a=0.000000",
			"device-0:register-a=1.000000",
			"device-0:register-b=10.000000",
		}
		got := getAsStrings(values)

		if !equalIgnoreOrder(expect, got) {
			t.Errorf("expect %#v but got %#v", expect, got)
		}
	})

	t.Run("filterSkipRegisterCategories", func(t *testing.T) {
		values := run(dataflow.RegisterFilter(
			[]string{"register-b"},
			[]string{"set-c"},
		))

		// check values
		expect := []string{
			"device-0:register-a=0.000000",
			"device-0:register-a=1.000000",
			"device-1:register-a=100.000000",
			"device-1:register-a=101.000000",
			"device-2:register-a=200.000000",
		}
		got := getAsStrings(values)

		if !equalIgnoreOrder(expect, got) {
			t.Errorf("expect %#v but got %#v", expect, got)
		}
	})
}
