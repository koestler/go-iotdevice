package httpDevice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/dataflow"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type TeracomDevice struct {
	ds *DeviceStruct
}

func (c *TeracomDevice) GetPath() string {
	return "status.json"
}

func (c *TeracomDevice) HandleResponse(body []byte) error {
	var status teracomStatusStruct

	// trim Byte Order Mark (BOM) before decoding
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))

	if err := json.Unmarshal(body, &status); err != nil {
		return fmt.Errorf("cannot parse json: %s", err)
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

func (c *TeracomDevice) CommandValueRequest(value dataflow.Value) (*http.Request, OnCommandSuccess, error) {
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
		c.relay(value.Register().Name(), value.Register().Description(), enum.Value(), value.Register().Commandable())
	}

	req, err := http.NewRequest("POST", "/monitor/monitor.htm", strings.NewReader(values.Encode()))
	return req, onSuccess, err
}

type teracomSensorValueStruct struct {
	Value string `json:"value"`
	Unit  string `json:"unit"`
	Alarm string `json:"alarm"`
	Min   string `json:"min"`
	Max   string `json:"max"`
	Hys   string `json:"hys"`
}

type teracomSensorStruct struct {
	Description string                   `json:"description"`
	ID          string                   `json:"id"`
	Item1       teracomSensorValueStruct `json:"item1"`
	Item2       teracomSensorValueStruct `json:"item2"`
}

type teracomAnalogStruct struct {
	Description string `json:"description"`
	Value       string `json:"value"`
	Unit        string `json:"unit"`
	Multiplier  string `json:"multiplier"`
	Offset      string `json:"offset"`
	Alarm       string `json:"alarm"`
	Min         string `json:"min"`
	Max         string `json:"max"`
	Hys         string `json:"hys"`
}

type teracomDigitalStruct struct {
	Description string `json:"description"`
	Value       string `json:"value"`
	Valuebin    string `json:"valuebin"`
	AlarmState  string `json:"alarmState"`
	Alarm       string `json:"alarm"`
}

type teracomRelayStruct struct {
	Description string `json:"description"`
	Value       string `json:"value"`
	Valuebin    string `json:"valuebin"`
	PulseWidth  string `json:"pulseWidth"`
	Control     string `json:"control"`
}

type teracomStatusStruct struct {
	Monitor struct {
		DeviceInfo struct {
			DeviceName  string `json:"DeviceName"`
			HostName    string `json:"HostName"`
			ID          string `json:"ID"`
			FwVer       string `json:"FwVer"`
			SysContact  string `json:"SysContact"`
			SysName     string `json:"SysName"`
			SysLocation string `json:"SysLocation"`
		} `json:"DeviceInfo"`
		S struct {
			S1 teracomSensorStruct `json:"S1"`
			S2 teracomSensorStruct `json:"S2"`
			S3 teracomSensorStruct `json:"S3"`
			S4 teracomSensorStruct `json:"S4"`
			S5 teracomSensorStruct `json:"S5"`
			S6 teracomSensorStruct `json:"S6"`
			S7 teracomSensorStruct `json:"S7"`
			S8 teracomSensorStruct `json:"S8"`
		}
		AI struct {
			AI1 teracomAnalogStruct `json:"AI1"`
			AI2 teracomAnalogStruct `json:"AI2"`
			AI3 teracomAnalogStruct `json:"AI3"`
			AI4 teracomAnalogStruct `json:"AI4"`
		} `json:"AI"`
		VI struct {
			VI1 teracomAnalogStruct `json:"VI1"`
			VI2 teracomAnalogStruct `json:"VI2"`
			VI3 teracomAnalogStruct `json:"VI3"`
			VI4 teracomAnalogStruct `json:"VI4"`
		} `json:"VI"`
		DI struct {
			DI1 teracomDigitalStruct `json:"DI1"`
			DI2 teracomDigitalStruct `json:"DI2"`
			DI3 teracomDigitalStruct `json:"DI3"`
			DI4 teracomDigitalStruct `json:"DI4"`
		} `json:"DI"`
		R struct {
			R1 teracomRelayStruct `json:"R1"`
			R2 teracomRelayStruct `json:"R2"`
			R3 teracomRelayStruct `json:"R3"`
			R4 teracomRelayStruct `json:"R4"`
		} `json:"R"`
		HTTPPush struct {
			Key        string `json:"Key"`
			PushPeriod string `json:"PushPeriod"`
		} `json:"HTTPPush"`
		Hwerr   string `json:"hwerr"`
		Alarmed string `json:"Alarmed"`

		Scannig string `json:"Scannig"`
		Time    struct {
			Date string `json:"Date"`
			Time string `json:"Time"`
		} `json:"Time"`
	} `json:"Monitor"`
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
	c.ds.StateStorage().Fill(dataflow.NewTextRegisterValue(c.ds.Name(), register, value))
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
	c.ds.StateStorage().Fill(dataflow.NewNumericRegisterValue(c.ds.Name(), register, floatValue))
}

