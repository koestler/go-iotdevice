// Package watcher implements a watchdog for long-running go routines. Whenever the routine stops and return an error,
// it is restarted periodically.
package watcher

import (
	"context"
	"log"
	"sync"
	"time"
)

type Watchable interface {
	Name() string
	Run(ctx context.Context) error
}

type Watcher[S Watchable] struct {
	service   S
	isRunning bool

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func RunWatcher[S Watchable](service S) (w *Watcher[S]) {
	ctx, cancel := context.WithCancel(context.Background())
	w = &Watcher[S]{
		service: service,
		ctx:     ctx,
		cancel:  cancel,
	}

	go func() {
		w.wg.Add(1)
		defer w.wg.Done()

		for {
			log.Printf("watcher[%s]: start", service.Name())

			err := service.Run(w.ctx)
			if err != nil {
				log.Printf("watcher[%s]: terminated with error: %s", service.Name(), err)
			} else {
				log.Printf("watcher[%s]: terminated", service.Name())
			}

			// wait 2s for restart
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}()

	return
}

func (w *Watcher[S]) Shutdown() {
	w.cancel()
	w.wg.Wait()
}

func (w *Watcher[S]) Name() string {
	return w.service.Name()
}

func (w *Watcher[S]) Service() S {
	return w.service
}

func (w *Watcher[S]) IsRunning() bool {
	return w.isRunning
}
