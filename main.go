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

	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/prompt"
	"github.com/phrazzld/architect/internal/spinner"
)

// Default constants
const (
	defaultOutputFile = "PLAN.md"
	defaultModel      = "gemini-2.5-pro-exp-03-25"
	apiKeyEnvVar      = "GEMINI_API_KEY"
	defaultFormat     = "<{path}>\n```\n{content}\n```\n</{path}>\n\n"

	// Default excludes inspired by common project types
	defaultExcludes = ".exe,.bin,.obj,.o,.a,.lib,.so,.dll,.dylib,.class,.jar,.pyc,.pyo,.pyd," +
		".zip,.tar,.gz,.rar,.7z,.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.odt,.ods,.odp," +
		".jpg,.jpeg,.png,.gif,.bmp,.tiff,.svg,.mp3,.wav,.ogg,.mp4,.avi,.mov,.wmv,.flv," +
		".iso,.img,.dmg,.db,.sqlite,.log"

	defaultExcludeNames = ".git,.hg,.svn,node_modules,bower_components,vendor,target,dist,build," +
		"out,tmp,coverage,__pycache__,*.pyc,*.pyo,.DS_Store,~$*,desktop.ini,Thumbs.db," +
		"package-lock.json,yarn.lock,go.sum,go.work"
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
	NoSpinner       bool
	ClarifyTask     bool
	Paths           []string
	ApiKey          string
}

func main() {
	// Parse command line flags
	config := parseFlags()

	// Setup logging
	logger := setupLogging(config)

	// Validate inputs
	validateInputs(config, logger)

	// Initialize API client
	ctx := context.Background()
	geminiClient := initGeminiClient(ctx, config, logger)
	defer geminiClient.Close()

	// If task clarification is enabled, let the user refine their task
	if config.ClarifyTask && !config.DryRun {
		config.TaskDescription = clarifyTaskDescription(ctx, config, geminiClient, logger)
	}

	// Gather context from files
	projectContext := gatherContext(ctx, config, geminiClient, logger)

	// Generate content if not in dry run mode
	if !config.DryRun {
		generateAndSavePlan(ctx, config, geminiClient, projectContext, logger)
	}
}

