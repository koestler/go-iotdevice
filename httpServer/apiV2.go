package httpServer

import "net/http"

func addApiV2Routes(mux *http.ServeMux, env *Environment) {
	setupConfig(mux, env)
	setupLogin(mux, env)
	setupRegisters(mux, env)
	setupValuesGetJson(mux, env)
	setupValuesPatch(mux, env)
	setupDocs(mux, env)
	setupValuesWs(mux, env)
}
