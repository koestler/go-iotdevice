package gpioDevice

import (
	"errors"
	"fmt"
	"log"
	"periph.io/x/host/v3"
	"sync"
)

var ErrHostInit = errors.New("init failed")

var hostInitSync sync.Once
var hostInitError error

func hostInitOnce() error {
	hostInitSync.Do(func() {
		hostInitError = hostInit()
	})
	return hostInitError
}

func hostInit() error {
	state, err := host.Init()
	if err != nil {
		return fmt.Errorf("host: %w: %s", ErrHostInit, err)
	}

	if len(state.Loaded) < 1 {
		err := fmt.Errorf("host: %w: no driver loaded", ErrHostInit)

		log.Print(err)

		log.Printf("host: Drivers skipped:\n")
		for _, failure := range state.Skipped {
			log.Printf("host - %s: %s\n", failure.D, failure.Err)
		}

		log.Printf("host: Drivers failed to load:\n")
		for _, failure := range state.Failed {
			log.Printf("host - %s: %v\n", failure.D, failure.Err)
		}

		return err
	}

	log.Printf("host: initialized. Using drivers:\n")
	for _, driver := range state.Loaded {
		log.Printf("host - %s\n", driver)
	}

	return nil
}
