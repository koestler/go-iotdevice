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
	switch s {
	case "Teracom":
		return HttpTeracomKind
	case "ShellyEm3":
		return HttpShellyEm3Kind
	default:
		return HttpUndefinedKind
	}
}
