// main.go
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/prompt"
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
	ClarifyTask     bool
	Paths           []string
	ApiKey          string
}

func main() {
	// Parse command line flags
	cliConfig := parseFlags()

	// Setup logging early for error reporting
	logger := setupLogging(cliConfig)

	// Initialize XDG-compliant configuration system
	configManager := initConfigSystem(logger)

	// Create a temporary NoopLogger for early initialization
	// This is needed because we need to load configuration before we can
	// initialize the real audit logger, but the config system needs an audit logger.
	// We'll use a NoopLogger during this bootstrapping phase.
	tempAuditLogger := auditlog.NewNoopLogger()

	// Update the configManager with the temporary audit logger (if supported)
	if manager, ok := configManager.(*config.Manager); ok {
		manager.AuditLogger = tempAuditLogger
	}

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

	// Convert CLI flags to the format needed for merging
	cliFlags := convertConfigToMap(cliConfig)

	// Merge CLI flags with loaded configuration
	if err := configManager.MergeWithFlags(cliFlags); err != nil {
		logger.Warn("Failed to merge CLI flags with configuration: %v", err)
	}

	// Get the final configuration
	appConfig := configManager.GetConfig()

	// Initialize structured audit logger with the loaded configuration
	auditLogger := initAuditLogger(appConfig, logger)
	defer auditLogger.Close() // Ensure logger is closed at program exit
	
	// Update the config manager with the real audit logger (if supported)
	if manager, ok := configManager.(*config.Manager); ok {
		manager.AuditLogger = auditLogger
	}

	// Log application startup
	startupEvent := auditlog.NewAuditEvent(
		"INFO",
		"ApplicationStart",
		"Architect tool started",
	).WithMetadata("version", "1.0.0") // TODO: Use actual version

	auditLogger.Log(startupEvent)

	// Create backfilled CLI config for backward compatibility
	config := backfillConfigFromAppConfig(cliConfig, appConfig)

	// Validate inputs
	validateInputs(config, logger)

	// Initialize API client
	ctx := context.Background()
	geminiClient := initGeminiClient(ctx, config, logger)
	defer geminiClient.Close()

	// If task clarification is enabled, let the user refine their task
	if config.ClarifyTask && !config.DryRun {
		config.TaskDescription = clarifyTaskDescriptionWithConfig(ctx, config, geminiClient, configManager, logger, auditLogger)
	}

	// Gather context from files
	projectContext := gatherContext(ctx, config, geminiClient, logger, auditLogger)

	// Generate content if not in dry run mode
	if !config.DryRun {
		generateAndSavePlanWithConfig(ctx, config, geminiClient, projectContext, configManager, logger, auditLogger)
	}

	// Log application shutdown
	shutdownEvent := auditlog.NewAuditEvent(
		"INFO",
		"ApplicationEnd",
		"Architect tool completed successfully",
	)
	auditLogger.Log(shutdownEvent)
}

// clarifyTaskDescription is a backward-compatible wrapper for clarification process
func clarifyTaskDescription(ctx context.Context, config *Configuration, geminiClient gemini.Client, logger logutil.LoggerInterface, auditLogger auditlog.StructuredLogger) string {
	// Use legacy version with default prompt manager
	promptManager := prompt.NewManager(logger)
	return clarifyTaskDescriptionWithPromptManager(ctx, config, geminiClient, promptManager, logger, auditLogger)
}

// clarifyTaskDescriptionWithConfig performs task clarification using the config system
func clarifyTaskDescriptionWithConfig(ctx context.Context, config *Configuration, geminiClient gemini.Client, configManager config.ManagerInterface, logger logutil.LoggerInterface, auditLogger auditlog.StructuredLogger) string {
	// Set up prompt manager with config support
	promptManager, err := prompt.SetupPromptManagerWithConfig(logger, configManager)
	if err != nil {
		logger.Error("Failed to set up prompt manager: %v", err)
		// Log the error
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"PromptManagerSetupError",
			"Failed to set up prompt manager with config",
		).WithErrorFromGoError(err))
		// Fall back to the non-config version
		return clarifyTaskDescription(ctx, config, geminiClient, logger, auditLogger)
	}

	return clarifyTaskDescriptionWithPromptManager(ctx, config, geminiClient, promptManager, logger, auditLogger)
}

