package device

import "github.com/koestler/go-iotdevice/dataflow"

const availabilityRegisterName = "Available"

var availabilityRegister = dataflow.NewRegisterStruct(
	availabilityRegisterName,
	availabilityRegisterName,
	availabilityRegisterName,
	dataflow.EnumRegister,
	map[int]string{
		0: "offline",
		1: "online",
	},
	"",
	1000,
	false,
)

func availabilityValueFilter(value dataflow.Value) bool {
	reg := value.Register()
	return reg.RegisterType() == dataflow.EnumRegister && reg.Name() == availabilityRegisterName
}
