package testutil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestSetupMockHTTPServer(t *testing.T) {
	t.Run("creates HTTP server", func(t *testing.T) {
		server := SetupMockHTTPServer(t)

		// Verify server exists and has URL
		if server == nil {
			t.Fatal("SetupMockHTTPServer returned nil")
		}
		if server.URL == "" {
			t.Error("Server URL is empty")
		}
		if !strings.HasPrefix(server.URL, "http://") {
			t.Errorf("Expected HTTP URL, got %s", server.URL)
		}

		// Verify server is reachable (even without handlers)
		resp, err := http.Get(server.URL + "/nonexistent")
		if err != nil {
			t.Errorf("Failed to reach server: %v", err)
		} else {
			_ = resp.Body.Close()
			// Should get 404 for nonexistent endpoint
			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("Expected 404, got %d", resp.StatusCode)
			}
		}
	})

	t.Run("creates unique servers for multiple calls", func(t *testing.T) {
		server1 := SetupMockHTTPServer(t)
		server2 := SetupMockHTTPServer(t)

		if server1.URL == server2.URL {
			t.Error("SetupMockHTTPServer should create unique servers")
		}
	})
}

func TestSetupMockTLSServer(t *testing.T) {
	t.Run("creates HTTPS server", func(t *testing.T) {
		server := SetupMockTLSServer(t)

		// Verify server exists and has HTTPS URL
		if server == nil {
			t.Fatal("SetupMockTLSServer returned nil")
		}
		if server.URL == "" {
			t.Error("Server URL is empty")
		}
		if !strings.HasPrefix(server.URL, "https://") {
			t.Errorf("Expected HTTPS URL, got %s", server.URL)
		}

		// Verify server is reachable using the test client that accepts self-signed certs
		client := server.Client()
		resp, err := client.Get(server.URL + "/nonexistent")
		if err != nil {
			t.Errorf("Failed to reach TLS server: %v", err)
		} else {
			_ = resp.Body.Close()
			// Should get 404 for nonexistent endpoint
			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("Expected 404, got %d", resp.StatusCode)
			}
		}
	})
}

func TestAddJSONHandler(t *testing.T) {
	server := SetupMockHTTPServer(t)

	testData := map[string]interface{}{
		"message": "Hello, World!",
		"number":  42,
		"boolean": true,
	}

	t.Run("returns JSON response", func(t *testing.T) {
		server.AddJSONHandler("/test", http.StatusOK, testData)

		resp, err := http.Get(server.URL + "/test")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Verify status code
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Verify content type
		contentType := resp.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		// Verify response body
		var responseData map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		if responseData["message"] != testData["message"] {
			t.Errorf("Expected message %v, got %v", testData["message"], responseData["message"])
		}
		if responseData["number"] != float64(42) { // JSON numbers become float64
			t.Errorf("Expected number 42, got %v", responseData["number"])
		}
		if responseData["boolean"] != testData["boolean"] {
			t.Errorf("Expected boolean %v, got %v", testData["boolean"], responseData["boolean"])
		}
	})

	t.Run("handles nil response", func(t *testing.T) {
		server.AddJSONHandler("/empty", http.StatusNoContent, nil)

		resp, err := http.Get(server.URL + "/empty")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", resp.StatusCode)
		}
	})
}

func TestAddTextHandler(t *testing.T) {
	server := SetupMockHTTPServer(t)
	testText := "This is a plain text response"

	t.Run("returns text response", func(t *testing.T) {
		server.AddTextHandler("/text", http.StatusOK, testText)

		resp, err := http.Get(server.URL + "/text")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Verify status code
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Verify content type
		contentType := resp.Header.Get("Content-Type")
		if contentType != "text/plain" {
			t.Errorf("Expected Content-Type text/plain, got %s", contentType)
		}

		// Verify response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if string(body) != testText {
			t.Errorf("Expected body %q, got %q", testText, string(body))
		}
	})
}

func TestAddErrorHandler(t *testing.T) {
	server := SetupMockHTTPServer(t)
	errorMessage := "Something went wrong"

	t.Run("returns error response", func(t *testing.T) {
		server.AddErrorHandler("/error", http.StatusInternalServerError, errorMessage)

		resp, err := http.Get(server.URL + "/error")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Verify status code
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", resp.StatusCode)
		}

		// Verify response body contains error message
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if !strings.Contains(string(body), errorMessage) {
			t.Errorf("Expected body to contain %q, got %q", errorMessage, string(body))
		}
	})
}

func TestAddMalformedJSONHandler(t *testing.T) {
	server := SetupMockHTTPServer(t)

	t.Run("returns malformed JSON", func(t *testing.T) {
		server.AddMalformedJSONHandler("/malformed")

		resp, err := http.Get(server.URL + "/malformed")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Verify status code is OK
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Verify content type
		contentType := resp.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		// Verify that JSON parsing fails
		var responseData map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err == nil {
			t.Error("Expected JSON decoding to fail, but it succeeded")
		}
	})
}

