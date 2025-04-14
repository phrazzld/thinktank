// internal/gemini/model_info_test.go
// Tests for fetchModelInfo and GetModelInfo methods
package gemini

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestFetchModelInfo(t *testing.T) {
	// This test verifies that fetchModelInfo correctly:
	// - Builds the proper URL based on endpoint and model name
	// - Makes an HTTP request using the injected HTTP client
	// - Handles success responses by parsing the JSON into ModelInfo
	// - Handles various error conditions (network errors, HTTP status errors, JSON parsing errors)

	// Test constants
	const (
		testAPIKey         = "test-api-key"
		testModelName      = "test-model"
		testCustomEndpoint = "https://custom-endpoint.example.com"
	)

	t.Run("URL construction with default endpoint", func(t *testing.T) {
		ctx := context.Background()
		mockClient := &http.Client{
			Transport: &mockHTTPTransport{
				inspectRequest: func(req *http.Request) {
					// URL should contain model name and API key
					expectedURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s?key=%s", testModelName, testAPIKey)
					if req.URL.String() != expectedURL {
						t.Errorf("Expected URL %q, got %q", expectedURL, req.URL.String())
					}
				},
				response: createSuccessResponse(ModelDetailsResponse{
					Name:             "models/test-model",
					InputTokenLimit:  32000,
					OutputTokenLimit: 8192,
				}),
			},
		}

		client := &geminiClient{
			apiKey:     testAPIKey,
			httpClient: mockClient,
			logger:     getTestLogger(),
		}

		info, err := client.fetchModelInfo(ctx, testModelName)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if info.Name != "models/test-model" {
			t.Errorf("Expected model name %q, got %q", "models/test-model", info.Name)
		}

		if info.InputTokenLimit != 32000 {
			t.Errorf("Expected input token limit %d, got %d", 32000, info.InputTokenLimit)
		}

		if info.OutputTokenLimit != 8192 {
			t.Errorf("Expected output token limit %d, got %d", 8192, info.OutputTokenLimit)
		}
	})

	t.Run("URL construction with custom endpoint", func(t *testing.T) {
		ctx := context.Background()
		mockClient := &http.Client{
			Transport: &mockHTTPTransport{
				inspectRequest: func(req *http.Request) {
					// URL should use custom endpoint
					expectedURL := fmt.Sprintf("%s/v1beta/models/%s", testCustomEndpoint, testModelName)
					if req.URL.String() != expectedURL {
						t.Errorf("Expected URL %q, got %q", expectedURL, req.URL.String())
					}
				},
				response: createSuccessResponse(ModelDetailsResponse{
					Name:             "models/test-model",
					InputTokenLimit:  32000,
					OutputTokenLimit: 8192,
				}),
			},
		}

		client := &geminiClient{
			apiKey:      testAPIKey,
			apiEndpoint: testCustomEndpoint,
			httpClient:  mockClient,
			logger:      getTestLogger(),
		}

		info, err := client.fetchModelInfo(ctx, testModelName)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if info.Name != "models/test-model" {
			t.Errorf("Expected model name %q, got %q", "models/test-model", info.Name)
		}
	})

	t.Run("Custom endpoint with trailing slash", func(t *testing.T) {
		ctx := context.Background()

		// Endpoint with trailing slash should be handled correctly
		customEndpointWithSlash := testCustomEndpoint + "/"

		mockClient := &http.Client{
			Transport: &mockHTTPTransport{
				inspectRequest: func(req *http.Request) {
					// URL should use custom endpoint without trailing slash
					expectedURL := fmt.Sprintf("%s/v1beta/models/%s", testCustomEndpoint, testModelName)
					if req.URL.String() != expectedURL {
						t.Errorf("Expected URL %q, got %q", expectedURL, req.URL.String())
					}
				},
				response: createSuccessResponse(ModelDetailsResponse{
					Name:             "models/test-model",
					InputTokenLimit:  32000,
					OutputTokenLimit: 8192,
				}),
			},
		}

		client := &geminiClient{
			apiKey:      testAPIKey,
			apiEndpoint: customEndpointWithSlash,
			httpClient:  mockClient,
			logger:      getTestLogger(),
		}

		_, err := client.fetchModelInfo(ctx, testModelName)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})

	t.Run("Success response parsing", func(t *testing.T) {
		ctx := context.Background()

		// Create response with detailed model information
		mockResponse := createSuccessResponse(ModelDetailsResponse{
			Name:                       "models/custom-model",
			BaseModelID:                "gemini-pro",
			Version:                    "001",
			DisplayName:                "Custom Model",
			Description:                "A test model",
			InputTokenLimit:            12345,
			OutputTokenLimit:           6789,
			SupportedGenerationMethods: []string{"generateContent", "countTokens"},
			Temperature:                0.5,
			TopP:                       0.95,
			TopK:                       40,
		})

		client := &geminiClient{
			apiKey:     testAPIKey,
			httpClient: newMockHTTPClient(mockResponse, nil),
			logger:     getTestLogger(),
		}

		info, err := client.fetchModelInfo(ctx, testModelName)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify the parsed fields
		if info.Name != "models/custom-model" {
			t.Errorf("Expected model name %q, got %q", "models/custom-model", info.Name)
		}

		if info.InputTokenLimit != 12345 {
			t.Errorf("Expected input token limit %d, got %d", 12345, info.InputTokenLimit)
		}

		if info.OutputTokenLimit != 6789 {
			t.Errorf("Expected output token limit %d, got %d", 6789, info.OutputTokenLimit)
		}
	})

	t.Run("Network error handling", func(t *testing.T) {
		ctx := context.Background()

		// Simulate network error
		mockClient := createNetworkErrorClient("connection refused")

		client := &geminiClient{
			apiKey:     testAPIKey,
			httpClient: mockClient,
			logger:     getTestLogger(),
		}

		info, err := client.fetchModelInfo(ctx, testModelName)

		// Should return error
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		// Should be an APIError
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Expected APIError, got %T: %v", err, err)
		}

		// Should have network error type
		if apiErr.Type != ErrorTypeNetwork {
			t.Errorf("Expected error type %q, got %q", ErrorTypeNetwork, apiErr.Type)
		}

		// Should contain original error
		if apiErr.Original == nil || !strings.Contains(apiErr.Original.Error(), "connection refused") {
			t.Errorf("Expected original error to contain %q, got %v", "connection refused", apiErr.Original)
		}

		// Result should be nil
		if info != nil {
			t.Errorf("Expected nil ModelInfo, got %+v", info)
		}
	})

	t.Run("HTTP 404 error handling", func(t *testing.T) {
		ctx := context.Background()

		// Create 404 error response
		mockResponse := createErrorResponse(http.StatusNotFound, `{"error": "Model not found"}`)

		client := &geminiClient{
			apiKey:     testAPIKey,
			httpClient: newMockHTTPClient(mockResponse, nil),
			logger:     getTestLogger(),
		}

		info, err := client.fetchModelInfo(ctx, testModelName)

		// Should return error
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		// Should be an APIError
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Expected APIError, got %T: %v", err, err)
		}

		// Should have not found error type
		if apiErr.Type != ErrorTypeNotFound {
			t.Errorf("Expected error type %q, got %q", ErrorTypeNotFound, apiErr.Type)
		}

		// Should have specific message about the model
		if !strings.Contains(apiErr.Message, testModelName) {
			t.Errorf("Expected message to contain model name %q, got %q", testModelName, apiErr.Message)
		}

		// Result should be nil
		if info != nil {
			t.Errorf("Expected nil ModelInfo, got %+v", info)
		}
	})

	t.Run("HTTP 401 error handling", func(t *testing.T) {
		ctx := context.Background()

		// Create 401 error response
		mockResponse := createErrorResponse(http.StatusUnauthorized, `{"error": "Invalid API key"}`)

		client := &geminiClient{
			apiKey:     testAPIKey,
			httpClient: newMockHTTPClient(mockResponse, nil),
			logger:     getTestLogger(),
		}

		info, err := client.fetchModelInfo(ctx, testModelName)

		// Should return error
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		// Should be an APIError
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Expected APIError, got %T: %v", err, err)
		}

		// Should have auth error type
		if apiErr.Type != ErrorTypeAuth {
			t.Errorf("Expected error type %q, got %q", ErrorTypeAuth, apiErr.Type)
		}

		// Result should be nil
		if info != nil {
			t.Errorf("Expected nil ModelInfo, got %+v", info)
		}
	})

	t.Run("HTTP 429 error handling", func(t *testing.T) {
		ctx := context.Background()

		// Create 429 error response
		mockResponse := createErrorResponse(http.StatusTooManyRequests, `{"error": "Rate limit exceeded"}`)

		client := &geminiClient{
			apiKey:     testAPIKey,
			httpClient: newMockHTTPClient(mockResponse, nil),
			logger:     getTestLogger(),
		}

		info, err := client.fetchModelInfo(ctx, testModelName)

		// Should return error
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		// Should be an APIError
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Expected APIError, got %T: %v", err, err)
		}

		// Should have rate limit error type
		if apiErr.Type != ErrorTypeRateLimit {
			t.Errorf("Expected error type %q, got %q", ErrorTypeRateLimit, apiErr.Type)
		}

		// Result should be nil
		if info != nil {
			t.Errorf("Expected nil ModelInfo, got %+v", info)
		}
	})

	t.Run("JSON parsing error handling", func(t *testing.T) {
		ctx := context.Background()

		// Create invalid JSON response
		invalidJSON := `{"name": "models/test-model", "inputTokenLimit": invalid}`
		mockResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(invalidJSON)),
			Header:     make(http.Header),
		}

		client := &geminiClient{
			apiKey:     testAPIKey,
			httpClient: newMockHTTPClient(mockResponse, nil),
			logger:     getTestLogger(),
		}

		info, err := client.fetchModelInfo(ctx, testModelName)

		// Should return error
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		// Should be an APIError
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Expected APIError, got %T: %v", err, err)
		}

		// Should have invalid request error type (since JSON parsing failed)
		if apiErr.Type != ErrorTypeInvalidRequest {
			t.Errorf("Expected error type %q, got %q", ErrorTypeInvalidRequest, apiErr.Type)
		}

		// Result should be nil
		if info != nil {
			t.Errorf("Expected nil ModelInfo, got %+v", info)
		}
	})

	t.Run("HTTP request creation error", func(t *testing.T) {
		// This is hard to simulate directly, but we can check that the fetchModelInfo
		// function properly converts any request creation errors to APIErrors.
		// In a real scenario, this could happen if the context is invalid or if the
		// URL cannot be parsed.

		// For now, we'll just verify that if there's a URL format issue, it's handled.
		ctx := context.Background()

		client := &geminiClient{
			apiKey:      testAPIKey,
			apiEndpoint: "invalid-url-\\with-backslash", // Invalid URL to force error
			httpClient:  &http.Client{},
			logger:      getTestLogger(),
		}

		// The code should not panic
		info, err := client.fetchModelInfo(ctx, testModelName)

		// Should return an error
		if err == nil {
			t.Fatal("Expected error for invalid URL, got nil")
		}

		// Result should be nil
		if info != nil {
			t.Errorf("Expected nil ModelInfo, got %+v", info)
		}
	})
}

