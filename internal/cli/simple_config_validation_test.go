// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePerformance(t *testing.T) {
	// Create test environment
	tempDir := t.TempDir()
	instFile := filepath.Join(tempDir, "instructions.md")
	require.NoError(t, os.WriteFile(instFile, []byte("test"), 0644))
	targetDir := filepath.Join(tempDir, "src")
	require.NoError(t, os.Mkdir(targetDir, 0755))

	// Set up environment
	_ = os.Setenv("OPENROUTER_API_KEY", "test-key")
	defer func() { _ = os.Unsetenv("OPENROUTER_API_KEY") }()

	config := SimplifiedConfig{
		InstructionsFile: instFile,
		TargetPath:       targetDir,
		Flags:            0x00,
	}

	// Warm up to avoid cold start effects
	_ = config.Validate()

	// Measure validation time
	start := time.Now()
	iterations := 1000
	for i := 0; i < iterations; i++ {
		err := config.Validate()
		require.NoError(t, err)
	}
	elapsed := time.Since(start)

	avgTime := elapsed / time.Duration(iterations)
	t.Logf("Average validation time: %v", avgTime)

	// Assert sub-millisecond performance
	assert.Less(t, avgTime, time.Millisecond,
		"Validation should complete in <1ms, got %v", avgTime)
}

// TestValidatePathLength tests path length validation (defensive programming)
func TestValidatePathLength(t *testing.T) {
	tests := []struct {
		name         string
		config       SimplifiedConfig
		wantErr      bool
		errorMessage string
	}{
		{
			name: "normal path length should pass",
			config: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "src/",
				Flags:            FlagDryRun, // Skip API key validation
			},
			wantErr: false,
		},
		{
			name: "very long target path should fail",
			config: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       strings.Repeat("a", 256), // Over 255 limit
				Flags:            FlagDryRun,
			},
			wantErr:      true,
			errorMessage: "path too long",
		},
		{
			name: "very long instructions path should fail",
			config: SimplifiedConfig{
				InstructionsFile: strings.Repeat("b", 253) + ".md", // 256 chars total, over limit
				TargetPath:       "src/",
				Flags:            FlagDryRun,
			},
			wantErr:      true,
			errorMessage: "path too long",
		},
		{
			name: "maximum valid path length should pass",
			config: SimplifiedConfig{
				InstructionsFile: strings.Repeat("c", 252) + ".md", // 255 total chars
				TargetPath:       strings.Repeat("d", 255),         // Exactly at limit
				Flags:            FlagDryRun,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				// For tests that should pass but might fail due to file not existing,
				// we accept either success or "does not exist" errors
				if err != nil {
					// Allow "does not exist" errors for very long valid paths
					// since we're testing length validation, not file existence
					assert.Contains(t, err.Error(), "does not exist",
						"Expected either success or 'does not exist' error, got: %v", err)
				}
			}
		})
	}
}

