// internal/gemini/mocks_test.go
// Common mock implementations for testing gemini package
//
// NOTE: This file contains some test helpers that are annotated with //nolint:unused.
// These are kept for future test expansion and represent a complete implementation of
// mock interfaces that might be needed for more comprehensive testing in the future.
// Please do not remove these annotated functions without checking whether they might be
// needed for upcoming test scenarios.
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

	"github.com/phrazzld/architect/internal/logutil"

	genai "github.com/google/generative-ai-go/genai"
)

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
// Kept for future test expansion
//
//nolint:unused // Kept for future expansion of HTTP error testing
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
// Kept for future test scenarios requiring URL-specific responses
//
//nolint:unused
type urlPatternTransport struct {
	urlToStatus    map[string]int
	urlToBody      map[string]string
	lastRequest    *http.Request
	inspectRequest func(*http.Request)
}

//nolint:unused
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
// Kept for future test expansion
//
//nolint:unused
func createStatusCodeMap(urlToStatus map[string]int, urlToBody map[string]string) *http.Client {
	transport := &urlPatternTransport{
		urlToStatus: urlToStatus,
		urlToBody:   urlToBody,
	}

	return &http.Client{Transport: transport}
}

// mockGenerativeModel is a test implementation of the genai model
// Kept for future test scenarios requiring direct model mocking
//
//nolint:unused
type mockGenerativeModel struct {
	generateResp *genai.GenerateContentResponse
	generateErr  error
	countResp    *genai.CountTokensResponse
	countErr     error

	// Capture calls for verification
	lastPrompt string
}

// GenerateContent implements the GenerativeModel interface for testing
//
//nolint:unused
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
//
//nolint:unused
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
//
//nolint:unused
func (m *mockGenerativeModel) SetTemperature(t float32) { /* No-op for testing */ }

// SetTopP implements the GenerativeModel interface for testing
//
//nolint:unused
func (m *mockGenerativeModel) SetTopP(p float32) { /* No-op for testing */ }

// SetTopK implements the GenerativeModel interface for testing
//
//nolint:unused
func (m *mockGenerativeModel) SetTopK(k int32) { /* No-op for testing */ }

// SetMaxOutputTokens implements the GenerativeModel interface for testing
//
//nolint:unused
func (m *mockGenerativeModel) SetMaxOutputTokens(tokens int32) { /* No-op for testing */ }

// Temperature implements the GenerativeModel interface for testing
//
//nolint:unused
func (m *mockGenerativeModel) Temperature() *float32 {
	t := float32(0.7)
	return &t
}

// TopP implements the GenerativeModel interface for testing
//
//nolint:unused
func (m *mockGenerativeModel) TopP() *float32 {
	p := float32(0.95)
	return &p
}

// TopK implements the GenerativeModel interface for testing
//
//nolint:unused
func (m *mockGenerativeModel) TopK() *int32 {
	k := int32(40)
	return &k
}

// MaxOutputTokens implements the GenerativeModel interface for testing
//
//nolint:unused
func (m *mockGenerativeModel) MaxOutputTokens() *int32 {
	tokens := int32(2048)
	return &tokens
}
