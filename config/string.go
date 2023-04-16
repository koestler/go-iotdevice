package config

func (dk VictronDeviceKind) String() string {
	switch dk {
	case VictronRandomBmvKind:
		return "RandomBmv"
	case VictronRandomSolarKind:
		return "RandomSolar"
	case VictronVedirectKind:
		return "Vedirect"
	default:
		return "Undefined"
	}
}

func VictronDeviceKindFromString(s string) VictronDeviceKind {
	if s == "RandomBmv" {
		return VictronRandomBmvKind
	}
	if s == "RandomSolar" {
		return VictronRandomSolarKind
	}
	if s == "Vedirect" {
		return VictronVedirectKind
	}
	return VictronUndefinedKind
}

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

func (dk HttpDeviceKind) String() string {
	switch dk {
	case HttpTeracomKind:
		return "Teracom"
	case HttpShellyEm3Kind:
		return "Shelly3m"
	default:
		return "Undefined"
	}
}

func HttpDeviceKindFromString(s string) HttpDeviceKind {
	if s == "Teracom" {
		return HttpTeracomKind
	}
	if s == "ShellyEm3" {
		return HttpShellyEm3Kind
	}

	return HttpUndefinedKind
}
