package httpDevice

import (
	"encoding/json"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/pkg/errors"
	"net/http"
)

type ShellyEm3Device struct {
	ds *DeviceStruct
}

func (c *ShellyEm3Device) GetPath() string {
	return "status"
}

func (c *ShellyEm3Device) HandleResponse(body []byte) error {
	var status ShellyEm3StatusStruct
	if err := json.Unmarshal(body, &status); err != nil {
		return fmt.Errorf("cannot parse json: %s", err)
	}
	c.extractRegistersAndValues(status)

	return nil
}

func (c *ShellyEm3Device) GetCategorySort(category string) int {
	switch category {
	case "Essential":
		return 0
	case "Energy Meters":
		return 1
	case "Relays":
		return 2
	case "Network Status":
		return 3
	case "Wifi":
		return 4
	case "System":
		return 5
	default:
		panic("unknown category: " + category)
	}
}

func (c *ShellyEm3Device) CommandValueRequest(value dataflow.Value) (*http.Request, OnCommandSuccess, error) {
	return nil, nil, errors.New("not implemented")
}

type ShellyEm3StatusStruct struct {
	WifiSta struct {
		Connected bool   `json:"connected"`
		Ssid      string `json:"ssid"`
		Ip        string `json:"ip"`
		Rssi      int    `json:"rssi"`
	} `json:"wifi_sta"`
	Cloud struct {
		Enabled   bool `json:"enabled"`
		Connected bool `json:"connected"`
	} `json:"cloud"`
	Mqtt struct {
		Connected bool `json:"connected"`
	} `json:"mqtt"`
	Time          string `json:"time"`
	Serial        int    `json:"serial"`
	HasUpdate     bool   `json:"has_update"`
	Mac           string `json:"mac"`
	CfgChangedCnt int    `json:"cfg_changed_cnt"`
	Relays        []struct {
		Ison           bool   `json:"ison"`
		HasTimer       bool   `json:"has_timer"`
		TimerStarted   int    `json:"timer_started"`
		TimerDuration  int    `json:"timer_duration"`
		TimerRemaining int    `json:"timer_remaining"`
		Overpower      bool   `json:"overpower"`
		IsValid        bool   `json:"is_valid"`
		Source         string `json:"source"`
	} `json:"relays"`
	Emeters []struct {
		Power         float64 `json:"power"`
		Pf            float64 `json:"pf"`
		Current       float64 `json:"current"`
		Voltage       float64 `json:"voltage"`
		IsValid       bool    `json:"is_valid"`
		Total         float64 `json:"total"`
		TotalReturned float64 `json:"total_returned"`
	} `json:"emeters"`
	TotalPower float64 `json:"total_power"`
	EmeterN    struct {
		Current  float64 `json:"current"`
		Ixsum    float64 `json:"ixsum"`
		Mismatch bool    `json:"mismatch"`
		IsValid  bool    `json:"is_valid"`
	} `json:"emeter_n"`
	Uptime int `json:"uptime"`
}

func (c *ShellyEm3Device) text(category, registerName, description, value string) {
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

func (c *ShellyEm3Device) number(category, registerName, description, unit string, value float64) {
	register := c.ds.addIgnoreRegister(
		category, registerName, description, unit, dataflow.NumberRegister, nil, false,
	)
	if register == nil {
		return
	}
	c.ds.StateStorage().Fill(dataflow.NewNumericRegisterValue(c.ds.Name(), register, value))
}

func (c *ShellyEm3Device) boolean(category, registerName, description string, value bool) {
	register := c.ds.addIgnoreRegister(
		category, registerName, description, "",
		dataflow.NumberRegister,
		map[int]string{
			0: "false",
			1: "true",
		},
		false,
	)
	if register == nil {
		return
	}

	var intValue = 0
	if value {
		intValue = 1
	}
	c.ds.StateStorage().Fill(dataflow.NewEnumRegisterValue(c.ds.Name(), register, intValue))
}

func (c *ShellyEm3Device) extractRegistersAndValues(s ShellyEm3StatusStruct) {
	// Essential
	c.number("Essential", "TotalPower", "Total Power", "W", s.TotalPower)

	// Meters
	{
		cat := "Energy Meters"
		for idx, e := range s.Emeters {
			regName := fmt.Sprintf("Emeter%d", idx+1)
			desc := fmt.Sprintf("P%d", idx+1)
			c.number("Essential", regName+"Power", desc+" Power", "W", e.Power)
			c.number(cat, regName+"Pf", desc+" Power Factor", "", e.Pf)
			c.number(cat, regName+"Current", desc+" Current", "A", e.Current)
			c.number("Essential", regName+"Voltage", desc+" Voltage", "V", e.Voltage)
			c.boolean(cat, regName+"IsValid", desc+" Is Valid", e.IsValid)
			c.number(cat, regName+"Total", desc+" Total", "Wh", e.Total)
			c.number(cat, regName+"TotalReturned", desc+" TotalReturned", "Wh", e.TotalReturned)
		}

		e := s.EmeterN
		c.number(cat, "NCurrent", "Neutral Current", "A", e.Current)
		c.boolean(cat, "Mismatch", "Mismatch", e.Mismatch)
		c.boolean(cat, "NIsValid", "Is Valid", e.IsValid)
	}

	{
		cat := "Relays"
		for idx, r := range s.Relays {
			regName := fmt.Sprintf("Relay%d", idx+1)
			desc := fmt.Sprintf("Relay%d", idx+1)
			c.boolean("Essential", regName+"IsOn", desc+" is on", r.Ison)
			c.boolean(cat, regName+"HasTimer", desc+" has timer", r.HasTimer)
			c.number(cat, regName+"TimerStarted", desc+" timer started", "", float64(r.TimerStarted))
			c.number(cat, regName+"TimerDuration", desc+" timer duration", "", float64(r.TimerDuration))
			c.number(cat, regName+"TimerRemaining", desc+" timer remaining", "", float64(r.TimerRemaining))
			c.boolean(cat, regName+"Overpower", desc+" over power", r.Overpower)
			c.boolean(cat, regName+"IsValid", desc+" is valid", r.IsValid)
			c.text(cat, regName+"Source", desc+" source", r.Source)
		}
	}

	{
		cat := "Network Status"
		c.boolean(cat, "WifiConnected", "Wifi Connected", s.WifiSta.Connected)
		c.boolean(cat, "CloudEnabled", "Cloud enabled", s.Cloud.Enabled)
		c.boolean(cat, "CloudConnected", "Cloud connected", s.Cloud.Connected)
		c.boolean(cat, "MqttConnected", "Mqtt Connected", s.Mqtt.Connected)
	}

	{
		cat := "Wifi"
		w := s.WifiSta
		c.text(cat, "WifiSsid", "SSID", w.Ssid)
		c.text(cat, "WifiIp", "IP", w.Ip)
		c.number(cat, "WifiRssi", "RSSI", "", float64(w.Rssi))
	}

	{
		cat := "System"
		c.text(cat, "Time", "Time", s.Time)
		c.number(cat, "Serial", "Serial", "", float64(s.Serial))
		c.boolean(cat, "HasUpdate", "Has update", s.HasUpdate)
		c.text(cat, "Mac", "MAC", s.Mac)
		c.number(cat, "CfgChangedCnt", "Configuration changed counter", "", float64(s.CfgChangedCnt))
		c.number(cat, "Uptime", "Uptime", "s", float64(s.Uptime))
	}
}
