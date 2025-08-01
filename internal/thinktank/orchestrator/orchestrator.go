// Package orchestrator is responsible for coordinating the core application workflow.
// It brings together various components like context gathering, API interaction,
// token management, and output writing to execute the main task defined
// by user instructions and configuration.
package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
	"github.com/phrazzld/thinktank/internal/thinktank/prompt"
)

// Orchestrator coordinates the main application logic.
// It depends on various services to perform tasks like interacting with the API,
// gathering context, writing files, logging audits, and handling rate limits.
type Orchestrator struct {
	apiService           interfaces.APIService
	contextGatherer      interfaces.ContextGatherer
	fileWriter           interfaces.FileWriter
	auditLogger          auditlog.AuditLogger
	rateLimiter          *ratelimit.RateLimiter
	config               *config.CliConfig
	logger               logutil.LoggerInterface
	consoleWriter        logutil.ConsoleWriter
	synthesisService     SynthesisService
	outputWriter         OutputWriter
	summaryWriter        SummaryWriter
	tokenCountingService interfaces.TokenCountingService
	modelRateLimiters    map[string]*ratelimit.RateLimiter // Per-model rate limiters for models with specific concurrency limits
	rateLimiterMutex     sync.RWMutex                      // Protects modelRateLimiters map
}

// NewOrchestrator creates a new instance of the Orchestrator.
// It requires all necessary dependencies to be provided during construction,
// ensuring that the orchestrator is properly configured to execute its tasks.
func NewOrchestrator(
	apiService interfaces.APIService,
	contextGatherer interfaces.ContextGatherer,
	fileWriter interfaces.FileWriter,
	auditLogger auditlog.AuditLogger,
	rateLimiter *ratelimit.RateLimiter,
	config *config.CliConfig,
	logger logutil.LoggerInterface,
	consoleWriter logutil.ConsoleWriter,
	tokenCountingService interfaces.TokenCountingService,
) *Orchestrator {
	// Create the output writer
	outputWriter := NewOutputWriter(fileWriter, auditLogger, logger)
	// Create the summary writer
	summaryWriter := NewSummaryWriter(logger, consoleWriter)
	// Create a synthesis service only if synthesis model is specified
	var synthesisService SynthesisService
	if config.SynthesisModel != "" {
		synthesisService = NewSynthesisService(apiService, auditLogger, logger, config.SynthesisModel)
	}
	return &Orchestrator{
		apiService:           apiService,
		contextGatherer:      contextGatherer,
		fileWriter:           fileWriter,
		auditLogger:          auditLogger,
		rateLimiter:          rateLimiter,
		config:               config,
		logger:               logger,
		consoleWriter:        consoleWriter,
		synthesisService:     synthesisService,
		outputWriter:         outputWriter,
		summaryWriter:        summaryWriter,
		tokenCountingService: tokenCountingService,
		modelRateLimiters:    make(map[string]*ratelimit.RateLimiter),
	}
}

