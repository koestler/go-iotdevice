package dataflow

type Value struct {
	Device        *Device
	Name          string
	Value         float64
	Unit          string
	RoundDecimals int
}
