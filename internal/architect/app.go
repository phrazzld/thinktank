// Package architect contains the core application logic for the architect tool
package architect

import (
	"context"
	"fmt"
	"os"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// Execute is the main entry point for the core application logic.
// It configures the necessary components and calls the internal run function.
func Execute(
	ctx context.Context,
	cliConfig *CliConfig,
	logger logutil.LoggerInterface,
) error {
	// 1. Read instructions from file
	instructionsContent, err := os.ReadFile(cliConfig.InstructionsFile)
	if err != nil {
		logger.Error("Failed to read instructions file %s: %v", cliConfig.InstructionsFile, err)
		return fmt.Errorf("failed to read instructions file %s: %w", cliConfig.InstructionsFile, err)
	}
	instructions := string(instructionsContent)
	logger.Info("Successfully read instructions from %s", cliConfig.InstructionsFile)

	// 2. Validate inputs
	if err := validateInputs(cliConfig, logger); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	// 3. Initialize API client
	apiService := NewAPIService(logger)
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

	// 4. Create token manager
	tokenManager := NewTokenManager(logger)

	// 5. Create context gatherer
	contextGatherer := NewContextGatherer(logger, cliConfig.DryRun, tokenManager)

	// 6. Create gather config
	gatherConfig := GatherConfig{
		Paths:        cliConfig.Paths,
		Include:      cliConfig.Include,
		Exclude:      cliConfig.Exclude,
		ExcludeNames: cliConfig.ExcludeNames,
		Format:       cliConfig.Format,
		Verbose:      cliConfig.Verbose,
		LogLevel:     cliConfig.LogLevel,
	}

	// 7. Gather context files
	contextFiles, contextStats, err := contextGatherer.GatherContext(ctx, geminiClient, gatherConfig)
	if err != nil {
		return fmt.Errorf("failed during project context gathering: %w", err)
	}

	// 8. Handle dry run mode
	if cliConfig.DryRun {
		err = contextGatherer.DisplayDryRunInfo(ctx, geminiClient, contextStats)
		if err != nil {
			logger.Error("Error displaying dry run information: %v", err)
			return fmt.Errorf("error displaying dry run information: %w", err)
		}
		return nil
	}

	// 9. Stitch prompt
	stitchedPrompt := StitchPrompt(instructions, contextFiles)
	logger.Info("Prompt constructed successfully")
	logger.Debug("Stitched prompt length: %d characters", len(stitchedPrompt))

	// 10. Check token limits
	logger.Info("Checking token limits...")
	tokenInfo, err := tokenManager.GetTokenInfo(ctx, geminiClient, stitchedPrompt)
	if err != nil {
		logger.Error("Token count check failed")

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Token count check failed: %s", apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			logger.Debug("Error details: %s", apiErr.DebugInfo())
		} else {
			logger.Error("Token count check failed: %v", err)
			logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		}

		return fmt.Errorf("token count check failed: %w", err)
	}

	// If token limit is exceeded, abort
	if tokenInfo.ExceedsLimit {
		logger.Error("Token limit exceeded")
		logger.Error("Token limit exceeded: %s", tokenInfo.LimitError)
		logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		return fmt.Errorf("token limit exceeded: %s", tokenInfo.LimitError)
	}

	logger.Info("Token check passed: %d / %d (%.1f%%)",
		tokenInfo.TokenCount, tokenInfo.InputLimit, tokenInfo.Percentage)

	// 11. Generate content
	logger.Info("Generating plan...")
	result, err := geminiClient.GenerateContent(ctx, stitchedPrompt)
	if err != nil {
		logger.Error("Generation failed")

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Error generating content: %s", apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			logger.Debug("Error details: %s", apiErr.DebugInfo())
		} else {
			logger.Error("Error generating content: %v", err)
		}

		return fmt.Errorf("plan generation failed: %w", err)
	}

	// 12. Process API response
	generatedPlan, err := apiService.ProcessResponse(result)
	if err != nil {
		// Get detailed error information
		errorDetails := apiService.GetErrorDetails(err)

		// Provide specific error messages based on error type
		if apiService.IsEmptyResponseError(err) {
			logger.Error("Received empty or invalid response from Gemini API")
			logger.Error("Error details: %s", errorDetails)
			return fmt.Errorf("failed to process API response due to empty content: %w", err)
		} else if apiService.IsSafetyBlockedError(err) {
			logger.Error("Content was blocked by Gemini safety filters")
			logger.Error("Error details: %s", errorDetails)
			return fmt.Errorf("failed to process API response due to safety restrictions: %w", err)
		} else {
			// Generic API error handling
			return fmt.Errorf("failed to process API response: %w", err)
		}
	}
	logger.Info("Plan generated successfully")

	// 13. Create file writer
	fileWriter := NewFileWriter(logger)

	// 14. Save output
	logger.Info("Writing plan to %s...", cliConfig.OutputFile)
	err = fileWriter.SaveToFile(generatedPlan, cliConfig.OutputFile)
	if err != nil {
		return fmt.Errorf("error saving plan to file: %w", err)
	}

	logger.Info("Plan successfully generated and saved to %s", cliConfig.OutputFile)
	return nil
}

