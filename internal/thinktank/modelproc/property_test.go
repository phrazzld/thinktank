package modelproc_test

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/misty-step/thinktank/internal/testutil"
	"github.com/misty-step/thinktank/internal/thinktank/modelproc"
)

// TestSanitizeFilename_Properties verifies invariants of the SanitizeFilename function
// using property-based testing with the Gopter library.
func TestSanitizeFilename_Properties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("SanitizeFilename preserves safety invariants", prop.ForAll(
		func(input string) bool {
			// Call the function under test
			result := modelproc.SanitizeFilename(input)

			// Property 1: Output never contains dangerous characters
			dangerousChars := []string{"/", "\\", ":", "*", "?", "\"", "'", "<", ">", "|"}
			for _, char := range dangerousChars {
				if strings.Contains(result, char) {
					t.Errorf("SanitizeFilename(%q) = %q contains dangerous character %q", input, result, char)
					return false
				}
			}

			// Property 2: Length relationship - output should not be longer than input
			// (characters are replaced, not added)
			if len(result) > len(input) {
				t.Errorf("SanitizeFilename(%q) = %q output longer than input (len=%d > %d)",
					input, result, len(result), len(input))
				return false
			}

			// Property 3: Idempotency - sanitizing a sanitized filename should not change it
			resanitized := modelproc.SanitizeFilename(result)
			if resanitized != result {
				t.Errorf("SanitizeFilename not idempotent: SanitizeFilename(%q) = %q, but SanitizeFilename(%q) = %q",
					input, result, result, resanitized)
				return false
			}

			// Property 4: Safe characters should remain unchanged
			safeChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-"
			for _, char := range safeChars {
				charStr := string(char)
				sanitized := modelproc.SanitizeFilename(charStr)
				if sanitized != charStr {
					t.Errorf("Safe character %q was changed to %q", charStr, sanitized)
					return false
				}
			}

			return true
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// TestSanitizeFilename_SpecificMappings tests that specific character mappings
// are consistent using property-based testing.
func TestSanitizeFilename_SpecificMappings(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("SanitizeFilename character mappings", prop.ForAll(
		func(prefix, suffix string) bool {
			// Test space mapping
			inputWithSpace := prefix + " " + suffix
			result := modelproc.SanitizeFilename(inputWithSpace)
			expected := prefix + "_" + suffix
			if result != expected {
				t.Errorf("Space mapping failed: SanitizeFilename(%q) = %q, expected %q",
					inputWithSpace, result, expected)
				return false
			}

			// Test slash mapping
			inputWithSlash := prefix + "/" + suffix
			result = modelproc.SanitizeFilename(inputWithSlash)
			expected = prefix + "-" + suffix
			if result != expected {
				t.Errorf("Slash mapping failed: SanitizeFilename(%q) = %q, expected %q",
					inputWithSlash, result, expected)
				return false
			}

			return true
		},
		gen.RegexMatch(`[a-zA-Z0-9]*`),
		gen.RegexMatch(`[a-zA-Z0-9]*`),
	))

	properties.TestingRun(t)
}

// TestSanitizeFilename_ModelNameProperties tests properties specific to model names
// using our custom generators.
func TestSanitizeFilename_ModelNameProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("SanitizeFilename model name properties", prop.ForAll(
		func(modelName string) bool {
			// Sanitize the model name
			sanitized := modelproc.SanitizeFilename(modelName)

			// Property: Sanitized model name should be non-empty for non-empty input
			if modelName != "" && sanitized == "" {
				t.Errorf("Sanitized model name should not be empty for non-empty input %q", modelName)
				return false
			}

			// Property: Should be suitable for filename (no path separators)
			if strings.Contains(sanitized, "/") || strings.Contains(sanitized, "\\") {
				t.Errorf("Sanitized model name %q contains path separators", sanitized)
				return false
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
					t.Errorf("Sanitized name %q should preserve alphanumeric content from %q", sanitized, modelName)
					return false
				}
			}

			return true
		},
		testutil.ModelName,
	))

	properties.TestingRun(t)
}

// TestSanitizeFilename_EmptyAndSpecialCases tests edge cases using property-based testing.
func TestSanitizeFilename_EmptyAndSpecialCases(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("SanitizeFilename special character handling", prop.ForAll(
		func(numChars int) bool {
			// Test with strings containing only special characters
			specialChars := []string{"/", "\\", ":", "*", "?", "\"", "'", "<", ">", "|"}

			// Generate strings of only special characters
			var input strings.Builder
			for i := 0; i < numChars; i++ {
				char := specialChars[i%len(specialChars)]
				input.WriteString(char)
			}

			inputStr := input.String()
			result := modelproc.SanitizeFilename(inputStr)

			// Property: Result should contain only replacement characters
			for _, char := range result {
				if char != '-' && char != '_' {
					t.Errorf("Result %q from input %q contains unexpected character %q", result, inputStr, string(char))
					return false
				}
			}

			// Property: Length should be preserved
			if len(result) != len(inputStr) {
				t.Errorf("Length not preserved: input %q (len=%d), result %q (len=%d)",
					inputStr, len(inputStr), result, len(result))
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// TestSanitizeFilename_UnicodeHandling tests how the function handles Unicode characters.
func TestSanitizeFilename_UnicodeHandling(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("SanitizeFilename Unicode handling", prop.ForAll(
		func(unicode, ascii string) bool {
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
					t.Errorf("Unicode handling failed: input %q (len=%d), result %q (len=%d)",
						combined, len(combined), result, len(result))
					return false
				}
			}

			return true
		},
		gen.RegexMatch(`[\p{L}\p{N}]*`),
		gen.RegexMatch(`[a-zA-Z0-9]*`),
	))

	properties.TestingRun(t)
}
