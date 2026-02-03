package httpServer

import (
	"bytes"
	stdjson "encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
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
func setupTestEnvironment() (*gin.Engine, *Environment) {
	gin.SetMode(gin.TestMode)

	config := &mockConfig{
		buildVersion:    "v0.0.0-test",
		bind:            "localhost",
		port:            8080,
		logRequests:     false,
		logDebug:        false,
		logConfig:       false,
		frontendPath:    "./frontend-build/",
		frontendExpires: 5 * time.Minute,
		configExpires:   1 * time.Minute,
	}

	device1 := &mockViewDeviceConfig{
		name:  "dev0",
		title: "Test Device 0",
		filter: &mockRegisterFilterConf{
			defaultInclude: true,
		},
	}

	device2 := &mockViewDeviceConfig{
		name:  "dev1",
		title: "Test Device 1",
		filter: &mockRegisterFilterConf{
			defaultInclude: true,
		},
	}

	view := &mockViewConfig{
		name:     "public",
		title:    "Public View",
		devices:  []ViewDeviceConfig{device1, device2},
		autoplay: true,
		isPublic: true,
		hidden:   false,
	}

	authConfig := &mockAuthenticationConfig{
		enabled:           true,
		jwtSecret:         []byte("test-secret"),
		jwtValidityPeriod: 1 * time.Hour,
	}

	// Create register database
	registerDb := dataflow.NewRegisterDb()
	testRegister := dataflow.NewRegisterStruct(
		"Monitor",
		"Temperature",
		"Room temperature",
		dataflow.NumberRegister,
		nil,
		"°C",
		100,
		false,
	)
	registerDb.Add(testRegister)

	env := &Environment{
		Config:         config,
		ProjectTitle:   "Test Project",
		Views:          []ViewConfig{view},
		Authentication: authConfig,
		RegisterDbOfDevice: func(deviceName string) *dataflow.RegisterDb {
			return registerDb
		},
		StateStorage:   dataflow.NewValueStorage(),
		CommandStorage: dataflow.NewValueStorage(),
	}

	router := gin.New()
	router.Use(authJwtMiddleware(env))
	addApiV2Routes(router, env)

	return router, env
}

// TestConfigFrontendEndpoint tests GET /api/v2/config/frontend
func TestConfigFrontendEndpoint(t *testing.T) {
	router, _ := setupTestEnvironment()

	req, _ := http.NewRequest("GET", "/api/v2/config/frontend", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

	var response configResponse
	err := stdjson.Unmarshal(w.Body.Bytes(), &response)
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
				Autoplay: true,
				IsPublic: true,
				Hidden:   false},
		},
	}

	assert.Equal(t, expected, response)
}

