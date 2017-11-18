package dataflow

import (
	"math"
)

func round(f float64) float64 {
	return math.Floor(f + .5)
}

func roundDecimals(number float64, decimals int) (float64) {
	shift := math.Pow(10, float64(decimals))
	return round(number*shift) / shift;
}

type Rounder struct {
	input, output chan Value
}

func RounderCreate() (*Rounder) {
	rounder := Rounder{
		input:  make(chan Value),
		output: make(chan Value),
	}

	go func() {
		defer close(rounder.output)
		for value := range rounder.input {
			value.Value = roundDecimals(value.Value, value.RoundDecimals)
			rounder.output <- value
		}
	}()

	return &rounder
}

func (rounder *Rounder) Fill(input <-chan Value) {
	go func() {
		for value := range input {
			rounder.input <- value
		}
	}()
}

func (rounder *Rounder) Drain() <-chan Value {
	return rounder.output
}

func (rounder *Rounder) Append(fillable Fillable) Fillable {
	fillable.Fill(rounder.Drain())
	return fillable
}