// TestValidateAPIKeyHelper tests the validateAPIKeyForModel helper function
func TestValidateAPIKeyHelper(t *testing.T) {
	// Save and restore environment
	oldGemini := os.Getenv("GEMINI_API_KEY")
	oldOpenAI := os.Getenv("OPENAI_API_KEY")
	defer func() {
		_ = os.Setenv("GEMINI_API_KEY", oldGemini)
		_ = os.Setenv("OPENAI_API_KEY", oldOpenAI)
	}()

	tests := []struct {
		name      string
		modelName string
		setup     func()
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid gemini model with key",
			modelName: "gemini-2.5-pro",
			setup: func() {
				_ = os.Setenv("GEMINI_API_KEY", "test-key")
			},
			wantErr: false,
		},
		{
			name:      "valid openai model with key",
			modelName: "gpt-4.1",
			setup: func() {
				_ = os.Setenv("OPENAI_API_KEY", "test-key")
			},
			wantErr: false,
		},
		{
			name:      "missing API key",
			modelName: "gemini-2.5-pro",
			setup: func() {
				_ = os.Unsetenv("GEMINI_API_KEY")
			},
			wantErr: true,
			errMsg:  "API key not set: please set OPENROUTER_API_KEY",
		},
		{
			name:      "unknown model",
			modelName: "invalid-model-xyz",
			wantErr:   true,
			errMsg:    "unknown model invalid-model-xyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			err := validateAPIKeyForModel(tt.modelName)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateAPIKeysForModels tests the batch API key validation helper
func TestValidateAPIKeysForModels(t *testing.T) {
	// Save and restore environment
	oldGemini := os.Getenv("GEMINI_API_KEY")
	oldOpenAI := os.Getenv("OPENAI_API_KEY")
	oldOpenRouter := os.Getenv("OPENROUTER_API_KEY")
	defer func() {
		_ = os.Setenv("GEMINI_API_KEY", oldGemini)
		_ = os.Setenv("OPENAI_API_KEY", oldOpenAI)
		_ = os.Setenv("OPENROUTER_API_KEY", oldOpenRouter)
	}()

	tests := []struct {
		name       string
		modelNames []string
		setup      func()
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "single model with key",
			modelNames: []string{"gemini-2.5-pro"},
			setup: func() {
				_ = os.Setenv("GEMINI_API_KEY", "test-key")
			},
			wantErr: false,
		},
		{
			name:       "multiple models, same provider",
			modelNames: []string{"gemini-2.5-pro", "gemini-2.5-flash"},
			setup: func() {
				_ = os.Setenv("GEMINI_API_KEY", "test-key")
			},
			wantErr: false,
		},
		{
			name:       "multiple models, different providers",
			modelNames: []string{"gemini-2.5-pro", "gpt-4.1"},
			setup: func() {
				_ = os.Setenv("GEMINI_API_KEY", "test-key")
				_ = os.Setenv("OPENAI_API_KEY", "test-key")
			},
			wantErr: false,
		},
		{
			name:       "synthesis mode models (typical use case)",
			modelNames: []string{"gemini-2.5-pro", "gpt-4.1"},
			setup: func() {
				_ = os.Setenv("GEMINI_API_KEY", "test-key")
				_ = os.Setenv("OPENAI_API_KEY", "test-key")
			},
			wantErr: false,
		},
		{
			name:       "missing key for first provider",
			modelNames: []string{"gemini-2.5-pro", "gpt-4.1"},
			setup: func() {
				_ = os.Unsetenv("GEMINI_API_KEY")
				_ = os.Setenv("OPENAI_API_KEY", "test-key")
			},
			wantErr: true,
			errMsg:  "API key not set: please set OPENROUTER_API_KEY",
		},
		{
			name:       "missing key for second provider",
			modelNames: []string{"gemini-2.5-pro", "gpt-4.1"},
			setup: func() {
				_ = os.Setenv("GEMINI_API_KEY", "test-key")
				_ = os.Unsetenv("OPENAI_API_KEY")
			},
			wantErr: true,
			errMsg:  "API key not set: please set OPENROUTER_API_KEY",
		},
		{
			name:       "unknown model in list",
			modelNames: []string{"gemini-2.5-pro", "invalid-model"},
			setup: func() {
				_ = os.Setenv("GEMINI_API_KEY", "test-key")
			},
			wantErr: true,
			errMsg:  "unknown model invalid-model",
		},
		{
			name:       "empty model list",
			modelNames: []string{},
			wantErr:    false, // Should succeed with no models to check
		},
		{
			name:       "duplicate providers optimization test",
			modelNames: []string{"gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.5-pro"},
			setup: func() {
				_ = os.Setenv("GEMINI_API_KEY", "test-key")
			},
			wantErr: false, // Should only check GEMINI_API_KEY once
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			err := validateAPIKeysForModels(tt.modelNames)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateAPIKeysPerformance measures batch validation performance
func TestValidateAPIKeysPerformance(t *testing.T) {
	// Set up environment
	_ = os.Setenv("GEMINI_API_KEY", "test-key")
	_ = os.Setenv("OPENAI_API_KEY", "test-key")
	_ = os.Setenv("OPENROUTER_API_KEY", "test-key")
	defer func() {
		_ = os.Unsetenv("GEMINI_API_KEY")
		_ = os.Unsetenv("OPENAI_API_KEY")
		_ = os.Unsetenv("OPENROUTER_API_KEY")
	}()

	// Test with synthesis mode models (common case)
	models := []string{"gemini-2.5-pro", "gpt-4.1"}

	// Warm up
	_ = validateAPIKeysForModels(models)

	// Measure performance
	start := time.Now()
	iterations := 10000
	for i := 0; i < iterations; i++ {
		err := validateAPIKeysForModels(models)
		require.NoError(t, err)
	}
	elapsed := time.Since(start)

	avgTime := elapsed / time.Duration(iterations)
	t.Logf("Average batch API key validation time: %v", avgTime)

	// Should be very fast - mostly map lookups and env var access
	assert.Less(t, avgTime, 10*time.Microsecond,
		"Batch API key validation should be <10μs, got %v", avgTime)
}

// TestParseOptionalFlags tests the enhanced flag parsing with abbreviations and validation.
// Following TDD - RED phase: this test will fail until we implement parseOptionalFlags().
func TestParseOptionalFlags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		errorMsg      string
		expectedFlags uint8
	}{
		{
			name:          "short model flag",
			args:          []string{"-m", "gemini-2.5-pro", "instructions.md", "src/"},
			expectedFlags: 0,
		},
		{
			name:          "short verbose flag",
			args:          []string{"-v", "instructions.md", "src/"},
			expectedFlags: FlagVerbose,
		},
		{
			name:          "mixed short and long flags",
			args:          []string{"-v", "--dry-run", "-m", "gpt-4", "instructions.md", "src/"},
			expectedFlags: FlagVerbose | FlagDryRun,
		},
		{
			name:        "unknown short flag",
			args:        []string{"-x", "instructions.md", "src/"},
			expectError: true,
			errorMsg:    "unknown flag: -x",
		},
		{
			name:        "short model without value",
			args:        []string{"-m"},
			expectError: true,
			errorMsg:    "flag needs an argument: -m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseSimplifiedArgs(tt.args)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedFlags, config.Flags)
				assert.Equal(t, "instructions.md", config.InstructionsFile)
				assert.Equal(t, "src/", config.TargetPath)
			}
		})
	}
}

