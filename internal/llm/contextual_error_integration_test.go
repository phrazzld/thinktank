package llm_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/misty-step/thinktank/internal/llm"
)

// TestErrorChainPreservation validates that error categories and information
// are preserved through the full call chain as used in production.
// This test follows the TDD RED phase - defining what we expect from our
// enhanced simple error system.
func TestErrorChainPreservation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		originalError     error
		provider          string
		message           string
		category          llm.ErrorCategory
		wantCategory      llm.ErrorCategory
		wantContains      []string
		wantCorrelationID string // This will initially fail - we'll implement this
	}{
		{
			name:          "API client auth error propagation",
			originalError: errors.New("invalid API key"),
			provider:      "openai",
			message:       "authentication failed",
			category:      llm.CategoryAuth,
			wantCategory:  llm.CategoryAuth,
			wantContains:  []string{"authentication failed", "invalid API key"},
		},
		{
			name:          "Model processor rate limit error",
			originalError: errors.New("rate limit exceeded"),
			provider:      "gemini",
			message:       "request rate limited",
			category:      llm.CategoryRateLimit,
			wantCategory:  llm.CategoryRateLimit,
			wantContains:  []string{"request rate limited", "rate limit exceeded"},
		},
		{
			name:          "Orchestrator network error wrapping",
			originalError: errors.New("connection timeout"),
			provider:      "orchestrator",
			message:       "network operation failed",
			category:      llm.CategoryNetwork,
			wantCategory:  llm.CategoryNetwork,
			wantContains:  []string{"network operation failed", "connection timeout"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use existing llm.Wrap as production does
			wrappedErr := llm.Wrap(tt.originalError, tt.provider, tt.message, tt.category)

			// Validate error is properly categorized
			if !llm.IsCategory(wrappedErr, tt.wantCategory) {
				t.Errorf("Expected error category %v, got %v", tt.wantCategory, getCategoryFromError(wrappedErr))
			}

			// Validate error message contains expected content
			errMsg := wrappedErr.Error()
			for _, wantStr := range tt.wantContains {
				if !strings.Contains(errMsg, wantStr) {
					t.Errorf("Expected error message to contain %q, got: %s", wantStr, errMsg)
				}
			}

			// Validate original error is preserved in chain
			if !errors.Is(wrappedErr, tt.originalError) {
				t.Errorf("Expected wrapped error to preserve original error in chain")
			}

			// This will initially fail - we want to add minimal correlation ID support
			// to the existing llm.Wrap function without breaking existing usage
			if tt.wantCorrelationID != "" {
				correlationID := extractCorrelationIDFromError(wrappedErr)
				if correlationID != tt.wantCorrelationID {
					t.Errorf("Expected correlation ID %q, got %q", tt.wantCorrelationID, correlationID)
				}
			}
		})
	}
}

// TestCorrelationIDSupport tests optional correlation ID functionality
// that we want to add to the existing llm.Wrap function.
// This follows TDD RED phase - test fails until we implement the feature.
func TestCorrelationIDSupport(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		baseError     error
		provider      string
		message       string
		category      llm.ErrorCategory
		correlationID string
		wantExtracted string
	}{
		{
			name:          "wrap with correlation ID",
			baseError:     errors.New("base error"),
			provider:      "test",
			message:       "test message",
			category:      llm.CategoryAuth,
			correlationID: "req-123-456",
			wantExtracted: "req-123-456",
		},
		{
			name:          "wrap without correlation ID",
			baseError:     errors.New("base error"),
			provider:      "test",
			message:       "test message",
			category:      llm.CategoryNetwork,
			correlationID: "", // empty correlation ID should work fine
			wantExtracted: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail initially - we need to add correlation ID support
			// to llm.Wrap function signature and implementation
			wrappedErr := wrapWithCorrelationID(tt.baseError, tt.provider, tt.message, tt.category, tt.correlationID)

			extracted := extractCorrelationIDFromError(wrappedErr)
			if extracted != tt.wantExtracted {
				t.Errorf("Expected correlation ID %q, got %q", tt.wantExtracted, extracted)
			}
		})
	}
}

