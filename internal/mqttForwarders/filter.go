package mqttForwarders

import (
	"github.com/koestler/go-iotdevice/v3/internal/dataflow"
	"github.com/koestler/go-iotdevice/v3/internal/device"
)

func availabilityRegisterFilter(r dataflow.Filterable) bool {
	// do not use Availability as a register in mqtt; availability is handled separately
	return r.Name() != device.AvailabilityRegisterName
}

func createRegisterValueFilter(registerFilter dataflow.RegisterFilterConf) dataflow.RegisterFilterFunc {
	f0 := availabilityRegisterFilter
	f1 := dataflow.RegisterFilter(registerFilter)
	return func(r dataflow.Filterable) bool {
		return f0(r) && f1(r)
	}
}

func createDeviceAndRegisterValueFilter(dev device.Device, registerFilter dataflow.RegisterFilterConf) dataflow.ValueFilterFunc {
	deviceName := dev.Name()
	f := createRegisterValueFilter(registerFilter)
	return func(curr, prev dataflow.Value) bool {
		return !prev.Equals(curr) && curr.DeviceName() == deviceName && f(curr.Register())
	}
}

func createWritableAndRegisterValueFilter(registerFilter dataflow.RegisterFilterConf) dataflow.RegisterFilterFunc {
	f0 := createRegisterValueFilter(registerFilter)
	return func(r dataflow.Filterable) bool {
		return r.Writable() && f0(r)
	}
}
