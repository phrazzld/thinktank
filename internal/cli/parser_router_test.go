package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestParsingModeString tests the string representation of parsing modes
func TestParsingModeString(t *testing.T) {
	testCases := []struct {
		mode     ParsingMode
		expected string
	}{
		{SimplifiedMode, "simplified"},
		{ComplexMode, "complex"},
		{ParsingMode(999), "unknown"},
	}

	for _, tc := range testCases {
		if got := tc.mode.String(); got != tc.expected {
			t.Errorf("ParsingMode(%d).String() = %q, want %q", int(tc.mode), got, tc.expected)
		}
	}
}

// TestDetectParsingMode tests the parsing mode detection logic
func TestDetectParsingMode(t *testing.T) {
	router := NewParserRouter(nil)

	testCases := []struct {
		name     string
		args     []string
		expected ParsingMode
	}{
		{
			name:     "SimplifiedMode_BasicCase",
			args:     []string{"thinktank", "instructions.txt", "./src"},
			expected: SimplifiedMode,
		},
		{
			name:     "SimplifiedMode_MarkdownInstructions",
			args:     []string{"instructions.md", "./target"},
			expected: SimplifiedMode,
		},
		{
			name:     "SimplifiedMode_WithFlags",
			args:     []string{"test.txt", "./src", "--dry-run", "--verbose"},
			expected: SimplifiedMode,
		},
		{
			name:     "ComplexMode_InstructionsFlag",
			args:     []string{"--instructions", "test.txt", "./src"},
			expected: ComplexMode,
		},
		{
			name:     "ComplexMode_ModelFlag",
			args:     []string{"--model", "gpt-4.1", "./src"},
			expected: ComplexMode,
		},
		{
			name:     "ComplexMode_NoFileExtension",
			args:     []string{"instructions", "./src"},
			expected: ComplexMode,
		},
		{
			name:     "ComplexMode_FirstArgIsFlag",
			args:     []string{"--dry-run", "instructions.txt"},
			expected: ComplexMode,
		},
		{
			name:     "ComplexMode_InsufficientArgs",
			args:     []string{"instructions.txt"},
			expected: ComplexMode,
		},
		{
			name:     "ComplexMode_EmptyArgs",
			args:     []string{},
			expected: ComplexMode,
		},
		{
			name:     "ComplexMode_OnlyFlags",
			args:     []string{"--help"},
			expected: ComplexMode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mode := router.detectParsingMode(tc.args)
			if mode != tc.expected {
				t.Errorf("detectParsingMode(%v) = %v, want %v", tc.args, mode, tc.expected)
			}
		})
	}
}

// TestParseArguments_SimplifiedMode tests parsing with simplified mode
func TestParseArguments_SimplifiedMode(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()
	instructionsFile := createTestFile(t, tempDir, "instructions.txt", "Test instructions")
	targetDir := createTestDir(t, tempDir, "src")

	logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)
	router := NewParserRouter(logger)

	args := []string{"thinktank", instructionsFile, targetDir, "--dry-run"}
	result := router.ParseArguments(args)

	// Verify parsing succeeded
	if !result.IsSuccess() {
		t.Fatalf("Expected successful parsing, got error: %v", result.Error)
	}

	// Verify mode detection
	if result.Mode != SimplifiedMode {
		t.Errorf("Expected SimplifiedMode, got %v", result.Mode)
	}

	// Verify config conversion
	if result.Config == nil {
		t.Fatal("Expected non-nil config")
	}

	if result.Config.InstructionsFile != instructionsFile {
		t.Errorf("Expected instructions file %s, got %s", instructionsFile, result.Config.InstructionsFile)
	}

	if len(result.Config.Paths) != 1 || result.Config.Paths[0] != targetDir {
		t.Errorf("Expected paths [%s], got %v", targetDir, result.Config.Paths)
	}

	if !result.Config.DryRun {
		t.Error("Expected DryRun to be true")
	}

	// Should not have deprecation warnings for simplified mode
	if result.HasDeprecationWarning() {
		t.Errorf("Unexpected deprecation warning for simplified mode: %v", result.Deprecation)
	}
}

