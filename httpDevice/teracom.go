package httpDevice

import (
	"encoding/xml"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type TeracomDevice struct {
	ds *DeviceStruct
}

func (c *TeracomDevice) GetPath() string {
	return "status.xml"
}

func (c *TeracomDevice) HandleResponse(body []byte) error {
	var status teracomStatusStruct
	if err := xml.Unmarshal(body, &status); err != nil {
		return fmt.Errorf("cannot parse xml: %s", err)
	}
	c.extractRegistersAndValues(status)

	return nil
}

func (c *TeracomDevice) GetCategorySort(category string) int {
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

func (c *TeracomDevice) ControlValueRequest(value dataflow.Value) (*http.Request, onControlSuccess, error) {
	if value.Register().RegisterType() != dataflow.EnumRegister {
		return nil, nil, fmt.Errorf("only enum implemented")
	}

	enum := value.(dataflow.EnumRegisterValue)

	// control relays
	var cmd string
	switch v := enum.Value(); v {
	case "ON":
		cmd = "ron"
	case "OFF":
		cmd = "rof"
	case "in pulse":
		cmd = "rpl"
	default:
		return nil, nil, fmt.Errorf("unsupported value=%s", v)
	}

	var param string
	switch name := value.Register().Name(); name {
	case "R1":
		param = "1"
	case "R2":
		param = "2"
	case "R3":
		param = "4"
	case "R4":
		param = "8"
	default:
		return nil, nil, fmt.Errorf("unsupported register Name=%s", name)
	}

	values := url.Values{}
	values.Set(cmd, param)

	onSuccess := func() {
		c.relay(value.Register().Name(), value.Register().Description(), enum.Value(), value.Register().Controllable())
	}

	req, err := http.NewRequest("POST", "/monitor/monitor.htm", strings.NewReader(values.Encode()))
	return req, onSuccess, err
}

type teracomSensorValueStruct struct {
	Value string `xml:"value"`
	Unit  string `xml:"unit"`
	Alarm string `xml:"alarm"`
	Min   string `xml:"min"`
	Max   string `xml:"max"`
	Hys   string `xml:"hys"`
}

type teracomSensorStruct struct {
	Description string                   `xml:"description"`
	ID          string                   `xml:"id"`
	Item1       teracomSensorValueStruct `xml:"item1"`
	Item2       teracomSensorValueStruct `xml:"item2"`
}

type teracomAnalogStruct struct {
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

type teracomDigitalStruct struct {
	Description string `xml:"description"`
	Value       string `xml:"value"`
	Valuebin    string `xml:"valuebin"`
	AlarmState  string `xml:"alarmState"`
	Alarm       string `xml:"alarm"`
}

type teracomRelayStruct struct {
	Description string `xml:"description"`
	Value       string `xml:"value"`
	Valuebin    string `xml:"valuebin"`
	PulseWidth  string `xml:"pulseWidth"`
	Control     string `xml:"control"`
}

type teracomStatusStruct struct {
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
		S1 teracomSensorStruct `xml:"S1"`
		S2 teracomSensorStruct `xml:"S2"`
		S3 teracomSensorStruct `xml:"S3"`
		S4 teracomSensorStruct `xml:"S4"`
		S5 teracomSensorStruct `xml:"S5"`
		S6 teracomSensorStruct `xml:"S6"`
		S7 teracomSensorStruct `xml:"S7"`
		S8 teracomSensorStruct `xml:"S8"`
	}
	AI struct {
		AI1 teracomAnalogStruct `xml:"AI1"`
		AI2 teracomAnalogStruct `xml:"AI2"`
		AI3 teracomAnalogStruct `xml:"AI3"`
		AI4 teracomAnalogStruct `xml:"AI4"`
	} `xml:"AI"`
	VI struct {
		VI1 teracomAnalogStruct `xml:"VI1"`
		VI2 teracomAnalogStruct `xml:"VI2"`
		VI3 teracomAnalogStruct `xml:"VI3"`
		VI4 teracomAnalogStruct `xml:"VI4"`
	} `xml:"VI"`
	DI struct {
		DI1 teracomDigitalStruct `xml:"DI1"`
		DI2 teracomDigitalStruct `xml:"DI2"`
		DI3 teracomDigitalStruct `xml:"DI3"`
		DI4 teracomDigitalStruct `xml:"DI4"`
	} `xml:"DI"`
	R struct {
		R1 teracomRelayStruct `xml:"R1"`
		R2 teracomRelayStruct `xml:"R2"`
		R3 teracomRelayStruct `xml:"R3"`
		R4 teracomRelayStruct `xml:"R4"`
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

func (c *TeracomDevice) text(category, registerName, description, value string) {
	if len(value) < 1 {
		return
	}

	register := c.ds.addIgnoreRegister(
		category, registerName, description, "", dataflow.TextRegister, nil, false,
	)
	if register == nil {
		return
	}
	c.ds.stateOutput <- dataflow.NewTextRegisterValue(c.ds.deviceConfig.Name(), register, value)
}

func (c *TeracomDevice) number(category, registerName, description, unit string, value string) {
	if value == "---" {
		// this is teracom's way of encoding null
		return
	}

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return
	}

	register := c.ds.addIgnoreRegister(
		category, registerName, description, unit, dataflow.NumberRegister, nil, false,
	)
	if register == nil {
		return
	}
	c.ds.stateOutput <- dataflow.NewNumericRegisterValue(c.ds.deviceConfig.Name(), register, floatValue)
}

func (c *TeracomDevice) enum(
	category, registerName, description string, enum map[int]string, strValue string, controllable bool,
) {
	register := c.ds.addIgnoreRegister(
		category, registerName, description, "", dataflow.EnumRegister, enum, controllable,
	)
	if register == nil {
		return
	}

	enumIdx := func(string) int {
		for idx, v := range enum {
			if v == strValue {
				return idx
			}
		}
		return -1
	}(strValue)

	c.ds.stateOutput <- dataflow.NewEnumRegisterValue(c.ds.deviceConfig.Name(), register, enumIdx)
}

func (c *TeracomDevice) relay(
	registerName, description string, strValue string, controllable bool,
) {
	c.enum("Relays", registerName, description,
		map[int]string{
			0: "OFF",
			1: "ON",
			2: "in pulse",
		},
		strValue,
		controllable,
	)
}

func (c *TeracomDevice) alarm(category, registerName, description string, strValue string) {
	enum := map[int]string{
		0: "OK",
		1: "ALARMED",
	}

	register := c.ds.addIgnoreRegister(
		category, registerName, description, "", dataflow.EnumRegister, enum, false,
	)
	if register == nil {
		return
	}

	enumIdx := 0
	if strValue == "1" {
		enumIdx = 1
	}

	c.ds.stateOutput <- dataflow.NewEnumRegisterValue(c.ds.deviceConfig.Name(), register, enumIdx)
}

func (c *TeracomDevice) extractRegistersAndValues(s teracomStatusStruct) {
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
	c.alarm(cat, "Alarmed", "Alarmed", s.Alarmed)
	c.text(cat, "Date", "Date", s.Time.Date)
	c.text(cat, "Time", "Time", s.Time.Time)

	// sensors
	sensor := func(sIdx int, s teracomSensorStruct) {
		if s.ID == "0000000000000000" {
			return
		}
		item := func(sIdx, vIdx int, s teracomSensorStruct, i teracomSensorValueStruct) {
			if i.Value == "---" {
				return
			}

			regName := fmt.Sprintf("S%dV%d", sIdx, vIdx)
			desc := s.Description

			c.number("Sensors", regName, desc, i.Unit, i.Value)
			c.alarm("Alarms", regName+"Alarm", desc, i.Alarm)

			c.number("Settings", regName+"Min", desc+" Min", i.Unit, i.Min)
			c.number("Settings", regName+"Max", desc+" Max", i.Unit, i.Max)
			c.number("Settings", regName+"Hys", desc+" Hysteresis", i.Unit, i.Hys)
		}

		item(sIdx, 1, s, s.Item1)
		item(sIdx, 2, s, s.Item2)

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
	analog := func(regNamePrefix, valueCat string, sIdx int, a teracomAnalogStruct) {
		if a.Value == "---" {
			return
		}

		regName := fmt.Sprintf("%s%d", regNamePrefix, sIdx)
		desc := a.Description

		c.number(valueCat, regName, desc, a.Unit, a.Value)
		c.alarm("Alarms", regName+"Alarm", desc, a.Alarm)

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
	digital := func(sIdx int, a teracomDigitalStruct) {
		regName := fmt.Sprintf("DI%d", sIdx)
		desc := a.Description

		c.enum("Digital Inputs", regName, desc,
			map[int]string{
				0: "OPEN",
				1: "CLOSED",
			},
			a.Value,
			false,
		)
		c.alarm("Alarms", regName+"Alarm", desc, a.Alarm)
	}
	digital(1, s.DI.DI1)
	digital(2, s.DI.DI2)
	digital(3, s.DI.DI3)
	digital(4, s.DI.DI4)

	// relays
	relay := func(sIdx int, r teracomRelayStruct) {
		regName := fmt.Sprintf("R%d", sIdx)
		desc := r.Description

		controllable := r.Control == "0"
		c.relay(regName, desc, r.Value, controllable)
		if !controllable {
			c.text("Relays", regName+"Control", desc+" is controlled by", r.Control)
		}
	}
	relay(1, s.R.R1)
	relay(2, s.R.R2)
	relay(3, s.R.R3)
	relay(4, s.R.R4)
}