// getRateLimiterForModel returns the appropriate rate limiter for a specific model.
// If the model has MaxConcurrentRequests set, it creates/returns a model-specific rate limiter.
// Otherwise, it returns the global rate limiter.
func (o *Orchestrator) getRateLimiterForModel(modelName string) *ratelimit.RateLimiter {
	// Get model info to check for specific concurrency limits
	modelInfo, err := models.GetModelInfo(modelName)
	if err != nil {
		// If we can't get model info, use global rate limiter
		return o.rateLimiter
	}

	// If no specific concurrency limit is set, use global rate limiter
	if modelInfo.MaxConcurrentRequests == nil {
		return o.rateLimiter
	}
	// Check if we already have a rate limiter for this model
	o.rateLimiterMutex.RLock()
	if limiter, exists := o.modelRateLimiters[modelName]; exists {
		o.rateLimiterMutex.RUnlock()
		return limiter
	}
	o.rateLimiterMutex.RUnlock()
	// Create a new model-specific rate limiter
	o.rateLimiterMutex.Lock()
	defer o.rateLimiterMutex.Unlock()
	// Double-check in case another goroutine created it while we were waiting for the lock
	if limiter, exists := o.modelRateLimiters[modelName]; exists {
		return limiter
	}
	// Create new rate limiter with model-specific concurrency limit
	// Use provider-aware rate limiting for RPM
	// Determine the effective rate limit for this model
	// Priority: model-specific > provider-specific config > provider default
	var effectiveRateLimit int
	if modelInfo.RateLimitRPM != nil {
		// Model has a specific rate limit override
		effectiveRateLimit = *modelInfo.RateLimitRPM
		o.logger.DebugContext(context.Background(), "Using model-specific rate limit for %s: %d RPM", modelName, effectiveRateLimit)
	} else {
		// Check provider-specific config, then fall back to provider default
		configRateLimit := o.config.GetProviderRateLimit(modelInfo.Provider)
		if configRateLimit > 0 {
			effectiveRateLimit = configRateLimit
			o.logger.DebugContext(context.Background(), "Using config rate limit for %s (%s): %d RPM", modelName, modelInfo.Provider, effectiveRateLimit)
		} else {
			// Use provider default
			effectiveRateLimit = models.GetProviderDefaultRateLimit(modelInfo.Provider)
			o.logger.DebugContext(context.Background(), "Using provider default rate limit for %s (%s): %d RPM", modelName, modelInfo.Provider, effectiveRateLimit)
		}
	}
	modelRateLimiter := ratelimit.NewRateLimiter(
		*modelInfo.MaxConcurrentRequests,
		effectiveRateLimit,
	)
	o.modelRateLimiters[modelName] = modelRateLimiter
	return modelRateLimiter
}

// Run executes the main application workflow, representing the core business logic.
// It functions as a coordinator, delegating specific tasks to helper methods.
//
// Workflow:
// 1. Setup context with correlation ID and validate configuration
// 2. Gather context from project files
// 3. Handle dry run mode (if enabled)
// 4. Build the complete prompt
// 5. Process models concurrently with error handling
// 6. Save outputs (either individually or via synthesis)
// 7. Generate and display execution summary
// 8. Handle and report any errors
//
// Each step is delegated to a specialized helper method, making the workflow
// clear and maintainable.
func (o *Orchestrator) Run(ctx context.Context, instructions string) error {
	// Setup: initialize context and validate configuration
	ctx, contextLogger, err := o.setupContext(ctx)
	if err != nil {
		return err
	}

	// Welcome message to the user
	o.consoleWriter.StatusMessage("Starting thinktank processing...")
	// Step 1: Gather file context for the prompt
	contextFiles, contextStats, err := o.gatherProjectContext(ctx)
	if err != nil {
		contextLogger.ErrorContext(ctx, "Failed to gather project context: %v", err)
		return err
	}
	// Step 2: Handle dry run mode (short-circuit if enabled)
	if dryRunExecuted, err := o.runDryRunFlow(ctx, contextStats); err != nil {
		return err
	} else if dryRunExecuted {
		return nil
	}
	// Step 3: Build the complete prompt
	stitchedPrompt := o.buildPrompt(ctx, instructions, contextFiles)
	// Step 4: Process all models and handle errors
	modelOutputs, processingErr, criticalErr := o.processModelsWithErrorHandling(ctx, stitchedPrompt, contextLogger)
	if criticalErr != nil {
		return criticalErr
	}
	// Step 5: Save outputs (via synthesis or individually)
	outputInfo, fileSaveErr := o.handleOutputFlow(ctx, instructions, modelOutputs)
	// Step 6: Generate and display the execution summary
	summary := o.generateResultsSummary(modelOutputs, outputInfo, processingErr)
	o.summaryWriter.DisplaySummary(ctx, summary)
	// Step 7: Final error processing and return
	return o.handleProcessingOutcome(ctx, processingErr, fileSaveErr, contextLogger)
}