// TestParseArguments_ComplexMode tests parsing with complex mode
func TestParseArguments_ComplexMode(t *testing.T) {
	logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)
	router := NewParserRouter(logger)

	args := []string{"--instructions", "test.txt", "./src"}
	result := router.ParseArguments(args)

	// Should detect complex mode
	if result.Mode != ComplexMode {
		t.Errorf("Expected ComplexMode, got %v", result.Mode)
	}

	// Should have an error since complex parsing isn't fully integrated yet
	if result.Error == nil {
		t.Error("Expected error for complex parsing (not yet integrated)")
	}

	// Should have deprecation warning
	if !result.HasDeprecationWarning() {
		t.Error("Expected deprecation warning for complex mode")
	}
}

// TestGenerateDeprecationWarning tests deprecation warning generation
func TestGenerateDeprecationWarning(t *testing.T) {
	router := NewParserRouter(nil)

	testCases := []struct {
		name          string
		args          []string
		expectWarning bool
		expectedFlag  string
		expectedMsg   string
	}{
		{
			name:          "InstructionsFlag",
			args:          []string{"--instructions", "test.txt"},
			expectWarning: true,
			expectedFlag:  "--instructions",
			expectedMsg:   "The --instructions flag is deprecated",
		},
		{
			name:          "InstructionsEquals",
			args:          []string{"--instructions=test.txt"},
			expectWarning: true,
			expectedFlag:  "--instructions=",
			expectedMsg:   "The --instructions=value format is deprecated",
		},
		{
			name:          "OutputDirWithoutPositional",
			args:          []string{"--output-dir", "output/"},
			expectWarning: true,
			expectedFlag:  "--output-dir",
			expectedMsg:   "Flag-only interface is deprecated",
		},
		{
			name:          "ComplexFlags",
			args:          []string{"--model", "gpt-4.1", "--exclude", "*.tmp"},
			expectWarning: true,
			expectedFlag:  "complex_flags",
			expectedMsg:   "Complex flag interface will be deprecated",
		},
		{
			name:          "SimplifiedMode_NoWarning",
			args:          []string{"instructions.txt", "./src", "--dry-run"},
			expectWarning: false,
		},
		{
			name:          "OutputDirWithPositional",
			args:          []string{"instructions.txt", "./src", "--output-dir", "output/"},
			expectWarning: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			warning := router.generateDeprecationWarning(tc.args)

			if tc.expectWarning {
				if warning == nil {
					t.Error("Expected deprecation warning, got nil")
					return
				}
				if warning.FlagUsed != tc.expectedFlag {
					t.Errorf("Expected flag %q, got %q", tc.expectedFlag, warning.FlagUsed)
				}
				if !strings.Contains(warning.Message, tc.expectedMsg) {
					t.Errorf("Expected message containing %q, got %q", tc.expectedMsg, warning.Message)
				}
				if warning.Suggestion == "" {
					t.Error("Expected non-empty suggestion")
				}
			} else {
				if warning != nil {
					t.Errorf("Expected no deprecation warning, got: %v", warning)
				}
			}
		})
	}
}

// TestContainsPositionalArgs tests the positional argument detection helper
func TestContainsPositionalArgs(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "HasPositional",
			args:     []string{"thinktank", "file.txt", "./src"},
			expected: true,
		},
		{
			name:     "HasPositionalAfterFlags",
			args:     []string{"--dry-run", "file.txt"},
			expected: true,
		},
		{
			name:     "OnlyFlags",
			args:     []string{"--dry-run", "--verbose"},
			expected: false,
		},
		{
			name:     "Empty",
			args:     []string{},
			expected: false,
		},
		{
			name:     "OnlyBinary",
			args:     []string{"thinktank"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := containsPositionalArgs(tc.args)
			if result != tc.expected {
				t.Errorf("containsPositionalArgs(%v) = %v, want %v", tc.args, result, tc.expected)
			}
		})
	}
}

// TestContainsComplexFlags tests the complex flag detection helper
func TestContainsComplexFlags(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "HasComplexFlag",
			args:     []string{"--instructions", "test.txt"},
			expected: true,
		},
		{
			name:     "HasComplexFlagEquals",
			args:     []string{"--model=gpt-4.1"},
			expected: true,
		},
		{
			name:     "HasMultipleComplexFlags",
			args:     []string{"--include", "*.go", "--exclude", "*.tmp"},
			expected: true,
		},
		{
			name:     "OnlySimpleFlags",
			args:     []string{"--dry-run", "--verbose"},
			expected: false,
		},
		{
			name:     "NoFlags",
			args:     []string{"instructions.txt", "./src"},
			expected: false,
		},
		{
			name:     "MixedFlags",
			args:     []string{"instructions.txt", "./src", "--model", "gpt-4.1"},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := containsComplexFlags(tc.args)
			if result != tc.expected {
				t.Errorf("containsComplexFlags(%v) = %v, want %v", tc.args, result, tc.expected)
			}
		})
	}
}

