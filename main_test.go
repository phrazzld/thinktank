// main_test.go
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/phrazzld/architect/cmd/architect"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// Re-export the functions and types that were used in tests from the original main.go

// osExit is a variable holding os.Exit function that can be mocked in tests
var osExit = os.Exit

// Constants referencing the config package defaults
const (
	defaultOutputFile       = config.DefaultOutputFile
	defaultModel            = config.DefaultModel
	apiKeyEnvVar            = config.APIKeyEnvVar
	defaultFormat           = config.DefaultFormat
	defaultExcludes         = config.DefaultExcludes
	defaultExcludeNames     = config.DefaultExcludeNames
	taskFlagDescription     = "Description of the task or goal for the plan (deprecated: use --task-file instead)."
	taskFileFlagDescription = "Path to a file containing the task description (required)."
)

// Define sentinel errors for task file validation
var (
	ErrTaskFileNotFound       = errors.New("task file not found")
	ErrTaskFileReadPermission = errors.New("task file permission denied")
	ErrTaskFileIsDir          = errors.New("task file path is a directory")
	ErrTaskFileEmpty          = errors.New("task file is empty")
)

// Configuration holds the parsed command-line options
type Configuration struct {
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
	ListExamples    bool
	ShowExample     string
	Paths           []string
	ApiKey          string
}

// validateInputsResult represents the result of input validation
type validateInputsResult struct {
	Valid        bool
	ErrorMessage string
	WarningsOnly bool
}

// tokenInfoResult and associated functions moved to cmd/architect/token.go
// Re-export for testing with original signature
type tokenInfoResult struct {
	tokenCount    int32
	inputLimit    int32 
	outputLimit   int32
	exceedsLimit  bool
	percentOfMax  float64
	exceededBy    int32
	modelName     string
	countingError error
	infoError     error
	limitError    error
	TokenCount    int32
	ModelInfo     *gemini.ModelInfo
	PercentOfMax  float64
	ExceedsLimit  bool
	ExceededBy    int32
	CountingError error
	InfoError     error
}

// Helper function to check if a tokenInfoResult is nil (for test compatibility)
func (t tokenInfoResult) IsNil() bool {
	return false
}

// Re-export functions that were used in tests
func parseFlags() *Configuration {
	config := &Configuration{}

	// Define flags
	taskFlag := flag.String("task", "", taskFlagDescription)
	taskFileFlag := flag.String("task-file", "", taskFileFlagDescription)
	outputFileFlag := flag.String("output", defaultOutputFile, "Output file path for the generated plan.")
	modelNameFlag := flag.String("model", defaultModel, "Gemini model to use for generation.")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging output (shorthand for --log-level=debug).")
	flag.String("log-level", "info", "Set logging level (debug, info, warn, error).")
	useColorsFlag := flag.Bool("color", true, "Enable/disable colored log output.")
	includeFlag := flag.String("include", "", "Comma-separated list of file extensions to include (e.g., .go,.md)")
	excludeFlag := flag.String("exclude", defaultExcludes, "Comma-separated list of file extensions to exclude.")
	excludeNamesFlag := flag.String("exclude-names", defaultExcludeNames, "Comma-separated list of file/dir names to exclude.")
	formatFlag := flag.String("format", defaultFormat, "Format string for each file. Use {path} and {content}.")
	dryRunFlag := flag.Bool("dry-run", false, "Show files that would be included and token count, but don't call the API.")
	confirmTokensFlag := flag.Int("confirm-tokens", 0, "Prompt for confirmation if token count exceeds this value (0 = never prompt)")
	promptTemplateFlag := flag.String("prompt-template", "", "Path to a custom prompt template file (.tmpl)")
	listExamplesFlag := flag.Bool("list-examples", false, "List available example prompt template files")
	showExampleFlag := flag.String("show-example", "", "Display the content of a specific example template")

	// Set custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s --task-file <path> [options] <path1> [path2...]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <path1> [path2...]   One or more file or directory paths for project context.\n\n")
		fmt.Fprintf(os.Stderr, "Example Commands:\n")
		fmt.Fprintf(os.Stderr, "  %s --list-examples                 List available example templates\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --show-example basic.tmpl       Display the content of a specific example template\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --show-example basic > my.tmpl  Save an example template to a file\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  %s: Required. Your Google AI Gemini API key.\n", apiKeyEnvVar)
	}

	flag.Parse()

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
	config.ListExamples = *listExamplesFlag
	config.ShowExample = *showExampleFlag
	config.Paths = flag.Args()
	config.ApiKey = os.Getenv(apiKeyEnvVar)

	return config
}

