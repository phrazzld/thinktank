package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// ParsingMode represents the parsing strategy to use for command-line arguments
type ParsingMode int

const (
	// SimplifiedMode uses the new simplified parser (positional arguments)
	SimplifiedMode ParsingMode = iota
	// ComplexMode uses the existing complex parser (flag-based arguments)
	ComplexMode
)

// String returns a string representation of the parsing mode
func (pm ParsingMode) String() string {
	switch pm {
	case SimplifiedMode:
		return "simplified"
	case ComplexMode:
		return "complex"
	default:
		return "unknown"
	}
}

// ParserRouter provides intelligent routing between simplified and complex parsers
// with backward compatibility and deprecation warnings.
type ParserRouter struct {
	logger    logutil.LoggerInterface
	telemetry *DeprecationTelemetry
	getenv    func(string) string
}

// NewParserRouter creates a new parser router with the given logger
func NewParserRouter(logger logutil.LoggerInterface) *ParserRouter {
	return &ParserRouter{
		logger: logger,
		getenv: os.Getenv,
	}
}

// NewParserRouterWithTelemetry creates a new parser router with telemetry enabled
func NewParserRouterWithTelemetry(logger logutil.LoggerInterface, enabled bool) *ParserRouter {
	router := &ParserRouter{
		logger: logger,
		getenv: os.Getenv,
	}

	if enabled {
		router.telemetry = NewDeprecationTelemetry()
	}

	return router
}

// NewParserRouterWithEnv creates a new parser router with custom environment function
func NewParserRouterWithEnv(logger logutil.LoggerInterface, getenv func(string) string) *ParserRouter {
	return &ParserRouter{
		logger: logger,
		getenv: getenv,
	}
}

// GetTelemetry returns the telemetry collector if enabled, nil otherwise
func (pr *ParserRouter) GetTelemetry() *DeprecationTelemetry {
	return pr.telemetry
}

// detectParsingMode determines which parser to use based on argument patterns.
// Heuristics (following Rob Pike's simplicity principle):
// 1. If first non-binary arg doesn't start with '-' AND second arg doesn't start with '-', use simplified
// 2. If first arg starts with '--' (complex flag), use complex
// 3. Default to complex for backward compatibility
func (pr *ParserRouter) detectParsingMode(args []string) ParsingMode {
	// Handle empty arguments
	if len(args) == 0 {
		return ComplexMode
	}

	// Skip binary name if present (binary names typically don't start with '-')
	processArgs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		// Check if this looks like a binary name (contains no file extension or is an executable)
		arg0 := args[0]
		if !strings.Contains(arg0, ".") || strings.HasSuffix(arg0, ".exe") || strings.Contains(arg0, "/") {
			processArgs = args[1:]
		}
	}

	// Need at least 2 arguments for simplified mode detection
	if len(processArgs) < 2 {
		return ComplexMode
	}

	// If first argument is a complex flag, definitely complex mode
	if strings.HasPrefix(processArgs[0], "--") {
		return ComplexMode
	}

	// Simplified mode heuristic: first two args are positional (not flags)
	if !strings.HasPrefix(processArgs[0], "-") && !strings.HasPrefix(processArgs[1], "-") {
		// Additional validation: first arg should look like instructions file
		if strings.HasSuffix(strings.ToLower(processArgs[0]), ".txt") ||
			strings.HasSuffix(strings.ToLower(processArgs[0]), ".md") {
			return SimplifiedMode
		}
	}

	// Default to complex mode for backward compatibility
	return ComplexMode
}

// ParseResult represents the outcome of parsing with metadata about the parsing strategy
type ParseResult struct {
	Config      *config.CliConfig
	Mode        ParsingMode
	Error       error
	Deprecation *DeprecationWarning
}

// DeprecationWarning contains information about deprecated usage patterns
type DeprecationWarning struct {
	Message    string
	Suggestion string
	FlagUsed   string
}

// ParseArguments intelligently routes to the appropriate parser and returns a unified result.
// This function embodies the migration strategy: detect mode, parse, convert, and warn.
func (pr *ParserRouter) ParseArguments(args []string) *ParseResult {
	mode := pr.detectParsingMode(args)

	// Log the detected parsing mode for analytics
	if pr.logger != nil {
		pr.logger.Debug("Detected parsing mode", "mode", mode.String(), "args_count", len(args))
	}

	result := &ParseResult{
		Mode: mode,
	}

	switch mode {
	case SimplifiedMode:
		result = pr.parseSimplified(args, result)
	case ComplexMode:
		result = pr.parseComplex(args, result)
	}

	// Add deprecation warnings for complex mode usage (even on errors)
	if mode == ComplexMode {
		result.Deprecation = pr.generateDeprecationWarning(args)
	}

	return result
}

