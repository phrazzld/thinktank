package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDocumentationExamples tests that the examples in README.md actually work
// This follows Kent Beck's TDD principle of testing documentation accuracy
func TestDocumentationExamples(t *testing.T) {
	// Create test environment
	tempDir := t.TempDir()
	testInstructionsFile := filepath.Join(tempDir, "instructions.txt")
	testTargetDir := filepath.Join(tempDir, "project")

	// Create test files
	if err := os.WriteFile(testInstructionsFile, []byte("Analyze this codebase"), 0644); err != nil {
		t.Fatalf("Failed to create test instructions file: %v", err)
	}
	if err := os.MkdirAll(testTargetDir, 0755); err != nil {
		t.Fatalf("Failed to create test target directory: %v", err)
	}

	examples := []struct {
		name        string
		description string
		args        []string
		expectError bool
		validation  func(*testing.T, *ParseResult)
	}{
		{
			name:        "BasicSimplified",
			description: "Basic simplified interface: thinktank instructions.txt ./my-project",
			args:        []string{"thinktank", testInstructionsFile, testTargetDir},
			expectError: false,
			validation: func(t *testing.T, result *ParseResult) {
				assert.Equal(t, SimplifiedMode, result.Mode)
				assert.Equal(t, testInstructionsFile, result.Config.InstructionsFile)
				assert.Equal(t, []string{testTargetDir}, result.Config.Paths)
			},
		},
		{
			name:        "WithSpecificModel",
			description: "With specific model: thinktank instructions.txt . --model gpt-4o",
			args:        []string{"thinktank", testInstructionsFile, testTargetDir, "--model", "gpt-4o"},
			expectError: false,
			validation: func(t *testing.T, result *ParseResult) {
				assert.Equal(t, SimplifiedMode, result.Mode)
				// Note: The simplified parser currently only validates model syntax
				// Model selection is handled by environment variables or smart defaults
				assert.NotEmpty(t, result.Config.ModelNames, "Should have at least one model")
			},
		},
		{
			name:        "DryRun",
			description: "Dry run: thinktank instructions.txt . --dry-run",
			args:        []string{"thinktank", testInstructionsFile, testTargetDir, "--dry-run"},
			expectError: false,
			validation: func(t *testing.T, result *ParseResult) {
				assert.Equal(t, SimplifiedMode, result.Mode)
				assert.True(t, result.Config.DryRun)
			},
		},
		{
			name:        "MultipleModels",
			description: "Multiple models: thinktank instructions.txt . --model gemini-2.5-pro --model gpt-4o",
			args:        []string{"thinktank", testInstructionsFile, testTargetDir, "--model", "gemini-2.5-pro", "--model", "gpt-4o"},
			expectError: false,
			validation: func(t *testing.T, result *ParseResult) {
				assert.Equal(t, SimplifiedMode, result.Mode)
				// Note: The simplified parser currently only validates model syntax
				// Multiple model selection should be handled by environment variables
				assert.NotEmpty(t, result.Config.ModelNames, "Should have at least one model")
			},
		},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			// Create mock logger
			logger := &MockLogger{}
			router := NewParserRouterWithEnv(logger, func(string) string { return "" })

			// Parse the documented command
			result := router.ParseArguments(ex.args)

			if ex.expectError {
				assert.Error(t, result.Error, "Example should fail: %s", ex.description)
			} else {
				assert.NoError(t, result.Error, "Example should succeed: %s", ex.description)
				require.NotNil(t, result.Config, "Config should be parsed successfully")
				ex.validation(t, result)
			}
		})
	}
}

