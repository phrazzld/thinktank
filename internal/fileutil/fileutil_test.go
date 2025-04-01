// internal/fileutil/fileutil_test.go
package fileutil

import (
	"context"
	"testing"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

func TestEstimateTokenCount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "Single word",
			input:    "hello",
			expected: 1,
		},
		{
			name:     "Two words",
			input:    "hello world",
			expected: 2,
		},
		{
			name:     "Multiple spaces",
			input:    "hello  world",
			expected: 2,
		},
		{
			name:     "With newlines",
			input:    "hello\nworld",
			expected: 2,
		},
		{
			name:     "With tabs",
			input:    "hello\tworld",
			expected: 2,
		},
		{
			name:     "Mixed whitespace",
			input:    "hello\n \t world",
			expected: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := estimateTokenCount(test.input)
			if result != test.expected {
				t.Errorf("estimateTokenCount(%q) = %d, expected %d", test.input, result, test.expected)
			}
		})
	}
}

func TestCalculateStatistics(t *testing.T) {
	input := "Hello\nWorld\nThis is a test."
	expectedChars := len(input)
	expectedLines := 3
	expectedTokens := 6 // "Hello", "World", "This", "is", "a", "test."

	chars, lines, tokens := CalculateStatistics(input)

	if chars != expectedChars {
		t.Errorf("Character count: got %d, want %d", chars, expectedChars)
	}

	if lines != expectedLines {
		t.Errorf("Line count: got %d, want %d", lines, expectedLines)
	}

	if tokens != expectedTokens {
		t.Errorf("Token count: got %d, want %d", tokens, expectedTokens)
	}
}

func TestCalculateStatisticsWithTokenCounting(t *testing.T) {
	input := "Hello world, this is a test of the token counting system."
	ctx := context.Background()

	// Setup a mock client that will return a predefined token count
	mockClient := &gemini.MockClient{
		CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
			return &gemini.TokenCount{Total: 15}, nil
		},
	}

	// Create a mock logger
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ", false)

	// Test with mock client
	chars, lines, tokens := CalculateStatisticsWithTokenCounting(ctx, mockClient, input, logger)

	// Verify character count
	expectedChars := len(input)
	if chars != expectedChars {
		t.Errorf("Character count: got %d, want %d", chars, expectedChars)
	}

	// Verify line count
	expectedLines := 1
	if lines != expectedLines {
		t.Errorf("Line count: got %d, want %d", lines, expectedLines)
	}

	// Verify token count from mock client
	expectedTokens := 15 // From our mock
	if tokens != expectedTokens {
		t.Errorf("Token count: got %d, want %d", tokens, expectedTokens)
	}

	// Now test fallback behavior when client is nil
	chars, lines, tokens = CalculateStatisticsWithTokenCounting(ctx, nil, input, logger)

	// Should use estimation
	expectedTokensFallback := estimateTokenCount(input)
	if tokens != expectedTokensFallback {
		t.Errorf("Fallback token count: got %d, want %d", tokens, expectedTokensFallback)
	}
}
