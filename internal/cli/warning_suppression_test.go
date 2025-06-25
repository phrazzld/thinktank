package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
)

// TestParseFlags_NoDeprecationWarnings_Flag tests that the --no-deprecation-warnings
// flag is recognized and sets the config field correctly
func TestParseFlags_NoDeprecationWarnings_Flag(t *testing.T) {
	args := []string{"thinktank", "--no-deprecation-warnings", "--dry-run", "src/"}
	cfg, err := ParseFlagsWithArgs(args)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !cfg.SuppressDeprecationWarnings {
		t.Error("Expected SuppressDeprecationWarnings to be true with --no-deprecation-warnings flag")
	}
}

// TestParseFlags_NoDeprecationWarnings_DefaultFalse tests that the flag defaults to false
func TestParseFlags_NoDeprecationWarnings_DefaultFalse(t *testing.T) {
	args := []string{"thinktank", "--dry-run", "src/"}
	cfg, err := ParseFlagsWithArgs(args)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cfg.SuppressDeprecationWarnings {
		t.Error("Expected SuppressDeprecationWarnings to default to false")
	}
}

// TestParseFlags_NoDeprecationWarnings_WithOtherFlags tests that the flag works
// in combination with other flags
func TestParseFlags_NoDeprecationWarnings_WithOtherFlags(t *testing.T) {
	args := []string{"thinktank", "--verbose", "--no-deprecation-warnings", "--dry-run", "src/"}
	cfg, err := ParseFlagsWithArgs(args)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !cfg.SuppressDeprecationWarnings {
		t.Error("Expected SuppressDeprecationWarnings to be true with --no-deprecation-warnings flag")
	}

	if !cfg.Verbose {
		t.Error("Expected Verbose to be true with --verbose flag")
	}

	if !cfg.DryRun {
		t.Error("Expected DryRun to be true with --dry-run flag")
	}
}

// Using the existing MockLogger from parser_router_test.go

// TestLogDeprecationWarning_WithConfig_Suppression tests that LogDeprecationWarning
// accepts a config parameter and suppresses warnings based on the config
func TestLogDeprecationWarning_WithConfig_Suppression(t *testing.T) {
	logger := &MockLogger{}
	router := NewParserRouter(logger)

	warning := &DeprecationWarning{
		Message:    "Test warning",
		FlagUsed:   "--test",
		Suggestion: "Use new approach",
	}

	config := &config.CliConfig{
		SuppressDeprecationWarnings: true,
	}

	router.LogDeprecationWarning(warning, config)

	// Should log to debug, not warn
	if len(logger.WarnCalls) != 0 {
		t.Error("Expected no warn calls when suppressed via config")
	}

	if len(logger.DebugCalls) != 1 {
		t.Error("Expected debug call when suppressed via config")
	}

	if !strings.Contains(logger.DebugCalls[0].Message, "suppressed") {
		t.Errorf("Expected debug message to indicate suppression, got: %s", logger.DebugCalls[0].Message)
	}
}

// TestLogDeprecationWarning_WithConfig_NotSuppressed tests that warnings are shown
// when suppression is disabled in config
func TestLogDeprecationWarning_WithConfig_NotSuppressed(t *testing.T) {
	logger := &MockLogger{}
	router := NewParserRouter(logger)

	warning := &DeprecationWarning{
		Message:    "Test warning",
		FlagUsed:   "--test",
		Suggestion: "Use new approach",
	}

	config := &config.CliConfig{
		SuppressDeprecationWarnings: false,
	}

	router.LogDeprecationWarning(warning, config)

	// Should show warning normally
	if len(logger.WarnCalls) != 1 {
		t.Error("Expected warn call when not suppressed via config")
	}

	if len(logger.DebugCalls) != 0 {
		t.Error("Expected no debug calls when not suppressed")
	}
}

