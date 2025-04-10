// Package architect provides the command-line interface for the architect tool
package architect

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/logutil"
)

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
	TaskDescription string
	TaskFile        string
	OutputFile      string
	ModelName       string
	Verbose         bool
	LogLevel        logutil.LogLevel
	UseColors       bool
	Include         string
	Exclude         string
	ExcludeNames    string
	Format          string
	DryRun          bool
	ConfirmTokens   int
	PromptTemplate  string
	ClarifyTask     bool
	Paths           []string
	ApiKey          string
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
	taskFlag := flagSet.String("task", "", "Description of the task or goal for the plan.")
	taskFileFlag := flagSet.String("task-file", "", "Path to a file containing the task description (alternative to --task).")
	outputFileFlag := flagSet.String("output", defaultOutputFile, "Output file path for the generated plan.")
	modelNameFlag := flagSet.String("model", defaultModel, "Gemini model to use for generation.")
	verboseFlag := flagSet.Bool("verbose", false, "Enable verbose logging output (shorthand for --log-level=debug).")
	flagSet.String("log-level", "info", "Set logging level (debug, info, warn, error).")
	useColorsFlag := flagSet.Bool("color", true, "Enable/disable colored log output.")
	includeFlag := flagSet.String("include", "", "Comma-separated list of file extensions to include (e.g., .go,.md)")
	excludeFlag := flagSet.String("exclude", defaultExcludes, "Comma-separated list of file extensions to exclude.")
	excludeNamesFlag := flagSet.String("exclude-names", defaultExcludeNames, "Comma-separated list of file/dir names to exclude.")
	formatFlag := flagSet.String("format", defaultFormat, "Format string for each file. Use {path} and {content}.")
	dryRunFlag := flagSet.Bool("dry-run", false, "Show files that would be included and token count, but don't call the API.")
	confirmTokensFlag := flagSet.Int("confirm-tokens", 0, "Prompt for confirmation if token count exceeds this value (0 = never prompt)")
	promptTemplateFlag := flagSet.String("prompt-template", "", "Path to a custom prompt template file (.tmpl)")
	clarifyTaskFlag := flagSet.Bool("clarify", false, "Enable interactive task clarification to refine your task description")

	// Set custom usage message
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s (--task \"<description>\" | --task-file <path>) [options] <path1> [path2...]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <path1> [path2...]   One or more file or directory paths for project context.\n\n")
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
	config.TaskDescription = *taskFlag
	config.TaskFile = *taskFileFlag
	config.OutputFile = *outputFileFlag
	config.ModelName = *modelNameFlag
	config.Verbose = *verboseFlag
	config.UseColors = *useColorsFlag
	config.Include = *includeFlag
	config.Exclude = *excludeFlag
	config.ExcludeNames = *excludeNamesFlag
	config.Format = *formatFlag
	config.DryRun = *dryRunFlag
	config.ConfirmTokens = *confirmTokensFlag
	config.PromptTemplate = *promptTemplateFlag
	config.ClarifyTask = *clarifyTaskFlag
	config.Paths = flagSet.Args()
	config.ApiKey = getenv(apiKeyEnvVar)

	// Basic validation
	if config.TaskDescription == "" && config.TaskFile == "" {
		return nil, fmt.Errorf("missing task description, use --task or --task-file")
	}

	if len(config.Paths) == 0 {
		return nil, fmt.Errorf("no paths specified for project context")
	}

	return config, nil
}

// SetupLogging initializes the logger based on configuration
func SetupLogging(config *CliConfig) logutil.LoggerInterface {
	return SetupLoggingCustom(config, flag.Lookup("log-level"), os.Stderr)
}

// SetupLoggingCustom initializes the logger with custom flag and writer for testing
func SetupLoggingCustom(config *CliConfig, logLevelFlag *flag.Flag, output io.Writer) logutil.LoggerInterface {
	var logLevel logutil.LogLevel

	// Determine log level
	if config.Verbose {
		logLevel = logutil.DebugLevel
	} else if logLevelFlag != nil {
		// Get the log level from the configuration
		logLevelValue := logLevelFlag.Value.String()
		var err error
		logLevel, err = logutil.ParseLogLevel(logLevelValue)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v. Defaulting to 'info' level.\n", err)
			logLevel = logutil.InfoLevel
		}
	} else {
		logLevel = logutil.InfoLevel
	}

	// Store the log level in the config
	config.LogLevel = logLevel

	// Create structured logger
	logger := logutil.NewLogger(logLevel, output, "[architect] ", config.UseColors)

	return logger
}

// ConvertConfigToMap converts a CliConfig to a map for use with config.Manager.MergeWithFlags
func ConvertConfigToMap(cliConfig *CliConfig) map[string]interface{} {
	return map[string]interface{}{
		"task":           cliConfig.TaskDescription,
		"taskFile":       cliConfig.TaskFile,
		"output":         cliConfig.OutputFile,
		"model":          cliConfig.ModelName,
		"verbose":        cliConfig.Verbose,
		"logLevel":       cliConfig.LogLevel.String(),
		"color":          cliConfig.UseColors,
		"include":        cliConfig.Include,
		"exclude":        cliConfig.Exclude,
		"excludeNames":   cliConfig.ExcludeNames,
		"format":         cliConfig.Format,
		"dryRun":         cliConfig.DryRun,
		"confirmTokens":  cliConfig.ConfirmTokens,
		"promptTemplate": cliConfig.PromptTemplate,
		"clarify":        cliConfig.ClarifyTask,
		"apiKey":         cliConfig.ApiKey,
	}
}
