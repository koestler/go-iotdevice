package generator_test

import (
	"sync"
	"testing"
)

type tracker[T comparable] struct {
	sync.Mutex
	track []T
}

func (tr *tracker[T]) Drain(c <-chan T) {
	for u := range c {
		tr.Lock()
		tr.track = append(tr.track, u)
		tr.Unlock()
	}
}

func (tr *tracker[T]) Latest() T {
	tr.Lock()
	defer tr.Unlock()
	if len(tr.track) == 0 {
		var null T
		return null
	}
	return tr.track[len(tr.track)-1]
}

func (tr *tracker[T]) AssertLatest(t *testing.T, expect T) {
	t.Helper()
	if got := tr.Latest(); got != expect {
		t.Errorf("got %v, want %v", got, expect)
	}
}

func (tr *tracker[T]) Assert(t *testing.T, expect []T) {
	t.Helper()
	tr.Lock()
	defer tr.Unlock()
	if len(tr.track) != len(expect) {
		t.Errorf("got %v, want %v", tr.track, expect)
		return
	}
	for i, got := range tr.track {
		if got != expect[i] {
			t.Errorf("got %v, want %v", got, expect[i])
		}
	}
}
