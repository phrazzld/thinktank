// Package orchestrator is responsible for coordinating the core application workflow.
// It brings together various components like context gathering, API interaction,
// token management, and output writing to execute the main task defined
// by user instructions and configuration.
package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
	"github.com/phrazzld/thinktank/internal/thinktank/modelproc"
	"github.com/phrazzld/thinktank/internal/thinktank/prompt"
)

// Orchestrator coordinates the main application logic.
// It depends on various services to perform tasks like interacting with the API,
// gathering context, writing files, logging audits, and handling rate limits.
type Orchestrator struct {
	apiService       interfaces.APIService
	contextGatherer  interfaces.ContextGatherer
	fileWriter       interfaces.FileWriter
	auditLogger      auditlog.AuditLogger
	rateLimiter      *ratelimit.RateLimiter
	config           *config.CliConfig
	logger           logutil.LoggerInterface
	synthesisService SynthesisService
	outputWriter     OutputWriter
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
) *Orchestrator {
	// Create the output writer
	outputWriter := NewOutputWriter(fileWriter, auditLogger, logger)

	// Create a synthesis service only if synthesis model is specified
	var synthesisService SynthesisService
	if config.SynthesisModel != "" {
		synthesisService = NewSynthesisService(apiService, auditLogger, logger, config.SynthesisModel)
	}

	return &Orchestrator{
		apiService:       apiService,
		contextGatherer:  contextGatherer,
		fileWriter:       fileWriter,
		auditLogger:      auditLogger,
		rateLimiter:      rateLimiter,
		config:           config,
		logger:           logger,
		synthesisService: synthesisService,
		outputWriter:     outputWriter,
	}
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
// 7. Handle and report any errors
//
// Each step is delegated to a specialized helper method, making the workflow
// clear and maintainable.
func (o *Orchestrator) Run(ctx context.Context, instructions string) error {
	// Setup: initialize context and validate configuration
	ctx, contextLogger, err := o.setupContext(ctx)
	if err != nil {
		return err
	}

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
	stitchedPrompt := o.buildPrompt(instructions, contextFiles)

	// Step 4: Process all models and handle errors
	modelOutputs, processingErr, criticalErr := o.processModelsWithErrorHandling(ctx, stitchedPrompt, contextLogger)
	if criticalErr != nil {
		return criticalErr
	}

	// Step 5: Save outputs (via synthesis or individually)
	fileSaveErr := o.handleOutputFlow(ctx, instructions, modelOutputs)

	// Step 6: Final error processing and return
	return o.handleProcessingOutcome(ctx, processingErr, fileSaveErr, contextLogger)
}

// gatherProjectContext collects relevant files from the project based on configuration.
func (o *Orchestrator) gatherProjectContext(ctx context.Context) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
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
		return nil, nil, fmt.Errorf("failed during project context gathering: %w", err)
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
// Returns an error if any of the outputs fail to save.
func (o *Orchestrator) runIndividualOutputFlow(ctx context.Context, modelOutputs map[string]string) error {
	// Get logger with context
	contextLogger := o.logger.WithContext(ctx)

	// Log that individual outputs are being saved
	contextLogger.InfoContext(ctx, "Processing completed, saving individual model outputs")
	contextLogger.DebugContext(ctx, "Collected %d model outputs", len(modelOutputs))

	// Use the OutputWriter to save individual model outputs
	savedCount, err := o.outputWriter.SaveIndividualOutputs(ctx, modelOutputs, o.config.OutputDir)
	if err != nil {
		contextLogger.ErrorContext(ctx, "Completed with errors: %d files saved successfully, %d files failed",
			savedCount, len(modelOutputs)-savedCount)
		return err
	}

	contextLogger.InfoContext(ctx, "All %d model outputs saved successfully", savedCount)
	return nil
}

// runSynthesisFlow handles the synthesis of multiple model outputs using the specified synthesis model.
// It synthesizes the results and saves the output to a file.
// Returns an error if synthesis fails or if the output cannot be saved.
func (o *Orchestrator) runSynthesisFlow(ctx context.Context, instructions string, modelOutputs map[string]string) error {
	// Get logger with context
	contextLogger := o.logger.WithContext(ctx)

	// Log that we're starting synthesis
	contextLogger.InfoContext(ctx, "Processing completed, synthesizing results with model: %s", o.config.SynthesisModel)
	contextLogger.DebugContext(ctx, "Synthesizing %d model outputs", len(modelOutputs))

	// Only proceed with synthesis if we have model outputs to synthesize
	if len(modelOutputs) == 0 {
		contextLogger.WarnContext(ctx, "No model outputs available for synthesis")
		return nil
	}

	// Attempt to synthesize results using the SynthesisService
	contextLogger.InfoContext(ctx, "Starting synthesis with model: %s", o.config.SynthesisModel)
	synthesisContent, err := o.synthesisService.SynthesizeResults(ctx, instructions, modelOutputs)
	if err != nil {
		// Process the error with specialized handling
		contextLogger.ErrorContext(ctx, "Synthesis failed: %v", err)
		return err
	}

	// Log synthesis success
	contextLogger.InfoContext(ctx, "Successfully synthesized results from %d model outputs", len(modelOutputs))
	contextLogger.DebugContext(ctx, "Synthesis output length: %d characters", len(synthesisContent))

	// Save the synthesis output using the OutputWriter
	if err := o.outputWriter.SaveSynthesisOutput(ctx, synthesisContent, o.config.SynthesisModel, o.config.OutputDir); err != nil {
		contextLogger.ErrorContext(ctx, "Failed to save synthesis output: %v", err)
		return err
	}

	contextLogger.InfoContext(ctx, "Successfully saved synthesis output")
	return nil
}

// handleDryRun displays context statistics without performing API calls.
func (o *Orchestrator) handleDryRun(ctx context.Context, stats *interfaces.ContextStats) error {
	err := o.contextGatherer.DisplayDryRunInfo(ctx, stats)
	if err != nil {
		o.logger.Error("Error displaying dry run information: %v", err)
		return fmt.Errorf("error displaying dry run information: %w", err)
	}
	return nil
}

// buildPrompt creates the complete prompt by combining instructions with context files.
func (o *Orchestrator) buildPrompt(instructions string, contextFiles []fileutil.FileMeta) string {
	stitchedPrompt := prompt.StitchPrompt(instructions, contextFiles)
	o.logger.Info("Prompt constructed successfully")
	o.logger.Debug("Stitched prompt length: %d characters", len(stitchedPrompt))
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

// processModels processes each model concurrently with rate limiting.
// This is a key orchestration method that manages the concurrent execution
// of model processing while respecting rate limits. It coordinates multiple
// goroutines, each handling a different model, and collects both outputs and
// errors that occur during processing. This approach significantly improves
// throughput when multiple models are specified.
//
// When the synthesis feature is used, this method collects and returns all
// model outputs in a map, which will later be used as input for the synthesis
// model. This enables combining insights from multiple models into a cohesive
// response.
//
// IMPORTANT: Only successful model outputs (where err == nil) are added to the
// modelOutputs map. Failed models are not included in the map at all, which ensures
// accurate success counting and prevents empty/failed outputs from being included
// in synthesis prompts.
//
// Returns:
// - A map of model names to their generated content (contains only successful models)
// - A slice of errors encountered during processing (empty if all models were successful)
func (o *Orchestrator) processModels(ctx context.Context, stitchedPrompt string) (map[string]string, []error) {
	var wg sync.WaitGroup
	resultChan := make(chan modelResult, len(o.config.ModelNames))

	// Launch a goroutine for each model
	for _, modelName := range o.config.ModelNames {
		wg.Add(1)
		go o.processModelWithRateLimit(ctx, modelName, stitchedPrompt, &wg, resultChan)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultChan)

	// Collect outputs and errors from the channel
	modelOutputs := make(map[string]string)
	var modelErrors []error

	// We're processing a channel that's already closed, so there's no race condition here
	for result := range resultChan {
		// Only store output for successful models
		if result.err == nil {
			modelOutputs[result.modelName] = result.content
		} else {
			// Collect errors
			modelErrors = append(modelErrors, result.err)
		}
	}

	return modelOutputs, modelErrors
}

// modelResult represents the result of processing a single model.
// It includes the model name, generated content, and any error encountered.
// This struct is crucial for the synthesis feature as it captures outputs
// from multiple models so they can be combined by a synthesis model.
type modelResult struct {
	modelName string // Name of the processed model
	content   string // Generated content from the model, which may be used for synthesis
	err       error  // Any error encountered during processing
}

// processModelWithRateLimit processes a single model with rate limiting.
// It acquires a rate limiting token, processes the model, and sends the result
// (containing model name, content, and any error) to the result channel.
func (o *Orchestrator) processModelWithRateLimit(
	ctx context.Context,
	modelName string,
	stitchedPrompt string,
	wg *sync.WaitGroup,
	resultChan chan<- modelResult,
) {
	defer wg.Done()

	// Get logger with context
	contextLogger := o.logger.WithContext(ctx)

	// Create a local variable to store the result to avoid accessing it from multiple goroutines
	var result modelResult
	result.modelName = modelName

	// Acquire rate limiting permission
	contextLogger.DebugContext(ctx, "Attempting to acquire rate limiter for model %s...", modelName)
	acquireStart := time.Now()
	if err := o.rateLimiter.Acquire(ctx, modelName); err != nil {
		contextLogger.ErrorContext(ctx, "Rate limiting error for model %s: %v", modelName, err)
		result.err = fmt.Errorf("model %s rate limit: %w", modelName, err)
		resultChan <- result
		return
	}
	acquireDuration := time.Since(acquireStart)
	contextLogger.DebugContext(ctx, "Rate limiter acquired for model %s (waited %v)", modelName, acquireDuration)

	// Release rate limiter when done
	defer func() {
		contextLogger.DebugContext(ctx, "Releasing rate limiter for model %s", modelName)
		o.rateLimiter.Release()
	}()

	// Create API service adapter and model processor
	apiServiceAdapter := &APIServiceAdapter{APIService: o.apiService}
	processor := modelproc.NewProcessor(
		apiServiceAdapter,
		o.fileWriter,
		o.auditLogger,
		o.logger,
		o.config,
	)

	// Process the model
	content, err := processor.Process(ctx, modelName, stitchedPrompt)
	if err != nil {
		contextLogger.ErrorContext(ctx, "Processing model %s failed: %v", modelName, err)
		result.err = fmt.Errorf("model %s: %w", modelName, err)
		resultChan <- result
		return
	}

	// Store the successful result in local variable before sending
	contextLogger.DebugContext(ctx, "Processing model %s completed successfully", modelName)
	result.content = content
	result.err = nil
	resultChan <- result
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

// GetModelParameters retrieves parameter values from the registry for a given model.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) GetModelParameters(modelName string) (map[string]interface{}, error) {
	if apiService, ok := a.APIService.(interface {
		GetModelParameters(string) (map[string]interface{}, error)
	}); ok {
		return apiService.GetModelParameters(modelName)
	}
	// Return empty parameters if the underlying implementation doesn't support this method
	return make(map[string]interface{}), nil
}

// GetModelDefinition retrieves the full model definition from the registry.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) GetModelDefinition(modelName string) (*registry.ModelDefinition, error) {
	if apiService, ok := a.APIService.(interface {
		GetModelDefinition(string) (*registry.ModelDefinition, error)
	}); ok {
		return apiService.GetModelDefinition(modelName)
	}
	// Return nil with error if the underlying implementation doesn't support this method
	return nil, errors.New("GetModelDefinition not supported by the underlying APIService implementation")
}

// GetModelTokenLimits retrieves token limits from the registry for a given model.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) GetModelTokenLimits(modelName string) (contextWindow, maxOutputTokens int32, err error) {
	if apiService, ok := a.APIService.(interface {
		GetModelTokenLimits(string) (int32, int32, error)
	}); ok {
		return apiService.GetModelTokenLimits(modelName)
	}
	// Return zero values with error if the underlying implementation doesn't support this method
	return 0, 0, errors.New("GetModelTokenLimits not supported by the underlying APIService implementation")
}

