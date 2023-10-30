package types

type MqttDeviceKind int

const (
	MqttDeviceUndefinedKind MqttDeviceKind = iota
	MqttDeviceGoIotdeviceKind
)

func (dk MqttDeviceKind) String() string {
	switch dk {
	case MqttDeviceGoIotdeviceKind:
		return "GoIotdevice"
	default:
		return "Undefined"
	}
}

func MqttDeviceKindFromString(s string) MqttDeviceKind {
	if s == "GoIotdevice" {
		return MqttDeviceGoIotdeviceKind
	}

	return MqttDeviceUndefinedKind
}
