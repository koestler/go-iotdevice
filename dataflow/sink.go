package dataflow

import "log"

func SinkLog(prefix string, input <-chan Value) {
	go func() {
		for value := range input {
			log.Printf("%s: %v", prefix, value)
		}
	}()
}
