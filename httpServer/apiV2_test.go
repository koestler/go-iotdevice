package httpServer

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/koestler/go-iotdevice/v3/dataflow"
	"github.com/stretchr/testify/assert"
)

// mockConfig implements the Config interface for testing
type mockConfig struct {
	buildVersion    string
	bind            string
	port            int
	logRequests     bool
	logDebug        bool
	logConfig       bool
	frontendProxy   *url.URL
	frontendPath    string
	frontendExpires time.Duration
	configExpires   time.Duration
}

func (m *mockConfig) BuildVersion() string           { return m.buildVersion }
func (m *mockConfig) Bind() string                   { return m.bind }
func (m *mockConfig) Port() int                      { return m.port }
func (m *mockConfig) LogRequests() bool              { return m.logRequests }
func (m *mockConfig) LogDebug() bool                 { return m.logDebug }
func (m *mockConfig) LogConfig() bool                { return m.logConfig }
func (m *mockConfig) FrontendProxy() *url.URL        { return m.frontendProxy }
func (m *mockConfig) FrontendPath() string           { return m.frontendPath }
func (m *mockConfig) FrontendExpires() time.Duration { return m.frontendExpires }
func (m *mockConfig) ConfigExpires() time.Duration   { return m.configExpires }

// mockViewConfig implements the ViewConfig interface for testing
type mockViewConfig struct {
	name     string
	title    string
	devices  []ViewDeviceConfig
	autoplay bool
	isPublic bool
	hidden   bool
	allowed  map[string]bool
}

func (m *mockViewConfig) Name() string                { return m.name }
func (m *mockViewConfig) Title() string               { return m.title }
func (m *mockViewConfig) Devices() []ViewDeviceConfig { return m.devices }
func (m *mockViewConfig) Autoplay() bool              { return m.autoplay }
func (m *mockViewConfig) IsPublic() bool              { return m.isPublic }
func (m *mockViewConfig) Hidden() bool                { return m.hidden }
func (m *mockViewConfig) IsAllowed(user string) bool {
	if m.allowed == nil {
		return true
	}
	return m.allowed[user]
}

// mockRegisterFilterConf implements the RegisterFilterConf interface for testing
type mockRegisterFilterConf struct {
	includeRegisters  []string
	skipRegisters     []string
	includeCategories []string
	skipCategories    []string
	defaultInclude    bool
}

func (m *mockRegisterFilterConf) IncludeRegisters() []string  { return m.includeRegisters }
func (m *mockRegisterFilterConf) SkipRegisters() []string     { return m.skipRegisters }
func (m *mockRegisterFilterConf) IncludeCategories() []string { return m.includeCategories }
func (m *mockRegisterFilterConf) SkipCategories() []string    { return m.skipCategories }
func (m *mockRegisterFilterConf) DefaultInclude() bool        { return m.defaultInclude }

// mockViewDeviceConfig implements the ViewDeviceConfig interface for testing
type mockViewDeviceConfig struct {
	name   string
	title  string
	filter dataflow.RegisterFilterConf
}

func (m *mockViewDeviceConfig) Name() string                        { return m.name }
func (m *mockViewDeviceConfig) Title() string                       { return m.title }
func (m *mockViewDeviceConfig) Filter() dataflow.RegisterFilterConf { return m.filter }

// mockAuthenticationConfig implements the AuthenticationConfig interface for testing
type mockAuthenticationConfig struct {
	enabled           bool
	jwtSecret         []byte
	jwtValidityPeriod time.Duration
	htaccessFile      string
}

func (m *mockAuthenticationConfig) Enabled() bool                    { return m.enabled }
func (m *mockAuthenticationConfig) JwtSecret() []byte                { return m.jwtSecret }
func (m *mockAuthenticationConfig) JwtValidityPeriod() time.Duration { return m.jwtValidityPeriod }
func (m *mockAuthenticationConfig) HtaccessFile() string             { return m.htaccessFile }