// gatherProjectContext collects relevant files from the project based on configuration.
func (o *Orchestrator) gatherProjectContext(ctx context.Context) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	// Notify user that context gathering is starting (skip for dry run since it has its own display)
	if !o.config.DryRun {
		o.consoleWriter.StatusMessage("Gathering project files...")
	}

	gatherConfig := interfaces.GatherConfig{
		Paths:        o.config.Paths,
		Include:      o.config.Include,
		Exclude:      o.config.Exclude,
		ExcludeNames: o.config.ExcludeNames,
		Format:       o.config.Format,
		Verbose:      o.config.Verbose,
		LogLevel:     o.config.LogLevel,
	}

	contextFiles, contextStats, err := o.contextGatherer.GatherContext(ctx, gatherConfig)
	if err != nil {
		return nil, nil, WrapOrchestratorError(
			ErrModelProcessingCancelled,
			fmt.Sprintf("failed during project context gathering: %v", err),
		)
	}

	return contextFiles, contextStats, nil
}

// runDryRunFlow handles the dry run mode by displaying statistics without performing API calls.
// It short-circuits the execution flow when in dry run mode.
// Returns:
// - bool: true if dry run was executed (to stop normal flow), false if normal processing should continue
// - error: any error that occurred during dry run handling
func (o *Orchestrator) runDryRunFlow(ctx context.Context, contextStats *interfaces.ContextStats) (bool, error) {
	// Early return if not in dry run mode
	if !o.config.DryRun {
		return false, nil
	}
	// Get logger with context
	contextLogger := o.logger.WithContext(ctx)

	// Log that we're in dry run mode
	contextLogger.InfoContext(ctx, "Running in dry-run mode")
	// Call the existing handleDryRun method
	err := o.handleDryRun(ctx, contextStats)
	if err != nil {
		return true, err
	}
	// Indicate dry run was handled successfully
	return true, nil
}

// runIndividualOutputFlow handles the saving of individual model outputs when no synthesis model is specified.
// It saves each model's output to a separate file in the output directory.
// Returns a map of model names to output file paths, and an error if any of the outputs fail to save.
func (o *Orchestrator) runIndividualOutputFlow(ctx context.Context, modelOutputs map[string]string) (map[string]string, error) {
	// Get logger with context
	contextLogger := o.logger.WithContext(ctx)
	// Log that individual outputs are being saved
	contextLogger.InfoContext(ctx, "Processing completed, saving individual model outputs")
	contextLogger.DebugContext(ctx, "Collected %d model outputs", len(modelOutputs))
	// Notify user that individual outputs are being saved
	o.consoleWriter.ShowFileOperations("Saving individual outputs...")

	// Use the OutputWriter to save individual model outputs
	savedCount, filePaths, err := o.outputWriter.SaveIndividualOutputs(ctx, modelOutputs, o.config.OutputDir)
	if err != nil {
		contextLogger.ErrorContext(ctx, "Completed with errors: %d files saved successfully, %d files failed",
			savedCount, len(modelOutputs)-savedCount)
		return filePaths, err
	}
	contextLogger.InfoContext(ctx, "All %d model outputs saved successfully", savedCount)
	// Notify user that individual outputs are complete
	o.consoleWriter.ShowFileOperations(fmt.Sprintf("● Outputs saved to: %s", o.config.OutputDir))

	return filePaths, nil
}

// runSynthesisFlow handles the synthesis of multiple model outputs using the specified synthesis model.
// It synthesizes the results and saves the output to a file.
// Returns the path to the synthesis file, and an error if synthesis fails or if the output cannot be saved.
func (o *Orchestrator) runSynthesisFlow(ctx context.Context, instructions string, modelOutputs map[string]string) (string, error) {
	// Get logger with context
	contextLogger := o.logger.WithContext(ctx)
	// Log that we're starting synthesis
	contextLogger.InfoContext(ctx, "Processing completed, synthesizing results with model: %s", o.config.SynthesisModel)
	contextLogger.DebugContext(ctx, "Synthesizing %d model outputs", len(modelOutputs))
	// Only proceed with synthesis if we have model outputs to synthesize
	if len(modelOutputs) == 0 {
		contextLogger.WarnContext(ctx, "No model outputs available for synthesis")
		return "", nil
	}
	// Report synthesis started
	o.consoleWriter.SynthesisStarted()
	// Attempt to synthesize results using the SynthesisService
	contextLogger.InfoContext(ctx, "Starting synthesis with model: %s", o.config.SynthesisModel)
	synthesisContent, err := o.synthesisService.SynthesizeResults(ctx, instructions, modelOutputs)
	if err != nil {
		// Process the error with specialized handling
		contextLogger.ErrorContext(ctx, "Synthesis failed: %v", err)
		return "", err
	}
	// Log synthesis success
	contextLogger.InfoContext(ctx, "Successfully synthesized results from %d model outputs", len(modelOutputs))
	contextLogger.DebugContext(ctx, "Synthesis output length: %d characters", len(synthesisContent))

	// Save the synthesis output using the OutputWriter
	outputPath, err := o.outputWriter.SaveSynthesisOutput(ctx, synthesisContent, o.config.SynthesisModel, o.config.OutputDir)
	if err != nil {
		contextLogger.ErrorContext(ctx, "Failed to save synthesis output: %v", err)
		return "", err
	}

	contextLogger.InfoContext(ctx, "Successfully saved synthesis output to %s", outputPath)

	// Report synthesis completed
	o.consoleWriter.SynthesisCompleted(outputPath)

	return outputPath, nil
}

