package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestComplexParserIntegration tests that the parser router can actually
// handle complex flag parsing instead of returning stub errors
func TestComplexParserIntegration(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()
	instructionsFile := createTestFile(t, tempDir, "instructions.txt", "Test instructions")
	targetDir := createTestDir(t, tempDir, "src")

	testCases := []struct {
		name          string
		args          []string
		expectSuccess bool
		expectWarning bool
		expectedFlag  string
	}{
		{
			name:          "ComplexModeInstructionsFlag",
			args:          []string{"thinktank", "--instructions", instructionsFile, targetDir},
			expectSuccess: true,
			expectWarning: true,
			expectedFlag:  "--instructions",
		},
		{
			name:          "ComplexModeModelFlag",
			args:          []string{"thinktank", "--model", "gpt-4.1", "--instructions", instructionsFile, targetDir},
			expectSuccess: true,
			expectWarning: true,
			expectedFlag:  "--instructions", // Should still warn about instructions
		},
		{
			name:          "ComplexModeOutputDir",
			args:          []string{"thinktank", "--instructions", instructionsFile, "--output-dir", "output", targetDir},
			expectSuccess: true,
			expectWarning: true,
			expectedFlag:  "--instructions",
		},
		{
			name:          "SimplifiedMode",
			args:          []string{"thinktank", instructionsFile, targetDir},
			expectSuccess: true,
			expectWarning: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)
			router := NewParserRouter(logger)

			result := router.ParseArguments(tc.args)

			// Check success expectation
			if tc.expectSuccess && !result.IsSuccess() {
				t.Errorf("Expected successful parsing, got error: %v", result.Error)
			}
			if !tc.expectSuccess && result.IsSuccess() {
				t.Error("Expected parsing failure, but got success")
			}

			// Check warning expectation
			if tc.expectWarning && !result.HasDeprecationWarning() {
				t.Error("Expected deprecation warning, but got none")
			}
			if !tc.expectWarning && result.HasDeprecationWarning() {
				t.Errorf("Unexpected deprecation warning: %v", result.Deprecation)
			}

			// Check specific flag being warned about
			if tc.expectWarning && result.HasDeprecationWarning() {
				if result.Deprecation.FlagUsed != tc.expectedFlag {
					t.Errorf("Expected warning about flag %q, got %q", tc.expectedFlag, result.Deprecation.FlagUsed)
				}
			}

			// Check that complex mode actually creates valid config
			if result.IsSuccess() && result.Config != nil {
				// Verify essential fields are populated
				if result.Mode == ComplexMode {
					if result.Config.InstructionsFile != instructionsFile {
						t.Errorf("Expected instructions file %s, got %s", instructionsFile, result.Config.InstructionsFile)
					}
					if len(result.Config.Paths) == 0 {
						t.Error("Expected paths to be populated from complex parsing")
					}
				}
			}
		})
	}
}

// TestComplexParsingActuallyWorks tests that complex flag parsing produces
// the same result whether we go through the router or call flags.go directly
func TestComplexParsingActuallyWorks(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()
	instructionsFile := createTestFile(t, tempDir, "instructions.txt", "Test instructions")
	targetDir := createTestDir(t, tempDir, "src")

	args := []string{"thinktank", "--instructions", instructionsFile, "--model", "gpt-4.1", "--dry-run", targetDir}

	// Parse using direct flags.go method
	directConfig, directErr := ParseFlagsWithArgsAndEnv(args, os.Getenv)

	// Parse using router (should delegate to complex parser)
	logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)
	router := NewParserRouter(logger)
	routerResult := router.ParseArguments(args)

	// Both should succeed
	if directErr != nil {
		t.Fatalf("Direct parsing failed: %v", directErr)
	}
	if !routerResult.IsSuccess() {
		t.Fatalf("Router parsing failed: %v", routerResult.Error)
	}

	// Results should be equivalent
	if directConfig.InstructionsFile != routerResult.Config.InstructionsFile {
		t.Errorf("Instructions file mismatch: direct=%s, router=%s",
			directConfig.InstructionsFile, routerResult.Config.InstructionsFile)
	}

	if directConfig.DryRun != routerResult.Config.DryRun {
		t.Errorf("DryRun flag mismatch: direct=%v, router=%v",
			directConfig.DryRun, routerResult.Config.DryRun)
	}

	if len(directConfig.ModelNames) != len(routerResult.Config.ModelNames) {
		t.Errorf("Model names count mismatch: direct=%d, router=%d",
			len(directConfig.ModelNames), len(routerResult.Config.ModelNames))
	}

	// Router should also produce deprecation warning for complex mode
	if !routerResult.HasDeprecationWarning() {
		t.Error("Expected deprecation warning for complex flag usage")
	}
}

