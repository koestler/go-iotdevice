package dataflow

type RegisterType int

const (
	NumberRegister RegisterType = iota
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
