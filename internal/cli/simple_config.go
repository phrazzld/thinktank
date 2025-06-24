// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"fmt"

	"github.com/phrazzld/thinktank/internal/config"
)

// SimplifiedConfig represents the essential configuration in exactly 33 bytes.
// Following Go's principle of "less is more", this struct contains only the
// absolutely necessary fields with smart defaults for everything else.
//
// Memory layout (33 bytes total on 64-bit systems):
// - InstructionsFile: 16 bytes (string header: ptr+len)
// - TargetPath: 16 bytes (string header: ptr+len)
// - Flags: 1 byte (bitfield for DryRun|Verbose|Synthesis)
type SimplifiedConfig struct {
	InstructionsFile string // 16 bytes (pointer + length on 64-bit)
	TargetPath       string // 16 bytes (pointer + length on 64-bit)
	Flags            uint8  // 1 byte bitfield
	// Note: Actual struct size may include padding. The 33-byte target
	// refers to the logical data size, not the Go struct alignment.
}

// Flag constants for bitwise operations - O(1) validation
const (
	FlagDryRun    uint8 = 1 << iota // 0x01
	FlagVerbose                     // 0x02
	FlagSynthesis                   // 0x04
	// 5 bits remaining for future expansion
)

// Smart defaults - applied during parsing/conversion
const (
	DefaultModel     = "gemini-2.5-pro"
	DefaultOutputDir = "."
)

// HasFlag checks if a flag is set using bitwise AND - O(1) operation
func (s *SimplifiedConfig) HasFlag(flag uint8) bool {
	return s.Flags&flag != 0
}

// SetFlag sets a flag using bitwise OR - O(1) operation
func (s *SimplifiedConfig) SetFlag(flag uint8) {
	s.Flags |= flag
}

// ClearFlag clears a flag using bitwise AND NOT - O(1) operation
func (s *SimplifiedConfig) ClearFlag(flag uint8) {
	s.Flags &^= flag
}

// ToCliConfig converts SimplifiedConfig to the full CliConfig for compatibility.
// This is where the "smart defaults" magic happens - O(1) expansion.
func (s *SimplifiedConfig) ToCliConfig() *config.CliConfig {
	cfg := config.NewDefaultCliConfig()

	// Set the essentials from simplified config
	cfg.InstructionsFile = s.InstructionsFile
	cfg.Paths = []string{s.TargetPath}

	// Apply flags with bitwise operations - O(1)
	cfg.DryRun = s.HasFlag(FlagDryRun)
	cfg.Verbose = s.HasFlag(FlagVerbose)

	// Smart model selection based on synthesis flag
	if s.HasFlag(FlagSynthesis) {
		// Use multiple models for synthesis
		cfg.ModelNames = []string{"gemini-2.5-pro", "gpt-4o"}
		cfg.SynthesisModel = "gemini-2.5-pro"
		cfg.OutputDir = "synthesis-output"
	} else {
		// Single model for normal operation
		cfg.ModelNames = []string{DefaultModel}
		cfg.OutputDir = DefaultOutputDir
	}

	return cfg
}

// Validate performs essential validation in O(1) time using bitwise operations
func (s *SimplifiedConfig) Validate() error {
	// Essential field validation
	if s.InstructionsFile == "" && !s.HasFlag(FlagDryRun) {
		return fmt.Errorf("instructions file required")
	}

	if s.TargetPath == "" {
		return fmt.Errorf("target path required")
	}

	return nil
}