func (c *TeracomDevice) enum(
	category, registerName, description string, enum map[int]string, strValue string, commandable bool,
) {
	register := c.ds.addIgnoreRegister(
		category, registerName, description, "", dataflow.EnumRegister, enum, commandable,
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

	c.ds.StateStorage().Fill(dataflow.NewEnumRegisterValue(c.ds.Name(), register, enumIdx))
}

func (c *TeracomDevice) relay(
	registerName, description string, strValue string, commandable bool,
) {
	c.enum("Relays", registerName, description,
		map[int]string{
			0: "OFF",
			1: "ON",
			2: "in pulse",
		},
		strValue,
		commandable,
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

	c.ds.StateStorage().Fill(dataflow.NewEnumRegisterValue(c.ds.Name(), register, enumIdx))
}

func (c *TeracomDevice) extractRegistersAndValues(s teracomStatusStruct) {
	m := s.Monitor

	// device info
	cat := "Device Info"
	c.text(cat, "DeviceName", "Device Name", m.DeviceInfo.DeviceName)
	c.text(cat, "HostName", "Host Name", m.DeviceInfo.HostName)
	c.text(cat, "Id", "Id", m.DeviceInfo.ID)
	c.text(cat, "FWVer", "Firmware Vesion", m.DeviceInfo.FwVer)
	c.text(cat, "SysContact", "System Contact", m.DeviceInfo.SysContact)
	c.text(cat, "SysName", "System Name", m.DeviceInfo.SysName)
	c.text(cat, "SysLocation", "System Location", m.DeviceInfo.SysLocation)

	// general
	cat = "General"
	c.text(cat, "Hwerr", "Hardware Error", m.Hwerr)
	c.alarm(cat, "Alarmed", "Alarmed", m.Alarmed)
	c.text(cat, "Date", "Date", m.Time.Date)
	c.text(cat, "Time", "Time", m.Time.Time)

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
	sensor(1, m.S.S1)
	sensor(2, m.S.S2)
	sensor(3, m.S.S3)
	sensor(4, m.S.S4)
	sensor(5, m.S.S5)
	sensor(6, m.S.S6)
	sensor(7, m.S.S7)
	sensor(8, m.S.S8)

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
	analog("AI", "Analog Inputs", 1, m.AI.AI1)
	analog("AI", "Analog Inputs", 2, m.AI.AI2)
	analog("AI", "Analog Inputs", 3, m.AI.AI3)
	analog("AI", "Analog Inputs", 4, m.AI.AI4)

	// virtual inputs
	analog("VI", "Virtual Inputs", 1, m.VI.VI1)
	analog("VI", "Virtual Inputs", 2, m.VI.VI2)
	analog("VI", "Virtual Inputs", 3, m.VI.VI3)
	analog("VI", "Virtual Inputs", 4, m.VI.VI4)

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
	digital(1, m.DI.DI1)
	digital(2, m.DI.DI2)
	digital(3, m.DI.DI3)
	digital(4, m.DI.DI4)

	// relays
	relay := func(sIdx int, r teracomRelayStruct) {
		regName := fmt.Sprintf("R%d", sIdx)
		desc := r.Description

		commandable := r.Control == "0"
		c.relay(regName, desc, r.Value, commandable)
		if !commandable {
			c.text("Relays", regName+"Control", desc+" is controlled by", r.Control)
		}
	}
	relay(1, m.R.R1)
	relay(2, m.R.R2)
	relay(3, m.R.R3)
	relay(4, m.R.R4)
}
