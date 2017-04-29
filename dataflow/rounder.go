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

func Rounder(input <-chan Value) <-chan Value {
	output := make(chan Value)

	go func() {
		for value := range input {
			value.Value = roundDecimals(value.Value, value.RoundDecimals)
			output <- value
		}
		close(output)
	}()

	return output
}
