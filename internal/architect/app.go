// Package architect contains the core application logic for the architect tool
package architect

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/ratelimit"
	"github.com/phrazzld/architect/internal/runutil"
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

	// Determine the output directory
	if cliConfig.OutputDir == "" {
		// Generate a unique run name (e.g., "curious-panther")
		runName := runutil.GenerateRunName()
		// Get the current working directory
		cwd, err := os.Getwd()
		if err != nil {
			logger.Error("Error getting current working directory: %v", err)
			return fmt.Errorf("error getting current working directory: %w", err)
		}
		// Set the output directory to the run name in the current working directory
		cliConfig.OutputDir = filepath.Join(cwd, runName)
		logger.Info("Generated output directory: %s", cliConfig.OutputDir)
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(cliConfig.OutputDir, 0755); err != nil {
		logger.Error("Error creating output directory %s: %v", cliConfig.OutputDir, err)
		return fmt.Errorf("error creating output directory %s: %w", cliConfig.OutputDir, err)
	}

	logger.Info("Using output directory: %s", cliConfig.OutputDir)

	// Log the start of the Execute operation
	inputs := map[string]interface{}{
		"instructions_file": cliConfig.InstructionsFile,
		"output_dir":        cliConfig.OutputDir,
		"audit_log_file":    cliConfig.AuditLogFile,
		"format":            cliConfig.Format,
		"paths_count":       len(cliConfig.Paths),
		"include":           cliConfig.Include,
		"exclude":           cliConfig.Exclude,
		"exclude_names":     cliConfig.ExcludeNames,
		"dry_run":           cliConfig.DryRun,
		"verbose":           cliConfig.Verbose,
		"model_names":       cliConfig.ModelNames,
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

	// Input validation is handled by cmd/architect/cli.go:ValidateInputs before execution

	// 3. Create shared components
	apiService := NewAPIService(logger)
	tokenManager := NewTokenManager(logger)
	contextGatherer := NewContextGatherer(logger, cliConfig.DryRun, tokenManager)

	// 4. Create gather config
	gatherConfig := GatherConfig{
		Paths:        cliConfig.Paths,
		Include:      cliConfig.Include,
		Exclude:      cliConfig.Exclude,
		ExcludeNames: cliConfig.ExcludeNames,
		Format:       cliConfig.Format,
		Verbose:      cliConfig.Verbose,
		LogLevel:     cliConfig.LogLevel,
	}

	// 5. Initialize the reference client for context gathering only
	// We'll create model-specific clients inside the loop later
	referenceClient, err := apiService.InitClient(ctx, cliConfig.ApiKey, cliConfig.ModelNames[0])
	if err != nil {
		errorDetails := apiService.GetErrorDetails(err)
		logger.Error("Error creating reference Gemini client: %s", errorDetails)
		return fmt.Errorf("failed to initialize reference API client: %w", err)
	}
	defer referenceClient.Close()

	// 6. Gather context files (model-independent step)
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

	contextFiles, contextStats, err := contextGatherer.GatherContext(ctx, referenceClient, gatherConfig)

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

	// 7. Handle dry run mode
	if cliConfig.DryRun {
		err = contextGatherer.DisplayDryRunInfo(ctx, referenceClient, contextStats)
		if err != nil {
			logger.Error("Error displaying dry run information: %v", err)
			return fmt.Errorf("error displaying dry run information: %w", err)
		}
		return nil
	}

	// 8. Stitch prompt (model-independent step)
	stitchedPrompt := StitchPrompt(instructions, contextFiles)
	logger.Info("Prompt constructed successfully")
	logger.Debug("Stitched prompt length: %d characters", len(stitchedPrompt))

	// 9. Process each model concurrently (with rate limiting)
	var wg sync.WaitGroup
	// Create a buffered error channel to collect errors from goroutines
	errChan := make(chan error, len(cliConfig.ModelNames))

	// Create rate limiter from configuration
	rateLimiter := ratelimit.NewRateLimiter(
		cliConfig.MaxConcurrentRequests,
		cliConfig.RateLimitRequestsPerMinute,
	)

	// Log rate limiting configuration
	if cliConfig.MaxConcurrentRequests > 0 {
		logger.Info("Concurrency limited to %d simultaneous requests", cliConfig.MaxConcurrentRequests)
	} else {
		logger.Info("No concurrency limit applied")
	}

	if cliConfig.RateLimitRequestsPerMinute > 0 {
		logger.Info("Rate limited to %d requests per minute per model", cliConfig.RateLimitRequestsPerMinute)
	} else {
		logger.Info("No rate limit applied")
	}

	logger.Info("Processing %d models concurrently...", len(cliConfig.ModelNames))

	// Launch a goroutine for each model
	for _, name := range cliConfig.ModelNames {
		// Capture the loop variable to avoid data race
		modelName := name

		// Add to wait group before launching goroutine
		wg.Add(1)

		// Launch goroutine to process this model
		go func() {
			// Ensure we signal completion when goroutine exits
			defer wg.Done()

			// Acquire rate limiting permission with context
			if err := rateLimiter.Acquire(ctx, modelName); err != nil {
				logger.Error("Rate limiting error for model %s: %v", modelName, err)
				errChan <- fmt.Errorf("model %s rate limit: %w", modelName, err)
				return
			}
			// Release rate limiter when done
			defer rateLimiter.Release()

			// Process the model
			err := processModelConcurrently(ctx, modelName, cliConfig, logger, apiService, auditLogger, tokenManager, stitchedPrompt)

			// If there was an error, send it to the error channel
			if err != nil {
				logger.Error("Processing model %s failed: %v", modelName, err)
				errChan <- fmt.Errorf("model %s: %w", modelName, err)
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Close the error channel
	close(errChan)

	// Collect any errors from the channel
	var modelErrors []error
	var rateLimitErrors []error

	for err := range errChan {
		modelErrors = append(modelErrors, err)

		// Check if it's specifically a rate limit error
		if strings.Contains(err.Error(), "rate limit") {
			rateLimitErrors = append(rateLimitErrors, err)
		}
	}

	// If there were any errors, return a combined error
	if len(modelErrors) > 0 {
		errMsg := "errors occurred during model processing:"
		for _, e := range modelErrors {
			errMsg += "\n  - " + e.Error()
		}

		// Add additional guidance if there were rate limit errors
		if len(rateLimitErrors) > 0 {
			errMsg += "\n\nTip: If you're encountering rate limit errors, consider adjusting the --max-concurrent and --rate-limit flags to prevent overwhelming the API."
		}

		return errors.New(errMsg)
	}

	return nil
}

// processModel handles the processing of a single model, from client initialization to output saving
func processModel(
	ctx context.Context,
	modelName string,
	cliConfig *CliConfig,
	logger logutil.LoggerInterface,
	apiService APIService,
	auditLogger auditlog.AuditLogger,
	tokenManager TokenManager,
	stitchedPrompt string,
) error {
	logger.Info("Processing model: %s", modelName)

	// 1. Initialize model-specific client
	geminiClient, err := apiService.InitClient(ctx, cliConfig.ApiKey, modelName)
	if err != nil {
		errorDetails := apiService.GetErrorDetails(err)
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Error creating Gemini client for model %s: %s", modelName, apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			if cliConfig.LogLevel == logutil.DebugLevel {
				logger.Debug("Error details: %s", apiErr.DebugInfo())
			}
		} else {
			logger.Error("Error creating Gemini client for model %s: %s", modelName, errorDetails)
		}
		return fmt.Errorf("failed to initialize API client for model %s: %w", modelName, err)
	}
	defer geminiClient.Close()

	// 2. Check token limits for this model
	logger.Info("Checking token limits for model %s...", modelName)

	// Log token check start with prompt information
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: "CheckTokensStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"prompt_length": len(stitchedPrompt),
			"model_name":    modelName,
		},
		Message: "Starting token count check for model " + modelName,
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

	tokenInfo, err := tokenManager.GetTokenInfo(ctx, geminiClient, stitchedPrompt)
	if err != nil {
		logger.Error("Token count check failed for model %s", modelName)

		// Determine error type for better categorization
		errorType := "TokenCheckError"
		errorMessage := fmt.Sprintf("Failed to check token count for model %s: %v", modelName, err)

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Token count check failed for model %s: %s", modelName, apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			logger.Debug("Error details: %s", apiErr.DebugInfo())
			errorType = "APIError"
			errorMessage = apiErr.Message
		} else {
			logger.Error("Token count check failed for model %s: %v", modelName, err)
			logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")
		}

		// Log the token check failure
		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "CheckTokens",
			Status:    "Failure",
			Inputs: map[string]interface{}{
				"prompt_length": len(stitchedPrompt),
				"model_name":    modelName,
			},
			Error: &auditlog.ErrorInfo{
				Message: errorMessage,
				Type:    errorType,
			},
			Message: "Token count check failed for model " + modelName,
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("token count check failed for model %s: %w", modelName, err)
	}

	// If token limit is exceeded, abort
	if tokenInfo.ExceedsLimit {
		logger.Error("Token limit exceeded for model %s", modelName)
		logger.Error("Token limit exceeded: %s", tokenInfo.LimitError)
		logger.Error("Try reducing context by using --include, --exclude, or --exclude-names flags")

		// Log the token limit exceeded case
		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp: time.Now().UTC(),
			Operation: "CheckTokens",
			Status:    "Failure",
			Inputs: map[string]interface{}{
				"prompt_length": len(stitchedPrompt),
				"model_name":    modelName,
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
			Message: "Token limit exceeded for model " + modelName,
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("token limit exceeded for model %s: %s", modelName, tokenInfo.LimitError)
	}

	logger.Info("Token check passed for model %s: %d / %d (%.1f%%)",
		modelName, tokenInfo.TokenCount, tokenInfo.InputLimit, tokenInfo.Percentage)

	// Log the successful token check
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now().UTC(),
		Operation: "CheckTokens",
		Status:    "Success",
		Inputs: map[string]interface{}{
			"prompt_length": len(stitchedPrompt),
			"model_name":    modelName,
		},
		Outputs: map[string]interface{}{
			"percentage": tokenInfo.Percentage,
		},
		TokenCounts: &auditlog.TokenCountInfo{
			PromptTokens: tokenInfo.TokenCount,
			TotalTokens:  tokenInfo.TokenCount,
			Limit:        tokenInfo.InputLimit,
		},
		Message: fmt.Sprintf("Token check passed for model %s: %d / %d (%.1f%%)",
			modelName, tokenInfo.TokenCount, tokenInfo.InputLimit, tokenInfo.Percentage),
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

	// 3. Generate content with this model
	logger.Info("Generating output with model %s...", modelName)

	// Log the start of content generation
	generateStartTime := time.Now()
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp: generateStartTime,
		Operation: "GenerateContentStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"model_name":    modelName,
			"prompt_length": len(stitchedPrompt),
		},
		Message: "Starting content generation with Gemini model " + modelName,
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

	result, err := geminiClient.GenerateContent(ctx, stitchedPrompt)

	// Calculate duration in milliseconds
	generateDurationMs := time.Since(generateStartTime).Milliseconds()

	if err != nil {
		logger.Error("Generation failed for model %s", modelName)

		// Determine error type for better categorization
		errorType := "ContentGenerationError"
		errorMessage := fmt.Sprintf("Failed to generate content with model %s: %v", modelName, err)

		// Check if it's an API error with enhanced details
		if apiErr, ok := gemini.IsAPIError(err); ok {
			logger.Error("Error generating content with model %s: %s", modelName, apiErr.Message)
			if apiErr.Suggestion != "" {
				logger.Error("Suggestion: %s", apiErr.Suggestion)
			}
			logger.Debug("Error details: %s", apiErr.DebugInfo())
			errorType = "APIError"
			errorMessage = apiErr.Message
		} else {
			logger.Error("Error generating content with model %s: %v", modelName, err)
		}

		// Log the content generation failure
		if logErr := auditLogger.Log(auditlog.AuditEntry{
			Timestamp:  time.Now().UTC(),
			Operation:  "GenerateContentEnd",
			Status:     "Failure",
			DurationMs: &generateDurationMs,
			Inputs: map[string]interface{}{
				"model_name":    modelName,
				"prompt_length": len(stitchedPrompt),
			},
			Error: &auditlog.ErrorInfo{
				Message: errorMessage,
				Type:    errorType,
			},
			Message: "Content generation failed for model " + modelName,
		}); logErr != nil {
			logger.Error("Failed to write audit log: %v", logErr)
		}

		return fmt.Errorf("output generation failed for model %s: %w", modelName, err)
	}

	// Log successful content generation
	if logErr := auditLogger.Log(auditlog.AuditEntry{
		Timestamp:  time.Now().UTC(),
		Operation:  "GenerateContentEnd",
		Status:     "Success",
		DurationMs: &generateDurationMs,
		Inputs: map[string]interface{}{
			"model_name":    modelName,
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
		Message: "Content generation completed successfully for model " + modelName,
	}); logErr != nil {
		logger.Error("Failed to write audit log: %v", logErr)
	}

	// 4. Process API response
	generatedOutput, err := apiService.ProcessResponse(result)
	if err != nil {
		// Get detailed error information
		errorDetails := apiService.GetErrorDetails(err)

		// Provide specific error messages based on error type
		if apiService.IsEmptyResponseError(err) {
			logger.Error("Received empty or invalid response from Gemini API for model %s", modelName)
			logger.Error("Error details: %s", errorDetails)
			return fmt.Errorf("failed to process API response for model %s due to empty content: %w", modelName, err)
		} else if apiService.IsSafetyBlockedError(err) {
			logger.Error("Content was blocked by Gemini safety filters for model %s", modelName)
			logger.Error("Error details: %s", errorDetails)
			return fmt.Errorf("failed to process API response for model %s due to safety restrictions: %w", modelName, err)
		} else {
			// Generic API error handling
			return fmt.Errorf("failed to process API response for model %s: %w", modelName, err)
		}
	}
	logger.Info("Output generated successfully with model %s", modelName)

	// 5. Sanitize model name for use in filename
	sanitizedModelName := sanitizeFilename(modelName)

	// 6. Construct output file path
	outputFilePath := filepath.Join(cliConfig.OutputDir, sanitizedModelName+".md")

	// 7. Save the output to file
	err = saveOutputToFile(logger, auditLogger, outputFilePath, generatedOutput)
	if err != nil {
		return fmt.Errorf("failed to save output for model %s: %w", modelName, err)
	}

	logger.Info("Successfully processed model: %s", modelName)
	return nil
}

// sanitizeFilename replaces characters that are not valid in filenames
func sanitizeFilename(filename string) string {
	// Replace slashes and other problematic characters with hyphens
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
	)
	return replacer.Replace(filename)
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

	// Determine the output directory
	if cliConfig.OutputDir == "" {
		// Generate a unique run name (e.g., "curious-panther")
		runName := runutil.GenerateRunName()
		// Get the current working directory
		cwd, err := os.Getwd()
		if err != nil {
			logger.Error("Error getting current working directory: %v", err)
			return fmt.Errorf("error getting current working directory: %w", err)
		}
		// Set the output directory to the run name in the current working directory
		cliConfig.OutputDir = filepath.Join(cwd, runName)
		logger.Info("Generated output directory: %s", cliConfig.OutputDir)
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(cliConfig.OutputDir, 0755); err != nil {
		logger.Error("Error creating output directory %s: %v", cliConfig.OutputDir, err)
		return fmt.Errorf("error creating output directory %s: %w", cliConfig.OutputDir, err)
	}

	logger.Info("Using output directory: %s", cliConfig.OutputDir)

	// Log the start of the RunInternal operation
	inputs := map[string]interface{}{
		"instructions_file": cliConfig.InstructionsFile,
		"output_dir":        cliConfig.OutputDir,
		"audit_log_file":    cliConfig.AuditLogFile,
		"format":            cliConfig.Format,
		"paths_count":       len(cliConfig.Paths),
		"include":           cliConfig.Include,
		"exclude":           cliConfig.Exclude,
		"exclude_names":     cliConfig.ExcludeNames,
		"dry_run":           cliConfig.DryRun,
		"verbose":           cliConfig.Verbose,
		"model_names":       cliConfig.ModelNames,
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

	// Input validation is handled by cmd/architect/cli.go:ValidateInputs before execution

	// 3. Create token manager
	tokenManager := NewTokenManager(logger)

	// 4. Create context gatherer
	contextGatherer := NewContextGatherer(logger, cliConfig.DryRun, tokenManager)

	// 5. Create gather config
	gatherConfig := GatherConfig{
		Paths:        cliConfig.Paths,
		Include:      cliConfig.Include,
		Exclude:      cliConfig.Exclude,
		ExcludeNames: cliConfig.ExcludeNames,
		Format:       cliConfig.Format,
		Verbose:      cliConfig.Verbose,
		LogLevel:     cliConfig.LogLevel,
	}

	// 6. Initialize the reference client for context gathering only
	// We'll create model-specific clients inside the loop later
	referenceClient, err := apiService.InitClient(ctx, cliConfig.ApiKey, cliConfig.ModelNames[0])
	if err != nil {
		errorDetails := apiService.GetErrorDetails(err)
		logger.Error("Error creating reference Gemini client: %s", errorDetails)
		return fmt.Errorf("failed to initialize reference API client: %w", err)
	}
	defer referenceClient.Close()

	// 7. Gather context files (model-independent step)
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

	contextFiles, contextStats, err := contextGatherer.GatherContext(ctx, referenceClient, gatherConfig)

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
		err = contextGatherer.DisplayDryRunInfo(ctx, referenceClient, contextStats)
		if err != nil {
			logger.Error("Error displaying dry run information: %v", err)
			return fmt.Errorf("error displaying dry run information: %w", err)
		}
		return nil
	}

	// 9. Stitch prompt (model-independent step)
	stitchedPrompt := StitchPrompt(instructions, contextFiles)
	logger.Info("Prompt constructed successfully")
	logger.Debug("Stitched prompt length: %d characters", len(stitchedPrompt))

	// 10. Process each model concurrently (with rate limiting)
	var wg sync.WaitGroup
	// Create a buffered error channel to collect errors from goroutines
	errChan := make(chan error, len(cliConfig.ModelNames))

	// Create rate limiter from configuration
	rateLimiter := ratelimit.NewRateLimiter(
		cliConfig.MaxConcurrentRequests,
		cliConfig.RateLimitRequestsPerMinute,
	)

	// Log rate limiting configuration
	if cliConfig.MaxConcurrentRequests > 0 {
		logger.Info("Concurrency limited to %d simultaneous requests", cliConfig.MaxConcurrentRequests)
	} else {
		logger.Info("No concurrency limit applied")
	}

	if cliConfig.RateLimitRequestsPerMinute > 0 {
		logger.Info("Rate limited to %d requests per minute per model", cliConfig.RateLimitRequestsPerMinute)
	} else {
		logger.Info("No rate limit applied")
	}

	logger.Info("Processing %d models concurrently...", len(cliConfig.ModelNames))

	// Launch a goroutine for each model
	for _, name := range cliConfig.ModelNames {
		// Capture the loop variable to avoid data race
		modelName := name

		// Add to wait group before launching goroutine
		wg.Add(1)

		// Launch goroutine to process this model
		go func() {
			// Ensure we signal completion when goroutine exits
			defer wg.Done()

			// Acquire rate limiting permission with context
			if err := rateLimiter.Acquire(ctx, modelName); err != nil {
				logger.Error("Rate limiting error for model %s: %v", modelName, err)
				errChan <- fmt.Errorf("model %s rate limit: %w", modelName, err)
				return
			}
			// Release rate limiter when done
			defer rateLimiter.Release()

			// Process the model
			err := processModelConcurrently(ctx, modelName, cliConfig, logger, apiService, auditLogger, tokenManager, stitchedPrompt)

			// If there was an error, send it to the error channel
			if err != nil {
				logger.Error("Processing model %s failed: %v", modelName, err)
				errChan <- fmt.Errorf("model %s: %w", modelName, err)
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Close the error channel
	close(errChan)

	// Collect any errors from the channel
	var modelErrors []error
	var rateLimitErrors []error

	for err := range errChan {
		modelErrors = append(modelErrors, err)

		// Check if it's specifically a rate limit error
		if strings.Contains(err.Error(), "rate limit") {
			rateLimitErrors = append(rateLimitErrors, err)
		}
	}

	// If there were any errors, return a combined error
	if len(modelErrors) > 0 {
		errMsg := "errors occurred during model processing:"
		for _, e := range modelErrors {
			errMsg += "\n  - " + e.Error()
		}

		// Add additional guidance if there were rate limit errors
		if len(rateLimitErrors) > 0 {
			errMsg += "\n\nTip: If you're encountering rate limit errors, consider adjusting the --max-concurrent and --rate-limit flags to prevent overwhelming the API."
		}

		return errors.New(errMsg)
	}

	return nil
}

// saveOutputToFile is a helper function that saves the generated output to a model-specific file
// and includes audit logging around the file writing operation.
// outputFilePath is the model-specific output file path (e.g., outputDir/model-name.md)
func saveOutputToFile(
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
	logger.Info("Writing output to %s...", outputFilePath)
	err := fileWriter.SaveToFile(content, outputFilePath)

	// Calculate duration in milliseconds
	saveDurationMs := time.Since(saveStartTime).Milliseconds()

	if err != nil {
		// Log failure to save output
		logger.Error("Error saving output to file %s: %v", outputFilePath, err)

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

		return fmt.Errorf("error saving output to file %s: %w", outputFilePath, err)
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

	logger.Info("Output successfully generated and saved to %s", outputFilePath)
	return nil
}

// processModelConcurrently encapsulates the logic to process a single model
// and is designed to be called from a goroutine. It handles client initialization,
// token checking, content generation, and output file saving for a single model.
func processModelConcurrently(
	ctx context.Context,
	modelName string,
	cliConfig *CliConfig,
	logger logutil.LoggerInterface,
	apiService APIService,
	auditLogger auditlog.AuditLogger,
	tokenManager TokenManager,
	stitchedPrompt string,
) error {
	return processModel(ctx, modelName, cliConfig, logger, apiService, auditLogger, tokenManager, stitchedPrompt)
}

// Note: HandleSpecialCommands, processTaskInput, and validateInputs functions have been removed
// as part of the refactoring to simplify the core application flow.
// The functionality has been replaced with direct reading of the instructions file,
// the prompt stitching logic, and comprehensive validation in cmd/architect/cli.go:ValidateInputs.
