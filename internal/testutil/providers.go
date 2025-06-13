package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockHTTPServer wraps httptest.Server with additional test utilities
type MockHTTPServer struct {
	*httptest.Server
	Handler *http.ServeMux
}

// SetupMockHTTPServer creates a mock HTTP server for testing external API calls.
// The server is automatically torn down when the test completes via t.Cleanup.
func SetupMockHTTPServer(t testing.TB) *MockHTTPServer {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	t.Cleanup(func() {
		server.Close()
	})

	return &MockHTTPServer{
		Server:  server,
		Handler: mux,
	}
}

// SetupMockTLSServer creates a mock HTTPS server for testing external API calls.
// The server is automatically torn down when the test completes via t.Cleanup.
func SetupMockTLSServer(t testing.TB) *MockHTTPServer {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewTLSServer(mux)

	t.Cleanup(func() {
		server.Close()
	})

	return &MockHTTPServer{
		Server:  server,
		Handler: mux,
	}
}

// AddJSONHandler adds a handler that returns JSON response for the given path
func (m *MockHTTPServer) AddJSONHandler(path string, statusCode int, response interface{}) {
	m.Handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		if response != nil {
			if err := json.NewEncoder(w).Encode(response); err != nil {
				// If encoding fails, return error response
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprintf(w, `{"error": "encoding failed: %s"}`, err.Error())
			}
		}
	})
}

// AddTextHandler adds a handler that returns plain text response for the given path
func (m *MockHTTPServer) AddTextHandler(path string, statusCode int, response string) {
	m.Handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(statusCode)
		_, _ = fmt.Fprint(w, response)
	})
}

// AddErrorHandler adds a handler that returns an HTTP error for the given path
func (m *MockHTTPServer) AddErrorHandler(path string, statusCode int, errorMessage string) {
	m.Handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, errorMessage, statusCode)
	})
}

// AddMalformedJSONHandler adds a handler that returns malformed JSON for testing error handling
func (m *MockHTTPServer) AddMalformedJSONHandler(path string) {
	m.Handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"malformed": json, "missing": "quotes"}`)
	})
}

// AddTimeoutHandler adds a handler that never responds (simulates timeout)
func (m *MockHTTPServer) AddTimeoutHandler(path string) {
	m.Handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		// Never write response to simulate timeout
		select {}
	})
}

// AddSlowHandler adds a handler that responds after a delay (for testing timeouts)
func (m *MockHTTPServer) AddSlowHandler(path string, statusCode int, response interface{}, delayFunc func()) {
	m.Handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if delayFunc != nil {
			delayFunc()
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if response != nil {
			_ = json.NewEncoder(w).Encode(response)
		}
	})
}

// AddCustomHandler adds a custom handler function for the given path
func (m *MockHTTPServer) AddCustomHandler(path string, handler http.HandlerFunc) {
	m.Handler.HandleFunc(path, handler)
}

// AddMethodHandler adds a handler that only responds to specific HTTP methods
func (m *MockHTTPServer) AddMethodHandler(path string, method string, statusCode int, response interface{}) {
	m.Handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, fmt.Sprintf("Method %s not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if response != nil {
			_ = json.NewEncoder(w).Encode(response)
		}
	})
}

// AddAuthHandler adds a handler that requires specific authorization header
func (m *MockHTTPServer) AddAuthHandler(path string, expectedAuth string, statusCode int, response interface{}) {
	m.Handler.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != expectedAuth {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if response != nil {
			_ = json.NewEncoder(w).Encode(response)
		}
	})
}

// Common response helpers for HTTP API testing scenarios

// CreateHTTPSuccessResponse creates a standard HTTP success response structure
func CreateHTTPSuccessResponse(content string) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"content": content,
		"tokens":  len(content),
	}
}

// CreateHTTPErrorResponse creates a standard HTTP error response structure
func CreateHTTPErrorResponse(errorType, message string) map[string]interface{} {
	return map[string]interface{}{
		"error": map[string]interface{}{
			"type":    errorType,
			"message": message,
		},
	}
}

// CreateHTTPAuthErrorResponse creates an HTTP authentication error response
func CreateHTTPAuthErrorResponse() map[string]interface{} {
	return CreateHTTPErrorResponse("authentication_error", "Invalid API key provided")
}

// CreateHTTPRateLimitResponse creates an HTTP rate limit error response
func CreateHTTPRateLimitResponse() map[string]interface{} {
	return CreateHTTPErrorResponse("rate_limit_exceeded", "Rate limit exceeded, please try again later")
}

// CreateHTTPSafetyErrorResponse creates an HTTP safety/content filter error response
func CreateHTTPSafetyErrorResponse() map[string]interface{} {
	return CreateHTTPErrorResponse("safety_error", "Content filtered due to safety policies")
}

// CreateHTTPQuotaErrorResponse creates an HTTP quota exceeded error response
func CreateHTTPQuotaErrorResponse() map[string]interface{} {
	return CreateHTTPErrorResponse("quota_exceeded", "API quota exceeded")
}

// WithMockProvider executes a function with a mock HTTP server configured as a provider.
// This is useful for table-driven tests or when you need to scope the server usage.
func WithMockProvider(t testing.TB, setupFunc func(*MockHTTPServer), testFunc func(baseURL string)) {
	t.Helper()

	server := SetupMockHTTPServer(t)
	if setupFunc != nil {
		setupFunc(server)
	}
	testFunc(server.URL)
}

// WithMockTLSProvider executes a function with a mock HTTPS server configured as a provider.
func WithMockTLSProvider(t testing.TB, setupFunc func(*MockHTTPServer), testFunc func(baseURL string)) {
	t.Helper()

	server := SetupMockTLSServer(t)
	if setupFunc != nil {
		setupFunc(server)
	}
	testFunc(server.URL)
}
