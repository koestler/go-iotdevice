package genset

import (
	"errors"
	"fmt"
)

var ErrFireDirected = errors.New("fire detected")
var ErrIOUnavailable = errors.New("I/O controller unavailable")
var ErrEngineTempLow = errors.New("engine temperature too low")
var ErrEngineTempHigh = errors.New("engine temperature too high")
var ErrAuxTemp0Low = errors.New("auxiliary temperature 0 too low")
var ErrAuxTemp0High = errors.New("auxiliary temperature 0 too high")
var ErrAuxTemp1Low = errors.New("auxiliary temperature 1 too low")
var ErrAuxTemp1High = errors.New("auxiliary temperature 1 too high")

// Comparisons are written negated so that NaN fails the checks.
func ioError(p Params, i Inputs) error {
	if i.FireDetected {
		return ErrFireDirected
	}

	if !i.IOAvailable {
		return ErrIOUnavailable
	}

	if !(i.EngineTemp >= p.EngineTempMin) {
		return ErrEngineTempLow
	}
	if !(i.EngineTemp <= p.EngineTempMax) {
		return ErrEngineTempHigh
	}

	if !(i.AuxTemp0 >= p.AuxTemp0Min) {
		return ErrAuxTemp0Low
	}
	if !(i.AuxTemp0 <= p.AuxTemp0Max) {
		return ErrAuxTemp0High
	}

	if !(i.AuxTemp1 >= p.AuxTemp1Min) {
		return ErrAuxTemp1Low
	}
	if !(i.AuxTemp1 <= p.AuxTemp1Max) {
		return ErrAuxTemp1High
	}

	return nil
}

var ErrOutputUnavailable = errors.New("output measurements unavailable")
var ErrFrequencyLow = errors.New("frequency too low")
var ErrFrequencyHigh = errors.New("frequency too high")
var ErrU1Low = errors.New("U1 too low")
var ErrU1High = errors.New("U1 too high")
var ErrU2Low = errors.New("U2 too low")
var ErrU2High = errors.New("U2 too high")
var ErrU3Low = errors.New("U3 too low")
var ErrU3High = errors.New("U3 too high")
var ErrP1High = errors.New("P1 too high")
var ErrP2High = errors.New("P2 too high")
var ErrP3High = errors.New("P3 too high")
var ErrPTotHigh = errors.New("total power too high")

// Comparisons are written negated so that NaN fails the checks.
func outputError(p Params, i Inputs) error {
	if !i.OutputAvailable {
		return fmt.Errorf("%w: output available = %t", ErrOutputUnavailable, i.OutputAvailable)
	}

	if !(i.F >= p.FMin) {
		return fmt.Errorf("%w: %f < %f", ErrFrequencyLow, i.F, p.FMin)
	}
	if !(i.F <= p.FMax) {
		return fmt.Errorf("%w: %f > %f", ErrFrequencyHigh, i.F, p.FMax)
	}

	if !(i.U1 >= p.UMin) {
		return fmt.Errorf("%w: %f < %f", ErrU1Low, i.U1, p.UMin)
	}
	if !(i.U1 <= p.UMax) {
		return fmt.Errorf("%w: %f > %f", ErrU1High, i.U1, p.UMax)
	}

	if !(i.P1 <= p.PMax) {
		return fmt.Errorf("%w: %f > %f", ErrP1High, i.P1, p.PMax)
	}

	if p.SinglePhase {
		if !(i.P1 <= p.PTotMax) {
			return fmt.Errorf("%w: %f > %f", ErrPTotHigh, i.P1, p.PTotMax)
		}
	} else {
		if !(i.U2 >= p.UMin) {
			return fmt.Errorf("%w: %f < %f", ErrU2Low, i.U2, p.UMin)
		}
		if !(i.U2 <= p.UMax) {
			return fmt.Errorf("%w: %f > %f", ErrU2High, i.U2, p.UMax)
		}

		if !(i.U3 >= p.UMin) {
			return fmt.Errorf("%w: %f < %f", ErrU3Low, i.U3, p.UMin)
		}
		if !(i.U3 <= p.UMax) {
			return fmt.Errorf("%w: %f > %f", ErrU3High, i.U3, p.UMax)
		}

		if !(i.P2 <= p.PMax) {
			return fmt.Errorf("%w: %f > %f", ErrP2High, i.P2, p.PMax)
		}

		if !(i.P3 <= p.PMax) {
			return fmt.Errorf("%w: %f > %f", ErrP3High, i.P3, p.PMax)
		}

		pTotal := i.P1 + i.P2 + i.P3
		if !(pTotal <= p.PTotMax) {
			return fmt.Errorf("%w: %f > %f", ErrPTotHigh, pTotal, p.PTotMax)
		}
	}

	return nil
}
