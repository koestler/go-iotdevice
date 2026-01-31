package gensetDevice

import (
	"errors"
	"fmt"
	"log"

	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/genset"
)

var ErrUnknownInputName = errors.New("unknown input name")

func (d *DeviceStruct) enumSetter(
	name string,
	reg dataflow.RegisterStruct,
	f func(bool) func(genset.Inputs) genset.Inputs,
) func(*genset.Controller, dataflow.Value) {
	return func(c *genset.Controller, v dataflow.Value) {
		if ev, ok := v.(dataflow.EnumRegisterValue); ok {
			c.UpdateInputs(f(ev.EnumIdx() != 0))
			d.StateStorage().Fill(dataflow.NewEnumRegisterValue(
				d.Config().Name(),
				reg,
				ev.EnumIdx(),
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
		if nv, ok := v.(dataflow.NumericRegisterValue); ok {
			c.UpdateInputs(f(nv.Value()))
			d.StateStorage().Fill(dataflow.NewNumericRegisterValue(
				d.Config().Name(),
				reg,
				nv.Value(),
			))
		} else {
			log.Printf("gensetDevice: %s: expected a number, got %s", name, v.Register().RegisterType())
		}
	}
}

func (d *DeviceStruct) inpSetter(name string) (func(*genset.Controller, dataflow.Value), error) {
	switch name {
	case "ArmSwitch":
		return d.enumSetter(name, ArmSwitchRegister,
			func(v bool) func(genset.Inputs) genset.Inputs {
				return func(i genset.Inputs) genset.Inputs {
					i.ArmSwitch = v
					return i
				}
			}), nil
	case "ArmSwitchRO":
		return d.enumSetter(name, ArmSwitchRegisterRO,
			func(v bool) func(genset.Inputs) genset.Inputs {
				return func(i genset.Inputs) genset.Inputs {
					i.ArmSwitch = v
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
	case "CommandSwitchRO":
		return d.enumSetter(name, CommandSwitchRegisterRO,
			func(v bool) func(genset.Inputs) genset.Inputs {
				return func(i genset.Inputs) genset.Inputs {
					i.CommandSwitch = v
					return i
				}
			}), nil
	case "ResetSwitch":
		return d.enumSetter(name, ResetSwitchRegister,
			func(v bool) func(genset.Inputs) genset.Inputs {
				return func(i genset.Inputs) genset.Inputs {
					i.ResetSwitch = v
					return i
				}
			}), nil
	case "ResetSwitchRO":
		return d.enumSetter(name, ResetSwitchRegisterRO,
			func(v bool) func(genset.Inputs) genset.Inputs {
				return func(i genset.Inputs) genset.Inputs {
					i.ResetSwitch = v
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
	case "U1":
		return d.numberSetter(name, U1Register, func(v float64) func(genset.Inputs) genset.Inputs {
			avg := NewAverage(d.gensetConfig.UAvgWindow())
			return func(i genset.Inputs) genset.Inputs {
				avg.Add(v)
				i.U1 = avg.Value()
				return i
			}
		}), nil
	case "U2":
		return d.numberSetter(name, U2Register, func(v float64) func(genset.Inputs) genset.Inputs {
			avg := NewAverage(d.gensetConfig.UAvgWindow())
			return func(i genset.Inputs) genset.Inputs {
				avg.Add(v)
				i.U2 = avg.Value()
				return i
			}
		}), nil
	case "U3":
		return d.numberSetter(name, U3Register, func(v float64) func(genset.Inputs) genset.Inputs {
			avg := NewAverage(d.gensetConfig.UAvgWindow())
			return func(i genset.Inputs) genset.Inputs {
				avg.Add(v)
				i.U3 = avg.Value()
				return i
			}
		}), nil
	case "P1":
		return d.numberSetter(name, P1Register, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.P1 = v
				return i
			}
		}), nil
	case "P2":
		return d.numberSetter(name, P2Register, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.P2 = v
				return i
			}
		}), nil
	case "P3":
		return d.numberSetter(name, P3Register, func(v float64) func(genset.Inputs) genset.Inputs {
			return func(i genset.Inputs) genset.Inputs {
				i.P3 = v
				return i
			}
		}), nil
	case "F":
		return d.numberSetter(name, FRegister, func(v float64) func(genset.Inputs) genset.Inputs {
			avg := NewAverage(d.gensetConfig.FAvgWindow())
			return func(i genset.Inputs) genset.Inputs {
				avg.Add(v)
				i.F = avg.Value()
				log.Printf("gensetDevice: F input add=%f, value=%f", v, i.F)
				return i
			}
		}), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownInputName, name)
	}
}