// setupTestEnvironment creates a test environment with router
func setupTestEnvironment(t *testing.T) *Environment {
	t.Helper()

	config := &mockConfig{
		buildVersion:    "v0.0.0-test",
		bind:            "localhost",
		port:            8080,
		logRequests:     true,
		logDebug:        true,
		logConfig:       true,
		frontendPath:    "./frontend-build/",
		frontendExpires: 5 * time.Minute,
		configExpires:   1 * time.Minute,
	}

	dev0 := &mockViewDeviceConfig{
		name:  "dev0",
		title: "Test Device 0",
		filter: &mockRegisterFilterConf{
			defaultInclude: true,
		},
	}

	dev1 := &mockViewDeviceConfig{
		name:  "dev1",
		title: "Test Device 1",
		filter: &mockRegisterFilterConf{
			defaultInclude: true,
		},
	}

	dev2 := &mockViewDeviceConfig{
		name:  "dev2",
		title: "Test Device 2",
		filter: &mockRegisterFilterConf{
			defaultInclude: true,
		},
	}

	publicView := &mockViewConfig{
		name:     "public",
		title:    "Public View",
		devices:  []ViewDeviceConfig{dev0, dev1},
		autoplay: false,
		isPublic: true,
		hidden:   false,
	}

	privateView := &mockViewConfig{
		name:     "private",
		title:    "Private View",
		devices:  []ViewDeviceConfig{dev2},
		autoplay: true,
		isPublic: false,
		hidden:   false,
	}

	forbiddenView := &mockViewConfig{
		name:     "forbidden",
		title:    "Forbidden View",
		devices:  []ViewDeviceConfig{dev1},
		autoplay: true,
		isPublic: false,
		hidden:   false,
		allowed: map[string]bool{
			"not-me": true,
		},
	}

	authConfig := &mockAuthenticationConfig{
		enabled:           true,
		jwtSecret:         []byte("test-secret"),
		jwtValidityPeriod: 1 * time.Hour,
		htaccessFile:      setupTestHtaccessFile(t),
	}

	// Create register database
	registerDb := dataflow.NewRegisterDb()
	registerDb.Add(dataflow.NewRegisterStruct(
		"Monitor",
		"Temperature",
		"Room temperature",
		dataflow.NumberRegister,
		nil,
		"°C",
		100,
		false,
	))
	registerDb.Add(dataflow.NewRegisterStruct(
		"Control",
		"Setpoint",
		"Temperature setpoint",
		dataflow.NumberRegister,
		nil,
		"°C",
		200,
		true, // writable
	))

	env := &Environment{
		Config:         config,
		ProjectTitle:   "Test Project",
		Views:          []ViewConfig{publicView, privateView, forbiddenView},
		Authentication: authConfig,
		RegisterDbOfDevice: func(deviceName string) *dataflow.RegisterDb {
			return registerDb
		},
		StateStorage:   dataflow.NewValueStorage(),
		CommandStorage: dataflow.NewValueStorage(),
	}

	return env
}

func setupRouter(t *testing.T, env *Environment) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	addApiV2Routes(mux, env)

	handler := middlewares(mux.ServeHTTP, env)

	return handler
}

const testUser = "testuser"
const testPassword = "testpass123"

func setupTestHtaccessFile(t *testing.T) string {
	t.Helper()

	// Password hash for "testpass123" using bcrypt
	htaccessContent := "testuser:$2y$05$.lI/iuiP7VfmuHPGJNk0W.YeGw8Y2AjrwwB4OYvmdZciD6BXfYyLS\n"

	htaccessPath := filepath.Join(t.TempDir(), "test-auth.passwd")
	err := os.WriteFile(htaccessPath, []byte(htaccessContent), 0644)
	assert.NoError(t, err)

	return htaccessPath
}

func setupToken(t *testing.T, env *Environment) string {
	token, err := createJwtToken(env.Authentication, "testuser")
	assert.NoError(t, err)
	return token
}

// TestConfigFrontendEndpoint tests GET /api/v2/config/frontend
func TestConfigFrontendEndpoint(t *testing.T) {
	env := setupTestEnvironment(t)
	router := setupRouter(t, env)

	t.Run("content", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/config/frontend", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		assertCacheControlHeader(t, w, env.Config.ConfigExpires())

		var response configResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Response should be valid JSON")

		expected := configResponse{
			ProjectTitle:   "Test Project",
			BackendVersion: "v0.0.0-test",
			Views: []viewResponse{
				{
					Name:  "public",
					Title: "Public View",
					Devices: []deviceViewResponse{
						{Name: "dev0", Title: "Test Device 0"},
						{Name: "dev1", Title: "Test Device 1"},
					},
					Autoplay: false,
					IsPublic: true,
					Hidden:   false,
				}, {
					Name:  "private",
					Title: "Private View",
					Devices: []deviceViewResponse{
						{Name: "dev2", Title: "Test Device 2"},
					},
					Autoplay: true,
					IsPublic: false,
					Hidden:   false,
				}, {
					Name:  "forbidden",
					Title: "Forbidden View",
					Devices: []deviceViewResponse{
						{Name: "dev1", Title: "Test Device 1"},
					},
					Autoplay: true,
					IsPublic: false,
					Hidden:   false,
				},
			},
		}
		assert.Equal(t, expected, response)
	})

	t.Run("ETag", func(t *testing.T) {
		// First request
		req, _ := http.NewRequest("GET", "/api/v2/config/frontend", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		etag := w.Header().Get("ETag")
		assert.NotEmpty(t, etag, "ETag header should be present")

		// Second request with If-None-Match header
		req2, _ := http.NewRequest("GET", "/api/v2/config/frontend", nil)
		req2.Header.Set("If-None-Match", etag)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusNotModified, w2.Code, "Should return 304 Not Modified when ETag matches")
	})
}

