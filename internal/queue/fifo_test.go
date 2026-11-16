package queue

import (
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
				if q.Enqueue(i) {
					t.Fatalf("enqueue %d: unexpected replace", i)
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
			if q.Enqueue(1) || q.Enqueue(2) {
				t.Fatal("unexpected replace")
			}
			// full: replaces the oldest element (1)
			if !q.Enqueue(3) {
				t.Fatal("expected replace")
			}

			// the run goroutine may already have offered the replaced element 1
			got := []int{receive(t, q.Dequeue()), receive(t, q.Dequeue())}
			if got[0] == 1 {
				got = []int{got[1], receive(t, q.Dequeue())}
			}
			if got[0] != 2 || got[1] != 3 {
				t.Fatalf("expected [2 3], got %v", got)
			}
		})
	})

	t.Run("length 1", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			q := NewFIFO[int](1)
			defer q.Close()

			if q.Enqueue(1) {
				t.Fatal("unexpected replace")
			}
			if !q.Enqueue(2) {
				t.Fatal("expected replace")
			}
			// the run goroutine may already have offered the replaced element 1
			got := receive(t, q.Dequeue())
			if got == 1 {
				got = receive(t, q.Dequeue())
			}
			if got != 2 {
				t.Fatalf("expected 2, got %d", got)
			}
		})
	})

	t.Run("zero length", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			NewFIFO[int](0)
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

	t.Run("next after close", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			q := NewFIFO[int](1)
			_ = q.Enqueue(1)
			e := <-q.Dequeue()
			q.Close()
			e.Next(true) // must not block
			if _, ok := <-q.Dequeue(); ok {
				t.Fatal("expected closed channel")
			}
		})
	})

	t.Run("next twice panics", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			q := NewFIFO[int](1)
			defer q.Close()
			_ = q.Enqueue(1)
			e := <-q.Dequeue()
			e.Next(false)
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			e.Next(true)
		})
	})

	t.Run("replace in-flight element", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			q := NewFIFO[int](2)
			defer q.Close()
			_ = q.Enqueue(1)
			_ = q.Enqueue(2)

			e := <-q.Dequeue() // 1 is in flight
			if !q.Enqueue(3) { // replaces 1
				t.Fatal("expected replace")
			}
			e.Next(true) // must not remove 2

			for _, want := range []int{2, 3} {
				if got := receive(t, q.Dequeue()); got != want {
					t.Fatalf("expected %d, got %d", want, got)
				}
			}
		})
	})

	t.Run("keep after replace", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			q := NewFIFO[int](2)
			defer q.Close()
			_ = q.Enqueue(1)
			_ = q.Enqueue(2)

			e := <-q.Dequeue()
			q.Enqueue(3) // replaces 1
			e.Next(false)

			for _, want := range []int{2, 3} {
				if got := receive(t, q.Dequeue()); got != want {
					t.Fatalf("expected %d, got %d", want, got)
				}
			}
		})
	})

	t.Run("wraparound", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			q := NewFIFO[int](3)
			defer q.Close()
			for _, batch := range [][]int{{1, 2}, {3, 4, 5}, {6, 7}} {
				for _, v := range batch {
					if q.Enqueue(v) {
						t.Fatal("unexpected replace")
					}
				}
				for _, want := range batch {
					if got := receive(t, q.Dequeue()); got != want {
						t.Fatalf("expected %d, got %d", want, got)
					}
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
						if q.Enqueue(i) {
							t.Errorf("enqueue: unexpected replace")
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
