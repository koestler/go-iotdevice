package dataflow

import "context"

type Subscription struct {
	ctx           context.Context
	outputChannel chan Value
	filter        Filter
}

func (s *Subscription) Drain() <-chan Value {
	return s.outputChannel
}

func filterByDevice(filter *Filter, deviceName string) bool {
	// list is empty -> every device is ok
	if len(filter.IncludeDevices) < 1 {
		return true
	}

	// keep if entry is present and true
	v, ok := filter.IncludeDevices[deviceName]
	return ok && v
}

func filterByRegisterName(filter *Filter, deviceName string, register Register) bool {
	// list is empty -> every device is ok
	if len(filter.SkipRegisterNames) < 1 {
		return true
	}

	// keep all but those that are present and false
	v, ok := filter.SkipRegisterNames[SkipRegisterNameStruct{
		Device:   deviceName,
		Register: register.Name(),
	}]
	return !(ok && v)
}

func filterByRegisterCategory(filter *Filter, deviceName string, register Register) bool {
	// list is empty -> every device is ok
	if len(filter.SkipRegisterCategories) < 1 {
		return true
	}

	// keep all but those that are present and false
	v, ok := filter.SkipRegisterCategories[SkipRegisterCategoryStruct{
		Device:   deviceName,
		Category: register.Category(),
	}]
	return !(ok && v)
}

func filterByRegister(filter *Filter, deviceName string, register Register) bool {
	return filterByRegisterName(filter, deviceName, register) &&
		filterByRegisterCategory(filter, deviceName, register)
}

func filterByValue(filter *Filter, value Value) bool {
	_, isNullValue := value.(NullRegisterValue)
	return !(filter.SkipNull && isNullValue)
}

func filterValue(filter *Filter, value Value) bool {
	register := value.Register()
	return filterByDevice(filter, value.DeviceName()) &&
		filterByRegister(filter, value.DeviceName(), register) &&
		filterByValue(filter, value)
}

func (s *Subscription) forward(newValue Value) {
	if filterValue(&s.filter, newValue) {
		// forward value
		s.outputChannel <- newValue
	}
}
