package httpServer

import (
	"github.com/pkg/errors"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
	"path/filepath"
)

func setupFrontend(mux *http.ServeMux, env *Environment) {
	config := env.Config
	frontendUrl := config.FrontendProxy()

	if frontendUrl != nil {
		// Setup proxy for all unhandled routes
		proxy := httputil.NewSingleHostReverseProxy(frontendUrl)
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})
		if config.LogConfig() {
			log.Printf("httpServer: GET /* -> proxy %s*", frontendUrl)
		}
	} else {
		frontendPath := path.Clean(config.FrontendPath())

		if len(frontendPath) > 0 {
			if frontendPathInfo, err := os.Lstat(frontendPath); err != nil {
				log.Printf("httpServer: given frontend path is not accessible: %s", err)
			} else if !frontendPathInfo.IsDir() {
				log.Printf("httpServer: given frontend path is not a directory")
			} else {
				err := filepath.Walk(frontendPath, func(filePath string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						return nil
					}

					route := filePath[len(frontendPath):] // remove frontendPath prefix
					if len(route) > 0 && route[0] != '/' {
						route = "/" + route
					}
					serveStatic(mux, config, route, filePath)
					return nil
				})

				// load index file single page frontend application
				for _, route := range append(getNames(env.Views), "", "login") {
					if route != "" {
						route = "/" + route
					}
					serveStatic(mux, config, route, frontendPath+"/index.html")
				}

				if err != nil {
					log.Printf("httpServer: failed to serve from local folder: %s", err)
				}
			}
		} else {
			log.Print("httpServer: no frontend configured")
		}

		// Fallback 404 handler
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			setCacheControlPublic(w, config.FrontendExpires())
			jsonErrorResponse(w, http.StatusNotFound, errors.New("route not found"))
		})
	}
}

func serveStatic(mux *http.ServeMux, config Config, route string, filePath string) {
	mux.HandleFunc("GET "+route, func(w http.ResponseWriter, r *http.Request) {
		setCacheControlPublic(w, config.FrontendExpires())
		// http.ServeFile calls http.serveContent which sets / checks Last-Modified / If-Modified-Since
		http.ServeFile(w, r, filePath)
	})
	if config.LogConfig() {
		log.Printf("httpServer: GET %s -> serve static %s", route, filePath)
	}
}

type Nameable interface {
	Name() string
}

func getNames[N Nameable](list []N) (ret []string) {
	ret = make([]string, len(list))
	for i, t := range list {
		ret[i] = t.Name()
	}
	return
}
