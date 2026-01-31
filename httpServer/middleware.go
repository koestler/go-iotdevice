package httpServer

import (
	"log"
	"net/http"
	"time"
)

// loggingMiddleware logs HTTP requests if enabled
func loggingMiddleware(config Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if !config.LogRequests() {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			log.Printf("[HTTP] %s %s %d %s", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
		})
	}
}

// recoveryMiddleware recovers from panics and logs them
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
