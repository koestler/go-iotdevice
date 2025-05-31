package gensetDevice

import (
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/koestler/go-iotdevice/v3/genset"
)

func boolToOnOff(b bool) int {
	if b {
		return 1
	}
	return 0
}

var OnOffEnum = map[int]string{
	boolToOnOff(false): "Off",
	boolToOnOff(true):  "On",
}

func addToRegisterDb(
	rdb *dataflow.RegisterDb, singlePhase bool, inputBindings []Binding,
) (
	commandRegisters []dataflow.RegisterStruct,
) {
	commandRegisters = make([]dataflow.RegisterStruct, 0)

	if isBindedInput("ArmSwitch", inputBindings) {
		rdb.AddStruct(ArmSwitchRegisterRO)
	} else {
		commandRegisters = append(commandRegisters, ArmSwitchRegister)
	}

	if isBindedInput("CommandSwitch", inputBindings) {
		rdb.AddStruct(CommandSwitchRegisterRO)
	} else {
		commandRegisters = append(commandRegisters, CommandSwitchRegister)
	}

	if isBindedInput("ResetSwitch", inputBindings) {
		rdb.AddStruct(ResetSwitchRegisterRO)
	} else {
		commandRegisters = append(commandRegisters, ResetSwitchRegister)
	}

	rdb.AddStruct(commandRegisters...)

	if singlePhase {
		rdb.AddStruct(InputRegisters1P...)
	} else {
		rdb.AddStruct(InputRegisters3P...)
	}
	rdb.AddStruct(StateRegisters...)
	rdb.AddStruct(OutputRegisters...)

	return commandRegisters
}

func isBindedInput(name string, inputBindings []Binding) bool {
	for _, b := range inputBindings {
		if b.RegisterName() == name {
			return true
		}
	}
	return false
}

func NewOnOffRegister(
	category, name, description string,
	sort int,
	writable bool,
) dataflow.RegisterStruct {
	return dataflow.NewRegisterStruct(
		category, name, description,
		dataflow.EnumRegister, OnOffEnum, "", sort, writable,
	)
}

func NewNumberRegister(
	category, name, description, unit string,
	sort int,
) dataflow.RegisterStruct {
	return dataflow.NewRegisterStruct(
		category, name, description,
		dataflow.NumberRegister, nil, unit, sort, false,
	)
}

var ArmSwitchRegister = NewOnOffRegister(
	"Switches",
	"ArmSwitch",
	"Arm switch",
	0,
	true,
)

var ArmSwitchRegisterRO = NewOnOffRegister(
	"Switches",
	"ArmSwitchRO",
	"Arm switch",
	0,
	false,
)

var CommandSwitchRegister = NewOnOffRegister(
	"Switches",
	"CommandSwitch",
	"Command switch",
	1,
	true,
)

var CommandSwitchRegisterRO = NewOnOffRegister(
	"Switches",
	"CommandSwitchRO",
	"Command switch",
	1,
	false,
)

var ResetSwitchRegister = NewOnOffRegister(
	"Switches",
	"ResetSwitch",
	"Emergency off and reset switch",
	2,
	true,
)

var ResetSwitchRegisterRO = NewOnOffRegister(
	"Switches",
	"ResetSwitchRO",
	"Emergency off and reset switch",
	2,
	false,
)

var IOAvailableRegister = NewOnOffRegister(
	"Inputs",
	"IOAvailable",
	"I/O controller available",
	10,
	false,
)

var FireDetectedRegister = NewOnOffRegister(
	"Inputs",
	"FireDetected",
	"Fire detected",
	11,
	false,
)

var EngineTempRegister = NewNumberRegister(
	"Inputs",
	"EngineTemp",
	"Engine temperature",
	"°C",
	12,
)

var AuxTemp0Register = NewNumberRegister(
	"Inputs",
	"AuxTemp0",
	"Auxiliary temperature 0",
	"°C",
	13,
)

var AuxTemp1Register = NewNumberRegister(
	"Inputs",
	"AuxTemp1",
	"Auxiliary temperature 1",
	"°C",
	14,
)

var OutputAvailableRegister = NewOnOffRegister(
	"Inputs",
	"OutputAvailable",
	"Output available",
	20,
	false,
)

var U0Register = NewNumberRegister(
	"Inputs",
	"U1",
	"Voltage U1",
	"V",
	21,
)

var U1Register = NewNumberRegister(
	"Inputs",
	"U2",
	"Voltage U2",
	"V",
	22,
)

var U2Register = NewNumberRegister(
	"Inputs",
	"U3",
	"Voltage U3",
	"V",
	23,
)

var P1Register = NewNumberRegister(
	"Inputs",
	"P1",
	"Load P1",
	"W",
	24,
)

var P2Register = NewNumberRegister(
	"Inputs",
	"P2",
	"Load P2",
	"W",
	25,
)

var P3Register = NewNumberRegister(
	"Inputs",
	"P3",
	"Load P3",
	"W",
	26,
)

var FRegister = NewNumberRegister(
	"Inputs",
	"F",
	"Frequency",
	"Hz",
	27,
)

var InputRegisters3P = []dataflow.RegisterStruct{
	IOAvailableRegister,
	FireDetectedRegister,
	EngineTempRegister,
	AuxTemp0Register,
	AuxTemp1Register,
	OutputAvailableRegister,
	U0Register,
	P1Register,
	FRegister,

	// only for 3-phase
	U1Register,
	U2Register,
	P2Register,
	P3Register,
}

var InputRegisters1P = InputRegisters3P[:len(InputRegisters3P)-4]

var StateRegister = dataflow.NewRegisterStruct(
	"State",
	"State",
	"Controller State",
	dataflow.EnumRegister,
	func(m map[genset.StateNode]string) map[int]string {
		enum := make(map[int]string, len(m))
		for k, v := range m {
			enum[int(k)] = v
		}
		return enum
	}(genset.StateNodeMap()),
	"",
	30,
	false,
)

var StateChangedRegister = dataflow.NewRegisterStruct(
	"State",
	"StateChanged",
	"Controller State Changed",
	dataflow.TextRegister,
	nil,
	"",
	31,
	false,
)

var StateRegisters = []dataflow.RegisterStruct{
	StateRegister,
	StateChangedRegister,
}

var IgnitionRegister = NewOnOffRegister(
	"Outputs",
	"Ignition",
	"Ignition",
	40,
	false,
)

var StarterRegister = NewOnOffRegister(
	"Outputs",
	"Starter",
	"Starter",
	41,
	false,
)

var FanRegister = NewOnOffRegister(
	"Outputs",
	"Fan",
	"Fan",
	42,
	false,
)

var PumpRegister = NewOnOffRegister(
	"Outputs",
	"Pump",
	"Pump",
	43,
	false,
)

var LoadRegister = NewOnOffRegister(
	"Outputs",
	"Load",
	"Load",
	44,
	false,
)

var TimeInStateRegister = NewNumberRegister(
	"Outputs",
	"TimeInState",
	"Time in state",
	"s",
	45,
)

var IoCheckRegister = NewOnOffRegister(
	"Outputs",
	"IoCheck",
	"I/O check",
	46,
	false,
)

var OutputCheckRegister = NewOnOffRegister(
	"Outputs",
	"OutputCheck",
	"Output check",
	47,
	false,
)

var OutputRegisters = []dataflow.RegisterStruct{
	IgnitionRegister,
	StarterRegister,
	FanRegister,
	PumpRegister,
	LoadRegister,
	TimeInStateRegister,
	IoCheckRegister,
	OutputCheckRegister,
}
