// Package testutil contains example integration tests demonstrating secure API key handling
package testutil

import (
	"testing"
)

// TestSecureAPIKeyHandling demonstrates the secure test configuration and API key handling
// This test shows how to safely use real API keys in integration tests
func TestSecureAPIKeyHandling(t *testing.T) {
	// Ensure we're running in a secure test environment
	EnsureTestEnvironment(t)

	// Create secure test configuration
	config := CreateSecureTestConfig()
	ValidateTestConfiguration(t, config)

	t.Run("OpenAI API Key Handling", func(t *testing.T) {
		// Safely retrieve test API key - will skip test if not provided or invalid
		apiKey := GetTestAPIKeyOptional(t, "OPENAI_TEST_API_KEY")
		if apiKey == "" {
			t.Skip("No valid OpenAI test API key provided (OPENAI_TEST_API_KEY), skipping integration test")
		}

		// If we reach this point, we have a valid test API key
		t.Logf("Successfully retrieved secure OpenAI test API key: %s...", apiKey[:10])
		// In a real integration test, you would use this key to create a client and test against the API
	})

	t.Run("Gemini API Key Handling", func(t *testing.T) {
		// Safely retrieve test API key - will skip test if not provided or invalid
		apiKey := GetTestAPIKeyOptional(t, "GEMINI_TEST_API_KEY")
		if apiKey == "" {
			t.Skip("No valid Gemini test API key provided (GEMINI_TEST_API_KEY), skipping integration test")
		}

		// If we reach this point, we have a valid test API key
		t.Logf("Successfully retrieved secure Gemini test API key: %s...", apiKey[:10])
		// In a real integration test, you would use this key to create a client and test against the API
	})

	t.Run("OpenRouter API Key Handling", func(t *testing.T) {
		// Safely retrieve test API key - will skip test if not provided or invalid
		apiKey := GetTestAPIKeyOptional(t, "OPENROUTER_TEST_API_KEY")
		if apiKey == "" {
			t.Skip("No valid OpenRouter test API key provided (OPENROUTER_TEST_API_KEY), skipping integration test")
		}

		// If we reach this point, we have a valid test API key
		t.Logf("Successfully retrieved secure OpenRouter test API key: %s...", apiKey[:10])
		// In a real integration test, you would use this key to create a client and test against the API
	})

	t.Run("Mandatory API Key Validation", func(t *testing.T) {
		// This test demonstrates GetTestAPIKey (as opposed to GetTestAPIKeyOptional)
		// It will skip the test if the key is not provided or doesn't have test- prefix

		// Test with a required API key (this will skip if not provided)
		t.Run("Required API Key Test", func(t *testing.T) {
			// Note: This will skip the test if REQUIRED_TEST_API_KEY is not set
			// or doesn't have the test- prefix
			apiKey := GetTestAPIKey(t, "REQUIRED_TEST_API_KEY")

			// If we reach this point, the API key is valid
			t.Logf("Got valid required test API key: %s...", apiKey[:10])

			// Use the API key for testing...
		})
	})

	t.Run("Security Validation Examples", func(t *testing.T) {
		// Test security validation with different configurations
		t.Run("Valid Test Configuration", func(t *testing.T) {
			config := CreateSecureTestConfig()
			ValidateTestConfiguration(t, config)
			t.Log("Security validation passed for secure test config")
		})

		t.Run("Environment Check", func(t *testing.T) {
			if !IsTestEnvironment() {
				t.Error("Expected to be running in test environment")
			}
			t.Log("Test environment check passed")
		})
	})
}

// TestAPIKeyHandlingExamples demonstrates various API key handling patterns
func TestAPIKeyHandlingExamples(t *testing.T) {
	t.Run("Pattern 1: Optional API Key with Fallback to Mock", func(t *testing.T) {
		apiKey := GetTestAPIKeyOptional(t, "EXAMPLE_TEST_API_KEY")

		if apiKey != "" {
			t.Logf("Using real API with key: %s...", apiKey[:10])
			// Use real API client with the secure test key
		} else {
			t.Log("Using mock implementation")
			// Use mock implementation for testing
		}
	})

	t.Run("Pattern 2: Required API Key (Skip if Not Available)", func(t *testing.T) {
		// This pattern is useful for critical integration tests that must use real APIs
		apiKey := GetTestAPIKey(t, "CRITICAL_TEST_API_KEY")

		// If we reach this line, we have a valid test API key
		t.Logf("Running critical test with API key: %s...", apiKey[:10])
		// Perform critical integration test...
	})

	t.Run("Pattern 3: Validate Test Environment", func(t *testing.T) {
		// For highly sensitive tests, ensure we're in a secure test environment
		EnsureTestEnvironment(t)

		apiKey := GetTestAPIKeyOptional(t, "SENSITIVE_TEST_API_KEY")
		if apiKey != "" {
			t.Logf("Running sensitive test with validated environment and API key: %s...", apiKey[:10])
			// Perform sensitive integration test with additional safety checks
		}
	})
}
