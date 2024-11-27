package gensetDevice

import "github.com/koestler/go-iotdevice/v3/dataflow"

type Register struct {
	dataflow.RegisterStruct
}

var OnOffEnum = map[int]string{
	0: "Off",
	1: "On",
}

func NewOnOffRegister(
	category, name, description string,
	sort int,
	writable bool,
) Register {
	return Register{
		RegisterStruct: dataflow.NewRegisterStruct(
			category, name, description,
			dataflow.EnumRegister, OnOffEnum, "", sort, writable,
		),
	}
}

func addToRegisterDb(rdb *dataflow.RegisterDb, registers []Register) {
	dataflowRegisters := make([]dataflow.RegisterStruct, len(registers))
	for i, r := range registers {
		dataflowRegisters[i] = r.RegisterStruct
	}
	rdb.AddStruct(dataflowRegisters...)
}

var InputRegisters = []Register{
	NewOnOffRegister(
		"Switches",
		"MasterSwitch",
		"Master switch",
		0,
		true,
	),
	NewOnOffRegister(
		"Switches",
		"ResetSwitch",
		"Emergency off and reset switch",
		1,
		true,
	),
	NewOnOffRegister(
		"I/O controller",
		"IOAvailable",
		"I/O controller available",
		10,
		false,
	),
}

var OutputRegisters = []Register{
	NewOnOffRegister(
		"Outputs",
		"Ignition",
		"Ignition",
		100,
		false,
	), NewOnOffRegister(
		"Outputs",
		"Starter",
		"Starter",
		101,
		false,
	), NewOnOffRegister(
		"Outputs",
		"Fan",
		"Fan",
		102,
		false,
	), NewOnOffRegister(
		"Outputs",
		"Output",
		"Output Relay",
		103,
		false,
	),
}
