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
// It coordinates the entire process from context gathering to model processing:
// 1. Gather context from project files based on configuration
// 2. Handle dry run mode by displaying statistics (if enabled)
// 3. Build the complete prompt by combining instructions with context
// 4. Process all configured models concurrently with rate limiting
// 5. Handle and format any errors that occurred during processing
// 6. Based on configuration, either:
//   - Save individual model outputs (when no synthesis model is specified)
//   - Synthesize results using a designated synthesis model
//
// The synthesis workflow branch (activated when SynthesisModel is specified in config)
// will send all individual model outputs to the synthesis model, which combines
// them into a unified response. This synthesis output is saved as
// <synthesis-model-name>-synthesis.md in the output directory, alongside the
// individual model outputs.
//
// The method enforces a clear separation of concerns by delegating specific tasks
// to helper methods, making the high-level workflow easy to understand.
func (o *Orchestrator) Run(ctx context.Context, instructions string) error {
	// Ensure context has a correlation ID for tracing and structured logging
	ctx = logutil.WithCorrelationID(ctx)

	// Create a logger with the context for all subsequent logging
	contextLogger := o.logger.WithContext(ctx)

	// Validate that models are specified
	if len(o.config.ModelNames) == 0 {
		return errors.New("no model names specified, at least one model is required")
	}

	// Log the start of processing
	contextLogger.InfoContext(ctx, "Starting processing")

	// STEP 1: Gather context from files
	contextLogger.DebugContext(ctx, "Gathering project context")
	contextFiles, contextStats, err := o.gatherProjectContext(ctx)
	if err != nil {
		contextLogger.ErrorContext(ctx, "Failed to gather project context: %v", err)
		return err
	}
	contextLogger.DebugContext(ctx, "Project context gathered: %d files", len(contextFiles))

	// STEP 2: Handle dry run mode (short-circuit if dry run)
	dryRunExecuted, err := o.runDryRunFlow(ctx, contextStats)
	if err != nil {
		return err
	}
	if dryRunExecuted {
		return nil
	}

	// STEP 3: Build prompt by combining instructions and context
	contextLogger.DebugContext(ctx, "Building prompt from instructions and context")
	stitchedPrompt := o.buildPrompt(instructions, contextFiles)
	contextLogger.DebugContext(ctx, "Prompt built successfully, length: %d characters", len(stitchedPrompt))

	// STEP 4: Process models concurrently
	contextLogger.InfoContext(ctx, "Beginning model processing")
	o.logRateLimitingConfiguration(ctx)
	modelOutputs, modelErrors := o.processModels(ctx, stitchedPrompt)

	// STEP 5: Handle model processing errors
	// Store any model errors to return later - we'll proceed with synthesis if possible
	var returnErr error
	if len(modelErrors) > 0 {
		// If ALL models failed (no outputs available), fail immediately
		if len(modelOutputs) == 0 {
			errorMsg := fmt.Sprintf("all models failed: %v", aggregateErrorMessages(modelErrors))
			contextLogger.ErrorContext(ctx, errorMsg)
			return errors.New(errorMsg)
		}

		// Otherwise, log errors but continue with available outputs
		// Note: modelOutputs only contains entries for successful models (no empty entries for failed models)
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
		returnErr = fmt.Errorf("processed %d/%d models successfully; %d failed: %v",
			len(modelOutputs), len(o.config.ModelNames), len(modelErrors),
			aggregateErrorMessages(modelErrors))
	}

	// STEP 6: Handle synthesis or individual model outputs based on configuration
	// We'll track file-save errors separately from model processing errors
	var fileSaveErrors error

	if o.config.SynthesisModel == "" {
		// No synthesis model specified - save individual model outputs
		contextLogger.InfoContext(ctx, "Processing completed, saving individual model outputs")
		contextLogger.DebugContext(ctx, "Collected %d model outputs", len(modelOutputs))

		// Use the OutputWriter to save individual model outputs
		savedCount, err := o.outputWriter.SaveIndividualOutputs(ctx, modelOutputs, o.config.OutputDir)
		if err != nil {
			contextLogger.ErrorContext(ctx, "Completed with errors: %d files saved successfully, %d files failed",
				savedCount, len(modelOutputs)-savedCount)
			fileSaveErrors = err
		} else {
			contextLogger.InfoContext(ctx, "All %d model outputs saved successfully", savedCount)
		}
	} else {
		// Synthesis model specified - process all outputs with synthesis model
		contextLogger.InfoContext(ctx, "Processing completed, synthesizing results with model: %s", o.config.SynthesisModel)
		contextLogger.DebugContext(ctx, "Synthesizing %d model outputs", len(modelOutputs))

		// Only proceed with synthesis if we have model outputs to synthesize
		if len(modelOutputs) > 0 {
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
				fileSaveErrors = err
			} else {
				contextLogger.InfoContext(ctx, "Successfully saved synthesis output")
			}
		} else {
			contextLogger.WarnContext(ctx, "No model outputs available for synthesis")
		}
	}

	// Return any model errors or file save errors that occurred, combining them if both exist
	if returnErr != nil && fileSaveErrors != nil {
		// Combine model and file save errors
		err := fmt.Errorf("model processing errors and file save errors occurred: %w; additionally: %v",
			returnErr, fileSaveErrors)
		contextLogger.ErrorContext(ctx, "Completed with both model and file errors: %v", err)
		return err
	} else if fileSaveErrors != nil {
		// Only file save errors occurred
		contextLogger.ErrorContext(ctx, "Completed with file save errors: %v", fileSaveErrors)
		return fileSaveErrors
	} else if returnErr != nil {
		// Only model errors
		contextLogger.ErrorContext(ctx, "Completed with model errors: %v", returnErr)
		return returnErr
	} else {
		// No errors
		contextLogger.InfoContext(ctx, "Processing completed successfully")
		return nil
	}
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
