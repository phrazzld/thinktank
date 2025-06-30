package cli

import (
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultiplePathsIntegration tests the full flow from CLI args to MinimalConfig
func TestMultiplePathsIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		wantPaths   []string
		wantDryRun  bool
		wantVerbose bool
		wantErr     bool
	}{
		{
			name:       "single path backward compatibility",
			args:       []string{"thinktank", "testdata/instructions.md", "testdata/", "--dry-run"},
			wantPaths:  []string{"testdata/"},
			wantDryRun: true,
		},
		{
			name:       "two paths",
			args:       []string{"thinktank", "testdata/instructions.md", "simple_parser.go", "simple_config.go", "--dry-run"},
			wantPaths:  []string{"simple_parser.go", "simple_config.go"},
			wantDryRun: true,
		},
		{
			name:        "multiple paths with flags",
			args:        []string{"thinktank", "testdata/instructions.md", "testdata/", "--verbose", "simple_parser.go", "--dry-run", "simple_config.go"},
			wantPaths:   []string{"testdata/", "simple_parser.go", "simple_config.go"},
			wantDryRun:  true,
			wantVerbose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Parse arguments
			simplifiedConfig, err := ParseSimpleArgsWithArgs(tt.args)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Simulate conversion to MinimalConfig (as done in main.go)
			minimalConfig := &config.MinimalConfig{
				InstructionsFile: simplifiedConfig.InstructionsFile,
				TargetPaths:      strings.Fields(simplifiedConfig.TargetPath),
				ModelNames:       []string{config.DefaultModel},
				DryRun:           simplifiedConfig.HasFlag(FlagDryRun),
				Verbose:          simplifiedConfig.HasFlag(FlagVerbose),
			}

			// Verify the results
			assert.Equal(t, tt.wantPaths, minimalConfig.TargetPaths, "Target paths should match")
			assert.Equal(t, tt.wantDryRun, minimalConfig.DryRun, "DryRun flag should match")
			assert.Equal(t, tt.wantVerbose, minimalConfig.Verbose, "Verbose flag should match")
		})
	}
}

// TestMultiplePathsUsageDocumentation verifies the usage message is updated
func TestMultiplePathsUsageDocumentation(t *testing.T) {
	t.Parallel()

	_, err := ParseSimpleArgsWithArgs([]string{"thinktank"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "target_path...", "Usage should indicate multiple paths with ...")
}
