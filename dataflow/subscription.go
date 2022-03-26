package dataflow

type Subscription struct {
	shutdownChannel chan struct{}
	outputChannel   chan Value
	filter          Filter
}

func (s *Subscription) Shutdown() {
	close(s.outputChannel)
	close(s.shutdownChannel)
}

func (s *Subscription) GetOutput() <-chan Value {
	return s.outputChannel
}

func filterByDevice(filter *Filter, deviceName string) bool {
	// list is empty -> every device is ok
	if len(filter.Devices) < 1 {
		return true
	}

	// only ok if present and true
	_, ok := filter.Devices[deviceName]
	return ok && filter.Devices[deviceName]
}

func filterByRegisterName(filter *Filter, registerName string) bool {
	// list is empty -> every device is ok
	if len(filter.RegisterNames) < 1 {
		return true
	}

	// only ok if present and true
	_, ok := filter.RegisterNames[registerName]
	return ok && filter.RegisterNames[registerName]
}

func filterValue(filter *Filter, value Value) bool {
	return filterByDevice(filter, value.DeviceName()) && filterByRegisterName(filter, value.Register().Name())
}

func (s *Subscription) forward(newValue Value) {
	if filterValue(&s.filter, newValue) {
		// forward value
		s.outputChannel <- newValue
	}
}
