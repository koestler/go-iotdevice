package types

type MqttDeviceKind int

const (
	MqttUndefinedKind MqttDeviceKind = iota
	MqttGoIotdeviceV3Kind
)

func (dk MqttDeviceKind) String() string {
	switch dk {
	case MqttGoIotdeviceV3Kind:
		return "GoIotdeviceV3"
	default:
		return "Undefined"
	}
}

func MqttDeviceKindFromString(s string) MqttDeviceKind {
	if s == "GoIotdeviceV3" {
		return MqttGoIotdeviceV3Kind
	}

	return MqttUndefinedKind
}
