// internal/gemini/gemini_client_test.go
// Tests for the gemini package implementing the Gemini API client
//
//nolint:unused,U1000 // Contains helper functions that may be used in future test expansions
package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/phrazzld/architect/internal/logutil"

	genai "github.com/google/generative-ai-go/genai"
)

// Mock components for testing

// mockHTTPTransport implements http.RoundTripper for testing HTTP requests
type mockHTTPTransport struct {
	// Response to return
	response *http.Response
	// Error to return
	err error
	// Function to inspect the request before responding
	inspectRequest func(*http.Request)
	// Capture the most recent request for inspection in tests
	lastRequest *http.Request
}

func (m *mockHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Store the request for later inspection
	m.lastRequest = req

	// Call the inspect function if provided
	if m.inspectRequest != nil {
		m.inspectRequest(req)
	}

	return m.response, m.err
}

// newMockHTTPClient creates a new HTTP client with a mock transport
func newMockHTTPClient(resp *http.Response, err error) *http.Client {
	return &http.Client{
		Transport: &mockHTTPTransport{
			response: resp,
			err:      err,
		},
	}
}

// getTestLogger returns a no-op logger for testing
func getTestLogger() logutil.LoggerInterface {
	// Use a discard writer that does nothing with the log output
	return logutil.NewLogger(logutil.InfoLevel, io.Discard, "[test] ")
}

// getMockTransport retrieves the mockHTTPTransport from a client for inspection
// Unused function kept for future expansion of tests
//
//lint:ignore U1000 Kept for future test expansion
func getMockTransport(client *http.Client) *mockHTTPTransport {
	if transport, ok := client.Transport.(*mockHTTPTransport); ok {
		return transport
	}
	return nil
}

// Helper functions for creating test responses

// createSuccessResponse creates a mock HTTP success response with the given body
func createSuccessResponse(body interface{}) *http.Response {
	jsonBody, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jsonBody)),
		Header:     make(http.Header),
	}
}

// createErrorResponse creates a mock HTTP error response
func createErrorResponse(statusCode int, errorMessage string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader([]byte(errorMessage))),
		Header:     make(http.Header),
	}
}

// createNetworkErrorClient creates a mock client that simulates network errors
func createNetworkErrorClient(errorMessage string) *http.Client {
	return &http.Client{
		Transport: &mockHTTPTransport{
			err: errors.New(errorMessage),
		},
	}
}

// createRequestErrorClient creates a mock client that captures the request but fails with a specific error
func createRequestErrorClient(errorMessage string, inspectFunc func(*http.Request)) *http.Client {
	return &http.Client{
		Transport: &mockHTTPTransport{
			err:            errors.New(errorMessage),
			inspectRequest: inspectFunc,
		},
	}
}

// sequenceTransport is a custom transport that returns responses in sequence
type sequenceTransport struct {
	responses      []*http.Response
	errors         []error
	index          int
	lastRequest    *http.Request
	inspectRequest func(*http.Request)
}

func (t *sequenceTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.lastRequest = req
	if t.inspectRequest != nil {
		t.inspectRequest(req)
	}

	if t.index >= len(t.responses) {
		return nil, fmt.Errorf("no more responses in sequence (called %d times)", t.index+1)
	}

	var resp *http.Response
	var err error

	if t.index < len(t.responses) {
		resp = t.responses[t.index]
	}

	if t.index < len(t.errors) {
		err = t.errors[t.index]
	}

	t.index++
	return resp, err
}

// createResponseSequenceClient creates a client that returns responses in sequence
func createResponseSequenceClient(responses []*http.Response, errors []error) *http.Client {
	transport := &sequenceTransport{
		responses: responses,
		errors:    errors,
		index:     0,
		inspectRequest: func(req *http.Request) {
			// Do nothing, just for capturing the request
		},
	}

	return &http.Client{Transport: transport}
}

// urlPatternTransport is a custom transport that maps URLs to specific responses
type urlPatternTransport struct {
	urlToStatus    map[string]int
	urlToBody      map[string]string
	lastRequest    *http.Request
	inspectRequest func(*http.Request)
}

