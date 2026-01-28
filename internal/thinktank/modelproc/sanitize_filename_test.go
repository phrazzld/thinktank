package modelproc_test

import (
	"testing"

	"github.com/misty-step/thinktank/internal/thinktank/modelproc"
)

// TestSanitizeFilename tests the filename sanitization function
func TestSanitizeFilename(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no special characters",
			input:    "simple-model-name",
			expected: "simple-model-name",
		},
		{
			name:     "forward slash",
			input:    "gpt/3.5/turbo",
			expected: "gpt-3.5-turbo",
		},
		{
			name:     "backward slash",
			input:    "model\\version",
			expected: "model-version",
		},
		{
			name:     "colon",
			input:    "claude:v1",
			expected: "claude-v1",
		},
		{
			name:     "asterisk",
			input:    "model*name",
			expected: "model-name",
		},
		{
			name:     "question mark",
			input:    "model?name",
			expected: "model-name",
		},
		{
			name:     "double quotes",
			input:    "model\"name",
			expected: "model-name",
		},
		{
			name:     "single quotes",
			input:    "model'name",
			expected: "model-name",
		},
		{
			name:     "less than",
			input:    "model<name",
			expected: "model-name",
		},
		{
			name:     "greater than",
			input:    "model>name",
			expected: "model-name",
		},
		{
			name:     "pipe",
			input:    "model|name",
			expected: "model-name",
		},
		{
			name:     "spaces become underscores",
			input:    "gemini pro model",
			expected: "gemini_pro_model",
		},
		{
			name:     "multiple special characters",
			input:    "gpt/3.5:turbo \"instruct\"",
			expected: "gpt-3.5-turbo_-instruct-",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "/\\:*?\"'<>|",
			expected: "----------",
		},
		{
			name:     "mixed with periods and hyphens",
			input:    "claude-3.5-sonnet",
			expected: "claude-3.5-sonnet",
		},
		{
			name:     "unicode and special characters",
			input:    "model/name:version",
			expected: "model-name-version",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := modelproc.SanitizeFilename(tc.input)
			if result != tc.expected {
				t.Errorf("SanitizeFilename(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestSanitizeFilename_EdgeCases tests edge cases for filename sanitization
func TestSanitizeFilename_EdgeCases(t *testing.T) {
	// Test very long filename
	longInput := "this-is-a-very-long-model-name-that-might-cause-issues-with-filesystem-limits-but-should-still-be-handled-correctly"
	result := modelproc.SanitizeFilename(longInput)
	if result != longInput {
		t.Errorf("Long filename should remain unchanged if no special characters: got %q", result)
	}

	// Test mixed case preservation
	mixedCase := "GPT-4/Turbo"
	expected := "GPT-4-Turbo"
	result = modelproc.SanitizeFilename(mixedCase)
	if result != expected {
		t.Errorf("SanitizeFilename(%q) = %q, expected %q", mixedCase, result, expected)
	}

	// Test numbers and dots
	numberDots := "gpt-3.5.turbo"
	result = modelproc.SanitizeFilename(numberDots)
	if result != numberDots {
		t.Errorf("Numbers and dots should remain unchanged: got %q", result)
	}
}
