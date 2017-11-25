package httpServer

import (
	"github.com/gorilla/mux"
	"net/http"
)

type WsRoute struct {
	Name        string
	Pattern     string
	HandlerFunc HandlerHandleFunc
}

type HttpRoute struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc HandlerHandleFunc
}

type WsRoutes []WsRoute
type HttpRoutes []HttpRoute

func newRouter(env *Environment) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// setup websocket routes
	for _, route := range wsRoutes {
		var handler http.Handler
		handler = Handler{Env: env, Handle: route.HandlerFunc}
		handler = Logger(handler, route.Name)

		router.Path(route.Pattern).
			Name(route.Name).
			Handler(handler)

	}

	// setup normal http routes
	for _, route := range httpRoutes {
		var handler http.Handler
		handler = Handler{Env: env, Handle: route.HandlerFunc}
		handler = Logger(handler, route.Name)

		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)

	}

	return router
}