// TestCorrelationIDIntegration tests the end-to-end correlation ID functionality
// from error creation through extraction in logging systems.
func TestCorrelationIDIntegration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		createError           func() error
		expectedCorrelationID string
		wantCategory          llm.ErrorCategory
	}{
		{
			name: "API client error with correlation ID propagates through chain",
			createError: func() error {
				baseErr := errors.New("network timeout")
				return llm.WrapWithCorrelationID(baseErr, "openai", "API request failed", llm.CategoryNetwork, "req-api-123")
			},
			expectedCorrelationID: "req-api-123",
			wantCategory:          llm.CategoryNetwork,
		},
		{
			name: "Model processor error without correlation ID",
			createError: func() error {
				baseErr := errors.New("token limit exceeded")
				return llm.Wrap(baseErr, "model-processor", "input too long", llm.CategoryInputLimit)
			},
			expectedCorrelationID: "",
			wantCategory:          llm.CategoryInputLimit,
		},
		{
			name: "Orchestrator error wrapping error with existing correlation ID",
			createError: func() error {
				// Start with an error that has a correlation ID
				baseErr := llm.WrapWithCorrelationID(errors.New("auth failed"), "gemini", "invalid key", llm.CategoryAuth, "req-orchestrator-456")
				// Wrap it again in orchestrator layer - correlation ID should be preserved
				return llm.Wrap(baseErr, "orchestrator", "model processing failed", llm.CategoryAuth)
			},
			expectedCorrelationID: "req-orchestrator-456",
			wantCategory:          llm.CategoryAuth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createError()

			// Validate error category is preserved
			if !llm.IsCategory(err, tt.wantCategory) {
				t.Errorf("Expected error category %v, got %v", tt.wantCategory, getCategoryFromError(err))
			}

			// Validate correlation ID extraction works correctly
			extractedID := llm.ExtractCorrelationID(err)
			if extractedID != tt.expectedCorrelationID {
				t.Errorf("Expected correlation ID %q, got %q", tt.expectedCorrelationID, extractedID)
			}

			// Validate that the error is still a proper LLMError
			if _, ok := llm.IsCategorizedError(err); !ok {
				t.Errorf("Expected error to implement CategorizedError interface")
			}
		})
	}
}

// TestProductionErrorPatterns validates that our current production error
// patterns continue to work exactly as they do today.
func TestProductionErrorPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pattern  func() error
		wantType string
		wantCat  llm.ErrorCategory
	}{
		{
			name: "API client error pattern",
			pattern: func() error {
				baseErr := errors.New("HTTP 401 Unauthorized")
				return llm.Wrap(baseErr, "openai", "authentication failed", llm.CategoryAuth)
			},
			wantType: "*llm.LLMError",
			wantCat:  llm.CategoryAuth,
		},
		{
			name: "model processor error pattern",
			pattern: func() error {
				baseErr := errors.New("token limit exceeded")
				return llm.Wrap(baseErr, "model-processor", "input too long", llm.CategoryInputLimit)
			},
			wantType: "*llm.LLMError",
			wantCat:  llm.CategoryInputLimit,
		},
		{
			name: "orchestrator error pattern",
			pattern: func() error {
				baseErr := errors.New("file write failed")
				return llm.Wrap(baseErr, "orchestrator", "output save failed", llm.CategoryServer)
			},
			wantType: "*llm.LLMError",
			wantCat:  llm.CategoryServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pattern()

			// Validate it's the expected type
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			// Validate it's categorized correctly
			if !llm.IsCategory(err, tt.wantCat) {
				t.Errorf("Expected category %v, got %v", tt.wantCat, getCategoryFromError(err))
			}

			// Validate it implements CategorizedError interface
			if _, ok := llm.IsCategorizedError(err); !ok {
				t.Errorf("Expected error to implement CategorizedError interface")
			}
		})
	}
}

// Helper functions that will initially fail - we'll implement these

// getCategoryFromError extracts the category from an error
func getCategoryFromError(err error) llm.ErrorCategory {
	if catErr, ok := llm.IsCategorizedError(err); ok {
		return catErr.Category()
	}
	return llm.CategoryUnknown
}

// extractCorrelationIDFromError extracts correlation ID from error
// Uses the existing ExtractCorrelationID function from llm package
func extractCorrelationIDFromError(err error) string {
	return llm.ExtractCorrelationID(err)
}

// wrapWithCorrelationID wraps error with correlation ID support
// Uses the new WrapWithCorrelationID function we just implemented
func wrapWithCorrelationID(err error, provider, message string, category llm.ErrorCategory, correlationID string) error {
	return llm.WrapWithCorrelationID(err, provider, message, category, correlationID)
}