// clarifyTaskDescriptionWithPromptManager is the core implementation of the task clarification process
func clarifyTaskDescriptionWithPromptManager(ctx context.Context, config *Configuration, geminiClient gemini.Client, promptManager prompt.ManagerInterface, logger logutil.LoggerInterface, auditLogger auditlog.StructuredLogger) string {
	// Spinner initialization removed

	// Original task description
	originalTask := config.TaskDescription

	// Log that we're starting the clarification process
	clarifyEvent := auditlog.NewAuditEvent(
		"INFO",
		"TaskClarificationStart",
		"Starting task clarification process",
	).WithInput(originalTask)
	auditLogger.Log(clarifyEvent)

	// Build prompt for clarification (template loading is handled internally)
	logger.Info("Analyzing task description...")
	logger.Debug("Analyzing task description...")
	data := &prompt.TemplateData{
		Task: originalTask,
	}

	clarifyPrompt, err := promptManager.BuildPrompt("clarify.tmpl", data)
	if err != nil {
		logger.Error("Failed to build clarification prompt: %v", err)
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"TaskClarificationError",
			"Failed to build clarification prompt",
		).WithErrorFromGoError(err))
		return originalTask
	}

	// Call Gemini to generate clarification questions
	logger.Info("Generating clarification questions...")
	logger.Debug("Generating clarification questions...")
	
	// Log API call
	apiCallEvent := auditlog.NewAuditEvent(
		"INFO",
		"APIRequest",
		"Calling Gemini API for clarification questions",
	).WithMetadata("model", config.ModelName)
	auditLogger.Log(apiCallEvent)
	
	result, err := geminiClient.GenerateContent(ctx, clarifyPrompt)
	if err != nil {
		logger.Error("Error generating clarification questions: %v", err)
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"APIError",
			"Error generating clarification questions",
		).WithErrorFromGoError(err))
		return originalTask
	}

	// Process the JSON response
	var clarificationData struct {
		Analysis  string   `json:"analysis"`
		Questions []string `json:"questions"`
	}

	// Parse JSON response - must be valid JSON as requested in the prompt
	err = json.Unmarshal([]byte(result.Content), &clarificationData)
	if err != nil {
		logger.Error("Failed to parse clarification response: %v", err)
		logger.Debug("Response content: %s", result.Content)
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"ResponseParsingError",
			"Failed to parse clarification response as JSON",
		).WithErrorFromGoError(err).
		WithMetadata("response", result.Content))
		return originalTask
	}

	// Stop spinner and start the interactive clarification process
	logger.Info("Task analysis complete")

	// Show the analysis to the user
	logger.Info("Task Analysis: %s", clarificationData.Analysis)
	
	// Log that we received analysis
	analysisEvent := auditlog.NewAuditEvent(
		"INFO",
		"TaskAnalysisComplete",
		"Received task analysis from API",
	).WithOutput(clarificationData.Analysis).
	WithMetadata("num_questions", len(clarificationData.Questions))
	auditLogger.Log(analysisEvent)

	// Present each question and collect answers
	var questionAnswers strings.Builder
	fmt.Println("\n🔍 Task Clarification:")

	reader := bufio.NewReader(os.Stdin)
	for i, question := range clarificationData.Questions {
		fmt.Printf("\n%d. %s\n", i+1, question)
		fmt.Print("   Your answer: ")

		answer, err := reader.ReadString('\n')
		if err != nil {
			logger.Error("Error reading input: %v", err)
			auditLogger.Log(auditlog.NewAuditEvent(
				"ERROR",
				"UserInputError",
				"Error reading user input",
			).WithErrorFromGoError(err))
			return originalTask
		}

		// Add the Q&A to our collection
		questionAnswers.WriteString(fmt.Sprintf("Question %d: %s\n", i+1, question))
		questionAnswers.WriteString(fmt.Sprintf("Answer %d: %s\n", i+1, strings.TrimSpace(answer)))
		
		// Log each Q&A
		qaEvent := auditlog.NewAuditEvent(
			"INFO",
			"UserClarification",
			fmt.Sprintf("User answered clarification question %d", i+1),
		).WithInput(question)
		.WithOutput(strings.TrimSpace(answer))
		auditLogger.Log(qaEvent)
	}

	// Now refine the task with the answers
	logger.Info("Refining task description...")
	logger.Debug("Refining task description...")

	// Build prompt for refinement (template loading is handled internally)
	refineData := &prompt.TemplateData{
		Task:    originalTask,
		Context: questionAnswers.String(),
	}

	refinePrompt, err := promptManager.BuildPrompt("refine.tmpl", refineData)
	if err != nil {
		logger.Error("Failed to build refinement prompt: %v", err)
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"TaskRefinementError",
			"Failed to build refinement prompt",
		).WithErrorFromGoError(err))
		return originalTask
	}

	// Log API call
	refineApiEvent := auditlog.NewAuditEvent(
		"INFO",
		"APIRequest",
		"Calling Gemini API for task refinement",
	).WithMetadata("model", config.ModelName)
	auditLogger.Log(refineApiEvent)

	// Call Gemini to generate refined task
	result, err = geminiClient.GenerateContent(ctx, refinePrompt)
	if err != nil {
		logger.Error("Error generating refined task: %v", err)
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"APIError",
			"Error generating refined task",
		).WithErrorFromGoError(err))
		return originalTask
	}

	// Process the JSON response
	var refinementData struct {
		RefinedTask string   `json:"refined_task"`
		KeyPoints   []string `json:"key_points"`
	}

	// Parse JSON response
	err = json.Unmarshal([]byte(result.Content), &refinementData)
	if err != nil {
		logger.Error("Failed to parse refinement response: %v", err)
		logger.Debug("Response content: %s", result.Content)
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"ResponseParsingError",
			"Failed to parse refinement response as JSON",
		).WithErrorFromGoError(err).
		WithMetadata("response", result.Content))
		return originalTask
	}

	// Stop spinner and show the refined task
	logger.Info("Task refinement complete")

	// Show the refinement results
	fmt.Println("\n✨ Refined Task Description:")
	fmt.Println(refinementData.RefinedTask)

	if len(refinementData.KeyPoints) > 0 {
		fmt.Println("\n🔑 Key Technical Points:")
		for i, point := range refinementData.KeyPoints {
			fmt.Printf("%d. %s\n", i+1, point)
		}
	}

	fmt.Println("\nProceeding with the refined task description...")

	// Log completion of clarification process
	completionEvent := auditlog.NewAuditEvent(
		"INFO",
		"TaskClarificationComplete",
		"Task clarification process completed successfully",
	).WithInput(originalTask)
	.WithOutput(refinementData.RefinedTask)
	.WithMetadata("key_points_count", len(refinementData.KeyPoints))
	auditLogger.Log(completionEvent)

	return refinementData.RefinedTask
}

// parseFlags handles command line argument parsing
func parseFlags() *Configuration {
	config := &Configuration{}

	// Define flags
	taskFlag := flag.String("task", "", "Description of the task or goal for the plan.")
	taskFileFlag := flag.String("task-file", "", "Path to a file containing the task description (alternative to --task).")
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
	clarifyTaskFlag := flag.Bool("clarify", false, "Enable interactive task clarification to refine your task description")

	// Set custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s (--task \"<description>\" | --task-file <path>) [options] <path1> [path2...]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <path1> [path2...]   One or more file or directory paths for project context.\n\n")
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
	config.ClarifyTask = *clarifyTaskFlag
	config.Paths = flag.Args()
	config.ApiKey = os.Getenv(apiKeyEnvVar)

	return config
}

