package gensetDevice

import (
	"errors"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/genset"
	"log"
)

var ErrUnknownInputName = errors.New("unknown input name")

func enumSetter(name string, f func(bool) func(genset.Inputs) genset.Inputs) func(*genset.Controller, dataflow.Value) {
	return func(c *genset.Controller, v dataflow.Value) {
		if v, ok := v.(dataflow.EnumRegisterValue); ok {
			c.UpdateInputs(f(v.EnumIdx() != 0))
		} else {
			log.Printf("gensetDevice: %s: expected an enum, got %s", name, v.Register().RegisterType())
		}
	}
}

func numberSetter(name string, f func(float64) func(genset.Inputs) genset.Inputs) func(*genset.Controller, dataflow.Value) {
	return func(c *genset.Controller, v dataflow.Value) {
		if v, ok := v.(dataflow.NumericRegisterValue); ok {
			c.UpdateInputs(f(v.Value()))
		} else {
			log.Printf("gensetDevice: %s: expected a number, got %s", name, v.Register().RegisterType())
		}
	}
}

func inpSetter(name string) (func(*genset.Controller, dataflow.Value), error) {
	switch name {
	case "ResetSwitch":
		return enumSetter(name, func(v bool) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.ResetSwitch = v
				return i
			}
		}), nil
	case "CommandSwitch":
		return enumSetter(name, func(v bool) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.CommandSwitch = v
				return i
			}
		}), nil
	case "IOAvailable":
		return enumSetter(name, func(v bool) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.IOAvailable = v
				return i
			}
		}), nil
	case "ArmSwitch":
		return enumSetter(name, func(v bool) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.ArmSwitch = v
				return i
			}
		}), nil
	case "FireDetected":
		return enumSetter(name, func(v bool) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.FireDetected = v
				return i
			}
		}), nil
	case "EngineTemp":
		return numberSetter(name, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.EngineTemp = v
				return i
			}
		}), nil
	case "AuxTemp0":
		return numberSetter(name, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.AuxTemp0 = v
				return i
			}
		}), nil
	case "AuxTemp1":
		return numberSetter(name, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.AuxTemp1 = v
				return i
			}
		}), nil
	case "OutputAvailable":
		return enumSetter(name, func(v bool) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.OutputAvailable = v
				return i
			}
		}), nil
	case "U0":
		return numberSetter(name, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.U0 = v
				return i
			}
		}), nil
	case "U1":
		return numberSetter(name, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.U1 = v
				return i
			}
		}), nil
	case "U2":
		return numberSetter(name, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.U2 = v
				return i
			}
		}), nil
	case "L0":
		return numberSetter(name, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.L0 = v
				return i
			}
		}), nil
	case "L1":
		return numberSetter(name, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.L1 = v
				return i
			}
		}), nil
	case "L2":
		return numberSetter(name, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.L2 = v
				return i
			}
		}), nil
	case "F":
		return numberSetter(name, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.F = v
				return i
			}
		}), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownInputName, name)
	}
}
