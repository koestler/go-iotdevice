package httpServer

import (
	"github.com/gorilla/mux"
	"github.com/lestrrat-go/apache-logformat"
	"io"
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

func newRouter(logger io.Writer, env *Environment) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// setup websocket routes
	for _, route := range wsRoutes {
		var handler http.Handler
		handler = Handler{Env: env, Handle: route.HandlerFunc}

		if logger != nil {

		}
		if logger != nil {
			handler = apachelog.CombinedLog.Wrap(handler, logger)
		}

		router.Path(route.Pattern).
			Name(route.Name).
			Handler(handler)

	}

	// setup normal http routes
	for _, route := range httpRoutes {
		var handler http.Handler
		handler = Handler{Env: env, Handle: route.HandlerFunc}
		if logger != nil {
			handler = apachelog.CombinedLog.Wrap(handler, logger)
		}

		router.Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)

	}

	return router
}