// setupLogging initializes the logger based on configuration
func setupLogging(config *Configuration) logutil.LoggerInterface {
	// Determine log level
	logLevel := logutil.InfoLevel
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
func readTaskFromFile(taskFilePath string, logger logutil.LoggerInterface) (string, error) {
	// Resolve the task file path (treating it as an "output" type since it's an input file in the working directory)
	resolvedPath, err := resolvePath(taskFilePath, "output", logger)
	if err != nil {
		return "", fmt.Errorf("error resolving task file path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		return "", fmt.Errorf("task file not found: %s", resolvedPath)
	}

	// Read file content
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("error reading task file: %w", err)
	}

	// Return content as string
	return string(content), nil
}

// validateInputs verifies required inputs are provided
func validateInputs(config *Configuration, logger logutil.LoggerInterface) {
	// Check for task description (not required in dry run mode)
	if config.TaskFile != "" {
		// Load task from file
		taskContent, err := readTaskFromFile(config.TaskFile, logger)
		if err != nil {
			logger.Error("Failed to read task file: %v", err)
			flag.Usage()
			os.Exit(1)
		}

		// Set task description from file content
		config.TaskDescription = taskContent
		logger.Debug("Loaded task description from file: %s", config.TaskFile)
	}

	// Check if task description is still empty after potentially loading from file
	if config.TaskDescription == "" && !config.DryRun {
		logger.Error("Either --task or --task-file must be provided (except in dry-run mode).")
		flag.Usage()
		os.Exit(1)
	}

	// Check if both --task and --task-file are provided
	if config.TaskDescription != "" && config.TaskFile != "" && flag.Lookup("task").Value.String() != "" {
		logger.Warn("Both --task and --task-file flags were provided. Using task from --task-file.")
	}

	// Check for input paths
	if len(config.Paths) == 0 {
		logger.Error("At least one file or directory path must be provided as an argument.")
		flag.Usage()
		os.Exit(1)
	}

	// Check for API key
	if config.ApiKey == "" {
		logger.Error("%s environment variable not set.", apiKeyEnvVar)
		flag.Usage()
		os.Exit(1)
	}
}

// initGeminiClient creates and initializes the Gemini API client
func initGeminiClient(ctx context.Context, config *Configuration, logger logutil.LoggerInterface) gemini.Client {
	client, err := gemini.NewClient(ctx, config.ApiKey, config.ModelName)
	if err != nil {
		logger.Fatal("Error creating Gemini client: %v", err)
	}
	return client
}

// gatherContext collects and processes files based on configuration
func gatherContext(ctx context.Context, config *Configuration, geminiClient gemini.Client, logger logutil.LoggerInterface, auditLogger auditlog.StructuredLogger) string {
	// Spinner initialization removed

	// Log that we're starting the context gathering process
	gatherEvent := auditlog.NewAuditEvent(
		"INFO",
		"ContextGatheringStart",
		"Starting project context gathering",
	).WithMetadata("dry_run", config.DryRun).
	WithMetadata("include_filter", config.Include).
	WithMetadata("exclude_filter", config.Exclude).
	WithMetadata("exclude_names", config.ExcludeNames)
	
	// Add paths to metadata as an array
	for i, path := range config.Paths {
		gatherEvent = gatherEvent.WithMetadata(fmt.Sprintf("path_%d", i+1), path)
	}
	auditLogger.Log(gatherEvent)

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

	// Gather project context with audit logging
	projectContext, processedFilesCount, err := fileutil.GatherProjectContextWithAuditLogging(config.Paths, fileConfig, auditLogger)
	if err != nil {
		logger.Error("Failed during project context gathering: %v", err)
		
		// Log error to audit log
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"ContextGatheringError",
			"Failed to gather project context",
		).WithErrorFromGoError(err))
		
		logger.Fatal("Failed during project context gathering: %v", err)
	}

	// Log warning if no files were processed
	if processedFilesCount == 0 {
		logger.Info("No files were processed for context. Check paths and filters.")
		logger.Warn("No files were processed for context. Check paths and filters.")
		
		// Log warning to audit log
		auditLogger.Log(auditlog.NewAuditEvent(
			"WARN",
			"EmptyContext",
			"No files were processed for context",
		).WithMetadata("paths", strings.Join(config.Paths, ", ")))
		
		return projectContext
	}

	// Update spinner message and calculate statistics
	logger.Info("Calculating token statistics...")
	logger.Debug("Calculating token statistics...")
	
	// Log that we're calculating statistics
	statsEvent := auditlog.NewAuditEvent(
		"INFO",
		"CalculatingStatistics",
		"Calculating token statistics for context",
	)
	auditLogger.Log(statsEvent)
	
	charCount, lineCount, tokenCount := fileutil.CalculateStatisticsWithTokenCounting(ctx, geminiClient, projectContext, logger)

	// Log statistics to audit log
	contextStats := auditlog.NewAuditEvent(
		"INFO",
		"ContextStatistics",
		"Context gathering statistics",
	).WithMetadata("file_count", processedFilesCount).
	WithMetadata("line_count", lineCount).
	WithMetadata("char_count", charCount).
	WithMetadata("token_count", tokenCount)
	auditLogger.Log(contextStats)

	// Handle dry run mode specific output
	if config.DryRun {
		logger.Info("Context gathered: %d files, %d lines, %d chars, %d tokens",
			processedFilesCount, lineCount, charCount, tokenCount)
		displayDryRunInfo(charCount, lineCount, tokenCount, processedFilesCount, processedFiles, ctx, geminiClient, logger, auditLogger)
	} else if config.LogLevel == logutil.DebugLevel || processedFilesCount > 0 {
		// Normal run mode
		logger.Info("Context gathered: %d files, %d lines, %d chars, %d tokens",
			processedFilesCount, lineCount, charCount, tokenCount)
		logger.Info("Context gathered: %d files processed, %d lines, %d chars, %d tokens.",
			processedFilesCount, lineCount, charCount, tokenCount)
	}

	// Log completion of context gathering
	completionEvent := auditlog.NewAuditEvent(
		"INFO",
		"ContextGatheringComplete",
		"Project context gathering completed successfully",
	).WithMetadata("file_count", processedFilesCount)
	.WithMetadata("token_count", tokenCount)
	auditLogger.Log(completionEvent)

	return projectContext
}

