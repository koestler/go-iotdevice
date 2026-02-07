package httpServer

import (
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/koestler/go-iotdevice/v3/dataflow"
)

type HttpServer struct {
	config Config
	server *http.Server
}

type RegisterDbOfDeviceFunc func(deviceName string) *dataflow.RegisterDb

type Environment struct {
	Config             Config
	ProjectTitle       string
	Views              []ViewConfig
	Authentication     AuthenticationConfig
	RegisterDbOfDevice RegisterDbOfDeviceFunc
	StateStorage       *dataflow.ValueStorage
	CommandStorage     *dataflow.ValueStorage
}

type Config interface {
	BuildVersion() string
	Bind() string
	Port() int
	LogRequests() bool
	LogDebug() bool
	LogConfig() bool
	FrontendProxy() *url.URL
	FrontendPath() string
	FrontendExpires() time.Duration
	ConfigExpires() time.Duration
}

type ViewConfig interface {
	Name() string
	Title() string
	Devices() []ViewDeviceConfig
	Autoplay() bool
	IsAllowed(user string) bool
	IsPublic() bool
	Hidden() bool
}

type ViewDeviceConfig interface {
	Name() string
	Title() string
	Filter() dataflow.RegisterFilterConf
}

type AuthenticationConfig interface {
	Enabled() bool
	JwtSecret() []byte
	JwtValidityPeriod() time.Duration
	HtaccessFile() string
}

func Run(env *Environment) (httpServer *HttpServer) {
	cfg := env.Config

	apiMux := http.NewServeMux()
	addApiV2Routes(apiMux, env)
	setupFrontend(apiMux, env.Config, env.Views)

	rootMux := http.NewServeMux()
	rootMux.HandleFunc("/", middlewares(apiMux.ServeHTTP, env))
	setupValuesWs(rootMux, env)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Bind(), cfg.Port()),
		Handler: rootMux,
	}

	go func() {
		if cfg.LogDebug() {
			log.Printf("httpServer: listening on %v", server.Addr)
		}
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Printf("httpServer: stopped due to error: %s", err)
		}
	}()

	return &HttpServer{
		config: cfg,
		server: server,
	}
}

func (s *HttpServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err != nil {
		log.Printf("httpServer: graceful shutdown failed: %s", err)
	}
}

// middlewares creates the chain: logging -> auth -> gzip -> mux
func middlewares(handler http.HandlerFunc, env *Environment) http.HandlerFunc {
	handler = gzipMiddleware(handler)
	handler = authJwtMiddleware(handler, env)
	if env.Config.LogRequests() {
		handler = loggingMiddleware(handler)
	}
	return handler
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(wrapped, r)
		log.Printf(
			"httpServer: %s %s %d %v %dB",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			time.Since(start),
			humanize.IBytes(uint64(wrapped.written)),
		)
	}
}

// responseWriter is a wrapper around http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("httpServer: response writer does not support hijacking")
	}
	return hj.Hijack()
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n
	return n, err
}

// gzipMiddleware provides gzip compression for responses
func gzipMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gzw := &gzipResponseWriter{ResponseWriter: w}

		// serve the request. the gzip writer is lazily initialized on the first write
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
	gz *gzip.Writer
}

func (gzw *gzipResponseWriter) Write(b []byte) (int, error) {
	if gzw.gz == nil {
		gzw.gz = gzip.NewWriter(gzw.ResponseWriter)
	}
	return gzw.gz.Write(b)
}
