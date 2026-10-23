package types

type HttpDeviceKind int

const (
	HttpUndefinedKind HttpDeviceKind = iota
	HttpTeracomKind
	HttpShellyEm3Kind
)

func (dk HttpDeviceKind) String() string {
	switch dk {
	case HttpTeracomKind:
		return "Teracom"
	case HttpShellyEm3Kind:
		return "ShellyEm3"
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
