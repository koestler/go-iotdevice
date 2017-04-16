package bmv

import (
	"github.com/koestler/go-ve-sensor/vedirect"
	"log"
)

type NumericValues map[string]NumericValue

type NumericValue struct {
	Value float64
	Unit  string
}

type Registers map[string]Register

type Register struct {
	Address uint16
	Factor  float64
	Unit    string
	Signed  bool
}

var RegisterList700 = Registers{
	"MainVoltage": Register{
		Address: 0xED8D,
		Factor:  0.01,
		Unit:    "V",
		Signed:  false,
	},
	"Current": Register{
		Address: 0xED8F,
		Factor:  0.1,
		Unit:    "A",
		Signed:  true,
	},
	"Power": Register{
		Address: 0xED8E,
		Factor:  1,
		Unit:    "W",
		Signed:  true,
	},
	"Consumed": Register{
		Address: 0xEEFF,
		Factor:  0.1,
		Unit:    "Ah",
		Signed:  true,
	},
	"StateOfCharge": Register{
		Address: 0x0FFF,
		Factor:  0.01,
		Unit:    "%",
		Signed:  false,
	},
	"TimeToGo": Register{
		Address: 0x0FFE,
		Factor:  1,
		Unit:    "min",
		Signed:  false,
	},
	"Temperature": Register{
		Address: 0xEDEC,
		Factor:  0.01,
		Unit:    "K",
		Signed:  false,
	},
	"DepthOfTheDeepestDischarge": Register{
		Address: 0x0300,
		Factor:  0.1,
		Unit:    "Ah",
		Signed:  true,
	},
	"DepthOfTheLastDischarge": Register{
		Address: 0x0301,
		Factor:  0.1,
		Unit:    "Ah",
		Signed:  true,
	},
	"DepthOfTheAverageDischarge": Register{
		Address: 0x0302,
		Factor:  0.1,
		Unit:    "Ah",
		Signed:  true,
	},
	"NumberOfCycles": Register{
		Address: 0x0303,
		Factor:  1,
		Unit:    "",
		Signed:  false,
	},
	"NumberOfFullDischarges": Register{
		Address: 0x0304,
		Factor:  1,
		Unit:    "",
		Signed:  false,
	},
	"CumulativeAmpHours": Register{
		Address: 0x0305,
		Factor:  0.1,
		Unit:    "Ah",
		Signed:  true,
	},
	"MainVoltageMinimum": Register{
		Address: 0x0306,
		Factor:  0.01,
		Unit:    "V",
		Signed:  false,
	},
	"MainVoltageMaximum": Register{
		Address: 0x0307,
		Factor:  0.01,
		Unit:    "V",
		Signed:  false,
	},
	"DaysSinceFullCharge": Register{
		Address: 0x0308,
		Factor:  1,
		Unit:    "d",
		Signed:  false,
	},
	"NumberOfAutomaticSynchronizations": Register{
		Address: 0x0309,
		Factor:  1,
		Unit:    "",
		Signed:  false,
	},
	"NumberOfLowMainVoltageAlarms": Register{
		Address: 0x030A,
		Factor:  1,
		Unit:    "",
		Signed:  false,
	},
	"NumberOfHighMainVoltageAlarms": Register{
		Address: 0x030B,
		Factor:  1,
		Unit:    "",
		Signed:  false,
	},
	"AmountOfDischargedEnergy": Register{
		Address: 0x0310,
		Factor:  0.01,
		Unit:    "kWh",
		Signed:  false,
	},
	"AmountOfChargedEnergy": Register{
		Address: 0x0311,
		Factor:  0.01,
		Unit:    "kWh",
		Signed:  false,
	},
}

var RegisterList702 = Registers{
	"AuxVoltage": Register{
		Address: 0xED7D,
		Factor:  0.01,
		Unit:    "V",
		Signed:  false,
	},
	"Synchronized": Register{
		Address: 0xEEB6,
		Factor:  1,
		Unit:    "1",
		Signed:  false,
	},
	"MidPointVoltage": Register{
		Address: 0x0382,
		Factor:  0.01,
		Unit:    "V",
		Signed:  false,
	},
	"MidPointVoltageDeviation": Register{
		Address: 0x0383,
		Factor:  0.1,
		Unit:    "%",
		Signed:  true,
	},
	"NumberOfLowAuxVoltageAlarms": Register{
		Address: 0x030C,
		Factor:  1,
		Unit:    "",
		Signed:  false,
	},
	"NumberOfHighAuxVoltageAlarms": Register{
		Address: 0x030D,
		Factor:  1,
		Unit:    "",
		Signed:  false,
	},
	"AuxVoltageMinimum": Register{
		Address: 0x030E,
		Factor:  0.01,
		Unit:    "V",
		Signed:  true,
	},
	"AuxVoltageMaximum": Register{
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
		log.Printf("bmv.RecvNumeric failed: %v", err)
		return
	}

	result = NumericValue{
		Value: value * reg.Factor,
		Unit:  reg.Unit,
	}

	return
}
