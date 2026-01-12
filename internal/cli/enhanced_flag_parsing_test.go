package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseOptionalFlags_AllFlags ensures comprehensive coverage of all supported flags
func TestParseOptionalFlags_AllFlags(t *testing.T) {
	// Note: Not using t.Parallel() as this is a straightforward sequential test

	tests := []struct {
		name          string
		args          []string
		expectedFlags uint8
		description   string
	}{
		{
			name:          "debug flag",
			args:          []string{"--debug", "instructions.md", "src/"},
			expectedFlags: FlagDebug,
			description:   "Debug flag should be parsed correctly",
		},
		{
			name:          "quiet flag",
			args:          []string{"--quiet", "instructions.md", "src/"},
			expectedFlags: FlagQuiet,
			description:   "Quiet flag should be parsed correctly",
		},
		{
			name:          "json-logs flag",
			args:          []string{"--json-logs", "instructions.md", "src/"},
			expectedFlags: FlagJsonLogs,
			description:   "JsonLogs flag should be parsed correctly",
		},
		{
			name:          "no-progress flag",
			args:          []string{"--no-progress", "instructions.md", "src/"},
			expectedFlags: FlagNoProgress,
			description:   "NoProgress flag should be parsed correctly",
		},
		{
			name:          "all boolean flags combined",
			args:          []string{"--dry-run", "--verbose", "--synthesis", "--debug", "--quiet", "--json-logs", "--no-progress", "instructions.md", "src/"},
			expectedFlags: FlagDryRun | FlagVerbose | FlagSynthesis | FlagDebug | FlagQuiet | FlagJsonLogs | FlagNoProgress,
			description:   "All boolean flags should work together",
		},
		{
			name:          "newer flags with value flags",
			args:          []string{"--debug", "--model", "gpt-4", "--quiet", "--output-dir", "./out", "instructions.md", "src/"},
			expectedFlags: FlagDebug | FlagQuiet,
			description:   "Newer flags should work with value flags",
		},
		{
			name:          "newer flags with equals syntax",
			args:          []string{"--json-logs", "--model=gemini-3-flash", "--no-progress", "--output-dir=./results", "instructions.md", "src/"},
			expectedFlags: FlagJsonLogs | FlagNoProgress,
			description:   "Newer flags should work with equals syntax for value flags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseSimplifiedArgs(tt.args)

			require.NoError(t, err, "Parsing should succeed for: %s", tt.description)
			require.NotNil(t, config, "Config should not be nil")

			assert.Equal(t, tt.expectedFlags, config.Flags, "Flags mismatch for: %s", tt.description)
			assert.Equal(t, "instructions.md", config.InstructionsFile, "Instructions file should be parsed correctly")
			assert.Equal(t, "src/", config.TargetPath, "Target path should be parsed correctly")

			// Verify individual flag checking works
			assert.Equal(t, (tt.expectedFlags&FlagDryRun) != 0, config.HasFlag(FlagDryRun))
			assert.Equal(t, (tt.expectedFlags&FlagVerbose) != 0, config.HasFlag(FlagVerbose))
			assert.Equal(t, (tt.expectedFlags&FlagSynthesis) != 0, config.HasFlag(FlagSynthesis))
			assert.Equal(t, (tt.expectedFlags&FlagDebug) != 0, config.HasFlag(FlagDebug))
			assert.Equal(t, (tt.expectedFlags&FlagQuiet) != 0, config.HasFlag(FlagQuiet))
			assert.Equal(t, (tt.expectedFlags&FlagJsonLogs) != 0, config.HasFlag(FlagJsonLogs))
			assert.Equal(t, (tt.expectedFlags&FlagNoProgress) != 0, config.HasFlag(FlagNoProgress))
		})
	}
}

