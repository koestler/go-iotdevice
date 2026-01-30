package gensetDevice

// Average computes a moving average over a fixed-size window.
type Average struct {
	window int
	idx    int //  current index in the circular buffer
	arr    []float64
}

// NewAverage creates a new Average with the specified window size.
func NewAverage(window int) *Average {
	return &Average{
		window: window,
		arr:    make([]float64, 0, window),
	}
}

func (a *Average) Add(value float64) {
	if len(a.arr) < a.window {
		a.arr = append(a.arr, value)
	} else {
		a.arr[a.idx] = value
	}

	a.idx += 1
	a.idx %= a.window
}

func (a *Average) Value() float64 {
	if len(a.arr) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range a.arr {
		sum += v
	}
	return sum / float64(len(a.arr))
}
