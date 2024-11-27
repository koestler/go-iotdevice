package genset_test

import (
	"sync"
	"testing"
)

type logger interface {
	Logf(format string, args ...any)
}

type tracker[T comparable] struct {
	lock   sync.RWMutex
	logger logger
	track  []T
}

func newTracker[T comparable](logger logger) *tracker[T] {
	return &tracker[T]{
		logger: logger,
	}
}

func (tr *tracker[T]) OnUpdateFunc() func(T) {
	return func(u T) {
		if tr.logger != nil {
			tr.logger.Logf("update: %v", u)
		}
		tr.lock.Lock()
		defer tr.lock.Unlock()
		tr.track = append(tr.track, u)
	}
}

func (tr *tracker[T]) Track() []T {
	tr.lock.RLock()
	defer tr.lock.RUnlock()
	return append([]T(nil), tr.track...)
}

func (tr *tracker[T]) Latest() (ok bool, r T) {
	tr.lock.RLock()
	defer tr.lock.RUnlock()

	if len(tr.track) == 0 {
		return false, r
	}
	return true, tr.track[len(tr.track)-1]
}

func (tr *tracker[T]) AssertLatest(t *testing.T, expect T) {
	t.Helper()
	if ok, got := tr.Latest(); !ok {
		t.Errorf("track empty, expect %v", expect)
	} else if got != expect {
		t.Errorf("AssertLatest failed\ngot\t\t%v,\nexpect\t%v", got, expect)
	}
}

func (tr *tracker[T]) Assert(t *testing.T, expect []T) {
	t.Helper()
	track := tr.Track()

	if len(track) != len(expect) {
		t.Errorf("Assert failed\ngot\t\t%v,\nexpect\t%v", tr.track, expect)
		return
	}
	for i, got := range track {
		if got != expect[i] {
			t.Errorf("Assert failed\ngot\t\t%v,\nexpect\t%v", tr.track, expect)
			return
		}
	}
}
