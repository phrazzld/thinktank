// Package architect provides the command-line interface for the architect tool
package architect

import (
	"context"
	"fmt"
	"os"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// APIServiceFactory defines a factory function type for creating APIService instances
type APIServiceFactory func(logger logutil.LoggerInterface) APIService

// Main is the entry point for the architect CLI
func Main() {
	// Create a base context
	ctx := context.Background()

	// Parse command line flags
	cliConfig, err := ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Setup logging early for error reporting
	logger := SetupLogging(cliConfig)
	logger.Info("Starting Architect - AI-assisted planning tool")

	// Initialize XDG-compliant configuration system
	configManager := config.NewManager(logger)

	// Load configuration from files
	err = configManager.LoadFromFiles()
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
		return
	}

	// Convert CLI flags to the format needed for merging
	cliFlags := ConvertConfigToMap(cliConfig)

	// Merge CLI flags with loaded configuration
	if err := configManager.MergeWithFlags(cliFlags); err != nil {
		logger.Warn("Failed to merge CLI flags with configuration: %v", err)
	}

	// Get the final configuration
	_ = configManager.GetConfig()

	// Call the core Run function with the real API service factory
	err = Run(ctx, cliConfig, logger, configManager, NewAPIService)
	if err != nil {
		logger.Error("Application failed: %v", err)
		os.Exit(1)
	}
}

// Run contains the core application logic, separated from the Main function
// to improve testability. It takes dependencies as parameters for better
// dependency injection support during testing.
func Run(
	ctx context.Context,
	cliConfig *CliConfig,
	logger logutil.LoggerInterface,
	configManager config.ManagerInterface,
	apiServiceFactory APIServiceFactory,
) error {
	// Process task input (from file or flag)
	taskDescription, err := processTaskInput(cliConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to process task input: %w", err)
	}

	// Validate inputs
	if err := validateInputs(cliConfig, taskDescription, logger); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	// Initialize API client
	apiService := apiServiceFactory(logger)
	geminiClient, err := apiService.InitClient(ctx, cliConfig.ApiKey, cliConfig.ModelName)
	if err != nil {
		// Handle API client initialization errors
		errorDetails := apiService.GetErrorDetails(err)
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Error creating Gemini client: %s", apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			if cliConfig.LogLevel == logutil.DebugLevel {
				logger.Debug("Error details: %s", apiErr.DebugInfo())
			}
		} else {
			logger.Error("Error creating Gemini client: %s", errorDetails)
		}
		return fmt.Errorf("failed to initialize API client: %w", err)
	}
	defer geminiClient.Close()

	// Create token manager
	tokenManager := NewTokenManager(logger)

	// Create context gatherer
	contextGatherer := NewContextGatherer(logger, cliConfig.DryRun, tokenManager)

	// Create gather config
	gatherConfig := GatherConfig{
		Paths:        cliConfig.Paths,
		Include:      cliConfig.Include,
		Exclude:      cliConfig.Exclude,
		ExcludeNames: cliConfig.ExcludeNames,
		Format:       cliConfig.Format,
		Verbose:      cliConfig.Verbose,
		LogLevel:     cliConfig.LogLevel,
	}

	// Gather context from files
	projectContext, contextStats, err := contextGatherer.GatherContext(ctx, geminiClient, gatherConfig)
	if err != nil {
		return fmt.Errorf("failed during project context gathering: %w", err)
	}

	// Handle dry run mode
	if cliConfig.DryRun {
		err = contextGatherer.DisplayDryRunInfo(ctx, geminiClient, contextStats)
		if err != nil {
			logger.Error("Error displaying dry run information: %v", err)
			return fmt.Errorf("error displaying dry run information: %w", err)
		}
		return nil
	}

	// Create output writer
	outputWriter := NewOutputWriter(logger, tokenManager)

	// Generate and save plan with configuration
	err = outputWriter.GenerateAndSavePlanWithConfig(
		ctx,
		geminiClient,
		taskDescription,
		projectContext,
		cliConfig.OutputFile,
		configManager,
	)
	if err != nil {
		return fmt.Errorf("error generating and saving plan: %w", err)
	}

	logger.Info("Plan successfully generated and saved to %s", cliConfig.OutputFile)
	return nil
}

// handleSpecialCommands processes special command flags like list-examples and show-example
// Returns true if a special command was executed
func handleSpecialCommands(cliConfig *CliConfig, logger logutil.LoggerInterface, configManager config.ManagerInterface) bool {
	// Create prompt builder
	promptBuilder := NewPromptBuilder(logger)

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
func processTaskInput(cliConfig *CliConfig, logger logutil.LoggerInterface) (string, error) {
	// If task file is provided, read from file
	if cliConfig.TaskFile != "" {
		// Create prompt builder
		promptBuilder := NewPromptBuilder(logger)

		// Read task from file using prompt builder
		content, err := promptBuilder.ReadTaskFromFile(cliConfig.TaskFile)
		if err != nil {
			return "", fmt.Errorf("failed to read task from file: %w", err)
		}
		return content, nil
	}

	// Otherwise, use the task description from CLI flags
	return cliConfig.TaskDescription, nil
}

// validateInputs verifies that all required inputs are provided
func validateInputs(cliConfig *CliConfig, taskDescription string, logger logutil.LoggerInterface) error {
	// Skip validation in dry-run mode
	if cliConfig.DryRun {
		return nil
	}

	// Validate task description
	if taskDescription == "" {
		return fmt.Errorf("task description is required (use --task or --task-file)")
	}

	// Validate paths
	if len(cliConfig.Paths) == 0 {
		return fmt.Errorf("at least one file or directory path must be provided")
	}

	// Validate API key
	if cliConfig.ApiKey == "" {
		return fmt.Errorf("%s environment variable not set", apiKeyEnvVar)
	}

	return nil
}
