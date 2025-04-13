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

	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/architect/prompt"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/ratelimit"
	"github.com/phrazzld/architect/internal/runutil"
)

// Execute is the main entry point for the core application logic.
// It configures the necessary components and calls the internal run function.
func Execute(
	ctx context.Context,
	cliConfig *config.CliConfig,
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

	// Set up the output directory
	if err := setupOutputDirectory(cliConfig, logger); err != nil {
		return err
	}

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
	stitchedPrompt := prompt.StitchPrompt(instructions, contextFiles)
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
			logger.Debug("Attempting to acquire rate limiter for model %s...", modelName)
			acquireStart := time.Now()
			if err := rateLimiter.Acquire(ctx, modelName); err != nil {
				logger.Error("Rate limiting error for model %s: %v", modelName, err)
				errChan <- fmt.Errorf("model %s rate limit: %w", modelName, err)
				return
			}
			acquireDuration := time.Since(acquireStart)
			logger.Debug("Rate limiter acquired for model %s (waited %v)", modelName, acquireDuration)

			// Release rate limiter when done
			defer func() {
				logger.Debug("Releasing rate limiter for model %s", modelName)
				rateLimiter.Release()
			}()

			// Create model processor
			fileWriter := NewFileWriter(logger)

			// Create adapters for interfaces
			apiServiceAdapter := &apiServiceAdapter{apiService}
			tokenManagerAdapter := &tokenManagerAdapter{tokenManager}

			processor := modelproc.NewProcessor(
				apiServiceAdapter,
				tokenManagerAdapter,
				fileWriter,
				auditLogger,
				logger,
				cliConfig,
			)

			// Process the model
			err := processor.Process(ctx, modelName, stitchedPrompt)

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

// RunInternal is the same as Execute but exported specifically for testing purposes.
// This allows integration tests to use this function directly and inject mocked services.
func RunInternal(
	ctx context.Context,
	cliConfig *config.CliConfig,
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

	// Set up the output directory
	if err := setupOutputDirectory(cliConfig, logger); err != nil {
		return err
	}

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
	stitchedPrompt := prompt.StitchPrompt(instructions, contextFiles)
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
			logger.Debug("Attempting to acquire rate limiter for model %s...", modelName)
			acquireStart := time.Now()
			if err := rateLimiter.Acquire(ctx, modelName); err != nil {
				logger.Error("Rate limiting error for model %s: %v", modelName, err)
				errChan <- fmt.Errorf("model %s rate limit: %w", modelName, err)
				return
			}
			acquireDuration := time.Since(acquireStart)
			logger.Debug("Rate limiter acquired for model %s (waited %v)", modelName, acquireDuration)

			// Release rate limiter when done
			defer func() {
				logger.Debug("Releasing rate limiter for model %s", modelName)
				rateLimiter.Release()
			}()

			// Create model processor
			fileWriter := NewFileWriter(logger)

			// Create adapters for interfaces
			apiServiceAdapter := &apiServiceAdapter{apiService}
			tokenManagerAdapter := &tokenManagerAdapter{tokenManager}

			processor := modelproc.NewProcessor(
				apiServiceAdapter,
				tokenManagerAdapter,
				fileWriter,
				auditLogger,
				logger,
				cliConfig,
			)

			// Process the model
			err := processor.Process(ctx, modelName, stitchedPrompt)

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

// Note: processModel, processModelConcurrently, sanitizeFilename, and saveOutputToFile functions
// have been removed as part of the refactoring. Their functionality has been moved to the
// ModelProcessor in the modelproc package.

// setupOutputDirectory ensures that the output directory is set and exists.
// If outputDir in cliConfig is empty, it generates a unique directory name.
func setupOutputDirectory(cliConfig *config.CliConfig, logger logutil.LoggerInterface) error {
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
	return nil
}

// apiServiceAdapter adapts APIService to modelproc.APIService
type apiServiceAdapter struct {
	apiService APIService
}

func (a *apiServiceAdapter) InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
	return a.apiService.InitClient(ctx, apiKey, modelName)
}

func (a *apiServiceAdapter) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	return a.apiService.ProcessResponse(result)
}

func (a *apiServiceAdapter) IsEmptyResponseError(err error) bool {
	return a.apiService.IsEmptyResponseError(err)
}

func (a *apiServiceAdapter) IsSafetyBlockedError(err error) bool {
	return a.apiService.IsSafetyBlockedError(err)
}

func (a *apiServiceAdapter) GetErrorDetails(err error) string {
	return a.apiService.GetErrorDetails(err)
}

// tokenResultAdapter adapts TokenResult to modelproc.TokenResult
func tokenResultAdapter(tr *TokenResult) *modelproc.TokenResult {
	return &modelproc.TokenResult{
		TokenCount:   tr.TokenCount,
		InputLimit:   tr.InputLimit,
		ExceedsLimit: tr.ExceedsLimit,
		LimitError:   tr.LimitError,
		Percentage:   tr.Percentage,
	}
}

// tokenManagerAdapter adapts TokenManager to modelproc.TokenManager
type tokenManagerAdapter struct {
	tokenManager TokenManager
}

func (t *tokenManagerAdapter) CheckTokenLimit(ctx context.Context, client gemini.Client, prompt string) error {
	return t.tokenManager.CheckTokenLimit(ctx, client, prompt)
}

func (t *tokenManagerAdapter) GetTokenInfo(ctx context.Context, client gemini.Client, prompt string) (*modelproc.TokenResult, error) {
	result, err := t.tokenManager.GetTokenInfo(ctx, client, prompt)
	if err != nil {
		return nil, err
	}
	return tokenResultAdapter(result), nil
}

func (t *tokenManagerAdapter) PromptForConfirmation(tokenCount int32, threshold int) bool {
	return t.tokenManager.PromptForConfirmation(tokenCount, threshold)
}