func setupLogging(config *Configuration) logutil.LoggerInterface {
	var logLevel logutil.LogLevel

	// Determine log level
	if config.Verbose {
		logLevel = logutil.DebugLevel
	} else {
		// Get the log level from the configuration
		logLevelValue := flag.Lookup("log-level").Value.String()
		var err error
		logLevel, err = logutil.ParseLogLevel(logLevelValue)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v. Defaulting to 'info' level.\n", err)
			logLevel = logutil.InfoLevel
		}
	}

	// Store the log level in the config
	config.LogLevel = logLevel

	// Create structured logger
	logger := logutil.NewLogger(logLevel, os.Stderr, "[architect] ", config.UseColors)

	// Configure global logger
	log.SetOutput(os.Stderr) // Ensure global logger is configured

	return logger
}

func readTaskFromFile(taskFilePath string, logger logutil.LoggerInterface) (string, error) {
	// Create prompt builder
	promptBuilder := architect.NewPromptBuilder(logger)

	// Read task file using the prompt builder
	content, err := promptBuilder.ReadTaskFromFile(taskFilePath)
	if err != nil {
		// We need to map generic errors to our sentinel errors for backward compatibility
		if strings.Contains(err.Error(), "task file not found") {
			return "", fmt.Errorf("%w: %s", ErrTaskFileNotFound, taskFilePath)
		}
		if strings.Contains(err.Error(), "task file permission denied") {
			return "", fmt.Errorf("%w: %s", ErrTaskFileReadPermission, taskFilePath)
		}
		if strings.Contains(err.Error(), "task file path is a directory") {
			return "", fmt.Errorf("%w: %s", ErrTaskFileIsDir, taskFilePath)
		}
		if strings.Contains(err.Error(), "task file is empty") {
			return "", fmt.Errorf("%w: %s", ErrTaskFileEmpty, taskFilePath)
		}
		// Generic error
		return "", err
	}

	return content, nil
}

func validateInputs(config *Configuration, logger logutil.LoggerInterface) {
	result := doValidateInputs(config, logger)

	// If validation failed
	if !result.Valid {
		// Log the appropriate error message
		logger.Error(result.ErrorMessage)

		// Show usage
		flag.Usage()

		// Exit with error code
		os.Exit(1)
	}
}