// clarifyTaskDescription performs an interactive process to refine the user's task description
func clarifyTaskDescription(ctx context.Context, config *Configuration, geminiClient gemini.Client, logger logutil.LoggerInterface) string {
	// Initialize spinner
	spinnerInstance := initSpinner(config, logger)

	// Initialize prompt manager for clarification templates
	promptManager := prompt.NewManager(logger)

	// Original task description
	originalTask := config.TaskDescription

	// Load clarification template
	spinnerInstance.Start("Analyzing task description...")
	err := promptManager.LoadTemplate("clarify.tmpl")
	if err != nil {
		spinnerInstance.StopFail(fmt.Sprintf("Failed to load clarification template: %v", err))
		logger.Error("Failed to load clarification template: %v", err)
		return originalTask
	}

	// Build prompt for clarification
	data := &prompt.TemplateData{
		Task: originalTask,
	}

	clarifyPrompt, err := promptManager.BuildPrompt("clarify.tmpl", data)
	if err != nil {
		spinnerInstance.StopFail(fmt.Sprintf("Failed to build clarification prompt: %v", err))
		logger.Error("Failed to build clarification prompt: %v", err)
		return originalTask
	}

	// Call Gemini to generate clarification questions
	spinnerInstance.UpdateMessage("Generating clarification questions...")
	result, err := geminiClient.GenerateContent(ctx, clarifyPrompt)
	if err != nil {
		spinnerInstance.StopFail(fmt.Sprintf("Error generating clarification questions: %v", err))
		logger.Error("Error generating clarification questions: %v", err)
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
		spinnerInstance.StopFail(fmt.Sprintf("Failed to parse clarification response: %v", err))
		logger.Error("Failed to parse clarification response: %v", err)
		logger.Debug("Response content: %s", result.Content)
		return originalTask
	}

	// Stop spinner and start the interactive clarification process
	spinnerInstance.Stop("Task analysis complete")

	// Show the analysis to the user
	logger.Info("Task Analysis: %s", clarificationData.Analysis)

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
			return originalTask
		}

		// Add the Q&A to our collection
		questionAnswers.WriteString(fmt.Sprintf("Question %d: %s\n", i+1, question))
		questionAnswers.WriteString(fmt.Sprintf("Answer %d: %s\n", i+1, strings.TrimSpace(answer)))
	}

	// Now refine the task with the answers
	spinnerInstance.Start("Refining task description...")

	// Load the refine template
	err = promptManager.LoadTemplate("refine.tmpl")
	if err != nil {
		spinnerInstance.StopFail(fmt.Sprintf("Failed to load refinement template: %v", err))
		logger.Error("Failed to load refinement template: %v", err)
		return originalTask
	}

	// Build prompt for refinement
	refineData := &prompt.TemplateData{
		Task:    originalTask,
		Context: questionAnswers.String(),
	}

	refinePrompt, err := promptManager.BuildPrompt("refine.tmpl", refineData)
	if err != nil {
		spinnerInstance.StopFail(fmt.Sprintf("Failed to build refinement prompt: %v", err))
		logger.Error("Failed to build refinement prompt: %v", err)
		return originalTask
	}

	// Call Gemini to generate refined task
	result, err = geminiClient.GenerateContent(ctx, refinePrompt)
	if err != nil {
		spinnerInstance.StopFail(fmt.Sprintf("Error generating refined task: %v", err))
		logger.Error("Error generating refined task: %v", err)
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
		spinnerInstance.StopFail(fmt.Sprintf("Failed to parse refinement response: %v", err))
		logger.Error("Failed to parse refinement response: %v", err)
		logger.Debug("Response content: %s", result.Content)
		return originalTask
	}

	// Stop spinner and show the refined task
	spinnerInstance.Stop("Task refinement complete")

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
	noSpinnerFlag := flag.Bool("no-spinner", false, "Disable spinner animation during API calls")
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
	config.NoSpinner = *noSpinnerFlag
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
	// Check if path is absolute, if not make it absolute
	if !filepath.IsAbs(taskFilePath) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("error getting current working directory: %w", err)
		}
		taskFilePath = filepath.Join(cwd, taskFilePath)
	}

	// Check if file exists
	if _, err := os.Stat(taskFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("task file not found: %s", taskFilePath)
	}

	// Read file content
	content, err := os.ReadFile(taskFilePath)
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
func gatherContext(ctx context.Context, config *Configuration, geminiClient gemini.Client, logger logutil.LoggerInterface) string {
	// Initialize spinner
	spinnerInstance := initSpinner(config, logger)

	// Log appropriate message based on mode and start spinner
	if config.DryRun {
		spinnerInstance.Start("Gathering files that would be included in context...")
		logger.Info("Dry run mode: gathering files that would be included in context...")
	} else {
		spinnerInstance.Start("Gathering project context...")
		logger.Info("Gathering project context...")
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
		spinnerInstance.StopFail(fmt.Sprintf("Failed during project context gathering: %v", err))
		logger.Fatal("Failed during project context gathering: %v", err)
	}

	// Log warning if no files were processed
	if processedFilesCount == 0 {
		spinnerInstance.Stop("No files were processed for context. Check paths and filters.")
		logger.Warn("No files were processed for context. Check paths and filters.")
		return projectContext
	}

	// Update spinner message and calculate statistics
	spinnerInstance.UpdateMessage("Calculating token statistics...")
	charCount, lineCount, tokenCount := fileutil.CalculateStatisticsWithTokenCounting(ctx, geminiClient, projectContext, logger)

	// Handle dry run mode specific output
	if config.DryRun {
		spinnerInstance.Stop(fmt.Sprintf("Context gathered: %d files, %d lines, %d chars, %d tokens",
			processedFilesCount, lineCount, charCount, tokenCount))
		displayDryRunInfo(charCount, lineCount, tokenCount, processedFilesCount, processedFiles, ctx, geminiClient, logger)
	} else if config.LogLevel == logutil.DebugLevel || processedFilesCount > 0 {
		// Normal run mode
		spinnerInstance.Stop(fmt.Sprintf("Context gathered: %d files, %d lines, %d chars, %d tokens",
			processedFilesCount, lineCount, charCount, tokenCount))
		logger.Info("Context gathered: %d files processed, %d lines, %d chars, %d tokens.",
			processedFilesCount, lineCount, charCount, tokenCount)
	}

	return projectContext
}

// displayDryRunInfo shows detailed information for dry run mode
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