// TestParseResult_Methods tests the ParseResult helper methods
func TestParseResult_Methods(t *testing.T) {
	t.Run("IsSuccess_True", func(t *testing.T) {
		result := &ParseResult{
			Config: &config.CliConfig{},
			Error:  nil,
		}
		if !result.IsSuccess() {
			t.Error("Expected IsSuccess() to return true")
		}
	})

	t.Run("IsSuccess_False_NoConfig", func(t *testing.T) {
		result := &ParseResult{
			Config: nil,
			Error:  nil,
		}
		if result.IsSuccess() {
			t.Error("Expected IsSuccess() to return false when config is nil")
		}
	})

	t.Run("IsSuccess_False_WithError", func(t *testing.T) {
		result := &ParseResult{
			Config: &config.CliConfig{},
			Error:  fmt.Errorf("test error"),
		}
		if result.IsSuccess() {
			t.Error("Expected IsSuccess() to return false when error is present")
		}
	})

	t.Run("HasDeprecationWarning_True", func(t *testing.T) {
		result := &ParseResult{
			Deprecation: &DeprecationWarning{Message: "test"},
		}
		if !result.HasDeprecationWarning() {
			t.Error("Expected HasDeprecationWarning() to return true")
		}
	})

	t.Run("HasDeprecationWarning_False", func(t *testing.T) {
		result := &ParseResult{
			Deprecation: nil,
		}
		if result.HasDeprecationWarning() {
			t.Error("Expected HasDeprecationWarning() to return false")
		}
	})
}

// TestLogDeprecationWarning tests deprecation warning logging
func TestLogDeprecationWarning(t *testing.T) {
	// Create a mock logger to capture log calls
	logger := &MockLogger{}
	router := NewParserRouter(logger)

	warning := &DeprecationWarning{
		Message:    "Test deprecation",
		Suggestion: "Use new interface",
		FlagUsed:   "--old-flag",
	}

	router.LogDeprecationWarning(warning, nil)

	// Verify logger was called
	if len(logger.WarnCalls) != 1 {
		t.Errorf("Expected 1 warn call, got %d", len(logger.WarnCalls))
	}

	if len(logger.WarnCalls) > 0 {
		call := logger.WarnCalls[0]
		if !strings.Contains(call.Message, "Deprecation warning") {
			t.Errorf("Expected message to contain 'Deprecation warning', got %q", call.Message)
		}
	}

	// Test with nil warning
	logger.Reset()
	router.LogDeprecationWarning(nil, nil)
	if len(logger.WarnCalls) != 0 {
		t.Error("Expected no warn calls for nil warning")
	}
}

// Helper functions for testing

// createTestFile creates a temporary file with content for testing
func createTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, name)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", filePath, err)
	}
	return filePath
}

// createTestDir creates a temporary directory for testing
func createTestDir(t *testing.T, parent, name string) string {
	t.Helper()
	dirPath := filepath.Join(parent, name)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory %s: %v", dirPath, err)
	}
	return dirPath
}

// MockLogger implements LoggerInterface for testing
type MockLogger struct {
	DebugCalls []MockLogCall
	WarnCalls  []MockLogCall
}

type MockLogCall struct {
	Message string
	KeyVals []interface{}
}

func (m *MockLogger) Debug(msg string, keyVals ...interface{}) {
	m.DebugCalls = append(m.DebugCalls, MockLogCall{
		Message: msg,
		KeyVals: keyVals,
	})
}
func (m *MockLogger) Info(msg string, keyVals ...interface{}) {}
func (m *MockLogger) Warn(msg string, keyVals ...interface{}) {
	m.WarnCalls = append(m.WarnCalls, MockLogCall{
		Message: msg,
		KeyVals: keyVals,
	})
}
func (m *MockLogger) Error(msg string, keyVals ...interface{})  {}
func (m *MockLogger) Fatal(msg string, keyVals ...interface{})  {}
func (m *MockLogger) Printf(format string, args ...interface{}) {}
func (m *MockLogger) Println(args ...interface{})               {}

