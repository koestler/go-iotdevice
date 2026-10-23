package httpServer

import (
	"errors"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
)

func setupFrontend(mux *http.ServeMux, config Config, views []ViewConfig) {
	frontendUrl := config.FrontendProxy()

	if frontendUrl != nil {
		setupFrontendReverseProxy(mux, config, frontendUrl)
		return
	}

	frontendPath := config.FrontendPath()
	if frontendPath != "" {
		frontendPath = filepath.Clean(frontendPath)
		err := setupFrontendStatic(mux, os.DirFS(frontendPath), config, views)
		if err != nil {
			log.Printf("httpServer: failed to serve from local folder: %s", err)
		}
	} else {
		log.Print("httpServer: no frontend configured")
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		setCacheControlPublic(w, config.FrontendExpires())
		jsonErrorResponse(w, http.StatusNotFound, errors.New("route not found"))
	})
}

func setupFrontendReverseProxy(mux *http.ServeMux, config Config, frontendUrl *url.URL) {
	mux.Handle("/", httputil.NewSingleHostReverseProxy(frontendUrl))
	if config.LogConfig() {
		log.Printf("httpServer: GET /* -> proxy %s*", frontendUrl)
	}
}

func setupFrontendStatic(mux *http.ServeMux, srcFSys fs.FS, config Config, views []ViewConfig) error {
	err := fs.WalkDir(srcFSys, ".", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		serveStatic(mux, config, srcFSys, filePath, "/"+filePath)
		return nil
	})
	if err != nil {
		return err
	}

	// load index file for single page frontend application
	indexRoutes := append(getViewNames(views), "", "login")
	for _, route := range indexRoutes {
		serveStatic(mux, config, srcFSys, "index.html", "/"+route)
	}

	return nil
}

func serveStatic(mux *http.ServeMux, config Config, fSys fs.FS, fileName, route string) {
	if route == "/" {
		// root node is handled as prefix
		route = "/{$}"
	}
	pattern := "GET " + route
	if config.LogConfig() {
		log.Printf("httpServer: %s -> serve static %s", pattern, fileName)
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		setCacheControlPublic(w, config.FrontendExpires())
		http.ServeFileFS(w, r, fSys, fileName)
	}
	mux.HandleFunc(pattern, gzipMiddleware(handler))
}

func getViewNames(views []ViewConfig) []string {
	ret := make([]string, len(views))
	for i, view := range views {
		ret[i] = view.Name()
	}
	return ret
}
