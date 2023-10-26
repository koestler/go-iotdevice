package dataflow

var EmptyFilter = func(Value) bool { return true }

func DeviceNameValueFilter(deviceName string) ValueFilterFunc {
	return func(value Value) bool {
		return value.DeviceName() == deviceName
	}
}

func RegisterValueFilter(registerFilter RegisterFilterConf) ValueFilterFunc {
	f := RegisterFilter(registerFilter)
	return func(value Value) bool {
		return f(value.Register())
	}
}

var AllValueFilter ValueFilterFunc = func(value Value) bool {
	return true
}

var NonNullValueFilter ValueFilterFunc = func(value Value) bool {
	_, isNullValue := value.(NullRegisterValue)
	return !isNullValue
}

func DeviceNonNullValueFilter(deviceName string) ValueFilterFunc {
	deviceFilter := DeviceNameValueFilter(deviceName)
	return func(value Value) bool {
		return NonNullValueFilter(value) && deviceFilter(value)
	}
}
