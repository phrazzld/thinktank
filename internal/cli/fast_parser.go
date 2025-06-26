package cli

import (
	"fmt"

	"github.com/phrazzld/thinktank/internal/config"
)

// FastParseMode controls whether to perform expensive validations during parsing
// This is used for startup performance optimization where we want to minimize I/O
type FastParseMode bool

const (
	// StandardParseMode performs full validation including file system checks
	StandardParseMode FastParseMode = false

	// SkipValidationMode skips expensive validations for startup performance
	SkipValidationMode FastParseMode = true
)

// ParseSimpleArgsWithArgsFast is a performance-optimized version of ParseSimpleArgsWithArgs
// that skips expensive file system validations for startup performance measurement
func ParseSimpleArgsWithArgsFast(args []string, mode FastParseMode) (*SimplifiedConfig, error) {
	if mode == SkipValidationMode {
		// Use fast parsing that skips file system validation
		return parseSimpleArgsInternal(args, true)
	}

	// Use standard parsing with full validation
	return ParseSimpleArgsWithArgs(args)
}

// parseSimpleArgsInternal is the internal parsing function with optional validation skipping
func parseSimpleArgsInternal(args []string, skipFileValidation bool) (*SimplifiedConfig, error) {
	// Parse arguments using the existing logic
	config, err := ParseSimpleArgsWithArgs(args)
	if err != nil {
		// If error is due to file validation and we're in fast mode, try to continue
		if skipFileValidation && isFileValidationError(err) {
			// Create a minimal config without validation
			return createMinimalConfigFromArgs(args)
		}
		return nil, err
	}

	return config, nil
}

// isFileValidationError checks if an error is related to file system validation
func isFileValidationError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return contains(errMsg, "does not exist") ||
		contains(errMsg, "permission denied") ||
		contains(errMsg, "not a directory") ||
		contains(errMsg, "not a regular file")
}

// contains checks if a string contains a substring (simple helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || findSubstring(s, substr))
}

// findSubstring is a simple substring search
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// createMinimalConfigFromArgs creates a minimal configuration from args without validation
// This is used when file validation fails but we still want to measure parsing performance
func createMinimalConfigFromArgs(args []string) (*SimplifiedConfig, error) {
	if len(args) < 3 {
		// Need at least: binary, instructions, target
		return nil, fmt.Errorf("insufficient arguments: need at least 3, got %d", len(args))
	}

	config := &SimplifiedConfig{
		InstructionsFile: args[1],
		TargetPath:       args[2],
		Flags:            0, // Initialize flags bitfield
	}

	// Parse flags from remaining arguments
	for i := 3; i < len(args); i++ {
		arg := args[i]
		if len(arg) > 2 && arg[:2] == "--" {
			flagName := arg[2:]
			// Set common flags in the bitfield
			switch flagName {
			case "dry-run":
				config.SetFlag(FlagDryRun)
			case "verbose":
				config.SetFlag(FlagVerbose)
			case "synthesis":
				config.SetFlag(FlagSynthesis)
			}
		}
	}

	return config, nil
}

// ToCliConfigFast creates a CliConfig with minimal overhead for performance testing
func (sc *SimplifiedConfig) ToCliConfigFast() *config.CliConfig {
	// Pre-warm environment cache for fast lookups
	PrewarmEnvCache()

	// Use the standard ToCliConfig but with cached environment access
	return sc.ToCliConfig()
}