// parseSimplified handles parsing using the simplified parser
func (pr *ParserRouter) parseSimplified(args []string, result *ParseResult) *ParseResult {
	simplifiedConfig, err := ParseSimpleArgsWithArgs(args)
	if err != nil {
		result.Error = fmt.Errorf("simplified parsing failed: %w", err)
		return result
	}

	// Convert to complex config for compatibility
	adapter := NewConfigAdapter(simplifiedConfig)
	result.Config = adapter.ToComplexConfig()

	if pr.logger != nil {
		pr.logger.Debug("Successfully parsed using simplified parser",
			"instructions", simplifiedConfig.InstructionsFile,
			"target", simplifiedConfig.TargetPath,
			"flags", simplifiedConfig.Flags)
	}

	return result
}

// parseComplex handles parsing using the existing complex parser
func (pr *ParserRouter) parseComplex(args []string, result *ParseResult) *ParseResult {
	config, err := ParseFlagsWithArgsAndEnv(args, pr.getenv)
	if err != nil {
		result.Error = fmt.Errorf("complex parsing failed: %w", err)
		return result
	}

	result.Config = config

	// Record telemetry if enabled
	if pr.telemetry != nil {
		pr.recordComplexFlagUsage(args)
	}

	if pr.logger != nil {
		pr.logger.Debug("Successfully parsed using complex parser",
			"instructions", config.InstructionsFile,
			"paths", len(config.Paths),
			"model_count", len(config.ModelNames))
	}

	return result
}

// generateDeprecationWarning creates a deprecation warning for complex flag usage
func (pr *ParserRouter) generateDeprecationWarning(args []string) *DeprecationWarning {
	// Look for common deprecated patterns - prioritize specific warnings
	for _, arg := range args {
		switch {
		case arg == "--instructions":
			return &DeprecationWarning{
				Message:    "The --instructions flag is deprecated",
				Suggestion: "Use positional arguments: thinktank instructions.txt target_path [flags...]",
				FlagUsed:   "--instructions",
			}
		case strings.HasPrefix(arg, "--instructions="):
			return &DeprecationWarning{
				Message:    "The --instructions=value format is deprecated",
				Suggestion: "Use positional arguments: thinktank instructions.txt target_path [flags...]",
				FlagUsed:   "--instructions=",
			}
		}
	}

	// Check for flag-only interface (no positional args)
	if !containsPositionalArgs(args) && containsComplexFlags(args) {
		// Find which flag to highlight for the warning
		for _, arg := range args {
			if arg == "--output-dir" {
				return &DeprecationWarning{
					Message:    "Flag-only interface is deprecated",
					Suggestion: "Use simplified interface: thinktank instructions.txt target_path --output-dir output/",
					FlagUsed:   "--output-dir",
				}
			}
		}
		// If no specific flag found, use general flag-only deprecation
		return &DeprecationWarning{
			Message:    "Complex flag interface will be deprecated in future versions",
			Suggestion: "Consider using the simplified interface: thinktank instructions.txt target_path [optional_flags...]",
			FlagUsed:   "complex_flags",
		}
	}

	// General deprecation for complex usage (only if has positional args + complex flags)
	// But skip if this looks like simplified mode with some complex flags mixed in
	if containsPositionalArgs(args) && containsComplexFlags(args) {
		// Check if this looks like simplified mode pattern: instructions.txt target_path [flags...]
		if looksLikeSimplifiedModePattern(args) {
			return nil // No warning for simplified mode with occasional complex flags
		}

		return &DeprecationWarning{
			Message:    "Complex flag interface will be deprecated in future versions",
			Suggestion: "Consider using the simplified interface: thinktank instructions.txt target_path [optional_flags...]",
			FlagUsed:   "complex_flags",
		}
	}

	return nil
}

// containsPositionalArgs checks if the arguments contain positional arguments
// This is smarter than just checking for non-dash args - it considers flag values
func containsPositionalArgs(args []string) bool {
	// Skip binary name if present
	processArgs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		// Check if this looks like a binary name
		arg0 := args[0]
		if !strings.Contains(arg0, ".") || strings.HasSuffix(arg0, ".exe") || strings.Contains(arg0, "/") {
			processArgs = args[1:]
		}
	}

	// Track flags that expect values to avoid counting their values as positional
	skipNext := false
	flagsWithValues := map[string]bool{
		"--instructions": true, "--output-dir": true, "--model": true,
		"--include": true, "--exclude": true, "--exclude-names": true,
		"--confirm-tokens": true, "--log-level": true, "--audit-log-file": true,
	}

	for i, arg := range processArgs {
		if skipNext {
			skipNext = false
			continue
		}

		if strings.HasPrefix(arg, "-") {
			// Check if this flag expects a value
			if flagsWithValues[arg] && i+1 < len(processArgs) {
				skipNext = true // Skip the next argument (flag value)
			}
		} else {
			// This is a non-flag argument that's not a flag value
			return true
		}
	}
	return false
}

