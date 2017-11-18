package vedevices

var RegisterListSolarBatterySettings = Registers{
	"AutomaticEqualizationMode": Register{
		Address:       0xEDFD,
		Factor:        1,
		Unit:          "every X day",
		Signed:        false,
		RoundDecimals: 0,
	},
	"BatteryBulkTimeLimit": Register{
		Address:       0xEDFC,
		Factor:        0.01,
		Unit:          "h",
		Signed:        false,
		RoundDecimals: 0,
	},
	"BatteryAbsorptionTimeLimit": Register{
		Address:       0xEDFB,
		Factor:        0.01,
		Unit:          "h",
		Signed:        false,
		RoundDecimals: 0,
	},
	"BatteryAbsorptionVoltage": Register{
		Address:       0xEDF7,
		Factor:        0.01,
		Unit:          "V",
		Signed:        false,
		RoundDecimals: 0,
	},
	"BatteryFloatVoltage": Register{
		Address:       0xEDF6,
		Factor:        0.01,
		Unit:          "V",
		Signed:        false,
		RoundDecimals: 0,
	},
	"BatteryEqualisationVoltage": Register{
		Address:       0xEDF4,
		Factor:        0.01,
		Unit:          "V",
		Signed:        false,
		RoundDecimals: 0,
	},
	"BatteryTempCompensation": Register{
		Address:       0xEDF2,
		Factor:        0.01,
		Unit:          "mV/K",
		Signed:        true,
		RoundDecimals: 0,
	},
	"BatteryType": Register{
		// todo: this is an enum
		Address:       0xEDF1,
		Factor:        1,
		Unit:          "",
		Signed:        false,
		RoundDecimals: 0,
	},
	"BatteryMaximumCurrent": Register{
		Address:       0xEDF0,
		Factor:        0.1,
		Unit:          "A",
		Signed:        false,
		RoundDecimals: 0,
	},
	"BatteryVoltage": Register{
		Address:       0xEDEF,
		Factor:        1,
		Unit:          "V",
		Signed:        false,
		RoundDecimals: 0,
	},
	"BatteryVoltageSetting": Register{
		Address:       0xEDEA,
		Factor:        1,
		Unit:          "V",
		Signed:        false,
		RoundDecimals: 0,
	},
	"BmsPresent": Register{
		Address:       0xEDE8,
		Factor:        1,
		Unit:          "",
		Signed:        false,
		RoundDecimals: 0,
	},
}

var RegisterListSolarChargerData = Registers{
	"ChargerMaximumCurrent": Register{
		Address:       0xEDDF,
		Factor:        0.01,
		Unit:          "A",
		Signed:        false,
		RoundDecimals: 0,
	},
	"SystemYield": Register{
		Address:       0xEDDD,
		Factor:        0.01,
		Unit:          "kWh",
		Signed:        false,
		RoundDecimals: 0,
	},
	"UserYield": Register{
		Address:       0xEDDC,
		Factor:        0.01,
		Unit:          "kWh",
		Signed:        false,
		RoundDecimals: 0,
	},
	"ChargerInternalTemperature": Register{
		Address:       0xEDDB,
		Factor:        0.01,
		Unit:          "C",
		Signed:        true,
		RoundDecimals: 2,
	},
	"ChargerErrorCode": Register{
		Address:       0xEDDA,
		Factor:        1,
		Unit:          "",
		Signed:        false,
		RoundDecimals: 0,
	},
	"ChargerCurrent": Register{
		Address:       0xEDD7,
		Factor:        0.1,
		Unit:          "A",
		Signed:        false,
		RoundDecimals: 1,
	},
	"ChargerVoltage": Register{
		Address:       0xEDD5,
		Factor:        0.01,
		Unit:          "V",
		Signed:        false,
		RoundDecimals: 2,
	},

	"AdditionalChargerStateInfo": Register{
		Address:       0xEDD4,
		Factor:        1,
		Unit:          "",
		Signed:        false,
		RoundDecimals: 0,
	},
	"YieldToday": Register{
		Address:       0xEDD3,
		Factor:        0.01,
		Unit:          "kWh",
		Signed:        false,
		RoundDecimals: 2,
	},
	"MaximumPowerToday": Register{
		Address:       0xEDD2,
		Factor:        1,
		Unit:          "W",
		Signed:        false,
		RoundDecimals: 0,
	},
	"YieldYesterday": Register{
		Address:       0xEDD1,
		Factor:        0.01,
		Unit:          "kWh",
		Signed:        false,
		RoundDecimals: 2,
	},
	"MaximumPowerYesterday": Register{
		Address:       0xEDD0,
		Factor:        1,
		Unit:          "W",
		Signed:        false,
		RoundDecimals: 0,
	},
	"Voltage settings range": Register{
		Address: 0xEDCE,
		// todo
		// Note 5: The low-byte is the minimum system voltage and the high byte is maximum system voltage
		// (both in 1V units). Available in firmware version 1.16 and higher.
		Factor:        1,
		Unit:          "",
		Signed:        false,
		RoundDecimals: 0,
	},
	"HistoryVersion": Register{
		Address:       0xEDCD,
		Factor:        1,
		Unit:          "",
		Signed:        false,
		RoundDecimals: 0,
	},
	"StreetlightVersion": Register{
		Address:       0xEDCC,
		Factor:        1,
		Unit:          "",
		Signed:        false,
		RoundDecimals: 0,
	},
}

var RegisterListSolarPanelData = Registers{
	"PanelPower": Register{
		Address:       0xEDBC,
		Factor:        0.01,
		Unit:          "W",
		Signed:        false,
		RoundDecimals: 2,
	},
	"PanelVoltage": Register{
		Address:       0xEDBB,
		Factor:        0.01,
		Unit:          "V",
		Signed:        false,
		RoundDecimals: 2,
	},
	"PanelCurrent": Register{
		Address:       0xEDBD,
		Factor:        0.1,
		Unit:          "A",
		Signed:        false,
		RoundDecimals: 1,
	},
	"PanelMaximumVoltage": Register{
		Address:       0xEDB8,
		Factor:        0.01,
		Unit:          "V",
		Signed:        false,
		RoundDecimals: 0,
	},
}

var RegisterListSolar = mergeRegisters(
	RegisterListSolarChargerData,
	RegisterListSolarPanelData,
)