// generateAndSavePlan creates and saves the plan to a file
func generateAndSavePlan(ctx context.Context, config *Configuration, geminiClient gemini.Client,
	projectContext string, logger logutil.LoggerInterface) {

	// Initialize spinner
	spinnerInstance := initSpinner(config, logger)

	// Construct prompt
	spinnerInstance.Start("Building prompt template...")
	prompt, err := buildPrompt(config, config.TaskDescription, projectContext, logger)
	if err != nil {
		spinnerInstance.StopFail(fmt.Sprintf("Failed to build prompt: %v", err))
		logger.Fatal("Failed to build prompt: %v", err)
	}
	spinnerInstance.Stop("Prompt template built successfully")

	// Debug logging of prompt details
	if config.LogLevel == logutil.DebugLevel {
		logger.Debug("Prompt length: %d characters", len(prompt))
		logger.Debug("Sending task to Gemini: %s", config.TaskDescription)
	}

	// Get token count for confirmation and limit checking
	spinnerInstance.Start("Checking token limits...")
	tokenInfo, err := getTokenInfo(ctx, geminiClient, prompt, logger)
	if err != nil {
		spinnerInstance.StopFail("Token count check failed")

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
		} else {
			logger.Error("Token count check failed: %v", err)
			logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		}

		logger.Fatal("Aborting generation to prevent API errors")
	}

	// If token limit is exceeded, abort
	if tokenInfo.exceedsLimit {
		spinnerInstance.StopFail("Token limit exceeded")
		logger.Error("Token limit exceeded: %s", tokenInfo.limitError)
		logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		logger.Fatal("Aborting generation to prevent API errors")
	}
	spinnerInstance.Stop(fmt.Sprintf("Token check passed: %d / %d (%.1f%%)",
		tokenInfo.tokenCount, tokenInfo.inputLimit, tokenInfo.percentage))

	// Log token usage for regular (non-debug) mode
	if config.LogLevel != logutil.DebugLevel {
		logger.Info("Token usage: %d / %d (%.1f%%)",
			tokenInfo.tokenCount,
			tokenInfo.inputLimit,
			tokenInfo.percentage)
	}

	// Prompt for confirmation if threshold is set and exceeded
	if !promptForConfirmation(tokenInfo.tokenCount, config.ConfirmTokens, logger) {
		logger.Info("Operation cancelled by user.")
		return
	}

	// Call Gemini API
	spinnerInstance.Start(fmt.Sprintf("Generating plan using model %s...", config.ModelName))
	result, err := geminiClient.GenerateContent(ctx, prompt)
	if err != nil {
		spinnerInstance.StopFail("Generation failed")

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
		} else {
			logger.Error("Error generating content: %v", err)
		}

		logger.Fatal("Plan generation failed")
	}

	// Process API response
	generatedPlan := processApiResponse(result, logger)
	spinnerInstance.Stop("Plan generated successfully")

	// Debug logging of results
	if config.LogLevel == logutil.DebugLevel {
		logger.Debug("Plan received from Gemini.")
		if result.TokenCount > 0 {
			logger.Debug("Token usage: %d tokens", result.TokenCount)
		}
	}

	// Write the plan to file
	spinnerInstance.Start(fmt.Sprintf("Writing plan to %s...", config.OutputFile))
	saveToFile(generatedPlan, config.OutputFile, logger)
	spinnerInstance.Stop(fmt.Sprintf("Plan saved to %s", config.OutputFile))
}

// processApiResponse extracts content from the API response and handles errors
func processApiResponse(result *gemini.GenerationResult, logger logutil.LoggerInterface) string {
	// Check for empty content
	if result.Content == "" {
		// Build an informative error message
		finishReason := ""
		if result.FinishReason != "" {
			finishReason = fmt.Sprintf(" (Finish Reason: %s)", result.FinishReason)
		}

		// Check for safety blocks
		safetyInfo := ""
		if len(result.SafetyRatings) > 0 {
			blocked := false
			for _, rating := range result.SafetyRatings {
				if rating.Blocked {
					blocked = true
					safetyInfo += fmt.Sprintf(" Blocked by Safety Category: %s;", rating.Category)
				}
			}
			if blocked {
				safetyInfo = " Safety Blocking:" + safetyInfo
			}
		}

		logger.Fatal("Received empty response from Gemini.%s%s", finishReason, safetyInfo)
	}

	// Check for whitespace-only content
	if strings.TrimSpace(result.Content) == "" {
		logger.Fatal("Gemini returned an empty plan text.")
	}

	return result.Content
}