// handleDryRun displays context statistics without performing API calls.
func (o *Orchestrator) handleDryRun(ctx context.Context, stats *interfaces.ContextStats) error {
	err := o.contextGatherer.DisplayDryRunInfo(ctx, stats)
	if err != nil {
		o.logger.ErrorContext(ctx, "Error displaying dry run information: %v", err)
		return WrapOrchestratorError(
			ErrModelProcessingCancelled,
			fmt.Sprintf("error displaying dry run information: %v", err),
		)
	}
	return nil
}

// buildPrompt creates the complete prompt by combining instructions with context files.
func (o *Orchestrator) buildPrompt(ctx context.Context, instructions string, contextFiles []fileutil.FileMeta) string {
	stitchedPrompt := prompt.StitchPrompt(instructions, contextFiles)
	o.logger.InfoContext(ctx, "Prompt constructed successfully")
	o.logger.DebugContext(ctx, "Stitched prompt length: %d characters", len(stitchedPrompt))
	return stitchedPrompt
}

// logRateLimitingConfiguration logs information about concurrency and rate limits.
func (o *Orchestrator) logRateLimitingConfiguration(ctx context.Context) {
	// Get logger with context
	contextLogger := o.logger.WithContext(ctx)

	if o.config.MaxConcurrentRequests > 0 {
		contextLogger.InfoContext(ctx, "Concurrency limited to %d simultaneous requests", o.config.MaxConcurrentRequests)
	} else {
		contextLogger.InfoContext(ctx, "No concurrency limit applied")
	}

	if o.config.RateLimitRequestsPerMinute > 0 {
		contextLogger.InfoContext(ctx, "Rate limited to %d requests per minute per model", o.config.RateLimitRequestsPerMinute)
	} else {
		contextLogger.InfoContext(ctx, "No rate limit applied")
	}

	contextLogger.InfoContext(ctx, "Processing %d models concurrently...", len(o.config.ModelNames))
}

// APIServiceAdapter adapts interfaces.APIService to modelproc.APIService.
// This adapter pattern resolves potential interface incompatibilities between
// packages without requiring changes to either interface. It allows the orchestrator
// to work with the interfaces.APIService while providing the modelproc package with
// a compatible interface implementation, preventing circular dependencies.
type APIServiceAdapter struct {
	APIService interfaces.APIService
}

// InitLLMClient initializes and returns a provider-agnostic LLM client.
// It delegates to the underlying APIService implementation's InitLLMClient method.
func (a *APIServiceAdapter) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	return a.APIService.InitLLMClient(ctx, apiKey, modelName, apiEndpoint)
}

// ProcessLLMResponse extracts content from the provider-agnostic API response.
// It delegates to the underlying APIService implementation's ProcessLLMResponse method.
func (a *APIServiceAdapter) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	return a.APIService.ProcessLLMResponse(result)
}

// IsEmptyResponseError checks if an error is related to empty API responses.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) IsEmptyResponseError(err error) bool {
	return a.APIService.IsEmptyResponseError(err)
}

// IsSafetyBlockedError checks if an error is related to safety filters.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) IsSafetyBlockedError(err error) bool {
	return a.APIService.IsSafetyBlockedError(err)
}

