package dataflow

type RegisterType int

const (
	TextRegister RegisterType = iota
	NumberRegister
	EnumRegister
)

func (rt RegisterType) String() string {
	switch rt {
	case TextRegister:
		return "string"
	case NumberRegister:
		return "number"
	case EnumRegister:
		return "enum"
	default:
		return ""
	}
}
