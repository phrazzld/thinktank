package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSafetyMarginCLIFlag_SimplestFailingTest tests that we can parse --token-safety-margin flag
// This should fail initially because the flag doesn't exist yet
func TestSafetyMarginCLIFlag_SimplestFailingTest(t *testing.T) {
	t.Parallel()

	args := []string{
		"thinktank",
		"--token-safety-margin", "30",
		"--dry-run",
		"../../README.md",
		".",
	}

	config, err := ParseSimpleArgsWithArgs(args)
	require.NoError(t, err)

	// This will fail because SafetyMargin field doesn't exist yet
	assert.Equal(t, uint8(30), config.SafetyMargin, "Safety margin should be parsed from CLI flag")
}

// TestSafetyMarginDefault tests that default safety margin is 20%
func TestSafetyMarginDefault(t *testing.T) {
	t.Parallel()

	args := []string{
		"thinktank",
		"--dry-run",
		"../../README.md",
		".",
	}

	config, err := ParseSimpleArgsWithArgs(args)
	require.NoError(t, err)

	// This will fail because SafetyMargin field doesn't exist yet
	assert.Equal(t, uint8(20), config.SafetyMargin, "Default safety margin should be 20%")
}

// TestSafetyMarginValidation tests that invalid safety margin values are rejected
func TestSafetyMarginValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		margin      string
		expectError bool
		errorSubstr string
	}{
		{
			name:        "valid minimum",
			margin:      "0",
			expectError: false,
		},
		{
			name:        "valid default",
			margin:      "20",
			expectError: false,
		},
		{
			name:        "valid maximum",
			margin:      "50",
			expectError: false,
		},
		{
			name:        "invalid negative",
			margin:      "-5",
			expectError: true,
			errorSubstr: "must be between 0% and 50%",
		},
		{
			name:        "invalid too high",
			margin:      "51",
			expectError: true,
			errorSubstr: "must be between 0% and 50%",
		},
		{
			name:        "invalid extreme",
			margin:      "100",
			expectError: true,
			errorSubstr: "must be between 0% and 50%",
		},
		{
			name:        "invalid non-numeric",
			margin:      "invalid",
			expectError: true,
			errorSubstr: "invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{
				"thinktank",
				"--token-safety-margin", tt.margin,
				"--dry-run",
				"../../README.md",
				".",
			}

			config, err := ParseSimpleArgsWithArgs(args)

			if tt.expectError {
				assert.Error(t, err, "Should have error for invalid safety margin")
				assert.Contains(t, err.Error(), tt.errorSubstr, "Error should contain expected substring")
			} else {
				assert.NoError(t, err, "Should not have error for valid safety margin")
				require.NotNil(t, config)
			}
		})
	}
}

// TestSafetyMarginEqualsFormat tests --token-safety-margin=value format
func TestSafetyMarginEqualsFormat(t *testing.T) {
	t.Parallel()

	args := []string{
		"thinktank",
		"--token-safety-margin=25",
		"--dry-run",
		"../../README.md",
		".",
	}

	config, err := ParseSimpleArgsWithArgs(args)
	require.NoError(t, err)

	// This will fail because SafetyMargin field doesn't exist yet
	assert.Equal(t, uint8(25), config.SafetyMargin, "Safety margin should be parsed from --flag=value format")
}