// ValidateModelParameter validates a parameter value against its constraints.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) ValidateModelParameter(modelName, paramName string, value interface{}) (bool, error) {
	if apiService, ok := a.APIService.(interface {
		ValidateModelParameter(string, string, interface{}) (bool, error)
	}); ok {
		return apiService.ValidateModelParameter(modelName, paramName, value)
	}
	// Return true if the underlying implementation doesn't support this method
	return true, nil
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
		return errors.New(errorMsg)
	}

	// Some operations succeeded but others failed
	return fmt.Errorf("processed %d/%d models successfully; %d failed: %v",
		successCount, totalCount, len(errs),
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
		return ctx, contextLogger, errors.New("no model names specified, at least one model is required")
	}

	// Log the start of processing
	contextLogger.InfoContext(ctx, "Starting processing")

	return ctx, contextLogger, nil
}

// processModelsWithErrorHandling processes models and handles any errors that occur.
// It runs the model processing and handles error aggregation and logging.
// Returns the model outputs, any processing errors for later handling, and a critical
// error that should interrupt processing immediately.
func (o *Orchestrator) processModelsWithErrorHandling(ctx context.Context, stitchedPrompt string, contextLogger logutil.LoggerInterface) (map[string]string, error, error) {
	// Start model processing
	contextLogger.InfoContext(ctx, "Beginning model processing")
	o.logRateLimitingConfiguration(ctx)
	modelOutputs, modelErrors := o.processModels(ctx, stitchedPrompt)

	// Handle model processing errors
	var returnErr error
	if len(modelErrors) > 0 {
		// If ALL models failed (no outputs available), fail immediately
		if len(modelOutputs) == 0 {
			returnErr = o.aggregateErrors(modelErrors, len(o.config.ModelNames), 0)
			contextLogger.ErrorContext(ctx, returnErr.Error())
			return nil, nil, returnErr
		}

		// Otherwise, log errors but continue with available outputs
		// Get list of successful model names for the log
		var successfulModels []string
		for modelName := range modelOutputs {
			successfulModels = append(successfulModels, modelName)
		}

		// Log a warning with detailed counts and successful model names
		contextLogger.WarnContext(ctx, "Some models failed but continuing with synthesis: %d/%d models successful, %d failed. Successful models: %v",
			len(modelOutputs), len(o.config.ModelNames), len(modelErrors), successfulModels)

		// Log individual error details
		for _, err := range modelErrors {
			contextLogger.ErrorContext(ctx, "%v", err)
		}

		// Create a descriptive error to return after processing is complete
		returnErr = o.aggregateErrors(modelErrors, len(o.config.ModelNames), len(modelOutputs))
	}

	return modelOutputs, returnErr, nil
}