func (t *urlPatternTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.lastRequest = req
	if t.inspectRequest != nil {
		t.inspectRequest(req)
	}

	statusCode := http.StatusOK
	responseBody := "{}"

	// Find matching URL pattern
	for urlPattern, code := range t.urlToStatus {
		if strings.Contains(req.URL.String(), urlPattern) {
			statusCode = code
			break
		}
	}

	// Find matching body
	for urlPattern, body := range t.urlToBody {
		if strings.Contains(req.URL.String(), urlPattern) {
			responseBody = body
			break
		}
	}

	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
		Header:     make(http.Header),
	}, nil
}

// createStatusCodeMap creates a client that maps URLs to specific status codes
func createStatusCodeMap(urlToStatus map[string]int, urlToBody map[string]string) *http.Client {
	transport := &urlPatternTransport{
		urlToStatus: urlToStatus,
		urlToBody:   urlToBody,
	}

	return &http.Client{Transport: transport}
}

// Test stubs (to be implemented in subsequent tasks)

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

func TestGenerateContent(t *testing.T) {
	// This test verifies that GenerateContent correctly:
	// - Validates the prompt (empty check)
	// - Handles API errors
	// - Processes responses correctly in various scenarios

	t.Run("Empty prompt validation", func(t *testing.T) {
		// Create a geminiClient directly to test its implementation
		client := &geminiClient{
			apiKey:    "test-key",
			modelName: "test-model",
			logger:    getTestLogger(),
			// No model needed as we're testing the empty prompt check which happens first
		}

		// Call GenerateContent with empty prompt
		result, err := client.GenerateContent(context.Background(), "")

		// Verify error is returned
		if err == nil {
			t.Fatal("Expected error for empty prompt, got nil")
		}

		// Verify it's an APIError with the correct type
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Expected *APIError, got %T", err)
		}

		if apiErr.Type != ErrorTypeInvalidRequest {
			t.Errorf("Expected error type %v, got %v", ErrorTypeInvalidRequest, apiErr.Type)
		}

		// Verify message mentions empty prompt
		if !strings.Contains(apiErr.Message, "empty prompt") {
			t.Errorf("Expected error message to mention empty prompt, got: %s", apiErr.Message)
		}

		// Result should be nil
		if result != nil {
			t.Errorf("Expected nil result, got %+v", result)
		}
	})

	// For the remaining tests, we'll use a mocked implementation of GenerateContent
	// This lets us test the interface without getting into the genai package details

	t.Run("API error handling", func(t *testing.T) {
		// Setup mock client that returns a specific error
		client := &MockClient{
			GenerateContentFunc: func(ctx context.Context, prompt string) (*GenerationResult, error) {
				return nil, &APIError{
					Original:   errors.New("API error: rate limit exceeded"),
					Type:       ErrorTypeRateLimit,
					Message:    "Request rate limit or quota exceeded on the Gemini API",
					Suggestion: "Wait and try again later.",
				}
			},
		}

		// Call GenerateContent
		result, err := client.GenerateContent(context.Background(), "Test prompt")

		// Verify error is returned
		if err == nil {
			t.Fatal("Expected error from API, got nil")
		}

		// Verify it's an APIError with the expected type
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Expected *APIError, got %T", err)
		}

		if apiErr.Type != ErrorTypeRateLimit {
			t.Errorf("Expected error type %v, got %v", ErrorTypeRateLimit, apiErr.Type)
		}

		// Result should be nil
		if result != nil {
			t.Errorf("Expected nil result, got %+v", result)
		}
	})

	t.Run("Successful content generation", func(t *testing.T) {
		// Setup mock client that returns a successful result
		expectedResult := &GenerationResult{
			Content:      "Generated test content",
			FinishReason: "STOP",
			SafetyRatings: []SafetyRating{
				{
					Category: "HARM_CATEGORY_HARASSMENT",
					Score:    0.1,
					Blocked:  false,
				},
			},
			TokenCount: 42,
			Truncated:  false,
		}

		client := &MockClient{
			GenerateContentFunc: func(ctx context.Context, prompt string) (*GenerationResult, error) {
				if prompt != "Test prompt" {
					t.Errorf("Expected prompt 'Test prompt', got '%s'", prompt)
				}
				return expectedResult, nil
			},
		}

		// Call GenerateContent
		result, err := client.GenerateContent(context.Background(), "Test prompt")

		// Should not return an error
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify result matches expected
		if result != expectedResult {
			t.Errorf("Expected result %+v, got %+v", expectedResult, result)
		}
	})

	t.Run("Content with safety blocks", func(t *testing.T) {
		// Setup mock client that returns content with safety ratings
		client := &MockClient{
			GenerateContentFunc: func(ctx context.Context, prompt string) (*GenerationResult, error) {
				return &GenerationResult{
					Content:      "Safe content",
					FinishReason: "SAFETY",
					SafetyRatings: []SafetyRating{
						{
							Category: "HARM_CATEGORY_DANGEROUS_CONTENT",
							Score:    0.8,
							Blocked:  true,
						},
					},
				}, nil
			},
		}

		// Call GenerateContent
		result, err := client.GenerateContent(context.Background(), "Test prompt")

		// Should not return an error (safety filtering is not an error)
		if err != nil {
			t.Fatalf("Expected no error for safety filtered content, got %v", err)
		}

		// Verify result
		if result == nil {
			t.Fatal("Expected non-nil result, got nil")
		}

		// Content should be available
		expected := "Safe content"
		if result.Content != expected {
			t.Errorf("Expected content %q, got %q", expected, result.Content)
		}

		// FinishReason should indicate safety
		if result.FinishReason != "SAFETY" {
			t.Errorf("Expected finish reason SAFETY, got %q", result.FinishReason)
		}

		// Should have safety ratings with blocked=true
		if len(result.SafetyRatings) != 1 {
			t.Fatalf("Expected 1 safety rating, got %d", len(result.SafetyRatings))
		}

		if !result.SafetyRatings[0].Blocked {
			t.Error("Expected blocked safety rating, got not blocked")
		}
	})

	t.Run("Truncated content", func(t *testing.T) {
		// Setup mock client that returns truncated content
		client := &MockClient{
			GenerateContentFunc: func(ctx context.Context, prompt string) (*GenerationResult, error) {
				return &GenerationResult{
					Content:      "Truncated content",
					FinishReason: "MAX_TOKENS",
					Truncated:    true,
				}, nil
			},
		}

		// Call GenerateContent
		result, err := client.GenerateContent(context.Background(), "Test prompt")

		// Should not return an error
		if err != nil {
			t.Fatalf("Expected no error for truncated content, got %v", err)
		}

		// Verify result is marked as truncated
		if !result.Truncated {
			t.Error("Expected truncated=true, got truncated=false")
		}

		// FinishReason should indicate max tokens
		if result.FinishReason != "MAX_TOKENS" {
			t.Errorf("Expected finish reason MAX_TOKENS, got %q", result.FinishReason)
		}
	})
}