// TestParseOptionalFlags_FlagManipulation tests the flag manipulation methods
func TestParseOptionalFlags_FlagManipulation(t *testing.T) {
	// Note: Not using t.Parallel() as this is a straightforward sequential test

	config := &SimplifiedConfig{
		InstructionsFile: "test.md",
		TargetPath:       "src/",
		Flags:            0,
	}

	// Test SetFlag
	config.SetFlag(FlagDebug)
	assert.True(t, config.HasFlag(FlagDebug), "SetFlag should set the debug flag")
	assert.False(t, config.HasFlag(FlagVerbose), "SetFlag should not affect other flags")

	config.SetFlag(FlagQuiet)
	assert.True(t, config.HasFlag(FlagDebug), "SetFlag should preserve existing flags")
	assert.True(t, config.HasFlag(FlagQuiet), "SetFlag should set the quiet flag")

	// Test ClearFlag
	config.ClearFlag(FlagDebug)
	assert.False(t, config.HasFlag(FlagDebug), "ClearFlag should clear the debug flag")
	assert.True(t, config.HasFlag(FlagQuiet), "ClearFlag should preserve other flags")

	// Test multiple flags
	config.SetFlag(FlagJsonLogs | FlagNoProgress)
	assert.True(t, config.HasFlag(FlagJsonLogs), "Multiple SetFlag should work with JsonLogs")
	assert.True(t, config.HasFlag(FlagNoProgress), "Multiple SetFlag should work with NoProgress")
	assert.True(t, config.HasFlag(FlagQuiet), "Multiple SetFlag should preserve existing flags")

	// Test ClearFlag with multiple flags
	config.ClearFlag(FlagQuiet | FlagJsonLogs)
	assert.False(t, config.HasFlag(FlagQuiet), "Multiple ClearFlag should clear Quiet")
	assert.False(t, config.HasFlag(FlagJsonLogs), "Multiple ClearFlag should clear JsonLogs")
	assert.True(t, config.HasFlag(FlagNoProgress), "Multiple ClearFlag should preserve other flags")
}

// TestParseOptionalFlags_EdgeCasesEnhanced tests additional edge cases for the enhanced parser
func TestParseOptionalFlags_EdgeCasesEnhanced(t *testing.T) {
	// Note: Not using t.Parallel() as this is a straightforward sequential test

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
		description string
	}{
		{
			name:        "single dash with multiple characters",
			args:        []string{"-abc", "instructions.md", "src/"},
			expectError: true,
			errorMsg:    "unknown flag: -abc",
			description: "Single dash with multiple characters should be rejected",
		},
		{
			name:        "unknown long flag",
			args:        []string{"--unknown", "instructions.md", "src/"},
			expectError: true,
			errorMsg:    "unknown flag: --unknown",
			description: "Unknown long flags should be rejected",
		},
		{
			name:        "hyphenated argument that's not a flag",
			args:        []string{"instructions.md", "src/", "some-file.txt"},
			expectError: false,
			description: "Hyphenated positional arguments should be accepted",
		},
		{
			name:        "flag-like positional argument",
			args:        []string{"instructions.md", "--not-a-flag"},
			expectError: true,
			errorMsg:    "unknown flag: --not-a-flag",
			description: "Flag-like strings are parsed as flags and rejected if unknown",
		},
		{
			name:        "empty equals value for model",
			args:        []string{"--model=", "instructions.md", "src/"},
			expectError: true,
			errorMsg:    "flag has empty value",
			description: "Empty value after equals should be rejected",
		},
		{
			name:        "empty equals value for output-dir",
			args:        []string{"--output-dir=", "instructions.md", "src/"},
			expectError: true,
			errorMsg:    "flag has empty value",
			description: "Empty value after equals should be rejected for output-dir",
		},
		{
			name:        "short flag missing argument",
			args:        []string{"-m"},
			expectError: true,
			errorMsg:    "flag needs an argument: -m",
			description: "Short model flag without value should be rejected",
		},
		{
			name:        "model flag at end without value",
			args:        []string{"instructions.md", "src/", "--model"},
			expectError: true,
			errorMsg:    "flag needs an argument: --model",
			description: "Model flag at end without value should be rejected",
		},
		{
			name:        "output-dir flag at end without value",
			args:        []string{"instructions.md", "src/", "--output-dir"},
			expectError: true,
			errorMsg:    "flag needs an argument: --output-dir",
			description: "Output-dir flag at end without value should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseSimplifiedArgs(tt.args)

			if tt.expectError {
				assert.Error(t, err, "Should return error for: %s", tt.description)
				assert.Nil(t, config, "Config should be nil on error")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Should not return error for: %s", tt.description)
				assert.NotNil(t, config, "Config should not be nil on success")
			}
		})
	}
}

