package mqttForwarders

import (
	"github.com/koestler/go-iotdevice/config"
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/device"
)

var availabilityRegisterFilter = func(r dataflow.Register) bool {
	// do not use Availability as a register in mqtt; availability is handled separately
	return r.Name() != device.AvailabilityRegisterName
}

func createRegisterValueFilter(registerFilter config.RegisterFilterConfig) dataflow.RegisterFilterFunc {
	f0 := availabilityRegisterFilter
	f1 := dataflow.RegisterFilter(registerFilter)
	return func(r dataflow.Register) bool {
		return f0(r) && f1(r)
	}
}

func createDeviceAndRegisterValueFilter(dev device.Device, registerFilter config.RegisterFilterConfig) dataflow.ValueFilterFunc {
	f0 := dataflow.DeviceNameValueFilter(dev.Name())
	f1 := createRegisterValueFilter(registerFilter)
	return func(v dataflow.Value) bool {
		return f0(v) && f1(v.Register())
	}
}
