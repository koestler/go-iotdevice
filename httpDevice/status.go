package httpDevice

import (
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"strconv"
)

type sensorValueStruct struct {
	Value string `xml:"value"`
	Unit  string `xml:"unit"`
	Alarm string `xml:"alarm"`
	Min   string `xml:"min"`
	Max   string `xml:"max"`
	Hys   string `xml:"hys"`
}

type sensorStruct struct {
	Description string            `xml:"description"`
	ID          string            `xml:"id"`
	Item1       sensorValueStruct `xml:"item1"`
	Item2       sensorValueStruct `xml:"item2"`
}

type analogStruct struct {
	Description string `xml:"description"`
	Value       string `xml:"value"`
	Unit        string `xml:"unit"`
	Multiplier  string `xml:"multiplier"`
	Offset      string `xml:"offset"`
	Alarm       string `xml:"alarm"`
	Min         string `xml:"min"`
	Max         string `xml:"max"`
	Hys         string `xml:"hys"`
}

type digitalStruct struct {
	Description string `xml:"description"`
	Value       string `xml:"value"`
	Valuebin    string `xml:"valuebin"`
	AlarmState  string `xml:"alarmState"`
	Alarm       string `xml:"alarm"`
}

type relayStruct struct {
	Description string `xml:"description"`
	Value       string `xml:"value"`
	Valuebin    string `xml:"valuebin"`
	PulseWidth  string `xml:"pulseWidth"`
	Control     string `xml:"control"`
}

type StatusStruct struct {
	DeviceInfo struct {
		DeviceName  string `xml:"DeviceName"`
		HostName    string `xml:"HostName"`
		ID          string `xml:"ID"`
		FwVer       string `xml:"FwVer"`
		SysContact  string `xml:"SysContact"`
		SysName     string `xml:"SysName"`
		SysLocation string `xml:"SysLocation"`
	} `xml:"DeviceInfo"`
	S struct {
		S1 sensorStruct `xml:"S1"`
		S2 sensorStruct `xml:"S2"`
		S3 sensorStruct `xml:"S3"`
		S4 sensorStruct `xml:"S4"`
		S5 sensorStruct `xml:"S5"`
		S6 sensorStruct `xml:"S6"`
		S7 sensorStruct `xml:"S7"`
		S8 sensorStruct `xml:"S8"`
	}
	AI struct {
		AI1 analogStruct `xml:"AI1"`
		AI2 analogStruct `xml:"AI2"`
		AI3 analogStruct `xml:"AI3"`
		AI4 analogStruct `xml:"AI4"`
	} `xml:"AI"`
	VI struct {
		VI1 analogStruct `xml:"VI1"`
		VI2 analogStruct `xml:"VI2"`
		VI3 analogStruct `xml:"VI3"`
		VI4 analogStruct `xml:"VI4"`
	} `xml:"VI"`
	DI struct {
		DI1 digitalStruct `xml:"DI1"`
		DI2 digitalStruct `xml:"DI2"`
		DI3 digitalStruct `xml:"DI3"`
		DI4 digitalStruct `xml:"DI4"`
	} `xml:"DI"`
	R struct {
		R1 relayStruct `xml:"R1"`
		R2 relayStruct `xml:"R2"`
		R3 relayStruct `xml:"R3"`
		R4 relayStruct `xml:"R4"`
	} `xml:"R"`
	HTTPPush struct {
		Key        string `xml:"Key"`
		PushPeriod string `xml:"PushPeriod"`
	} `xml:"HTTPPush"`
	Hwerr   string `xml:"hwerr"`
	Alarmed string `xml:"Alarmed"`
	Scannig string `xml:"Scannig"`
	Time    struct {
		Date string `xml:"Date"`
		Time string `xml:"Time"`
	} `xml:"Time"`
}

func (c *DeviceStruct) text(category, registerName, description, value string) {
	if len(value) < 1 {
		return
	}

	register := c.addIgnoreRegister(category, registerName, description, "", "text")
	c.output <- dataflow.NewTextRegisterValue(c.deviceConfig.Name(), register, value)
}

func (c *DeviceStruct) number(category, registerName, description, unit string, value string) {
	if value == "---" {
		// this is teracom's way of encoding null
		return
	}

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return
	}

	register := c.addIgnoreRegister(category, registerName, description, unit, "numeric")
	c.output <- dataflow.NewNumericRegisterValue(c.deviceConfig.Name(), register, floatValue)
}

func (c *DeviceStruct) boolean(category, registerName, description string, value string) {
	numericValue := func(value string) float64 {
		if value == "ON" {
			return 1
		}

		if value == "OPEN" {
			return 1
		}

		return 0
	}(value)

	register := c.addIgnoreRegister(category, registerName, description, "", "numeric")
	c.output <- dataflow.NewNumericRegisterValue(c.deviceConfig.Name(), register, numericValue)
}

func getCategorySort(category string) int {
	switch category {
	case "Sensors":
		return 0
	case "Analog Inputs":
		return 1
	case "Virtual Inputs":
		return 2
	case "Digital Inputs":
		return 3
	case "Relays":
		return 4
	case "Alarms":
		return 5
	case "General":
		return 6
	case "Device Info":
		return 7
	case "Settings":
		return 8
	default:
		panic("unknown category: " + category)
	}
}

