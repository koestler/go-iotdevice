package dataflow

type Subscription struct {
	shutdownChannel chan struct{}
	outputChannel   chan Value
	filter          Filter
}

func (s *Subscription) Shutdown() {
	close(s.shutdownChannel)
}

func (s *Subscription) GetOutput() <-chan Value {
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

func filterByRegisterName(filter *Filter, register Register) bool {
	// list is empty -> every device is ok
	if len(filter.SkipRegisterNames) < 1 {
		return true
	}

	// keep all but those that are present and false
	v, ok := filter.SkipRegisterNames[register.Name()]
	return !(ok && !v)
}

func filterByRegisterCategory(filter *Filter, register Register) bool {
	// list is empty -> every device is ok
	if len(filter.SkipRegisterCategories) < 1 {
		return true
	}

	// keep all but those that are present and false
	v, ok := filter.SkipRegisterCategories[register.Category()]
	return !(ok && !v)
}

func filterByRegister(filter *Filter, register Register) bool {
	return filterByRegisterName(filter, register) &&
		filterByRegisterCategory(filter, register)
}

func filterValue(filter *Filter, value Value) bool {
	return filterByDevice(filter, value.DeviceName()) &&
		filterByRegister(filter, value.Register())
}

func (s *Subscription) forward(newValue Value) {
	if filterValue(&s.filter, newValue) {
		// forward value
		s.outputChannel <- newValue
	}
}