// RunInternal is the same as Execute but exported specifically for testing purposes.
// This allows integration tests to use this function directly and inject mocked services.
func RunInternal(
	ctx context.Context,
	cliConfig *CliConfig,
	logger logutil.LoggerInterface,
	configManager config.ManagerInterface,
	apiService APIService,
) error {
	// 1. Read instructions from file
	instructionsContent, err := os.ReadFile(cliConfig.InstructionsFile)
	if err != nil {
		logger.Error("Failed to read instructions file %s: %v", cliConfig.InstructionsFile, err)
		return fmt.Errorf("failed to read instructions file %s: %w", cliConfig.InstructionsFile, err)
	}
	instructions := string(instructionsContent)
	logger.Info("Successfully read instructions from %s", cliConfig.InstructionsFile)

	// 2. Validate inputs
	if err := validateInputs(cliConfig, logger); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	// 3. Initialize the client
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

	// 4. Create token manager
	tokenManager := NewTokenManager(logger)

	// 5. Create context gatherer
	contextGatherer := NewContextGatherer(logger, cliConfig.DryRun, tokenManager)

	// 6. Create gather config
	gatherConfig := GatherConfig{
		Paths:        cliConfig.Paths,
		Include:      cliConfig.Include,
		Exclude:      cliConfig.Exclude,
		ExcludeNames: cliConfig.ExcludeNames,
		Format:       cliConfig.Format,
		Verbose:      cliConfig.Verbose,
		LogLevel:     cliConfig.LogLevel,
	}

	// 7. Gather context files
	contextFiles, contextStats, err := contextGatherer.GatherContext(ctx, geminiClient, gatherConfig)
	if err != nil {
		return fmt.Errorf("failed during project context gathering: %w", err)
	}

	// 8. Handle dry run mode
	if cliConfig.DryRun {
		err = contextGatherer.DisplayDryRunInfo(ctx, geminiClient, contextStats)
		if err != nil {
			logger.Error("Error displaying dry run information: %v", err)
			return fmt.Errorf("error displaying dry run information: %w", err)
		}
		return nil
	}

	// 9. Stitch prompt
	stitchedPrompt := StitchPrompt(instructions, contextFiles)
	logger.Info("Prompt constructed successfully")
	logger.Debug("Stitched prompt length: %d characters", len(stitchedPrompt))

	// 10. Check token limits
	logger.Info("Checking token limits...")
	tokenInfo, err := tokenManager.GetTokenInfo(ctx, geminiClient, stitchedPrompt)
	if err != nil {
		logger.Error("Token count check failed")

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Token count check failed: %s", apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			logger.Debug("Error details: %s", apiErr.DebugInfo())
		} else {
			logger.Error("Token count check failed: %v", err)
			logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		}

		return fmt.Errorf("token count check failed: %w", err)
	}

	// If token limit is exceeded, abort
	if tokenInfo.ExceedsLimit {
		logger.Error("Token limit exceeded")
		logger.Error("Token limit exceeded: %s", tokenInfo.LimitError)
		logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		return fmt.Errorf("token limit exceeded: %s", tokenInfo.LimitError)
	}

	logger.Info("Token check passed: %d / %d (%.1f%%)",
		tokenInfo.TokenCount, tokenInfo.InputLimit, tokenInfo.Percentage)

	// 11. Generate content
	logger.Info("Generating plan...")
	result, err := geminiClient.GenerateContent(ctx, stitchedPrompt)
	if err != nil {
		logger.Error("Generation failed")

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Error generating content: %s", apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			logger.Debug("Error details: %s", apiErr.DebugInfo())
		} else {
			logger.Error("Error generating content: %v", err)
		}

		return fmt.Errorf("plan generation failed: %w", err)
	}

	// 12. Process API response
	generatedPlan, err := apiService.ProcessResponse(result)
	if err != nil {
		// Get detailed error information
		errorDetails := apiService.GetErrorDetails(err)

		// Provide specific error messages based on error type
		if apiService.IsEmptyResponseError(err) {
			logger.Error("Received empty or invalid response from Gemini API")
			logger.Error("Error details: %s", errorDetails)
			return fmt.Errorf("failed to process API response due to empty content: %w", err)
		} else if apiService.IsSafetyBlockedError(err) {
			logger.Error("Content was blocked by Gemini safety filters")
			logger.Error("Error details: %s", errorDetails)
			return fmt.Errorf("failed to process API response due to safety restrictions: %w", err)
		} else {
			// Generic API error handling
			return fmt.Errorf("failed to process API response: %w", err)
		}
	}
	logger.Info("Plan generated successfully")

	// 13. Create file writer
	fileWriter := NewFileWriter(logger)

	// 14. Save output
	logger.Info("Writing plan to %s...", cliConfig.OutputFile)
	err = fileWriter.SaveToFile(generatedPlan, cliConfig.OutputFile)
	if err != nil {
		return fmt.Errorf("error saving plan to file: %w", err)
	}

	logger.Info("Plan successfully generated and saved to %s", cliConfig.OutputFile)
	return nil
}

// Note: HandleSpecialCommands and processTaskInput functions have been removed
// as part of the refactoring to simplify the core application flow.
// The functionality has been replaced with direct reading of the instructions file
// and the prompt stitching logic.

// validateInputs verifies that all required inputs are provided
func validateInputs(cliConfig *CliConfig, logger logutil.LoggerInterface) error {
	// Skip validation in dry-run mode
	if cliConfig.DryRun {
		return nil
	}

	// Validate instructions file
	if cliConfig.InstructionsFile == "" {
		return fmt.Errorf("instructions file is required (use --instructions)")
	}

	// Validate paths
	if len(cliConfig.Paths) == 0 {
		return fmt.Errorf("at least one file or directory path must be provided")
	}

	// Validate API key
	if cliConfig.ApiKey == "" {
		return fmt.Errorf("%s environment variable not set", APIKeyEnvVar)
	}

	return nil
}