// GetErrorDetails extracts detailed information from an error.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) GetErrorDetails(err error) string {
	return a.APIService.GetErrorDetails(err)
}

// GetModelParameters retrieves parameter values from the models package for a given model.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) GetModelParameters(ctx context.Context, modelName string) (map[string]interface{}, error) {
	return a.APIService.GetModelParameters(ctx, modelName)
}

// GetModelDefinition retrieves the full model definition.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) GetModelDefinition(ctx context.Context, modelName string) (*models.ModelInfo, error) {
	return a.APIService.GetModelDefinition(ctx, modelName)
}

// GetModelTokenLimits retrieves token limits from the models package for a given model.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) GetModelTokenLimits(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error) {
	return a.APIService.GetModelTokenLimits(ctx, modelName)
}

// ValidateModelParameter validates a parameter value against its constraints.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) ValidateModelParameter(ctx context.Context, modelName, paramName string, value interface{}) (bool, error) {
	return a.APIService.ValidateModelParameter(ctx, modelName, paramName, value)
}

// aggregateErrors combines multiple errors into a single error with context.
// It takes a slice of errors, total expected count, successful count, and creates
// a descriptive error message that includes statistics and all individual error messages.
// Returns nil if the error slice is empty.
func (o *Orchestrator) aggregateErrors(errs []error, totalCount, successCount int) error {
	if len(errs) == 0 {
		return nil
	}

	// If all operations failed (no successes)
	if successCount == 0 {
		errorMsg := fmt.Sprintf("all models failed: %v", aggregateErrorMessages(errs))
		return fmt.Errorf("%w: %s", ErrAllProcessingFailed, errorMsg)
	}

	// Some operations succeeded but others failed
	return fmt.Errorf("%w: processed %d/%d models successfully; %d failed: %v",
		ErrPartialProcessingFailure, successCount, totalCount, len(errs),
		aggregateErrorMessages(errs))
}

// setupContext handles the initial setup of the context, validation, and logging.
// It ensures that the context has a correlation ID and validates that the required
// configuration values are present.
// Returns the enhanced context, a context logger, and an error if validation fails.
func (o *Orchestrator) setupContext(ctx context.Context) (context.Context, logutil.LoggerInterface, error) {
	// Ensure context has a correlation ID for tracing and structured logging
	ctx = logutil.WithCorrelationID(ctx)

	// Create a logger with the context for all subsequent logging
	contextLogger := o.logger.WithContext(ctx)

	// Validate that models are specified
	if len(o.config.ModelNames) == 0 {
		return ctx, contextLogger, fmt.Errorf("%w: at least one model is required", ErrNoValidModels)
	}

	// Log the start of processing
	contextLogger.InfoContext(ctx, "Starting processing")

	return ctx, contextLogger, nil
}

// generateResultsSummary creates a ResultsSummary object containing information about
// the processing results, including model successes, failures, and output file paths.
func (o *Orchestrator) generateResultsSummary(
	modelOutputs map[string]string,
	outputInfo *OutputInfo,
	processingErr error,
) *ResultsSummary {
	summary := &ResultsSummary{
		TotalModels:      len(o.config.ModelNames),
		SuccessfulModels: len(modelOutputs),
	}

	// Add successful model names
	for modelName := range modelOutputs {
		summary.SuccessfulNames = append(summary.SuccessfulNames, modelName)
	}

	// Add synthesis path if available
	if outputInfo.SynthesisFilePath != "" {
		summary.SynthesisPath = outputInfo.SynthesisFilePath
	}

	// Add individual output paths if available
	for _, path := range outputInfo.IndividualFilePaths {
		summary.OutputPaths = append(summary.OutputPaths, path)
	}

	// Determine failed models (those in config.ModelNames but not in modelOutputs)
	successMap := make(map[string]bool)
	for modelName := range modelOutputs {
		successMap[modelName] = true
	}

	for _, modelName := range o.config.ModelNames {
		if !successMap[modelName] {
			summary.FailedModels = append(summary.FailedModels, modelName)
		}
	}

	return summary
}

