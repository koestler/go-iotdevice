package httpServer

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
	"path/filepath"
)

func setupFrontend(engine *gin.Engine, config Config) {
	frontendUrl := config.FrontendProxy()

	if frontendUrl != nil {
		engine.NoRoute(func(c *gin.Context) {
			proxy := httputil.NewSingleHostReverseProxy(frontendUrl)
			proxy.ServeHTTP(c.Writer, c.Request)
		})
		if config.LogConfig() {
			log.Printf("httpServer: /* -> proxy %s*", frontendUrl)
		}
	} else {
		frontendPath := path.Clean(config.FrontendPath())

		if len(frontendPath) > 0 {
			if frontendPathInfo, err := os.Lstat(frontendPath); err != nil {
				log.Printf("httpServer: given frontend path is not accessible: %s", err)
			} else if !frontendPathInfo.IsDir() {
				log.Printf("httpServer: given frontend path is not a directory")
			} else {
				err := filepath.Walk(frontendPath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						return nil
					}

					route := path[len(frontendPath):]
					serveStatic(engine, config, route, path)
					return nil
				})

				// load index file single page frontend application
				path := frontendPath + "/index.html"
				for _, route := range append(config.GetViewNames(), "", "login") {
					route = "/" + route
					serveStatic(engine, config, route, path)
				}

				if err != nil {
					log.Printf("httpServer: failed to serve from local folder: %s", err)
				}
			}
		} else {
			log.Print("httpServer: no frontend configured")
		}
		engine.NoRoute(func(c *gin.Context) {
			setCacheControlPublic(c, config.FrontendExpires())
			jsonErrorResponse(c, http.StatusNotFound, errors.New("route not found"))
		})
	}
}

func serveStatic(engine *gin.Engine, config Config, route, filePath string) {
	engine.GET(route, func(c *gin.Context) {
		setCacheControlPublic(c, config.FrontendExpires())
		// c.File calls http.serveContent which sets / checks Last-Modified / If-Modified-Since
		c.File(filePath)
	})
	if config.LogConfig() {
		log.Printf("httpServer: %s -> serve static %s", route, filePath)
	}
}
