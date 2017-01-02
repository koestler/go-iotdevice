package bmv

import (
	"github.com/koestler/go-ve-sensor/vedirect"
	"log"
)

type Register struct {
	Name    string
	Address uint16
	Factor  float64
	Unit    string
	Signed  bool
}

type NumericValue struct {
	Name  string
	Value float64
	Unit  string
}

var RegisterList700 = []Register{
	Register{
		Name:    "MainVoltage",
		Address: 0xED8D,
		Factor:  0.01,
		Unit:    "V",
		Signed:  false,
	},
	Register{
		Name:    "Current",
		Address: 0xED8F,
		Factor:  0.1,
		Unit:    "A",
		Signed:  true,
	},
	Register{
		Name:    "Power",
		Address: 0xED8E,
		Factor:  1,
		Unit:    "W",
		Signed:  true,
	},
	Register{
		Name:    "Consumed",
		Address: 0xEEFF,
		Factor:  0.1,
		Unit:    "Ah",
		Signed:  true,
	},
	Register{
		Name:    "StateOfCharge",
		Address: 0x0FFF,
		Factor:  0.01,
		Unit:    "%",
		Signed:  false,
	},
	Register{
		Name:    "TimeToGo",
		Address: 0x0FFE,
		Factor:  1,
		Unit:    "min",
		Signed:  false,
	},
	Register{
		Name:    "Temperature",
		Address: 0xEDEC,
		Factor:  0.01,
		Unit:    "K",
		Signed:  false,
	},
	Register{
		Name:    "DepthOfTheDeepestDischarge",
		Address: 0x0300,
		Factor:  0.1,
		Unit:    "Ah",
		Signed:  true,
	},
	Register{
		Name:    "DepthOfTheLastDischarge",
		Address: 0x0301,
		Factor:  0.1,
		Unit:    "Ah",
		Signed:  true,
	},
	Register{
		Name:    "DepthOfTheAverageDischarge",
		Address: 0x0302,
		Factor:  0.1,
		Unit:    "Ah",
		Signed:  true,
	},
	Register{
		Name:    "NumberOfCycles",
		Address: 0x0303,
		Factor:  1,
		Unit:    "",
		Signed:  false,
	},
	Register{
		Name:    "NumberOfFullDischarges",
		Address: 0x0304,
		Factor:  1,
		Unit:    "",
		Signed:  false,
	},
	Register{
		Name:    "CumulativeAmpHours",
		Address: 0x0305,
		Factor:  0.1,
		Unit:    "Ah",
		Signed:  true,
	},
	Register{
		Name:    "MainVoltageMinimum",
		Address: 0x0306,
		Factor:  0.01,
		Unit:    "V",
		Signed:  false,
	},
	Register{
		Name:    "MainVoltageMaximum",
		Address: 0x0307,
		Factor:  0.01,
		Unit:    "V",
		Signed:  false,
	},
	Register{
		Name:    "DaysSinceFullChrage",
		Address: 0x0308,
		Factor:  1,
		Unit:    "d",
		Signed:  false,
	},
	Register{
		Name:    "NumberOfAutomaticSynchronizations",
		Address: 0x0309,
		Factor:  1,
		Unit:    "",
		Signed:  false,
	},
	Register{
		Name:    "NumberOfLowMainVoltageAlarms",
		Address: 0x030A,
		Factor:  1,
		Unit:    "",
		Signed:  false,
	},
	Register{
		Name:    "NumberOfHighMainVoltageAlarms",
		Address: 0x030B,
		Factor:  1,
		Unit:    "",
		Signed:  false,
	},
	Register{
		Name:    "AmountOfDischargedEnergy",
		Address: 0x0310,
		Factor:  0.01,
		Unit:    "kWh",
		Signed:  false,
	},
	Register{
		Name:    "AmountOfChargedEnergy",
		Address: 0x0311,
		Factor:  0.01,
		Unit:    "kWh",
		Signed:  false,
	},
}

var RegisterList702 = []Register{
	Register{
		Name:    "AuxVoltage",
		Address: 0xED7D,
		Factor:  0.01,
		Unit:    "V",
		Signed:  false,
	},
	Register{
		Name:    "Synchronized",
		Address: 0xEEB6,
		Factor:  1,
		Unit:    "1",
		Signed:  false,
	},
	Register{
		Name:    "MidPointVoltage",
		Address: 0x0382,
		Factor:  0.01,
		Unit:    "V",
		Signed:  false,
	},
	Register{
		Name:    "MidPointVoltageDeviation",
		Address: 0x0383,
		Factor:  0.1,
		Unit:    "%",
		Signed:  true,
	},
	Register{
		Name:    "NumberOfLowAuxVoltageAlarms",
		Address: 0x030C,
		Factor:  1,
		Unit:    "",
		Signed:  false,
	},
	Register{
		Name:    "NumberOfHighAuxVoltageAlarms",
		Address: 0x030D,
		Factor:  1,
		Unit:    "",
		Signed:  false,
	},
	Register{
		Name:    "AuxVoltageMinimum",
		Address: 0x030E,
		Factor:  0.01,
		Unit:    "V",
		Signed:  true,
	},
	Register{
		Name:    "AuxVoltageMaximum",
		Address: 0x030F,
		Factor:  0.01,
		Unit:    "V",
		Signed:  true,
	},
}

func (reg Register) RecvNumeric(vd *vedirect.Vedirect) (result NumericValue, err error) {
	var value float64

	if reg.Signed {
		var intValue int64
		intValue, err = vd.VeCommandGetInt(reg.Address)
		value = float64(intValue)
	} else {
		var intValue uint64
		intValue, err = vd.VeCommandGetUint(reg.Address)
		value = float64(intValue)
	}

	if err != nil {
		log.Printf("bmv.BmvGetResgite failed: %v", err)
		return
	}

	result = NumericValue{
		Name:  reg.Name,
		Value: value * reg.Factor,
		Unit:  reg.Unit,
	}

	return
}