// TestParseOptionalFlags_Comprehensive tests all flag formats and edge cases
func TestParseOptionalFlags_Comprehensive(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		errorMsg      string
		expectedFlags uint8
	}{
		{
			name:          "all long flags",
			args:          []string{"--verbose", "--dry-run", "--synthesis", "instructions.md", "src/"},
			expectedFlags: FlagVerbose | FlagDryRun | FlagSynthesis,
		},
		{
			name:          "mixed short and long",
			args:          []string{"-v", "--synthesis", "-m", "claude-3", "instructions.md", "src/"},
			expectedFlags: FlagVerbose | FlagSynthesis,
		},
		{
			name:          "flags with equals syntax",
			args:          []string{"--model=gpt-4", "--output-dir=./out", "-v", "instructions.md", "src/"},
			expectedFlags: FlagVerbose,
		},
		{
			name:          "flags before and after positional args",
			args:          []string{"-v", "instructions.md", "--dry-run", "src/", "--synthesis"},
			expectedFlags: FlagVerbose | FlagDryRun | FlagSynthesis,
		},
		{
			name:          "no flags, just positional",
			args:          []string{"instructions.md", "src/"},
			expectedFlags: 0,
		},
		{
			name:        "unknown short flag",
			args:        []string{"-z", "instructions.md", "src/"},
			expectError: true,
			errorMsg:    "unknown flag: -z",
		},
		{
			name:        "invalid single dash format",
			args:        []string{"-abc", "instructions.md", "src/"},
			expectError: true,
			errorMsg:    "unknown flag: -abc",
		},
		{
			name:        "empty model value with equals",
			args:        []string{"--model=", "instructions.md", "src/"},
			expectError: true,
			errorMsg:    "flag has empty value",
		},
		{
			name:        "empty output-dir value with equals",
			args:        []string{"--output-dir=", "instructions.md", "src/"},
			expectError: true,
			errorMsg:    "flag has empty value",
		},
		{
			name:          "model and output-dir together",
			args:          []string{"--model", "claude-3", "--output-dir", "./tmp", "instructions.md", "src/"},
			expectedFlags: 0,
		},
		{
			name:        "only flags, no positional args",
			args:        []string{"-v", "--dry-run"},
			expectError: true,
			errorMsg:    "insufficient arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseSimplifiedArgs(tt.args)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedFlags, config.Flags)
				assert.Equal(t, "instructions.md", config.InstructionsFile)
				assert.Equal(t, "src/", config.TargetPath)
			}
		})
	}
}

