package gensetDevice

import "testing"

func TestAverage_Window1(t *testing.T) {
	avg := NewAverage(1)

	// Before entering any value
	if got := avg.Value(); got != 0.0 {
		t.Errorf("Before adding values: got %.2f, want 0.0", got)
	}

	// After entering 1 value
	avg.Add(10.0)
	if got := avg.Value(); got != 10.0 {
		t.Errorf("After 1 value: got %.2f, want 10.0", got)
	}

	// After entering 2 values (2x window): avg of [20] = 20
	avg.Add(20.0)
	if got := avg.Value(); got != 20.0 {
		t.Errorf("After 2 values: got %.2f, want 20.0", got)
	}
}

func TestAverage_Window2(t *testing.T) {
	avg := NewAverage(2)

	// Before entering any value
	if got := avg.Value(); got != 0.0 {
		t.Errorf("Before adding values: got %.2f, want 0.0", got)
	}

	// After entering 1 value
	avg.Add(10.0)
	if got := avg.Value(); got != 10.0 {
		t.Errorf("After 1 value: got %.2f, want 10.0", got)
	}

	// After entering 2 values (window full): avg of [10, 20] = 15
	avg.Add(20.0)
	if got := avg.Value(); got != 15.0 {
		t.Errorf("After 2 values (window full): got %.2f, want 15.0", got)
	}

	// After entering 4 values (2x window): avg of [30, 40] = 35
	avg.Add(30.0)
	avg.Add(40.0)
	if got := avg.Value(); got != 35.0 {
		t.Errorf("After 4 values: got %.2f, want 35.0", got)
	}
}

func TestAverage_Window4(t *testing.T) {
	avg := NewAverage(4)

	// Before entering any value
	if got := avg.Value(); got != 0.0 {
		t.Errorf("Before adding values: got %.2f, want 0.0", got)
	}

	// After entering 1 value
	avg.Add(10.0)
	if got := avg.Value(); got != 10.0 {
		t.Errorf("After 1 value: got %.2f, want 10.0", got)
	}

	// After entering 2 values: avg of [10, 20] = 15
	avg.Add(20.0)
	if got := avg.Value(); got != 15.0 {
		t.Errorf("After 2 values: got %.2f, want 15.0", got)
	}

	// After entering 4 values (window full): avg of [10, 20, 30, 40] = 25
	avg.Add(30.0)
	avg.Add(40.0)
	if got := avg.Value(); got != 25.0 {
		t.Errorf("After 4 values (window full): got %.2f, want 25.0", got)
	}

	// After entering 5 values (window + 1): avg of [20, 30, 40, 50] = 35
	avg.Add(50.0)
	if got := avg.Value(); got != 35.0 {
		t.Errorf("After 5 values: got %.2f, want 35.0", got)
	}

	// After entering 8 values (2x window): avg of [50, 60, 70, 80] = 65
	avg.Add(60.0)
	avg.Add(70.0)
	avg.Add(80.0)
	if got := avg.Value(); got != 65.0 {
		t.Errorf("After 8 values: got %.2f, want 65.0", got)
	}
}
