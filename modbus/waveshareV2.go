package modbus

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