// looksLikeSimplifiedModePattern checks if args follow simplified mode pattern:
// instructions.txt target_path [flags...]
func looksLikeSimplifiedModePattern(args []string) bool {
	// Skip binary name if present
	processArgs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		arg0 := args[0]
		if !strings.Contains(arg0, ".") || strings.HasSuffix(arg0, ".exe") || strings.Contains(arg0, "/") {
			processArgs = args[1:]
		}
	}

	if len(processArgs) < 2 {
		return false
	}

	// First arg should be instructions file (.txt or .md)
	firstArg := processArgs[0]
	if strings.HasPrefix(firstArg, "-") {
		return false
	}
	if !strings.HasSuffix(strings.ToLower(firstArg), ".txt") &&
		!strings.HasSuffix(strings.ToLower(firstArg), ".md") {
		return false
	}

	// Second arg should be target path (not a flag)
	secondArg := processArgs[1]
	if strings.HasPrefix(secondArg, "-") {
		return false
	}

	// Everything after that should be flags or flag values
	return true
}

// containsComplexFlags checks if the arguments contain complex flags
func containsComplexFlags(args []string) bool {
	complexFlags := []string{
		"--instructions", "--output-dir", "--model", "--include", "--exclude",
		"--exclude-names", "--confirm-tokens", "--log-level", "--audit-log-file",
		"--force-overwrite",
	}

	for _, arg := range args {
		for _, flag := range complexFlags {
			if arg == flag || strings.HasPrefix(arg, flag+"=") {
				return true
			}
		}
	}
	return false
}

// LogDeprecationWarning logs a deprecation warning if present
func (pr *ParserRouter) LogDeprecationWarning(warning *DeprecationWarning, config *config.CliConfig) {
	if warning == nil || pr.logger == nil {
		return
	}

	// Determine if warnings should be suppressed with clear precedence:
	// CLI flag > environment variable > default (false)
	shouldSuppress := false
	if config != nil {
		if config.SuppressDeprecationWarnings {
			// CLI flag explicitly set to suppress
			shouldSuppress = true
		} else {
			// CLI flag is false - check environment variable for backward compatibility
			shouldSuppress = os.Getenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS") != ""
		}
	} else {
		// No config - fall back to environment variable
		shouldSuppress = os.Getenv("THINKTANK_SUPPRESS_DEPRECATION_WARNINGS") != ""
	}

	if shouldSuppress {
		// Still log to structured logs for analytics, but skip stderr output
		pr.logger.Debug("Deprecation warning suppressed: %s (flag: %s, suggestion: %s)",
			warning.Message, warning.FlagUsed, warning.Suggestion)
		return
	}

	// Use format string logging for compatibility
	pr.logger.Warn("Deprecation warning: %s (flag: %s, suggestion: %s)",
		warning.Message, warning.FlagUsed, warning.Suggestion)

	// Also write to stderr for immediate user visibility
	fmt.Fprintf(os.Stderr, "DEPRECATION WARNING: %s\n", warning.Message)
	fmt.Fprintf(os.Stderr, "Suggestion: %s\n", warning.Suggestion)
	fmt.Fprintf(os.Stderr, "\n")
}

// IsSuccess returns true if parsing succeeded without errors
func (pr *ParseResult) IsSuccess() bool {
	return pr.Error == nil && pr.Config != nil
}

// HasDeprecationWarning returns true if there are deprecation warnings
func (pr *ParseResult) HasDeprecationWarning() bool {
	return pr.Deprecation != nil
}

// recordComplexFlagUsage records usage of complex flags for telemetry analysis
func (pr *ParserRouter) recordComplexFlagUsage(args []string) {
	if pr.telemetry == nil {
		return
	}

	// Track overall complex flag usage
	pr.telemetry.RecordUsage("complex_flags")

	// Record specific deprecated flags
	deprecatedFlags := []string{
		"--instructions", "--output-dir", "--model", "--include", "--exclude",
		"--exclude-names", "--log-level", "--audit-log-file",
	}

	for _, arg := range args {
		for _, flag := range deprecatedFlags {
			if arg == flag || strings.HasPrefix(arg, flag+"=") {
				pr.telemetry.RecordUsage(flag)
			}
		}
	}

	// Record usage pattern for migration analysis
	pattern := strings.Join(args, " ")
	pr.telemetry.RecordUsagePattern(pattern, args)
}
