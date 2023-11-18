package modbusDevice

import "github.com/koestler/go-iotdevice/dataflow"

type FinderRegister struct {
	dataflow.RegisterStruct
	addressBegin uint16
	addressEnd   uint16
}

type FinderRegisterType int

const (
	FinderT1 FinderRegisterType = iota
	FinderT2
	FinderT3
	FinderT4
	FinderT5
	FinderT6
	FinderT7
	FinderT8
	FinderT9
	FinderT10
	FinderT_Str2
	FinderT_Str4
	FinderT_Str6
	FinderT_Str8
	FinderT_Str16
	FinderT_Str20
	FinderT16
	FinderT17
	FinderT_Time
	FinderT_TimeIEC
	FinderT_Data
	FinderT_Str40
	FinderT_float
	FinderT9A
	FinderT10A
	Finder18
	FinderT_unix
)

func FinderRegisterTypeToType(t FinderRegisterType) dataflow.RegisterType {
	switch t {
	case FinderT1, FinderT2, FinderT3, FinderT4, FinderT5, FinderT6, FinderT7, FinderT8, FinderT9, FinderT10:
		return dataflow.NumberRegister
	case FinderT_Str2, FinderT_Str4, FinderT_Str6, FinderT_Str8, FinderT_Str16, FinderT_Str20:
		return dataflow.TextRegister
	case FinderT16, FinderT17:
		return dataflow.NumberRegister
	case FinderT_Time, FinderT_TimeIEC:
		return dataflow.TextRegister
	case FinderT_Data, FinderT_Str40:
		return dataflow.TextRegister
	case FinderT_float:
		return dataflow.NumberRegister
	case FinderT9A:
		return dataflow.TextRegister
	case FinderT10A, Finder18, FinderT_unix:
		return dataflow.NumberRegister
	default:
		return -1
	}
}

func NewFinderRegister(
	category, name, description string,
	registerType FinderRegisterType,
	addressBegin, addressEnd uint16,
	enum map[int]string,
	unit string,
	sort int,
) FinderRegister {
	return FinderRegister{
		dataflow.NewRegisterStruct(
			category, name, description,
			FinderRegisterTypeToType(registerType),
			enum,
			unit,
			sort,
			false,
		),
		addressBegin,
		addressEnd,
	}
}
