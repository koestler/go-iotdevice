package config

import (
	"bytes"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/koestler/go-iotdevice/v3/internal/types"
)

const (
	InvalidSyntaxConfig = `- -`

	InvalidUnknownVersionConfig = `
Version: 42
`

	ValidCompleteConfig = `
Version: 2
ProjectTitle: Configurable Title of Project
LogConfig: false
LogWorkerStart: false
LogStateStorageDebug: true
LogCommandStorageDebug: true

HttpServer:
  Bind: "[::1]"
  Port: 8001
  LogRequests: false
  FrontendProxy: "http://127.0.0.1:3000/"
  FrontendPath: ./my-frontend/
  DocsPath: /tmp/mydocs/
  FrontendExpires: 1m
  ConfigExpires: 2m
  LogDebug: true

Authentication:
  JwtSecret: 'aiziax9Hied0ier9Yo0Lo6bi3xahth7o'
  JwtValidityPeriod: 2h
  HtaccessFile: ./my-auth.passwd

MqttClients:
  0-local:
    Broker: tcp://mqtt.example.com:1883
    ProtocolVersion: 5

    User: dev
    Password: zee4AhRi
    ClientId: server42

    KeepAlive: 2m
    ConnectRetryDelay: 20s
    ConnectTimeout: 10s
    TopicPrefix: my-prefix/
    ReadOnly: false
    MaxBacklogSize: 42

    MqttDevices:
      bmv1:
        MqttTopics:
          - stat/go-iotdevice/bmv1/+

    AvailabilityClient:
      Enabled: false
      TopicTemplate: '%Prefix%tele-X/%ClientId%/status'
      Retain: false
      Qos: 0

    AvailabilityDevice:
      Enabled: true
      TopicTemplate: '%Prefix%tele-X/%ClientId%/%DeviceName%/status'
      Retain: false
      Qos: 1
      Devices:
        bmv0:

    Structure:
      Enabled: false
      TopicTemplate: '%Prefix%struct-X/go-iotdevice/%DeviceName%'
      Interval: 20s
      Retain: false
      Qos: 2
      Devices:
        bmv0:
          Filter:
            IncludeRegisters:
              - reg-inc-a
              - reg-inc-b        
            SkipRegisters:
              - reg-skip-a
              - reg-skip-b
            IncludeCategories:
              - cat-inc-a
              - cat-inc-b
            SkipCategories:
              - cat-skip-a
              - cat-skip-b
            DefaultInclude: False


    Telemetry:
      Enabled: false
      TopicTemplate: '%Prefix%tele-X/go-iotdevice/%DeviceName%/state'
      Interval: 30s
      Retain: true
      Qos: 2
      Devices:
        modbus-rtu0:
          Filter:
            IncludeRegisters:
              - BatteryVoltage
              - Power            
            SkipRegisters:
            IncludeCategories:
              - Essential
            SkipCategories:
            DefaultInclude: False
        tcw241:

    Realtime:
      Enabled: false
      TopicTemplate: '%Prefix%real-X/go-iotdevice/%DeviceName%/%RegisterName%'
      Interval: 2s
      Retain: true

    HomeassistantDiscovery:
      Enabled: true
      TopicTemplate: 'my-hass/%Component%/%NodeId%/%ObjectId%/config'
      Interval: 0s
      Retain: true
      Qos: 2
      Devices:

    Command:
      Enabled: false
      TopicTemplate: '%Prefix%cmnd-X/go-iotdevice/%DeviceName%/%RegisterName%'

    LogDebug: true
    LogMessages: true

  1-remote:
    Broker: "ssl://eu1.cloud.thethings.network:8883"

  2-readonly:
    Broker: "tcp://example.com:1883"
    ReadOnly: true

Modbus:
  bus0:
    Device: /dev/ttyACM0
    BaudRate: 9600
    ReadTimeout: 200ms
    LogDebug: true

VictronDevices:
  bmv0:
    Filter:
      SkipRegisters:
        - Temperature
        - AuxVoltage
      SkipCategories:
        - Settings
    RestartInterval: 400ms
    RestartIntervalMaxBackoff: 2m
    LogDebug: true
    LogComDebug: true
    Device: /dev/serial/by-id/usb-VictronEnergy_BV_VE_Direct_cable_VEHTVQT-if00-port0
    Kind: Vedirect
    PollInterval: 700ms
    IoLog: /tmp/bmv0.log

ModbusDevices:
  modbus-rtu0:
    Filter:
      IncludeRegisters:
        - a
        - b
      SkipRegisters:
        - c
        - d
      IncludeCategories:
        - A
        - B
        - C
      SkipCategories:
        - D
        - E
        - F
      DefaultInclude: False
    RestartInterval:
    RestartIntervalMaxBackoff:
    LogDebug: false
    LogComDebug: false
    Bus: bus0
    Kind: WaveshareRtuRelay8
    Address: 0x01
    Relays:
      CH1:
        Description: Lamp
        OpenLabel: Off
        ClosedLabel: On
    PollInterval: 1s

GpioDevices:
  gpio0:
    Filter:
      IncludeRegisters:
      SkipRegisters:
      IncludeCategories:
      SkipCategories:
      DefaultInclude: True
    RestartInterval: 400ms
    RestartIntervalMaxBackoff: 200s
    LogDebug: true
    Chip: gpiochip42
    InputDebounce: 10ms
    InputOptions: ["WithPullUp"]
    OutputOptions: ["AsPushPull"]
    Inputs:
      Switch0:
        Pin: GPIO2
        Description: Switch 0
        LowLabel: Pressed
        HighLabel: Released
      Switch1:
        Pin: GPIO3
        Description: Switch 1
    Outputs:
      Relay0:
        Pin: GPIO4
        Description: Relay 0
        LowLabel: Off
        HighLabel: On
      Relay1:
        Pin: GPIO5
        Description: Relay 1

HttpDevices:
  tcw241:
    RestartInterval: 1m
    RestartIntervalMaxBackoff: 2m
    LogDebug: false
    LogComDebug: false
    Url: http://control0/
    Kind: Teracom
    Username: admin
    Password: my-secret
    PollInterval: 5s

MqttDevices:
  bmv1:
    RestartInterval: 50ms
    RestartIntervalMaxBackoff: 30s
    LogDebug: false
    LogComDebug: true
    Kind: GoIotdeviceV3

GensetDevices:
  genset0:
    RestartInterval: 60ms
    RestartIntervalMaxBackoff: 3m
    LogDebug: false
    LogComDebug: true

    InputBindings:
      tcw241:
        IOAvailable: Available
        ArmSwitch: DI0
        ResetSwitch: DI1
        FireDetected: DI2

    OutputBindings:
      modbus-rtu0:
        Ignition: CH0
        Starter: CH1
        Fan: CH2
        Pump: CH3
        Load: CH4

    PrimingTimeout: 20s
    CrankingTimeout: 19s
    StabilizingTimeout: 18s
    WarmUpTimeout: 17s
    WarmUpMinTime: 16s
    WarmUpTemp: 110
    EngineCoolDownTimeout: 15s
    EngineCoolDownMinTime: 14s
    EngineCoolDownTemp: 120
    EnclosureCoolDownTimeout: 13s
    EnclosureCoolDownMinTime: 12s
    EnclosureCoolDownTemp: 130

    EngineTempMin: 140
    EngineTempMax: 150
    AuxTemp0Min: 160
    AuxTemp0Max: 170
    AuxTemp1Min: 180
    AuxTemp1Max: 190

    SinglePhase: true
    UMin: 200
    UMax: 210
    UAvgWindow: 2
    FMin: 220
    FMax: 230
    FAvgWindow: 2
    PMax: 240
    PTotMax: 250

Views:
  - Name: private
    Title: Private
    Devices:
      - Name: bmv0
        Title: Battery Monitor
        Filter:
          SkipRegisters:
            - field-a
            - field-b
          SkipCategories:
            - cat-a
            - cat-b
            - cat-c
      - Name: modbus-rtu0
        Title: Relay Board
    Autoplay: false
    AllowedUsers:
      - test0
    Hidden: true
  - Name: public
    Title: Public
    Devices:
      - Name: bmv0
        Title: Bmv 0
`

	ValidDefaultConfig = `
Version: 2

HttpServer:
  Bind: "[::1]"

Authentication:
  HtaccessFile: ./my-auth.passwd

MqttClients:
  0-local:
    Broker: tcp://mqtt.example.com:1883
    MqttDevices:
      bmv1:
        MqttTopics:
          - stat/go-iotdevice/bmv1/+

Modbus:
  bus0:
    Device: /dev/ttyUSB1
    BaudRate: 1200

VictronDevices:
  bmv0:
    Kind: RandomBmv

ModbusDevices:
  modbus-rtu0:
    Bus: bus0
    Kind: WaveshareRtuRelay8
    Address: 0x02

GpioDevices:
  gpio0:
    Inputs:
      Switch0:
        Pin: GPIO2
        Description: Switch 0

HttpDevices:
  tcw241:
    Url: http://control0/
    Kind: Teracom

MqttDevices:
  bmv1:
    Kind: GoIotdeviceV3

GensetDevices:
  genset0:

    InputBindings:
      tcw241:
        IOAvailable: Available

    OutputBindings:
      modbus-rtu0:
        Ignition: CH0

Views:
  - Name: private
    Title: Private
    Devices:
      - Name: bmv0
        Title: Battery Monitor
`
)

func containsError(needle string, err []error) bool {
	for _, e := range err {
		if strings.Contains(e.Error(), needle) {
			return true
		}
	}
	return false
}

func TestReadConfig_InvalidSyntax(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidSyntaxConfig), true)
	if len(err) != 1 {
		t.Fatalf("expect one error for invalid file, got: %v", err)
	}
	if !strings.Contains(err[0].Error(), "cannot unmarshal") {
		t.Errorf("expect a yaml unmarshal error but got: %v", err[0])
	}
}

func TestReadConfig_NoVersion(t *testing.T) {
	_, err := ReadConfig([]byte("LogConfig: true"), true)

	if !containsError("Version must be defined", err) {
		t.Errorf("expect 'version must be defined' error, but got %s", err)
	}
}

func TestReadConfig_InvalidUnknownVersion(t *testing.T) {
	_, err := ReadConfig([]byte(InvalidUnknownVersionConfig), true)
	if len(err) != 1 || !strings.Contains(err[0].Error(), "version=42 is not supported") {
		t.Errorf("expect 1 error: 'version=42 is not supported' but got: %v", err)
	}
}

