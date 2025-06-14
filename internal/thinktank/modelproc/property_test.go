package modelproc_test

import (
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/phrazzld/thinktank/internal/thinktank/modelproc"
)

// TestSanitizeFilename_Properties verifies invariants of the SanitizeFilename function
// using property-based testing with the Rapid library.
func TestSanitizeFilename_Properties(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate arbitrary input string
		input := rapid.String().Draw(t, "input")

		// Call the function under test
		result := modelproc.SanitizeFilename(input)

		// Property 1: Output never contains dangerous characters
		dangerousChars := []string{"/", "\\", ":", "*", "?", "\"", "'", "<", ">", "|"}
		for _, char := range dangerousChars {
			if strings.Contains(result, char) {
				t.Fatalf("SanitizeFilename(%q) = %q contains dangerous character %q", input, result, char)
			}
		}

		// Property 2: Length relationship - output should not be longer than input
		// (characters are replaced, not added)
		if len(result) > len(input) {
			t.Fatalf("SanitizeFilename(%q) = %q output longer than input (len=%d > %d)",
				input, result, len(result), len(input))
		}

		// Property 3: Idempotency - sanitizing a sanitized filename should not change it
		resanitized := modelproc.SanitizeFilename(result)
		if resanitized != result {
			t.Fatalf("SanitizeFilename not idempotent: SanitizeFilename(%q) = %q, but SanitizeFilename(%q) = %q",
				input, result, result, resanitized)
		}

		// Property 4: Safe characters should remain unchanged
		safeChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-"
		for _, char := range safeChars {
			charStr := string(char)
			sanitized := modelproc.SanitizeFilename(charStr)
			if sanitized != charStr {
				t.Fatalf("Safe character %q was changed to %q", charStr, sanitized)
			}
		}
	})
}

// TestSanitizeFilename_SpecificMappings tests that specific character mappings
// are consistent using property-based testing.
func TestSanitizeFilename_SpecificMappings(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate string with specific characters we want to test
		prefix := rapid.StringMatching(`[a-zA-Z0-9]*`).Draw(t, "prefix")
		suffix := rapid.StringMatching(`[a-zA-Z0-9]*`).Draw(t, "suffix")

		// Test space mapping
		inputWithSpace := prefix + " " + suffix
		result := modelproc.SanitizeFilename(inputWithSpace)
		expected := prefix + "_" + suffix
		if result != expected {
			t.Fatalf("Space mapping failed: SanitizeFilename(%q) = %q, expected %q",
				inputWithSpace, result, expected)
		}

		// Test slash mapping
		inputWithSlash := prefix + "/" + suffix
		result = modelproc.SanitizeFilename(inputWithSlash)
		expected = prefix + "-" + suffix
		if result != expected {
			t.Fatalf("Slash mapping failed: SanitizeFilename(%q) = %q, expected %q",
				inputWithSlash, result, expected)
		}
	})
}

// TestSanitizeFilename_ModelNameProperties tests properties specific to model names
// using our custom generators.
func TestSanitizeFilename_ModelNameProperties(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Use our custom model name generator
		modelName := testutil.ModelName.Draw(t, "modelName")

		// Sanitize the model name
		sanitized := modelproc.SanitizeFilename(modelName)

		// Property: Sanitized model name should be non-empty for non-empty input
		if modelName != "" && sanitized == "" {
			t.Fatalf("Sanitized model name should not be empty for non-empty input %q", modelName)
		}

		// Property: Should be suitable for filename (no path separators)
		if strings.Contains(sanitized, "/") || strings.Contains(sanitized, "\\") {
			t.Fatalf("Sanitized model name %q contains path separators", sanitized)
		}

		// Property: Should preserve alphanumeric characters
		hasAlphaNum := false
		for _, char := range modelName {
			if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
				hasAlphaNum = true
				break
			}
		}

		if hasAlphaNum {
			// Should have at least some alphanumeric content preserved
			preservedAlphaNum := false
			for _, char := range sanitized {
				if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
					preservedAlphaNum = true
					break
				}
			}
			if !preservedAlphaNum {
				t.Fatalf("Sanitized name %q should preserve alphanumeric content from %q", sanitized, modelName)
			}
		}
	})
}

// TestSanitizeFilename_EmptyAndSpecialCases tests edge cases using property-based testing.
func TestSanitizeFilename_EmptyAndSpecialCases(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Test with strings containing only special characters
		specialChars := []string{"/", "\\", ":", "*", "?", "\"", "'", "<", ">", "|"}

		// Generate strings of only special characters
		numChars := rapid.IntRange(1, 10).Draw(t, "numChars")
		var input strings.Builder
		for i := 0; i < numChars; i++ {
			char := rapid.SampledFrom(specialChars).Draw(t, "char")
			input.WriteString(char)
		}

		inputStr := input.String()
		result := modelproc.SanitizeFilename(inputStr)

		// Property: Result should contain only replacement characters
		for _, char := range result {
			if char != '-' && char != '_' {
				t.Fatalf("Result %q from input %q contains unexpected character %q", result, inputStr, string(char))
			}
		}

		// Property: Length should be preserved
		if len(result) != len(inputStr) {
			t.Fatalf("Length not preserved: input %q (len=%d), result %q (len=%d)",
				inputStr, len(inputStr), result, len(result))
		}
	})
}

// TestSanitizeFilename_UnicodeHandling tests how the function handles Unicode characters.
func TestSanitizeFilename_UnicodeHandling(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate strings with Unicode characters mixed with ASCII
		unicode := rapid.StringMatching(`[\p{L}\p{N}]*`).Draw(t, "unicode")
		ascii := rapid.StringMatching(`[a-zA-Z0-9]*`).Draw(t, "ascii")

		// Combine them in various ways
		combined := unicode + ascii
		result := modelproc.SanitizeFilename(combined)

		// Property: Unicode letters and numbers should be preserved
		// (SanitizeFilename only replaces specific problematic characters)
		expectedLength := len(combined)
		if len(result) != expectedLength {
			// Only expect length change if input contained characters that get replaced
			containsProblematic := strings.ContainsAny(combined, "/\\:*?\"'<>| ")
			if !containsProblematic {
				t.Fatalf("Unicode handling failed: input %q (len=%d), result %q (len=%d)",
					combined, len(combined), result, len(result))
			}
		}
	})
}
