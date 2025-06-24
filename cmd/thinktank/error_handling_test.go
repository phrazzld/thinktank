// Package thinktank provides the command-line interface for the thinktank tool
package main

import (
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
)

// TestGetFriendlyErrorMessage has been moved to internal/cli package
// This package should not duplicate internal logic

// TestSanitizeErrorMessage has been moved to internal/cli package
// This package should not duplicate internal logic

func TestLLMErrorHandling(t *testing.T) {
	// CPU-bound test: error categorization logic with no I/O dependencies
	t.Parallel()

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
