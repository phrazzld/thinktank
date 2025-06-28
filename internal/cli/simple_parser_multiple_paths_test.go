package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseSimpleArgs_MultiplePaths tests parsing multiple target paths
func TestParseSimpleArgs_MultiplePaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		wantInstFile string
		wantPaths    []string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "single path - backward compatibility",
			args:         []string{"thinktank", "testdata/instructions.md", "testdata/", "--dry-run"},
			wantInstFile: "testdata/instructions.md",
			wantPaths:    []string{"testdata/"},
			wantErr:      false,
		},
		{
			name:         "two paths",
			args:         []string{"thinktank", "testdata/instructions.md", "simple_parser.go", "simple_config.go", "--dry-run"},
			wantInstFile: "testdata/instructions.md",
			wantPaths:    []string{"simple_parser.go", "simple_config.go"},
			wantErr:      false,
		},
		{
			name:         "multiple paths - files and directories",
			args:         []string{"thinktank", "testdata/instructions.md", "testdata/", "simple_parser.go", "simple_config.go", "--dry-run"},
			wantInstFile: "testdata/instructions.md",
			wantPaths:    []string{"testdata/", "simple_parser.go", "simple_config.go"},
			wantErr:      false,
		},
		{
			name:         "paths with flags at the end",
			args:         []string{"thinktank", "testdata/instructions.md", "testdata/", "simple_parser.go", "--dry-run", "--verbose"},
			wantInstFile: "testdata/instructions.md",
			wantPaths:    []string{"testdata/", "simple_parser.go"},
			wantErr:      false,
		},
		{
			name:         "flags between paths",
			args:         []string{"thinktank", "testdata/instructions.md", "testdata/", "--verbose", "simple_parser.go", "--dry-run", "simple_config.go"},
			wantInstFile: "testdata/instructions.md",
			wantPaths:    []string{"testdata/", "simple_parser.go", "simple_config.go"},
			wantErr:      false,
		},
		// Note: Paths with spaces are not supported when using multiple paths
		// This is a known limitation of the space-joining approach
		// Users should use paths without spaces or wait for future enhancement
		{
			name:        "no target paths",
			args:        []string{"thinktank", "testdata/instructions.md"},
			wantErr:     true,
			errContains: "usage:",
		},
		{
			name:        "only flags after instructions",
			args:        []string{"thinktank", "testdata/instructions.md", "--dry-run", "--verbose"},
			wantErr:     true,
			errContains: "at least one target path required",
		},
		{
			name:        "path that looks like unknown flag",
			args:        []string{"thinktank", "testdata/instructions.md", "testdata/", "--unknown-flag"},
			wantErr:     true,
			errContains: "unknown flag: --unknown-flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config, err := ParseSimpleArgsWithArgs(tt.args)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, config)
			assert.Equal(t, tt.wantInstFile, config.InstructionsFile)

			// Parse the space-joined paths
			actualPaths := strings.Fields(config.TargetPath)
			assert.Equal(t, tt.wantPaths, actualPaths, "Target paths should match")
		})
	}
}

// TestParseSimpleArgs_MultiplePathsWithValidation tests multiple paths with filesystem validation
func TestParseSimpleArgs_MultiplePathsWithValidation(t *testing.T) {
	// Skip this test during regular runs as it requires specific filesystem setup
	t.Skip("Integration test - requires specific filesystem setup")

	tests := []struct {
		name        string
		args        []string
		setupFiles  []string // Files to create before test
		wantErr     bool
		errContains string
	}{
		{
			name:       "all paths exist",
			args:       []string{"thinktank", "instructions.md", "testdata/", "README.md"},
			setupFiles: []string{"instructions.md", "README.md"},
			wantErr:    false,
		},
		{
			name:        "one path doesn't exist",
			args:        []string{"thinktank", "instructions.md", "testdata/", "nonexistent.go"},
			setupFiles:  []string{"instructions.md"},
			wantErr:     true,
			errContains: "does not exist",
		},
		{
			name:        "instructions file doesn't exist",
			args:        []string{"thinktank", "missing.md", "testdata/"},
			setupFiles:  []string{},
			wantErr:     true,
			errContains: "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would require filesystem setup in real implementation
			// For now, it's a placeholder for integration testing
		})
	}
}