// TestLogDeprecationWarning_PrecedenceCliOverEnv tests that CLI flag takes precedence
// over environment variable when the CLI flag is explicitly set to true
func TestLogDeprecationWarning_PrecedenceCliOverEnv(t *testing.T) {
	// Set environment variable to NOT suppress, but CLI flag to suppress
	_ = os.Setenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS", "")
	defer func() { _ = os.Unsetenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS") }()

	logger := &MockLogger{}
	router := NewParserRouter(logger)

	warning := &DeprecationWarning{
		Message:    "Test",
		FlagUsed:   "--test",
		Suggestion: "New way",
	}

	config := &config.CliConfig{
		SuppressDeprecationWarnings: true, // CLI flag explicitly true
	}

	router.LogDeprecationWarning(warning, config)

	// Should suppress because CLI flag overrides ENV
	if len(logger.WarnCalls) != 0 {
		t.Error("Expected no warn calls when CLI flag true overrides ENV false")
	}

	if len(logger.DebugCalls) != 1 {
		t.Error("Expected debug call when CLI flag true overrides ENV false")
	}
}

// TestLogDeprecationWarning_EnvironmentVariableBackwardCompatibility tests that
// the environment variable still works when no CLI flag is set
func TestLogDeprecationWarning_EnvironmentVariableBackwardCompatibility(t *testing.T) {
	// Set environment variable to suppress
	_ = os.Setenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS", "true")
	defer func() { _ = os.Unsetenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS") }()

	logger := &MockLogger{}
	router := NewParserRouter(logger)

	warning := &DeprecationWarning{
		Message:    "Test",
		FlagUsed:   "--test",
		Suggestion: "New way",
	}

	// Config with default value (false) - should respect environment variable
	config := &config.CliConfig{
		SuppressDeprecationWarnings: false, // Default value
	}

	router.LogDeprecationWarning(warning, config)

	// Should suppress because of environment variable
	if len(logger.WarnCalls) != 0 {
		t.Error("Expected no warn calls when suppressed via environment variable")
	}

	if len(logger.DebugCalls) != 1 {
		t.Error("Expected debug call when suppressed via environment variable")
	}
}

// TestDeprecationSuppression_EndToEnd tests the complete flow from CLI parsing
// to warning suppression
func TestDeprecationSuppression_EndToEnd(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()
	instructionsFile := createTestFile(t, tempDir, "instructions.txt", "Test instructions")
	targetDir := createTestDir(t, tempDir, "src")

	// Test 1: With suppression flag, warnings should be suppressed
	t.Run("WithSuppressionFlag", func(t *testing.T) {
		args := []string{"thinktank", "--no-deprecation-warnings", "--instructions", instructionsFile, "--dry-run", targetDir}
		config, err := ParseFlagsWithArgs(args)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !config.SuppressDeprecationWarnings {
			t.Error("Expected SuppressDeprecationWarnings to be true")
		}

		// Test that the router respects the config
		logger := &MockLogger{}
		router := NewParserRouter(logger)

		warning := &DeprecationWarning{
			Message:    "Test warning",
			FlagUsed:   "--instructions",
			Suggestion: "Use simplified interface",
		}

		router.LogDeprecationWarning(warning, config)

		// Should be suppressed
		if len(logger.WarnCalls) != 0 {
			t.Error("Expected no warn calls when suppressed via CLI flag")
		}

		if len(logger.DebugCalls) != 1 {
			t.Error("Expected debug call when suppressed via CLI flag")
		}
	})

	// Test 2: Without suppression flag, warnings should be shown
	t.Run("WithoutSuppressionFlag", func(t *testing.T) {
		args := []string{"thinktank", "--instructions", instructionsFile, "--dry-run", targetDir}
		config, err := ParseFlagsWithArgs(args)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if config.SuppressDeprecationWarnings {
			t.Error("Expected SuppressDeprecationWarnings to be false")
		}

		// Test that the router respects the config
		logger := &MockLogger{}
		router := NewParserRouter(logger)

		warning := &DeprecationWarning{
			Message:    "Test warning",
			FlagUsed:   "--instructions",
			Suggestion: "Use simplified interface",
		}

		router.LogDeprecationWarning(warning, config)

		// Should show warning
		if len(logger.WarnCalls) != 1 {
			t.Error("Expected warn call when not suppressed")
		}
	})

	// Test 3: Environment variable works with default config
	t.Run("WithEnvironmentVariable", func(t *testing.T) {
		_ = os.Setenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS", "true")
		defer func() { _ = os.Unsetenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS") }()

		args := []string{"thinktank", "--instructions", instructionsFile, "--dry-run", targetDir}
		config, err := ParseFlagsWithArgs(args)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Config should still be false (environment variable doesn't affect parsing)
		if config.SuppressDeprecationWarnings {
			t.Error("Expected SuppressDeprecationWarnings to be false in config")
		}

		// But router should still suppress due to environment variable
		logger := &MockLogger{}
		router := NewParserRouter(logger)

		warning := &DeprecationWarning{
			Message:    "Test warning",
			FlagUsed:   "--instructions",
			Suggestion: "Use simplified interface",
		}

		router.LogDeprecationWarning(warning, config)

		// Should be suppressed due to environment variable
		if len(logger.WarnCalls) != 0 {
			t.Error("Expected no warn calls when suppressed via environment variable")
		}

		if len(logger.DebugCalls) != 1 {
			t.Error("Expected debug call when suppressed via environment variable")
		}
	})
}
