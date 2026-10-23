package dataflow

type RegisterType int

const (
	UndefinedRegister RegisterType = iota
	NumberRegister
	TextRegister
	EnumRegister
)

func (rt RegisterType) String() string {
	switch rt {
	case NumberRegister:
		return "number"
	case TextRegister:
		return "string"
	case EnumRegister:
		return "enum"
	default:
		return ""
	}
}

func RegisterTypeFromString(s string) RegisterType {
	switch s {
	case "number":
		return NumberRegister
	case "string":
		return TextRegister
	case "enum":
		return EnumRegister
	default:
		return UndefinedRegister
	}
}
