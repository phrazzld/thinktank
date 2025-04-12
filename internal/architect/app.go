// Package architect contains the core application logic for the architect tool
package architect

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// Execute is the main entry point for the core application logic.
// It configures the necessary components and calls the internal run function.
func Execute(
	ctx context.Context,
	cliConfig *CliConfig,
	logger logutil.LoggerInterface,
	auditLogger auditlog.AuditLogger,
) (err error) {
	// Use a deferred function to ensure ExecuteEnd is always logged
	defer func() {
		status := "Success"
		var errorInfo *auditlog.ErrorInfo
		if err != nil {
			status = "Failure"
			errorInfo = &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "ExecutionError",
			}
		}

		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "ExecuteEnd",
			Status:    status,
			Error:     errorInfo,
			Message:   fmt.Sprintf("Execution completed with status: %s", status),
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}
	}()
	// Log the start of the Execute operation
	inputs := map[string]interface{}{
		"instructions_file": cliConfig.InstructionsFile,
		"output_file":       cliConfig.OutputFile,
		"audit_log_file":    cliConfig.AuditLogFile,
		"format":            cliConfig.Format,
		"paths_count":       len(cliConfig.Paths),
		"include":           cliConfig.Include,
		"exclude":           cliConfig.Exclude,
		"exclude_names":     cliConfig.ExcludeNames,
		"dry_run":           cliConfig.DryRun,
		"verbose":           cliConfig.Verbose,
		"model_name":        cliConfig.ModelName,
		"confirm_tokens":    cliConfig.ConfirmTokens,
		"log_level":         cliConfig.LogLevel,
	}

	if err := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: "ExecuteStart",
		Status:    "InProgress",
		Inputs:    inputs,
		Message:   "Starting execution of architect tool",
	}); err != nil {
		logger.Error("Failed to write audit log: %v", err)
	}

	// 1. Read instructions from file
	instructionsContent, err := os.ReadFile(cliConfig.InstructionsFile)
	if err != nil {
		logger.Error("Failed to read instructions file %s: %v", cliConfig.InstructionsFile, err)

		// Log the failure to read the instructions file to the audit log
		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "ReadInstructions",
			Status:    "Failure",
			Inputs: map[string]interface{}{
				"path": cliConfig.InstructionsFile,
			},
			Error: &auditlog.ErrorInfo{
				Message: fmt.Sprintf("Failed to read instructions file: %v", err),
				Type:    "FileIOError",
			},
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("failed to read instructions file %s: %w", cliConfig.InstructionsFile, err)
	}
	instructions := string(instructionsContent)
	logger.Info("Successfully read instructions from %s", cliConfig.InstructionsFile)

	// Log the successful reading of the instructions file to the audit log
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: "ReadInstructions",
		Status:    "Success",
		Inputs: map[string]interface{}{
			"path": cliConfig.InstructionsFile,
		},
		Outputs: map[string]interface{}{
			"content_length": len(instructions),
		},
		Message: "Successfully read instructions file",
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

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
	// Log the start of context gathering
	gatherStartTime := time.Now()
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: gatherStartTime,
		Operation: "GatherContextStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"paths":         cliConfig.Paths,
			"include":       cliConfig.Include,
			"exclude":       cliConfig.Exclude,
			"exclude_names": cliConfig.ExcludeNames,
			"format":        cliConfig.Format,
		},
		Message: "Starting to gather project context files",
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

	contextFiles, contextStats, err := contextGatherer.GatherContext(ctx, geminiClient, gatherConfig)

	// Calculate duration in milliseconds
	gatherDurationMs := time.Since(gatherStartTime).Milliseconds()

	if err != nil {
		// Log the failure of context gathering
		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp:  time.Now().UTC(),
			Operation:  "GatherContextEnd",
			Status:     "Failure",
			DurationMs: &gatherDurationMs,
			Inputs: map[string]interface{}{
				"paths":         cliConfig.Paths,
				"include":       cliConfig.Include,
				"exclude":       cliConfig.Exclude,
				"exclude_names": cliConfig.ExcludeNames,
			},
			Error: &auditlog.ErrorInfo{
				Message: fmt.Sprintf("Failed to gather project context: %v", err),
				Type:    "ContextGatheringError",
			},
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}
		return fmt.Errorf("failed during project context gathering: %w", err)
	}

	// Log the successful completion of context gathering
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp:  time.Now().UTC(),
		Operation:  "GatherContextEnd",
		Status:     "Success",
		DurationMs: &gatherDurationMs,
		Inputs: map[string]interface{}{
			"paths":         cliConfig.Paths,
			"include":       cliConfig.Include,
			"exclude":       cliConfig.Exclude,
			"exclude_names": cliConfig.ExcludeNames,
		},
		Outputs: map[string]interface{}{
			"processed_files_count": contextStats.ProcessedFilesCount,
			"char_count":            contextStats.CharCount,
			"line_count":            contextStats.LineCount,
			"token_count":           contextStats.TokenCount,
			"files_count":           len(contextFiles),
		},
		Message: "Successfully gathered project context files",
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
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

	// Log token check start with prompt information
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: "CheckTokensStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"prompt_length": len(stitchedPrompt),
			"model_name":    cliConfig.ModelName,
		},
		Message: "Starting token count check",
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

	tokenInfo, err := tokenManager.GetTokenInfo(ctx, geminiClient, stitchedPrompt)

	if err != nil {
		logger.Error("Token count check failed")

		// Determine error type for better categorization
		errorType := "TokenCheckError"
		errorMessage := fmt.Sprintf("Failed to check token count: %v", err)

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Token count check failed: %s", apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			logger.Debug("Error details: %s", apiErr.DebugInfo())
			errorType = "APIError"
			errorMessage = apiErr.Message
		} else {
			logger.Error("Token count check failed: %v", err)
			logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		}

		// Log the token check failure
		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "CheckTokens",
			Status:    "Failure",
			Inputs: map[string]interface{}{
				"prompt_length": len(stitchedPrompt),
				"model_name":    cliConfig.ModelName,
			},
			Error: &auditlog.ErrorInfo{
				Message: errorMessage,
				Type:    errorType,
			},
			Message: "Token count check failed",
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("token count check failed: %w", err)
	}

	// If token limit is exceeded, abort
	if tokenInfo.ExceedsLimit {
		logger.Error("Token limit exceeded")
		logger.Error("Token limit exceeded: %s", tokenInfo.LimitError)
		logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")

		// Log the token limit exceeded case
		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "CheckTokens",
			Status:    "Failure",
			Inputs: map[string]interface{}{
				"prompt_length": len(stitchedPrompt),
				"model_name":    cliConfig.ModelName,
			},
			TokenCounts: &auditlog.TokenCountInfo{
				PromptTokens: tokenInfo.TokenCount,
				TotalTokens:  tokenInfo.TokenCount,
				Limit:        tokenInfo.InputLimit,
			},
			Error: &auditlog.ErrorInfo{
				Message: tokenInfo.LimitError,
				Type:    "TokenLimitExceededError",
			},
			Message: "Token limit exceeded",
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("token limit exceeded: %s", tokenInfo.LimitError)
	}

	logger.Info("Token check passed: %d / %d (%.1f%%)",
		tokenInfo.TokenCount, tokenInfo.InputLimit, tokenInfo.Percentage)

	// Log the successful token check
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: "CheckTokens",
		Status:    "Success",
		Inputs: map[string]interface{}{
			"prompt_length": len(stitchedPrompt),
			"model_name":    cliConfig.ModelName,
		},
		Outputs: map[string]interface{}{
			"percentage": tokenInfo.Percentage,
		},
		TokenCounts: &auditlog.TokenCountInfo{
			PromptTokens: tokenInfo.TokenCount,
			TotalTokens:  tokenInfo.TokenCount,
			Limit:        tokenInfo.InputLimit,
		},
		Message: fmt.Sprintf("Token check passed: %d / %d (%.1f%%)",
			tokenInfo.TokenCount, tokenInfo.InputLimit, tokenInfo.Percentage),
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

	// 11. Generate content
	logger.Info("Generating plan...")

	// Log the start of content generation
	generateStartTime := time.Now()
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: generateStartTime,
		Operation: "GenerateContentStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"model_name":    cliConfig.ModelName,
			"prompt_length": len(stitchedPrompt),
		},
		Message: "Starting content generation with Gemini",
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

	result, err := geminiClient.GenerateContent(ctx, stitchedPrompt)

	// Calculate duration in milliseconds
	generateDurationMs := time.Since(generateStartTime).Milliseconds()

	if err != nil {
		logger.Error("Generation failed")

		// Determine error type for better categorization
		errorType := "ContentGenerationError"
		errorMessage := fmt.Sprintf("Failed to generate content: %v", err)

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Error generating content: %s", apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			logger.Debug("Error details: %s", apiErr.DebugInfo())
			errorType = "APIError"
			errorMessage = apiErr.Message
		} else {
			logger.Error("Error generating content: %v", err)
		}

		// Log the content generation failure
		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp:  time.Now().UTC(),
			Operation:  "GenerateContentEnd",
			Status:     "Failure",
			DurationMs: &generateDurationMs,
			Inputs: map[string]interface{}{
				"model_name":    cliConfig.ModelName,
				"prompt_length": len(stitchedPrompt),
			},
			Error: &auditlog.ErrorInfo{
				Message: errorMessage,
				Type:    errorType,
			},
			Message: "Content generation failed",
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("plan generation failed: %w", err)
	}

	// Log successful content generation
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp:  time.Now().UTC(),
		Operation:  "GenerateContentEnd",
		Status:     "Success",
		DurationMs: &generateDurationMs,
		Inputs: map[string]interface{}{
			"model_name":    cliConfig.ModelName,
			"prompt_length": len(stitchedPrompt),
		},
		Outputs: map[string]interface{}{
			"finish_reason":      result.FinishReason,
			"has_safety_ratings": len(result.SafetyRatings) > 0,
		},
		TokenCounts: &auditlog.TokenCountInfo{
			PromptTokens: int32(tokenInfo.TokenCount),
			OutputTokens: int32(result.TokenCount),
			TotalTokens:  int32(tokenInfo.TokenCount + result.TokenCount),
		},
		Message: "Content generation completed successfully",
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
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

	// 13 & 14. Use the helper function to save the plan to file
	err = savePlanToFile(logger, auditLogger, cliConfig.OutputFile, generatedPlan)
	if err != nil {
		return err // Error already logged by savePlanToFile
	}

	return nil
}