func doValidateInputs(config *Configuration, logger logutil.LoggerInterface) validateInputsResult {
	// Initialize result with default values
	result := validateInputsResult{
		Valid:        true,
		ErrorMessage: "",
		WarningsOnly: false,
	}

	// Track whether task has been successfully loaded
	taskLoaded := false

	if config.TaskFile != "" {
		// Task file provided - this is the preferred path
		taskContent, err := readTaskFromFile(config.TaskFile, logger)
		if err != nil {
			// Set validation failed
			result.Valid = false

			// Specific error handling
			switch {
			case errors.Is(err, ErrTaskFileNotFound):
				result.ErrorMessage = fmt.Sprintf("Task file not found. Please check the path: %s", config.TaskFile)
				logger.Error(result.ErrorMessage)
			case errors.Is(err, ErrTaskFileReadPermission):
				result.ErrorMessage = fmt.Sprintf("Cannot read task file due to permissions. Please check permissions for: %s", config.TaskFile)
				logger.Error(result.ErrorMessage)
			case errors.Is(err, ErrTaskFileIsDir):
				result.ErrorMessage = fmt.Sprintf("The specified task file path is a directory, not a file: %s", config.TaskFile)
				logger.Error(result.ErrorMessage)
			case errors.Is(err, ErrTaskFileEmpty):
				result.ErrorMessage = fmt.Sprintf("The task file is empty or contains only whitespace: %s", config.TaskFile)
				logger.Error(result.ErrorMessage)
			default:
				// Generic fallback with more specific underlying error
				result.ErrorMessage = fmt.Sprintf("Failed to load task file '%s': %v", config.TaskFile, err)
				logger.Error(result.ErrorMessage)
			}
			return result
		}

		// Set task description from file content
		config.TaskDescription = taskContent
		taskLoaded = true
		logger.Debug("Loaded task description from file: %s", config.TaskFile)

		// Check if --task was also unnecessarily provided
		if getTaskFlagValue() != "" {
			logger.Warn("Both --task and --task-file flags were provided. Using task from --task-file. The --task flag is deprecated.")
		}
	} else if getTaskFlagValue() != "" && !config.DryRun {
		// Task file NOT provided, but deprecated --task IS provided (and not dry run)
		logger.Warn("The --task flag is deprecated and will be removed in a future version. Please use --task-file instead.")
		// config.TaskDescription is already set from parseFlags
		taskLoaded = true
	}

	// Check if a task is loaded (unless in dry-run mode)
	if !taskLoaded && !config.DryRun {
		result.Valid = false
		result.ErrorMessage = "The required --task-file flag is missing."
		logger.Error(result.ErrorMessage)
		return result
	}

	// Check for input paths
	if len(config.Paths) == 0 {
		result.Valid = false
		result.ErrorMessage = "At least one file or directory path must be provided as an argument."
		logger.Error(result.ErrorMessage)
		return result
	}

	// Check for API key
	if config.ApiKey == "" {
		result.Valid = false
		result.ErrorMessage = fmt.Sprintf("%s environment variable not set.", apiKeyEnvVar)
		logger.Error(result.ErrorMessage)
		return result
	}

	return result
}

// getTaskFlagValue returns the value of the task flag
// This function is extracted to allow mocking in tests
var getTaskFlagValue = func() string {
	if f := flag.Lookup("task"); f != nil {
		return f.Value.String()
	}
	return ""
}

func convertConfigToMap(cliConfig *Configuration) map[string]interface{} {
	// Create a map of CLI flags suitable for merging
	return map[string]interface{}{
		"task_description":    cliConfig.TaskDescription,
		"task_file":           cliConfig.TaskFile,
		"output_file":         cliConfig.OutputFile,
		"model":               cliConfig.ModelName,
		"verbose":             cliConfig.Verbose,
		"log_level":           cliConfig.LogLevel,
		"use_colors":          cliConfig.UseColors,
		"include":             cliConfig.Include,
		"format":              cliConfig.Format,
		"dry_run":             cliConfig.DryRun,
		"confirm_tokens":      cliConfig.ConfirmTokens,
		"paths":               cliConfig.Paths,
		"api_key":             cliConfig.ApiKey,
		"templates.default":   cliConfig.PromptTemplate, // Map prompt template to config format
		"excludes.extensions": cliConfig.Exclude,
		"excludes.names":      cliConfig.ExcludeNames,
	}
}

func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func backfillConfigFromAppConfig(cliConfig *Configuration, appConfig *config.AppConfig) *Configuration {
	// Start with CLI config to preserve fields not in AppConfig
	config := *cliConfig

	// Override with values from AppConfig only if they weren't explicitly set via CLI
	if flag.Lookup("output").Value.String() == defaultOutputFile {
		config.OutputFile = appConfig.OutputFile
	}
	if flag.Lookup("model").Value.String() == defaultModel {
		config.ModelName = appConfig.ModelName
	}
	if !isFlagSet("verbose") {
		config.Verbose = appConfig.Verbose
	}
	if !isFlagSet("color") {
		config.UseColors = appConfig.UseColors
	}
	if !isFlagSet("include") {
		config.Include = appConfig.Include
	}
	if !isFlagSet("exclude") {
		config.Exclude = appConfig.Excludes.Extensions
	}
	if !isFlagSet("exclude-names") {
		config.ExcludeNames = appConfig.Excludes.Names
	}
	if !isFlagSet("format") {
		config.Format = appConfig.Format
	}
	if !isFlagSet("confirm-tokens") {
		config.ConfirmTokens = appConfig.ConfirmTokens
	}
	// ClarifyTask handling has been removed
	if !isFlagSet("prompt-template") && appConfig.Templates.Default != "" {
		config.PromptTemplate = appConfig.Templates.Default
	}

	return &config
}

