package device

import "github.com/koestler/go-iotdevice/dataflow"

var availabilityRegister = dataflow.CreateRegisterStruct(
	"Availability",
	"Connection",
	"Connection",
	dataflow.EnumRegister,
	map[int]string{
		0: "disconnected",
		1: "connected",
	},
	"",
	1000,
	false,
)

func GetAvailabilityRegister() dataflow.Register {
	return availabilityRegister
}