// RunInternal is the same as Execute but exported specifically for testing purposes.
// This allows integration tests to use this function directly and inject mocked services.
func RunInternal(
	ctx context.Context,
	cliConfig *CliConfig,
	logger logutil.LoggerInterface,
	apiService APIService,
	auditLogger auditlog.AuditLogger,
) (err error) {
	// Use a deferred function to ensure ExecuteEnd is always logged
	defer func() {
		status := "Success"
		var errorInfo *auditlog.ErrorInfo
		if err != nil {
			status = "Failure"
			errorInfo = &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "ExecutionError",
			}
		}

		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "ExecuteEnd",
			Status:    status,
			Error:     errorInfo,
			Message:   fmt.Sprintf("Execution completed with status: %s", status),
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}
	}()
	// Log the start of the RunInternal operation
	inputs := map[string]interface{}{
		"instructions_file": cliConfig.InstructionsFile,
		"output_file":       cliConfig.OutputFile,
		"audit_log_file":    cliConfig.AuditLogFile,
		"format":            cliConfig.Format,
		"paths_count":       len(cliConfig.Paths),
		"include":           cliConfig.Include,
		"exclude":           cliConfig.Exclude,
		"exclude_names":     cliConfig.ExcludeNames,
		"dry_run":           cliConfig.DryRun,
		"verbose":           cliConfig.Verbose,
		"model_name":        cliConfig.ModelName,
		"confirm_tokens":    cliConfig.ConfirmTokens,
		"log_level":         cliConfig.LogLevel,
	}

	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: "ExecuteStart",
		Status:    "InProgress",
		Inputs:    inputs,
		Message:   "Starting execution of architect tool (RunInternal)",
	}); logErr != nil {

		logger.Error("Failed to write audit log: %v", logErr)

	}

	// 1. Read instructions from file
	instructionsContent, err := os.ReadFile(cliConfig.InstructionsFile)
	if err != nil {
		logger.Error("Failed to read instructions file %s: %v", cliConfig.InstructionsFile, err)

		// Log the failure to read the instructions file to the audit log
		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "ReadInstructions",
			Status:    "Failure",
			Inputs: map[string]interface{}{
				"path": cliConfig.InstructionsFile,
			},
			Error: &auditlog.ErrorInfo{
				Message: fmt.Sprintf("Failed to read instructions file: %v", err),
				Type:    "FileIOError",
			},
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("failed to read instructions file %s: %w", cliConfig.InstructionsFile, err)
	}
	instructions := string(instructionsContent)
	logger.Info("Successfully read instructions from %s", cliConfig.InstructionsFile)

	// Log the successful reading of the instructions file to the audit log
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: "ReadInstructions",
		Status:    "Success",
		Inputs: map[string]interface{}{
			"path": cliConfig.InstructionsFile,
		},
		Outputs: map[string]interface{}{
			"content_length": len(instructions),
		},
		Message: "Successfully read instructions file",
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

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
	// Log the start of context gathering
	gatherStartTime := time.Now()
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: gatherStartTime,
		Operation: "GatherContextStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"paths":         cliConfig.Paths,
			"include":       cliConfig.Include,
			"exclude":       cliConfig.Exclude,
			"exclude_names": cliConfig.ExcludeNames,
			"format":        cliConfig.Format,
		},
		Message: "Starting to gather project context files",
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

	contextFiles, contextStats, err := contextGatherer.GatherContext(ctx, geminiClient, gatherConfig)

	// Calculate duration in milliseconds
	gatherDurationMs := time.Since(gatherStartTime).Milliseconds()

	if err != nil {
		// Log the failure of context gathering
		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp:  time.Now().UTC(),
			Operation:  "GatherContextEnd",
			Status:     "Failure",
			DurationMs: &gatherDurationMs,
			Inputs: map[string]interface{}{
				"paths":         cliConfig.Paths,
				"include":       cliConfig.Include,
				"exclude":       cliConfig.Exclude,
				"exclude_names": cliConfig.ExcludeNames,
			},
			Error: &auditlog.ErrorInfo{
				Message: fmt.Sprintf("Failed to gather project context: %v", err),
				Type:    "ContextGatheringError",
			},
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}
		return fmt.Errorf("failed during project context gathering: %w", err)
	}

	// Log the successful completion of context gathering
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp:  time.Now().UTC(),
		Operation:  "GatherContextEnd",
		Status:     "Success",
		DurationMs: &gatherDurationMs,
		Inputs: map[string]interface{}{
			"paths":         cliConfig.Paths,
			"include":       cliConfig.Include,
			"exclude":       cliConfig.Exclude,
			"exclude_names": cliConfig.ExcludeNames,
		},
		Outputs: map[string]interface{}{
			"processed_files_count": contextStats.ProcessedFilesCount,
			"char_count":            contextStats.CharCount,
			"line_count":            contextStats.LineCount,
			"token_count":           contextStats.TokenCount,
			"files_count":           len(contextFiles),
		},
		Message: "Successfully gathered project context files",
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
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

	// Log token check start with prompt information
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: "CheckTokensStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"prompt_length": len(stitchedPrompt),
			"model_name":    cliConfig.ModelName,
		},
		Message: "Starting token count check",
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

	tokenInfo, err := tokenManager.GetTokenInfo(ctx, geminiClient, stitchedPrompt)

	if err != nil {
		logger.Error("Token count check failed")

		// Determine error type for better categorization
		errorType := "TokenCheckError"
		errorMessage := fmt.Sprintf("Failed to check token count: %v", err)

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Token count check failed: %s", apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			logger.Debug("Error details: %s", apiErr.DebugInfo())
			errorType = "APIError"
			errorMessage = apiErr.Message
		} else {
			logger.Error("Token count check failed: %v", err)
			logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		}

		// Log the token check failure
		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "CheckTokens",
			Status:    "Failure",
			Inputs: map[string]interface{}{
				"prompt_length": len(stitchedPrompt),
				"model_name":    cliConfig.ModelName,
			},
			Error: &auditlog.ErrorInfo{
				Message: errorMessage,
				Type:    errorType,
			},
			Message: "Token count check failed",
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("token count check failed: %w", err)
	}

	// If token limit is exceeded, abort
	if tokenInfo.ExceedsLimit {
		logger.Error("Token limit exceeded")
		logger.Error("Token limit exceeded: %s", tokenInfo.LimitError)
		logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")

		// Log the token limit exceeded case
		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "CheckTokens",
			Status:    "Failure",
			Inputs: map[string]interface{}{
				"prompt_length": len(stitchedPrompt),
				"model_name":    cliConfig.ModelName,
			},
			TokenCounts: &auditlog.TokenCountInfo{
				PromptTokens: tokenInfo.TokenCount,
				TotalTokens:  tokenInfo.TokenCount,
				Limit:        tokenInfo.InputLimit,
			},
			Error: &auditlog.ErrorInfo{
				Message: tokenInfo.LimitError,
				Type:    "TokenLimitExceededError",
			},
			Message: "Token limit exceeded",
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("token limit exceeded: %s", tokenInfo.LimitError)
	}

	logger.Info("Token check passed: %d / %d (%.1f%%)",
		tokenInfo.TokenCount, tokenInfo.InputLimit, tokenInfo.Percentage)

	// Log the successful token check
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: "CheckTokens",
		Status:    "Success",
		Inputs: map[string]interface{}{
			"prompt_length": len(stitchedPrompt),
			"model_name":    cliConfig.ModelName,
		},
		Outputs: map[string]interface{}{
			"percentage": tokenInfo.Percentage,
		},
		TokenCounts: &auditlog.TokenCountInfo{
			PromptTokens: tokenInfo.TokenCount,
			TotalTokens:  tokenInfo.TokenCount,
			Limit:        tokenInfo.InputLimit,
		},
		Message: fmt.Sprintf("Token check passed: %d / %d (%.1f%%)",
			tokenInfo.TokenCount, tokenInfo.InputLimit, tokenInfo.Percentage),
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

	// 11. Generate content
	logger.Info("Generating plan...")

	// Log the start of content generation
	generateStartTime := time.Now()
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: generateStartTime,
		Operation: "GenerateContentStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"model_name":    cliConfig.ModelName,
			"prompt_length": len(stitchedPrompt),
		},
		Message: "Starting content generation with Gemini",
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

	result, err := geminiClient.GenerateContent(ctx, stitchedPrompt)

	// Calculate duration in milliseconds
	generateDurationMs := time.Since(generateStartTime).Milliseconds()

	if err != nil {
		logger.Error("Generation failed")

		// Determine error type for better categorization
		errorType := "ContentGenerationError"
		errorMessage := fmt.Sprintf("Failed to generate content: %v", err)

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Error generating content: %s", apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			logger.Debug("Error details: %s", apiErr.DebugInfo())
			errorType = "APIError"
			errorMessage = apiErr.Message
		} else {
			logger.Error("Error generating content: %v", err)
		}

		// Log the content generation failure
		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp:  time.Now().UTC(),
			Operation:  "GenerateContentEnd",
			Status:     "Failure",
			DurationMs: &generateDurationMs,
			Inputs: map[string]interface{}{
				"model_name":    cliConfig.ModelName,
				"prompt_length": len(stitchedPrompt),
			},
			Error: &auditlog.ErrorInfo{
				Message: errorMessage,
				Type:    errorType,
			},
			Message: "Content generation failed",
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("plan generation failed: %w", err)
	}

	// Log successful content generation
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp:  time.Now().UTC(),
		Operation:  "GenerateContentEnd",
		Status:     "Success",
		DurationMs: &generateDurationMs,
		Inputs: map[string]interface{}{
			"model_name":    cliConfig.ModelName,
			"prompt_length": len(stitchedPrompt),
		},
		Outputs: map[string]interface{}{
			"finish_reason":      result.FinishReason,
			"has_safety_ratings": len(result.SafetyRatings) > 0,
		},
		TokenCounts: &auditlog.TokenCountInfo{
			PromptTokens: int32(tokenInfo.TokenCount),
			OutputTokens: int32(result.TokenCount),
			TotalTokens:  int32(tokenInfo.TokenCount + result.TokenCount),
		},
		Message: "Content generation completed successfully",
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
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

	// 13 & 14. Use the helper function to save the plan to file
	err = savePlanToFile(logger, auditLogger, cliConfig.OutputFile, generatedPlan)
	if err != nil {
		return err // Error already logged by savePlanToFile
	}

	return nil
}

