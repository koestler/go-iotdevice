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
	switch s {
	case "RandomBmv":
		return VictronRandomBmvKind
	case "RandomSolar":
		return VictronRandomSolarKind
	case "Vedirect":
		return VictronVedirectKind
	default:
		return VictronUndefinedKind
	}
}
