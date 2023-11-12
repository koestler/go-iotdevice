package types

type MqttDeviceKind int

const (
	MqttDeviceUndefinedKind MqttDeviceKind = iota
	MqttDeviceGoIotdeviceV3Kind
)

func (dk MqttDeviceKind) String() string {
	switch dk {
	case MqttDeviceGoIotdeviceV3Kind:
		return "GoIotdeviceV3"
	default:
		return "Undefined"
	}
}

func MqttDeviceKindFromString(s string) MqttDeviceKind {
	if s == "GoIotdeviceV3" {
		return MqttDeviceGoIotdeviceV3Kind
	}

	return MqttDeviceUndefinedKind
}
