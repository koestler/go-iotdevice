package main

/*
func setupBmvDevices() {
	log.Printf("main: setup Bmv Devices")

	configs := config.GetVedeviceConfigs()

	sources := make([]dataflow.Drainable, 0, len(configs))

	// get devices from database and create them
	for _, c := range configs {
		log.Printf(
			"bmvDevices: setup name=%v model=%v device=%v",
			c.Name, c.Model, c.Device,
		)

		// register device in storage
		device := storage.DeviceCreate(c.Name, c.Model, c.FrontendConfig)

		// setup the datasource
		if "dummy" == c.Device {
			sources = append(sources, vedevices.CreateDummySource(device, c))
		} else {
			if err, source := vedevices.CreateSource(device, c); err == nil {
				sources = append(sources, source)
			} else {
				log.Printf("bmvDevices: error during CreateSource: %v", err)
			}
		}
	}

	// append them as sources to the raw storage
	for _, source := range sources {
		source.Append(rawStorage)
	}
}
*/
