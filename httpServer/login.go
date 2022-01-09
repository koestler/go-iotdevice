package httpServer

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tg123/go-htpasswd"
	"golang.org/x/sync/semaphore"
	"log"
	"net/http"
	"time"
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
// @Summary Login endpoint
// @Description Creates a new JWT token used for authentication if a valid user / password is given.
// @ID login
// @Accept json
// @Produce json
// @Param request body loginRequest true "user info"
// @Success 200 {object} loginResponse
// @Failure 422 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Failure 503 {object} ErrorResponse
// @Router /login [post]
func setupLogin(r *gin.RouterGroup, env *Environment) {
	if !env.Auth.Enabled() {
		disableLogin(r, env.Config)
		return
	}

	// setup htpasswd module
	authChecker, err := htpasswd.New(env.Auth.HtaccessFile(), htpasswd.DefaultSystems, nil)
	if err != nil {
		log.Printf("httpServer: cannot load htaccess file: %s", err)
		disableLogin(r, env.Config)
		return
	}

	r.POST("login", func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			jsonErrorResponse(c, http.StatusUnprocessableEntity, errors.New("Invalid json body provided"))
			return
		}

		reloadAuthChecker(authChecker, env.Config)
		if !authChecker.Match(req.User, req.Password) {
			jsonErrorResponse(c, http.StatusUnauthorized, errors.New("Invalid credentials"))
			return
		}

		tokenStr, err := createJwtToken(env.Auth, req.User)
		if err != nil {
			jsonErrorResponse(c, http.StatusInternalServerError, errors.New("Cannot create token"))
			return
		}

		// get allowed views
		allowedViews := make([]string, 0)
		for _, v := range env.Views {
			if v.IsAllowed(req.User) {
				allowedViews = append(allowedViews, v.Name())
			}
		}

		c.JSON(http.StatusOK, loginResponse{Token: tokenStr, User: req.User, AllowedViews: allowedViews})
	})
	if env.Config.LogConfig() {
		log.Printf("httpServer: %slogin -> serve login", r.BasePath())
	}
}

func disableLogin(r *gin.RouterGroup, config Config) {
	r.POST("login", func(c *gin.Context) {
		jsonErrorResponse(c, http.StatusServiceUnavailable, errors.New("Authentication module is disabled"))
	})
	if config.LogConfig() {
		log.Printf("httpServer: %slogin -> login disabled", r.BasePath())
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
		log.Printf("httpServer: login: auth file reloaded")
	}
}
