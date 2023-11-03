package types

type VictronDeviceKind int

const (
	VictronUndefinedKind VictronDeviceKind = iota
	VictronRandomBmvKind
	VictronRandomSolarKind
	VictronVedirectKind
	VictronVebusKind
)

func (dk VictronDeviceKind) String() string {
	switch dk {
	case VictronRandomBmvKind:
		return "RandomBmv"
	case VictronRandomSolarKind:
		return "RandomSolar"
	case VictronVedirectKind:
		return "Vedirect"
	case VictronVebusKind:
		return "Vebus"
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
	case "Vebus":
		return VictronVebusKind
	default:
		return VictronUndefinedKind
	}
}