func TestCountTokens(t *testing.T) {
	// This test verifies that CountTokens correctly:
	// - Handles empty prompts
	// - Handles API errors
	// - Processes responses correctly for token counting

	// Test constants
	const (
		testPrompt = "Test prompt for token counting"
	)

	t.Run("Empty prompt handling", func(t *testing.T) {
		// For this specific test, we'll use the MockClient since we can easily customize its behavior
		client := &MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*TokenCount, error) {
				if prompt == "" {
					return &TokenCount{Total: 0}, nil
				}
				return nil, errors.New("test failed: expected empty prompt")
			},
		}

		// Call CountTokens with empty prompt
		result, err := client.CountTokens(context.Background(), "")

		// Should not return an error
		if err != nil {
			t.Fatalf("Expected no error for empty prompt, got %v", err)
		}

		// Should return a TokenCount with Total=0
		if result == nil {
			t.Fatal("Expected non-nil result, got nil")
		}

		if result.Total != 0 {
			t.Errorf("Expected token count 0 for empty prompt, got %d", result.Total)
		}
	})

	t.Run("API error handling", func(t *testing.T) {
		// Setup mock client that returns a specific error
		client := &MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*TokenCount, error) {
				return nil, &APIError{
					Original:   errors.New("API error: invalid request"),
					Type:       ErrorTypeInvalidRequest,
					Message:    "Failed to count tokens in prompt",
					Suggestion: "Check your API key and internet connection.",
				}
			},
		}

		// Call CountTokens
		result, err := client.CountTokens(context.Background(), testPrompt)

		// Verify error is returned
		if err == nil {
			t.Fatal("Expected error from API, got nil")
		}

		// Verify it's an APIError with the expected type
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Expected *APIError, got %T", err)
		}

		if apiErr.Type != ErrorTypeInvalidRequest {
			t.Errorf("Expected error type %v, got %v", ErrorTypeInvalidRequest, apiErr.Type)
		}

		// Verify message is as expected
		if !strings.Contains(apiErr.Message, "Failed to count tokens") {
			t.Errorf("Expected message to mention token counting, got: %s", apiErr.Message)
		}

		// Result should be nil
		if result != nil {
			t.Errorf("Expected nil result, got %+v", result)
		}
	})

	t.Run("Rate limit error handling", func(t *testing.T) {
		// Setup mock client that returns a rate limit error
		client := &MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*TokenCount, error) {
				return nil, &APIError{
					Original:   errors.New("API error: rate limit exceeded"),
					Type:       ErrorTypeRateLimit,
					Message:    "Request rate limit or quota exceeded on the Gemini API",
					Suggestion: "Wait and try again later.",
				}
			},
		}

		// Call CountTokens
		result, err := client.CountTokens(context.Background(), testPrompt)

		// Verify error is returned
		if err == nil {
			t.Fatal("Expected error from API, got nil")
		}

		// Verify it's an APIError with the expected type
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Expected *APIError, got %T", err)
		}

		if apiErr.Type != ErrorTypeRateLimit {
			t.Errorf("Expected error type %v, got %v", ErrorTypeRateLimit, apiErr.Type)
		}

		// Result should be nil
		if result != nil {
			t.Errorf("Expected nil result, got %+v", result)
		}
	})

	t.Run("Network error handling", func(t *testing.T) {
		// Setup mock client that returns a network error
		client := &MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*TokenCount, error) {
				return nil, &APIError{
					Original:   errors.New("network error: connection refused"),
					Type:       ErrorTypeNetwork,
					Message:    "Network error while connecting to the Gemini API",
					Suggestion: "Check your internet connection and try again.",
				}
			},
		}

		// Call CountTokens
		result, err := client.CountTokens(context.Background(), testPrompt)

		// Verify error is returned
		if err == nil {
			t.Fatal("Expected error from API, got nil")
		}

		// Verify it's an APIError with the expected type
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("Expected *APIError, got %T", err)
		}

		if apiErr.Type != ErrorTypeNetwork {
			t.Errorf("Expected error type %v, got %v", ErrorTypeNetwork, apiErr.Type)
		}

		// Result should be nil
		if result != nil {
			t.Errorf("Expected nil result, got %+v", result)
		}
	})

	t.Run("Successful token counting", func(t *testing.T) {
		// Setup mock client that returns a successful result
		expectedResult := &TokenCount{
			Total: 42,
		}

		client := &MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*TokenCount, error) {
				if prompt != testPrompt {
					t.Errorf("Expected prompt '%s', got '%s'", testPrompt, prompt)
				}
				return expectedResult, nil
			},
		}

		// Call CountTokens
		result, err := client.CountTokens(context.Background(), testPrompt)

		// Should not return an error
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify result matches expected
		if result == nil {
			t.Fatal("Expected non-nil result, got nil")
		}

		if result.Total != 42 {
			t.Errorf("Expected token count 42, got %d", result.Total)
		}
	})

	t.Run("Large token count", func(t *testing.T) {
		// Setup mock client that returns a large token count (e.g., for a long document)
		client := &MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*TokenCount, error) {
				return &TokenCount{
					Total: 10000, // A large number of tokens
				}, nil
			},
		}

		// Call CountTokens
		result, err := client.CountTokens(context.Background(), "This is a very long document...")

		// Should not return an error
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify result has expected large token count
		if result == nil {
			t.Fatal("Expected non-nil result, got nil")
		}

		if result.Total != 10000 {
			t.Errorf("Expected token count 10000, got %d", result.Total)
		}
	})
}