// TestParseOptionalFlags_PositionalArguments tests handling of positional arguments with flags
func TestParseOptionalFlags_PositionalArguments(t *testing.T) {
	// Note: Not using t.Parallel() as this is a straightforward sequential test

	tests := []struct {
		name                 string
		args                 []string
		expectedInstructions string
		expectedTarget       string
		expectedFlags        uint8
		description          string
	}{
		{
			name:                 "flags before positional args",
			args:                 []string{"--verbose", "--debug", "test.md", "src/"},
			expectedInstructions: "test.md",
			expectedTarget:       "src/",
			expectedFlags:        FlagVerbose | FlagDebug,
			description:          "Flags before positional arguments should work",
		},
		{
			name:                 "flags after positional args",
			args:                 []string{"test.md", "src/", "--quiet", "--no-progress"},
			expectedInstructions: "test.md",
			expectedTarget:       "src/",
			expectedFlags:        FlagQuiet | FlagNoProgress,
			description:          "Flags after positional arguments should work",
		},
		{
			name:                 "flags mixed with positional args",
			args:                 []string{"--dry-run", "test.md", "--json-logs", "src/", "--synthesis"},
			expectedInstructions: "test.md",
			expectedTarget:       "src/",
			expectedFlags:        FlagDryRun | FlagJsonLogs | FlagSynthesis,
			description:          "Flags mixed between positional arguments should work",
		},
		{
			name:                 "value flags with positional args",
			args:                 []string{"--model", "gpt-4", "test.md", "--output-dir", "./out", "src/"},
			expectedInstructions: "test.md",
			expectedTarget:       "src/",
			expectedFlags:        0,
			description:          "Value flags should not interfere with positional argument parsing",
		},
		{
			name:                 "equals syntax with positional args",
			args:                 []string{"--model=gemini-3-flash", "test.md", "--output-dir=./results", "src/"},
			expectedInstructions: "test.md",
			expectedTarget:       "src/",
			expectedFlags:        0,
			description:          "Equals syntax should not interfere with positional argument parsing",
		},
		{
			name:                 "complex mixed scenario",
			args:                 []string{"-v", "--model=gpt-4", "my-instructions.md", "--debug", "--output-dir", "./output", "my-source/", "--quiet"},
			expectedInstructions: "my-instructions.md",
			expectedTarget:       "my-source/",
			expectedFlags:        FlagVerbose | FlagDebug | FlagQuiet,
			description:          "Complex mixed scenario should parse correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseSimplifiedArgs(tt.args)

			require.NoError(t, err, "Parsing should succeed for: %s", tt.description)
			require.NotNil(t, config, "Config should not be nil")

			assert.Equal(t, tt.expectedInstructions, config.InstructionsFile, "Instructions file mismatch for: %s", tt.description)
			assert.Equal(t, tt.expectedTarget, config.TargetPath, "Target path mismatch for: %s", tt.description)
			assert.Equal(t, tt.expectedFlags, config.Flags, "Flags mismatch for: %s", tt.description)
		})
	}
}

// TestParseOptionalFlags_ShortFlagsEnhanced tests short flag combinations and edge cases
func TestParseOptionalFlags_ShortFlagsEnhanced(t *testing.T) {
	// Note: Not using t.Parallel() as this is a straightforward sequential test

	tests := []struct {
		name          string
		args          []string
		expectedFlags uint8
		expectError   bool
		errorMsg      string
		description   string
	}{
		{
			name:          "short verbose flag",
			args:          []string{"-v", "test.md", "src/"},
			expectedFlags: FlagVerbose,
			description:   "Short verbose flag should work",
		},
		{
			name:          "short model flag with value",
			args:          []string{"-m", "gpt-4", "test.md", "src/"},
			expectedFlags: 0,
			description:   "Short model flag with value should work (value ignored)",
		},
		{
			name:          "short flags combined with long flags",
			args:          []string{"-v", "--debug", "-m", "claude-3", "--quiet", "test.md", "src/"},
			expectedFlags: FlagVerbose | FlagDebug | FlagQuiet,
			description:   "Short and long flags should work together",
		},
		{
			name:        "unknown short flag",
			args:        []string{"-x", "test.md", "src/"},
			expectError: true,
			errorMsg:    "unknown flag: -x",
			description: "Unknown short flags should be rejected",
		},
		{
			name:        "short model flag without value",
			args:        []string{"-m"},
			expectError: true,
			errorMsg:    "flag needs an argument: -m",
			description: "Short model flag without value should be rejected",
		},
		{
			name:        "short flag with multiple characters",
			args:        []string{"-vv", "test.md", "src/"},
			expectError: true,
			errorMsg:    "unknown flag: -vv",
			description: "Short flags with multiple characters should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseSimplifiedArgs(tt.args)

			if tt.expectError {
				assert.Error(t, err, "Should return error for: %s", tt.description)
				assert.Nil(t, config, "Config should be nil on error")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				require.NoError(t, err, "Parsing should succeed for: %s", tt.description)
				require.NotNil(t, config, "Config should not be nil")
				assert.Equal(t, tt.expectedFlags, config.Flags, "Flags mismatch for: %s", tt.description)
			}
		})
	}
}
