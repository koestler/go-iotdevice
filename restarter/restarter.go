// Package restarter implements a watchdog for long-running go routines. Whenever the routine stops and return an error,
// it is restarted periodically.
package restarter

import (
	"context"
	"log"
	"sync"
	"time"
)

type Restartable interface {
	Name() string
	Run(ctx context.Context) error
}

type Restarter[S Restartable] struct {
	service   S
	isRunning bool

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func RunRestarter[S Restartable](service S) (w *Restarter[S]) {
	ctx, cancel := context.WithCancel(context.Background())
	w = &Restarter[S]{
		service: service,
		ctx:     ctx,
		cancel:  cancel,
	}

	go func() {
		w.wg.Add(1)
		defer w.wg.Done()

		for {
			log.Printf("restarter[%s]: start", service.Name())

			err := service.Run(w.ctx)
			if err != nil {
				log.Printf("restarter[%s]: terminated with error: %s", service.Name(), err)
			} else {
				log.Printf("restarter[%s]: terminated", service.Name())
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

func (w *Restarter[S]) Shutdown() {
	w.cancel()
	w.wg.Wait()
}

func (w *Restarter[S]) Name() string {
	return w.service.Name()
}

func (w *Restarter[S]) Service() S {
	return w.service
}

func (w *Restarter[S]) IsRunning() bool {
	return w.isRunning
}