func TestAddMethodHandler(t *testing.T) {
	server := SetupMockHTTPServer(t)
	testData := map[string]string{"method": "POST"}

	t.Run("responds to correct method", func(t *testing.T) {
		server.AddMethodHandler("/method", "POST", http.StatusOK, testData)

		resp, err := http.Post(server.URL+"/method", "application/json", nil)
		if err != nil {
			t.Fatalf("Failed to make POST request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("rejects incorrect method", func(t *testing.T) {
		server.AddMethodHandler("/method-only", "PUT", http.StatusOK, testData)

		resp, err := http.Get(server.URL + "/method-only")
		if err != nil {
			t.Fatalf("Failed to make GET request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", resp.StatusCode)
		}
	})
}

func TestAddAuthHandler(t *testing.T) {
	server := SetupMockHTTPServer(t)
	expectedAuth := "Bearer test-token"
	testData := map[string]string{"authenticated": "true"}

	t.Run("responds with correct auth", func(t *testing.T) {
		server.AddAuthHandler("/auth", expectedAuth, http.StatusOK, testData)

		req, err := http.NewRequest("GET", server.URL+"/auth", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Authorization", expectedAuth)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("rejects incorrect auth", func(t *testing.T) {
		server.AddAuthHandler("/auth-required", expectedAuth, http.StatusOK, testData)

		req, err := http.NewRequest("GET", server.URL+"/auth-required", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer wrong-token")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("rejects missing auth", func(t *testing.T) {
		server.AddAuthHandler("/auth-missing", expectedAuth, http.StatusOK, testData)

		resp, err := http.Get(server.URL + "/auth-missing")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}
	})
}

func TestAddCustomHandler(t *testing.T) {
	server := SetupMockHTTPServer(t)

	t.Run("uses custom handler function", func(t *testing.T) {
		customHandler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom-Header", "test-value")
			w.WriteHeader(http.StatusTeapot)
			_, _ = fmt.Fprint(w, "I'm a teapot")
		}

		server.AddCustomHandler("/custom", customHandler)

		resp, err := http.Get(server.URL + "/custom")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Verify custom status code
		if resp.StatusCode != http.StatusTeapot {
			t.Errorf("Expected status 418, got %d", resp.StatusCode)
		}

		// Verify custom header
		headerValue := resp.Header.Get("X-Custom-Header")
		if headerValue != "test-value" {
			t.Errorf("Expected X-Custom-Header test-value, got %s", headerValue)
		}

		// Verify custom body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if string(body) != "I'm a teapot" {
			t.Errorf("Expected body 'I'm a teapot', got %q", string(body))
		}
	})
}

func TestAddSlowHandler(t *testing.T) {
	server := SetupMockHTTPServer(t)

	t.Run("responds after delay", func(t *testing.T) {
		delayExecuted := false
		delayFunc := func() {
			delayExecuted = true
			time.Sleep(10 * time.Millisecond) // Small delay for testing
		}

		testData := map[string]string{"delayed": "true"}
		server.AddSlowHandler("/slow", http.StatusOK, testData, delayFunc)

		start := time.Now()
		resp, err := http.Get(server.URL + "/slow")
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Verify delay was executed
		if !delayExecuted {
			t.Error("Delay function was not executed")
		}

		// Verify some delay occurred (should be at least 10ms)
		if duration < 10*time.Millisecond {
			t.Errorf("Request completed too quickly: %v", duration)
		}

		// Verify response
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestHTTPResponseHelpers(t *testing.T) {
	t.Run("CreateHTTPSuccessResponse", func(t *testing.T) {
		content := "Test content"
		response := CreateHTTPSuccessResponse(content)

		if response["success"] != true {
			t.Errorf("Expected success to be true, got %v", response["success"])
		}
		if response["content"] != content {
			t.Errorf("Expected content %q, got %v", content, response["content"])
		}
		if response["tokens"] != len(content) {
			t.Errorf("Expected tokens %d, got %v", len(content), response["tokens"])
		}
	})

	t.Run("CreateHTTPErrorResponse", func(t *testing.T) {
		errorType := "test_error"
		message := "Test error message"
		response := CreateHTTPErrorResponse(errorType, message)

		errorData, ok := response["error"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected error to be a map")
		}

		if errorData["type"] != errorType {
			t.Errorf("Expected error type %q, got %v", errorType, errorData["type"])
		}
		if errorData["message"] != message {
			t.Errorf("Expected error message %q, got %v", message, errorData["message"])
		}
	})

	t.Run("CreateHTTPAuthErrorResponse", func(t *testing.T) {
		response := CreateHTTPAuthErrorResponse()

		errorData, ok := response["error"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected error to be a map")
		}

		if errorData["type"] != "authentication_error" {
			t.Errorf("Expected authentication_error, got %v", errorData["type"])
		}
	})

	t.Run("CreateHTTPRateLimitResponse", func(t *testing.T) {
		response := CreateHTTPRateLimitResponse()

		errorData, ok := response["error"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected error to be a map")
		}

		if errorData["type"] != "rate_limit_exceeded" {
			t.Errorf("Expected rate_limit_exceeded, got %v", errorData["type"])
		}
	})

	t.Run("CreateHTTPSafetyErrorResponse", func(t *testing.T) {
		response := CreateHTTPSafetyErrorResponse()

		errorData, ok := response["error"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected error to be a map")
		}

		if errorData["type"] != "safety_error" {
			t.Errorf("Expected safety_error, got %v", errorData["type"])
		}
	})

	t.Run("CreateHTTPQuotaErrorResponse", func(t *testing.T) {
		response := CreateHTTPQuotaErrorResponse()

		errorData, ok := response["error"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected error to be a map")
		}

		if errorData["type"] != "quota_exceeded" {
			t.Errorf("Expected quota_exceeded, got %v", errorData["type"])
		}
	})
}

func TestWithMockProvider(t *testing.T) {
	t.Run("provides server to test function", func(t *testing.T) {
		var capturedURL string
		testData := map[string]string{"test": "data"}

		WithMockProvider(t,
			func(server *MockHTTPServer) {
				server.AddJSONHandler("/test", http.StatusOK, testData)
			},
			func(baseURL string) {
				capturedURL = baseURL

				// Verify we can make requests to the provided URL
				resp, err := http.Get(baseURL + "/test")
				if err != nil {
					t.Errorf("Failed to make request to provided URL: %v", err)
					return
				}
				defer func() { _ = resp.Body.Close() }()

				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			})

		// Verify URL was captured
		if capturedURL == "" {
			t.Error("URL was not provided to test function")
		}
		if !strings.HasPrefix(capturedURL, "http://") {
			t.Errorf("Expected HTTP URL, got %s", capturedURL)
		}
	})
}

func TestWithMockTLSProvider(t *testing.T) {
	t.Run("provides TLS server to test function", func(t *testing.T) {
		var capturedURL string
		testData := map[string]string{"secure": "data"}

		WithMockTLSProvider(t,
			func(server *MockHTTPServer) {
				server.AddJSONHandler("/secure", http.StatusOK, testData)
			},
			func(baseURL string) {
				capturedURL = baseURL

				// We need to use the server's client for TLS requests
				// Note: In a real test, we'd need access to the server's client
				// For now, just verify the URL format
			})

		// Verify URL was captured and is HTTPS
		if capturedURL == "" {
			t.Error("URL was not provided to test function")
		}
		if !strings.HasPrefix(capturedURL, "https://") {
			t.Errorf("Expected HTTPS URL, got %s", capturedURL)
		}
	})
}

// TestIntegrationWorkflow demonstrates realistic usage patterns
func TestIntegrationWorkflow(t *testing.T) {
	t.Run("complete API testing workflow", func(t *testing.T) {
		server := SetupMockHTTPServer(t)

		// Setup multiple endpoints to simulate a real API
		server.AddJSONHandler("/api/success", http.StatusOK, CreateHTTPSuccessResponse("Test response"))
		server.AddJSONHandler("/api/auth-error", http.StatusUnauthorized, CreateHTTPAuthErrorResponse())
		server.AddJSONHandler("/api/rate-limit", http.StatusTooManyRequests, CreateHTTPRateLimitResponse())
		server.AddMalformedJSONHandler("/api/malformed")
		server.AddErrorHandler("/api/server-error", http.StatusInternalServerError, "Internal server error")

		// Setup auth endpoint
		server.AddAuthHandler("/api/secure", "Bearer valid-token", http.StatusOK,
			map[string]string{"data": "secure content"})

		// Setup method-specific endpoint
		server.AddMethodHandler("/api/post-only", "POST", http.StatusCreated,
			map[string]string{"created": "true"})

		// Test success endpoint
		resp, err := http.Get(server.URL + "/api/success")
		if err != nil {
			t.Errorf("Failed to test success endpoint: %v", err)
		} else {
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Success endpoint returned %d", resp.StatusCode)
			}
		}

		// Test auth error endpoint
		resp, err = http.Get(server.URL + "/api/auth-error")
		if err != nil {
			t.Errorf("Failed to test auth error endpoint: %v", err)
		} else {
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Auth error endpoint returned %d", resp.StatusCode)
			}
		}

		// Test auth success
		req, _ := http.NewRequest("GET", server.URL+"/api/secure", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("Failed to test secure endpoint: %v", err)
		} else {
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Secure endpoint returned %d", resp.StatusCode)
			}
		}

		// Test method validation
		resp, err = http.Post(server.URL+"/api/post-only", "application/json", nil)
		if err != nil {
			t.Errorf("Failed to test POST endpoint: %v", err)
		} else {
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusCreated {
				t.Errorf("POST endpoint returned %d", resp.StatusCode)
			}
		}
	})
}
