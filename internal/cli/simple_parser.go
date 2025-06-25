// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"fmt"
	"os"
	"strings"
)

// ParseSimpleArgs parses the simplified command line interface in O(n) time using os.Args.
// Interface: thinktank instructions.txt ./src [--model gpt-4] [--output-dir ./out] [--dry-run] [--verbose] [--synthesis]
//
// This function embodies Rob Pike's philosophy: "Simplicity is the ultimate sophistication."
// It performs a single pass through os.Args with explicit state tracking and clear error messages.
func ParseSimpleArgs() (*SimplifiedConfig, error) {
	return ParseSimpleArgsWithArgs(os.Args)
}

// ParseSimpleArgsWithArgs parses arguments with dependency injection for testing.
// This enables comprehensive testing of the parsing logic without subprocess execution.
func ParseSimpleArgsWithArgs(args []string) (*SimplifiedConfig, error) {
	if len(args) < 3 {
		binary := "thinktank"
		if len(args) > 0 {
			binary = args[0]
		}
		return nil, fmt.Errorf("usage: %s instructions.txt target_path [flags...]", binary)
	}

	config := &SimplifiedConfig{
		InstructionsFile: args[1],
		TargetPath:       args[2],
		Flags:            0,
	}

	// Single pass through remaining arguments - O(n) time complexity
	for i := 3; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--dry-run":
			config.SetFlag(FlagDryRun)

		case arg == "--verbose":
			config.SetFlag(FlagVerbose)

		case arg == "--synthesis":
			config.SetFlag(FlagSynthesis)

		case arg == "--model":
			// --model flag requires a value
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--model flag requires a value")
			}
			i++ // Skip the model value - we use smart default in ToCliConfig()
			// Note: The specific model is handled by ToCliConfig() for now
			// This maintains the 33-byte SimplifiedConfig constraint

		case arg == "--output-dir":
			// --output-dir flag requires a value
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--output-dir flag requires a value")
			}
			i++ // Skip the output dir value - we use smart default in ToCliConfig()
			// Note: The specific output dir is handled by ToCliConfig() for now

		case strings.HasPrefix(arg, "--model="):
			// Handle --model=value format
			value := strings.TrimPrefix(arg, "--model=")
			if value == "" {
				return nil, fmt.Errorf("--model flag requires a non-empty value")
			}
			// Value stored implicitly - handled by ToCliConfig()

		case strings.HasPrefix(arg, "--output-dir="):
			// Handle --output-dir=value format
			value := strings.TrimPrefix(arg, "--output-dir=")
			if value == "" {
				return nil, fmt.Errorf("--output-dir flag requires a non-empty value")
			}
			// Value stored implicitly - handled by ToCliConfig()

		default:
			// Unknown flag - fail fast with clear error message
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	// Validate the parsed configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// SimpleParseResult represents the outcome of argument parsing with structured error handling
type SimpleParseResult struct {
	Config *SimplifiedConfig
	Error  error
}

// ParseSimpleArgsWithResult returns a structured result for better error handling patterns.
// This follows Go's explicit error handling philosophy while providing richer context.
func ParseSimpleArgsWithResult(args []string) *SimpleParseResult {
	config, err := ParseSimpleArgsWithArgs(args)
	return &SimpleParseResult{
		Config: config,
		Error:  err,
	}
}

// IsSuccess returns true if parsing succeeded
func (r *SimpleParseResult) IsSuccess() bool {
	return r.Error == nil && r.Config != nil
}

// MustConfig returns the config or panics if parsing failed.
// Use this only when you're certain parsing will succeed (e.g., in tests with known-good inputs).
func (r *SimpleParseResult) MustConfig() *SimplifiedConfig {
	if !r.IsSuccess() {
		panic(fmt.Sprintf("parsing failed: %v", r.Error))
	}
	return r.Config
}
