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
		i := i
		subscription := storage.Subscribe(ctx, dataflow.Filter{})
		go func() {
			counter := 0
			defer wg.Done()
			defer func() {
				counts <- counter
			}()
			for v := range subscription.GetOutput() {
				t.Logf("worker %d got %s", i, v)
				counter += 1
			}
		}()
	}

	expected := fillSetA(storage)
	storage.Wait()
	cancel()
	wg.Wait()
	close(counts)

	{
		i := 0
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
