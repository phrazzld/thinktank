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
// Supports multiple target paths: thinktank instructions.txt path1 path2 [flags...]
func ParseSimpleArgsWithArgs(args []string) (*SimplifiedConfig, error) {
	// Check for help flag early (handle empty args gracefully)
	if len(args) > 1 {
		for _, arg := range args[1:] {
			if arg == "--help" || arg == "-h" {
				return &SimplifiedConfig{
					Flags: FlagHelp,
				}, nil
			}
		}
	}

	if len(args) < 3 {
		binary := "thinktank"
		if len(args) > 0 {
			binary = args[0]
		}
		return nil, fmt.Errorf("usage: %s instructions.txt target_path... [flags...]", binary)
	}

	// First pass: separate positional args from flags
	var instructionsFile string
	var targetPaths []string
	flags := uint8(0)

	// Track if we've seen the instructions file
	seenInstructions := false

	// Single pass through all arguments - O(n) time complexity
	for i := 1; i < len(args); i++ {
		arg := args[i]

		// Check if this is a flag
		switch {
		case arg == "--help" || arg == "-h":
			flags |= FlagHelp

		case arg == "--dry-run":
			flags |= FlagDryRun

		case arg == "--verbose":
			flags |= FlagVerbose

		case arg == "--synthesis":
			flags |= FlagSynthesis

		case arg == "--debug":
			flags |= FlagDebug

		case arg == "--quiet":
			flags |= FlagQuiet

		case arg == "--json-logs":
			flags |= FlagJsonLogs

		case arg == "--no-progress":
			flags |= FlagNoProgress

		case arg == "--model":
			// --model flag requires a value
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--model flag requires a value")
			}
			i++ // Skip the model value - we use smart default in ToCliConfig()

		case arg == "--output-dir":
			// --output-dir flag requires a value
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--output-dir flag requires a value")
			}
			i++ // Skip the output dir value - we use smart default in ToCliConfig()

		case strings.HasPrefix(arg, "--model="):
			// Handle --model=value format
			value := strings.TrimPrefix(arg, "--model=")
			if value == "" {
				return nil, fmt.Errorf("--model flag requires a non-empty value")
			}

		case strings.HasPrefix(arg, "--output-dir="):
			// Handle --output-dir=value format
			value := strings.TrimPrefix(arg, "--output-dir=")
			if value == "" {
				return nil, fmt.Errorf("--output-dir flag requires a non-empty value")
			}

		case strings.HasPrefix(arg, "--"):
			// Unknown flag - fail fast with clear error message
			return nil, fmt.Errorf("unknown flag: %s", arg)

		default:
			// This is a positional argument
			if !seenInstructions {
				instructionsFile = arg
				seenInstructions = true
			} else {
				targetPaths = append(targetPaths, arg)
			}
		}
	}

	// If help is requested, bypass validation
	if flags&FlagHelp != 0 {
		return &SimplifiedConfig{
			Flags: flags,
		}, nil
	}

	// Validate we have the required positional arguments
	if instructionsFile == "" {
		return nil, fmt.Errorf("instructions file required")
	}
	if len(targetPaths) == 0 {
		return nil, fmt.Errorf("at least one target path required")
	}

	// Join multiple paths with spaces for SimplifiedConfig
	// This maintains the 33-byte struct while supporting multiple paths
	targetPath := strings.Join(targetPaths, " ")

	config := &SimplifiedConfig{
		InstructionsFile: instructionsFile,
		TargetPath:       targetPath,
		Flags:            flags,
	}

	// Skip validation if help is requested
	if !config.HelpRequested() {
		// Validate the parsed configuration
		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("invalid configuration: %w", err)
		}
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