// check that a complex example setting all available options is correctly read
func TestReadConfig_Complete(t *testing.T) {
	config, err := ReadConfig([]byte(ValidCompleteConfig), true)
	if len(err) > 0 {
		t.Errorf("did not expect any errors, got: %v", err)
	}

	// General Section
	if expect, got := 2, config.Version(); expect != got {
		t.Errorf("expect Version to be %d but got %d", expect, got)
	}

	if expect, got := "Configurable Title of Project", config.ProjectTitle(); expect != got {
		t.Errorf("expect ProjectTitle to be '%s' but got '%s'", expect, got)
	}

	if config.LogConfig() {
		t.Errorf("expect LogConfig to be false")
	}

	if config.LogWorkerStart() {
		t.Errorf("expect LogWorkerStart to be false")
	}

	if !config.LogStateStorageDebug() {
		t.Errorf("expect LogStateStorageDebug to be true")
	}

	if !config.LogCommandStorageDebug() {
		t.Errorf("expect LogCommandStorageDebug to be true")
	}

	{
		hs := config.HttpServer()
		if !hs.Enabled() {
			t.Error("expect HttpServer->Enabled to be true")
		}

		if expect, got := "[::1]", hs.Bind(); expect != got {
			t.Errorf("expect HttpServer->Bind to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 8001, hs.Port(); expect != got {
			t.Errorf("expect HttpServer->Port to be %d but got %d", expect, got)
		}

		if hs.LogRequests() {
			t.Error("expect HttpServer->LogRequests to be false")
		}

		if expect, gotA := "http://127.0.0.1:3000/", hs.FrontendProxy(); gotA != nil {
			if got := gotA.String(); expect != got {
				t.Errorf("expect HttpServer->FrontendProxy to be '%s' but got '%s'", expect, got)
			}
		} else {
			t.Error("expect HttpServer->FrontendProxy not to be nil")
		}

		if expect, got := "./my-frontend/", hs.FrontendPath(); expect != got {
			t.Errorf("expect HttpServer->FrontendPath to be '%s' but got '%s'", expect, got)
		}

		if expect, got := "/tmp/mydocs/", hs.DocsPath(); expect != got {
			t.Errorf("expect HttpServer->DocsPath to be '%s' but got '%s'", expect, got)
		}

		if expect, got := time.Minute, hs.FrontendExpires(); expect != got {
			t.Errorf("expect HttpServer->FrontendExpires to be %s but got %s", expect, got)
		}

		if expect, got := 2*time.Minute, hs.ConfigExpires(); expect != got {
			t.Errorf("expect HttpServer->ConfigExpires to be %s but got %s", expect, got)
		}

		if !hs.LogDebug() {
			t.Error("expect HttpServer->LogDebug to be true")
		}
	}

	{
		a := config.Authentication()

		if !a.Enabled() {
			t.Error("expect Authentication->Enabled to be true")
		}

		if expect, got := "aiziax9Hied0ier9Yo0Lo6bi3xahth7o", string(a.JwtSecret()); expect != got {
			t.Errorf("expect Authentication->JwtSecret to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 2*time.Hour, a.JwtValidityPeriod(); expect != got {
			t.Errorf("expect Authentication->JwtValidityPeriod to be %s but got %s", expect, got)
		}

		if expect, got := "./my-auth.passwd", a.HtaccessFile(); expect != got {
			t.Errorf("expect Authentication->HtaccessFile to be '%s' but got '%s'", expect, got)
		}
	}

	if expect, got := 3, len(config.MqttClients()); expect != got {
		t.Errorf("expect length of config.MqttClients to be %d but got %d", expect, got)
	} else {
		{
			prefix := "MqttClients->0-local"

			mc := config.MqttClients()[0]

			if expect, got := "0-local", mc.Name(); expect != got {
				t.Errorf("expect Name of first MqttClient to be '%s' but got '%s'", expect, got)
			}

			if expect, got := "tcp://mqtt.example.com:1883", mc.Broker().String(); expect != got {
				t.Errorf("expect %s->Broker to be '%s' but got '%s'", prefix, expect, got)
			}

			if expect, got := 5, mc.ProtocolVersion(); expect != got {
				t.Errorf("expect %s->ProtocolVersion to be %d but got %d", prefix, expect, got)
			}

			if expect, got := "dev", mc.User(); expect != got {
				t.Errorf("expect %s->User to be '%s' but got '%s'", prefix, expect, got)
			}

			if expect, got := "zee4AhRi", mc.Password(); expect != got {
				t.Errorf("expect %s->Password to be '%s' but got '%s'", prefix, expect, got)
			}

			if expect, got := "server42", mc.ClientId(); expect != got {
				t.Errorf("expect %s->ClientId to be '%s' but got '%s'", prefix, expect, got)
			}

			if expect, got := 2*time.Minute, mc.KeepAlive(); expect != got {
				t.Errorf("expect %s->KeepAlive to be '%s' but got '%s'", prefix, expect, got)
			}

			if expect, got := 20*time.Second, mc.ConnectRetryDelay(); expect != got {
				t.Errorf("expect %s->ConnectRetryDelay to be '%s' but got '%s'", prefix, expect, got)
			}

			if expect, got := 10*time.Second, mc.ConnectTimeout(); expect != got {
				t.Errorf("expect %s->ConnectTimeout to be '%s' but got '%s'", prefix, expect, got)
			}

			if expect, got := "my-prefix/", mc.TopicPrefix(); expect != got {
				t.Errorf("expect %s->TopicPrefix to be '%s' but got '%s'", prefix, expect, got)
			}

			if got := mc.ReadOnly(); got {
				t.Errorf("expect %s->ReadOnly to be false", prefix)
			}

			if expect, got := 42, mc.MaxBacklogSize(); expect != got {
				t.Errorf("expect %s->MaxBacklogSize to be %d but got %d", prefix, expect, got)
			}

			if expect, got := 1, len(mc.MqttDevices()); expect != got {
				t.Errorf("expect length of %s->MqttDevices to be %d but got %d", prefix, expect, got)
			} else {
				d := mc.MqttDevices()[0]

				if expect, got := "bmv1", d.Name(); expect != got {
					t.Errorf("expect Name of first %s->MqttDevices to be '%s' but got '%s'", prefix, expect, got)
				}

				if expect, got := []string{"stat/go-iotdevice/bmv1/+"}, d.MqttTopics(); !reflect.DeepEqual(expect, got) {
					t.Errorf("expect %s->MqttDevices->bmv1->MqttTopics to be %v but got %v", prefix, expect, got)
				}
			}

			{
				mcSect := mc.AvailabilityClient()
				sPrefix := prefix + "->AvailabilityClient"

				if expect, got := "my-prefix/tele-X/server42/status", mc.AvailabilityClientTopic(); expect != got {
					t.Errorf("expect %s->AvailabilityClientTopic to be '%s' but got '%s'", prefix, expect, got)
				}

				if got := mcSect.Enabled(); got {
					t.Errorf("expect %s->Enabled to be false", sPrefix)
				}

				if expect, got := "%Prefix%tele-X/%ClientId%/status", mcSect.TopicTemplate(); expect != got {
					t.Errorf("expect %s->TopicTemplate to be '%s' but got '%s'", sPrefix, expect, got)
				}

				if expect, got := 0*time.Second, mcSect.Interval(); expect != got {
					t.Errorf("expect %s->Interval to be '%s' but got '%s'", sPrefix, expect, got)
				}

				if got := mcSect.Retain(); got {
					t.Errorf("expect %s->Retain to be false", sPrefix)
				}

				if expect, got := byte(0), mcSect.Qos(); expect != got {
					t.Errorf("expect %s->Qos to be %d but got %d", sPrefix, expect, got)
				}

				if expect, got := []string{}, getNames(mcSect.Devices()); !reflect.DeepEqual(expect, got) {
					t.Errorf("expect %s->Devices to be %v but got %v", sPrefix, expect, got)
				}
			}

			{
				mcSect := mc.AvailabilityDevice()
				sPrefix := prefix + "->AvailabilityDevice"

				if expect, got := "my-prefix/tele-X/server42/my-dev/status", mc.AvailabilityDeviceTopic("my-dev"); expect != got {
					t.Errorf("expect %s->AvailabilityDeviceTopic to be '%s' but got '%s'", prefix, expect, got)
				}

				if got := mcSect.Enabled(); !got {
					t.Errorf("expect %s->Enabled to be true", sPrefix)
				}

				if expect, got := "%Prefix%tele-X/%ClientId%/%DeviceName%/status", mcSect.TopicTemplate(); expect != got {
					t.Errorf("expect %s->TopicTemplate to be '%s' but got '%s'", sPrefix, expect, got)
				}

				if expect, got := 0*time.Second, mcSect.Interval(); expect != got {
					t.Errorf("expect %s->Interval to be '%s' but got '%s'", sPrefix, expect, got)
				}

				if got := mcSect.Retain(); got {
					t.Errorf("expect %s->Retain to be false", sPrefix)
				}

				if expect, got := byte(1), mcSect.Qos(); expect != got {
					t.Errorf("expect %s->Qos to be %d but got %d", sPrefix, expect, got)
				}

				if expect, got := []string{"bmv0"}, getNames(mcSect.Devices()); !reflect.DeepEqual(expect, got) {
					t.Errorf("expect %s->Devices to be %v but got %v", sPrefix, expect, got)
				} else {
					prefix := sPrefix + "->Devices->bmv0"
					rf := mcSect.Devices()[0].Filter()

					if got := rf.IncludeRegisters(); len(got) > 0 {
						t.Errorf("expect %s->IncludeRegisters to be empty, got %v", prefix, got)
					}

					if got := rf.SkipRegisters(); len(got) > 0 {
						t.Errorf("expect %s->SkipRegisters to be empty but got %v", prefix, got)
					}

					if got := rf.IncludeCategories(); len(got) > 0 {
						t.Errorf("expect %s->IncludeCategories to be empty but got %v", prefix, got)
					}

					if got := rf.SkipCategories(); len(got) > 0 {
						t.Errorf("expect %s->SkipCategories to be empty but got %v", prefix, got)
					}

					if got := rf.DefaultInclude(); !got {
						t.Errorf("expect %s->DefaultInclude to be true", prefix)
					}
				}
			}

			{
				mcSect := mc.Structure()
				sPrefix := prefix + "->Structure"

				if expect, got := "my-prefix/struct-X/go-iotdevice/my-dev", mc.StructureTopic("my-dev"); expect != got {
					t.Errorf("expect %s->StructureTopic to be '%s' but got '%s'", prefix, expect, got)
				}

				if got := mcSect.Enabled(); got {
					t.Errorf("expect %s->Enabled to be false", sPrefix)
				}

				if expect, got := "%Prefix%struct-X/go-iotdevice/%DeviceName%", mcSect.TopicTemplate(); expect != got {
					t.Errorf("expect %s->TopicTemplate to be '%s' but got '%s'", sPrefix, expect, got)
				}

				if expect, got := 20*time.Second, mcSect.Interval(); expect != got {
					t.Errorf("expect %s->Interval to be '%s' but got '%s'", sPrefix, expect, got)
				}

				if got := mcSect.Retain(); got {
					t.Errorf("expect %s->Retain to be false", sPrefix)
				}

				if expect, got := byte(2), mcSect.Qos(); expect != got {
					t.Errorf("expect %s->Qos to be %d but got %d", sPrefix, expect, got)
				}

				if expect, got := []string{"bmv0"}, getNames(mcSect.Devices()); !reflect.DeepEqual(expect, got) {
					t.Errorf("expect %s->Devices to be %v but got %v", sPrefix, expect, got)
				} else {
					prefix := sPrefix + "->Devices->bmv0"
					rf := mcSect.Devices()[0].Filter()

					if expect, got := []string{"reg-inc-a", "reg-inc-b"}, rf.IncludeRegisters(); !reflect.DeepEqual(expect, got) {
						t.Errorf("expect %s->IncludeRegisters to be %v but got %#v", prefix, expect, got)
					}

					if expect, got := []string{"reg-skip-a", "reg-skip-b"}, rf.SkipRegisters(); !reflect.DeepEqual(expect, got) {
						t.Errorf("expect %s->SkipRegisters to be %v but got %v", prefix, expect, got)
					}

					if expect, got := []string{"cat-inc-a", "cat-inc-b"}, rf.IncludeCategories(); !reflect.DeepEqual(expect, got) {
						t.Errorf("expect %s->IncludeCategories to be %v but got %v", prefix, expect, got)
					}

					if expect, got := []string{"cat-skip-a", "cat-skip-b"}, rf.SkipCategories(); !reflect.DeepEqual(expect, got) {
						t.Errorf("expect %s->SkipCategories to be %v but got %v", prefix, expect, got)
					}

					if got := rf.DefaultInclude(); got {
						t.Errorf("expect %s->DefaultInclude to be false", prefix)
					}
				}
			}

			{
				mcSect := mc.Telemetry()
				sPrefix := prefix + "->Telemetry"

				if expect, got := "my-prefix/tele-X/go-iotdevice/my-dev/state", mc.TelemetryTopic("my-dev"); expect != got {
					t.Errorf("expect %s->TelemetryTopic to be '%s' but got '%s'", prefix, expect, got)
				}

				if got := mcSect.Enabled(); got {
					t.Errorf("expect %s->Enabled to be false", sPrefix)
				}

				if expect, got := "%Prefix%tele-X/go-iotdevice/%DeviceName%/state", mcSect.TopicTemplate(); expect != got {
					t.Errorf("expect %s->TopicTemplate to be '%s' but got '%s'", sPrefix, expect, got)
				}

				if expect, got := 30*time.Second, mcSect.Interval(); expect != got {
					t.Errorf("expect %s->Interval to be '%s' but got '%s'", sPrefix, expect, got)
				}

				if got := mcSect.Retain(); !got {
					t.Errorf("expect %s->Retain to be true", sPrefix)
				}

				if expect, got := byte(2), mcSect.Qos(); expect != got {
					t.Errorf("expect %s->Qos to be %d but got %d", sPrefix, expect, got)
				}

				if expect, got := []string{"modbus-rtu0", "tcw241"}, getNames(mcSect.Devices()); !reflect.DeepEqual(expect, got) {
					t.Errorf("expect %s->Devices to be %v but got %v", sPrefix, expect, got)
				} else {
					prefix := sPrefix + "->Devices->modbus-rtu0"
					rf := mcSect.Devices()[0].Filter()

					if expect, got := []string{"BatteryVoltage", "Power"}, rf.IncludeRegisters(); !reflect.DeepEqual(expect, got) {
						t.Errorf("expect %s->IncludeRegisters to be %v but got %#v", prefix, expect, got)
					}

					if got := rf.SkipRegisters(); len(got) > 0 {
						t.Errorf("expect %s->SkipRegisters to be empty but got %v", prefix, got)
					}

					if expect, got := []string{"Essential"}, rf.IncludeCategories(); !reflect.DeepEqual(expect, got) {
						t.Errorf("expect %s->IncludeCategories to be %v but got %v", prefix, expect, got)
					}

					if got := rf.SkipCategories(); len(got) > 0 {
						t.Errorf("expect %s->SkipCategories to be empty but got %v", prefix, got)
					}

					if got := rf.DefaultInclude(); got {
						t.Errorf("expect %s->DefaultInclude to be false", prefix)
					}
				}
			}

			{
				mcSect := mc.Realtime()
				sPrefix := prefix + "->Realtime"

				if expect, got := "my-prefix/real-X/go-iotdevice/my-dev/my-reg", mc.RealtimeTopic("my-dev", "my-reg"); expect != got {
					t.Errorf("expect %s->RealtimeTopic to be '%s' but got '%s'", prefix, expect, got)
				}

				if got := mcSect.Enabled(); got {
					t.Errorf("expect %s->Enabled to be false", sPrefix)
				}

				if expect, got := "%Prefix%real-X/go-iotdevice/%DeviceName%/%RegisterName%", mcSect.TopicTemplate(); expect != got {
					t.Errorf("expect %s->TopicTemplate to be '%s' but got '%s'", sPrefix, expect, got)
				}

				if expect, got := 2*time.Second, mcSect.Interval(); expect != got {
					t.Errorf("expect %s->Interval to be '%s' but got '%s'", sPrefix, expect, got)
				}

				if got := mcSect.Retain(); !got {
					t.Errorf("expect %s->Retain to be true", sPrefix)
				}

				if expect, got := byte(1), mcSect.Qos(); expect != got {
					t.Errorf("expect %s->Qos to be %d but got %d", sPrefix, expect, got)
				}

				if expect, got := []string{"bmv0", "genset0", "gpio0", "modbus-rtu0", "tcw241"}, getNames(mcSect.Devices()); !reflect.DeepEqual(expect, got) {
					t.Errorf("expect %s->Devices to be %v but got %v", sPrefix, expect, got)
				} else {
					prefix := sPrefix + "->Devices->bmv0"
					rf := mcSect.Devices()[0].Filter()

					if got := rf.IncludeRegisters(); len(got) > 0 {
						t.Errorf("expect %s->IncludeRegisters to be empty, got %v", prefix, got)
					}

					if got := rf.SkipRegisters(); len(got) > 0 {
						t.Errorf("expect %s->SkipRegisters to be empty but got %v", prefix, got)
					}

					if got := rf.IncludeCategories(); len(got) > 0 {
						t.Errorf("expect %s->IncludeCategories to be empty but got %v", prefix, got)
					}

					if got := rf.SkipCategories(); len(got) > 0 {
						t.Errorf("expect %s->SkipCategories to be empty but got %v", prefix, got)
					}

					if got := rf.DefaultInclude(); !got {
						t.Errorf("expect %s->DefaultInclude to be true", prefix)
					}
				}
			}

			{
				mcSect := mc.HomeassistantDiscovery()
				sPrefix := prefix + "->HomeassistantDiscovery"

				if expect, got := "my-hass/a/b/c/config", mc.HomeassistantDiscoveryTopic("a", "b", "c"); expect != got {
					t.Errorf("expect %s->HomeassistantDiscoveryTopic to be '%s' but got '%s'", prefix, expect, got)
				}

				if got := mcSect.Enabled(); !got {
					t.Errorf("expect %s->Enabled to be true", sPrefix)
				}

				if expect, got := "my-hass/%Component%/%NodeId%/%ObjectId%/config", mcSect.TopicTemplate(); expect != got {
					t.Errorf("expect %s->TopicTemplate to be '%s' but got '%s'", sPrefix, expect, got)
				}

				if expect, got := 0*time.Second, mcSect.Interval(); expect != got {
					t.Errorf("expect %s->Interval to be '%s' but got '%s'", sPrefix, expect, got)
				}

				if got := mcSect.Retain(); !got {
					t.Errorf("expect %s->Retain to be true", sPrefix)
				}

				if expect, got := byte(2), mcSect.Qos(); expect != got {
					t.Errorf("expect %s->Qos to be %d but got %d", sPrefix, expect, got)
				}

				if expect, got := []string{"bmv0", "genset0", "gpio0", "modbus-rtu0", "tcw241"}, getNames(mcSect.Devices()); !reflect.DeepEqual(expect, got) {
					t.Errorf("expect %s->Devices to be %v but got %v", sPrefix, expect, got)
				} else {
					prefix := sPrefix + "->Devices->bmv0"
					rf := mcSect.Devices()[0].Filter()

					if got := rf.IncludeRegisters(); len(got) > 0 {
						t.Errorf("expect %s->IncludeRegisters to be empty, got %v", prefix, got)
					}

					if got := rf.SkipRegisters(); len(got) > 0 {
						t.Errorf("expect %s->SkipRegisters to be empty but got %v", prefix, got)
					}

					if got := rf.IncludeCategories(); len(got) > 0 {
						t.Errorf("expect %s->IncludeCategories to be empty but got %v", prefix, got)
					}

					if got := rf.SkipCategories(); len(got) > 0 {
						t.Errorf("expect %s->SkipCategories to be empty but got %v", prefix, got)
					}

					if got := rf.DefaultInclude(); !got {
						t.Errorf("expect %s->DefaultInclude to be true", prefix)
					}
				}
			}

			{
				mcSect := mc.Command()
				sPrefix := prefix + "->Command"

				if expect, got := "my-prefix/cmnd-X/go-iotdevice/my-dev/my-reg", mc.CommandTopic("my-dev", "my-reg"); expect != got {
					t.Errorf("expect %s->CommandTopic to be '%s' but got '%s'", prefix, expect, got)
				}

				if got := mcSect.Enabled(); got {
					t.Errorf("expect %s->Enabled to be false", sPrefix)
				}

				if expect, got := "%Prefix%cmnd-X/go-iotdevice/%DeviceName%/%RegisterName%", mcSect.TopicTemplate(); expect != got {
					t.Errorf("expect %s->TopicTemplate to be '%s' but got '%s'", sPrefix, expect, got)
				}

				if expect, got := 0*time.Second, mcSect.Interval(); expect != got {
					t.Errorf("expect %s->Interval to be '%s' but got '%s'", sPrefix, expect, got)
				}

				if got := mcSect.Retain(); got {
					t.Errorf("expect %s->Retain to be false", sPrefix)
				}

				if expect, got := byte(1), mcSect.Qos(); expect != got {
					t.Errorf("expect %s->Qos to be %d but got %d", sPrefix, expect, got)
				}

				if expect, got := []string{"bmv0", "genset0", "gpio0", "modbus-rtu0", "tcw241"}, getNames(mcSect.Devices()); !reflect.DeepEqual(expect, got) {
					t.Errorf("expect %s->Devices to be %v but got %v", sPrefix, expect, got)
				} else {
					prefix := sPrefix + "->Devices->bmv0"
					rf := mcSect.Devices()[0].Filter()

					if got := rf.IncludeRegisters(); len(got) > 0 {
						t.Errorf("expect %s->IncludeRegisters to be empty, got %v", prefix, got)
					}

					if got := rf.SkipRegisters(); len(got) > 0 {
						t.Errorf("expect %s->SkipRegisters to be empty but got %v", prefix, got)
					}

					if got := rf.IncludeCategories(); len(got) > 0 {
						t.Errorf("expect %s->IncludeCategories to be empty but got %v", prefix, got)
					}

					if got := rf.SkipCategories(); len(got) > 0 {
						t.Errorf("expect %s->SkipCategories to be empty but got %v", prefix, got)
					}

					if got := rf.DefaultInclude(); !got {
						t.Errorf("expect %s->DefaultInclude to be true", prefix)
					}
				}
			}

			if !mc.LogDebug() {
				t.Error("expect MqttClients->0-local->LogDebug to be true")
			}

			if !mc.LogMessages() {
				t.Error("expect MqttClients->0-local->LogMessages to be true")
			}
		}

		{
			mc := config.MqttClients()[1]

			if expect, got := "1-remote", mc.Name(); expect != got {
				t.Errorf("expect Name of second MqttClient to be '%s' but got '%s'", expect, got)
			}

			if expect, got := "ssl://eu1.cloud.thethings.network:8883", mc.Broker().String(); expect != got {
				t.Errorf("expect MqttClients->1-remote->Broker to be '%s' but got '%s'", expect, got)
			}

			if expect, got := 5, mc.ProtocolVersion(); expect != got {
				t.Errorf("expect MqttClients->1-remote->ProtocolVersion to be %d but got %d", expect, got)
			}

			if expect, got := "", mc.User(); expect != got {
				t.Errorf("expect MqttClients->1-remote->User to be '%s' but got '%s'", expect, got)
			}
		}

		{
			mc := config.MqttClients()[2]

			if expect, got := "2-readonly", mc.Name(); expect != got {
				t.Errorf("expect Name of third MqttClient to be '%s' but got '%s'", expect, got)
			}

			if got := mc.ReadOnly(); !got {
				t.Errorf("expect MqttClients->2-readonly->ReadOnly to be true")
			}

			if got := mc.AvailabilityClient().Enabled(); got {
				t.Error("expect MqttClients->2-readonly->AvailabilityClient->Enabled to be false")
			}

			if got := mc.AvailabilityDevice().Enabled(); got {
				t.Error("expect MqttClients->2-readonly->AvailabilityDevice->Enabled to be false")
			}

			if got := mc.Structure().Enabled(); got {
				t.Error("expect MqttClients->2-readonly->Structure->Enabled to be false")
			}

			if got := mc.Telemetry().Enabled(); got {
				t.Error("expect MqttClients->2-readonly->Telemetry->Enabled to be false")
			}

			if got := mc.Realtime().Enabled(); got {
				t.Error("expect MqttClients->2-readonly->Realtime->Enabled to be false")
			}

			if got := mc.HomeassistantDiscovery().Enabled(); got {
				t.Error("expect MqttClients->2-readonly->HomeassistantDiscovery->Enabled to be false")
			}

			if got := mc.Command().Enabled(); got {
				t.Error("expect MqttClients->2-readonly->Command->Enabled to be false")
			}

		}

	}

	if expect, got := 1, len(config.Modbus()); expect != got {
		t.Errorf("expect length of config.Modbus to be %d but got %d", expect, got)
	} else {
		mb := config.Modbus()[0]

		if expect, got := "bus0", mb.Name(); expect != got {
			t.Errorf("expect Name of first Modbus to be '%s' but got '%s'", expect, got)
		}

		if expect, got := "/dev/ttyACM0", mb.Device(); expect != got {
			t.Errorf("expect Modbus->bus0->Device to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 9600, mb.BaudRate(); expect != got {
			t.Errorf("expect Modbus->bus0->BaudRate to be %d but got %d", expect, got)
		}

		if expect, got := 200*time.Millisecond, mb.ReadTimeout(); expect != got {
			t.Errorf("expect Modbus->bus0->ReadTimeout to be %s but got %s", expect, got)
		}

		if !mb.LogDebug() {
			t.Error("expect Modbus->bus0->LogDebug to be true")
		}
	}

	if expect, got := 1, len(config.VictronDevices()); expect != got {
		t.Errorf("expect length of config.VictronDevices to be %d but got %d", expect, got)
	} else {
		vd := config.VictronDevices()[0]

		if expect, got := "bmv0", vd.Name(); expect != got {
			t.Errorf("expect Name of first VictronDevice to be '%s' but got '%s'", expect, got)
		}

		{
			prefix := "VictronDevices->bmv0->General->Filter"
			rf := vd.Filter()

			if got := rf.IncludeRegisters(); len(got) > 0 {
				t.Errorf("expect %s->IncludeRegisters to be empty, got %v", prefix, got)
			}

			if expect, got := []string{"Temperature", "AuxVoltage"}, rf.SkipRegisters(); !reflect.DeepEqual(expect, got) {
				t.Errorf("expect %s->SkipRegisters to be %v but got %v", prefix, expect, got)
			}

			if got := rf.IncludeCategories(); len(got) > 0 {
				t.Errorf("expect %s->IncludeCategories to be empty but got %v", prefix, got)
			}

			if expect, got := []string{"Settings"}, rf.SkipCategories(); !reflect.DeepEqual(expect, got) {
				t.Errorf("expect %s->SkipCategories to be %v but got %v", prefix, expect, got)
			}

			if got := rf.DefaultInclude(); !got {
				t.Errorf("expect %s->DefaultInclude to be true", prefix)
			}
		}

		if expect, got := 400*time.Millisecond, vd.RestartInterval(); expect != got {
			t.Errorf("expect VictronDevices->bmv0->General->RestartInterval to be %s but got %s", expect, got)
		}

		if expect, got := 2*time.Minute, vd.RestartIntervalMaxBackoff(); expect != got {
			t.Errorf("expect VictronDevices->bmv0->General->RestartIntervalMaxBackoff to be %s but got %s", expect, got)
		}

		if !vd.LogDebug() {
			t.Error("expect VictronDevices->bmv0->General->LogDebug to be true")
		}

		if !vd.LogComDebug() {
			t.Error("expect VictronDevices->bmv0->General->LogComDebug to be true")
		}

		if expect, got := "/dev/serial/by-id/usb-VictronEnergy_BV_VE_Direct_cable_VEHTVQT-if00-port0", vd.Device(); expect != got {
			t.Errorf("expect VictronDevices->bmv0->Device to be '%s' but got '%s'", expect, got)
		}

		if expect, got := types.VictronVedirectKind, vd.Kind(); expect != got {
			t.Errorf("expect VictronDevices->bmv0->Kind to be %s but got %s", expect, got)
		}

		if expect, got := "/tmp/bmv0.log", vd.IoLog(); expect != got {
			t.Errorf("expect VictronDevices->bmv0->IoLog to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 700*time.Millisecond, vd.PollInterval(); expect != got {
			t.Errorf("expect VictronDevices->bmv0->PollInterval to be %s but got %s", expect, got)
		}
	}

	if expect, got := 1, len(config.ModbusDevices()); expect != got {
		t.Errorf("expect length of config.ModbusDevices to be %d but got %d", expect, got)
	} else {
		md := config.ModbusDevices()[0]

		if expect, got := "modbus-rtu0", md.Name(); expect != got {
			t.Errorf("expect Name of first ModbusDevice to be '%s' but got '%s'", expect, got)
		}

		{
			prefix := "ModbusDevices->modbus-rtu0->General->FilterRegister"
			rf := md.Filter()

			if expect, got := []string{"a", "b"}, rf.IncludeRegisters(); !reflect.DeepEqual(expect, got) {
				t.Errorf("expect %s->IncludeRegisters to be %v but got %#v", prefix, expect, got)
			}

			if expect, got := []string{"c", "d"}, rf.SkipRegisters(); !reflect.DeepEqual(expect, got) {
				t.Errorf("expect %s->SkipRegisters to be %v but got %v", prefix, expect, got)
			}

			if expect, got := []string{"A", "B", "C"}, rf.IncludeCategories(); !reflect.DeepEqual(expect, got) {
				t.Errorf("expect %s->IncludeCategories to be %v but got %v", prefix, expect, got)
			}

			if expect, got := []string{"D", "E", "F"}, rf.SkipCategories(); !reflect.DeepEqual(expect, got) {
				t.Errorf("expect %s->SkipCategories to be %v but got %v", prefix, expect, got)
			}

			if got := rf.DefaultInclude(); got {
				t.Errorf("expect %s->DefaultInclude to be false", prefix)
			}
		}

		if expect, got := 200*time.Millisecond, md.RestartInterval(); expect != got {
			t.Errorf("expect ModbusDevices->modbus-rtu0->General->RestartInterval to be %s but got %s", expect, got)
		}

		if expect, got := time.Minute, md.RestartIntervalMaxBackoff(); expect != got {
			t.Errorf("expect ModbusDevices->modbus-rtu0->General->RestartIntervalMaxBackoff to be %s but got %s", expect, got)
		}

		if md.LogDebug() {
			t.Error("expect ModbusDevices->modbus-rtu0->General->LogDebug to be false")
		}

		if md.LogComDebug() {
			t.Error("expect ModbusDevices->modbus-rtu0->General->LogComDebug to be false")
		}

		if expect, got := "bus0", md.Bus(); expect != got {
			t.Errorf("expect ModbusDevices->modbus-rtu0->Bus to be '%s' but got '%s'", expect, got)
		}

		if expect, got := types.ModbusWaveshareRtuRelay8Kind, md.Kind(); expect != got {
			t.Errorf("expect ModbusDevices->modbus-rtu0->Kind to be %s but got %s", expect, got)
		}

		if expect, got := byte(0x01), md.Address(); expect != got {
			t.Errorf("expect ModbusDevices->modbus-rtu0->Address to be 0x%x but got 0x%x", expect, got)
		}

		if expect, got := "Lamp", md.RelayDescription("CH1"); expect != got {
			t.Errorf("expect ModbusDevices->modbus-rtu0->Relays->CH1->Description to be '%s' but got '%s'", expect, got)
		}

		if expect, got := "Off", md.RelayOpenLabel("CH1"); expect != got {
			t.Errorf("expect ModbusDevices->modbus-rtu0->Relays->CH1->OpenLabel to be '%s' but got '%s'", expect, got)
		}

		if expect, got := "On", md.RelayClosedLabel("CH1"); expect != got {
			t.Errorf("expect ModbusDevices->modbus-rtu0->Relays->CH1->ClosedLabel to be '%s' but got '%s'", expect, got)
		}
	}

	if expect, got := 1, len(config.GpioDevices()); expect != got {
		t.Errorf("expect length of config.GpioDevices to be %d but got %d", expect, got)
	} else {
		gd := config.GpioDevices()[0]

		if expect, got := "gpio0", gd.Name(); expect != got {
			t.Errorf("expect Name of first GpioDevices to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 400*time.Millisecond, gd.RestartInterval(); expect != got {
			t.Errorf("expect GpioDevices->gpio0->General->RestartInterval to be %s but got %s", expect, got)
		}

		if expect, got := 200*time.Second, gd.RestartIntervalMaxBackoff(); expect != got {
			t.Errorf("expect GpioDevices->gpio0->General->RestartIntervalMaxBackoff to be %s but got %s", expect, got)
		}

		if !gd.LogDebug() {
			t.Error("expect GpioDevices->gpio0->General->LogDebug to be true")
		}

		if gd.LogComDebug() {
			t.Error("expect GpioDevices->gpio0->General->LogComDebug to be false")
		}

		if expect, got := "gpiochip42", gd.Chip(); expect != got {
			t.Errorf("expect GpioDevices->gpio0->Chip to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 10*time.Millisecond, gd.InputDebounce(); expect != got {
			t.Errorf("expect GpioDevices->gpio0->InputDebounce to be %s but got %s", expect, got)
		}

		if expect, got := []string{"WithPullUp"}, gd.InputOptions(); !reflect.DeepEqual(expect, got) {
			t.Errorf("expect GpioDevices->gpio0->InputOptions to be %v but got %v", expect, got)
		}

		if expect, got := []string{"AsPushPull"}, gd.OutputOptions(); !reflect.DeepEqual(expect, got) {
			t.Errorf("expect GpioDevices->gpio0->OutputOptions to be %v but got %v", expect, got)
		}

		if expect, got := 2, len(gd.Inputs()); expect != got {
			t.Errorf("expect length of GpioDevices->gpio0->Inputs to be %d but got %d", expect, got)
		} else {
			{
				in := gd.Inputs()[0]

				if expect, got := "Switch0", in.Name(); expect != got {
					t.Errorf("expect Name of first GpioDevices->gpio0->Inputs to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "GPIO2", in.Pin(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Inputs->in0->Pin to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "Switch 0", in.Description(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Inputs->in0->Description to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "Pressed", in.LowLabel(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Inputs->in0->LowLabel to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "Released", in.HighLabel(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Inputs->in0->HighLabel to be '%s' but got '%s'", expect, got)
				}
			}
			{
				in := gd.Inputs()[1]

				if expect, got := "Switch1", in.Name(); expect != got {
					t.Errorf("expect Name of second GpioDevices->gpio0->Inputs to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "GPIO3", in.Pin(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Inputs->in0->Pin to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "Switch 1", in.Description(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Inputs->in0->Description to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "low", in.LowLabel(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Inputs->in0->LowLabel to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "high", in.HighLabel(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Inputs->in0->HighLabel to be '%s' but got '%s'", expect, got)
				}
			}
		}

		if expect, got := 2, len(gd.Outputs()); expect != got {
			t.Errorf("expect length of GpioDevices->gpio0->Outputs to be %d but got %d", expect, got)
		} else {
			{
				in := gd.Outputs()[0]

				if expect, got := "Relay0", in.Name(); expect != got {
					t.Errorf("expect Name of first GpioDevices->gpio0->Outputs to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "GPIO4", in.Pin(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Outputs->in0->Pin to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "Relay 0", in.Description(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Outputs->in0->Description to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "Off", in.LowLabel(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Outputs->in0->LowLabel to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "On", in.HighLabel(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Outputs->in0->HighLabel to be '%s' but got '%s'", expect, got)
				}
			}
			{
				in := gd.Outputs()[1]

				if expect, got := "Relay1", in.Name(); expect != got {
					t.Errorf("expect Name of second GpioDevices->gpio0->Outputs to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "GPIO5", in.Pin(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Outputs->in0->Pin to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "Relay 1", in.Description(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Outputs->in0->Description to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "low", in.LowLabel(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Outputs->in0->LowLabel to be '%s' but got '%s'", expect, got)
				}

				if expect, got := "high", in.HighLabel(); expect != got {
					t.Errorf("expect GpioDevices->gpio0->Outputs->in0->HighLabel to be '%s' but got '%s'", expect, got)
				}
			}
		}
	}

	if expect, got := 1, len(config.HttpDevices()); expect != got {
		t.Errorf("expect length of config.HttpDevices to be %d but got %d", expect, got)
	} else {
		hd := config.HttpDevices()[0]

		if expect, got := "tcw241", hd.Name(); expect != got {
			t.Errorf("expect Name of first HttpDevice to be '%s' but got '%s'", expect, got)
		}

		if expect, got := time.Minute, hd.RestartInterval(); expect != got {
			t.Errorf("expect HttpDevices->tcw241->General->RestartInterval to be %s but got %s", expect, got)
		}

		if expect, got := 2*time.Minute, hd.RestartIntervalMaxBackoff(); expect != got {
			t.Errorf("expect HttpDevices->tcw241->General->RestartIntervalMaxBackoff to be %s but got %s", expect, got)
		}

		if hd.LogDebug() {
			t.Error("expect HttpDevices->tcw241->General->LogDebug to be false")
		}

		if hd.LogComDebug() {
			t.Error("expect HttpDevices->tcw241->General->LogComDebug to be false")
		}

		if expect, got := "http://control0/", hd.Url().String(); expect != got {
			t.Errorf("expect HttpDevices->tcw241->Url to be '%s' but got '%s'", expect, got)
		}

		if expect, got := types.HttpTeracomKind, hd.Kind(); expect != got {
			t.Errorf("expect HttpDevices->tcw241->Kind to be %s but got %s", expect, got)
		}

		if expect, got := "admin", hd.Username(); expect != got {
			t.Errorf("expect HttpDevices->tcw241->Username to be '%s' but got '%s'", expect, got)
		}

		if expect, got := "my-secret", hd.Password(); expect != got {
			t.Errorf("expect HttpDevices->tcw241->Password to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 5*time.Second, hd.PollInterval(); expect != got {
			t.Errorf("expect HttpDevices->tcw241->PollInterval to be %s but got %s", expect, got)
		}
	}

	if expect, got := 1, len(config.MqttDevices()); expect != got {
		t.Errorf("expect length of config.MqttDevices to be %d but got %d", expect, got)
	} else {
		vd := config.MqttDevices()[0]

		if expect, got := "bmv1", vd.Name(); expect != got {
			t.Errorf("expect Name of first MqttDevice to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 50*time.Millisecond, vd.RestartInterval(); expect != got {
			t.Errorf("expect MqttDevices->bmv1->General->RestartInterval to be %s but got %s", expect, got)
		}

		if expect, got := 30*time.Second, vd.RestartIntervalMaxBackoff(); expect != got {
			t.Errorf("expect MqttDevices->bmv1->General->RestartIntervalMaxBackoff to be %s but got %s", expect, got)
		}

		if vd.LogDebug() {
			t.Error("expect MqttDevices->bmv1->General->LogDebug to be false")
		}

		if !vd.LogComDebug() {
			t.Error("expect MqttDevices->bmv1->General->LogComDebug to be true")
		}

		if expect, got := types.MqttGoIotdeviceV3Kind, vd.Kind(); expect != got {
			t.Errorf("expect MqttDevices->bmv1->Kind to be %v but got %v", expect, got)
		}
	}

	if expect, got := 1, len(config.GensetDevices()); expect != got {
		t.Errorf("expect length of config.GensetDevices to be %d but got %d", expect, got)
	} else {
		gd := config.GensetDevices()[0]

		if expect, got := "genset0", gd.Name(); expect != got {
			t.Errorf("expect Name of first GensetDevice to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 60*time.Millisecond, gd.RestartInterval(); expect != got {
			t.Errorf("expect GensetDevice->genset0->General->RestartInterval to be %s but got %s", expect, got)
		}

		if expect, got := 3*time.Minute, gd.RestartIntervalMaxBackoff(); expect != got {
			t.Errorf("expect GensetDevice->genset0->General->RestartIntervalMaxBackoff to be %s but got %s", expect, got)
		}

		if gd.LogDebug() {
			t.Error("expect GensetDevice->genset0->General->LogDebug to be false")
		}

		if !gd.LogComDebug() {
			t.Error("expect GensetDevice->genset0->General->LogComDebug to be true")
		}

		{
			ib := gd.InputBindings()
			if expect, got := 4, len(ib); expect != got {
				t.Errorf("expect GensetDevice->genset0->InputBindings to have %d items but got %d", expect, got)
			} else {
				{
					b := ib[0]
					if expect, got := "tcw241", b.DeviceName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->InputBindings->0->DeviceName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "Available", b.RegisterName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->InputBindings->0->RegisterName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "IOAvailable", b.Name(); expect != got {
						t.Errorf("expect GensetDevice->genset0->InputBindings->0->Name to be '%s' but got '%s'", expect, got)
					}
				}
				{
					b := ib[1]
					if expect, got := "tcw241", b.DeviceName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->InputBindings->1->DeviceName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "DI0", b.RegisterName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->InputBindings->1->RegisterName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "ArmSwitch", b.Name(); expect != got {
						t.Errorf("expect GensetDevice->genset0->InputBindings->1->Name to be '%s' but got '%s'", expect, got)
					}
				}
				{
					b := ib[2]
					if expect, got := "tcw241", b.DeviceName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->InputBindings->2->DeviceName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "DI1", b.RegisterName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->InputBindings->2->RegisterName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "ResetSwitch", b.Name(); expect != got {
						t.Errorf("expect GensetDevice->genset0->InputBindings->2->Name to be '%s' but got '%s'", expect, got)
					}
				}
				{
					b := ib[3]
					if expect, got := "tcw241", b.DeviceName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->InputBindings->3->DeviceName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "DI2", b.RegisterName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->InputBindings->3->RegisterName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "FireDetected", b.Name(); expect != got {
						t.Errorf("expect GensetDevice->genset0->InputBindings->3->Name to be '%s' but got '%s'", expect, got)
					}
				}
			}
		}

		{
			ob := gd.OutputBindings()
			if expect, got := 5, len(ob); expect != got {
				t.Errorf("expect GensetDevice->genset0->OutputBindings to have %d items but got %d", expect, got)
			} else {
				{
					b := ob[0]
					if expect, got := "modbus-rtu0", b.DeviceName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->0->DeviceName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "CH0", b.RegisterName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->0->RegisterName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "Ignition", b.Name(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->0->Name to be '%s' but got '%s'", expect, got)
					}
				}
				{
					b := ob[1]
					if expect, got := "modbus-rtu0", b.DeviceName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->1->DeviceName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "CH1", b.RegisterName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->1->RegisterName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "Starter", b.Name(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->1->Name to be '%s' but got '%s'", expect, got)
					}
				}
				{
					b := ob[2]
					if expect, got := "modbus-rtu0", b.DeviceName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->2->DeviceName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "CH2", b.RegisterName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->2->RegisterName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "Fan", b.Name(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->2->Name to be '%s' but got '%s'", expect, got)
					}
				}
				{
					b := ob[3]
					if expect, got := "modbus-rtu0", b.DeviceName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->3->DeviceName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "CH3", b.RegisterName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->3->RegisterName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "Pump", b.Name(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->3->Name to be '%s' but got '%s'", expect, got)
					}
				}
				{
					b := ob[4]
					if expect, got := "modbus-rtu0", b.DeviceName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->4->DeviceName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "CH4", b.RegisterName(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->4->RegisterName to be '%s' but got '%s'", expect, got)
					}
					if expect, got := "Load", b.Name(); expect != got {
						t.Errorf("expect GensetDevice->genset0->OutputBindings->4->Name to be '%s' but got '%s'", expect, got)
					}
				}
			}
		}

		if expect, got := 20*time.Second, gd.PrimingTimeout(); expect != got {
			t.Errorf("expect GensetDevice->genset0->PrimingTimeout to be %s but got %s", expect, got)
		}

		if expect, got := 19*time.Second, gd.CrankingTimeout(); expect != got {
			t.Errorf("expect GensetDevice->genset0->CrankingTimeout to be %s but got %s", expect, got)
		}

		if expect, got := 18*time.Second, gd.StabilizingTimeout(); expect != got {
			t.Errorf("expect GensetDevice->genset0->StabilizingTimeout to be %s but got %s", expect, got)
		}

		if expect, got := 17*time.Second, gd.WarmUpTimeout(); expect != got {
			t.Errorf("expect GensetDevice->genset0->WarmUpTimeout to be %s but got %s", expect, got)
		}

		if expect, got := 16*time.Second, gd.WarmUpMinTime(); expect != got {
			t.Errorf("expect GensetDevice->genset0->WarmUpMinTime to be %s but got %s", expect, got)
		}

		if expect, got := 110.0, gd.WarmUpTemp(); expect != got {
			t.Errorf("expect GensetDevice->genset0->WarmUpTemp to be %f but got %f", expect, got)
		}

		if expect, got := 15*time.Second, gd.EngineCoolDownTimeout(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EngineCoolDownTimeout to be %s but got %s", expect, got)
		}

		if expect, got := 14*time.Second, gd.EngineCoolDownMinTime(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EngineCoolDownMinTime to be %s but got %s", expect, got)
		}

		if expect, got := 120.0, gd.EngineCoolDownTemp(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EngineCoolDownTemp to be %f but got %f", expect, got)
		}

		if expect, got := 13*time.Second, gd.EnclosureCoolDownTimeout(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EnclosureCoolDownTimeout to be %s but got %s", expect, got)
		}

		if expect, got := 12*time.Second, gd.EnclosureCoolDownMinTime(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EnclosureCoolDownMinTime to be %s but got %s", expect, got)
		}

		if expect, got := 130.0, gd.EnclosureCoolDownTemp(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EnclosureCoolDownTemp to be %f but got %f", expect, got)
		}

		if expect, got := 140.0, gd.EngineTempMin(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EngineTempMin to be %f but got %f", expect, got)
		}

		if expect, got := 150.0, gd.EngineTempMax(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EngineTempMax to be %f but got %f", expect, got)
		}

		if expect, got := 160.0, gd.AuxTemp0Min(); expect != got {
			t.Errorf("expect GensetDevice->genset0->AuxTemp0Min to be %f but got %f", expect, got)
		}

		if expect, got := 170.0, gd.AuxTemp0Max(); expect != got {
			t.Errorf("expect GensetDevice->genset0->AuxTemp0Max to be %f but got %f", expect, got)
		}

		if expect, got := 180.0, gd.AuxTemp1Min(); expect != got {
			t.Errorf("expect GensetDevice->genset0->AuxTemp1Min to be %f but got %f", expect, got)
		}

		if expect, got := 190.0, gd.AuxTemp1Max(); expect != got {
			t.Errorf("expect GensetDevice->genset0->AuxTemp1Max to be %f but got %f", expect, got)
		}

		if !gd.SinglePhase() {
			t.Error("expect GensetDevice->genset0->SinglePhase to be true")
		}

		if expect, got := 200.0, gd.UMin(); expect != got {
			t.Errorf("expect GensetDevice->genset0->UMin to be %f but got %f", expect, got)
		}

		if expect, got := 210.0, gd.UMax(); expect != got {
			t.Errorf("expect GensetDevice->genset0->UMax to be %f but got %f", expect, got)
		}

		if expect, got := 2, gd.UAvgWindow(); expect != got {
			t.Errorf("expect GensetDevice->genset0->UAvgWindow to be %d but got %d", expect, got)
		}

		if expect, got := 220.0, gd.FMin(); expect != got {
			t.Errorf("expect GensetDevice->genset0->FMin to be %f but got %f", expect, got)
		}

		if expect, got := 230.0, gd.FMax(); expect != got {
			t.Errorf("expect GensetDevice->genset0->FMax to be %f but got %f", expect, got)
		}

		if expect, got := 2, gd.FAvgWindow(); expect != got {
			t.Errorf("expect GensetDevice->genset0->FAvgWindow to be %d but got %d", expect, got)
		}

		if expect, got := 240.0, gd.PMax(); expect != got {
			t.Errorf("expect GensetDevice->genset0->PMax to be %f but got %f", expect, got)
		}

		if expect, got := 250.0, gd.PTotMax(); expect != got {
			t.Errorf("expect GensetDevice->genset0->PTotMax to be %f but got %f", expect, got)
		}
	}

	if expect, got := 2, len(config.Views()); expect != got {
		t.Errorf("expect length of config.Views to be %d but got %d", expect, got)
	} else {
		{
			v := config.Views()[0]

			if expect, got := "private", v.Name(); expect != got {
				t.Errorf("expect Name of first View to be '%s' but got '%s'", expect, got)
			}

			if expect, got := "Private", v.Title(); expect != got {
				t.Errorf("expect Views->private->Title to be '%s' but got '%s'", expect, got)
			}

			if v.Autoplay() {
				t.Error("expect Views->private->Autoplay to be false")
			}

			if !v.Hidden() {
				t.Error("expect Views->private->Hidden to be true")
			}

			if !v.IsAllowed("test0") || v.IsAllowed("other") {
				t.Error("expect Views->private->IsAllowed to be true only for test0")
			}

			if expect, got := 2, len(v.Devices()); expect != got {
				t.Fatalf("expect length of Views->private->Devices to be %d but got %d", expect, got)
			}

			{
				i := 0
				d := v.Devices()[i]
				if expect, got := "bmv0", d.Name(); expect != got {
					t.Errorf("expect Views->%s->Devices->%d->Name to be '%s' but got '%s'", v.Name(), i, expect, got)
				}

				if expect, got := "Battery Monitor", d.Title(); expect != got {
					t.Errorf("expect Views->%s->Devices->%d->Title to be '%s' but got '%s'", v.Name(), i, expect, got)
				}

				if expect, got := []string{"field-a", "field-b"}, d.Filter().SkipRegisters(); !reflect.DeepEqual(expect, got) {
					t.Errorf("expect Views->%s->Devices->%d->Filter->SkipRegisters to be %v but got %v", v.Name(), i, expect, got)
				}

				if expect, got := []string{"cat-a", "cat-b", "cat-c"}, d.Filter().SkipCategories(); !reflect.DeepEqual(expect, got) {
					t.Errorf("expect Views->%s->Devices->%d->Filter->SkipCategories to be %v but got %v", v.Name(), i, expect, got)
				}
			}

			{
				i := 1
				d := v.Devices()[i]
				if expect, got := "modbus-rtu0", d.Name(); expect != got {
					t.Errorf("expect Views->%s->Devices->%d->Name to be '%s' but got '%s'", v.Name(), i, expect, got)
				}

				if expect, got := "Relay Board", d.Title(); expect != got {
					t.Errorf("expect Views->%s->Devices->%d->Title to be '%s' but got '%s'", v.Name(), i, expect, got)
				}
			}

		}

		{
			v := config.Views()[1]

			if expect, got := "public", v.Name(); expect != got {
				t.Errorf("expect Name of second View to be '%s' but got '%s'", expect, got)
			}

			if expect, got := "Public", v.Title(); expect != got {
				t.Errorf("expect Views->public->Title to be '%s' but got '%s'", expect, got)
			}

			if expect, got := 1, len(v.Devices()); expect != got {
				t.Fatalf("expect length of Views->public->Devices to be %d but got %d", expect, got)
			}

			{
				i := 0
				d := v.Devices()[i]
				if expect, got := "bmv0", d.Name(); expect != got {
					t.Errorf("expect Views->%s->Devices->%d->Name to be '%s' but got '%s'", v.Name(), i, expect, got)
				}

				if expect, got := "Bmv 0", d.Title(); expect != got {
					t.Errorf("expect Views->%s->Devices->%d->Title to be '%s' but got '%s'", v.Name(), i, expect, got)
				}
			}
		}
	}
}

func TestReadConfig_Default(t *testing.T) {
	config, err := ReadConfig([]byte(ValidDefaultConfig), true)
	if len(err) > 0 {
		t.Errorf("did not expect any errors, got: %v", err)
	}

	config2, err := ReadConfig([]byte(ValidDefaultConfig), true)
	if len(err) > 0 {
		t.Errorf("did not expect any errors, got: %v", err)
	}

	// General Section
	if expect, got := 2, config.Version(); expect != got {
		t.Errorf("expect Version to be %d but got %d", expect, got)
	}

	if expect, got := "go-iotdevice", config.ProjectTitle(); expect != got {
		t.Errorf("expect ProjectTitle to be '%s' but got '%s'", expect, got)
	}

	if !config.LogConfig() {
		t.Errorf("expect LogConfig to be true")
	}

	if !config.LogWorkerStart() {
		t.Errorf("expect LogWorkerStart to be true")
	}

	if config.LogStateStorageDebug() {
		t.Errorf("expect LogStateStorageDebug to be false")
	}

	if config.LogCommandStorageDebug() {
		t.Errorf("expect LogCommandStorageDebug to be false")
	}

	{
		hs := config.HttpServer()
		if !hs.Enabled() {
			t.Error("expect HttpServer->Enabled to be true")
		}

		if expect, got := "[::1]", hs.Bind(); expect != got {
			t.Errorf("expect HttpServer->Bind to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 8000, hs.Port(); expect != got {
			t.Errorf("expect HttpServer->Port to be %d but got %d", expect, got)
		}

		if !hs.LogRequests() {
			t.Error("expect HttpServer->LogRequests to be true")
		}

		if got := hs.FrontendProxy(); got != nil {
			t.Errorf("expect HttpServer->FrontendProxy to be unset but got %v", got)
		}

		if expect, got := "./frontend-build/", hs.FrontendPath(); expect != got {
			t.Errorf("expect HttpServer->FrontendPath to be '%s' but got '%s'", expect, got)
		}

		if expect, got := "./docs/", hs.DocsPath(); expect != got {
			t.Errorf("expect HttpServer->DocsPath to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 5*time.Minute, hs.FrontendExpires(); expect != got {
			t.Errorf("expect HttpServer->FrontendExpires to be %s but got %s", expect, got)
		}

		if expect, got := time.Minute, hs.ConfigExpires(); expect != got {
			t.Errorf("expect HttpServer->ConfigExpires to be %s but got %s", expect, got)
		}

		if hs.LogDebug() {
			t.Error("expect HttpServer->LogDebug to be false")
		}
	}

	{
		a := config.Authentication()

		if !a.Enabled() {
			t.Error("expect Authentication->Enabled to be true")
		}

		if got0, got1 := string(a.JwtSecret()), string(config2.Authentication().JwtSecret()); got0 == got1 {
			t.Errorf("expect Authentication->JwtSecret to be different when rereading the config, got '%s' and '%s'", got0, got1)
		}

		if expect, got := time.Hour, a.JwtValidityPeriod(); expect != got {
			t.Errorf("expect Authentication->JwtValidityPeriod to be %s but got %s", expect, got)
		}

		if expect, got := "./my-auth.passwd", a.HtaccessFile(); expect != got {
			t.Errorf("expect Authentication->HtaccessFile to be '%s' but got '%s'", expect, got)
		}
	}

	if expect, got := 1, len(config.MqttClients()); expect != got {
		t.Errorf("expect length of config.MqttClients to be %d but got %d", expect, got)
	} else {
		mc := config.MqttClients()[0]

		if expect, got := "0-local", mc.Name(); expect != got {
			t.Errorf("expect Name of first MqttClient to be '%s' but got '%s'", expect, got)
		}

		if expect, got := "tcp://mqtt.example.com:1883", mc.Broker().String(); expect != got {
			t.Errorf("expect MqttClients->0-local->Broker to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 5, mc.ProtocolVersion(); expect != got {
			t.Errorf("expect MqttClients->0-local->ProtocolVersion to be %d but got %d", expect, got)
		}

		if expect, got := "", mc.User(); expect != got {
			t.Errorf("expect MqttClients->0-local->User to be '%s' but got '%s'", expect, got)
		}

		if expect, got := "", mc.Password(); expect != got {
			t.Errorf("expect MqttClients->0-local->Password to be '%s' but got '%s'", expect, got)
		}

		if got0, got1 := mc.ClientId(), config2.MqttClients()[0].ClientId(); got0 == got1 {
			t.Errorf("expect MqttClients->0-local->ClientId to be different when rereading the config, got '%s' and '%s'", got0, got1)
		}

		if expect, got := time.Minute, mc.KeepAlive(); expect != got {
			t.Errorf("expect MqttClients->0-local->KeepAlive to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 10*time.Second, mc.ConnectRetryDelay(); expect != got {
			t.Errorf("expect MqttClients->0-local->ConnectRetryDelay to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 5*time.Second, mc.ConnectTimeout(); expect != got {
			t.Errorf("expect MqttClients->0-local->ConnectTimeout to be '%s' but got '%s'", expect, got)
		}

		if expect, got := "go-iotdevice/", mc.TopicPrefix(); expect != got {
			t.Errorf("expect MqttClients->0-local->TopicPrefix to be '%s' but got '%s'", expect, got)
		}

		if got := mc.ReadOnly(); got {
			t.Errorf("expect MqttClients->0-local->ReadOnly to be false")
		}

		if expect, got := 512, mc.MaxBacklogSize(); expect != got {
			t.Errorf("expect MqttClients->0-local->MaxBacklogSize to be %d but got %d", expect, got)
		}

		{
			mcSect := mc.AvailabilityClient()
			prefix := "MqttClients->0-local->AvailabilityClient"

			{
				expect := "go-iotdevice/avail/%ClientId%"
				expect = strings.Replace(expect, "%ClientId%", mc.ClientId(), 1)
				if got := mc.AvailabilityClientTopic(); expect != got {
					t.Errorf("expect %s->Topic to be '%s' but got '%s'", prefix, expect, got)
				}
			}

			if got := mcSect.Enabled(); !got {
				t.Errorf("expect %s->Enabled to be true", prefix)
			}

			if expect, got := "%Prefix%avail/%ClientId%", mcSect.TopicTemplate(); expect != got {
				t.Errorf("expect %s->TopicTemplate to be '%s' but got '%s'", prefix, expect, got)
			}

			if expect, got := 0*time.Second, mcSect.Interval(); expect != got {
				t.Errorf("expect %s->Interval to be '%s' but got '%s'", prefix, expect, got)
			}

			if got := mcSect.Retain(); !got {
				t.Errorf("expect %s->Retain to be true", prefix)
			}

			if expect, got := byte(1), mcSect.Qos(); expect != got {
				t.Errorf("expect %s->Qos to be %d but got %d", prefix, expect, got)
			}

			if expect, got := []string{}, getNames(mcSect.Devices()); !reflect.DeepEqual(expect, got) {
				t.Errorf("expect %s->Devices to be %v but got %v", prefix, expect, got)
			}
		}

		{
			mcSect := mc.AvailabilityDevice()
			prefix := "MqttClients->0-local->AvailabilityDevice"

			if expect, got := "go-iotdevice/avail/my-dev", mc.AvailabilityDeviceTopic("my-dev"); expect != got {
				t.Errorf("expect %s->AvailabilityDeviceTopic to be '%s' but got '%s'", prefix, expect, got)
			}

			if got := mcSect.Enabled(); !got {
				t.Errorf("expect %s->Enabled to be true", prefix)
			}

			if expect, got := "%Prefix%avail/%DeviceName%", mcSect.TopicTemplate(); expect != got {
				t.Errorf("expect %s->TopicTemplate to be '%s' but got '%s'", prefix, expect, got)
			}

			if expect, got := 0*time.Second, mcSect.Interval(); expect != got {
				t.Errorf("expect %s->Interval to be '%s' but got '%s'", prefix, expect, got)
			}

			if got := mcSect.Retain(); !got {
				t.Errorf("expect %s->Retain to be true", prefix)
			}

			if expect, got := byte(1), mcSect.Qos(); expect != got {
				t.Errorf("expect %s->Qos to be %d but got %d", prefix, expect, got)
			}

			if expect, got := []string{"bmv0", "genset0", "gpio0", "modbus-rtu0", "tcw241"}, getNames(mcSect.Devices()); !reflect.DeepEqual(expect, got) {
				t.Errorf("expect %s->Devices to be %v but got %v", prefix, expect, got)
			} else {
				prefix := prefix + "->Devices->bmv0"
				rf := mcSect.Devices()[0].Filter()

				if got := rf.IncludeRegisters(); len(got) > 0 {
					t.Errorf("expect %s->IncludeRegisters to be empty, got %v", prefix, got)
				}

				if got := rf.SkipRegisters(); len(got) > 0 {
					t.Errorf("expect %s->SkipRegisters to be empty but got %v", prefix, got)
				}

				if got := rf.IncludeCategories(); len(got) > 0 {
					t.Errorf("expect %s->IncludeCategories to be empty but got %v", prefix, got)
				}

				if got := rf.SkipCategories(); len(got) > 0 {
					t.Errorf("expect %s->SkipCategories to be empty but got %v", prefix, got)
				}

				if got := rf.DefaultInclude(); !got {
					t.Errorf("expect %s->DefaultInclude to be true", prefix)
				}
			}
		}

		{
			mcSect := mc.Structure()
			prefix := "MqttClients->0-local->Structure"

			if expect, got := "go-iotdevice/struct/my-dev", mc.StructureTopic("my-dev"); expect != got {
				t.Errorf("expect %s->StructureTopic to be '%s' but got '%s'", prefix, expect, got)
			}

			if got := mcSect.Enabled(); got {
				t.Errorf("expect %s->Enabled to be false", prefix)
			}

			if expect, got := "%Prefix%struct/%DeviceName%", mcSect.TopicTemplate(); expect != got {
				t.Errorf("expect %s->TopicTemplate to be '%s' but got '%s'", prefix, expect, got)
			}

			if expect, got := 0*time.Second, mcSect.Interval(); expect != got {
				t.Errorf("expect %s->Interval to be '%s' but got '%s'", prefix, expect, got)
			}

			if got := mcSect.Retain(); !got {
				t.Errorf("expect %s->Retain to be true", prefix)
			}

			if expect, got := byte(1), mcSect.Qos(); expect != got {
				t.Errorf("expect %s->Qos to be %d but got %d", prefix, expect, got)
			}

			if expect, got := []string{"bmv0", "genset0", "gpio0", "modbus-rtu0", "tcw241"}, getNames(mcSect.Devices()); !reflect.DeepEqual(expect, got) {
				t.Errorf("expect %s->Devices to be %v but got %v", prefix, expect, got)
			} else {
				prefix := prefix + "->Devices->bmv0"
				rf := mcSect.Devices()[0].Filter()

				if got := rf.IncludeRegisters(); len(got) > 0 {
					t.Errorf("expect %s->IncludeRegisters to be empty, got %v", prefix, got)
				}

				if got := rf.SkipRegisters(); len(got) > 0 {
					t.Errorf("expect %s->SkipRegisters to be empty but got %v", prefix, got)
				}

				if got := rf.IncludeCategories(); len(got) > 0 {
					t.Errorf("expect %s->IncludeCategories to be empty but got %v", prefix, got)
				}

				if got := rf.SkipCategories(); len(got) > 0 {
					t.Errorf("expect %s->SkipCategories to be empty but got %v", prefix, got)
				}

				if got := rf.DefaultInclude(); !got {
					t.Errorf("expect %s->DefaultInclude to be true", prefix)
				}
			}
		}

		{
			mcSect := mc.Telemetry()
			prefix := "MqttClients->0-local->Telemetry"

			if expect, got := "go-iotdevice/tele/my-dev", mc.TelemetryTopic("my-dev"); expect != got {
				t.Errorf("expect %s->TelemetryTopic to be '%s' but got '%s'", prefix, expect, got)
			}

			if got := mcSect.Enabled(); got {
				t.Errorf("expect %s->Enabled to be false", prefix)
			}

			if expect, got := "%Prefix%tele/%DeviceName%", mcSect.TopicTemplate(); expect != got {
				t.Errorf("expect %s->TopicTemplate to be '%s' but got '%s'", prefix, expect, got)
			}

			if expect, got := 1*time.Second, mcSect.Interval(); expect != got {
				t.Errorf("expect %s->Interval to be '%s' but got '%s'", prefix, expect, got)
			}

			if got := mcSect.Retain(); got {
				t.Errorf("expect %s->Retain to be false", prefix)
			}

			if expect, got := byte(1), mcSect.Qos(); expect != got {
				t.Errorf("expect %s->Qos to be %d but got %d", prefix, expect, got)
			}

			if expect, got := []string{"bmv0", "genset0", "gpio0", "modbus-rtu0", "tcw241"}, getNames(mcSect.Devices()); !reflect.DeepEqual(expect, got) {
				t.Errorf("expect %s->Devices to be %v but got %v", prefix, expect, got)
			} else {
				prefix := prefix + "->Devices->bmv0"
				rf := mcSect.Devices()[0].Filter()

				if got := rf.IncludeRegisters(); len(got) > 0 {
					t.Errorf("expect %s->IncludeRegisters to be empty, got %v", prefix, got)
				}

				if got := rf.SkipRegisters(); len(got) > 0 {
					t.Errorf("expect %s->SkipRegisters to be empty but got %v", prefix, got)
				}

				if got := rf.IncludeCategories(); len(got) > 0 {
					t.Errorf("expect %s->IncludeCategories to be empty but got %v", prefix, got)
				}

				if got := rf.SkipCategories(); len(got) > 0 {
					t.Errorf("expect %s->SkipCategories to be empty but got %v", prefix, got)
				}

				if got := rf.DefaultInclude(); !got {
					t.Errorf("expect %s->DefaultInclude to be true", prefix)
				}
			}
		}

		{
			mcSect := mc.Realtime()
			prefix := "MqttClients->0-local->Realtime"

			if expect, got := "go-iotdevice/real/my-dev/reg-name", mc.RealtimeTopic("my-dev", "reg-name"); expect != got {
				t.Errorf("expect %s->RealtimeTopic to be '%s' but got '%s'", prefix, expect, got)
			}

			if got := mcSect.Enabled(); got {
				t.Errorf("expect %s->Enabled to be false", prefix)
			}

			if expect, got := "%Prefix%real/%DeviceName%/%RegisterName%", mcSect.TopicTemplate(); expect != got {
				t.Errorf("expect %s->TopicTemplate to be '%s' but got '%s'", prefix, expect, got)
			}

			if expect, got := 0*time.Second, mcSect.Interval(); expect != got {
				t.Errorf("expect %s->Interval to be '%s' but got '%s'", prefix, expect, got)
			}

			if got := mcSect.Retain(); got {
				t.Errorf("expect %s->Retain to be false", prefix)
			}

			if expect, got := byte(1), mcSect.Qos(); expect != got {
				t.Errorf("expect %s->Qos to be %d but got %d", prefix, expect, got)
			}

			if expect, got := []string{"bmv0", "genset0", "gpio0", "modbus-rtu0", "tcw241"}, getNames(mcSect.Devices()); !reflect.DeepEqual(expect, got) {
				t.Errorf("expect %s->Devices to be %v but got %v", prefix, expect, got)
			} else {
				prefix := prefix + "->Devices->bmv0"
				rf := mcSect.Devices()[0].Filter()

				if got := rf.IncludeRegisters(); len(got) > 0 {
					t.Errorf("expect %s->IncludeRegisters to be empty, got %v", prefix, got)
				}

				if got := rf.SkipRegisters(); len(got) > 0 {
					t.Errorf("expect %s->SkipRegisters to be empty but got %v", prefix, got)
				}

				if got := rf.IncludeCategories(); len(got) > 0 {
					t.Errorf("expect %s->IncludeCategories to be empty but got %v", prefix, got)
				}

				if got := rf.SkipCategories(); len(got) > 0 {
					t.Errorf("expect %s->SkipCategories to be empty but got %v", prefix, got)
				}

				if got := rf.DefaultInclude(); !got {
					t.Errorf("expect %s->DefaultInclude to be true", prefix)
				}
			}
		}

		{
			mcSect := mc.HomeassistantDiscovery()
			prefix := "MqttClients->0-local->HomeassistantDiscovery"

			if expect, got := "homeassistant/a/b/c/config", mc.HomeassistantDiscoveryTopic("a", "b", "c"); expect != got {
				t.Errorf("expect %s->HomeassistantDiscoveryTopic to be '%s' but got '%s'", prefix, expect, got)
			}

			if got := mcSect.Enabled(); got {
				t.Errorf("expect %s->Enabled to be false", prefix)
			}

			if expect, got := "homeassistant/%Component%/%NodeId%/%ObjectId%/config", mcSect.TopicTemplate(); expect != got {
				t.Errorf("expect %s->TopicTemplate to be '%s' but got '%s'", prefix, expect, got)
			}

			if expect, got := 0*time.Second, mcSect.Interval(); expect != got {
				t.Errorf("expect %s->Interval to be '%s' but got '%s'", prefix, expect, got)
			}

			if got := mcSect.Retain(); got {
				t.Errorf("expect %s->Retain to be false", prefix)
			}

			if expect, got := byte(1), mcSect.Qos(); expect != got {
				t.Errorf("expect %s->Qos to be %d but got %d", prefix, expect, got)
			}

			if expect, got := []string{"bmv0", "genset0", "gpio0", "modbus-rtu0", "tcw241"}, getNames(mcSect.Devices()); !reflect.DeepEqual(expect, got) {
				t.Errorf("expect %s->Devices to be %v but got %v", prefix, expect, got)
			} else {
				prefix := prefix + "->Devices->bmv0"
				rf := mcSect.Devices()[0].Filter()

				if got := rf.IncludeRegisters(); len(got) > 0 {
					t.Errorf("expect %s->IncludeRegisters to be empty, got %v", prefix, got)
				}

				if got := rf.SkipRegisters(); len(got) > 0 {
					t.Errorf("expect %s->SkipRegisters to be empty but got %v", prefix, got)
				}

				if got := rf.IncludeCategories(); len(got) > 0 {
					t.Errorf("expect %s->IncludeCategories to be empty but got %v", prefix, got)
				}

				if got := rf.SkipCategories(); len(got) > 0 {
					t.Errorf("expect %s->SkipCategories to be empty but got %v", prefix, got)
				}

				if got := rf.DefaultInclude(); !got {
					t.Errorf("expect %s->DefaultInclude to be true", prefix)
				}
			}
		}

		{
			mcSect := mc.Command()
			prefix := "MqttClients->0-local->Command"

			if expect, got := "go-iotdevice/cmnd/my-dev/reg-name", mc.CommandTopic("my-dev", "reg-name"); expect != got {
				t.Errorf("expect %s->CommandTopic to be '%s' but got '%s'", prefix, expect, got)
			}

			if got := mcSect.Enabled(); got {
				t.Errorf("expect %s->Enabled to be false", prefix)
			}

			if expect, got := "%Prefix%cmnd/%DeviceName%/%RegisterName%", mcSect.TopicTemplate(); expect != got {
				t.Errorf("expect %s->TopicTemplate to be '%s' but got '%s'", prefix, expect, got)
			}

			if expect, got := 0*time.Second, mcSect.Interval(); expect != got {
				t.Errorf("expect %s->Interval to be '%s' but got '%s'", prefix, expect, got)
			}

			if got := mcSect.Retain(); got {
				t.Errorf("expect %s->Retain to be false", prefix)
			}

			if expect, got := byte(1), mcSect.Qos(); expect != got {
				t.Errorf("expect %s->Qos to be %d but got %d", prefix, expect, got)
			}

			if expect, got := []string{"bmv0", "genset0", "gpio0", "modbus-rtu0", "tcw241"}, getNames(mcSect.Devices()); !reflect.DeepEqual(expect, got) {
				t.Errorf("expect %s->Devices to be %v but got %v", prefix, expect, got)
			} else {
				prefix := prefix + "->Devices->bmv0"
				rf := mcSect.Devices()[0].Filter()

				if got := rf.IncludeRegisters(); len(got) > 0 {
					t.Errorf("expect %s->IncludeRegisters to be empty, got %v", prefix, got)
				}

				if got := rf.SkipRegisters(); len(got) > 0 {
					t.Errorf("expect %s->SkipRegisters to be empty but got %v", prefix, got)
				}

				if got := rf.IncludeCategories(); len(got) > 0 {
					t.Errorf("expect %s->IncludeCategories to be empty but got %v", prefix, got)
				}

				if got := rf.SkipCategories(); len(got) > 0 {
					t.Errorf("expect %s->SkipCategories to be empty but got %v", prefix, got)
				}

				if got := rf.DefaultInclude(); !got {
					t.Errorf("expect %s->DefaultInclude to be true", prefix)
				}
			}
		}

		if got := mc.LogDebug(); got {
			t.Error("expect MqttClients->0-local->LogDebug to be false")
		}

		if got := mc.LogMessages(); got {
			t.Error("expect MqttClients->0-local->LogMessages to be false")
		}
	}

	if expect, got := 1, len(config.Modbus()); expect != got {
		t.Errorf("expect length of config.Modbus to be %d but got %d", expect, got)
	} else {
		mb := config.Modbus()[0]

		if expect, got := "bus0", mb.Name(); expect != got {
			t.Errorf("expect Name of first Modbus to be '%s' but got '%s'", expect, got)
		}

		if expect, got := "/dev/ttyUSB1", mb.Device(); expect != got {
			t.Errorf("expect Modbus->bus0->Device to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 1200, mb.BaudRate(); expect != got {
			t.Errorf("expect Modbus->bus0->BaudRate to be %d but got %d", expect, got)
		}

		if expect, got := 100*time.Millisecond, mb.ReadTimeout(); expect != got {
			t.Errorf("expect Modbus->bus0->ReadTimeout to be %s but got %s", expect, got)
		}

		if mb.LogDebug() {
			t.Error("expect Modbus->bus0->LogDebug to be false")
		}
	}

	if expect, got := 1, len(config.VictronDevices()); expect != got {
		t.Errorf("expect length of config.VictronDevices to be %d but got %d", expect, got)
	} else {
		vd := config.VictronDevices()[0]

		if expect, got := "bmv0", vd.Name(); expect != got {
			t.Errorf("expect Name of first VictronDevice to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 200*time.Millisecond, vd.RestartInterval(); expect != got {
			t.Errorf("expect VictronDevices->bmv0->General->RestartInterval to be %s but got %s", expect, got)
		}

		if expect, got := time.Minute, vd.RestartIntervalMaxBackoff(); expect != got {
			t.Errorf("expect VictronDevices->bmv0->General->RestartIntervalMaxBackoff to be %s but got %s", expect, got)
		}

		if vd.LogDebug() {
			t.Error("expect VictronDevices->bmv0->General->LogDebug to be false")
		}

		if vd.LogComDebug() {
			t.Error("expect VictronDevices->bmv0->General->LogComDebug to be false")
		}

		if expect, got := "", vd.Device(); expect != got {
			t.Errorf("expect VictronDevices->bmv0->Device to be '%s' but got '%s'", expect, got)
		}

		if expect, got := types.VictronRandomBmvKind, vd.Kind(); expect != got {
			t.Errorf("expect VictronDevices->bmv0->Kind to be %s but got %s", expect, got)
		}

		if expect, got := "", vd.IoLog(); expect != got {
			t.Errorf("expect VictronDevices->bmv0->IoLog to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 500*time.Millisecond, vd.PollInterval(); expect != got {
			t.Errorf("expect VictronDevices->bmv0->PollInterval to be %s but got %s", expect, got)
		}
	}

	if expect, got := 1, len(config.ModbusDevices()); expect != got {
		t.Errorf("expect length of config.ModbusDevices to be %d but got %d", expect, got)
	} else {
		md := config.ModbusDevices()[0]

		if expect, got := "modbus-rtu0", md.Name(); expect != got {
			t.Errorf("expect Name of first ModbusDevice to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 200*time.Millisecond, md.RestartInterval(); expect != got {
			t.Errorf("expect ModbusDevices->modbus-rtu0->General->RestartInterval to be %s but got %s", expect, got)
		}

		if expect, got := time.Minute, md.RestartIntervalMaxBackoff(); expect != got {
			t.Errorf("expect ModbusDevices->modbus-rtu0->General->RestartIntervalMaxBackoff to be %s but got %s", expect, got)
		}

		if md.LogDebug() {
			t.Error("expect ModbusDevices->modbus-rtu0->General->LogDebug to be false")
		}

		if md.LogComDebug() {
			t.Error("expect ModbusDevices->modbus-rtu0->General->LogComDebug to be false")
		}

		if expect, got := "bus0", md.Bus(); expect != got {
			t.Errorf("expect ModbusDevices->modbus-rtu0->Bus to be '%s' but got '%s'", expect, got)
		}

		if expect, got := types.ModbusWaveshareRtuRelay8Kind, md.Kind(); expect != got {
			t.Errorf("expect ModbusDevices->modbus-rtu0->Kind to be %s but got %s", expect, got)
		}

		if expect, got := byte(0x02), md.Address(); expect != got {
			t.Errorf("expect ModbusDevices->modbus-rtu0->Address to be 0x%x but got 0x%x", expect, got)
		}
	}

	if expect, got := 1, len(config.GpioDevices()); expect != got {
		t.Errorf("expect length of config.GpioDevices to be %d but got %d", expect, got)
	} else {
		gd := config.GpioDevices()[0]

		if expect, got := "gpio0", gd.Name(); expect != got {
			t.Errorf("expect Name of first GpioDevices to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 200*time.Millisecond, gd.RestartInterval(); expect != got {
			t.Errorf("expect GpioDevices->gpio0->General->RestartInterval to be %s but got %s", expect, got)
		}

		if expect, got := time.Minute, gd.RestartIntervalMaxBackoff(); expect != got {
			t.Errorf("expect GpioDevices->gpio0->General->RestartIntervalMaxBackoff to be %s but got %s", expect, got)
		}

		if gd.LogDebug() {
			t.Error("expect GpioDevices->gpio0->General->LogDebug to be false")
		}

		if gd.LogComDebug() {
			t.Error("expect GpioDevices->gpio0->General->LogComDebug to be false")
		}

		if expect, got := "gpiochip0", gd.Chip(); expect != got {
			t.Errorf("expect GpioDevices->gpio0->Chip to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 100*time.Millisecond, gd.InputDebounce(); expect != got {
			t.Errorf("expect GpioDevices->gpio0->InputDebounce to be '%s' but got '%s'", expect, got)
		}

		if got := gd.InputOptions(); len(got) != 0 {
			t.Errorf("expect GpioDevices->gpio0->InputOptions to be empty but got %v", got)
		}

		if got := gd.OutputOptions(); len(got) != 0 {
			t.Errorf("expect GpioDevices->gpio0->OutputOptions to be empty but got %v", got)
		}

		if expect, got := 1, len(gd.Inputs()); expect != got {
			t.Errorf("expect length of GpioDevices->gpio0->Inputs to be %d but got %d", expect, got)
		} else {
			in := gd.Inputs()[0]

			if expect, got := "Switch0", in.Name(); expect != got {
				t.Errorf("expect Name of first GpioDevices->gpio0->Inputs to be '%s' but got '%s'", expect, got)
			}

			if expect, got := "GPIO2", in.Pin(); expect != got {
				t.Errorf("expect GpioDevices->gpio0->Inputs->in0->Pin to be '%s' but got '%s'", expect, got)
			}

			if expect, got := "Switch 0", in.Description(); expect != got {
				t.Errorf("expect GpioDevices->gpio0->Inputs->in0->Description to be '%s' but got '%s'", expect, got)
			}

			if expect, got := "low", in.LowLabel(); expect != got {
				t.Errorf("expect GpioDevices->gpio0->Inputs->in0->LowLabel to be '%s' but got '%s'", expect, got)
			}

			if expect, got := "high", in.HighLabel(); expect != got {
				t.Errorf("expect GpioDevices->gpio0->Inputs->in0->HighLabel to be '%s' but got '%s'", expect, got)
			}
		}
	}

	if expect, got := 1, len(config.HttpDevices()); expect != got {
		t.Errorf("expect length of config.HttpDevices to be %d but got %d", expect, got)
	} else {
		hd := config.HttpDevices()[0]

		if expect, got := "tcw241", hd.Name(); expect != got {
			t.Errorf("expect Name of first HttpDevice to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 200*time.Millisecond, hd.RestartInterval(); expect != got {
			t.Errorf("expect HttpDevices->tcw241->General->RestartInterval to be %s but got %s", expect, got)
		}

		if expect, got := time.Minute, hd.RestartIntervalMaxBackoff(); expect != got {
			t.Errorf("expect HttpDevices->tcw241->General->RestartIntervalMaxBackoff to be %s but got %s", expect, got)
		}

		if hd.LogDebug() {
			t.Error("expect HttpDevices->tcw241->General->LogDebug to be false")
		}

		if hd.LogComDebug() {
			t.Error("expect HttpDevices->tcw241->General->LogComDebug to be false")
		}

		if expect, got := "http://control0/", hd.Url().String(); expect != got {
			t.Errorf("expect HttpDevices->tcw241->Url to be '%s' but got '%s'", expect, got)
		}

		if expect, got := types.HttpTeracomKind, hd.Kind(); expect != got {
			t.Errorf("expect HttpDevices->tcw241->Kind to be %s but got %s", expect, got)
		}

		if expect, got := "", hd.Username(); expect != got {
			t.Errorf("expect HttpDevices->tcw241->Username to be '%s' but got '%s'", expect, got)
		}

		if expect, got := "", hd.Password(); expect != got {
			t.Errorf("expect HttpDevices->tcw241->Password to be '%s' but got '%s'", expect, got)
		}

		if expect, got := time.Second, hd.PollInterval(); expect != got {
			t.Errorf("expect HttpDevices->tcw241->PollInterval to be %s but got %s", expect, got)
		}
	}

	if expect, got := 1, len(config.MqttDevices()); expect != got {
		t.Errorf("expect length of config.MqttDevices to be %d but got %d", expect, got)
	} else {
		vd := config.MqttDevices()[0]

		if expect, got := "bmv1", vd.Name(); expect != got {
			t.Errorf("expect Name of first MqttDevice to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 200*time.Millisecond, vd.RestartInterval(); expect != got {
			t.Errorf("expect MqttDevices->bmv1->General->RestartInterval to be %s but got %s", expect, got)
		}

		if expect, got := time.Minute, vd.RestartIntervalMaxBackoff(); expect != got {
			t.Errorf("expect MqttDevices->bmv1->General->RestartIntervalMaxBackoff to be %s but got %s", expect, got)
		}

		if vd.LogDebug() {
			t.Error("expect MqttDevices->bmv1->General->LogDebug to be false")
		}

		if vd.LogComDebug() {
			t.Error("expect MqttDevices->bmv1->General->LogComDebug to be false")
		}

		if expect, got := types.MqttGoIotdeviceV3Kind, vd.Kind(); expect != got {
			t.Errorf("expect MqttDevices->bmv1->Kind to be %v but got %v", expect, got)
		}
	}

	if expect, got := 1, len(config.GensetDevices()); expect != got {
		t.Errorf("expect length of config.GensetDevices to be %d but got %d", expect, got)
	} else {
		gd := config.GensetDevices()[0]

		if expect, got := "genset0", gd.Name(); expect != got {
			t.Errorf("expect Name of first GensetDevice to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 200*time.Millisecond, gd.RestartInterval(); expect != got {
			t.Errorf("expect GensetDevice->genset0->General->RestartInterval to be %s but got %s", expect, got)
		}

		if expect, got := time.Minute, gd.RestartIntervalMaxBackoff(); expect != got {
			t.Errorf("expect GensetDevice->genset0->General->RestartIntervalMaxBackoff to be %s but got %s", expect, got)
		}

		if gd.LogDebug() {
			t.Error("expect GensetDevice->genset0->General->LogDebug to be false")
		}

		if gd.LogComDebug() {
			t.Error("expect GensetDevice->genset0->General->LogComDebug to be false")
		}

		if expect, got := 10*time.Second, gd.PrimingTimeout(); expect != got {
			t.Errorf("expect GensetDevice->genset0->PrimingTimeout to be %s but got %s", expect, got)
		}

		if expect, got := 10*time.Second, gd.CrankingTimeout(); expect != got {
			t.Errorf("expect GensetDevice->genset0->CrankingTimeout to be %s but got %s", expect, got)
		}

		if expect, got := 3*time.Second, gd.StabilizingTimeout(); expect != got {
			t.Errorf("expect GensetDevice->genset0->StabilizingTimeout to be %s but got %s", expect, got)
		}

		if expect, got := 10*time.Minute, gd.WarmUpTimeout(); expect != got {
			t.Errorf("expect GensetDevice->genset0->WarmUpTimeout to be %s but got %s", expect, got)
		}

		if expect, got := 2*time.Minute, gd.WarmUpMinTime(); expect != got {
			t.Errorf("expect GensetDevice->genset0->WarmUpMinTime to be %s but got %s", expect, got)
		}

		if expect, got := 50.0, gd.WarmUpTemp(); expect != got {
			t.Errorf("expect GensetDevice->genset0->WarmUpTemp to be %f but got %f", expect, got)
		}

		if expect, got := 5*time.Minute, gd.EngineCoolDownTimeout(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EngineCoolDownTimeout to be %s but got %s", expect, got)
		}

		if expect, got := 2*time.Minute, gd.EngineCoolDownMinTime(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EngineCoolDownMinTime to be %s but got %s", expect, got)
		}

		if expect, got := 70.0, gd.EngineCoolDownTemp(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EngineCoolDownTemp to be %f but got %f", expect, got)
		}

		if expect, got := 10*time.Minute, gd.EnclosureCoolDownTimeout(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EnclosureCoolDownTimeout to be %s but got %s", expect, got)
		}

		if expect, got := 2*time.Minute, gd.EnclosureCoolDownMinTime(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EnclosureCoolDownMinTime to be %s but got %s", expect, got)
		}

		if expect, got := 30.0, gd.EnclosureCoolDownTemp(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EnclosureCoolDownTemp to be %f but got %f", expect, got)
		}

		if expect, got := -20.0, gd.EngineTempMin(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EngineTempMin to be %f but got %f", expect, got)
		}

		if expect, got := 90.0, gd.EngineTempMax(); expect != got {
			t.Errorf("expect GensetDevice->genset0->EngineTempMax to be %f but got %f", expect, got)
		}

		if expect, got := -20.0, gd.AuxTemp0Min(); expect != got {
			t.Errorf("expect GensetDevice->genset0->AuxTemp0Min to be %f but got %f", expect, got)
		}

		if expect, got := 120.0, gd.AuxTemp0Max(); expect != got {
			t.Errorf("expect GensetDevice->genset0->AuxTemp0Max to be %f but got %f", expect, got)
		}

		if expect, got := -20.0, gd.AuxTemp1Min(); expect != got {
			t.Errorf("expect GensetDevice->genset0->AuxTemp1Min to be %f but got %f", expect, got)
		}

		if expect, got := 120.0, gd.AuxTemp1Max(); expect != got {
			t.Errorf("expect GensetDevice->genset0->AuxTemp1Max to be %f but got %f", expect, got)
		}

		if gd.SinglePhase() {
			t.Error("expect GensetDevice->genset0->SinglePhase to be false")
		}

		if expect, got := 220.0, gd.UMin(); expect != got {
			t.Errorf("expect GensetDevice->genset0->UMin to be %f but got %f", expect, got)
		}

		if expect, got := 240.0, gd.UMax(); expect != got {
			t.Errorf("expect GensetDevice->genset0->UMax to be %f but got %f", expect, got)
		}

		if expect, got := 3, gd.UAvgWindow(); expect != got {
			t.Errorf("expect GensetDevice->genset0->UAvgWindow to be %d but got %d", expect, got)
		}

		if expect, got := 45.0, gd.FMin(); expect != got {
			t.Errorf("expect GensetDevice->genset0->FMin to be %f but got %f", expect, got)
		}

		if expect, got := 55.0, gd.FMax(); expect != got {
			t.Errorf("expect GensetDevice->genset0->FMax to be %f but got %f", expect, got)
		}

		if expect, got := 3, gd.FAvgWindow(); expect != got {
			t.Errorf("expect GensetDevice->genset0->FAvgWindow to be %d but got %d", expect, got)
		}

		if expect, got := 1e6, gd.PMax(); expect != got {
			t.Errorf("expect GensetDevice->genset0->PMax to be %f but got %f", expect, got)
		}

		if expect, got := 1e6, gd.PTotMax(); expect != got {
			t.Errorf("expect GensetDevice->genset0->PTotMax to be %f but got %f", expect, got)
		}
	}

	if expect, got := 1, len(config.Views()); expect != got {
		t.Errorf("expect length of config.Views to be %d but got %d", expect, got)
	} else {
		v := config.Views()[0]

		if expect, got := "private", v.Name(); expect != got {
			t.Errorf("expect Name of first View to be '%s' but got '%s'", expect, got)
		}

		if expect, got := "Private", v.Title(); expect != got {
			t.Errorf("expect Views->private->Title to be '%s' but got '%s'", expect, got)
		}

		if expect, got := 1, len(v.Devices()); expect != got {
			t.Errorf("expect length of Views->private->Devices to be %d but got %d", expect, got)
		} else {
			i := 0
			d := v.Devices()[i]
			if expect, got := "bmv0", d.Name(); expect != got {
				t.Errorf("expect Views->%s->Devices->%d->Name to be '%s' but got '%s'", v.Name(), i, expect, got)
			}

			if expect, got := "Battery Monitor", d.Title(); expect != got {
				t.Errorf("expect Views->%s->Devices->%d->Title to be '%s' but got '%s'", v.Name(), i, expect, got)
			}
		}
	}
}

// check that configuration files in the documentation do not contain any errors
func TestReadConfig_DocumentationFullConfig(t *testing.T) {
	_, err := ReadConfigFile("", "../../documentation/full-config.yaml", true)
	if len(err) > 0 {
		t.Errorf("did not expect any error, got %v", err)
	}
}

func TestReadConfig_DocumentationConfig(t *testing.T) {
	_, err := ReadConfigFile("", "../../documentation/config.yaml", true)
	if len(err) > 0 {
		t.Errorf("did not expect any error, got %v", err)
	}
}

func TestPrintConfig(t *testing.T) {
	config, err := ReadConfig([]byte(ValidCompleteConfig), true)
	if len(err) > 0 {
		t.Errorf("did not expect any errors, got: %v", err)
	}

	// test config output does not crash
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	if err := config.PrintConfig(); err != nil {
		t.Errorf("expect no error. Got: %s", err)
	}
	t.Log(buf.String())
}

func getNames[N Nameable](list []N) (ret []string) {
	ret = make([]string, len(list))
	for i, t := range list {
		ret[i] = t.Name()
	}
	return
}