// handleOutputFlow decides whether to use synthesis or individual output flow
// based on configuration and handles the saving of outputs accordingly.
// Returns an OutputInfo struct containing the paths to generated files, and any error from the output handling.
func (o *Orchestrator) handleOutputFlow(ctx context.Context, instructions string, modelOutputs map[string]string) (*OutputInfo, error) {
	outputInfo := NewOutputInfo()

	if o.config.SynthesisModel == "" {
		// No synthesis model specified - save individual model outputs
		filePaths, err := o.runIndividualOutputFlow(ctx, modelOutputs)
		if filePaths != nil {
			outputInfo.IndividualFilePaths = filePaths
		}
		return outputInfo, err
	}

	// Synthesis model specified - save both individual outputs and synthesis
	// Get logger with context
	contextLogger := o.logger.WithContext(ctx)

	// First, save individual model outputs
	filePaths, individualErr := o.runIndividualOutputFlow(ctx, modelOutputs)
	if filePaths != nil {
		outputInfo.IndividualFilePaths = filePaths
	}

	// Then, run synthesis flow
	synthesisPath, synthesisErr := o.runSynthesisFlow(ctx, instructions, modelOutputs)

	if synthesisErr != nil {
		// If synthesis fails, log it but still return individual outputs
		contextLogger.WarnContext(ctx, "Synthesis failed, but individual outputs were saved: %v", synthesisErr)
		return outputInfo, synthesisErr
	}

	if synthesisPath != "" {
		outputInfo.SynthesisFilePath = synthesisPath
	}

	// Return individual error if synthesis succeeded but individual saving failed
	return outputInfo, individualErr
}

// handleProcessingOutcome combines and reports any errors from model processing and file saving.
// It formats and logs appropriate error messages based on the types of errors encountered.
// It also categorizes errors using the llm.ErrorCategory system for consistent handling.
func (o *Orchestrator) handleProcessingOutcome(ctx context.Context, processingErr, fileSaveErr error, contextLogger logutil.LoggerInterface) error {
	if processingErr != nil && fileSaveErr != nil {
		// Combine model and file save errors
		// Detect the categories of both errors
		procCategory := llm.CategoryUnknown
		if catErr, ok := llm.IsCategorizedError(processingErr); ok {
			procCategory = catErr.Category()
		}

		// Default file errors to server category
		fileCategory := llm.CategoryServer
		if catErr, ok := llm.IsCategorizedError(fileSaveErr); ok {
			fileCategory = catErr.Category()
		}

		// Use the more severe category for the combined error
		category := procCategory
		if fileCategory == llm.CategoryAuth ||
			fileCategory == llm.CategoryInsufficientCredits ||
			fileCategory == llm.CategoryRateLimit {
			category = fileCategory
		}

		// Create a combined error message
		msg := fmt.Sprintf("model processing errors and file save errors occurred: %v; additionally: %v",
			processingErr, fileSaveErr)
		err := llm.Wrap(processingErr, "orchestrator", msg, category)

		contextLogger.ErrorContext(ctx, "Completed with both model and file errors: %v", err)

		// Log the completion outcome with audit logger
		o.logAuditEvent(ctx, "ExecuteEnd", "Failure",
			map[string]interface{}{
				"error_type":     "combined_errors",
				"error_category": category.String(),
			},
			nil, err)

		return err
	} else if fileSaveErr != nil {
		// Only file save errors occurred
		// Check if this is a synthesis model error that needs to be handled specially
		if strings.Contains(fileSaveErr.Error(), "synthesis model") {
			// For synthesis model errors, preserve the original error message for better debugging
			var wrappedErr error
			if catErr, ok := llm.IsCategorizedError(fileSaveErr); ok {
				wrappedErr = llm.Wrap(fileSaveErr, "orchestrator",
					fileSaveErr.Error(), catErr.Category())
			} else {
				wrappedErr = llm.Wrap(fileSaveErr, "orchestrator",
					fileSaveErr.Error(), llm.CategoryInvalidRequest)
			}

			contextLogger.ErrorContext(ctx, "Completed with synthesis error: %v", wrappedErr)

			// Log the completion outcome with audit logger
			o.logAuditEvent(ctx, "ExecuteEnd", "Failure",
				map[string]interface{}{
					"error_type":     "synthesis_model_errors",
					"error_category": CategorizeOrchestratorError(wrappedErr).String(),
				},
				nil, wrappedErr)

			return wrappedErr
		} else {
			// File system errors are usually server-side issues
			var wrappedErr error
			if catErr, ok := llm.IsCategorizedError(fileSaveErr); ok {
				wrappedErr = llm.Wrap(fileSaveErr, "orchestrator",
					"file save operation failed", catErr.Category())
			} else {
				wrappedErr = llm.Wrap(fileSaveErr, "orchestrator",
					"file save operation failed", llm.CategoryServer)
			}

			contextLogger.ErrorContext(ctx, "Completed with file save errors: %v", wrappedErr)

			// Log the completion outcome with audit logger
			o.logAuditEvent(ctx, "ExecuteEnd", "Failure",
				map[string]interface{}{
					"error_type":     "file_save_errors",
					"error_category": CategorizeOrchestratorError(wrappedErr).String(),
				},
				nil, wrappedErr)

			return wrappedErr
		}
	} else if processingErr != nil {
		// Only model errors - these should already be LLMError types
		// from the processor, but we'll ensure consistency
		var wrappedErr error
		if catErr, ok := llm.IsCategorizedError(processingErr); ok {
			wrappedErr = llm.Wrap(processingErr, "orchestrator",
				"model processing errors occurred", catErr.Category())
		} else {
			wrappedErr = llm.Wrap(processingErr, "orchestrator",
				"model processing errors occurred", llm.CategoryInvalidRequest)
		}

		contextLogger.ErrorContext(ctx, "Completed with model errors: %v", wrappedErr)

		// Log the completion outcome with audit logger
		o.logAuditEvent(ctx, "ExecuteEnd", "Failure",
			map[string]interface{}{
				"error_type":     "model_processing_errors",
				"error_category": CategorizeOrchestratorError(wrappedErr).String(),
			},
			nil, wrappedErr)

		return wrappedErr
	} else {
		// No errors
		contextLogger.InfoContext(ctx, "Processing completed successfully")

		// Log the successful completion with audit logger
		o.logAuditEvent(ctx, "ExecuteEnd", "Success",
			map[string]interface{}{"status": "complete"},
			map[string]interface{}{"outcome": "success"}, nil)

		return nil
	}
}

