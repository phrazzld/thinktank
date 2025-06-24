// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"testing"
	"unsafe"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestSimplifiedConfigSize validates the struct is compact and memory-efficient
func TestSimplifiedConfigSize(t *testing.T) {
	// Test that the SimplifiedConfig struct is significantly smaller than CliConfig
	var config SimplifiedConfig
	actualSize := unsafe.Sizeof(config)

	// The struct should be compact - much smaller than the 200+ byte CliConfig
	// Go's alignment may add padding, but the logical data is 33 bytes
	assert.LessOrEqual(t, actualSize, uintptr(48), "SimplifiedConfig should be ≤48 bytes with alignment")
	assert.GreaterOrEqual(t, actualSize, uintptr(33), "SimplifiedConfig should be ≥33 bytes of logical data")

	// Verify the logical data size is exactly 33 bytes (2 strings + 1 byte)
	logicalSize := unsafe.Sizeof(config.InstructionsFile) + unsafe.Sizeof(config.TargetPath) + unsafe.Sizeof(config.Flags)
	assert.Equal(t, uintptr(33), logicalSize, "Logical data size should be exactly 33 bytes")
}

// TestSimplifiedConfigFields validates the required fields are present
func TestSimplifiedConfigFields(t *testing.T) {
	config := SimplifiedConfig{
		InstructionsFile: "test.md",
		TargetPath:       "src/",
		Flags:            0x01, // FlagDryRun
	}

	assert.Equal(t, "test.md", config.InstructionsFile)
	assert.Equal(t, "src/", config.TargetPath)
	assert.Equal(t, uint8(0x01), config.Flags)
}

// TestFlagOperations validates O(1) bitfield operations for boolean flags
func TestFlagOperations(t *testing.T) {
	tests := []struct {
		name      string
		initial   uint8
		flag      uint8
		operation string
		expected  uint8
		hasFlag   bool
	}{
		{
			name:      "set dry run flag",
			initial:   0x00,
			flag:      FlagDryRun,
			operation: "set",
			expected:  FlagDryRun,
			hasFlag:   true,
		},
		{
			name:      "set verbose flag",
			initial:   0x00,
			flag:      FlagVerbose,
			operation: "set",
			expected:  FlagVerbose,
			hasFlag:   true,
		},
		{
			name:      "set synthesis flag",
			initial:   0x00,
			flag:      FlagSynthesis,
			operation: "set",
			expected:  FlagSynthesis,
			hasFlag:   true,
		},
		{
			name:      "set multiple flags",
			initial:   FlagDryRun,
			flag:      FlagVerbose,
			operation: "set",
			expected:  FlagDryRun | FlagVerbose,
			hasFlag:   true,
		},
		{
			name:      "clear flag",
			initial:   FlagDryRun | FlagVerbose,
			flag:      FlagDryRun,
			operation: "clear",
			expected:  FlagVerbose,
			hasFlag:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "src/",
				Flags:            tt.initial,
			}

			switch tt.operation {
			case "set":
				config.SetFlag(tt.flag)
			case "clear":
				config.ClearFlag(tt.flag)
			}

			assert.Equal(t, tt.expected, config.Flags, "Flag operation result mismatch")
			assert.Equal(t, tt.hasFlag, config.HasFlag(tt.flag), "HasFlag result mismatch")
		})
	}
}

// TestFlagConstants validates the flag bit patterns are correct
func TestFlagConstants(t *testing.T) {
	// Verify flag constants have correct bit patterns for O(1) operations
	assert.Equal(t, uint8(0x01), FlagDryRun, "FlagDryRun should be 0x01")
	assert.Equal(t, uint8(0x02), FlagVerbose, "FlagVerbose should be 0x02")
	assert.Equal(t, uint8(0x04), FlagSynthesis, "FlagSynthesis should be 0x04")

	// Verify flags are mutually exclusive bit patterns
	assert.Equal(t, uint8(0x00), FlagDryRun&FlagVerbose, "Flags should be mutually exclusive")
	assert.Equal(t, uint8(0x00), FlagDryRun&FlagSynthesis, "Flags should be mutually exclusive")
	assert.Equal(t, uint8(0x00), FlagVerbose&FlagSynthesis, "Flags should be mutually exclusive")
}

// TestToCliConfig validates conversion from SimplifiedConfig to CliConfig
func TestToCliConfig(t *testing.T) {
	tests := []struct {
		name     string
		simple   SimplifiedConfig
		validate func(*testing.T, *config.CliConfig)
	}{
		{
			name: "basic conversion",
			simple: SimplifiedConfig{
				InstructionsFile: "instructions.md",
				TargetPath:       "./src",
				Flags:            0x00,
			},
			validate: func(t *testing.T, cfg *config.CliConfig) {
				assert.Equal(t, "instructions.md", cfg.InstructionsFile)
				assert.Equal(t, []string{"./src"}, cfg.Paths)
				assert.False(t, cfg.DryRun)
				assert.False(t, cfg.Verbose)
				assert.Equal(t, []string{DefaultModel}, cfg.ModelNames)
				assert.Equal(t, DefaultOutputDir, cfg.OutputDir)
			},
		},
		{
			name: "with dry run flag",
			simple: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "./code",
				Flags:            FlagDryRun,
			},
			validate: func(t *testing.T, cfg *config.CliConfig) {
				assert.True(t, cfg.DryRun)
				assert.False(t, cfg.Verbose)
			},
		},
		{
			name: "with verbose flag",
			simple: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "./code",
				Flags:            FlagVerbose,
			},
			validate: func(t *testing.T, cfg *config.CliConfig) {
				assert.False(t, cfg.DryRun)
				assert.True(t, cfg.Verbose)
			},
		},
		{
			name: "with synthesis flag",
			simple: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "./code",
				Flags:            FlagSynthesis,
			},
			validate: func(t *testing.T, cfg *config.CliConfig) {
				assert.NotEmpty(t, cfg.SynthesisModel, "Should set synthesis model")
				assert.Greater(t, len(cfg.ModelNames), 1, "Should use multiple models for synthesis")
			},
		},
		{
			name: "with multiple flags",
			simple: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "./code",
				Flags:            FlagVerbose | FlagSynthesis,
			},
			validate: func(t *testing.T, cfg *config.CliConfig) {
				assert.True(t, cfg.Verbose)
				assert.NotEmpty(t, cfg.SynthesisModel)
				assert.Greater(t, len(cfg.ModelNames), 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.simple.ToCliConfig()
			assert.NotNil(t, cfg, "ToCliConfig should not return nil")
			tt.validate(t, cfg)
		})
	}
}

// TestValidate validates the O(1) validation logic
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  SimplifiedConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "src/",
				Flags:            0x00,
			},
			wantErr: false,
		},
		{
			name: "missing instructions file",
			config: SimplifiedConfig{
				InstructionsFile: "",
				TargetPath:       "src/",
				Flags:            0x00,
			},
			wantErr: true,
			errMsg:  "instructions file required",
		},
		{
			name: "missing target path",
			config: SimplifiedConfig{
				InstructionsFile: "test.md",
				TargetPath:       "",
				Flags:            0x00,
			},
			wantErr: true,
			errMsg:  "target path required",
		},
		{
			name: "dry run allows missing instructions",
			config: SimplifiedConfig{
				InstructionsFile: "",
				TargetPath:       "src/",
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
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
