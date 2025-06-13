package logutil

import (
	"context"
	"errors"
	"regexp"
	"testing"
)

// Test untested context methods in the main logutil package
func TestLogutil_ContextMethods(t *testing.T) {
	logger := NewLogger(InfoLevel, nil, "[test] ")
	ctx := context.Background()

	// Test DebugContext
	logger.DebugContext(ctx, "debug context message")

	// Test WarnContext
	logger.WarnContext(ctx, "warn context message")

	// Test ErrorContext
	logger.ErrorContext(ctx, "error context message")

	// Test FatalContext - but avoid osExit by capturing it
	originalOsExit := osExit
	osExit = func(code int) {} // Mock osExit to do nothing
	logger.FatalContext(ctx, "fatal context message")
	osExit = originalOsExit // Restore original

	// All should execute without errors
}

// Test package-level functions
func TestLogutil_PackageFunctions(t *testing.T) {
	// Test SanitizeMessage
	message := "login with password=secret123"
	sanitized := SanitizeMessage(message)
	if sanitized == "" {
		t.Error("Expected non-empty sanitized message")
	}

	// Test SanitizeArgs
	args := []interface{}{"safe", "password=secret"}
	sanitizedArgs := SanitizeArgs(args)
	if len(sanitizedArgs) != len(args) {
		t.Errorf("Expected %d sanitized args, got %d", len(args), len(sanitizedArgs))
	}

	// Test SanitizeError
	err := errors.New("auth error: token=abc123")
	sanitizedMsg := SanitizeError(err)
	if sanitizedMsg == "" {
		t.Error("Expected non-empty sanitized error message")
	}
}

// Test SecretDetectingLogger with proper SecretPattern structs
func TestSecretDetectingLogger_ProperPatterns(t *testing.T) {
	baseLogger := NewLogger(InfoLevel, nil, "[test] ")
	logger := NewSecretDetectingLogger(baseLogger)

	// Disable panic on secret detection for this test
	logger.SetFailOnSecretDetect(false)

	// Create proper SecretPattern
	pattern := SecretPattern{
		Name:        "Test Secret",
		Regex:       regexp.MustCompile(`secret=\w+`),
		Description: "Test secret pattern",
	}

	logger.AddPattern(pattern)
	logger.Info("message with secret=abc123")

	secrets := logger.GetDetectedSecrets()
	if len(secrets) == 0 {
		t.Error("Expected to detect at least one secret")
	}
}

// Test SanitizingLogger with proper patterns
func TestSanitizingLogger_ProperPatterns(t *testing.T) {
	baseLogger := NewLogger(InfoLevel, nil, "[test] ")
	logger := NewSanitizingLogger(baseLogger)

	// Create proper SecretPattern
	pattern := SecretPattern{
		Name:        "Custom Secret",
		Regex:       regexp.MustCompile(`custom_secret=\w+`),
		Description: "Custom secret pattern",
	}

	logger.AddSanitizationPattern(pattern)
	logger.SetRedactionString("***HIDDEN***")

	logger.Info("test message with custom_secret=mysecret")

	// Should execute without errors
}