// savePlanToFile is a helper function that saves the generated plan to a file
// and includes audit logging around the file writing operation.
func savePlanToFile(
	logger logutil.LoggerInterface,
	auditLogger auditlog.AuditLogger,
	outputFilePath string,
	content string,
) error {
	// Create file writer
	fileWriter := NewFileWriter(logger)

	// Log the start of output saving
	saveStartTime := time.Now()
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: saveStartTime,
		Operation: "SaveOutputStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"output_path":    outputFilePath,
			"content_length": len(content),
		},
		Message: "Starting to save output to file",
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

	// Save output file
	logger.Info("Writing plan to %s...", outputFilePath)
	err := fileWriter.SaveToFile(content, outputFilePath)

	// Calculate duration in milliseconds
	saveDurationMs := time.Since(saveStartTime).Milliseconds()

	if err != nil {
		// Log failure to save output
		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp:  time.Now().UTC(),
			Operation:  "SaveOutputEnd",
			Status:     "Failure",
			DurationMs: &saveDurationMs,
			Inputs: map[string]interface{}{
				"output_path": outputFilePath,
			},
			Error: &auditlog.ErrorInfo{
				Message: fmt.Sprintf("Failed to save output to file: %v", err),
				Type:    "FileIOError",
			},
			Message: "Failed to save output to file",
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("error saving plan to file: %w", err)
	}

	// Log successful saving of output
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp:  time.Now().UTC(),
		Operation:  "SaveOutputEnd",
		Status:     "Success",
		DurationMs: &saveDurationMs,
		Inputs: map[string]interface{}{
			"output_path": outputFilePath,
		},
		Outputs: map[string]interface{}{
			"content_length": len(content),
		},
		Message: "Successfully saved output to file",
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

	logger.Info("Plan successfully generated and saved to %s", outputFilePath)
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

	// Validate model names
	if len(cliConfig.ModelNames) == 0 {
		return fmt.Errorf("at least one model must be specified")
	}

	return nil
}
