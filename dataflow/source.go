package dataflow

import (
	"github.com/koestler/go-ve-sensor/bmv"
	"time"
	"math/rand"
)

type Source struct {
	outputChain chan Value
}

func SourceCreateBmvStartDummy(device *Device, registers bmv.Registers) (*Source) {
	output := make(chan Value)
	go func() {
		defer close(output)
		for _ = range time.Tick(time.Second) {
			for name, register := range registers {
				output <- Value{
					Device:        device,
					Name:          name,
					Value:         1e2 * rand.Float64() * register.Factor,
					Unit:          register.Unit,
					RoundDecimals: register.RoundDecimals,
				}
			}
		}
	}()
	return &Source{
		outputChain: output,
	}
}

func (source *Source) Drain() <-chan Value {
	return source.outputChain
}

func (source *Source) Append(fillable Fillable) Fillable {
	fillable.Fill(source.Drain())
	return fillable
}