// displayDryRunInfo shows detailed information for dry run mode
func displayDryRunInfo(charCount int, lineCount int, tokenCount int, processedFilesCount int,
	processedFiles []string, ctx context.Context, geminiClient gemini.Client, logger logutil.LoggerInterface, auditLogger auditlog.StructuredLogger) {

	// Log that we're displaying dry run info
	dryRunEvent := auditlog.NewAuditEvent(
		"INFO",
		"DryRunInfo",
		"Displaying dry run information",
	).WithMetadata("file_count", processedFilesCount).
	WithMetadata("line_count", lineCount).
	WithMetadata("char_count", charCount).
	WithMetadata("token_count", tokenCount)
	auditLogger.Log(dryRunEvent)

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
			
			auditLogger.Log(auditlog.NewAuditEvent(
				"WARN",
				"ModelInfoAPIError",
				"Could not get model information due to API error",
			).WithError(apiErr.Message))
		} else {
			logger.Warn("Could not get model information: %v", modelInfoErr)
			
			auditLogger.Log(auditlog.NewAuditEvent(
				"WARN",
				"ModelInfoError",
				"Could not get model information",
			).WithError(modelInfoErr.Error()))
		}
	} else {
		// Convert to int32 for comparison with model limits
		tokenCountInt32 := int32(tokenCount)
		percentOfLimit := float64(tokenCountInt32) / float64(modelInfo.InputTokenLimit) * 100
		logger.Info("Token usage: %d / %d (%.1f%% of model's limit)",
			tokenCountInt32, modelInfo.InputTokenLimit, percentOfLimit)

		// Log token information
		tokenInfoEvent := auditlog.NewAuditEvent(
			"INFO",
			"TokenUsageInfo",
			"Token usage information",
		).WithMetadata("token_count", tokenCountInt32)
		.WithMetadata("token_limit", modelInfo.InputTokenLimit)
		.WithMetadata("percentage", percentOfLimit)
		
		// Check if token count exceeds limit
		if tokenCountInt32 > modelInfo.InputTokenLimit {
			logger.Error("WARNING: Token count exceeds model's limit by %d tokens",
				tokenCountInt32-modelInfo.InputTokenLimit)
			logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
			
			tokenInfoEvent = tokenInfoEvent.WithMetadata("exceeds_limit", true)
			.WithMetadata("excess_tokens", tokenCountInt32-modelInfo.InputTokenLimit)
			auditLogger.Log(tokenInfoEvent)
			
			// Log a separate error event
			auditLogger.Log(auditlog.NewAuditEvent(
				"ERROR",
				"TokenLimitExceeded",
				"Token count exceeds model limit",
			).WithMetadata("token_count", tokenCountInt32)
			.WithMetadata("token_limit", modelInfo.InputTokenLimit)
			.WithMetadata("excess_tokens", tokenCountInt32-modelInfo.InputTokenLimit))
		} else {
			logger.Info("Context size is within the model's token limit")
			
			tokenInfoEvent = tokenInfoEvent.WithMetadata("exceeds_limit", false)
			auditLogger.Log(tokenInfoEvent)
		}
	}

	logger.Info("Dry run completed successfully.")
	logger.Info("To generate content, run without the --dry-run flag.")
	
	// Log dry run completion
	auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"DryRunComplete",
		"Dry run completed successfully",
	))
}

// generateAndSavePlan is a backward-compatible wrapper for plan generation
func generateAndSavePlan(ctx context.Context, config *Configuration, geminiClient gemini.Client,
	projectContext string, logger logutil.LoggerInterface, auditLogger auditlog.StructuredLogger) {

	// Use the legacy version without config system support
	promptManager := prompt.NewManager(logger)
	generateAndSavePlanWithPromptManager(ctx, config, geminiClient, projectContext, promptManager, logger, auditLogger)
}

// generateAndSavePlanWithConfig creates and saves the plan to a file using the config system
func generateAndSavePlanWithConfig(ctx context.Context, config *Configuration, geminiClient gemini.Client,
	projectContext string, configManager config.ManagerInterface, logger logutil.LoggerInterface, auditLogger auditlog.StructuredLogger) {

	// Set up a prompt manager with config support
	promptManager, err := prompt.SetupPromptManagerWithConfig(logger, configManager)
	if err != nil {
		logger.Error("Failed to set up prompt manager: %v", err)
		// Log the error
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"PromptManagerSetupError",
			"Failed to set up prompt manager with config",
		).WithErrorFromGoError(err))
		// Fall back to non-config version
		generateAndSavePlan(ctx, config, geminiClient, projectContext, logger, auditLogger)
		return
	}

	generateAndSavePlanWithPromptManager(ctx, config, geminiClient, projectContext, promptManager, logger, auditLogger)
}

