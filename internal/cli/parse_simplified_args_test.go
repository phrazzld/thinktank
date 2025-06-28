// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseSimplifiedArgs_Basic tests the most basic parsing scenario
// This is our first failing test - the minimal case that drives the design
func TestParseSimplifiedArgs_Basic(t *testing.T) {
	args := []string{"instructions.md", "src/"}

	config, err := ParseSimplifiedArgs(args)

	require.NoError(t, err)
	assert.Equal(t, "instructions.md", config.InstructionsFile)
	assert.Equal(t, "src/", config.TargetPath)
	assert.Equal(t, uint8(0), config.Flags) // No flags set
}

// TestParseSimplifiedArgs_EmptyArgs tests the error case for empty arguments
func TestParseSimplifiedArgs_EmptyArgs(t *testing.T) {
	args := []string{}

	config, err := ParseSimplifiedArgs(args)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "insufficient arguments")
}

// TestParseSimplifiedArgs_InsufficientArgs tests missing target path
func TestParseSimplifiedArgs_InsufficientArgs(t *testing.T) {
	args := []string{"instructions.md"}

	config, err := ParseSimplifiedArgs(args)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "target path required")
}

// TestParseSimplifiedArgs_TableDriven tests comprehensive parsing scenarios
func TestParseSimplifiedArgs_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
		expected    *SimplifiedConfig
	}{
		{
			name: "basic positional args",
			args: []string{"instructions.md", "src/"},
			expected: &SimplifiedConfig{
				InstructionsFile: "instructions.md",
				TargetPath:       "src/",
				Flags:            0,
			},
		},
		{
			name: "with dry-run flag",
			args: []string{"--dry-run", "instructions.md", "src/"},
			expected: &SimplifiedConfig{
				InstructionsFile: "instructions.md",
				TargetPath:       "src/",
				Flags:            FlagDryRun,
			},
		},
		{
			name: "with verbose flag",
			args: []string{"--verbose", "instructions.md", "src/"},
			expected: &SimplifiedConfig{
				InstructionsFile: "instructions.md",
				TargetPath:       "src/",
				Flags:            FlagVerbose,
			},
		},
		{
			name: "with synthesis flag",
			args: []string{"--synthesis", "instructions.md", "src/"},
			expected: &SimplifiedConfig{
				InstructionsFile: "instructions.md",
				TargetPath:       "src/",
				Flags:            FlagSynthesis,
			},
		},
		{
			name: "with multiple flags",
			args: []string{"--dry-run", "--verbose", "instructions.md", "src/"},
			expected: &SimplifiedConfig{
				InstructionsFile: "instructions.md",
				TargetPath:       "src/",
				Flags:            FlagDryRun | FlagVerbose,
			},
		},
		{
			name: "with model flag (ignored for simplified config)",
			args: []string{"--model", "gpt-4", "instructions.md", "src/"},
			expected: &SimplifiedConfig{
				InstructionsFile: "instructions.md",
				TargetPath:       "src/",
				Flags:            0,
			},
		},
		{
			name: "with output-dir flag (ignored for simplified config)",
			args: []string{"--output-dir", "output/", "instructions.md", "src/"},
			expected: &SimplifiedConfig{
				InstructionsFile: "instructions.md",
				TargetPath:       "src/",
				Flags:            0,
			},
		},
		{
			name:        "empty args",
			args:        []string{},
			expectError: true,
			errorMsg:    "insufficient arguments",
		},
		{
			name:        "only one arg",
			args:        []string{"instructions.md"},
			expectError: true,
			errorMsg:    "target path required",
		},
		{
			name:        "invalid flag",
			args:        []string{"--invalid-flag", "instructions.md", "src/"},
			expectError: true,
			errorMsg:    "unknown flag",
		},
		{
			name:        "flag without value at end",
			args:        []string{"instructions.md", "src/", "--model"},
			expectError: true,
			errorMsg:    "flag needs an argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseSimplifiedArgs(tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				assert.Equal(t, tt.expected.InstructionsFile, config.InstructionsFile)
				assert.Equal(t, tt.expected.TargetPath, config.TargetPath)
				assert.Equal(t, tt.expected.Flags, config.Flags)

				// Verify flag operations work correctly
				assert.Equal(t, tt.expected.HasFlag(FlagDryRun), config.HasFlag(FlagDryRun))
				assert.Equal(t, tt.expected.HasFlag(FlagVerbose), config.HasFlag(FlagVerbose))
				assert.Equal(t, tt.expected.HasFlag(FlagSynthesis), config.HasFlag(FlagSynthesis))
			}
		})
	}
}

