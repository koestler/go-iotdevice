package httpServer

import (
	"compress/gzip"
	"log"
	"net/http"
	"strings"
)

// gzipMiddleware provides gzip compression for responses
func gzipMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gzw := &gzipResponseWriter{ResponseWriter: w}

		// serve the request. the gzip writer is lazily initialized
		next(gzw, r)

		if gzw.gz != nil {
			err := gzw.gz.Close()
			if err != nil {
				log.Printf("httpServer: error closing gzip writer: %s", err)
			}
		}
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	initialized bool
	gz          *gzip.Writer
}

func (gzw *gzipResponseWriter) Initialize() {
	if gzw.initialized {
		return
	}
	gzw.initialized = true

	header := gzw.Header()
	header.Set("Content-Encoding", "gzip")
	header.Del("Content-Length") // content length will be different after gzip
}

func (gzw *gzipResponseWriter) WriteHeader(code int) {
	gzw.Initialize()
	gzw.ResponseWriter.WriteHeader(code)
}

func (gzw *gzipResponseWriter) Write(b []byte) (int, error) {
	gzw.Initialize()

	if gzw.gz == nil {
		gzw.gz = gzip.NewWriter(gzw.ResponseWriter)
	}
	return gzw.gz.Write(b)
}
