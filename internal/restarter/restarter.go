// Package restarter implements a watchdog for long-running go routines. Whenever the routine stops and return an error,
// it is restarted periodically.
package restarter

import (
	"context"
	"log"
	"sync"
	"time"
)

type Config interface {
	RestartInterval() time.Duration
	RestartIntervalMaxBackoff() time.Duration
}

type Restartable interface {
	Name() string
	Run(ctx context.Context) (err error, immediateError bool)
}

type Restarter[S Restartable] struct {
	config    Config
	service   S
	isRunning bool

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func CreateRestarter[S Restartable](config Config, service S) (w *Restarter[S]) {
	ctx, cancel := context.WithCancel(context.Background())
	w = &Restarter[S]{
		config:  config,
		service: service,
		ctx:     ctx,
		cancel:  cancel,
	}

	return
}

func (w *Restarter[S]) Run() {
	// check if context is already canceled
	if w.ctx.Err() != nil {
		return
	}

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		immediateErrorsInARow := 0
		first := true
		for {
			if first {
				first = false
			} else {
				log.Printf("restarter[%s]: start", w.service.Name())
			}

			start := time.Now()
			err, immediateError := w.service.Run(w.ctx)
			if err == nil {
				// shutdown
				return
			} else {
				log.Printf("restarter[%s]: terminated with error: %s", w.service.Name(), err)
				runningFor := time.Since(start)

				if immediateError {
					immediateErrorsInARow += 1
				} else {
					immediateErrorsInARow = 0
				}

				retryIn := w.getRestartInterval(immediateErrorsInARow)

				log.Printf("restarter[%s]: error after %s, expoential backoff, retry in %s", w.service.Name(), runningFor, retryIn)

				select {
				case <-w.ctx.Done():
					return
				case <-time.After(retryIn):
				}
			}
		}
	}()
}

func (w *Restarter[S]) Shutdown() {
	w.cancel()
	w.wg.Wait()
}

func (w *Restarter[S]) GetCtx() context.Context {
	return w.ctx
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

func (w *Restarter[S]) getRestartInterval(errorsInARow int) time.Duration {
	if errorsInARow > 16 {
		errorsInARow = 16
	}
	var backoffFactor uint64 = 1 << errorsInARow // 2^errorsInARow
	interval := w.config.RestartInterval() * time.Duration(backoffFactor)
	max := w.config.RestartIntervalMaxBackoff()
	if interval > max {
		return max
	}
	return interval
}