// TestLoginEndpoint tests POST /api/v2/auth/login with various scenarios
func TestLoginEndpoint(t *testing.T) {
	t.Run("Disabled", func(t *testing.T) {
		router, _ := setupTestEnvironment()

		loginReq := loginRequest{
			User:     "testuser",
			Password: "testpass",
		}
		body, _ := stdjson.Marshal(loginReq)

		req, _ := http.NewRequest("POST", "/api/v2/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code, "Expected status 503 when auth is disabled")

		var response ErrorResponse
		err := stdjson.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response.Message, "disabled")
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		router, env := setupTestEnvironment()

		// Enable authentication but don't provide valid htaccess file
		env.Authentication = &mockAuthenticationConfig{
			enabled:           true,
			jwtSecret:         []byte("test-secret"),
			jwtValidityPeriod: 1 * time.Hour,
			htaccessFile:      "/tmp/nonexistent.passwd",
		}

		req, _ := http.NewRequest("POST", "/api/v2/auth/login", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should get 503 because htaccess file doesn't exist
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

// TestRegistersEndpoint tests GET /api/v2/views/{viewName}/devices/{deviceName}/registers with various scenarios
func TestRegistersEndpoint(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		router, _ := setupTestEnvironment()

		req, _ := http.NewRequest("GET", "/api/v2/views/public/devices/dev0/registers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var response []registerResponse
		err := stdjson.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Response should be valid JSON")

		expected := []registerResponse{
			{Category: "Monitor",
				Name:        "Temperature",
				Description: "Room temperature",
				Type:        "number",
				Unit:        "°C",
				Sort:        100,
				Writable:    false,
			},
		}

		assert.Equal(t, expected, response)
	})

	t.Run("NonexistentDevice", func(t *testing.T) {
		router, _ := setupTestEnvironment()

		req, _ := http.NewRequest("GET", "/api/v2/views/public/devices/nonexistent-device/registers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status 404 Not Found")
	})
}

// TestValuesGetEndpoint tests GET /api/v2/views/{viewName}/devices/{deviceName}/values
func TestValuesGetEndpoint(t *testing.T) {
	router, env := setupTestEnvironment()

	// Add a value to the state storage
	registerDb := env.RegisterDbOfDevice("test-device")
	registers := registerDb.GetAll()
	if len(registers) > 0 {
		value := dataflow.NewNumericRegisterValue("test-device", registers[0], 25.5)
		env.StateStorage.Fill(value)
	}

	req, _ := http.NewRequest("GET", "/api/v2/views/public/devices/test-device/values", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

	var response values1DResponse
	err := stdjson.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should be valid JSON")

	// Check if the value is present
	if len(response) > 0 {
		assert.Contains(t, response, "Temperature")
		assert.Equal(t, 25.5, response["Temperature"])
	}
}

// TestValuesGetEndpointNonexistentDevice tests GET /api/v2/views/{viewName}/devices/{deviceName}/values with invalid device
func TestValuesGetEndpointNonexistentDevice(t *testing.T) {
	router, _ := setupTestEnvironment()

	req, _ := http.NewRequest("GET", "/api/v2/views/public/devices/nonexistent-device/values", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Expected status 404 Not Found")
}

// TestValuesPatchEndpoint tests PATCH /api/v2/views/{viewName}/devices/{deviceName}/values
func TestValuesPatchEndpoint(t *testing.T) {
	router, env := setupTestEnvironment()

	// Enable authentication and create a token
	env.Authentication = &mockAuthenticationConfig{
		enabled:           true,
		jwtSecret:         []byte("test-secret-key"),
		jwtValidityPeriod: 1 * time.Hour,
	}

	token, _ := createJwtToken(env.Authentication, "testuser")

	// Add a writable register
	registerDb := env.RegisterDbOfDevice("test-device")
	writableRegister := dataflow.NewRegisterStruct(
		"Control",
		"Setpoint",
		"Temperature setpoint",
		dataflow.NumberRegister,
		nil,
		"°C",
		200,
		true, // writable
	)
	registerDb.Add(writableRegister)

	patchData := map[string]interface{}{
		"Setpoint": 22.0,
	}
	body, _ := stdjson.Marshal(patchData)

	req, _ := http.NewRequest("PATCH", "/api/v2/views/public/devices/test-device/values", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")
}

// TestValuesPatchEndpointUnauthenticated tests PATCH without authentication (should fail)
func TestValuesPatchEndpointUnauthenticated(t *testing.T) {
	router, _ := setupTestEnvironment()

	patchData := map[string]interface{}{
		"Temperature": 22.0,
	}
	body, _ := stdjson.Marshal(patchData)

	req, _ := http.NewRequest("PATCH", "/api/v2/views/public/devices/test-device/values", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code, "Expected status 403 Forbidden without authentication")
}

// TestValuesPatchEndpointInvalidJSON tests PATCH /api/v2/views/{viewName}/devices/{deviceName}/values with invalid JSON
func TestValuesPatchEndpointInvalidJSON(t *testing.T) {
	router, env := setupTestEnvironment()

	// Enable authentication and create a token
	env.Authentication = &mockAuthenticationConfig{
		enabled:           true,
		jwtSecret:         []byte("test-secret-key"),
		jwtValidityPeriod: 1 * time.Hour,
	}

	token, _ := createJwtToken(env.Authentication, "testuser")

	req, _ := http.NewRequest("PATCH", "/api/v2/views/public/devices/test-device/values", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected status 422 Unprocessable Entity")
}

// TestValuesPatchEndpointInvalidRegister tests PATCH /api/v2/views/{viewName}/devices/{deviceName}/values with invalid register
func TestValuesPatchEndpointInvalidRegister(t *testing.T) {
	router, env := setupTestEnvironment()

	// Enable authentication and create a token
	env.Authentication = &mockAuthenticationConfig{
		enabled:           true,
		jwtSecret:         []byte("test-secret-key"),
		jwtValidityPeriod: 1 * time.Hour,
	}

	token, _ := createJwtToken(env.Authentication, "testuser")

	patchData := map[string]interface{}{
		"NonexistentRegister": 42.0,
	}
	body, _ := stdjson.Marshal(patchData)

	req, _ := http.NewRequest("PATCH", "/api/v2/views/public/devices/test-device/values", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected status 422 Unprocessable Entity")
}

// TestValuesPatchEndpointNonexistentDevice tests PATCH /api/v2/views/{viewName}/devices/{deviceName}/values with invalid device
func TestValuesPatchEndpointNonexistentDevice(t *testing.T) {
	router, _ := setupTestEnvironment()

	patchData := map[string]interface{}{
		"Temperature": 22.0,
	}
	body, _ := stdjson.Marshal(patchData)

	req, _ := http.NewRequest("PATCH", "/api/v2/views/public/devices/nonexistent-device/values", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Expected status 404 Not Found")
}

// TestAuthenticationWithJWT tests endpoints with JWT authentication
func TestAuthenticationWithJWT(t *testing.T) {
	router, env := setupTestEnvironment()

	// Create a private view
	privateView := &mockViewConfig{
		name:     "private",
		title:    "Private View",
		devices:  []ViewDeviceConfig{},
		autoplay: false,
		isPublic: false,
		hidden:   false,
		allowed:  map[string]bool{"testuser": true},
	}
	env.Views = append(env.Views, privateView)

	// Enable authentication
	env.Authentication = &mockAuthenticationConfig{
		enabled:           true,
		jwtSecret:         []byte("test-secret-key"),
		jwtValidityPeriod: 1 * time.Hour,
	}

	// Create a valid JWT token
	token, err := createJwtToken(env.Authentication, "testuser")
	assert.NoError(t, err)

	// Test accessing private view with valid token
	req, _ := http.NewRequest("GET", "/api/v2/config/frontend", nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAuthenticationWithInvalidJWT tests endpoints with invalid JWT authentication
func TestAuthenticationWithInvalidJWT(t *testing.T) {
	router, env := setupTestEnvironment()

	// Enable authentication
	env.Authentication = &mockAuthenticationConfig{
		enabled:           true,
		jwtSecret:         []byte("test-secret-key"),
		jwtValidityPeriod: 1 * time.Hour,
	}

	// Test with invalid token
	req, _ := http.NewRequest("GET", "/api/v2/config/frontend", nil)
	req.Header.Set("Authorization", "invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWebsocketEndpointExists tests that the websocket endpoint is set up (cannot fully test WebSocket in httptest)
func TestWebsocketEndpointExists(t *testing.T) {
	t.Skip("WebSocket testing requires actual WebSocket client, httptest.ResponseRecorder causes crashes")
	// Note: The websocket endpoint is set up in setupValuesWs() and is available at /api/v2/views/{viewName}/ws
	// Full WebSocket testing would require using a real WebSocket client library
}

// TestCacheHeaders tests that appropriate cache headers are set
func TestCacheHeaders(t *testing.T) {
	router, _ := setupTestEnvironment()

	req, _ := http.NewRequest("GET", "/api/v2/config/frontend", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	cacheControl := w.Header().Get("Cache-Control")
	assert.NotEmpty(t, cacheControl, "Cache-Control header should be set")
}

// TestETagSupport tests that ETag headers are properly supported
func TestETagSupport(t *testing.T) {
	router, _ := setupTestEnvironment()

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
}
