// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"fmt"
	"strings"

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

// ParseSimplifiedArgs parses command line arguments into a SimplifiedConfig
// Expected format: [flags...] <instructions-file> <target-path>
// Supported flags: --model, --output-dir, --dry-run, --verbose, --synthesis
func ParseSimplifiedArgs(args []string) (*SimplifiedConfig, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("insufficient arguments: instructions file and target path required")
	}

	config := &SimplifiedConfig{
		Flags: 0,
	}

	var positionalArgs []string

	// Parse flags and collect positional arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if !strings.HasPrefix(arg, "--") {
			// Not a flag, collect as positional argument
			positionalArgs = append(positionalArgs, arg)
			continue
		}

		// Handle flags
		switch {
		case arg == "--dry-run":
			config.SetFlag(FlagDryRun)
		case arg == "--verbose":
			config.SetFlag(FlagVerbose)
		case arg == "--synthesis":
			config.SetFlag(FlagSynthesis)
		case arg == "--model":
			// --model requires a value, but we ignore it for SimplifiedConfig
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag needs an argument: %s", arg)
			}
			i++ // Skip the value
		case arg == "--output-dir":
			// --output-dir requires a value, but we ignore it for SimplifiedConfig
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag needs an argument: %s", arg)
			}
			i++ // Skip the value
		case strings.HasPrefix(arg, "--model="):
			// Handle --model=value format
			value := strings.TrimPrefix(arg, "--model=")
			if value == "" {
				return nil, fmt.Errorf("flag has empty value: %s", arg)
			}
			// Value ignored for SimplifiedConfig
		case strings.HasPrefix(arg, "--output-dir="):
			// Handle --output-dir=value format
			value := strings.TrimPrefix(arg, "--output-dir=")
			if value == "" {
				return nil, fmt.Errorf("flag has empty value: %s", arg)
			}
			// Value ignored for SimplifiedConfig
		default:
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	// Validate we have enough positional arguments
	if len(positionalArgs) < 2 {
		if len(positionalArgs) == 0 {
			return nil, fmt.Errorf("insufficient arguments: instructions file and target path required")
		}
		return nil, fmt.Errorf("target path required")
	}

	// Assign positional arguments
	config.InstructionsFile = positionalArgs[0]
	config.TargetPath = positionalArgs[1]

	return config, nil
}
