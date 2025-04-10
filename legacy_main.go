//go:build ignore
// +build ignore

// legacy_main.go - Contains the original transitional main logic for reference/testing
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
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/prompt"
)

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

// OriginalMain is the original entry point for the architect CLI
// It's a refactored version of the original main function that gradually uses
// the new components while maintaining backward compatibility
func OriginalMain() int {
	// Create a base context
	ctx := context.Background()

	// Parse command line flags
	cliConfig := parseFlags()

	// Setup logging early for error reporting
	logger := setupLogging(cliConfig)
	logger.Info("Starting Architect - AI-assisted planning tool")

	// Initialize XDG-compliant configuration system
	configManager := initConfigSystem(logger)

	// Load configuration from files
	err := configManager.LoadFromFiles()
	if err != nil {
		logger.Warn("Failed to load configuration: %v", err)
		logger.Info("Using default configuration")
	}

	// Ensure configuration directories exist
	if err := configManager.EnsureConfigDirs(); err != nil {
		logger.Warn("Failed to create configuration directories: %v", err)
	}

	// Handle special subcommands before regular flow
	if handleSpecialCommands(cliConfig, logger, configManager) {
		// Special command was executed, exit the program
		return 0
	}

	// Convert CLI flags to the format needed for merging
	cliFlags := convertConfigToMap(cliConfig)

	// Merge CLI flags with loaded configuration
	if err := configManager.MergeWithFlags(cliFlags); err != nil {
		logger.Warn("Failed to merge CLI flags with configuration: %v", err)
	}

	// Get the final configuration
	appConfig := configManager.GetConfig()

	// Create backfilled CLI config for backward compatibility
	config := backfillConfigFromAppConfig(cliConfig, appConfig)

	// Process task input from file or flag
	taskDescription := processTaskInput(config, logger)

	// Validate inputs - exit with code 1 if validation fails
	result := doValidateInputs(config, logger)
	if !result.Valid {
		logger.Error(result.ErrorMessage)
		flag.Usage()
		return 1
	}

	// Initialize API client using the new APIService
	apiService := architect.NewAPIService(logger)
	geminiClient, err := apiService.InitClient(ctx, config.ApiKey, config.ModelName)
	if err != nil {
		// Get detailed error information
		errorDetails := apiService.GetErrorDetails(err)

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Error creating Gemini client: %s", apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			// Log more details in debug mode
			if config.LogLevel == logutil.DebugLevel {
				logger.Debug("Error details: %s", apiErr.DebugInfo())
			}
		} else {
			logger.Error("Error creating Gemini client: %s", errorDetails)
		}

		logger.Fatal("Failed to initialize API client")
		return 1
	}
	defer geminiClient.Close()

	// Create token manager and context gatherer using the new components
	tokenManager := architect.NewTokenManager(logger)
	contextGatherer := architect.NewContextGatherer(logger, config.DryRun, tokenManager)

	// Create gather config
	gatherConfig := architect.GatherConfig{
		Paths:        config.Paths,
		Include:      config.Include,
		Exclude:      config.Exclude,
		ExcludeNames: config.ExcludeNames,
		Format:       config.Format,
		Verbose:      config.Verbose,
		LogLevel:     config.LogLevel,
	}

	// Gather context from files
	projectContext, contextStats, err := contextGatherer.GatherContext(ctx, geminiClient, gatherConfig)
	if err != nil {
		logger.Fatal("Failed during project context gathering: %v", err)
		return 1
	}

	// Handle dry run mode
	if config.DryRun {
		err = contextGatherer.DisplayDryRunInfo(ctx, geminiClient, contextStats)
		if err != nil {
			logger.Error("Error displaying dry run information: %v", err)
			return 1
		}
		return 0
	}

	// Create output writer
	outputWriter := architect.NewOutputWriter(logger, tokenManager)

	// Generate content if not in dry run mode
	err = outputWriter.GenerateAndSavePlanWithConfig(
		ctx,
		geminiClient,
		taskDescription,
		projectContext,
		config.OutputFile,
		configManager,
	)
	if err != nil {
		logger.Fatal("Error generating and saving plan: %v", err)
		return 1
	}

	logger.Info("Plan successfully generated and saved to %s", config.OutputFile)
	return 0
}

// clarifyTaskDescription function removed

// parseFlags handles command line argument parsing
// Transitional implementation - moved to cmd/architect/cli.go
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

// setupLogging initializes the logger based on configuration
// Transitional implementation - moved to cmd/architect/cli.go
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

// readTaskFromFile reads task description from a file
// Transitional implementation - moved to cmd/architect/prompt.go
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