func (c *DeviceStruct) extractRegistersAndValues(s StatusStruct) {
	// device info
	cat := "Device Info"
	c.text(cat, "DeviceName", "Device Name", s.DeviceInfo.DeviceName)
	c.text(cat, "HostName", "Host Name", s.DeviceInfo.HostName)
	c.text(cat, "Id", "Id", s.DeviceInfo.ID)
	c.text(cat, "FWVer", "Firmware Vesion", s.DeviceInfo.FwVer)
	c.text(cat, "SysContact", "System Contact", s.DeviceInfo.SysContact)
	c.text(cat, "SysName", "System Name", s.DeviceInfo.SysName)
	c.text(cat, "SysLocation", "System Location", s.DeviceInfo.SysLocation)

	// general
	cat = "General"
	c.text(cat, "Hwerr", "Hardware Error", s.Hwerr)
	c.text(cat, "Alarmed", "Alarmed", s.Alarmed)
	c.text(cat, "Date", "Date", s.Time.Date)
	c.text(cat, "Time", "Time", s.Time.Time)

	// sensors
	sensor := func(sIdx int, s sensorStruct) {
		if s.ID == "0000000000000000" {
			return
		}
		item := func(sIdx, vIdx int, s sensorStruct, i sensorValueStruct, multi bool) {
			regName := fmt.Sprintf("S%dV%d", sIdx, vIdx)
			desc := s.Description
			if multi {
				desc = fmt.Sprintf("%s - %s", desc, i.Unit)
			}

			c.number("Sensors", regName, desc, i.Unit, i.Value)
			c.number("Alarms", regName+"Alarm", desc+" Alarm", "", i.Alarm)

			c.number("Settings", regName+"Min", desc+" Min", i.Unit, i.Min)
			c.number("Settings", regName+"Max", desc+" Max", i.Unit, i.Max)
			c.number("Settings", regName+"Hys", desc+" Hysteresis", i.Unit, i.Hys)
		}

		multi := s.Item2.Value != "---"
		item(sIdx, 1, s, s.Item1, multi)
		if multi {
			item(sIdx, 2, s, s.Item2, true)
		}

		regName := fmt.Sprintf("S%d", sIdx)
		c.text("Settings", regName+"Id", s.Description+" Id", s.ID)
	}
	sensor(1, s.S.S1)
	sensor(2, s.S.S2)
	sensor(3, s.S.S3)
	sensor(4, s.S.S4)
	sensor(5, s.S.S5)
	sensor(6, s.S.S6)
	sensor(7, s.S.S7)
	sensor(8, s.S.S8)

	// analog inputs
	analog := func(regNamePrefix, valueCat string, sIdx int, a analogStruct) {
		if a.Value == "---" {
			return
		}

		regName := fmt.Sprintf("%s%d", regNamePrefix, sIdx)
		desc := a.Description

		c.number(valueCat, regName, desc, a.Unit, a.Value)
		c.number("Alarms", regName+"Alarm", desc+" Alarm", "", a.Alarm)

		c.number("Settings", regName+"Min", desc+" Min", a.Unit, a.Min)
		c.number("Settings", regName+"Max", desc+" Max", a.Unit, a.Max)
		c.number("Settings", regName+"Hys", desc+" Hysteresis", a.Unit, a.Hys)
		c.number("Settings", regName+"Offset", desc+" Offset", a.Unit, a.Offset)
		c.number("Settings", regName+"Multiplier", desc+" Multiplier", a.Unit, a.Multiplier)
	}
	analog("AI", "Analog Inputs", 1, s.AI.AI1)
	analog("AI", "Analog Inputs", 2, s.AI.AI2)
	analog("AI", "Analog Inputs", 3, s.AI.AI3)
	analog("AI", "Analog Inputs", 4, s.AI.AI4)

	// virtual inputs
	analog("VI", "Virtual Inputs", 1, s.VI.VI1)
	analog("VI", "Virtual Inputs", 2, s.VI.VI2)
	analog("VI", "Virtual Inputs", 3, s.VI.VI3)
	analog("VI", "Virtual Inputs", 4, s.VI.VI4)

	// digital inputs
	digital := func(sIdx int, a digitalStruct) {
		regName := fmt.Sprintf("DI%d", sIdx)
		desc := a.Description

		c.boolean("Digital Inputs", regName, desc, a.Value)
		c.number("Alarms", regName+"Alarm", desc+" Alarm", "", a.Alarm)
	}
	digital(1, s.DI.DI1)
	digital(2, s.DI.DI2)
	digital(3, s.DI.DI3)
	digital(4, s.DI.DI4)

	// relays
	relay := func(sIdx int, r relayStruct) {
		regName := fmt.Sprintf("R%d", sIdx)
		desc := r.Description

		c.boolean("Relays", regName, desc, r.Value)
		if r.Control != "0" {
			c.text("Relays", regName+"Control", desc+" is controlled by", r.Control)
		}
	}
	relay(1, s.R.R1)
	relay(2, s.R.R2)
	relay(3, s.R.R3)
	relay(4, s.R.R4)
}
