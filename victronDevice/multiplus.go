package victronDevice

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/mk2driver"
)

var RegisterListMultiplus = map[string]dataflow.RegisterStruct{
	"BatteryVoltage": dataflow.NewRegisterStruct(
		"Battery",
		"BatteryVoltage",
		"Battery Voltage",
		dataflow.NumberRegister,
		nil,
		"V",
		100,
		false,
	),
	"BatteryCurrent": dataflow.NewRegisterStruct(
		"Battery",
		"BatteryCurrent",
		"Battery Current",
		dataflow.NumberRegister,
		nil,
		"A",
		101,
		false,
	),

	"InputVoltage": dataflow.NewRegisterStruct(
		"Input",
		"InputVoltage",
		"Input Voltage",
		dataflow.NumberRegister,
		nil,
		"V",
		200,
		false,
	),
	"InputCurrent": dataflow.NewRegisterStruct(
		"Input",
		"InputCurrent",
		"Input Current",
		dataflow.NumberRegister,
		nil,
		"A",
		201,
		false,
	),
	"InputFrequency": dataflow.NewRegisterStruct(
		"Input",
		"InputFrequency",
		"Input Frequency",
		dataflow.NumberRegister,
		nil,
		"Hz",
		202,
		false,
	),

	"OutputVoltage": dataflow.NewRegisterStruct(
		"Output",
		"OutputVoltage",
		"Output Voltage",
		dataflow.NumberRegister,
		nil,
		"V",
		300,
		false,
	),
	"OutputCurrent": dataflow.NewRegisterStruct(
		"Output",
		"OutputCurrent",
		"Output Current",
		dataflow.NumberRegister,
		nil,
		"A",
		301,
		false,
	),
	"OutputFrequency": dataflow.NewRegisterStruct(
		"Output",
		"OutputFrequency",
		"Output Frequency",
		dataflow.NumberRegister,
		nil,
		"Hz",
		302,
		false,
	),

	"Mains": dataflow.NewRegisterStruct(
		"Charger",
		"Mains",
		"Mains",
		dataflow.EnumRegister,
		onOffEnum,
		"",
		400,
		false,
	),
	"ChargerMode": dataflow.NewRegisterStruct(
		"Charger",
		"ChargerMode",
		"Charger Mode",
		dataflow.EnumRegister,
		chargerModeEnum,
		"",
		401,
		false,
	),

	"Inverter": dataflow.NewRegisterStruct(
		"Inverter",
		"Inverter",
		"Inverter",
		dataflow.EnumRegister,
		onOffEnum,
		"",
		500,
		false,
	),
	"Overload": dataflow.NewRegisterStruct(
		"Inverter",
		"Overload",
		"Overload",
		dataflow.EnumRegister,
		faultEnum,
		"",
		501,
		false,
	),
	"LowBattery": dataflow.NewRegisterStruct(
		"Inverter",
		"LowBattery",
		"Low Battery",
		dataflow.EnumRegister,
		faultEnum,
		"",
		502,
		false,
	),
	"Temperature": dataflow.NewRegisterStruct(
		"Inverter",
		"Temperature",
		"Temperature",
		dataflow.EnumRegister,
		faultEnum,
		"",
		503,
		false,
	),

	"Version": dataflow.NewRegisterStruct(
		"Product",
		"Version",
		"Version",
		dataflow.TextRegister,
		nil,
		"",
		900,
		false,
	),
}

var onOffEnum = map[int]string{
	0: "Off",
	1: "On",
}

func ledStateToOnOffEnum(ledState mk2driver.LEDstate) int {
	if ledState != mk2driver.LedOff {
		return 1
	}
	return 0
}

var chargerModeEnum = map[int]string{
	0: "Off",
	1: "Bulk",
	2: "Absorption",
	3: "Float",
}

func ledStateToChargerModeEnum(leds map[mk2driver.Led]mk2driver.LEDstate) int {
	if leds[mk2driver.LedFloat] != mk2driver.LedOff {
		return 3
	}
	if leds[mk2driver.LedAbsorption] != mk2driver.LedOff {
		return 2
	}
	if leds[mk2driver.LedBulk] != mk2driver.LedOff {
		return 1
	}
	return 0
}

var faultEnum = map[int]string{
	0: "Ok",
	1: "Warning",
	2: "Error",
}

func ledStateToFaultEnum(ledState mk2driver.LEDstate) int {
	switch ledState {
	case mk2driver.LedOff:
		return 0
	case mk2driver.LedBlink:
		return 1
	case mk2driver.LedOn:
		return 2
	default:
		return 0

	}
}