// validateInputsResult represents the result of input validation
// Transitional implementation - moved to cmd/architect/cli.go
type validateInputsResult struct {
	// Whether validation passed successfully
	Valid bool
	// Error message if validation failed
	ErrorMessage string
	// Whether validation concerns should be treated as warnings only
	WarningsOnly bool
}

// validateInputs verifies required inputs are provided
// Transitional implementation - moved to cmd/architect/cli.go
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

// doValidateInputs performs the actual validation logic and returns a result
// This function is extracted to allow testing without os.Exit
// Transitional implementation - moved to cmd/architect/cli.go
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

// initGeminiClient function moved to cmd/architect/api.go

// gatherContext collects and processes files based on configuration
// Transitional implementation - moved to cmd/architect/context.go
func gatherContext(ctx context.Context, config *Configuration, geminiClient gemini.Client, logger logutil.LoggerInterface) string {
	// Spinner initialization removed

	// Log appropriate message based on mode and start spinner
	if config.DryRun {
		logger.Info("Gathering files that would be included in context...")
		logger.Debug("Gathering files that would be included in context...")
		logger.Info("Dry run mode: gathering files that would be included in context...")
	} else {
		logger.Info("Gathering project context...")
		logger.Debug("Gathering project context...")
	}

	// Setup file processing configuration
	fileConfig := fileutil.NewConfig(config.Verbose, config.Include, config.Exclude, config.ExcludeNames, config.Format, logger)

	// Track processed files for dry run mode
	var processedFiles []string
	if config.DryRun {
		processedFiles = make([]string, 0)
		collector := func(path string) {
			processedFiles = append(processedFiles, path)
		}
		fileConfig.SetFileCollector(collector)
	}

	// Gather project context
	projectContext, processedFilesCount, err := fileutil.GatherProjectContext(config.Paths, fileConfig)
	if err != nil {
		logger.Error("Failed during project context gathering: %v", err)
		logger.Fatal("Failed during project context gathering: %v", err)
	}

	// Log warning if no files were processed
	if processedFilesCount == 0 {
		logger.Info("No files were processed for context. Check paths and filters.")
		logger.Warn("No files were processed for context. Check paths and filters.")
		return projectContext
	}

	// Update spinner message and calculate statistics
	logger.Info("Calculating token statistics...")
	logger.Debug("Calculating token statistics...")
	charCount, lineCount, tokenCount := fileutil.CalculateStatisticsWithTokenCounting(ctx, geminiClient, projectContext, logger)

	// Handle dry run mode specific output
	if config.DryRun {
		logger.Info("Context gathered: %d files, %d lines, %d chars, %d tokens",
			processedFilesCount, lineCount, charCount, tokenCount)
		displayDryRunInfo(charCount, lineCount, tokenCount, processedFilesCount, processedFiles, ctx, geminiClient, logger)
	} else if config.LogLevel == logutil.DebugLevel || processedFilesCount > 0 {
		// Normal run mode
		logger.Info("Context gathered: %d files, %d lines, %d chars, %d tokens",
			processedFilesCount, lineCount, charCount, tokenCount)
		logger.Info("Context gathered: %d files processed, %d lines, %d chars, %d tokens.",
			processedFilesCount, lineCount, charCount, tokenCount)
	}

	return projectContext
}

// displayDryRunInfo shows detailed information for dry run mode
// Transitional implementation - moved to cmd/architect/context.go
func displayDryRunInfo(charCount int, lineCount int, tokenCount int, processedFilesCount int,
	processedFiles []string, ctx context.Context, geminiClient gemini.Client, logger logutil.LoggerInterface) {

	logger.Info("Files that would be included in context:")
	if processedFilesCount == 0 {
		logger.Info("  No files matched the current filters.")
	} else {
		for i, file := range processedFiles {
			logger.Info("  %d. %s", i+1, file)
		}
	}

	logger.Info("Context statistics:")
	logger.Info("  Files: %d", processedFilesCount)
	logger.Info("  Lines: %d", lineCount)
	logger.Info("  Characters: %d", charCount)
	logger.Info("  Tokens: %d", tokenCount)

	// Get model info for token limit comparison
	modelInfo, modelInfoErr := geminiClient.GetModelInfo(ctx)
	if modelInfoErr != nil {
		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(modelInfoErr); ok {
			logger.Warn("Could not get model information: %s", apiErr.Message)
			// Only show detailed info in debug logs
			logger.Debug("Model info error details: %s", apiErr.DebugInfo())
		} else {
			logger.Warn("Could not get model information: %v", modelInfoErr)
		}
	} else {
		// Convert to int32 for comparison with model limits
		tokenCountInt32 := int32(tokenCount)
		percentOfLimit := float64(tokenCountInt32) / float64(modelInfo.InputTokenLimit) * 100
		logger.Info("Token usage: %d / %d (%.1f%% of model's limit)",
			tokenCountInt32, modelInfo.InputTokenLimit, percentOfLimit)

		// Check if token count exceeds limit
		if tokenCountInt32 > modelInfo.InputTokenLimit {
			logger.Error("WARNING: Token count exceeds model's limit by %d tokens",
				tokenCountInt32-modelInfo.InputTokenLimit)
			logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		} else {
			logger.Info("Context size is within the model's token limit")
		}
	}

	logger.Info("Dry run completed successfully.")
	logger.Info("To generate content, run without the --dry-run flag.")
}

