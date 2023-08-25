package dataflow

var EmptyFilter = func(Value) bool { return true }

func DeviceFilter(deviceName string) FilterFunc {
	return func(value Value) bool {
		return value.DeviceName() == deviceName
	}
}

func RegisterFilter(
	skipFields []string,
	skipCategories []string,
) FilterFunc {
	skipFieldsMap := SliceToMap(skipFields)
	skipCategoriesMap := SliceToMap(skipCategories)

	return func(value Value) bool {
		reg := value.Register()

		if _, ok := skipFieldsMap[reg.Name()]; ok {
			return false
		}

		if _, ok := skipCategoriesMap[reg.Category()]; ok {
			return false
		}

		return true
	}
}

var NullFilter FilterFunc = func(value Value) bool {
	_, isNullValue := value.(NullRegisterValue)
	return isNullValue
}

func DeviceNonNullFilter(deviceName string) FilterFunc {
	deviceFilter := DeviceFilter(deviceName)
	return func(value Value) bool {
		return NullFilter(value) && deviceFilter(value)
	}
}

func SliceToMap[T comparable](inp []T) map[T]struct{} {
	oup := make(map[T]struct{}, len(inp))
	for _, v := range inp {
		oup[v] = struct{}{}
	}
	return oup
}
