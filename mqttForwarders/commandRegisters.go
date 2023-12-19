package mqttForwarders

import "github.com/koestler/go-iotdevice/v3/dataflow"

func getCommandFilter(cfg Config, devName string) dataflow.RegisterFilterFunc {
	// by default, nothing is commandable
	filter := func(dataflow.Register) bool {
		return false
	}

	// when command is enabled, use filter of given device
	var dev MqttDeviceSectionConfig
	if cfg.Command().Enabled() {
		dev = getCommandDevice(cfg, devName)
	}
	if dev != nil {
		commandEnabledFilter := dataflow.RegisterFilter(dev.Filter())
		filter = func(r dataflow.Register) bool {
			return r.Commandable() && commandEnabledFilter(r)
		}
	}

	return filter
}

func getCommandDevice(cfg Config, deviceName string) (r MqttDeviceSectionConfig) {
	for _, c := range cfg.Command().Devices() {
		if c.Name() == deviceName {
			return c
		}
	}
	return nil
}
