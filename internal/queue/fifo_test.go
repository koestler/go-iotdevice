package queue

import (
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

func TestFifo(t *testing.T) {
	t.Run("order", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			q := NewFIFO[int](10)
			defer q.Close()

			for i := range 5 {
				if err := q.Enqueue(i); err != nil {
					t.Fatalf("enqueue %d: %v", i, err)
				}
			}
			for i := range 5 {
				if got := receive(t, q.Dequeue()); got != i {
					t.Errorf("expected %d, got %d", i, got)
				}
			}
		})
	})

	t.Run("full", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			q := NewFIFO[int](2)
			defer q.Close()

			// no reader: the routine holds the front element until it is received
			if err := q.Enqueue(1); err != nil {
				t.Fatal(err)
			}
			if err := q.Enqueue(2); err != nil {
				t.Fatal(err)
			}
			if err := q.Enqueue(3); !errors.Is(err, ErrFull) {
				t.Fatalf("expected ErrFull, got %v", err)
			}

			receive(t, q.Dequeue())
			if err := q.Enqueue(3); err != nil {
				t.Fatalf("expected space after dequeue, got %v", err)
			}
		})
	})

	t.Run("zero length", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			q := NewFIFO[int](0)
			defer q.Close()
			if err := q.Enqueue(1); !errors.Is(err, ErrFull) {
				t.Fatalf("expected ErrFull, got %v", err)
			}
		})
	})

	t.Run("dequeue blocks until value", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			q := NewFIFO[int](1)
			defer q.Close()

			select {
			case <-q.Dequeue():
				t.Fatal("received value from empty fifo")
			case <-time.After(time.Second):
			}

			go func() {
				time.Sleep(time.Second)
				_ = q.Enqueue(42)
			}()
			if got := receive(t, q.Dequeue()); got != 42 {
				t.Errorf("expected 42, got %d", got)
			}
		})
	})

	t.Run("keep redelivers", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			q := NewFIFO[int](2)
			defer q.Close()
			_ = q.Enqueue(1)
			_ = q.Enqueue(2)

			e := <-q.Dequeue()
			e.Next(false)
			e = <-q.Dequeue()
			if e.Value != 1 {
				t.Fatalf("expected 1 redelivered, got %d", e.Value)
			}
			e.Next(true)
			if got := receive(t, q.Dequeue()); got != 2 {
				t.Errorf("expected 2, got %d", got)
			}
		})
	})

	t.Run("close", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			q := NewFIFO[int](1)
			q.Close()

			select {
			case _, ok := <-q.Dequeue():
				if ok {
					t.Fatal("expected closed channel")
				}
			case <-time.After(time.Second):
				t.Fatal("dequeue channel not closed")
			}
		})
	})

	t.Run("close with pending value", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			q := NewFIFO[int](1)
			_ = q.Enqueue(1)
			q.Close()

			deadline := time.After(time.Second)
			for {
				select {
				case _, ok := <-q.Dequeue():
					if !ok {
						return
					}
				case <-deadline:
					t.Fatal("dequeue channel not closed")
				}
			}
		})
	})

	t.Run("concurrent", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			const producers, perProducer = 8, 500
			q := NewFIFO[int](producers * perProducer)
			defer q.Close()

			var wg sync.WaitGroup
			for range producers {
				wg.Go(func() {
					for i := range perProducer {
						if err := q.Enqueue(i); err != nil {
							t.Errorf("enqueue: %v", err)
							return
						}
					}
				})
			}

			counts := make(map[int]int)
			for range producers * perProducer {
				counts[receive(t, q.Dequeue())]++
			}
			wg.Wait()

			for i := range perProducer {
				if counts[i] != producers {
					t.Errorf("value %d received %d times, expected %d", i, counts[i], producers)
				}
			}
		})
	})
}

func receive[T any](t *testing.T, ch <-chan Element[T]) T {
	t.Helper()
	select {
	case v, ok := <-ch:
		if !ok {
			t.Fatal("channel unexpectedly closed")
		}
		v.Next(true)
		return v.Value
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for value")
	}
	panic("unreachable")
}
