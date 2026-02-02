// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/misty-step/thinktank/internal/models"
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
	var metricsOutput string
	flags := uint8(0)
	safetyMargin := uint8(10) // Default 10% safety margin

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
				return nil, fmt.Errorf("--model flag requires a value%s", getModelSuggestion())
			}
			i++ // Skip the model value - we use smart default in ToCliConfig()

		case arg == "--output-dir":
			// --output-dir flag requires a value
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--output-dir flag requires a value")
			}
			i++ // Skip the output dir value - we use smart default in ToCliConfig()

		case arg == "--token-safety-margin":
			// --token-safety-margin flag requires a value
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--token-safety-margin flag requires a value")
			}
			i++ // Move to next argument
			marginValue := args[i]
			parsedMargin, err := parseAndValidateSafetyMargin(marginValue)
			if err != nil {
				return nil, fmt.Errorf("invalid --token-safety-margin value: %w", err)
			}
			safetyMargin = parsedMargin

		case strings.HasPrefix(arg, "--model="):
			// Handle --model=value format
			value := strings.TrimPrefix(arg, "--model=")
			if value == "" {
				return nil, fmt.Errorf("--model flag requires a non-empty value%s", getModelSuggestion())
			}

		case strings.HasPrefix(arg, "--output-dir="):
			// Handle --output-dir=value format
			value := strings.TrimPrefix(arg, "--output-dir=")
			if value == "" {
				return nil, fmt.Errorf("--output-dir flag requires a non-empty value")
			}

		case strings.HasPrefix(arg, "--token-safety-margin="):
			// Handle --token-safety-margin=value format
			value := strings.TrimPrefix(arg, "--token-safety-margin=")
			if value == "" {
				return nil, fmt.Errorf("--token-safety-margin flag requires a non-empty value")
			}
			parsedMargin, err := parseAndValidateSafetyMargin(value)
			if err != nil {
				return nil, fmt.Errorf("invalid --token-safety-margin value: %w", err)
			}
			safetyMargin = parsedMargin

		case arg == "--metrics-output":
			// --metrics-output flag requires a value
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--metrics-output flag requires a value")
			}
			i++
			metricsOutput = args[i]

		case strings.HasPrefix(arg, "--metrics-output="):
			// Handle --metrics-output=value format
			value := strings.TrimPrefix(arg, "--metrics-output=")
			if value == "" {
				return nil, fmt.Errorf("--metrics-output flag requires a non-empty value")
			}
			metricsOutput = value

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
		MetricsOutput:    metricsOutput,
		Flags:            flags,
		SafetyMargin:     safetyMargin,
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

// getModelSuggestion returns a formatted suggestion of popular models
func getModelSuggestion() string {
	popularModels := models.GetCoreCouncilModels()
	suggestion := "\n\nPopular models:\n"
	limit := 5
	if len(popularModels) < limit {
		limit = len(popularModels)
	}
	for i := 0; i < limit; i++ {
		suggestion += fmt.Sprintf("  - %s\n", popularModels[i])
	}
	suggestion += "\nSee available models at: https://openrouter.ai/models"
	return suggestion
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

// parseAndValidateSafetyMargin parses and validates a safety margin value.
// The safety margin represents the percentage of context window reserved for output tokens.
// Valid range: 0-50% (0 = no safety margin, 50 = half context reserved for output).
// This prevents token overflow and ensures models have adequate output space.
func parseAndValidateSafetyMargin(value string) (uint8, error) {
	// Parse the string as an integer
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid syntax: %w", err)
	}

	// Validate range (0% to 50%)
	if parsed < 0 || parsed > 50 {
		return 0, fmt.Errorf("safety margin must be between 0%% and 50%%, got %d%%", parsed)
	}

	return uint8(parsed), nil
}
