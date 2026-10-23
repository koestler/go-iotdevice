package types

type VictronDeviceKind int

const (
	VictronUndefinedKind VictronDeviceKind = iota
	VictronRandomBmvKind
	VictronRandomSolarKind
	VictronVedirectKind
)

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