// generateAndSavePlan is a backward-compatible wrapper for plan generation
// Transitional implementation - moved to cmd/architect/output.go
func generateAndSavePlan(ctx context.Context, config *Configuration, geminiClient gemini.Client,
	projectContext string, logger logutil.LoggerInterface) {

	// Create token manager and output writer
	tokenManager := architect.NewTokenManager(logger)
	outputWriter := architect.NewOutputWriter(logger, tokenManager)

	// Create a fallback prompt manager without config
	promptManager := prompt.NewManager(logger)

	// Call the method from the output writer
	err := outputWriter.GenerateAndSavePlan(ctx, geminiClient, config.TaskDescription, projectContext, config.OutputFile, promptManager)
	if err != nil {
		logger.Fatal("Error generating and saving plan: %v", err)
	}
}

// generateAndSavePlanWithConfig creates and saves the plan to a file using the config system
// Transitional implementation - moved to cmd/architect/output.go
func generateAndSavePlanWithConfig(ctx context.Context, config *Configuration, geminiClient gemini.Client,
	projectContext string, configManager config.ManagerInterface, logger logutil.LoggerInterface) {

	// Create token manager and output writer
	tokenManager := architect.NewTokenManager(logger)
	outputWriter := architect.NewOutputWriter(logger, tokenManager)

	// Call the method from the output writer
	err := outputWriter.GenerateAndSavePlanWithConfig(ctx, geminiClient, config.TaskDescription, projectContext, config.OutputFile, configManager)
	if err != nil {
		logger.Fatal("Error generating and saving plan with config: %v", err)
	}
}

// generateAndSavePlanWithPromptManager is the core implementation of plan generation
// Transitional implementation - moved to cmd/architect/output.go
func generateAndSavePlanWithPromptManager(ctx context.Context, config *Configuration, geminiClient gemini.Client,
	projectContext string, promptManager prompt.ManagerInterface, logger logutil.LoggerInterface) {

	// Create token manager and output writer
	tokenManager := architect.NewTokenManager(logger)
	outputWriter := architect.NewOutputWriter(logger, tokenManager)

	// Call the method from the output writer
	err := outputWriter.GenerateAndSavePlan(ctx, geminiClient, config.TaskDescription, projectContext, config.OutputFile, promptManager)
	if err != nil {
		logger.Fatal("Error generating and saving plan with prompt manager: %v", err)
	}
}

// processApiResponse function moved to cmd/architect/api.go
// Re-export for testing with original signature
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

// saveToFile writes the generated plan to the specified file
// Transitional implementation - moved to cmd/architect/output.go
func saveToFile(content string, outputFile string, logger logutil.LoggerInterface) {
	// Create an output writer
	tokenManager := architect.NewTokenManager(logger)
	outputWriter := architect.NewOutputWriter(logger, tokenManager)

	// Call the method from the output writer
	err := outputWriter.SaveToFile(content, outputFile)
	if err != nil {
		logger.Fatal("Error saving to file: %v", err)
	}
}

// initSpinner function removed

// promptForConfirmation function moved to TokenManager
// Re-export for testing - maintaining original signature
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
	limitError    error // Added to satisfy tests
	// Keep the capitalized fields for our implementation
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

// Re-export for testing
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

// Re-export for testing
func checkTokenLimit(ctx context.Context, client gemini.Client, prompt string, logger logutil.LoggerInterface) error {
	tokenManager := architect.NewTokenManager(logger)
	return tokenManager.CheckTokenLimit(ctx, client, prompt)
}

// initConfigSystem initializes the configuration system
// Transitional implementation - moved to cmd/architect/cli.go
func initConfigSystem(logger logutil.LoggerInterface) config.ManagerInterface {
	return config.NewManager(logger)
}

// convertConfigToMap converts the CLI Configuration struct to a map for merging with loaded config
// Transitional implementation - moved to cmd/architect/cli.go
// listExampleTemplates displays a list of available example templates
// Transitional implementation - moved to cmd/architect/prompt.go
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

