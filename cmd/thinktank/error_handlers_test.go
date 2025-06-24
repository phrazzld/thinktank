package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
)

// TestCreateStandardErrorWithMessage tests the creation of user-friendly error messages based on category
func TestCreateStandardErrorWithMessage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		provider    string
		category    llm.ErrorCategory
		originalErr error
		details     string
		expectMsg   string
		expectSugg  string
	}{
		{
			name:        "Auth error message",
			provider:    "test-provider",
			category:    llm.CategoryAuth,
			originalErr: errors.New("auth failed"),
			details:     "",
			expectMsg:   "Authentication failed with the test-provider API",
			expectSugg:  "Check that your API key is valid",
		},
		{
			name:        "Rate limit error message",
			provider:    "test-provider",
			category:    llm.CategoryRateLimit,
			originalErr: errors.New("too many requests"),
			details:     "",
			expectMsg:   "Request rate limit exceeded on the test-provider API",
			expectSugg:  "Wait and try again later",
		},
		{
			name:        "Server error message with details",
			provider:    "test-provider",
			category:    llm.CategoryServer,
			originalErr: errors.New("internal error"),
			details:     "500 Internal Server Error",
			expectMsg:   "test-provider API server error occurred (500 Internal Server Error)",
			expectSugg:  "This is typically a temporary issue",
		},
		{
			name:        "Network error message",
			provider:    "test-provider",
			category:    llm.CategoryNetwork,
			originalErr: errors.New("connection failed"),
			details:     "",
			expectMsg:   "Network error while connecting to the test-provider API",
			expectSugg:  "Check your internet connection",
		},
		{
			name:        "Input limit error message",
			provider:    "test-provider",
			category:    llm.CategoryInputLimit,
			originalErr: errors.New("too many tokens"),
			details:     "",
			expectMsg:   "Input token limit exceeded for the selected model",
			expectSugg:  "Reduce the input size",
		},
		{
			name:        "Content filtered error message",
			provider:    "test-provider",
			category:    llm.CategoryContentFiltered,
			originalErr: errors.New("content blocked"),
			details:     "",
			expectMsg:   "Content was filtered by safety settings",
			expectSugg:  "Your prompt or content may have triggered safety filters",
		},
		{
			name:        "Unknown error message",
			provider:    "test-provider",
			category:    llm.CategoryUnknown,
			originalErr: errors.New("something weird happened"),
			details:     "",
			expectMsg:   "Error calling test-provider API",
			expectSugg:  "Check the logs for more details",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llmErr := llm.CreateStandardErrorWithMessage(tt.provider, tt.category, tt.originalErr, tt.details)

			if !strings.Contains(llmErr.Message, tt.expectMsg) {
				t.Errorf("Expected message to contain %q, got %q", tt.expectMsg, llmErr.Message)
			}

			if !strings.Contains(llmErr.Suggestion, tt.expectSugg) {
				t.Errorf("Expected suggestion to contain %q, got %q", tt.expectSugg, llmErr.Suggestion)
			}

			// Verify original error is wrapped
			if !errors.Is(llmErr, tt.originalErr) {
				t.Errorf("Expected original error to be wrapped, but it was not")
			}
		})
	}
}

// TestFormatAPIError tests the creation of LLMError from generic errors
func TestFormatAPIError(t *testing.T) {
	tests := []struct {
		name           string
		provider       string
		err            error
		statusCode     int
		responseBody   string
		expectCategory llm.ErrorCategory
	}{
		{
			name:           "Auth error from status code",
			provider:       "test-provider",
			err:            errors.New("unauthorized"),
			statusCode:     401,
			responseBody:   "{\"error\":\"invalid_auth\"}",
			expectCategory: llm.CategoryAuth,
		},
		{
			name:           "Rate limit error from status code",
			provider:       "test-provider",
			err:            errors.New("too many requests"),
			statusCode:     429,
			responseBody:   "{\"error\":\"rate_limited\"}",
			expectCategory: llm.CategoryRateLimit,
		},
		{
			name:           "Server error from status code",
			provider:       "test-provider",
			err:            errors.New("server error"),
			statusCode:     500,
			responseBody:   "{\"error\":\"internal_error\"}",
			expectCategory: llm.CategoryServer,
		},
		{
			name:           "Auth error from message",
			provider:       "test-provider",
			err:            errors.New("invalid API key provided"),
			statusCode:     0,
			responseBody:   "",
			expectCategory: llm.CategoryAuth,
		},
		{
			name:           "Input limit error from message",
			provider:       "test-provider",
			err:            errors.New("token limit exceeded for prompt"),
			statusCode:     0,
			responseBody:   "",
			expectCategory: llm.CategoryInputLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llmErr := llm.FormatAPIError(tt.provider, tt.err, tt.statusCode, tt.responseBody)

			if llmErr.ErrorCategory != tt.expectCategory {
				t.Errorf("Expected category %v, got %v", tt.expectCategory, llmErr.ErrorCategory)
			}

			if llmErr.Provider != tt.provider {
				t.Errorf("Expected provider %q, got %q", tt.provider, llmErr.Provider)
			}

			// Check response body in details
			if tt.responseBody != "" && !strings.Contains(llmErr.Details, tt.responseBody) {
				t.Errorf("Expected details to contain response body %q", tt.responseBody)
			}

			// Verify original error is wrapped
			if !errors.Is(llmErr, tt.err) {
				t.Errorf("Expected original error to be wrapped, but it was not")
			}
		})
	}
}

