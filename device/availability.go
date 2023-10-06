package device

import "github.com/koestler/go-iotdevice/dataflow"

const availabilityRegisterName = "Available"
const availabilityOfflineValue = "offline"
const availabilityOnlineValue = "online"

var availabilityRegister = dataflow.NewRegisterStruct(
	availabilityRegisterName,
	availabilityRegisterName,
	availabilityRegisterName,
	dataflow.EnumRegister,
	map[int]string{
		0: availabilityOfflineValue,
		1: availabilityOnlineValue,
	},
	"",
	1000,
	false,
)