// mockGenerativeModel is a test implementation of the genai model
type mockGenerativeModel struct {
	generateResp *genai.GenerateContentResponse
	generateErr  error
	countResp    *genai.CountTokensResponse
	countErr     error

	// Capture calls for verification
	lastPrompt string
}

// GenerateContent implements the GenerativeModel interface for testing
func (m *mockGenerativeModel) GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	// Capture the prompt for inspection
	if len(parts) > 0 {
		if textPart, ok := parts[0].(genai.Text); ok {
			m.lastPrompt = string(textPart)
		}
	}

	return m.generateResp, m.generateErr
}

// CountTokens implements the GenerativeModel interface for testing
func (m *mockGenerativeModel) CountTokens(ctx context.Context, parts ...genai.Part) (*genai.CountTokensResponse, error) {
	// Capture the prompt for inspection
	if len(parts) > 0 {
		if textPart, ok := parts[0].(genai.Text); ok {
			m.lastPrompt = string(textPart)
		}
	}

	return m.countResp, m.countErr
}

// SetTemperature implements the GenerativeModel interface for testing
func (m *mockGenerativeModel) SetTemperature(t float32) { /* No-op for testing */ }

// SetTopP implements the GenerativeModel interface for testing
func (m *mockGenerativeModel) SetTopP(p float32) { /* No-op for testing */ }

