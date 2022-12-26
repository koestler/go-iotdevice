package teracomDevice

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
		AI1 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Unit        string `xml:"unit"`
			Multiplier  string `xml:"multiplier"`
			Offset      string `xml:"offset"`
			Alarm       string `xml:"alarm"`
			Min         string `xml:"min"`
			Max         string `xml:"max"`
			Hys         string `xml:"hys"`
		} `xml:"AI1"`
		AI2 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Unit        string `xml:"unit"`
			Multiplier  string `xml:"multiplier"`
			Offset      string `xml:"offset"`
			Alarm       string `xml:"alarm"`
			Min         string `xml:"min"`
			Max         string `xml:"max"`
			Hys         string `xml:"hys"`
		} `xml:"AI2"`
		AI3 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Unit        string `xml:"unit"`
			Multiplier  string `xml:"multiplier"`
			Offset      string `xml:"offset"`
			Alarm       string `xml:"alarm"`
			Min         string `xml:"min"`
			Max         string `xml:"max"`
			Hys         string `xml:"hys"`
		} `xml:"AI3"`
		AI4 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Unit        string `xml:"unit"`
			Multiplier  string `xml:"multiplier"`
			Offset      string `xml:"offset"`
			Alarm       string `xml:"alarm"`
			Min         string `xml:"min"`
			Max         string `xml:"max"`
			Hys         string `xml:"hys"`
		} `xml:"AI4"`
	} `xml:"AI"`
	VI struct {
		VI1 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Unit        string `xml:"unit"`
			Multiplier  string `xml:"multiplier"`
			Offset      string `xml:"offset"`
			Alarm       string `xml:"alarm"`
			Min         string `xml:"min"`
			Max         string `xml:"max"`
			Hys         string `xml:"hys"`
		} `xml:"VI1"`
		VI2 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Unit        string `xml:"unit"`
			Multiplier  string `xml:"multiplier"`
			Offset      string `xml:"offset"`
			Alarm       string `xml:"alarm"`
			Min         string `xml:"min"`
			Max         string `xml:"max"`
			Hys         string `xml:"hys"`
		} `xml:"VI2"`
		VI3 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Unit        string `xml:"unit"`
			Multiplier  string `xml:"multiplier"`
			Offset      string `xml:"offset"`
			Alarm       string `xml:"alarm"`
			Min         string `xml:"min"`
			Max         string `xml:"max"`
			Hys         string `xml:"hys"`
		} `xml:"VI3"`
		VI4 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Unit        string `xml:"unit"`
			Multiplier  string `xml:"multiplier"`
			Offset      string `xml:"offset"`
			Alarm       string `xml:"alarm"`
			Min         string `xml:"min"`
			Max         string `xml:"max"`
			Hys         string `xml:"hys"`
		} `xml:"VI4"`
	} `xml:"VI"`
	DI struct {
		DI1 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Valuebin    string `xml:"valuebin"`
			AlarmState  string `xml:"alarmState"`
			Alarm       string `xml:"alarm"`
		} `xml:"DI1"`
		DI2 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Valuebin    string `xml:"valuebin"`
			AlarmState  string `xml:"alarmState"`
			Alarm       string `xml:"alarm"`
		} `xml:"DI2"`
		DI3 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Valuebin    string `xml:"valuebin"`
			AlarmState  string `xml:"alarmState"`
			Alarm       string `xml:"alarm"`
		} `xml:"DI3"`
		DI4 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Valuebin    string `xml:"valuebin"`
			AlarmState  string `xml:"alarmState"`
			Alarm       string `xml:"alarm"`
		} `xml:"DI4"`
	} `xml:"DI"`
	R struct {
		R1 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Valuebin    string `xml:"valuebin"`
			PulseWidth  string `xml:"pulseWidth"`
			Control     string `xml:"control"`
		} `xml:"R1"`
		R2 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Valuebin    string `xml:"valuebin"`
			PulseWidth  string `xml:"pulseWidth"`
			Control     string `xml:"control"`
		} `xml:"R2"`
		R3 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Valuebin    string `xml:"valuebin"`
			PulseWidth  string `xml:"pulseWidth"`
			Control     string `xml:"control"`
		} `xml:"R3"`
		R4 struct {
			Description string `xml:"description"`
			Value       string `xml:"value"`
			Valuebin    string `xml:"valuebin"`
			PulseWidth  string `xml:"pulseWidth"`
			Control     string `xml:"control"`
		} `xml:"R4"`
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

func (c *DeviceStruct) text(category, registerName, description, unit string, value string) {
	if len(value) < 1 {
		return
	}

	register := c.addIgnoreRegister(category, registerName, description, unit, "text")
	c.output <- dataflow.NewTextRegisterValue(c.deviceConfig.Name(), register, value)
}

func (c *DeviceStruct) number(category, registerName, description, unit string, value string) {
	if value == "---" {
		// this is teracom's way of encoding null
		return
	}

	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return
	}

	register := c.addIgnoreRegister(category, registerName, description, unit, "numeric")
	c.output <- dataflow.NewNumericRegisterValue(c.deviceConfig.Name(), register, floatVal)
}

func (c *DeviceStruct) extractRegistersAndValues(s StatusStruct) {
	// sensors
	sensor := func(sIdx int, s sensorStruct) {
		if s.ID != "0000000000000000" {
			item := func(sIdx, vIdx int, s sensorStruct, i sensorValueStruct, multi bool) {
				regName := fmt.Sprintf("S%dV%d", sIdx, vIdx)
				description := s.Description
				if multi {
					description = fmt.Sprintf("%s - %s", description, i.Unit)
				}

				c.number("Sensors", regName, description, i.Unit, i.Value)
				c.text("Alarms", regName+"Alarm", description+" Alarm", "", i.Alarm)

				c.number("Sensor Config", regName+"Min", description+" Min", i.Unit, i.Min)
				c.number("Sensor Config", regName+"Max", description+" Max", i.Unit, i.Max)
				c.number("Sensor Config", regName+"Hys", description+" Hysteresis", i.Unit, i.Max)
			}

			multi := s.Item2.Value != "---"
			item(sIdx, 1, s, s.Item1, multi)
			if multi {
				item(sIdx, 2, s, s.Item2, true)
			}

			regName := fmt.Sprintf("S%d", sIdx)
			c.text("Sensor Config", regName+"Id", s.Description+" Id", "", s.ID)
		}
	}
	sensor(1, s.S.S1)
	sensor(2, s.S.S2)
	sensor(3, s.S.S3)
	sensor(4, s.S.S4)
	sensor(5, s.S.S5)
	sensor(6, s.S.S6)
	sensor(7, s.S.S7)
	sensor(8, s.S.S8)

	// device info
	cat := "Device Info"
	c.text(cat, "DeviceName", "Device Name", "", s.DeviceInfo.DeviceName)
	c.text(cat, "HostName", "Host Name", "", s.DeviceInfo.HostName)
	c.text(cat, "Id", "Id", "", s.DeviceInfo.ID)
	c.text(cat, "FWVer", "Firmware Vesion", "", s.DeviceInfo.FwVer)
	c.text(cat, "SysContact", "System Contact", "", s.DeviceInfo.SysContact)
	c.text(cat, "SysName", "System Name", "", s.DeviceInfo.SysName)
	c.text(cat, "SysLocation", "System Location", "", s.DeviceInfo.SysLocation)
}
