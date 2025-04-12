// Package architect provides the command-line interface for the architect tool
package architect

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/logutil"
)

// stringSliceFlag is a slice of strings that implements flag.Value interface
// to handle repeatable flags for multiple values
type stringSliceFlag []string

// String returns a comma-separated string representation of the flag values
// This method is required by the flag.Value interface
func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ",")
}

// Set appends the value to the slice of values
// This method is required by the flag.Value interface
func (s *stringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// Constants referencing the config package defaults
const (
	defaultOutputFile   = config.DefaultOutputFile
	defaultModel        = config.DefaultModel
	apiKeyEnvVar        = config.APIKeyEnvVar
	defaultFormat       = config.DefaultFormat
	defaultExcludes     = config.DefaultExcludes
	defaultExcludeNames = config.DefaultExcludeNames
)

// CliConfig holds the parsed command-line options
type CliConfig struct {
	InstructionsFile string
	OutputDir        string
	AuditLogFile     string // Path to write structured audit logs (JSON Lines)
	ModelNames       []string
	Verbose          bool
	LogLevel         logutil.LogLevel
	Include          string
	Exclude          string
	ExcludeNames     string
	Format           string
	DryRun           bool
	ConfirmTokens    int
	Paths            []string
	ApiKey           string
}

// ValidateInputs checks if the configuration is valid and returns an error if not
func ValidateInputs(config *CliConfig, logger logutil.LoggerInterface) error {
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

	// Check for API key
	if config.ApiKey == "" {
		logger.Error("%s environment variable not set.", apiKeyEnvVar)
		return fmt.Errorf("API key not set")
	}

	// Check for model names
	if len(config.ModelNames) == 0 && !config.DryRun {
		logger.Error("At least one model must be specified with --model flag.")
		return fmt.Errorf("no models specified")
	}

	return nil
}

// ParseFlags handles command line argument parsing and returns the configuration
func ParseFlags() (*CliConfig, error) {
	return ParseFlagsWithEnv(flag.CommandLine, os.Args[1:], os.Getenv)
}

// ParseFlagsWithEnv handles command-line flag parsing with custom flag set and environment lookup
// This improves testability by allowing tests to provide mock flag sets and environment functions
func ParseFlagsWithEnv(flagSet *flag.FlagSet, args []string, getenv func(string) string) (*CliConfig, error) {
	config := &CliConfig{}

	// Define flags
	instructionsFileFlag := flagSet.String("instructions", "", "Path to a file containing the static instructions for the LLM.")
	outputDirFlag := flagSet.String("output-dir", "", "Directory path to store generated plans (one per model).")
	verboseFlag := flagSet.Bool("verbose", false, "Enable verbose logging output (shorthand for --log-level=debug).")
	logLevelFlag := flagSet.String("log-level", "info", "Set logging level (debug, info, warn, error).")
	includeFlag := flagSet.String("include", "", "Comma-separated list of file extensions to include (e.g., .go,.md)")
	excludeFlag := flagSet.String("exclude", defaultExcludes, "Comma-separated list of file extensions to exclude.")
	excludeNamesFlag := flagSet.String("exclude-names", defaultExcludeNames, "Comma-separated list of file/dir names to exclude.")
	formatFlag := flagSet.String("format", defaultFormat, "Format string for each file. Use {path} and {content}.")
	dryRunFlag := flagSet.Bool("dry-run", false, "Show files that would be included and token count, but don't call the API.")
	confirmTokensFlag := flagSet.Int("confirm-tokens", 0, "Prompt for confirmation if token count exceeds this value (0 = never prompt)")
	auditLogFileFlag := flagSet.String("audit-log-file", "", "Path to write structured audit logs (JSON Lines). Disabled if empty.")

	// Define the model flag using our custom stringSliceFlag type to support multiple values
	modelFlag := &stringSliceFlag{}
	flagSet.Var(modelFlag, "model", fmt.Sprintf("Gemini model to use for generation (repeatable). Default: %s", defaultModel))

	// Set custom usage message
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s --instructions <file> [options] <path1> [path2...]\n\n", os.Args[0])

		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <path1> [path2...]   One or more file or directory paths for project context.\n\n")

		fmt.Fprintf(os.Stderr, "Example Commands:\n")
		fmt.Fprintf(os.Stderr, "  %s --instructions instructions.txt ./src                        Generate plan using default model\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --instructions instructions.txt --output-dir custom-dir ./       Generate plans in custom directory\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --instructions instructions.txt --model model1 --model model2 ./  Generate plans for multiple models\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --dry-run ./                                                 Show files without generating plan\n\n", os.Args[0])

		fmt.Fprintf(os.Stderr, "Options:\n")
		flagSet.PrintDefaults()

		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s: Required. Your Google AI Gemini API key.\n", apiKeyEnvVar)
	}

	// Parse the flags
	if err := flagSet.Parse(args); err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	// Store flag values in configuration
	config.InstructionsFile = *instructionsFileFlag

	// Set output directory
	config.OutputDir = *outputDirFlag

	config.AuditLogFile = *auditLogFileFlag
	config.Verbose = *verboseFlag
	config.Include = *includeFlag
	config.Exclude = *excludeFlag
	config.ExcludeNames = *excludeNamesFlag
	config.Format = *formatFlag
	config.DryRun = *dryRunFlag
	config.ConfirmTokens = *confirmTokensFlag
	config.Paths = flagSet.Args()

	// Set model names from the flag, defaulting to a single default model if none provided
	if len(*modelFlag) > 0 {
		config.ModelNames = *modelFlag
	} else {
		// If no models were specified on the command line, use the default model
		config.ModelNames = []string{defaultModel}
	}

	// Determine initial log level from flag
	parsedLogLevel := logutil.InfoLevel // Default
	if *logLevelFlag != "info" {
		ll, err := logutil.ParseLogLevel(*logLevelFlag)
		if err == nil {
			parsedLogLevel = ll
		}
	}
	config.LogLevel = parsedLogLevel

	// Apply verbose override *after* parsing the specific level
	if config.Verbose {
		config.LogLevel = logutil.DebugLevel
	}
	config.ApiKey = getenv(apiKeyEnvVar)

	// ParseFlagsWithEnv no longer does logical validation (just parsing errors)
	// Validation is now exclusively handled by ValidateInputs
	return config, nil
}

// SetupLogging initializes the logger based on configuration
func SetupLogging(config *CliConfig) logutil.LoggerInterface {
	return SetupLoggingCustom(config, flag.Lookup("log-level"), os.Stderr)
}

// SetupLoggingCustom initializes the logger with custom flag and writer for testing
func SetupLoggingCustom(config *CliConfig, _ *flag.Flag, output io.Writer) logutil.LoggerInterface {
	// Apply verbose override if set
	if config.Verbose {
		config.LogLevel = logutil.DebugLevel
	}

	// Use the LogLevel set in the config
	logger := logutil.NewLogger(config.LogLevel, output, "[architect] ")
	return logger
}