// handleOutputFlow decides whether to use synthesis or individual output flow
// based on configuration and handles the saving of outputs accordingly.
func (o *Orchestrator) handleOutputFlow(ctx context.Context, instructions string, modelOutputs map[string]string) error {
	if o.config.SynthesisModel == "" {
		// No synthesis model specified - save individual model outputs
		return o.runIndividualOutputFlow(ctx, modelOutputs)
	}

	// Synthesis model specified - process all outputs with synthesis model
	return o.runSynthesisFlow(ctx, instructions, modelOutputs)
}

// handleProcessingOutcome combines and reports any errors from model processing and file saving.
// It formats and logs appropriate error messages based on the types of errors encountered.
func (o *Orchestrator) handleProcessingOutcome(ctx context.Context, processingErr, fileSaveErr error, contextLogger logutil.LoggerInterface) error {
	if processingErr != nil && fileSaveErr != nil {
		// Combine model and file save errors
		err := fmt.Errorf("model processing errors and file save errors occurred: %w; additionally: %v",
			processingErr, fileSaveErr)
		contextLogger.ErrorContext(ctx, "Completed with both model and file errors: %v", err)
		return err
	} else if fileSaveErr != nil {
		// Only file save errors occurred
		contextLogger.ErrorContext(ctx, "Completed with file save errors: %v", fileSaveErr)
		return fileSaveErr
	} else if processingErr != nil {
		// Only model errors
		contextLogger.ErrorContext(ctx, "Completed with model errors: %v", processingErr)
		return processingErr
	} else {
		// No errors
		contextLogger.InfoContext(ctx, "Processing completed successfully")
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

	// Log using the audit logger's LogOp method
	if logErr := o.auditLogger.LogOp(op, status, inputs, outputs, err); logErr != nil {
		// Log any errors that occur during audit logging using the regular logger
		o.logger.WarnContext(ctx, "Failed to write audit log: %v", logErr)
	}
}
