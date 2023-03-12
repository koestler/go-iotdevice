package modbusDevice

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
)

type WaveshareFunction byte

const (
	WaveshareFunctionReadRelay             WaveshareFunction = 0x01
	WaveshareFunctionReadAddressAndVersion WaveshareFunction = 0x03
	WaveshareFunctionWriteRelay            WaveshareFunction = 0x05
	WaveshareFunctionSetBaudRateAndAddress WaveshareFunction = 0x06
	WaveshareFunctionWriteAllRelays        WaveshareFunction = 0x0F
)

const (
	WaveshareCommandOpen  uint16 = 0xFF00
	WaveshareCommandClose uint16 = 0x0000
	WaveshareCommandFlip  uint16 = 0x5500
)

func startWaveshareRtuRelay8(c *DeviceStruct, output dataflow.Fillable) error {
	log.Printf("device[%s]: start waveshare RTU Relay 8 source", c.deviceConfig.Name())
	return nil
}