// generateAndSavePlanWithPromptManager is the core implementation of plan generation
func generateAndSavePlanWithPromptManager(ctx context.Context, config *Configuration, geminiClient gemini.Client,
	projectContext string, promptManager prompt.ManagerInterface, logger logutil.LoggerInterface, auditLogger auditlog.StructuredLogger) {

	// Spinner initialization removed

	// Log that we're starting the plan generation process
	genStartEvent := auditlog.NewAuditEvent(
		"INFO",
		"PlanGenerationStart",
		"Starting plan generation process",
	).WithInput(config.TaskDescription)
	.WithMetadata("model", config.ModelName)
	.WithMetadata("output_file", config.OutputFile)
	auditLogger.Log(genStartEvent)

	// Construct prompt using the provided prompt manager
	logger.Info("Building prompt template...")
	logger.Debug("Building prompt template...")
	generatedPrompt, err := buildPromptWithManager(config, config.TaskDescription, projectContext, promptManager, logger)
	if err != nil {
		logger.Error("Failed to build prompt: %v", err)
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"PromptBuildingError",
			"Failed to build prompt template",
		).WithErrorFromGoError(err))
		logger.Fatal("Failed to build prompt: %v", err)
	}
	logger.Info("Prompt template built successfully")

	// Debug logging of prompt details
	if config.LogLevel == logutil.DebugLevel {
		logger.Debug("Prompt length: %d characters", len(generatedPrompt))
		logger.Debug("Sending task to Gemini: %s", config.TaskDescription)
	}

	// Get token count for confirmation and limit checking
	logger.Info("Checking token limits...")
	logger.Debug("Checking token limits...")
	
	tokenCountEvent := auditlog.NewAuditEvent(
		"INFO",
		"TokenCountCheck",
		"Checking token count and limits",
	).WithMetadata("prompt_length", len(generatedPrompt))
	auditLogger.Log(tokenCountEvent)
	
	tokenInfo, err := getTokenInfo(ctx, geminiClient, generatedPrompt, logger, auditLogger)
	if err != nil {
		logger.Error("Token count check failed")

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Token count check failed: %s", apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			// Log more details in debug mode
			if config.LogLevel == logutil.DebugLevel {
				logger.Debug("Error details: %s", apiErr.DebugInfo())
			}
			
			auditLogger.Log(auditlog.NewAuditEvent(
				"ERROR",
				"TokenCountAPIError",
				"Token count check failed due to API error",
			).WithError(apiErr.Message)
			.WithMetadata("suggestion", apiErr.Suggestion))
		} else {
			logger.Error("Token count check failed: %v", err)
			logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
			
			auditLogger.Log(auditlog.NewAuditEvent(
				"ERROR",
				"TokenCountError",
				"Token count check failed",
			).WithErrorFromGoError(err))
		}

		logger.Fatal("Aborting generation to prevent API errors")
	}

	// If token limit is exceeded, abort
	if tokenInfo.exceedsLimit {
		logger.Error("Token limit exceeded")
		logger.Error("Token limit exceeded: %s", tokenInfo.limitError)
		logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"TokenLimitExceeded",
			"Token count exceeds model limit",
		).WithMetadata("token_count", tokenInfo.tokenCount)
		.WithMetadata("token_limit", tokenInfo.inputLimit)
		.WithMetadata("error", tokenInfo.limitError))
		
		logger.Fatal("Aborting generation to prevent API errors")
	}
	
	logger.Info("Token check passed: %d / %d (%.1f%%)",
		tokenInfo.tokenCount, tokenInfo.inputLimit, tokenInfo.percentage)

	// Log token usage for regular (non-debug) mode
	if config.LogLevel != logutil.DebugLevel {
		logger.Info("Token usage: %d / %d (%.1f%%)",
			tokenInfo.tokenCount,
			tokenInfo.inputLimit,
			tokenInfo.percentage)
	}
	
	// Log token success info
	auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"TokenCheckPassed",
		"Token count is within model limits",
	).WithMetadata("token_count", tokenInfo.tokenCount)
	.WithMetadata("token_limit", tokenInfo.inputLimit)
	.WithMetadata("percentage", tokenInfo.percentage))

	// Prompt for confirmation if threshold is set and exceeded
	if !promptForConfirmation(tokenInfo.tokenCount, config.ConfirmTokens, logger, auditLogger) {
		logger.Info("Operation cancelled by user.")
		
		auditLogger.Log(auditlog.NewAuditEvent(
			"INFO",
			"UserCancellation",
			"User cancelled plan generation",
		).WithMetadata("token_count", tokenInfo.tokenCount)
		.WithMetadata("confirmation_threshold", config.ConfirmTokens))
		
		return
	}

	// Call Gemini API
	logger.Info("Generating plan using model %s...", config.ModelName)
	logger.Debug("Generating plan using model %s...", config.ModelName)
	
	// Log API call
	apiCallEvent := auditlog.NewAuditEvent(
		"INFO",
		"APIRequest",
		"Calling Gemini API for plan generation",
	).WithMetadata("model", config.ModelName)
	auditLogger.Log(apiCallEvent)
	
	result, err := geminiClient.GenerateContent(ctx, generatedPrompt)
	if err != nil {
		logger.Error("Generation failed")

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Error generating content: %s", apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			// Log more details in debug mode
			if config.LogLevel == logutil.DebugLevel {
				logger.Debug("Error details: %s", apiErr.DebugInfo())
			}
			
			auditLogger.Log(auditlog.NewAuditEvent(
				"ERROR",
				"APIError",
				"Error generating plan content",
			).WithError(apiErr.Message)
			.WithMetadata("suggestion", apiErr.Suggestion))
		} else {
			logger.Error("Error generating content: %v", err)
			
			auditLogger.Log(auditlog.NewAuditEvent(
				"ERROR",
				"GenerationError",
				"Error generating plan content",
			).WithErrorFromGoError(err))
		}

		logger.Fatal("Plan generation failed")
	}

	// Process API response
	generatedPlan := processApiResponse(result, logger, auditLogger)
	logger.Info("Plan generated successfully")
	
	// Log successful generation
	successEvent := auditlog.NewAuditEvent(
		"INFO",
		"PlanGenerationSuccess",
		"Successfully generated plan content",
	).WithMetadata("content_length", len(generatedPlan))
	if result.TokenCount > 0 {
		successEvent = successEvent.WithMetadata("output_tokens", result.TokenCount)
	}
	auditLogger.Log(successEvent)

	// Debug logging of results
	if config.LogLevel == logutil.DebugLevel {
		logger.Debug("Plan received from Gemini.")
		if result.TokenCount > 0 {
			logger.Debug("Token usage: %d tokens", result.TokenCount)
		}
	}

	// Write the plan to file
	logger.Info("Writing plan to %s...", config.OutputFile)
	logger.Debug("Writing plan to %s...", config.OutputFile)
	
	// Log file saving
	saveEvent := auditlog.NewAuditEvent(
		"INFO",
		"SavingPlan",
		"Saving generated plan to file",
	).WithMetadata("output_file", config.OutputFile)
	auditLogger.Log(saveEvent)
	
	saveToFile(generatedPlan, config.OutputFile, logger, auditLogger)
	logger.Info("Plan saved to %s", config.OutputFile)
	
	// Log completion
	completionEvent := auditlog.NewAuditEvent(
		"INFO",
		"PlanGenerationComplete",
		"Plan generation process completed successfully",
	).WithMetadata("output_file", config.OutputFile)
	auditLogger.Log(completionEvent)
}

