package vedevices

// todo: understand the enums
var RegisterListSolarBatterySettings = Registers{
	/*
		Register{
			Name:          "AutomaticEqualizationMode",
			Address:       0xEDFD,
			Factor:        1,
			Unit:          "every X day",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "BatteryBulkTimeLimit",
			Address:       0xEDFC,
			Factor:        0.01,
			Unit:          "h",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "BatteryAbsorptionTimeLimit",
			Address:       0xEDFB,
			Factor:        0.01,
			Unit:          "h",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "BatteryAbsorptionVoltage",
			Address:       0xEDF7,
			Factor:        0.01,
			Unit:          "V",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "BatteryFloatVoltage",
			Address:       0xEDF6,
			Factor:        0.01,
			Unit:          "V",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "BatteryEqualisationVoltage",
			Address:       0xEDF4,
			Factor:        0.01,
			Unit:          "V",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "BatteryTempCompensation",
			Address:       0xEDF2,
			Factor:        0.01,
			Unit:          "mV/K",
			Signed:        true,
			RoundDecimals: 0,
		},
		Register{
			Name: "BatteryType",
			// todo: this is an enum
			Address:       0xEDF1,
			Factor:        1,
			Unit:          "",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "BatteryMaximumCurrent",
			Address:       0xEDF0,
			Factor:        0.1,
			Unit:          "A",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "BatteryVoltage",
			Address:       0xEDEF,
			Factor:        1,
			Unit:          "V",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "BatteryVoltageSetting",
			Address:       0xEDEA,
			Factor:        1,
			Unit:          "V",
			Signed:        false,
			RoundDecimals: 0,
		},
		// BmsPresent 0xEDE8 skipped, Introduced in firmware version 1.17

	*/
}

var RegisterListSolarChargerData = Registers{
	/*
		Register{
			Name:          "ChargerMaximumCurrent",
			Address:       0xEDDF,
			Factor:        0.01,
			Unit:          "A",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "SystemYield",
			Address:       0xEDDD,
			Factor:        0.01,
			Unit:          "kWh",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "UserYield",
			Address:       0xEDDC,
			Factor:        0.01,
			Unit:          "kWh",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "ChargerInternalTemperature",
			Address:       0xEDDB,
			Factor:        0.01,
			Unit:          "C",
			Signed:        true,
			RoundDecimals: 2,
		},
		Register{
			Name:          "ChargerErrorCode",
			Address:       0xEDDA,
			Factor:        1,
			Unit:          "",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "ChargerCurrent",
			Address:       0xEDD7,
			Factor:        0.1,
			Unit:          "A",
			Signed:        false,
			RoundDecimals: 1,
		},
		Register{
			Name:          "ChargerVoltage",
			Address:       0xEDD5,
			Factor:        0.01,
			Unit:          "V",
			Signed:        false,
			RoundDecimals: 2,
		},

		Register{
			Name:          "AdditionalChargerStateInfo",
			Address:       0xEDD4,
			Factor:        1,
			Unit:          "",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "YieldToday",
			Address:       0xEDD3,
			Factor:        0.01,
			Unit:          "kWh",
			Signed:        false,
			RoundDecimals: 2,
		},
		Register{
			Name:          "MaximumPowerToday",
			Address:       0xEDD2,
			Factor:        1,
			Unit:          "W",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "YieldYesterday",
			Address:       0xEDD1,
			Factor:        0.01,
			Unit:          "kWh",
			Signed:        false,
			RoundDecimals: 2,
		},
		Register{
			Name:          "MaximumPowerYesterday",
			Address:       0xEDD0,
			Factor:        1,
			Unit:          "W",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:    "VoltageSettingsRange",
			Address: 0xEDCE,
			// todo
			// Note 5: The low-byte is the minimum system voltage and the high byte is maximum system voltage
			// (both in 1V units). Available in firmware version 1.16 and higher.
			Factor:        1,
			Unit:          "",
			Signed:        false,
			RoundDecimals: 0,
		},
		Register{
			Name:          "HistoryVersion",
			Address:       0xEDCD,
			Factor:        1,
			Unit:          "",
			Signed:        false,
			RoundDecimals: 0,
		},
		// StreetlightVersion 0xEDCC skipped; only available in firmware version 1.16 and higher
	*/
}

var RegisterListSolarPanelData = Registers{
	/*
		Register{
			Name:          "PanelPower",
			Address:       0xEDBC,
			Factor:        0.01,
			Unit:          "W",
			Signed:        false,
			RoundDecimals: 2,
		},
		Register{
			Name:          "PanelVoltage",
			Address:       0xEDBB,
			Factor:        0.01,
			Unit:          "V",
			Signed:        false,
			RoundDecimals: 2,
		},
				Register{
			        Name: "PanelCurrent",
					Address:       0xEDBD,
					Factor:        0.1,
					Unit:          "A",
					Signed:        false,
					RoundDecimals: 1,
				},
		Register{
			Name:          "PanelMaximumVoltage",
			Address:       0xEDB8,
			Factor:        0.01,
			Unit:          "V",
			Signed:        false,
			RoundDecimals: 0,
		},

	*/
}

var RegisterListSolar = mergeRegisters(
	RegisterListSolarBatterySettings,
	RegisterListSolarChargerData,
	RegisterListSolarPanelData,
)
