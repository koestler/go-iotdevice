package gensetDevice

import (
	"errors"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/genset"
	"log"
)

var ErrUnknownInputName = errors.New("unknown input name")

func (d *DeviceStruct) enumSetter(
	name string,
	reg dataflow.RegisterStruct,
	f func(bool) func(genset.Inputs) genset.Inputs,
) func(*genset.Controller, dataflow.Value) {
	return func(c *genset.Controller, v dataflow.Value) {
		if v, ok := v.(dataflow.EnumRegisterValue); ok {
			c.UpdateInputs(f(v.EnumIdx() != 0))
			d.StateStorage().Fill(dataflow.NewEnumRegisterValue(
				d.Config().Name(),
				reg,
				v.EnumIdx(),
			))
		} else {
			log.Printf("gensetDevice: %s: expected an enum, got %s", name, v.Register().RegisterType())
		}
	}
}

func (d *DeviceStruct) numberSetter(
	name string,
	reg dataflow.RegisterStruct,
	f func(float64) func(genset.Inputs) genset.Inputs,
) func(*genset.Controller, dataflow.Value) {
	return func(c *genset.Controller, v dataflow.Value) {
		if v, ok := v.(dataflow.NumericRegisterValue); ok {
			c.UpdateInputs(f(v.Value()))
			d.StateStorage().Fill(dataflow.NewNumericRegisterValue(
				d.Config().Name(),
				reg,
				v.Value(),
			))
		} else {
			log.Printf("gensetDevice: %s: expected a number, got %s", name, v.Register().RegisterType())
		}
	}
}

func (d *DeviceStruct) inpSetter(name string) (func(*genset.Controller, dataflow.Value), error) {
	switch name {
	case "ResetSwitch":
		return d.enumSetter(name, ResetSwitchRegister,
			func(v bool) func(genset.Inputs) genset.Inputs {
				return func(i genset.Inputs) genset.Inputs {
					i.ResetSwitch = v
					return i
				}
			}), nil
	case "CommandSwitch":
		return d.enumSetter(name, CommandSwitchRegister,
			func(v bool) func(genset.Inputs) genset.Inputs {
				return func(i genset.Inputs) genset.Inputs {
					i.CommandSwitch = v
					return i
				}
			}), nil
	case "IOAvailable":
		return d.enumSetter(name, IOAvailableRegister,
			func(v bool) func(genset.Inputs) genset.Inputs {
				return func(i genset.Inputs) genset.Inputs {
					i.IOAvailable = v
					return i
				}
			}), nil
	case "ArmSwitch":
		return d.enumSetter(name, ArmSwitchRegister,
			func(v bool) func(genset.Inputs) genset.Inputs {
				return func(i genset.Inputs) genset.Inputs {
					i.ArmSwitch = v
					return i
				}
			}), nil
	case "FireDetected":
		return d.enumSetter(name, FireDetectedRegister, func(v bool) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.FireDetected = v
				return i
			}
		}), nil
	case "EngineTemp":
		return d.numberSetter(name, EngineTempRegister, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.EngineTemp = v
				return i
			}
		}), nil
	case "AuxTemp0":
		return d.numberSetter(name, AuxTemp0Register, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.AuxTemp0 = v
				return i
			}
		}), nil
	case "AuxTemp1":
		return d.numberSetter(name, AuxTemp1Register, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.AuxTemp1 = v
				return i
			}
		}), nil
	case "OutputAvailable":
		return d.enumSetter(name, OutputAvailableRegister, func(v bool) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.OutputAvailable = v
				return i
			}
		}), nil
	case "U0":
		return d.numberSetter(name, U0Register, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.U0 = v
				return i
			}
		}), nil
	case "U1":
		return d.numberSetter(name, U1Register, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.U1 = v
				return i
			}
		}), nil
	case "U2":
		return d.numberSetter(name, U2Register, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.U2 = v
				return i
			}
		}), nil
	case "L0":
		return d.numberSetter(name, L0Register, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.L0 = v
				return i
			}
		}), nil
	case "L1":
		return d.numberSetter(name, L1Register, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.L1 = v
				return i
			}
		}), nil
	case "L2":
		return d.numberSetter(name, L2Register, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.L2 = v
				return i
			}
		}), nil
	case "F":
		return d.numberSetter(name, FRegister, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.F = v
				return i
			}
		}), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownInputName, name)
	}
}