// processApiResponse extracts content from the API response and handles errors
func processApiResponse(result *gemini.GenerationResult, logger logutil.LoggerInterface, auditLogger auditlog.StructuredLogger) string {
	// Check for empty content
	if result.Content == "" {
		// Build an informative error message
		finishReason := ""
		if result.FinishReason != "" {
			finishReason = fmt.Sprintf(" (Finish Reason: %s)", result.FinishReason)
		}

		// Check for safety blocks
		safetyInfo := ""
		safetyBlocked := false
		blockedCategories := []string{}

		if len(result.SafetyRatings) > 0 {
			for _, rating := range result.SafetyRatings {
				if rating.Blocked {
					safetyBlocked = true
					blockedCategories = append(blockedCategories, rating.Category)
					safetyInfo += fmt.Sprintf(" Blocked by Safety Category: %s;", rating.Category)
				}
			}
			if safetyBlocked {
				safetyInfo = " Safety Blocking:" + safetyInfo
			}
		}

		// Log error to audit log
		errorEvent := auditlog.NewAuditEvent(
			"ERROR",
			"EmptyAPIResponse",
			"Received empty response from Gemini API",
		)
		if result.FinishReason != "" {
			errorEvent = errorEvent.WithMetadata("finish_reason", result.FinishReason)
		}
		if safetyBlocked {
			errorEvent = errorEvent.WithMetadata("safety_blocked", true)
			// Add each blocked category
			for i, category := range blockedCategories {
				errorEvent = errorEvent.WithMetadata(fmt.Sprintf("blocked_category_%d", i+1), category)
			}
		}
		auditLogger.Log(errorEvent)

		logger.Fatal("Received empty response from Gemini.%s%s", finishReason, safetyInfo)
	}

	// Check for whitespace-only content
	if strings.TrimSpace(result.Content) == "" {
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR", 
			"EmptyContentResponse",
			"Gemini returned a whitespace-only response",
		))
		logger.Fatal("Gemini returned an empty plan text.")
	}

	// Log successful response processing
	auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"ResponseProcessed",
		"Successfully processed API response",
	).WithMetadata("content_length", len(result.Content)))

	return result.Content
}

// saveToFile writes the generated plan to the specified file
func saveToFile(content string, outputFile string, logger logutil.LoggerInterface, auditLogger auditlog.StructuredLogger) {
	// Log that we're resolving the output path
	auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"ResolvingOutputPath",
		"Resolving output file path",
	).WithMetadata("output_file", outputFile))

	// Resolve the output file path
	outputPath, err := resolvePath(outputFile, "output", logger)
	if err != nil {
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"PathResolutionError",
			"Error resolving output file path",
		).WithErrorFromGoError(err)
		.WithMetadata("output_file", outputFile))
		logger.Fatal("Error resolving output file path: %v", err)
	}

	// Log resolved path
	auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"OutputPathResolved",
		"Output path resolution successful",
	).WithMetadata("resolved_path", outputPath))

	// Ensure the directory exists
	dir := filepath.Dir(outputPath)
	if dir != "." {
		auditLogger.Log(auditlog.NewAuditEvent(
			"INFO",
			"EnsureDirectoryExists",
			"Ensuring output directory exists",
		).WithMetadata("directory", dir))

		if err := os.MkdirAll(dir, 0755); err != nil {
			auditLogger.Log(auditlog.NewAuditEvent(
				"ERROR",
				"DirectoryCreationError",
				"Error creating directory for output file",
			).WithErrorFromGoError(err)
			.WithMetadata("directory", dir))
			logger.Fatal("Error creating directory for output file: %v", err)
		}
	}

	// Write to file
	logger.Info("Writing plan to %s...", outputPath)

	auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"WritingOutput",
		"Writing content to output file",
	).WithMetadata("file_path", outputPath)
	.WithMetadata("content_length", len(content)))

	err = os.WriteFile(outputPath, []byte(content), 0644)
	if err != nil {
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"FileWriteError",
			"Error writing content to file",
		).WithErrorFromGoError(err)
		.WithMetadata("file_path", outputPath))
		logger.Fatal("Error writing plan to file %s: %v", outputPath, err)
	}

	// Log successful file write
	auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"OutputSaved",
		"Successfully saved output to file",
	).WithMetadata("file_path", outputPath))

	logger.Info("Successfully generated plan and saved to %s", outputPath)
}

// initSpinner function removed