// TestParseSimplifiedArgs_EdgeCases tests additional edge cases for 100% coverage
func TestParseSimplifiedArgs_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
		expected    *SimplifiedConfig
	}{
		{
			name: "flags with equals syntax",
			args: []string{"--model=gpt-4", "--output-dir=./out", "instructions.md", "src/"},
			expected: &SimplifiedConfig{
				InstructionsFile: "instructions.md",
				TargetPath:       "src/",
				Flags:            0,
			},
		},
		{
			name: "mixed flag and equals syntax",
			args: []string{"--model", "gpt-4", "--output-dir=./out", "--verbose", "instructions.md", "src/"},
			expected: &SimplifiedConfig{
				InstructionsFile: "instructions.md",
				TargetPath:       "src/",
				Flags:            FlagVerbose,
			},
		},
		{
			name:        "model with equals but empty value",
			args:        []string{"--model=", "instructions.md", "src/"},
			expectError: true,
			errorMsg:    "empty value",
		},
		{
			name:        "output-dir with equals but empty value",
			args:        []string{"--output-dir=", "instructions.md", "src/"},
			expectError: true,
			errorMsg:    "empty value",
		},
		{
			name: "all flags combined with values",
			args: []string{"--model", "claude-3", "--output-dir", "./results", "--dry-run", "--verbose", "--synthesis", "test.md", "src/"},
			expected: &SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "src/",
				Flags:            FlagDryRun | FlagVerbose | FlagSynthesis,
			},
		},
		{
			name: "flags before and after positional args",
			args: []string{"--verbose", "instructions.md", "--dry-run", "src/", "--synthesis"},
			expected: &SimplifiedConfig{
				InstructionsFile: "instructions.md",
				TargetPath:       "src/",
				Flags:            FlagVerbose | FlagDryRun | FlagSynthesis,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseSimplifiedArgs(tt.args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				assert.Equal(t, tt.expected.InstructionsFile, config.InstructionsFile)
				assert.Equal(t, tt.expected.TargetPath, config.TargetPath)
				assert.Equal(t, tt.expected.Flags, config.Flags)
			}
		})
	}
}

// TestParseSimplifiedArgs_Integration tests the integration with other components
func TestParseSimplifiedArgs_Integration(t *testing.T) {
	// Create temporary test files and directories
	tempDir := t.TempDir()
	validInstFile := filepath.Join(tempDir, "instructions.md")
	require.NoError(t, os.WriteFile(validInstFile, []byte("test instructions"), 0644))

	validTargetDir := filepath.Join(tempDir, "src")
	require.NoError(t, os.Mkdir(validTargetDir, 0755))

	// Set up API key for validation
	oldGeminiKey := os.Getenv("GEMINI_API_KEY")
	_ = os.Setenv("GEMINI_API_KEY", "test-key")
	defer func() { _ = os.Setenv("GEMINI_API_KEY", oldGeminiKey) }()

	t.Run("parsed config should validate", func(t *testing.T) {
		config, err := ParseSimplifiedArgs([]string{validInstFile, validTargetDir})
		require.NoError(t, err)
		assert.NoError(t, config.Validate())
	})

	t.Run("dry-run flag allows empty instructions", func(t *testing.T) {
		config, err := ParseSimplifiedArgs([]string{"--dry-run", "", validTargetDir})
		require.NoError(t, err)
		assert.True(t, config.HasFlag(FlagDryRun))
		assert.NoError(t, config.Validate()) // Should validate because dry-run allows empty instructions
	})

	t.Run("exactly three positional args", func(t *testing.T) {
		config, err := ParseSimplifiedArgs([]string{validInstFile, validTargetDir, "extra-arg"})
		require.NoError(t, err)
		assert.Equal(t, validInstFile, config.InstructionsFile)
		assert.Equal(t, validTargetDir, config.TargetPath)
		// Extra args are ignored
	})

	t.Run("only flags no positional", func(t *testing.T) {
		config, err := ParseSimplifiedArgs([]string{"--verbose", "--dry-run"})
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "insufficient arguments")
	})

	t.Run("output-dir without value at end", func(t *testing.T) {
		config, err := ParseSimplifiedArgs([]string{"instructions.md", "src/", "--output-dir"})
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "flag needs an argument")
	})
}
