package testutil

import (
	"os"
	"strings"
	"testing"
)

// GetTestAPIKey safely retrieves a test API key from environment variables.
// The key must have a "test-" prefix to ensure it's not a production key.
// If the key is not provided or doesn't have the proper prefix, the test is skipped.
func GetTestAPIKey(t *testing.T, envVar string) string {
	t.Helper()

	key := os.Getenv(envVar)
	if key == "" {
		t.Skipf("Test API key %s not provided - skipping test requiring real API", envVar)
	}

	// Verify it's a test key (must have test prefix)
	if !strings.HasPrefix(key, "test-") {
		t.Fatalf("API key %s must be a test key (must start with 'test-'), got: %s", envVar, key)
	}

	return key
}

// GetTestAPIKeyOptional retrieves a test API key but doesn't skip the test if not provided.
// Returns an empty string if the key is not set or invalid.
// This is useful for tests that can work with or without real API keys.
func GetTestAPIKeyOptional(t *testing.T, envVar string) string {
	t.Helper()

	key := os.Getenv(envVar)
	if key == "" {
		return ""
	}

	// Verify it's a test key if provided
	if !strings.HasPrefix(key, "test-") {
		t.Logf("Warning: API key %s doesn't have test prefix, ignoring", envVar)
		return ""
	}

	return key
}

// ValidateTestConfiguration ensures the test configuration doesn't contain production URLs or keys.
// This helps prevent accidental production API usage in tests.
func ValidateTestConfiguration(t *testing.T, config interface{}) {
	t.Helper()

	// Add validation logic based on configuration type
	// This is a placeholder for configuration validation
	// Implementation would depend on the specific config struct

	// Example validation patterns:
	// - No production URLs (e.g., api.openai.com, generativelanguage.googleapis.com)
	// - No production API key patterns
	// - Ensure test environment flags are set

	// For now, this is a stub that can be expanded based on specific needs
}

// CreateSecureTestConfig creates a test configuration with secure defaults.
// This ensures tests use safe, non-production settings by default.
func CreateSecureTestConfig() map[string]interface{} {
	return map[string]interface{}{
		"environment":     "test",
		"debug_mode":      true,
		"timeout_seconds": 30,
		"retry_count":     1,
		// Add other secure test defaults
	}
}

// IsTestEnvironment checks if we're running in a test environment.
// This can be used to apply additional safety checks in shared code.
func IsTestEnvironment() bool {
	// Check for common test environment indicators
	if os.Getenv("THINKTANK_ENV") == "test" {
		return true
	}

	// Check if we're running under go test
	if strings.Contains(os.Args[0], ".test") || strings.Contains(os.Args[0], "test") {
		return true
	}

	return false
}

// EnsureTestEnvironment fails the test if not running in a secure test environment.
// Use this for tests that require extra security validation.
func EnsureTestEnvironment(t *testing.T) {
	t.Helper()

	if !IsTestEnvironment() {
		t.Fatal("This test must only run in a secure test environment")
	}

	// Additional safety checks can be added here
	// e.g., checking for test database, test API endpoints, etc.
}
