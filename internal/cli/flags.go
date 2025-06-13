// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// stringSliceFlag is a slice of strings that implements flag.Value interface
// to handle repeatable flags for multiple values
type stringSliceFlag []string

// String implements the flag.Value interface
func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ",")
}

// Set implements the flag.Value interface
func (s *stringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// ValidateInputs validates the configuration and inputs before executing the main logic
func ValidateInputs(config *config.CliConfig, logger logutil.LoggerInterface) error {
	return ValidateInputsWithEnv(config, logger, os.Getenv)
}

// ValidateInputsWithEnv validates the configuration and inputs with a custom environment getter
func ValidateInputsWithEnv(config *config.CliConfig, logger logutil.LoggerInterface, getenv func(string) string) error {
	// Check for instructions file
	if config.InstructionsFile == "" && !config.DryRun {
		logger.Error("The required --instructions flag is missing.")
		return fmt.Errorf("missing required --instructions flag")
	}

	// Check for input paths
	if len(config.Paths) == 0 {
		logger.Error("At least one file or directory path must be provided as an argument.")
		return fmt.Errorf("no paths specified")
	}

	// Check for model names
	if len(config.ModelNames) == 0 && !config.DryRun {
		logger.Error("At least one model must be specified with --model flag.")
		return fmt.Errorf("no models specified")
	}

	// Validate synthesis model if provided
	if config.SynthesisModel != "" {
		logger.Debug("Validating synthesis model: %s", config.SynthesisModel)
		// Basic model validation based on naming patterns
		isLikelyValid := false
		if strings.HasPrefix(strings.ToLower(config.SynthesisModel), "gpt-") ||
			strings.HasPrefix(strings.ToLower(config.SynthesisModel), "text-") ||
			strings.HasPrefix(strings.ToLower(config.SynthesisModel), "gemini-") ||
			strings.HasPrefix(strings.ToLower(config.SynthesisModel), "claude-") ||
			strings.Contains(strings.ToLower(config.SynthesisModel), "openai") ||
			strings.Contains(strings.ToLower(config.SynthesisModel), "openrouter/") {
			isLikelyValid = true
		}

		if !isLikelyValid {
			logger.Error("Invalid synthesis model name pattern: '%s'", config.SynthesisModel)
			return fmt.Errorf("invalid synthesis model: '%s' does not match any known model pattern", config.SynthesisModel)
		}
	}

	return nil
}

// ParseFlags parses command line flags and returns a CliConfig
func ParseFlags() (*config.CliConfig, error) {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	return ParseFlagsWithEnv(flagSet, os.Args[1:], os.Getenv)
}

// ParseFlagsWithEnv parses command line flags with custom environment and flag set
func ParseFlagsWithEnv(flagSet *flag.FlagSet, args []string, getenv func(string) string) (*config.CliConfig, error) {
	cfg := config.NewDefaultCliConfig()

	// Define flags
	instructionsFileFlag := flagSet.String("instructions", "", "Path to a file containing the static instructions for the LLM.")
	outputDirFlag := flagSet.String("output-dir", "", "Directory path to store generated plans (one per model).")
	synthesisModelFlag := flagSet.String("synthesis-model", "", "Optional: Model to use for synthesizing results from multiple models.")
	verboseFlag := flagSet.Bool("verbose", false, "Enable verbose logging output (shorthand for --log-level=debug).")
	logLevelFlag := flagSet.String("log-level", "info", "Set logging level (debug, info, warn, error).")
	includeFlag := flagSet.String("include", "", "Comma-separated list of file extensions to include (e.g., .go,.md)")
	excludeFlag := flagSet.String("exclude", config.DefaultExcludes, "Comma-separated list of file extensions to exclude.")
	excludeNamesFlag := flagSet.String("exclude-names", config.DefaultExcludeNames, "Comma-separated list of file/dir names to exclude.")
	formatFlag := flagSet.String("format", config.DefaultFormat, "Format string for each file. Use {path} and {content}.")
	dryRunFlag := flagSet.Bool("dry-run", false, "Show files that would be included and token count, but don't call the API.")
	auditLogFileFlag := flagSet.String("audit-log-file", "", "Path to write structured audit logs (JSON Lines). Disabled if empty.")
	partialSuccessOkFlag := flagSet.Bool("partial-success-ok", false, "Return exit code 0 if any model succeeds and a synthesis file is generated, even if some models fail.")

	// Rate limiting flags
	maxConcurrentFlag := flagSet.Int("max-concurrent", 5, "Maximum number of concurrent API requests (0 = no limit)")
	rateLimitRPMFlag := flagSet.Int("rate-limit", 60, "Maximum requests per minute (RPM) per model (0 = no limit)")

	// Timeout flag
	timeoutFlag := flagSet.Duration("timeout", config.DefaultTimeout, "Global timeout for the entire operation (e.g., 60s, 2m, 1h)")

	// Permission flags
	dirPermFlag := flagSet.String("dir-permissions", fmt.Sprintf("%#o", config.DefaultDirPermissions), "Directory creation permissions (octal, e.g., 0750)")
	filePermFlag := flagSet.String("file-permissions", fmt.Sprintf("%#o", config.DefaultFilePermissions), "File creation permissions (octal, e.g., 0640)")

	// Define the model flag using our custom stringSliceFlag type to support multiple values
	modelFlag := &stringSliceFlag{}
	flagSet.Var(modelFlag, "model", fmt.Sprintf("Model to use for generation (repeatable). Can be Gemini (e.g., %s) or OpenAI (e.g., gpt-4) models. Default: %s", config.DefaultModel, config.DefaultModel))

	// Parse the flags
	if err := flagSet.Parse(args); err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	// Store flag values in configuration
	cfg.InstructionsFile = *instructionsFileFlag
	cfg.OutputDir = *outputDirFlag
	cfg.SynthesisModel = *synthesisModelFlag
	cfg.AuditLogFile = *auditLogFileFlag
	cfg.Verbose = *verboseFlag
	cfg.Include = *includeFlag
	cfg.Exclude = *excludeFlag
	cfg.ExcludeNames = *excludeNamesFlag
	cfg.Format = *formatFlag
	cfg.DryRun = *dryRunFlag
	cfg.PartialSuccessOk = *partialSuccessOkFlag
	cfg.Paths = flagSet.Args()

	// Store rate limiting configuration
	cfg.MaxConcurrentRequests = *maxConcurrentFlag
	cfg.RateLimitRequestsPerMinute = *rateLimitRPMFlag

	// Store timeout configuration
	cfg.Timeout = *timeoutFlag

	// Parse and store permissions
	if dirPerm, err := parseOctalPermission(*dirPermFlag); err == nil {
		cfg.DirPermissions = dirPerm
	} else {
		return nil, fmt.Errorf("invalid directory permission format: %w", err)
	}

	if filePerm, err := parseOctalPermission(*filePermFlag); err == nil {
		cfg.FilePermissions = filePerm
	} else {
		return nil, fmt.Errorf("invalid file permission format: %w", err)
	}

	// Set model names from the flag, defaulting to a single default model if none provided
	if len(*modelFlag) > 0 {
		cfg.ModelNames = *modelFlag
	} else {
		// If no models were specified on the command line, use the default model
		cfg.ModelNames = []string{config.DefaultModel}
	}

	// Determine initial log level from flag
	parsedLogLevel := logutil.InfoLevel // Default
	if *logLevelFlag != "info" {
		ll, err := logutil.ParseLogLevel(*logLevelFlag)
		if err == nil {
			parsedLogLevel = ll
		}
	}
	cfg.LogLevel = parsedLogLevel

	// Apply verbose override *after* parsing the specific level
	if cfg.Verbose {
		cfg.LogLevel = logutil.DebugLevel
	}

	cfg.APIEndpoint = getenv(config.APIEndpointEnvVar)

	return cfg, nil
}

// parseOctalPermission converts a string representation of an octal permission
// to an os.FileMode
func parseOctalPermission(permStr string) (os.FileMode, error) {
	// Parse the octal permission string
	n, err := strconv.ParseUint(strings.TrimPrefix(permStr, "0"), 8, 32)
	if err != nil {
		return 0, err
	}
	return os.FileMode(n), nil
}

// SetupLogging configures and returns a logger based on the configuration
func SetupLogging(config *config.CliConfig) logutil.LoggerInterface {
	// Use the log level that was already parsed in ParseFlags
	logLevel := config.LogLevel

	// Verbose flag overrides log level (should already be handled in ParseFlags, but double-check)
	if config.Verbose {
		logLevel = logutil.DebugLevel
	}

	// Create a structured logger with stream separation for CLI usage
	return logutil.NewSlogLoggerWithStreamSeparationFromLogLevel(os.Stdout, os.Stderr, logLevel)
}
