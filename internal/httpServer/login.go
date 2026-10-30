package httpServer

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"errors"
	"github.com/tg123/go-htpasswd"
)

type loginRequest struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type loginResponse struct {
	User         string   `json:"user"`
	Token        string   `json:"token"`
	AllowedViews []string `json:"allowedViews"`
}

// setupLogin godoc
// @Summary Login
// @Description Creates a new JWT token used for authentication.
// @Accept json
// @Produce json
// @Param request body loginRequest true "user info"
// @Success 200 {object} loginResponse
// @Failure 422 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /auth/login [post]
func setupLogin(mux *http.ServeMux, env *Environment) {
	if !env.Authentication.Enabled() {
		disableLogin(mux, env.Config)
		return
	}

	// setup htpasswd module
	authChecker, err := htpasswd.New(env.Authentication.HtaccessFile(), htpasswd.DefaultSystems, nil)
	if err != nil {
		log.Printf("httpServer: cannot load htaccess file: %s", err)
		disableLogin(mux, env.Config)
		return
	}

	mux.HandleFunc("POST /api/v2/auth/login", gzipMiddleware(loginHandler(env, authChecker)))
	if env.Config.LogConfig() {
		log.Printf("httpServer: POST /api/v2/auth/login -> serve login")
	}
}

func loginHandler(env *Environment, authChecker *htpasswd.File) http.HandlerFunc {
	reloader := &authReloader{file: authChecker, config: env.Config}
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := decodeJsonBody(w, r, &req); err != nil {
			respondInvalidBody(w, err)
			return
		}

		reloader.reload()
		if !authChecker.Match(req.User, req.Password) {
			jsonErrorResponse(w, http.StatusUnauthorized, errors.New("invalid credentials"))
			return
		}

		tokenStr, err := createJwtToken(env.Authentication, req.User)
		if err != nil {
			jsonErrorResponse(w, http.StatusInternalServerError, errors.New("cannot create token"))
			return
		}

		// get allowed views
		allowedViews := make([]string, 0)
		for _, v := range env.Views {
			if v.IsAllowed(req.User) {
				allowedViews = append(allowedViews, v.Name())
			}
		}

		response := loginResponse{Token: tokenStr, User: req.User, AllowedViews: allowedViews}
		jsonBytes, _ := json.Marshal(response)
		setContentTypeJsonHeader(w)
		w.WriteHeader(http.StatusOK)
		writeLogError(w, jsonBytes)
	}
}

func disableLogin(mux *http.ServeMux, config Config) {
	mux.HandleFunc("POST /api/v2/auth/login", func(w http.ResponseWriter, r *http.Request) {
		jsonErrorResponse(w, http.StatusServiceUnavailable, errors.New("authentication module is disabled"))
	})
	if config.LogConfig() {
		log.Printf("httpServer: POST /api/v2/auth/login -> login disabled")
	}
}

type authReloader struct {
	file   *htpasswd.File
	config Config

	mu         sync.Mutex
	lastReload time.Time
}

func (r *authReloader) reload() {
	// make sure this code runs only one at a time
	if !r.mu.TryLock() {
		return
	}
	defer r.mu.Unlock()

	// make sure reload happens no more than once a second
	now := time.Now()
	if r.lastReload.Add(time.Second).After(now) {
		return
	}
	r.lastReload = now

	// reload the file
	err := r.file.Reload(func(err error) {
		log.Printf("httpServer: login: error while reading htaccess file line: %s", err)
	})
	if err != nil {
		log.Printf("httpServer: login: error while reading htaccess file: %s", err)
		return
	}

	if r.config.LogDebug() {
		log.Printf("httpServer: login: authentication file reloaded")
	}
}