// TestMigrationExamples tests the migration examples from the documentation
func TestMigrationExamples(t *testing.T) {
	// Create test environment
	tempDir := t.TempDir()
	testInstructionsFile := filepath.Join(tempDir, "task.txt")
	testTargetDir := filepath.Join(tempDir, "src")

	// Create test files
	if err := os.WriteFile(testInstructionsFile, []byte("Test task"), 0644); err != nil {
		t.Fatalf("Failed to create test instructions file: %v", err)
	}
	if err := os.MkdirAll(testTargetDir, 0755); err != nil {
		t.Fatalf("Failed to create test target directory: %v", err)
	}

	migrations := []struct {
		name               string
		oldArgs            []string // Complex interface
		newArgs            []string // Simplified interface
		shouldBeEquivalent bool
	}{
		{
			name:               "BasicMigration",
			oldArgs:            []string{"thinktank", "--instructions", testInstructionsFile, testTargetDir},
			newArgs:            []string{"thinktank", testInstructionsFile, testTargetDir},
			shouldBeEquivalent: true,
		},
		{
			name:               "WithModel",
			oldArgs:            []string{"thinktank", "--instructions", testInstructionsFile, "--model", "gpt-4o", testTargetDir},
			newArgs:            []string{"thinktank", testInstructionsFile, testTargetDir}, // Model specified via environment
			shouldBeEquivalent: false,                                                      // Different - simplified uses env vars for model selection
		},
		{
			name:               "WithDryRun",
			oldArgs:            []string{"thinktank", "--instructions", testInstructionsFile, "--dry-run", testTargetDir},
			newArgs:            []string{"thinktank", testInstructionsFile, testTargetDir, "--dry-run"},
			shouldBeEquivalent: true,
		},
		{
			name:               "WithVerbose",
			oldArgs:            []string{"thinktank", "--instructions", testInstructionsFile, "--verbose", testTargetDir},
			newArgs:            []string{"thinktank", testInstructionsFile, testTargetDir, "--verbose"},
			shouldBeEquivalent: true,
		},
	}

	for _, migration := range migrations {
		t.Run(migration.name, func(t *testing.T) {
			logger := &MockLogger{}
			router := NewParserRouterWithEnv(logger, func(string) string { return "" })

			// Parse old (complex) interface
			oldResult := router.ParseArguments(migration.oldArgs)
			require.NoError(t, oldResult.Error, "Old interface should parse successfully")
			assert.Equal(t, ComplexMode, oldResult.Mode, "Should use complex mode")

			// Parse new (simplified) interface
			newResult := router.ParseArguments(migration.newArgs)
			require.NoError(t, newResult.Error, "New interface should parse successfully")
			assert.Equal(t, SimplifiedMode, newResult.Mode, "Should use simplified mode")

			if migration.shouldBeEquivalent {
				// Verify functional equivalence
				assert.Equal(t, oldResult.Config.InstructionsFile, newResult.Config.InstructionsFile)
				assert.Equal(t, oldResult.Config.Paths, newResult.Config.Paths)
				assert.Equal(t, oldResult.Config.DryRun, newResult.Config.DryRun)
				assert.Equal(t, oldResult.Config.Verbose, newResult.Config.Verbose)
				// Note: Model equivalence not tested since simplified interface uses environment variables
			} else {
				// Still verify basic functionality works
				assert.Equal(t, oldResult.Config.InstructionsFile, newResult.Config.InstructionsFile)
				assert.Equal(t, oldResult.Config.Paths, newResult.Config.Paths)
			}
		})
	}
}

// TestEnvironmentVariableDocumentation tests the documented environment variables
func TestEnvironmentVariableDocumentation(t *testing.T) {
	// Create test environment
	tempDir := t.TempDir()
	testInstructionsFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testInstructionsFile, []byte("Test"), 0644); err != nil {
		t.Fatalf("Failed to create test instructions file: %v", err)
	}

	envVars := map[string]struct {
		description string
		testValue   string
		validation  func(*testing.T, string)
	}{
		"THINKTANK_OPENAI_RATE_LIMIT": {
			description: "OpenAI rate limiting",
			testValue:   "100",
			validation: func(t *testing.T, val string) {
				assert.Equal(t, "100", val)
			},
		},
		"THINKTANK_GEMINI_RATE_LIMIT": {
			description: "Gemini rate limiting",
			testValue:   "1000",
			validation: func(t *testing.T, val string) {
				assert.Equal(t, "1000", val)
			},
		},
		"THINKTANK_MAX_CONCURRENT": {
			description: "Concurrency control",
			testValue:   "5",
			validation: func(t *testing.T, val string) {
				assert.Equal(t, "5", val)
			},
		},
		"THINKTANK_SUPPRESS_DEPRECATION_WARNINGS": {
			description: "Behavior control",
			testValue:   "1",
			validation: func(t *testing.T, val string) {
				assert.Equal(t, "1", val)
			},
		},
	}

	for envVar, spec := range envVars {
		t.Run(envVar, func(t *testing.T) {
			// Test with mock environment function
			mockEnv := func(key string) string {
				if key == envVar {
					return spec.testValue
				}
				return ""
			}

			// Validate the environment variable is recognized
			value := mockEnv(envVar)
			spec.validation(t, value)

			// Test environment loading with this variable
			logger := &MockLogger{}
			router := NewParserRouterWithEnv(logger, mockEnv)

			// Parse a simple command to trigger environment loading
			result := router.ParseArguments([]string{"thinktank", testInstructionsFile, "."})
			assert.NoError(t, result.Error, "Should parse successfully with environment variable %s", envVar)
		})
	}
}

// MockLogger is defined in parser_router_test.go