func listExampleTemplates(logger logutil.LoggerInterface, configManager config.ManagerInterface) {
	// Create prompt builder
	promptBuilder := architect.NewPromptBuilder(logger)

	// Call the method from the prompt builder
	err := promptBuilder.ListExampleTemplates(configManager)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing example templates: %v\n", err)
		os.Exit(1)
	}
}

func showExampleTemplate(name string, logger logutil.LoggerInterface, configManager config.ManagerInterface) {
	// Create prompt builder
	promptBuilder := architect.NewPromptBuilder(logger)

	// Call the method from the prompt builder
	err := promptBuilder.ShowExampleTemplate(name, configManager)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func getTokenInfo(ctx context.Context, client gemini.Client, prompt string, logger logutil.LoggerInterface) (tokenInfoResult, error) {
	// Create result struct
	result := tokenInfoResult{}
	
	// Get token count
	tokenCount, err := client.CountTokens(ctx, prompt)
	if err != nil {
		result.CountingError = err
		result.countingError = err
		return result, err
	}
	
	// Store token count
	if tokenCount != nil {
		result.TokenCount = tokenCount.Total
		result.tokenCount = tokenCount.Total
	}
	
	// Get model info
	modelInfo, err := client.GetModelInfo(ctx)
	if err != nil {
		result.InfoError = err
		result.infoError = err
		return result, err
	}
	
	// Store model info
	result.ModelInfo = modelInfo
	if modelInfo != nil {
		result.modelName = modelInfo.Name
		result.inputLimit = modelInfo.InputTokenLimit
		result.outputLimit = modelInfo.OutputTokenLimit
		
		// Calculate percentage of limit
		result.PercentOfMax = float64(result.TokenCount) / float64(modelInfo.InputTokenLimit) * 100
		result.percentOfMax = result.PercentOfMax
		
		// Check if token count exceeds limit
		if result.TokenCount > modelInfo.InputTokenLimit {
			result.ExceedsLimit = true
			result.exceedsLimit = true
			result.ExceededBy = result.TokenCount - modelInfo.InputTokenLimit
			result.exceededBy = result.ExceededBy
			result.limitError = fmt.Errorf("token limit exceeded by %d tokens", result.ExceededBy)
		}
	}
	
	return result, nil
}

func checkTokenLimit(ctx context.Context, client gemini.Client, prompt string, logger logutil.LoggerInterface) error {
	tokenManager := architect.NewTokenManager(logger)
	return tokenManager.CheckTokenLimit(ctx, client, prompt)
}

func processApiResponse(response *gemini.GenerationResult, logger logutil.LoggerInterface) string {
	// Process the generationResult and return the text
	if response == nil {
		logger.Error("Empty response from API")
		return ""
	}

	// Check for safety issues
	if response.FinishReason == "SAFETY" {
		logger.Warn("Content blocked due to safety concerns")
		return ""
	}

	// Return the text content
	return response.Content
}

func promptForConfirmation(tokenCount int32, confirmTokens int, logger logutil.LoggerInterface) bool {
	if confirmTokens <= 0 {
		return true // No confirmation needed
	}
	
	// Format the message
	msg := fmt.Sprintf("Token count (%d) exceeds confirmation threshold (%d). Proceed?", tokenCount, confirmTokens)
	
	// Read from stdin
	reader := bufio.NewReader(os.Stdin)
	logger.Warn("%s (y/N): ", msg)
	
	response, err := reader.ReadString('\n')
	if err != nil {
		logger.Error("Error reading input: %v", err)
		return false
	}
	
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}