// promptForConfirmation asks for user confirmation to proceed
func promptForConfirmation(tokenCount int32, threshold int, logger logutil.LoggerInterface, auditLogger auditlog.StructuredLogger) bool {
	if threshold <= 0 || int32(threshold) > tokenCount {
		// No confirmation needed if threshold is disabled (0) or token count is below threshold
		if threshold > 0 {
			auditLogger.Log(auditlog.NewAuditEvent(
				"INFO",
				"ConfirmationNotNeeded",
				"Token count below confirmation threshold",
			).WithMetadata("token_count", tokenCount)
			.WithMetadata("threshold", threshold))
		}
		return true
	}

	// Log that we're prompting for confirmation
	auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"ConfirmationPrompt",
		"Token count exceeds confirmation threshold, prompting for user confirmation",
	).WithMetadata("token_count", tokenCount)
	.WithMetadata("threshold", threshold))

	logger.Info("Token count (%d) exceeds confirmation threshold (%d).", tokenCount, threshold)
	logger.Info("Do you want to proceed with the API call? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		logger.Error("Error reading input: %v", err)
		
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"UserInputError",
			"Error reading user confirmation input",
		).WithErrorFromGoError(err))
		
		return false
	}

	// Trim whitespace and convert to lowercase
	response = strings.ToLower(strings.TrimSpace(response))

	// Log user response
	confirmed := response == "y" || response == "yes"
	auditLogger.Log(auditlog.NewAuditEvent(
		"INFO",
		"UserConfirmation",
		"User confirmation response received",
	).WithMetadata("confirmed", confirmed)
	.WithMetadata("response", response))

	// Only proceed if the user explicitly confirms with 'y' or 'yes'
	return confirmed
}

// tokenInfoResult holds information about token counts and limits
type tokenInfoResult struct {
	tokenCount   int32
	inputLimit   int32
	exceedsLimit bool
	limitError   string
	percentage   float64
}

// getTokenInfo gets token count information and checks limits
func getTokenInfo(ctx context.Context, geminiClient gemini.Client, prompt string, logger logutil.LoggerInterface, auditLogger auditlog.StructuredLogger) (*tokenInfoResult, error) {
	// Create result structure
	result := &tokenInfoResult{
		exceedsLimit: false,
	}

	// Log that we're getting model info
	modelInfoEvent := auditlog.NewAuditEvent(
		"INFO",
		"FetchingModelInfo",
		"Retrieving model information for token limit check"
	)
	auditLogger.Log(modelInfoEvent)

	// Get model information (limits)
	modelInfo, err := geminiClient.GetModelInfo(ctx)
	if err != nil {
		// Pass through API errors directly for better error messages
		if apiErr, ok := gemini.IsAPIError(err); ok {
			auditLogger.Log(auditlog.NewAuditEvent(
				"ERROR",
				"ModelInfoAPIError",
				"Failed to retrieve model information",
			).WithError(apiErr.Message))
			return nil, err
		}

		// Wrap other errors
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"ModelInfoError",
			"Failed to retrieve model information",
		).WithErrorFromGoError(err))
		return nil, fmt.Errorf("failed to get model info for token limit check: %w", err)
	}

	// Store input limit
	result.inputLimit = modelInfo.InputTokenLimit
	
	// Log model info received
	modelInfoReceivedEvent := auditlog.NewAuditEvent(
		"INFO",
		"ModelInfoReceived",
		"Retrieved model information successfully",
	).WithMetadata("model_name", modelInfo.Name)
	.WithMetadata("input_token_limit", modelInfo.InputTokenLimit)
	.WithMetadata("output_token_limit", modelInfo.OutputTokenLimit)
	auditLogger.Log(modelInfoReceivedEvent)

	// Log that we're counting tokens
	countTokensEvent := auditlog.NewAuditEvent(
		"INFO",
		"CountingTokens",
		"Counting tokens in prompt",
	).WithMetadata("prompt_length", len(prompt))
	auditLogger.Log(countTokensEvent)

	// Count tokens in the prompt
	tokenResult, err := geminiClient.CountTokens(ctx, prompt)
	if err != nil {
		// Pass through API errors directly for better error messages
		if apiErr, ok := gemini.IsAPIError(err); ok {
			auditLogger.Log(auditlog.NewAuditEvent(
				"ERROR",
				"TokenCountAPIError",
				"Failed to count tokens",
			).WithError(apiErr.Message))
			return nil, err
		}

		// Wrap other errors
		auditLogger.Log(auditlog.NewAuditEvent(
			"ERROR",
			"TokenCountError",
			"Failed to count tokens",
		).WithErrorFromGoError(err))
		return nil, fmt.Errorf("failed to count tokens for token limit check: %w", err)
	}

	// Store token count
	result.tokenCount = tokenResult.Total

	// Calculate percentage of limit
	result.percentage = float64(result.tokenCount) / float64(result.inputLimit) * 100

	// Log token usage information
	logger.Debug("Token usage: %d / %d (%.1f%%)",
		result.tokenCount,
		result.inputLimit,
		result.percentage)
		
	// Log token count received
	tokenCountEvent := auditlog.NewAuditEvent(
		"INFO",
		"TokenCountReceived",
		"Received token count information",
	).WithMetadata("token_count", result.tokenCount)
	.WithMetadata("token_limit", result.inputLimit)
	.WithMetadata("percentage", result.percentage)
	auditLogger.Log(tokenCountEvent)

	// Check if the prompt exceeds the token limit
	if result.tokenCount > result.inputLimit {
		result.exceedsLimit = true
		result.limitError = fmt.Sprintf("prompt exceeds token limit (%d tokens > %d token limit)",
			result.tokenCount, result.inputLimit)
			
		// Log token limit exceeded
		auditLogger.Log(auditlog.NewAuditEvent(
			"WARN",
			"TokenLimitExceeded",
			"Prompt exceeds token limit",
		).WithMetadata("token_count", result.tokenCount)
		.WithMetadata("token_limit", result.inputLimit)
		.WithMetadata("excess_tokens", result.tokenCount - result.inputLimit))
	}

	return result, nil
}

