package genset

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIoError(t *testing.T) {
	p := Params{EngineTempMin: 10, EngineTempMax: 90, AuxTemp0Max: 100, AuxTemp1Max: 100}
	ok := Inputs{IOAvailable: true, EngineTemp: 50}

	assert.NoError(t, ioError(p, ok))

	i := ok
	i.EngineTemp = 95
	assert.ErrorIs(t, ioError(p, i), ErrEngineTempHigh)

	i = ok
	i.EngineTemp = math.NaN()
	assert.Error(t, ioError(p, i))
}

func TestOutputError(t *testing.T) {
	p := Params{SinglePhase: true, UMin: 225, UMax: 235, FMin: 49, FMax: 51, PMax: 1000, PTotMax: 1000}
	ok := Inputs{OutputAvailable: true, U1: 230, F: 50}

	assert.NoError(t, outputError(p, ok))

	i := ok
	i.U1 = 200
	assert.ErrorIs(t, outputError(p, i), ErrU1Low)

	i = ok
	i.P1 = math.NaN()
	assert.Error(t, outputError(p, i))
}
