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

func SendDisconnected(deviceName string, output chan dataflow.Value) {
	output <- dataflow.NewEnumRegisterValue(deviceName, GetAvailabilityRegister(), 0, false)
}

func SendConnteced(deviceName string, output chan dataflow.Value) {
	output <- dataflow.NewEnumRegisterValue(deviceName, GetAvailabilityRegister(), 1, false)
}