// TestIsErrorCategoryHelpers tests the various IsXXX helper functions for error categories
func TestIsErrorCategoryHelpers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		err      error
		category llm.ErrorCategory
		testFunc func(error) bool
	}{
		{"IsAuth", llm.New("test", "", 0, "", "", nil, llm.CategoryAuth), llm.CategoryAuth, llm.IsAuth},
		{"IsRateLimit", llm.New("test", "", 0, "", "", nil, llm.CategoryRateLimit), llm.CategoryRateLimit, llm.IsRateLimit},
		{"IsInvalidRequest", llm.New("test", "", 0, "", "", nil, llm.CategoryInvalidRequest), llm.CategoryInvalidRequest, llm.IsInvalidRequest},
		{"IsNotFound", llm.New("test", "", 0, "", "", nil, llm.CategoryNotFound), llm.CategoryNotFound, llm.IsNotFound},
		{"IsServer", llm.New("test", "", 0, "", "", nil, llm.CategoryServer), llm.CategoryServer, llm.IsServer},
		{"IsNetwork", llm.New("test", "", 0, "", "", nil, llm.CategoryNetwork), llm.CategoryNetwork, llm.IsNetwork},
		{"IsCancelled", llm.New("test", "", 0, "", "", nil, llm.CategoryCancelled), llm.CategoryCancelled, llm.IsCancelled},
		{"IsInputLimit", llm.New("test", "", 0, "", "", nil, llm.CategoryInputLimit), llm.CategoryInputLimit, llm.IsInputLimit},
		{"IsContentFiltered", llm.New("test", "", 0, "", "", nil, llm.CategoryContentFiltered), llm.CategoryContentFiltered, llm.IsContentFiltered},
		{"IsInsufficientCredits", llm.New("test", "", 0, "", "", nil, llm.CategoryInsufficientCredits), llm.CategoryInsufficientCredits, llm.IsInsufficientCredits},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test positive case
			if !tt.testFunc(tt.err) {
				t.Errorf("Expected %s(%v) to return true", tt.name, tt.err)
			}

			// Test negative case with different category
			otherCategory := (tt.category + 1) % 11 // Wrap around to ensure different category
			if otherCategory == tt.category {
				otherCategory = (otherCategory + 1) % 11
			}
			otherErr := llm.New("test", "", 0, "", "", nil, otherCategory)
			if tt.testFunc(otherErr) {
				t.Errorf("Expected %s(%v) to return false for category %v", tt.name, otherErr, otherCategory)
			}

			// Test with nil error
			if tt.testFunc(nil) {
				t.Errorf("Expected %s(nil) to return false", tt.name)
			}

			// Test with regular error
			if tt.testFunc(errors.New("regular error")) {
				t.Errorf("Expected %s(regular error) to return false", tt.name)
			}
		})
	}
}

