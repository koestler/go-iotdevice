package webserver

var wsRoutes = WsRoutes{
	WsRoute{
		"ws-test",
		"/ws/v0/RoundedValues",
		HandleWsRoundedValues,
	},
}

var httpRoutes = HttpRoutes{
	HttpRoute{
		"DeviceIndex",
		"GET",
		"/api/v0/device/",
		HandleDeviceIndex,
	},
	HttpRoute{
		"DeviceIndex",
		"GET",
		"/api/v0/device/{DeviceId:[a-zA-Z0-9\\-]{1,32}}/RoundedValues",
		HandleDeviceGetRoundedValues,
	},
	HttpRoute{
		"DeviceIndex",
		"GET",
		"/api/v0/ws/RoundedValues",
		HandleDeviceGetRoundedValues,
	},
	HttpRoute{
		"DeviceIndex",
		"GET",
		"/api/v0/ws/RoundedValues",
		HandleDeviceGetRoundedValues,
	},
	HttpRoute{
		"Index",
		"GET",
		"/api",
		HandleApiNotFound,
	},
	HttpRoute{
		"Assets",
		"GET",
		"/{Path:.+}",
		HandleAssetsGet,
	},
}
