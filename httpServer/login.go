package httpServer

import (
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/tg123/go-htpasswd"
	"golang.org/x/sync/semaphore"
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
// @Router /authentication/login [post]
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

	mux.HandleFunc("POST /api/v2/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonErrorResponse(w, http.StatusUnprocessableEntity, errors.New("Invalid json body provided"))
			return
		}

		reloadAuthChecker(authChecker, env.Config)
		if !authChecker.Match(req.User, req.Password) {
			jsonErrorResponse(w, http.StatusUnauthorized, errors.New("Invalid credentials"))
			return
		}

		tokenStr, err := createJwtToken(env.Authentication, req.User)
		if err != nil {
			jsonErrorResponse(w, http.StatusInternalServerError, errors.New("Cannot create token"))
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
		w.Write(jsonBytes)
	})
	if env.Config.LogConfig() {
		log.Printf("httpServer: POST /api/v2/auth/login -> serve login")
	}
}

func disableLogin(mux *http.ServeMux, config Config) {
	mux.HandleFunc("POST /api/v2/auth/login", func(w http.ResponseWriter, r *http.Request) {
		jsonErrorResponse(w, http.StatusServiceUnavailable, errors.New("Authentication module is disabled"))
	})
	if config.LogConfig() {
		log.Printf("httpServer: POST /api/v2/auth/login -> login disabled")
	}
}

var sem = semaphore.NewWeighted(1)
var lastAuthReload time.Time

func reloadAuthChecker(file *htpasswd.File, config Config) {
	// make sure this code run only once at a time
	if !sem.TryAcquire(1) {
		return
	}
	defer sem.Release(1)

	// make sure reload happens no more than once a second
	now := time.Now()
	if lastAuthReload.Add(time.Second).After(now) {
		return
	}
	lastAuthReload = now

	// reload the file
	err := file.Reload(func(err error) {
		log.Printf("httpServer: login: error while reading htaccess file line: %s", err)
	})
	if err != nil {
		log.Printf("httpServer: login: error while reading htaccess file: %s", err)
	}

	if config.LogDebug() {
		log.Printf("httpServer: login: authentication file reloaded")
	}
}