// checkTokenLimit verifies that the prompt doesn't exceed the model's token limit
// Deprecated: Use getTokenInfo instead
func checkTokenLimit(ctx context.Context, geminiClient gemini.Client, prompt string, logger logutil.LoggerInterface) error {
	// Create a no-op audit logger for backward compatibility
	noopAuditLogger := auditlog.NewNoopLogger()
	tokenInfo, err := getTokenInfo(ctx, geminiClient, prompt, logger, noopAuditLogger)
	if err != nil {
		return err
	}

	if tokenInfo.exceedsLimit {
		return fmt.Errorf(tokenInfo.limitError)
	}

	return nil
}

// initConfigSystem initializes the configuration system
// The audit logger will be set later after configuration is loaded
func initConfigSystem(logger logutil.LoggerInterface) config.ManagerInterface {
	return config.NewManager(logger)
}

// getCacheDir returns the XDG cache directory for the application
// This can be overridden in tests
var getCacheDir = func() string {
	return filepath.Join(xdg.CacheHome, "architect")
}

// getConfigDir returns the XDG config directory for the application
// This can be overridden in tests
var getConfigDir = func() string {
	return filepath.Join(xdg.ConfigHome, "architect")
}

// resolvePath converts a relative path to an absolute path based on the path type
// Supported path types: "log", "config", "output"
// If the path is already absolute, it is returned unchanged
func resolvePath(path string, pathType string, logger logutil.LoggerInterface) (string, error) {
	// Check if path is empty
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// If path is already absolute, return it as is
	if filepath.IsAbs(path) {
		return path, nil
	}

	// Resolve based on path type
	switch pathType {
	case "log":
		// Log files go to XDG_CACHE_HOME/architect
		cacheDir := getCacheDir()
		resolved := filepath.Join(cacheDir, path)
		logger.Debug("Resolved log path '%s' to '%s'", path, resolved)
		return resolved, nil

	case "config":
		// Config files go to XDG_CONFIG_HOME/architect
		configDir := getConfigDir()
		resolved := filepath.Join(configDir, path)
		logger.Debug("Resolved config path '%s' to '%s'", path, resolved)
		return resolved, nil

	case "output":
		// Output files go to current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
		resolved := filepath.Join(cwd, path)
		logger.Debug("Resolved output path '%s' to '%s'", path, resolved)
		return resolved, nil

	default:
		return "", fmt.Errorf("unsupported path type: %s", pathType)
	}
}

// initAuditLogger initializes and returns a structured audit logger based on the configuration
func initAuditLogger(appConfig *config.AppConfig, logger logutil.LoggerInterface) auditlog.StructuredLogger {
	// Check if audit logging is enabled
	if !appConfig.AuditLogEnabled {
		logger.Debug("Audit logging is disabled, using NoopLogger")
		return auditlog.NewNoopLogger()
	}

	// Determine the log file path
	logPath := appConfig.AuditLogFile
	if logPath == "" {
		// Use default path for audit logs
		logPath = "audit.log"
	}

	// Resolve the path (handles both relative and absolute paths)
	resolvedPath, err := resolvePath(logPath, "log", logger)
	if err != nil {
		logger.Error("Failed to resolve audit log path '%s': %v", logPath, err)
		logger.Warn("Falling back to NoopLogger (audit events will be discarded)")
		return auditlog.NewNoopLogger()
	}

	// Create the audit logger
	auditLogger, err := auditlog.NewFileLogger(resolvedPath)
	if err != nil {
		// Log the error but continue with a NoopLogger
		logger.Error("Failed to create audit log file at %s: %v", resolvedPath, err)
		logger.Warn("Falling back to NoopLogger (audit events will be discarded)")
		return auditlog.NewNoopLogger()
	}

	logger.Info("Audit logging enabled, writing to: %s", resolvedPath)
	return auditLogger
}

// convertConfigToMap converts the CLI Configuration struct to a map for merging with loaded config
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
		"clarify_task":        cliConfig.ClarifyTask,
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
	if !isFlagSet("clarify") {
		config.ClarifyTask = appConfig.ClarifyTask
	}
	if !isFlagSet("prompt-template") && appConfig.Templates.Default != "" {
		config.PromptTemplate = appConfig.Templates.Default
	}

	return &config
}

// isFlagSet checks if a flag was explicitly set on the command line
func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// buildPrompt constructs the prompt string for the Gemini API.
func buildPrompt(config *Configuration, task string, context string, logger logutil.LoggerInterface) (string, error) {
	// Use config-less version for backward compatibility
	return buildPromptWithManager(config, task, context, prompt.NewManager(logger), logger)
}

// buildPromptWithConfig constructs the prompt string using the configuration system
func buildPromptWithConfig(config *Configuration, task string, context string, configManager config.ManagerInterface, logger logutil.LoggerInterface) (string, error) {
	// Create a prompt manager with config support
	promptManager, err := prompt.SetupPromptManagerWithConfig(logger, configManager)
	if err != nil {
		return "", fmt.Errorf("failed to set up prompt manager: %w", err)
	}

	return buildPromptWithManager(config, task, context, promptManager, logger)
}

// buildPromptWithManager constructs the prompt string using the provided prompt manager.
// This function is exported for testing purposes.
func buildPromptWithManager(config *Configuration, task string, context string, promptManager prompt.ManagerInterface, logger logutil.LoggerInterface) (string, error) {
	// Create template data
	data := &prompt.TemplateData{
		Task:    task,
		Context: context, // context already has the <context> tags from fileutil
	}

	// Determine which template to use
	templateName := "default.tmpl"
	if config.PromptTemplate != "" {
		templateName = config.PromptTemplate
		logger.Debug("Using custom prompt template: %s", templateName)
	}

	// Build the prompt (template loading is now handled by the manager)
	generatedPrompt, err := promptManager.BuildPrompt(templateName, data)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt: %w", err)
	}

	return generatedPrompt, nil
}