// Context methods to satisfy LoggerInterface
func (m *MockLogger) DebugContext(ctx context.Context, msg string, keyVals ...interface{}) {}
func (m *MockLogger) InfoContext(ctx context.Context, msg string, keyVals ...interface{})  {}
func (m *MockLogger) WarnContext(ctx context.Context, msg string, keyVals ...interface{})  {}
func (m *MockLogger) ErrorContext(ctx context.Context, msg string, keyVals ...interface{}) {}
func (m *MockLogger) FatalContext(ctx context.Context, msg string, keyVals ...interface{}) {}
func (m *MockLogger) WithContext(ctx context.Context) logutil.LoggerInterface              { return m }

func (m *MockLogger) Reset() {
	m.DebugCalls = nil
	m.WarnCalls = nil
}

// TestLogDeprecationWarning_Suppression tests warning suppression with environment variable
func TestLogDeprecationWarning_Suppression(t *testing.T) {
	// Save original environment and restore after test
	originalEnv := os.Getenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS")
	defer func() {
		if originalEnv == "" {
			_ = os.Unsetenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS")
		} else {
			_ = os.Setenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS", originalEnv)
		}
	}()

	// Create a mock logger to capture log calls
	logger := &MockLogger{}
	router := NewParserRouter(logger)

	warning := &DeprecationWarning{
		Message:    "Test deprecation",
		Suggestion: "Use new interface",
		FlagUsed:   "--old-flag",
	}

	t.Run("WithoutSuppression", func(t *testing.T) {
		// Ensure env var is not set
		_ = os.Unsetenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS")
		logger.Reset()

		router.LogDeprecationWarning(warning, nil)

		// Should log warning normally
		if len(logger.WarnCalls) != 1 {
			t.Errorf("Expected 1 warn call without suppression, got %d", len(logger.WarnCalls))
		}

		if len(logger.WarnCalls) > 0 {
			call := logger.WarnCalls[0]
			if !strings.Contains(call.Message, "Deprecation warning") {
				t.Errorf("Expected warn message to contain 'Deprecation warning', got %q", call.Message)
			}
		}
	})

	t.Run("WithSuppressionEnabled", func(t *testing.T) {
		// Set suppression environment variable
		_ = os.Setenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS", "true")
		logger.Reset()

		router.LogDeprecationWarning(warning, nil)

		// Should not log warning to Warn level
		if len(logger.WarnCalls) != 0 {
			t.Errorf("Expected 0 warn calls with suppression enabled, got %d", len(logger.WarnCalls))
		}

		// Should log to Debug level when suppressed
		if len(logger.DebugCalls) != 1 {
			t.Errorf("Expected 1 debug call with suppression enabled, got %d", len(logger.DebugCalls))
		}

		if len(logger.DebugCalls) > 0 {
			call := logger.DebugCalls[0]
			if !strings.Contains(call.Message, "Deprecation warning suppressed") {
				t.Errorf("Expected debug message to contain 'Deprecation warning suppressed', got %q", call.Message)
			}
		}
	})

	t.Run("WithSuppressionEmptyValue", func(t *testing.T) {
		// Set suppression environment variable to empty (should not suppress)
		_ = os.Setenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS", "")
		logger.Reset()

		router.LogDeprecationWarning(warning, nil)

		// Should log warning normally since empty value means not suppressed
		if len(logger.WarnCalls) != 1 {
			t.Errorf("Expected 1 warn call with empty suppression env var, got %d", len(logger.WarnCalls))
		}
	})

	t.Run("WithSuppressionAnyValue", func(t *testing.T) {
		// Test with various non-empty values - all should suppress
		testValues := []string{"1", "false", "off", "anything"}

		for _, value := range testValues {
			t.Run("Value_"+value, func(t *testing.T) {
				_ = os.Setenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS", value)
				logger.Reset()

				router.LogDeprecationWarning(warning, nil)

				// Should suppress regardless of specific value, as long as it's non-empty
				if len(logger.WarnCalls) != 0 {
					t.Errorf("Expected 0 warn calls with suppression value %q, got %d", value, len(logger.WarnCalls))
				}
			})
		}
	})
}