func assertCacheControlHeader(t *testing.T, w *httptest.ResponseRecorder, expected time.Duration) {
	if expected > 0 {
		expectedHeader := fmt.Sprintf("public, max-age=%.0f", expected.Seconds())
		cacheControl := w.Header().Get("Cache-Control")
		assert.Equal(t, expectedHeader, cacheControl, "Cache-Control header should match expected value")
	}
}

// TestLoginEndpoint tests POST /api/v2/auth/login with various scenarios
func TestLoginEndpoint(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		env := setupTestEnvironment(t)
		router := setupRouter(t, env)

		loginReq := loginRequest{
			User:     testUser,
			Password: testPassword,
		}
		body, _ := json.Marshal(loginReq)

		req, _ := http.NewRequest("POST", "/api/v2/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK for successful login")

		var response loginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, loginReq.User, response.User)
		assert.Equal(t, []string{"public", "private"}, response.AllowedViews)

		assert.NotEmpty(t, response.Token)
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		env := setupTestEnvironment(t)
		router := setupRouter(t, env)

		loginReq := loginRequest{
			User:     "testuser",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(loginReq)

		req, _ := http.NewRequest("POST", "/api/v2/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected status 401 Unauthorized for invalid credentials")

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response.Message, "Invalid credentials")
	})

	t.Run("Disabled", func(t *testing.T) {
		env := setupTestEnvironment(t)

		// Disable authentication
		env.Authentication = &mockAuthenticationConfig{
			enabled: false,
		}

		router := setupRouter(t, env)

		loginReq := loginRequest{
			User:     "testuser",
			Password: "testpass",
		}
		body, _ := json.Marshal(loginReq)

		req, _ := http.NewRequest("POST", "/api/v2/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code, "Expected status 503 when auth is disabled")

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response.Message, "disabled")
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		env := setupTestEnvironment(t)
		router := setupRouter(t, env)

		req, _ := http.NewRequest("POST", "/api/v2/auth/login", bytes.NewBuffer([]byte("invalid json")))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected status 422 for invalid JSON")
	})

	t.Run("MissingHtaccessFile", func(t *testing.T) {
		env := setupTestEnvironment(t)

		// Enable authentication but don't provide valid htaccess file
		env.Authentication = &mockAuthenticationConfig{
			enabled:           true,
			jwtSecret:         []byte("test-secret"),
			jwtValidityPeriod: 1 * time.Hour,
			htaccessFile:      "/tmp/nonexistent.passwd",
		}

		router := setupRouter(t, env)

		loginReq := loginRequest{
			User:     "testuser",
			Password: "testpass123",
		}
		body, _ := json.Marshal(loginReq)

		req, _ := http.NewRequest("POST", "/api/v2/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should get 503 because htaccess file doesn't exist
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestRegistersEndpoint tests GET /api/v2/views/{viewName}/devices/{deviceName}/registers with various scenarios
func TestRegistersEndpoint(t *testing.T) {
	env := setupTestEnvironment(t)
	router := setupRouter(t, env)
	t.Run("okPublic", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/views/public/devices/dev0/registers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")
		assertCacheControlHeader(t, w, RegistersExpires)

		var response []registerResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Response should be valid JSON")

		expected := []registerResponse{
			{
				Category:    "Monitor",
				Name:        "Temperature",
				Description: "Room temperature",
				Type:        "number",
				Unit:        "°C",
				Sort:        100,
				Writable:    false,
			},
			{
				Category:    "Control",
				Name:        "Setpoint",
				Description: "Temperature setpoint",
				Type:        "number",
				Unit:        "°C",
				Sort:        200,
				Writable:    true,
			},
		}

		assert.Equal(t, expected, response)
	})

	t.Run("okPrivate", func(t *testing.T) {
		token := setupToken(t, env)
		req, _ := http.NewRequest("GET", "/api/v2/views/private/devices/dev2/registers", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var response []registerResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Response should be valid JSON")
		assert.Len(t, response, 2)
	})

	t.Run("UnauthorizedPrivate", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/views/private/devices/dev2/registers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Expected status 403 Forbidden for private view without token")
	})

	t.Run("ForbiddenPrivate", func(t *testing.T) {
		token := setupToken(t, env)
		req, _ := http.NewRequest("GET", "/api/v2/views/forbidden/devices/dev1/registers", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Expected status 403 Forbidden for private view without token")
	})

	t.Run("NonexistentDevice", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/views/public/devices/nonexistent-device/registers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status 404 Not Found")
	})
}

// TestValuesGetEndpoint tests GET /api/v2/views/{viewName}/devices/{deviceName}/values
func TestValuesGetEndpoint(t *testing.T) {
	env := setupTestEnvironment(t)
	router := setupRouter(t, env)
	token := setupToken(t, env)

	// Add a value to the state storage
	devName := "dev0"
	registerDb := env.RegisterDbOfDevice(devName)
	registers := registerDb.GetAll()
	for _, reg := range registers {
		value := dataflow.NewNumericRegisterValue(devName, reg, 25.5)
		env.StateStorage.Fill(value)
	}
	env.StateStorage.Wait()

	t.Run("okPublic", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/views/public/devices/dev0/values", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var response values1DResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Response should be valid JSON")

		// Check if the value is present
		assert.Contains(t, response, "Temperature")
		assert.Equal(t, 25.5, response["Temperature"])
	})

	t.Run("okPrivate", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/views/private/devices/dev2/values", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")
	})

	t.Run("wrongToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/views/private/devices/dev2/values", nil)
		req.Header.Set("Authorization", "invalid")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Expected status 403 Forbidden for invalid token")
	})

	t.Run("UnauthorizedPrivate", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/views/private/devices/dev2/values", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Expected status 403 Forbidden for private view without token")
	})

	t.Run("ForbiddenPrivate", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/views/forbidden/devices/dev1/values", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Expected status 403 Forbidden for private view without token")
	})
}

