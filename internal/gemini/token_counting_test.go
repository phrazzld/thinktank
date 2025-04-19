// internal/gemini/token_counting_test.go
// Tests for token counting functionality in the gemini client

package gemini

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/llm"
)

func TestCountTokens(t *testing.T) {
	t.Skip("TODO: Update this test for the new llm.LLMClient token interface")
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
				// Using the llm package for error categories
				return nil, CreateAPIError(
					llm.CategoryInvalidRequest,
					"Failed to count tokens in prompt",
					errors.New("API error: invalid request"),
					"Check your API key and internet connection.")
			},
		}

		// Call CountTokens
		result, err := client.CountTokens(context.Background(), testPrompt)

		// Verify error is returned
		if err == nil {
			t.Fatal("Expected error from API, got nil")
		}

		// Verify it's an APIError with the expected type
		apiErr, ok := IsGeminiError(err)
		if !ok {
			t.Fatalf("Expected *APIError, got %T", err)
		}

		if GetErrorType(apiErr) != ErrorTypeInvalidRequest {
			t.Errorf("Expected error type %d, got %d", ErrorTypeInvalidRequest, GetErrorType(apiErr))
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
				return nil, CreateAPIError(
					llm.CategoryRateLimit,
					"Request rate limit or quota exceeded on the Gemini API",
					errors.New("API error: rate limit exceeded"),
					"Wait and try again later.")
			},
		}

		// Call CountTokens
		result, err := client.CountTokens(context.Background(), testPrompt)

		// Verify error is returned
		if err == nil {
			t.Fatal("Expected error from API, got nil")
		}

		// Verify it's an APIError with the expected type
		apiErr, ok := IsGeminiError(err)
		if !ok {
			t.Fatalf("Expected *APIError, got %T", err)
		}

		if GetErrorType(apiErr) != ErrorTypeRateLimit {
			t.Errorf("Expected error type %d, got %d", ErrorTypeRateLimit, GetErrorType(apiErr))
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
				return nil, CreateAPIError(
					llm.CategoryNetwork,
					"Network error while connecting to the Gemini API",
					errors.New("network error: connection refused"),
					"Check your internet connection and try again.")
			},
		}

		// Call CountTokens
		result, err := client.CountTokens(context.Background(), testPrompt)

		// Verify error is returned
		if err == nil {
			t.Fatal("Expected error from API, got nil")
		}

		// Verify it's an APIError with the expected type
		apiErr, ok := IsGeminiError(err)
		if !ok {
			t.Fatalf("Expected *APIError, got %T", err)
		}

		if GetErrorType(apiErr) != ErrorTypeNetwork {
			t.Errorf("Expected error type %d, got %d", ErrorTypeNetwork, GetErrorType(apiErr))
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
