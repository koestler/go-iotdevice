package generator_test

import (
	"testing"
)

type logger interface {
	Logf(format string, args ...any)
}

type tracker[T comparable] struct {
	req    chan chan []T
	logger logger
	track  []T
}

func newTracker[T comparable](logger logger) *tracker[T] {
	return &tracker[T]{
		req:    make(chan chan []T),
		logger: logger,
	}
}

func (tr *tracker[T]) Drain(c <-chan T) {
	for {
		select {
		case u, ok := <-c:
			if !ok {
				return
			}
			if tr.logger != nil {
				tr.logger.Logf("got %v", u)
			}
			tr.track = append(tr.track, u)
		case r := <-tr.req:
			r <- tr.track
		}
	}
}

func (tr *tracker[T]) Track() []T {
	r := make(chan []T)
	tr.req <- r
	return <-r
}

func (tr *tracker[T]) Latest() (ok bool, r T) {
	track := tr.Track()
	if len(track) == 0 {
		return false, r
	}
	return true, track[len(track)-1]
}

func (tr *tracker[T]) AssertLatest(t *testing.T, expect T) {
	t.Helper()
	if ok, got := tr.Latest(); !ok {
		t.Errorf("track empty, expected %v", expect)
	} else if got != expect {
		t.Errorf("got %v, expected %v", got, expect)
	}
}

func (tr *tracker[T]) Assert(t *testing.T, expect []T) {
	t.Helper()

	track := tr.Track()

	if len(track) != len(expect) {
		t.Errorf("got %v, want %v", tr.track, expect)
		return
	}
	for i, got := range track {
		if got != expect[i] {
			t.Errorf("got %v, want %v", got, expect[i])
		}
	}
}
