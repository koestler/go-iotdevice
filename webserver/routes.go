package webserver

var HttpRoutes = Routes{
	Route{
		"DeviceIndex",
		"GET",
		"/api/v0/device/",
		HandleDeviceIndex,
	},
	Route{
		"DeviceIndex",
		"GET",
		"/api/v0/device/{DeviceId:[a-zA-Z0-9\\-]{1,32}}/RoundedValues",
		HandleDeviceGetRoundedValues,
	},
	Route{
		"Index",
		"GET",
		"/",
		HandleAssetsGet,
	},
	Route{
		"Assets",
		"GET",
		"/{Path:.+}",
		HandleAssetsGet,
	},
}