func TestGetModelInfo(t *testing.T) {
	// This test verifies that GetModelInfo correctly:
	// - Checks and uses the cache when available
	// - Calls fetchModelInfo when cache misses
	// - Handles and caches errors by returning default values
	// - Properly updates the cache with new values

	// Test constants
	const (
		testAPIKey    = "test-api-key"
		testModelName = "test-model"
	)

	t.Run("First call with empty cache fetches from API", func(t *testing.T) {
		// Create a client with a mock HTTP transport
		mockTransport := &mockHTTPTransport{
			response: createSuccessResponse(ModelDetailsResponse{
				Name:             "models/test-model",
				DisplayName:      "Test Model",
				InputTokenLimit:  32000,
				OutputTokenLimit: 8192,
			}),
		}

		client := &geminiClient{
			apiKey:         testAPIKey,
			modelName:      testModelName,
			logger:         getTestLogger(),
			httpClient:     &http.Client{Transport: mockTransport},
			modelInfoCache: make(map[string]*ModelInfo),
			modelInfoMutex: sync.RWMutex{},
		}

		// Call GetModelInfo
		info, err := client.GetModelInfo(context.Background())

		// Should not return an error
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify result matches expected
		if info == nil {
			t.Fatal("Expected non-nil result, got nil")
		}

		if info.Name != "models/test-model" {
			t.Errorf("Expected model name 'models/test-model', got '%s'", info.Name)
		}

		if info.InputTokenLimit != 32000 {
			t.Errorf("Expected input token limit 32000, got %d", info.InputTokenLimit)
		}

		if info.OutputTokenLimit != 8192 {
			t.Errorf("Expected output token limit 8192, got %d", info.OutputTokenLimit)
		}

		// Verify that the HTTP request was made
		if mockTransport.lastRequest == nil {
			t.Fatal("Expected HTTP request to be made, got nil")
		}
	})

	t.Run("Second call uses cached value", func(t *testing.T) {
		// Create a client with cached value
		cachedInfo := &ModelInfo{
			Name:             "models/cached-model",
			InputTokenLimit:  30000,
			OutputTokenLimit: 7000,
		}

		client := &geminiClient{
			apiKey:    testAPIKey,
			modelName: testModelName,
			logger:    getTestLogger(),
			// Use a transport that would fail if called - to prove cache is used
			httpClient: &http.Client{
				Transport: &mockHTTPTransport{
					err: errors.New("this should not be called"),
				},
			},
			modelInfoCache: map[string]*ModelInfo{
				testModelName: cachedInfo,
			},
			modelInfoMutex: sync.RWMutex{},
		}

		// Call GetModelInfo
		info, err := client.GetModelInfo(context.Background())

		// Should not return an error
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify result matches cached value
		if info != cachedInfo {
			t.Error("Expected cached model info to be returned")
		}
	})

	t.Run("API error returns default values", func(t *testing.T) {
		// Create a client with a mock transport that returns an error
		client := &geminiClient{
			apiKey:         testAPIKey,
			modelName:      testModelName,
			logger:         getTestLogger(),
			httpClient:     createNetworkErrorClient("connection refused"),
			modelInfoCache: make(map[string]*ModelInfo),
			modelInfoMutex: sync.RWMutex{},
		}

		// Call GetModelInfo
		info, err := client.GetModelInfo(context.Background())

		// Should not return an error, even though API call failed
		if err != nil {
			t.Fatalf("Expected no error for API failure (should return defaults), got %v", err)
		}

		// Verify result contains default values
		if info == nil {
			t.Fatal("Expected non-nil result, got nil")
		}

		if info.Name != testModelName {
			t.Errorf("Expected model name '%s', got '%s'", testModelName, info.Name)
		}

		// Default values from the code
		if info.InputTokenLimit != 30720 {
			t.Errorf("Expected default input token limit 30720, got %d", info.InputTokenLimit)
		}

		if info.OutputTokenLimit != 8192 {
			t.Errorf("Expected default output token limit 8192, got %d", info.OutputTokenLimit)
		}
	})

	t.Run("Default values are cached to prevent repeated failures", func(t *testing.T) {
		// Create a client with a mock transport that returns an error
		client := &geminiClient{
			apiKey:    testAPIKey,
			modelName: testModelName,
			logger:    getTestLogger(),
			httpClient: createResponseSequenceClient(
				[]*http.Response{nil},                   // First call - nil response
				[]error{errors.New("first call fails")}, // First call - error
			),
			modelInfoCache: make(map[string]*ModelInfo),
			modelInfoMutex: sync.RWMutex{},
		}

		// First call - should fail API but return defaults
		info1, err := client.GetModelInfo(context.Background())
		if err != nil {
			t.Fatalf("Expected no error (should return defaults), got %v", err)
		}

		// Change the client to one that would panic if called
		// This proves the second call uses the cache
		client.httpClient = &http.Client{
			Transport: &mockHTTPTransport{
				err: errors.New("this should not be called"),
			},
		}

		// Second call - should use cached defaults
		info2, err := client.GetModelInfo(context.Background())
		if err != nil {
			t.Fatalf("Expected no error for second call, got %v", err)
		}

		// Both calls should return the same object (cached)
		if info1 != info2 {
			t.Error("Expected both calls to return the same cached object")
		}
	})

	t.Run("Invalid JSON response", func(t *testing.T) {
		// Create invalid JSON response
		invalidJSON := `{"name": "models/test-model", "inputTokenLimit": invalid}`
		mockResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(invalidJSON)),
			Header:     make(http.Header),
		}

		client := &geminiClient{
			apiKey:         testAPIKey,
			modelName:      testModelName,
			logger:         getTestLogger(),
			httpClient:     newMockHTTPClient(mockResponse, nil),
			modelInfoCache: make(map[string]*ModelInfo),
			modelInfoMutex: sync.RWMutex{},
		}

		// Call GetModelInfo
		info, err := client.GetModelInfo(context.Background())

		// Should not return an error, even with invalid JSON (should return defaults)
		if err != nil {
			t.Fatalf("Expected no error (should return defaults), got %v", err)
		}

		// Verify result contains default values
		if info == nil {
			t.Fatal("Expected non-nil result, got nil")
		}

		// Default values
		if info.InputTokenLimit != 30720 {
			t.Errorf("Expected default input token limit 30720, got %d", info.InputTokenLimit)
		}
	})

	t.Run("Custom endpoint URL construction", func(t *testing.T) {
		customEndpoint := "https://custom-endpoint.example.com"

		// Create a client with a mock transport that verifies the URL
		mockClient := &http.Client{
			Transport: &mockHTTPTransport{
				inspectRequest: func(req *http.Request) {
					expectedURL := fmt.Sprintf("%s/v1beta/models/%s", customEndpoint, testModelName)
					if req.URL.String() != expectedURL {
						t.Errorf("Expected URL %q, got %q", expectedURL, req.URL.String())
					}
				},
				response: createSuccessResponse(ModelDetailsResponse{
					Name:             "models/test-model",
					InputTokenLimit:  32000,
					OutputTokenLimit: 8192,
				}),
			},
		}

		client := &geminiClient{
			apiKey:         testAPIKey,
			modelName:      testModelName,
			apiEndpoint:    customEndpoint,
			logger:         getTestLogger(),
			httpClient:     mockClient,
			modelInfoCache: make(map[string]*ModelInfo),
			modelInfoMutex: sync.RWMutex{},
		}

		// Call GetModelInfo
		_, _ = client.GetModelInfo(context.Background())
		// URL verification is done in the inspectRequest callback
	})
}
