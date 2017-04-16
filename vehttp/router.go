package vehttp

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func newRouter(routes Routes) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)

	}
	return router
}

func Run(bind string, port int, routes Routes) {
	router := newRouter(routes)

	log.Printf("vehttp: listening on port %v", port)
	log.Fatal(router, http.ListenAndServe(bind+":"+strconv.Itoa(port), router))
}
