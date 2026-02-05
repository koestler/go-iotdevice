package httpServer

import (
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
)

func setupFrontend(mux *http.ServeMux, config Config, views []ViewConfig) {
	frontendUrl := config.FrontendProxy()

	if frontendUrl != nil {
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
					if len(route) == 0 || route[0] != '/' {
						route = "/" + route
					}
					serveStatic(mux, config, route, filePath)
					return nil
				})

				// load index file for single page frontend application
				indexPath := frontendPath + "/index.html"
				spaRoutes := append(getNames(views), "", "login")
				for _, route := range spaRoutes {
					if route == "" {
						route = "/"
					} else {
						route = "/" + route
					}
					serveStatic(mux, config, route, indexPath)
				}

				if err != nil {
					log.Printf("httpServer: failed to serve from local folder: %s", err)
				}
			}
		} else {
			log.Print("httpServer: no frontend configured")
		}
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			setCacheControlPublic(w, config.FrontendExpires())
			jsonErrorResponse(w, http.StatusNotFound, errors.New("route not found"))
		})
	}
}

func serveStatic(mux *http.ServeMux, config Config, route, filePath string) {
	// In Go 1.22+, paths ending without "/" are exact matches by default
	// Only "/" matches as prefix; to make "/" exact, we need "/{$}"
	pattern := "GET " + route
	if route == "/" {
		pattern = "GET /{$}"
	}
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		setCacheControlPublic(w, config.FrontendExpires())
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