// aggregateErrorMessages combines multiple error messages into a single string.
// It takes a slice of errors and returns a string with each error message separated
// by a semicolon and space.
func aggregateErrorMessages(errs []error) string {
	if len(errs) == 0 {
		return ""
	}

	var messages []string
	for _, err := range errs {
		if err != nil {
			messages = append(messages, err.Error())
		}
	}

	return strings.Join(messages, "; ")
}

// logAuditEvent is a helper method that constructs an AuditEntry and logs it via the audit logger.
// It simplifies and standardizes audit event logging across the orchestrator.
//
// Parameters:
//   - ctx: context containing correlation ID and cancellation signals
//   - op: operation name (e.g., "ModelProcessing", "SynthesisStart")
//   - status: status of the operation (e.g., "Success", "Failure", "InProgress")
//   - inputs: map of operation inputs (e.g., model names, instruction details)
//   - outputs: map of operation outputs (e.g., file paths, processing statistics)
//   - err: optional error to include in the audit entry
//
// This method extracts the correlation ID from context, handles error information properly,
// and ensures consistent audit logging across all orchestrator operations.
func (o *Orchestrator) logAuditEvent(
	ctx context.Context,
	op string,
	status string,
	inputs map[string]interface{},
	outputs map[string]interface{},
	err error,
) {
	// Ensure inputs and outputs are never nil
	if inputs == nil {
		inputs = make(map[string]interface{})
	}
	if outputs == nil {
		outputs = make(map[string]interface{})
	}

	// Add correlation ID from context if available
	correlationID := logutil.GetCorrelationID(ctx)
	if correlationID != "" {
		inputs["correlation_id"] = correlationID
	}

	// Log using the audit logger's LogOp method with context
	if logErr := o.auditLogger.LogOp(ctx, op, status, inputs, outputs, err); logErr != nil {
		// Log any errors that occur during audit logging using the regular logger
		o.logger.WarnContext(ctx, "Failed to write audit log: %v", logErr)
	}
}
