package httpServer

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const (
	testIndexHTMLContent = "<html><body>Index</body></html>"
	testStyleCSSContent  = "body { margin: 0; }"
)

func TestSetupFrontend(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tempDir := t.TempDir()

	// Create test files
	indexHTML := filepath.Join(tempDir, "index.html")
	err := os.WriteFile(indexHTML, []byte(testIndexHTMLContent), 0644)
	assert.NoError(t, err)

	styleCSS := filepath.Join(tempDir, "style.css")
	err = os.WriteFile(styleCSS, []byte(testStyleCSSContent), 0644)
	assert.NoError(t, err)

	config := &mockConfig{
		buildVersion:    "test-version",
		bind:            "localhost",
		port:            8080,
		logRequests:     false,
		logDebug:        false,
		logConfig:       false,
		frontendProxy:   nil,
		frontendPath:    tempDir,
		frontendExpires: time.Hour,
		configExpires:   time.Minute,
	}

	views := []ViewConfig{
		&mockViewConfig{
			name:     "private",
			title:    "Private View",
			devices:  []ViewDeviceConfig{},
			autoplay: false,
			isPublic: false,
			hidden:   false,
			allowed:  nil,
		},
		&mockViewConfig{
			name:     "public",
			title:    "Public View",
			devices:  []ViewDeviceConfig{},
			autoplay: false,
			isPublic: true,
			hidden:   false,
			allowed:  nil,
		},
	}

	engine := gin.New()
	setupFrontend(engine, config, views)

	t.Run("ServeStyleCSS", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/style.css", nil)
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), testStyleCSSContent)
	})

	t.Run("ServeRootAsIndex", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), testIndexHTMLContent)
	})

	t.Run("ServePrivateViewRouteAsIndex", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/private", nil)
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), testIndexHTMLContent)
	})

	t.Run("ServePublicViewRouteAsIndex", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/public", nil)
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), testIndexHTMLContent)
	})

	t.Run("ServeLoginRouteAsIndex", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/login", nil)
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), testIndexHTMLContent)
	})

	t.Run("NotFoundRoute", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/nonexistent", nil)
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "route not found")
	})
}

func TestSetupFrontendEmptyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock config with empty frontend path
	config := &mockConfig{
		buildVersion:    "test-version",
		bind:            "localhost",
		port:            8080,
		logRequests:     false,
		logDebug:        false,
		logConfig:       false,
		frontendProxy:   nil,
		frontendPath:    "",
		frontendExpires: time.Hour,
		configExpires:   time.Minute,
	}

	views := []ViewConfig{}

	engine := gin.New()
	setupFrontend(engine, config, views)

	t.Run("NotFoundWhenNoFrontend", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
