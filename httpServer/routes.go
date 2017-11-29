package httpServer

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
		"DeviceRoundedValues",
		"GET",
		"/api/v0/device/{DeviceId:[a-zA-Z0-9\\-]{1,32}}/RoundedValues",
		HandleDeviceGetRoundedValues,
	},
	HttpRoute{
		"DevicePicture",
		"GET",
		"/api/v0/device/{DeviceId:[a-zA-Z0-9\\-]{1,32}}/Picture",
		HandleDeviceGetPicture,
	},
	HttpRoute{
		"DeviceRoundedValuesWebSocket",
		"GET",
		"/api/v0/ws/RoundedValues",
		HandleWsRoundedValues,
	},
	HttpRoute{
		"ApiIndex",
		"GET",
		"/api{Path:.*}",
		HandleApiNotFound,
	},
	HttpRoute{
		"Assets",
		"GET",
		"/{Path:.+}",
		HandleAssetsGet,
	},
}