// TestIsCategorizedError tests the IsCategorizedError helper function
func TestIsCategorizedError(t *testing.T) {
	t.Parallel(
	// Test with nil error
	)

	if catErr, ok := llm.IsCategorizedError(nil); ok || catErr != nil {
		t.Errorf("Expected IsCategorizedError(nil) to return (nil, false), got (%v, %v)", catErr, ok)
	}

	// Test with standard error
	stdErr := errors.New("standard error")
	if catErr, ok := llm.IsCategorizedError(stdErr); ok || catErr != nil {
		t.Errorf("Expected IsCategorizedError(stdErr) to return (nil, false), got (%v, %v)", catErr, ok)
	}

	// Test with LLMError
	llmErr := llm.New("test", "", 0, "test error", "", nil, llm.CategoryAuth)
	catErr, ok := llm.IsCategorizedError(llmErr)
	if !ok || catErr == nil {
		t.Errorf("Expected IsCategorizedError(llmErr) to return (non-nil, true), got (%v, %v)", catErr, ok)
	}
	if ok && catErr.Category() != llm.CategoryAuth {
		t.Errorf("Expected category Auth, got %v", catErr.Category())
	}

	// Test with wrapped LLMError
	wrappedErr := fmt.Errorf("wrapped: %w", llmErr)
	catErr, ok = llm.IsCategorizedError(wrappedErr)
	if !ok || catErr == nil {
		t.Errorf("Expected IsCategorizedError(wrappedErr) to return (non-nil, true), got (%v, %v)", catErr, ok)
	}
	if ok && catErr.Category() != llm.CategoryAuth {
		t.Errorf("Expected category Auth, got %v", catErr.Category())
	}

	// Test with MockError
	mockErr := &llm.MockError{
		Message:       "mock error",
		ErrorCategory: llm.CategoryRateLimit,
	}
	catErr, ok = llm.IsCategorizedError(mockErr)
	if !ok || catErr == nil {
		t.Errorf("Expected IsCategorizedError(mockErr) to return (non-nil, true), got (%v, %v)", catErr, ok)
	}
	if ok && catErr.Category() != llm.CategoryRateLimit {
		t.Errorf("Expected category RateLimit, got %v", catErr.Category())
	}
}

// TestMockError tests the MockError implementation of CategorizedError
func TestMockError(t *testing.T) {
	t.Parallel()
	mockErr := &llm.MockError{
		Message:       "test mock error",
		ErrorCategory: llm.CategoryNetwork,
	}

	// Test Error() method
	if mockErr.Error() != "test mock error" {
		t.Errorf("Expected Error() to return %q, got %q", "test mock error", mockErr.Error())
	}

	// Test Category() method
	if mockErr.Category() != llm.CategoryNetwork {
		t.Errorf("Expected Category() to return %v, got %v", llm.CategoryNetwork, mockErr.Category())
	}

	// Test with IsCategory
	if !llm.IsNetwork(mockErr) {
		t.Errorf("Expected IsNetwork(mockErr) to return true")
	}

	// Test with errors.As
	var catErr llm.CategorizedError
	if !errors.As(mockErr, &catErr) {
		t.Errorf("Expected errors.As to recognize MockError as CategorizedError")
	}
}

// TestCategorizedErrorConversion tests error wrapping and unwrapping with category preservation
func TestCategorizedErrorConversion(t *testing.T) {
	t.Parallel(
	// Create a base error
	)

	baseErr := errors.New("base error")

	// Wrap it with the llm.Wrap function
	wrapped := llm.Wrap(baseErr, "test-provider", "wrapped error", llm.CategoryAuth)

	// Test that it preserves the original error
	if !errors.Is(wrapped, baseErr) {
		t.Errorf("Expected wrapped error to be the original error")
	}

	// Test that category is preserved
	if wrapped.Category() != llm.CategoryAuth {
		t.Errorf("Expected Category() to return %v, got %v", llm.CategoryAuth, wrapped.Category())
	}

	// Test re-wrapping an LLMError
	rewrapped := llm.Wrap(wrapped, "new-provider", "rewrapped error", llm.CategoryRateLimit)

	// Check that provider is updated
	if rewrapped.Provider != "new-provider" {
		t.Errorf("Expected Provider to be updated to %q, got %q", "new-provider", rewrapped.Provider)
	}

	// Check that message is updated
	if rewrapped.Message != "rewrapped error" {
		t.Errorf("Expected Message to be updated to %q, got %q", "rewrapped error", rewrapped.Message)
	}

	// Check that category is updated
	if rewrapped.Category() != llm.CategoryRateLimit {
		t.Errorf("Expected Category to be updated to %v, got %v", llm.CategoryRateLimit, rewrapped.Category())
	}

	// Test that original error is still preserved after rewrapping
	if !errors.Is(rewrapped, baseErr) {
		t.Errorf("Expected rewrapped error to still be the original base error")
	}

	// Test wrapping nil error
	if llm.Wrap(nil, "test", "message", llm.CategoryAuth) != nil {
		t.Errorf("Expected Wrap(nil, ...) to return nil")
	}
}