// showExampleTemplate displays the content of a specific example template
// Transitional implementation - moved to cmd/architect/prompt.go
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

// Transitional implementation - moved to cmd/architect/cli.go
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

// backfillConfigFromAppConfig creates a Configuration object from AppConfig for backward compatibility
// Transitional implementation - moved to cmd/architect/cli.go
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

// isFlagSet checks if a flag was explicitly set on the command line
// Transitional implementation - moved to cmd/architect/cli.go
func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// getTaskFlagValue returns the value of the task flag
// This function is extracted to allow mocking in tests
// Transitional implementation - moved to cmd/architect/cli.go
var getTaskFlagValue = func() string {
	if f := flag.Lookup("task"); f != nil {
		return f.Value.String()
	}
	return ""
}

// buildPrompt constructs the prompt string for the Gemini API.
// Transitional implementation - moved to cmd/architect/prompt.go
// nolint:unused
func buildPrompt(config *Configuration, task string, context string, logger logutil.LoggerInterface) (string, error) {
	// Create prompt builder
	promptBuilder := architect.NewPromptBuilder(logger)

	// Call the method from the prompt builder
	return promptBuilder.BuildPrompt(task, context, config.PromptTemplate)
}

// buildPromptWithConfig constructs the prompt string using the configuration system
// Transitional implementation - moved to cmd/architect/prompt.go
// nolint:unused
func buildPromptWithConfig(config *Configuration, task string, context string, configManager config.ManagerInterface, logger logutil.LoggerInterface) (string, error) {
	// Create prompt builder
	promptBuilder := architect.NewPromptBuilder(logger)

	// Call the method from the prompt builder
	return promptBuilder.BuildPromptWithConfig(task, context, config.PromptTemplate, configManager)
}

// buildPromptWithManager constructs the prompt string using the provided prompt manager.
// Transitional implementation - moved to cmd/architect/prompt.go
// This function is exported for testing purposes.
func buildPromptWithManager(config *Configuration, task string, context string, promptManager prompt.ManagerInterface, logger logutil.LoggerInterface) (string, error) {
	// No need to create a prompt builder in this transitional function
	// Just use the prompt manager directly

	// Adapt to the new interface by using BuildPrompt and passing the prompt manager's result
	// This is a bit of a hack for backward compatibility, but it works for transitional period
	customTemplateName := config.PromptTemplate

	// Create template data
	data := &prompt.TemplateData{
		Task:    task,
		Context: context,
	}

	// Determine which template to use
	templateName := "default.tmpl"
	if customTemplateName != "" {
		templateName = customTemplateName
		logger.Debug("Using custom prompt template: %s", templateName)
	}

	// Build the prompt (template loading is now handled by the manager)
	return promptManager.BuildPrompt(templateName, data)
}

// handleSpecialCommands processes special command flags like list-examples and show-example
// Returns true if a special command was executed
// This is a bridge to the architect.HandleSpecialCommands function that maintains backward compatibility
func handleSpecialCommands(cliConfig *Configuration, logger logutil.LoggerInterface, configManager config.ManagerInterface) bool {
	// Create prompt builder
	promptBuilder := architect.NewPromptBuilder(logger)
	
	// Handle special subcommands before regular flow
	if cliConfig.ListExamples {
		err := promptBuilder.ListExampleTemplates(configManager)
		if err != nil {
			logger.Error("Error listing example templates: %v", err)
			os.Exit(1)
		}
		return true
	}

	if cliConfig.ShowExample != "" {
		err := promptBuilder.ShowExampleTemplate(cliConfig.ShowExample, configManager)
		if err != nil {
			logger.Error("Error showing example template: %v", err)
			os.Exit(1)
		}
		return true
	}

	// No special commands were executed
	return false
}

// processTaskInput extracts task description from file or flag
// This is a bridge to the architect.ProcessTaskInput function that maintains backward compatibility
func processTaskInput(cliConfig *Configuration, logger logutil.LoggerInterface) string {
	// If task file is provided, read from file
	if cliConfig.TaskFile != "" {
		// Create prompt builder
		promptBuilder := architect.NewPromptBuilder(logger)
		
		// Read task from file using prompt builder
		content, err := promptBuilder.ReadTaskFromFile(cliConfig.TaskFile)
		if err != nil {
			logger.Error("Failed to read task from file: %v", err)
			os.Exit(1)
		}
		return content
	}
	
	// Otherwise, use the task description from CLI flags
	return cliConfig.TaskDescription
}

// Renamed from main for compatibility with legacy tests
func legacyEntry() {
	exitCode := OriginalMain()
	os.Exit(exitCode)
}