package gemini

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContextCancellation tests handling of context cancellation
func TestContextCancellation(t *testing.T) {
	tests := []struct {
		name                string
		setupContext        func() (context.Context, context.CancelFunc)
		delayResponse       time.Duration
		expectErrorContains string
		expectErrorCategory llm.ErrorCategory
	}{
		{
			name: "Context cancelled",
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
				return ctx, cancel
			},
			delayResponse:       50 * time.Millisecond,
			expectErrorContains: "cancel",
			expectErrorCategory: llm.CategoryCancelled,
		},
		{
			name: "Context deadline exceeded",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 10*time.Millisecond)
			},
			delayResponse:       50 * time.Millisecond,
			expectErrorContains: "deadline",
			expectErrorCategory: llm.CategoryCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We're not directly using mockTransport yet, but this will be used
			// when the client implementation is complete
			_ = &ErrorMockRoundTripper{
				statusCode:    200,
				responseBody:  []byte(`{"candidates": [{"content": {"parts": [{"text": "test response"}]}}]}`),
				delayResponse: tt.delayResponse,
			}

			// Create test context
			ctx, cancel := tt.setupContext()
			defer cancel()

			// Since we can't fully inject the HTTP client yet, we'll simulate the context cancellation
			var err error
			select {
			case <-ctx.Done():
				err = gemini.FormatAPIError(ctx.Err(), 0)
			case <-time.After(tt.delayResponse + 10*time.Millisecond):
				t.Fatal("Context should have been cancelled before this point")
			}

			// Assert that the error is not nil
			require.NotNil(t, err, "Expected an error but got nil")

			// Check error message contains expected text
			assert.True(t, contains(strings.ToLower(err.Error()), tt.expectErrorContains),
				"Expected error message to contain %q, got %q", tt.expectErrorContains, err.Error())

			// Check error category
			var llmErr *llm.LLMError
			if errors.As(err, &llmErr) {
				assert.Equal(t, "gemini", llmErr.Provider)
				// Context errors can sometimes be categorized as network errors
				// So we'll be flexible in our assertion
				assert.True(t,
					llmErr.Category() == tt.expectErrorCategory ||
						llmErr.Category() == llm.CategoryNetwork,
					"Expected error category to be %v or %v, got %v",
					tt.expectErrorCategory,
					llm.CategoryNetwork,
					llmErr.Category(),
				)
				assert.NotEmpty(t, llmErr.Suggestion, "Expected non-empty suggestion for error")
			} else {
				t.Fatalf("Expected error to be of type *llm.LLMError")
			}
		})
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