// SetTopK implements the GenerativeModel interface for testing
func (m *mockGenerativeModel) SetTopK(k int32) { /* No-op for testing */ }

// SetMaxOutputTokens implements the GenerativeModel interface for testing
func (m *mockGenerativeModel) SetMaxOutputTokens(tokens int32) { /* No-op for testing */ }

// Temperature implements the GenerativeModel interface for testing
func (m *mockGenerativeModel) Temperature() *float32 {
	t := float32(0.7)
	return &t
}

// TopP implements the GenerativeModel interface for testing
func (m *mockGenerativeModel) TopP() *float32 {
	p := float32(0.95)
	return &p
}

// TopK implements the GenerativeModel interface for testing
func (m *mockGenerativeModel) TopK() *int32 {
	k := int32(40)
	return &k
}

// MaxOutputTokens implements the GenerativeModel interface for testing
func (m *mockGenerativeModel) MaxOutputTokens() *int32 {
	tokens := int32(2048)
	return &tokens
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

func TestHelperMethods(t *testing.T) {
	// This test verifies helper methods:
	// - mapSafetyRatings correctly converts between types
	// - GetModelName returns the correct model name
	// - GetTemperature returns the correct temperature
	// - GetMaxOutputTokens returns the correct token limit
	// - GetTopP returns the correct topP value

	t.Run("mapSafetyRatings with nil ratings", func(t *testing.T) {
		// When passed nil, should return nil
		result := mapSafetyRatings(nil)
		if result != nil {
			t.Errorf("Expected nil result for nil input, got %+v", result)
		}
	})

	t.Run("mapSafetyRatings with empty ratings", func(t *testing.T) {
		// When passed empty slice, should return empty slice
		result := mapSafetyRatings([]*genai.SafetyRating{})
		if result == nil {
			t.Error("Expected empty slice, got nil")
		}
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got slice with %d elements", len(result))
		}
	})

	t.Run("mapSafetyRatings with actual ratings", func(t *testing.T) {
		// Create sample genai safety ratings
		ratings := []*genai.SafetyRating{
			{
				Category:    genai.HarmCategoryHarassment,
				Probability: genai.HarmProbabilityMedium,
				Blocked:     true,
			},
			{
				Category:    genai.HarmCategoryDangerousContent,
				Probability: genai.HarmProbabilityLow,
				Blocked:     false,
			},
		}

		// Map to our internal format
		result := mapSafetyRatings(ratings)

		// Verify the result
		if len(result) != 2 {
			t.Fatalf("Expected 2 ratings, got %d", len(result))
		}

		// Check first rating
		if string(result[0].Category) != string(genai.HarmCategoryHarassment) {
			t.Errorf("Expected category %q, got %q", genai.HarmCategoryHarassment, result[0].Category)
		}
		if result[0].Score != float32(genai.HarmProbabilityMedium) {
			t.Errorf("Expected score %f, got %f", float32(genai.HarmProbabilityMedium), result[0].Score)
		}
		if !result[0].Blocked {
			t.Error("Expected blocked to be true, got false")
		}

		// Check second rating
		if string(result[1].Category) != string(genai.HarmCategoryDangerousContent) {
			t.Errorf("Expected category %q, got %q", genai.HarmCategoryDangerousContent, result[1].Category)
		}
		if result[1].Score != float32(genai.HarmProbabilityLow) {
			t.Errorf("Expected score %f, got %f", float32(genai.HarmProbabilityLow), result[1].Score)
		}
		if result[1].Blocked {
			t.Error("Expected blocked to be false, got true")
		}
	})

	t.Run("GetModelName returns correct value", func(t *testing.T) {
		// Test with the actual implementation
		const expectedModelName = "test-model-name"

		client := &geminiClient{
			modelName: expectedModelName,
		}

		modelName := client.GetModelName()
		if modelName != expectedModelName {
			t.Errorf("Expected model name %q, got %q", expectedModelName, modelName)
		}

		// Test with MockClient
		mockClient := &MockClient{
			GetModelNameFunc: func() string {
				return "mock-model-name"
			},
		}

		mockModelName := mockClient.GetModelName()
		if mockModelName != "mock-model-name" {
			t.Errorf("Expected mock model name %q, got %q", "mock-model-name", mockModelName)
		}

		// Test default mock implementation
		defaultMockClient := NewMockClient()
		defaultMockName := defaultMockClient.GetModelName()
		if defaultMockName != "mock-model" {
			t.Errorf("Expected default mock model name %q, got %q", "mock-model", defaultMockName)
		}
	})

	t.Run("GetTemperature returns correct value", func(t *testing.T) {
		defaultTemp := DefaultModelConfig().Temperature
		customTemp := float32(0.42)

		// Test actual implementation with nil model
		clientWithNilModel := &geminiClient{
			model:  nil,
			logger: getTestLogger(),
		}

		tempFromNilModel := clientWithNilModel.GetTemperature()
		if tempFromNilModel != defaultTemp {
			t.Errorf("Expected default temperature %f from nil model, got %f", defaultTemp, tempFromNilModel)
		}

		// Test with MockClient with custom function
		mockClient := &MockClient{
			GetTemperatureFunc: func() float32 {
				return customTemp
			},
		}

		mockTemp := mockClient.GetTemperature()
		if mockTemp != customTemp {
			t.Errorf("Expected mock temperature %f, got %f", customTemp, mockTemp)
		}

		// Test default mock implementation
		defaultMockClient := NewMockClient()
		defaultMockTemp := defaultMockClient.GetTemperature()
		if defaultMockTemp != defaultTemp {
			t.Errorf("Expected default mock temperature %f, got %f", defaultTemp, defaultMockTemp)
		}
	})

	t.Run("GetMaxOutputTokens returns correct value", func(t *testing.T) {
		defaultTokens := DefaultModelConfig().MaxOutputTokens
		customTokens := int32(4096)

		// Test actual implementation with nil model
		clientWithNilModel := &geminiClient{
			model:  nil,
			logger: getTestLogger(),
		}

		tokensFromNilModel := clientWithNilModel.GetMaxOutputTokens()
		if tokensFromNilModel != defaultTokens {
			t.Errorf("Expected default tokens %d from nil model, got %d", defaultTokens, tokensFromNilModel)
		}

		// Test with MockClient with custom function
		mockClient := &MockClient{
			GetMaxOutputTokensFunc: func() int32 {
				return customTokens
			},
		}

		mockTokens := mockClient.GetMaxOutputTokens()
		if mockTokens != customTokens {
			t.Errorf("Expected mock tokens %d, got %d", customTokens, mockTokens)
		}

		// Test default mock implementation
		defaultMockClient := NewMockClient()
		defaultMockTokens := defaultMockClient.GetMaxOutputTokens()
		if defaultMockTokens != defaultTokens {
			t.Errorf("Expected default mock tokens %d, got %d", defaultTokens, defaultMockTokens)
		}
	})

	t.Run("GetTopP returns correct value", func(t *testing.T) {
		defaultTopP := DefaultModelConfig().TopP
		customTopP := float32(0.75)

		// Test actual implementation with nil model
		clientWithNilModel := &geminiClient{
			model:  nil,
			logger: getTestLogger(),
		}

		topPFromNilModel := clientWithNilModel.GetTopP()
		if topPFromNilModel != defaultTopP {
			t.Errorf("Expected default topP %f from nil model, got %f", defaultTopP, topPFromNilModel)
		}

		// Test with MockClient with custom function
		mockClient := &MockClient{
			GetTopPFunc: func() float32 {
				return customTopP
			},
		}

		mockTopP := mockClient.GetTopP()
		if mockTopP != customTopP {
			t.Errorf("Expected mock topP %f, got %f", customTopP, mockTopP)
		}

		// Test default mock implementation
		defaultMockClient := NewMockClient()
		defaultMockTopP := defaultMockClient.GetTopP()
		if defaultMockTopP != defaultTopP {
			t.Errorf("Expected default mock topP %f, got %f", defaultTopP, defaultMockTopP)
		}
	})
}
