package types

type ModbusDeviceKind int

const (
	ModbusUndefinedKind ModbusDeviceKind = iota
	ModbusWaveshareRtuRelay8Kind
)

func (dk ModbusDeviceKind) String() string {
	switch dk {
	case ModbusWaveshareRtuRelay8Kind:
		return "WaveshareRtuRelay8"
	default:
		return "Undefined"
	}
}

func ModbusDeviceKindFromString(s string) ModbusDeviceKind {
	if s == "WaveshareRtuRelay8" {
		return ModbusWaveshareRtuRelay8Kind
	}

	return ModbusUndefinedKind
}
