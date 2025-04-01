// main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
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

	// Gather context from files
	projectContext := gatherContext(ctx, config, geminiClient, logger)

	// Generate content if not in dry run mode
	if !config.DryRun {
		generateAndSavePlan(ctx, config, geminiClient, projectContext, logger)
	}
}

// parseFlags handles command line argument parsing
func parseFlags() *Configuration {
	config := &Configuration{}

	// Define flags
	taskFlag := flag.String("task", "", "Required: Description of the task or goal for the plan.")
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

	// Set custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s --task \"<your task description>\" [options] <path1> [path2...]\n\n", os.Args[0])
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
	config.OutputFile = *outputFileFlag
	config.ModelName = *modelNameFlag
	config.Verbose = *verboseFlag
	config.UseColors = *useColorsFlag
	config.Include = *includeFlag
	config.Exclude = *excludeFlag
	config.ExcludeNames = *excludeNamesFlag
	config.Format = *formatFlag
	config.DryRun = *dryRunFlag
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

// validateInputs verifies required inputs are provided
func validateInputs(config *Configuration, logger logutil.LoggerInterface) {
	// Check for task description (not required in dry run mode)
	if config.TaskDescription == "" && !config.DryRun {
		logger.Error("--task flag is required (except in dry-run mode).")
		flag.Usage()
		os.Exit(1)
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
	// Log appropriate message based on mode
	if config.DryRun {
		logger.Info("Dry run mode: gathering files that would be included in context...")
	} else {
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
		logger.Fatal("Failed during project context gathering: %v", err)
	}

	// Log warning if no files were processed
	if processedFilesCount == 0 {
		logger.Warn("No files were processed for context. Check paths and filters.")
	}

	// Calculate and log statistics
	charCount, lineCount, tokenCount := fileutil.CalculateStatisticsWithTokenCounting(ctx, geminiClient, projectContext, logger)

	// Handle dry run mode specific output
	if config.DryRun {
		displayDryRunInfo(charCount, lineCount, tokenCount, processedFilesCount, processedFiles, ctx, geminiClient, logger)
	} else if config.LogLevel == logutil.DebugLevel || processedFilesCount > 0 {
		// Normal run mode
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
	if modelInfoErr == nil {
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

	// Construct prompt
	prompt := buildPrompt(config.TaskDescription, projectContext)

	// Debug logging of prompt details
	if config.LogLevel == logutil.DebugLevel {
		logger.Debug("Prompt length: %d characters", len(prompt))
		logger.Debug("Sending task to Gemini: %s", config.TaskDescription)
	}

	// Check token limits before proceeding
	if err := checkTokenLimit(ctx, geminiClient, prompt, logger); err != nil {
		logger.Error("Token limit check failed: %v", err)
		logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		logger.Fatal("Aborting generation to prevent API errors")
	}

	// Call Gemini API
	logger.Info("Generating plan using model %s...", config.ModelName)
	result, err := geminiClient.GenerateContent(ctx, prompt)
	if err != nil {
		logger.Fatal("Error generating content: %v", err)
	}

	// Process API response
	generatedPlan := processApiResponse(result, logger)

	// Debug logging of results
	if config.LogLevel == logutil.DebugLevel {
		logger.Debug("Plan received from Gemini.")
		if result.TokenCount > 0 {
			logger.Debug("Token usage: %d tokens", result.TokenCount)
		}
	}

	// Write the plan to file
	saveToFile(generatedPlan, config.OutputFile, logger)
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

// checkTokenLimit verifies that the prompt doesn't exceed the model's token limit
func checkTokenLimit(ctx context.Context, geminiClient gemini.Client, prompt string, logger logutil.LoggerInterface) error {
	// Get model information (limits)
	modelInfo, err := geminiClient.GetModelInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get model info for token limit check: %w", err)
	}

	// Count tokens in the prompt
	tokenResult, err := geminiClient.CountTokens(ctx, prompt)
	if err != nil {
		return fmt.Errorf("failed to count tokens for token limit check: %w", err)
	}

	// Log token usage information
	logger.Debug("Token usage: %d / %d (%.1f%%)",
		tokenResult.Total,
		modelInfo.InputTokenLimit,
		float64(tokenResult.Total)/float64(modelInfo.InputTokenLimit)*100)

	// Check if the prompt exceeds the token limit
	if tokenResult.Total > modelInfo.InputTokenLimit {
		return fmt.Errorf("prompt exceeds token limit (%d tokens > %d token limit)",
			tokenResult.Total, modelInfo.InputTokenLimit)
	}

	return nil
}

// buildPrompt constructs the prompt string for the Gemini API.
func buildPrompt(task string, context string) string {
	return fmt.Sprintf(`You are an expert software architect and senior engineer.
Your goal is to create a detailed, actionable technical plan in Markdown format.

**Task:**
%s

**Project Context:**
Below is the relevant code context from the project. Analyze it carefully to understand the current state.
%s

**Instructions:**
Based on the task and the provided context, generate a technical plan named PLAN.md. The plan should include the following sections:

1.  **Overview:** Briefly explain the goal of the plan and the changes involved.
2.  **Task Breakdown:** A detailed list of specific, sequential tasks required to implement the feature or fix.
    *   For each task, estimate the effort (e.g., S, M, L) or time.
    *   Mention the primary files/modules likely to be affected.
3.  **Implementation Details:** Provide specific guidance for the more complex tasks. Include:
    *   Key functions, classes, or components to modify or create.
    *   Data structures or API changes needed.
    *   Code snippets or pseudocode where helpful.
4.  **Potential Challenges & Considerations:** Identify possible risks, edge cases, dependencies, or areas needing further investigation.
5.  **Testing Strategy:** Outline how the changes should be tested (unit tests, integration tests, manual testing steps).
6.  **Open Questions:** List any ambiguities or points needing clarification before starting implementation.

Format the entire response as a single Markdown document suitable for direct use as `+"`PLAN.md`"+`. Do not include any introductory or concluding remarks outside the Markdown plan itself. Ensure the markdown is well-formatted.
`, task, context) // context already has the <context> tags from fileutil
}
