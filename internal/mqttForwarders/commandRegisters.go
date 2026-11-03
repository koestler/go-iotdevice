package mqttForwarders

import (
	"slices"

	"github.com/koestler/go-iotdevice/v3/internal/dataflow"
)

func getCommandFilter(cfg Config, devName string) dataflow.RegisterFilterFunc {
	// when command is enabled, use filter of given device
	if cfg.Command().Enabled() {
		if dev := findSectionDevice(cfg.Command(), devName); dev != nil {
			return createWritableAndRegisterValueFilter(dev.Filter())
		}
	}

	// by default, nothing is writable
	return func(dataflow.Filterable) bool {
		return false
	}
}

// findSectionDevice returns the device config of the given name or nil when the section does not contain it.
func findSectionDevice(section interface {
	Devices() []MqttDeviceSectionConfig
}, deviceName string) MqttDeviceSectionConfig {
	i := slices.IndexFunc(section.Devices(), func(d MqttDeviceSectionConfig) bool {
		return d.Name() == deviceName
	})
	if i < 0 {
		return nil
	}
	return section.Devices()[i]
}
