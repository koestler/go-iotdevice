package httpServer

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetupDocs(t *testing.T) {
	tempDir := t.TempDir()
	const content = "swagger: '2.0'"
	assert.NoError(t, os.WriteFile(filepath.Join(tempDir, "swagger.yaml"), []byte(content), 0644))

	mux := http.NewServeMux()
	setupDocs(mux, &Environment{Config: &mockConfig{
		docsPath:        tempDir,
		frontendExpires: time.Hour,
	}})

	t.Run("existing file", func(t *testing.T) {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v2/docs/swagger.yaml", nil))
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, content, w.Body.String())
	})

	t.Run("missing file", func(t *testing.T) {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v2/docs/swagger.json", nil))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