// TestValuesPatchEndpoint tests PATCH /api/v2/views/{viewName}/devices/{deviceName}/values
func TestValuesPatchEndpoint(t *testing.T) {
	env := setupTestEnvironment(t)
	router := setupRouter(t, env)
	token := setupToken(t, env)

	t.Run("okPublic", func(t *testing.T) {
		patchData := map[string]interface{}{
			"Setpoint": 22.0,
		}
		body, _ := json.Marshal(patchData)

		req, _ := http.NewRequest("PATCH", "/api/v2/views/public/devices/dev0/values", bytes.NewBuffer(body))
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")
	})

	t.Run("ForbiddenWithoutAuth", func(t *testing.T) {
		patchData := map[string]interface{}{
			"Setpoint": 22.0,
		}
		body, _ := json.Marshal(patchData)

		req, _ := http.NewRequest("PATCH", "/api/v2/views/public/devices/dev0/values", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Expected status 403 Forbidden without authentication")
	})

	t.Run("nonWritableRegister", func(t *testing.T) {
		patchData := map[string]interface{}{
			"Temperature": 22.0,
		}
		body, _ := json.Marshal(patchData)

		req, _ := http.NewRequest("PATCH", "/api/v2/views/public/devices/dev0/values", bytes.NewBuffer(body))
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Expected status 403 Forbidden without authentication")
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", "/api/v2/views/public/devices/dev0/values", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected status 422 Unprocessable Entity")
	})

	t.Run("InvalidRegister", func(t *testing.T) {
		patchData := map[string]interface{}{
			"NonexistentRegister": 42.0,
		}
		body, _ := json.Marshal(patchData)

		req, _ := http.NewRequest("PATCH", "/api/v2/views/public/devices/dev0/values", bytes.NewBuffer(body))
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected status 422 Unprocessable Entity")
	})

	t.Run("NonexistentDevice", func(t *testing.T) {
		patchData := map[string]interface{}{
			"Setpoint": 22.0,
		}
		body, _ := json.Marshal(patchData)

		req, _ := http.NewRequest("PATCH", "/api/v2/views/public/devices/nonexistent-device/values", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status 404 Not Found")
	})
}
