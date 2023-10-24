package device

import "github.com/koestler/go-iotdevice/dataflow"

const AvailabilityRegisterName = "Available"
const AvailabilityOfflineValue = "offline"
const AvailabilityOnlineValue = "online"

var availabilityRegister = dataflow.NewRegisterStruct(
	AvailabilityRegisterName,
	AvailabilityRegisterName,
	AvailabilityRegisterName,
	dataflow.EnumRegister,
	map[int]string{
		0: AvailabilityOfflineValue,
		1: AvailabilityOnlineValue,
	},
	"",
	1000,
	false,
)

func GetAvailabilityRegister() dataflow.Register {
	return availabilityRegister
}