// BenchmarkParseOptionalFlags measures parsing performance to ensure <100μs target
func BenchmarkParseOptionalFlags(b *testing.B) {
	// Typical CLI arguments with mixed flags
	args := []string{"-v", "--model", "gemini-2.5-pro", "--output-dir", "./out",
		"--dry-run", "instructions.md", "src/"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := &SimplifiedConfig{}
		_, err := parseOptionalFlags(args, config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestParseOptionalFlagsPerformance validates sub-100μs parsing requirement
func TestParseOptionalFlagsPerformance(t *testing.T) {
	args := []string{"-v", "-m", "gpt-4", "--synthesis", "--dry-run", "test.md", "src/"}

	config := &SimplifiedConfig{}
	// Warm up
	_, _ = parseOptionalFlags(args, config)

	// Measure performance
	start := time.Now()
	iterations := 1000
	for i := 0; i < iterations; i++ {
		config := &SimplifiedConfig{}
		_, err := parseOptionalFlags(args, config)
		require.NoError(t, err)
	}
	elapsed := time.Since(start)

	avgTime := elapsed / time.Duration(iterations)
	t.Logf("Average flag parsing time: %v", avgTime)

	// Assert performance target: <100μs as specified in TODO requirements
	assert.Less(t, avgTime, 100*time.Microsecond,
		"Flag parsing should complete in <100μs, got %v", avgTime)
}

// TestValidatePositionalArgs tests the new positional argument validation function.
// This is our first failing test following TDD - RED phase.
func TestValidatePositionalArgs(t *testing.T) {
	t.Run("valid txt file and directory", func(t *testing.T) {
		tempDir := t.TempDir()
		// Create test files
		instructionsFile := filepath.Join(tempDir, "test.txt")
		targetDir := filepath.Join(tempDir, "src")

		err := os.WriteFile(instructionsFile, []byte("test instructions"), 0644)
		require.NoError(t, err)
		err = os.Mkdir(targetDir, 0755)
		require.NoError(t, err)

		// This should pass - basic case
		err = validatePositionalArgs(instructionsFile, targetDir)
		assert.NoError(t, err)
	})

	t.Run("valid md file and go file", func(t *testing.T) {
		tempDir := t.TempDir()
		// Create test files
		instructionsFile := filepath.Join(tempDir, "test.md")
		targetFile := filepath.Join(tempDir, "main.go")

		err := os.WriteFile(instructionsFile, []byte("# Test instructions"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(targetFile, []byte("package main"), 0644)
		require.NoError(t, err)

		// This should pass - md file and regular file
		err = validatePositionalArgs(instructionsFile, targetFile)
		assert.NoError(t, err)
	})

	t.Run("missing target path", func(t *testing.T) {
		tempDir := t.TempDir()
		instructionsFile := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(instructionsFile, []byte("test"), 0644)
		require.NoError(t, err)

		// Empty target path should fail
		err = validatePositionalArgs(instructionsFile, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target path required")
	})

	t.Run("missing instructions file", func(t *testing.T) {
		tempDir := t.TempDir()
		targetDir := filepath.Join(tempDir, "src")
		err := os.Mkdir(targetDir, 0755)
		require.NoError(t, err)

		// Empty instructions file should fail
		err = validatePositionalArgs("", targetDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "instructions file required")
	})

	t.Run("invalid instructions file extension", func(t *testing.T) {
		tempDir := t.TempDir()
		// Create file with invalid extension
		instructionsFile := filepath.Join(tempDir, "test.py")
		targetDir := filepath.Join(tempDir, "src")

		err := os.WriteFile(instructionsFile, []byte("test"), 0644)
		require.NoError(t, err)
		err = os.Mkdir(targetDir, 0755)
		require.NoError(t, err)

		// Should fail due to invalid extension
		err = validatePositionalArgs(instructionsFile, targetDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported instructions file extension")
	})

	t.Run("missing instructions file extension", func(t *testing.T) {
		tempDir := t.TempDir()
		// Create file with no extension
		instructionsFile := filepath.Join(tempDir, "test")
		targetDir := filepath.Join(tempDir, "src")

		err := os.WriteFile(instructionsFile, []byte("test"), 0644)
		require.NoError(t, err)
		err = os.Mkdir(targetDir, 0755)
		require.NoError(t, err)

		// Should fail due to missing extension
		err = validatePositionalArgs(instructionsFile, targetDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing extension")
	})

	t.Run("non-existent target path", func(t *testing.T) {
		tempDir := t.TempDir()
		instructionsFile := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(instructionsFile, []byte("test"), 0644)
		require.NoError(t, err)

		// Non-existent target should fail
		err = validatePositionalArgs(instructionsFile, "/non/existent/path")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target path does not exist")
	})

	t.Run("non-existent instructions file", func(t *testing.T) {
		tempDir := t.TempDir()
		targetDir := filepath.Join(tempDir, "src")
		err := os.Mkdir(targetDir, 0755)
		require.NoError(t, err)

		// Non-existent instructions file should fail
		err = validatePositionalArgs("/non/existent/instructions.txt", targetDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "instructions file does not exist")
	})
}
