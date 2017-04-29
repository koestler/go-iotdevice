package dataflow

import "log"

func SinkLog(value <-chan Value) {
	go func() {
		for {
			log.Printf("value sinked: %v", <-value)
		}
	}()
}