// saveToFile writes the generated plan to the specified file
func saveToFile(content string, outputFile string, logger logutil.LoggerInterface) {
	// Ensure output path is absolute
	outputPath := outputFile
	if !filepath.IsAbs(outputPath) {
		cwd, err := os.Getwd()
		if err != nil {
			logger.Fatal("Error getting current working directory: %v", err)
		}
		outputPath = filepath.Join(cwd, outputPath)
	}

	// Write to file
	logger.Info("Writing plan to %s...", outputPath)
	err := os.WriteFile(outputPath, []byte(content), 0644)
	if err != nil {
		logger.Fatal("Error writing plan to file %s: %v", outputPath, err)
	}

	logger.Info("Successfully generated plan and saved to %s", outputPath)
}

// initSpinner creates and configures a spinner instance based on config
func initSpinner(config *Configuration, logger logutil.LoggerInterface) *spinner.Spinner {
	// Configure spinner options based on user preferences
	spinnerOptions := &spinner.Options{
		Enabled:     !config.NoSpinner,
		CharSet:     14,  // Use circle dots by default
		RefreshRate: 100, // ms
		Output:      os.Stdout,
		Prefix:      " ",
		Suffix:      " ",
	}

	// If debug mode, disable spinner to avoid cluttering logs
	if config.LogLevel == logutil.DebugLevel {
		spinnerOptions.Enabled = false
	}

	// Create spinner instance
	return spinner.New(logger, spinnerOptions)
}

// promptForConfirmation asks for user confirmation to proceed
func promptForConfirmation(tokenCount int32, threshold int, logger logutil.LoggerInterface) bool {
	if threshold <= 0 || int32(threshold) > tokenCount {
		// No confirmation needed if threshold is disabled (0) or token count is below threshold
		return true
	}

	logger.Info("Token count (%d) exceeds confirmation threshold (%d).", tokenCount, threshold)
	logger.Info("Do you want to proceed with the API call? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		logger.Error("Error reading input: %v", err)
		return false
	}

	// Trim whitespace and convert to lowercase
	response = strings.ToLower(strings.TrimSpace(response))

	// Only proceed if the user explicitly confirms with 'y' or 'yes'
	return response == "y" || response == "yes"
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
func getTokenInfo(ctx context.Context, geminiClient gemini.Client, prompt string, logger logutil.LoggerInterface) (*tokenInfoResult, error) {
	// Create result structure
	result := &tokenInfoResult{
		exceedsLimit: false,
	}

	// Get model information (limits)
	modelInfo, err := geminiClient.GetModelInfo(ctx)
	if err != nil {
		// Pass through API errors directly for better error messages
		if _, ok := gemini.IsAPIError(err); ok {
			return nil, err
		}

		// Wrap other errors
		return nil, fmt.Errorf("failed to get model info for token limit check: %w", err)
	}

	// Store input limit
	result.inputLimit = modelInfo.InputTokenLimit

	// Count tokens in the prompt
	tokenResult, err := geminiClient.CountTokens(ctx, prompt)
	if err != nil {
		// Pass through API errors directly for better error messages
		if _, ok := gemini.IsAPIError(err); ok {
			return nil, err
		}

		// Wrap other errors
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

	// Check if the prompt exceeds the token limit
	if result.tokenCount > result.inputLimit {
		result.exceedsLimit = true
		result.limitError = fmt.Sprintf("prompt exceeds token limit (%d tokens > %d token limit)",
			result.tokenCount, result.inputLimit)
	}

	return result, nil
}

// checkTokenLimit verifies that the prompt doesn't exceed the model's token limit
// Deprecated: Use getTokenInfo instead
func checkTokenLimit(ctx context.Context, geminiClient gemini.Client, prompt string, logger logutil.LoggerInterface) error {
	tokenInfo, err := getTokenInfo(ctx, geminiClient, prompt, logger)
	if err != nil {
		return err
	}

	if tokenInfo.exceedsLimit {
		return fmt.Errorf(tokenInfo.limitError)
	}

	return nil
}

// buildPrompt constructs the prompt string for the Gemini API.
func buildPrompt(config *Configuration, task string, context string, logger logutil.LoggerInterface) (string, error) {
	return buildPromptWithManager(config, task, context, prompt.NewManager(logger), logger)
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

	// Try to load the template
	err := promptManager.LoadTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt template: %w", err)
	}

	// Build the prompt
	generatedPrompt, err := promptManager.BuildPrompt(filepath.Base(templateName), data)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt: %w", err)
	}

	return generatedPrompt, nil
}
