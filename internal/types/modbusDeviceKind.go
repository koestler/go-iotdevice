package types

type ModbusDeviceKind int

const (
	ModbusUndefinedKind ModbusDeviceKind = iota
	ModbusWaveshareRtuRelay8Kind
	ModbusFinder7M38Kind
)

func (dk ModbusDeviceKind) String() string {
	switch dk {
	case ModbusWaveshareRtuRelay8Kind:
		return "WaveshareRtuRelay8"
	case ModbusFinder7M38Kind:
		return "Finder7M38"
	default:
		return "Undefined"
	}
}

func ModbusDeviceKindFromString(s string) ModbusDeviceKind {
	switch s {
	case "WaveshareRtuRelay8":
		return ModbusWaveshareRtuRelay8Kind
	case "Finder7M38":
		return ModbusFinder7M38Kind
	default:
		return ModbusUndefinedKind
	}

}