// TestMainCliFlowIntegration tests that the main CLI flow actually uses
// the deprecation warnings system
func TestMainCliFlowIntegration(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()
	instructionsFile := createTestFile(t, tempDir, "instructions.txt", "Test instructions")
	targetDir := createTestDir(t, tempDir, "src")

	// Override os.Args for testing
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"thinktank", "--instructions", instructionsFile, "--dry-run", targetDir}

	// This should trigger deprecation warnings in the actual CLI flow
	config, err := ParseFlags()
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}

	// Verify the config is populated correctly
	if config.InstructionsFile != instructionsFile {
		t.Errorf("Expected instructions file %s, got %s", instructionsFile, config.InstructionsFile)
	}

	if !config.DryRun {
		t.Error("Expected DryRun to be true")
	}

	// Test that we can parse with simplified mode too
	os.Args = []string{"thinktank", instructionsFile, targetDir, "--dry-run"}

	configSimplified, err := ParseFlags()
	if err != nil {
		t.Fatalf("ParseFlags with simplified mode failed: %v", err)
	}

	// Should get equivalent config
	if configSimplified.InstructionsFile != instructionsFile {
		t.Errorf("Expected simplified instructions file %s, got %s", instructionsFile, configSimplified.InstructionsFile)
	}

	if !configSimplified.DryRun {
		t.Error("Expected simplified DryRun to be true")
	}
}

// TestDeprecationWarningContent tests the specific content and quality of
// deprecation warnings for each major flag
func TestDeprecationWarningContent(t *testing.T) {
	logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)
	router := NewParserRouter(logger)

	testCases := []struct {
		name             string
		args             []string
		expectedFlag     string
		expectedMessage  string
		expectSuggestion bool
	}{
		{
			name:             "InstructionsFlag",
			args:             []string{"--instructions", "test.txt"},
			expectedFlag:     "--instructions",
			expectedMessage:  "deprecated",
			expectSuggestion: true,
		},
		{
			name:             "ModelFlag",
			args:             []string{"--model", "gpt-4.1", "file.txt"},
			expectedFlag:     "complex_flags",
			expectedMessage:  "deprecated",
			expectSuggestion: true,
		},
		{
			name:             "OutputDirFlag",
			args:             []string{"--output-dir", "output/"},
			expectedFlag:     "--output-dir",
			expectedMessage:  "deprecated",
			expectSuggestion: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			warning := router.generateDeprecationWarning(tc.args)

			if warning == nil {
				t.Fatal("Expected deprecation warning, got nil")
			}

			if warning.FlagUsed != tc.expectedFlag {
				t.Errorf("Expected flag %q, got %q", tc.expectedFlag, warning.FlagUsed)
			}

			if !strings.Contains(strings.ToLower(warning.Message), tc.expectedMessage) {
				t.Errorf("Expected message to contain %q, got %q", tc.expectedMessage, warning.Message)
			}

			if tc.expectSuggestion && warning.Suggestion == "" {
				t.Error("Expected suggestion, got empty string")
			}
		})
	}
}

// TestDeprecationTelemetryCollection tests that we can track which
// deprecated flags are being used most frequently
func TestDeprecationTelemetryCollection(t *testing.T) {
	// Create temporary files for testing
	tempDir := t.TempDir()
	instructionsFile := createTestFile(t, tempDir, "instructions.txt", "Test instructions")
	targetDir := createTestDir(t, tempDir, "src")

	// Create router with telemetry enabled
	router := NewParserRouterWithTelemetry(nil, true)

	// Simulate various usage patterns
	testCases := [][]string{
		{"thinktank", "--instructions", instructionsFile, targetDir},
		{"thinktank", "--instructions", instructionsFile, "--model", "gpt-4", targetDir},
		{"thinktank", "--instructions", instructionsFile, "--output-dir", "output", targetDir},
		{"thinktank", "--model", "gpt-4", "--instructions", instructionsFile, targetDir},
		{"thinktank", "--instructions", instructionsFile, targetDir}, // Duplicate pattern
	}

	// Process all test cases
	for _, args := range testCases {
		result := router.ParseArguments(args)
		if !result.IsSuccess() {
			t.Errorf("Parsing failed for args %v: %v", args, result.Error)
		}
	}

	// Get telemetry data
	telemetry := router.GetTelemetry()
	if telemetry == nil {
		t.Fatal("Expected telemetry to be available")
	}

	// Check usage statistics
	stats := telemetry.GetUsageStats()

	// Should have recorded complex flag usage
	if stats["complex_flags"] != 5 {
		t.Errorf("Expected 5 complex_flags usage records, got %d", stats["complex_flags"])
	}

	// Should have recorded instructions flag usage
	if stats["--instructions"] != 5 {
		t.Errorf("Expected 5 --instructions usage records, got %d", stats["--instructions"])
	}

	// Should have recorded model flag usage
	if stats["--model"] != 2 {
		t.Errorf("Expected 2 --model usage records, got %d", stats["--model"])
	}

	// Test most common flags functionality
	mostCommon := telemetry.GetMostCommonFlags(3)
	if len(mostCommon) < 2 {
		t.Errorf("Expected at least 2 most common flags, got %d", len(mostCommon))
	}

	// The most common should be --instructions and complex_flags
	found_instructions := false
	found_complex := false
	for _, flag := range mostCommon {
		if flag.Flag == "--instructions" && flag.Count == 5 {
			found_instructions = true
		}
		if flag.Flag == "complex_flags" && flag.Count == 5 {
			found_complex = true
		}
	}

	if !found_instructions {
		t.Error("Expected --instructions to be in most common flags with count 5")
	}
	if !found_complex {
		t.Error("Expected complex_flags to be in most common flags with count 5")
	}
}
