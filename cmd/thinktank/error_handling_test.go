// Package thinktank provides the command-line interface for the thinktank tool
package thinktank

import (
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
)

func TestGetFriendlyErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "An unknown error occurred",
		},
		{
			name:     "authentication error",
			err:      errors.New("failed to authenticate with the API"),
			expected: "Authentication error: Please check your API key and permissions",
		},
		{
			name:     "rate limit error",
			err:      errors.New("rate limit exceeded"),
			expected: "Rate limit exceeded: Too many requests. Please try again later or adjust rate limits.",
		},
		{
			name:     "timeout error",
			err:      errors.New("operation timed out after 5m"),
			expected: "Operation timed out. Consider using a longer timeout with the --timeout flag.",
		},
		{
			name:     "not found error",
			err:      errors.New("model not found: claude-3"),
			expected: "Resource not found. Please check that the specified file paths or models exist.",
		},
		{
			name:     "file permission error",
			err:      errors.New("file permission denied: /tmp/output.txt"),
			expected: "File permission error: Please check file permissions and try again.",
		},
		{
			name:     "file error",
			err:      errors.New("file not readable"),
			expected: "File error: file not readable",
		},
		{
			name:     "flag error",
			err:      errors.New("unknown flag: --invalid"),
			expected: "Invalid command line arguments. Use --help to see usage instructions.",
		},
		{
			name:     "context cancelled error",
			err:      errors.New("context cancelled"),
			expected: "Operation was cancelled. This might be due to timeout or user interruption.",
		},
		{
			name:     "network error",
			err:      errors.New("network connection failed"),
			expected: "Network error: Please check your internet connection and try again.",
		},
		{
			name:     "generic error",
			err:      errors.New("something unexpected happened"),
			expected: "something unexpected happened",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFriendlyErrorMessage(tt.err)
			if result != tt.expected {
				t.Errorf("getFriendlyErrorMessage() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizeErrorMessage(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		shouldContain string
	}{
		{
			name:          "no sensitive information",
			message:       "Simple error message",
			shouldContain: "Simple error message",
		},
		{
			name:          "contains API key - OpenAI",
			message:       "Failed API call with key sk_abcdef1234567890abcdef12345",
			shouldContain: "[REDACTED]",
		},
		{
			name:          "contains API key - Gemini",
			message:       "Failed API call with key_abcdef1234567890abcdef12345",
			shouldContain: "[REDACTED]",
		},
		{
			name:          "contains long alphanumeric string",
			message:       "Error with token 1234567890abcdef1234567890abcdef",
			shouldContain: "[REDACTED]",
		},
		{
			name:          "contains URL with credentials",
			message:       "Failed to connect to https://user:password@example.com",
			shouldContain: "[REDACTED]",
		},
		{
			name:          "contains environment variable",
			message:       "OPENAI_API_KEY=sk-1234567890abcdef",
			shouldContain: "[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeErrorMessage(tt.message)
			if !strings.Contains(result, tt.shouldContain) {
				t.Errorf("sanitizeErrorMessage(%q) = %q, should contain %q", tt.message, result, tt.shouldContain)
			}
		})
	}
}

func TestLLMErrorHandling(t *testing.T) {
	// Create test LLMErrors with different categories
	authError := llm.New("test", "AUTH_ERR", 401, "Authentication failed", "req123", errors.New("invalid key"), llm.CategoryAuth)
	rateLimitError := llm.New("test", "RATE_LIMIT", 429, "Rate limit exceeded", "req456", errors.New("too many requests"), llm.CategoryRateLimit)
	serverError := llm.New("test", "SERVER_ERR", 500, "Server error", "req789", errors.New("internal server error"), llm.CategoryServer)

	// Verify error categories
	if !llm.IsAuth(authError) {
		t.Errorf("Expected authError to be identified as CategoryAuth")
	}

	if !llm.IsRateLimit(rateLimitError) {
		t.Errorf("Expected rateLimitError to be identified as CategoryRateLimit")
	}

	if !llm.IsServer(serverError) {
		t.Errorf("Expected serverError to be identified as CategoryServer")
	}

	// Test wrapped errors
	wrappedErr := errors.New("original error")
	wrappedLLMErr := llm.Wrap(wrappedErr, "test", "wrapped error", llm.CategoryNetwork)

	if !llm.IsNetwork(wrappedLLMErr) {
		t.Errorf("Expected wrapped error to be identified as CategoryNetwork")
	}

	if !errors.Is(wrappedLLMErr, wrappedErr) {
		t.Errorf("errors.Is() failed for wrapped error")
	}
}
