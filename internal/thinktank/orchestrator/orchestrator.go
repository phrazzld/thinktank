// Package orchestrator is responsible for coordinating the core application workflow.
// It brings together various components like context gathering, API interaction,
// token management, and output writing to execute the main task defined
// by user instructions and configuration.
package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
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
	apiService      interfaces.APIService
	contextGatherer interfaces.ContextGatherer
	fileWriter      interfaces.FileWriter
	auditLogger     auditlog.AuditLogger
	rateLimiter     *ratelimit.RateLimiter
	config          *config.CliConfig
	logger          logutil.LoggerInterface
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
	return &Orchestrator{
		apiService:      apiService,
		contextGatherer: contextGatherer,
		fileWriter:      fileWriter,
		auditLogger:     auditLogger,
		rateLimiter:     rateLimiter,
		config:          config,
		logger:          logger,
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
	// Validate that models are specified
	if len(o.config.ModelNames) == 0 {
		return errors.New("no model names specified, at least one model is required")
	}

	// STEP 1: Gather context from files
	contextFiles, contextStats, err := o.gatherProjectContext(ctx)
	if err != nil {
		return err
	}

	// STEP 2: Handle dry run mode (short-circuit if dry run)
	if o.config.DryRun {
		return o.handleDryRun(ctx, contextStats)
	}

	// STEP 3: Build prompt by combining instructions and context
	stitchedPrompt := o.buildPrompt(instructions, contextFiles)

	// STEP 4: Process models concurrently
	o.logRateLimitingConfiguration()
	modelOutputs, modelErrors := o.processModels(ctx, stitchedPrompt)

	// STEP 5: Handle any errors from model processing
	if len(modelErrors) > 0 {
		return o.aggregateAndFormatErrors(modelErrors)
	}

	// STEP 6: Handle synthesis or individual model outputs based on configuration
	if o.config.SynthesisModel == "" {
		// No synthesis model specified - save individual model outputs
		o.logger.Info("Processing completed, saving individual model outputs")
		o.logger.Debug("Collected %d model outputs", len(modelOutputs))

		// Track stats for logging
		savedCount := 0
		errorCount := 0

		// Iterate over the model outputs and save each to a file
		for modelName, content := range modelOutputs {
			// Sanitize model name for use in filename
			sanitizedModelName := sanitizeFilename(modelName)

			// Construct output file path
			outputFilePath := filepath.Join(o.config.OutputDir, sanitizedModelName+".md")

			// Save the output to file
			o.logger.Debug("Saving output for model %s to %s", modelName, outputFilePath)
			if err := o.fileWriter.SaveToFile(content, outputFilePath); err != nil {
				o.logger.Error("Failed to save output for model %s: %v", modelName, err)
				errorCount++
			} else {
				savedCount++
				o.logger.Info("Successfully saved output for model %s", modelName)
			}
		}

		// Log summary of file operations
		if errorCount > 0 {
			o.logger.Error("Completed with errors: %d files saved successfully, %d files failed",
				savedCount, errorCount)
		} else {
			o.logger.Info("All %d model outputs saved successfully", savedCount)
		}
	} else {
		// Synthesis model specified - process all outputs with synthesis model
		o.logger.Info("Processing completed, synthesizing results with model: %s", o.config.SynthesisModel)
		o.logger.Debug("Synthesizing %d model outputs", len(modelOutputs))

		// Only proceed with synthesis if we have model outputs to synthesize
		if len(modelOutputs) > 0 {
			// Attempt to synthesize results
			synthesisContent, err := o.synthesizeResults(ctx, instructions, modelOutputs)
			if err != nil {
				// Process the error with specialized handling
				return o.handleSynthesisError(err)
			}

			// Log synthesis success
			o.logger.Info("Successfully synthesized results from %d model outputs", len(modelOutputs))
			o.logger.Debug("Synthesis output length: %d characters", len(synthesisContent))

			// Sanitize model name for use in filename
			sanitizedModelName := sanitizeFilename(o.config.SynthesisModel)

			// Construct output file path with -synthesis suffix
			outputFilePath := filepath.Join(o.config.OutputDir, sanitizedModelName+"-synthesis.md")

			// Save the synthesis output to file
			o.logger.Debug("Saving synthesis output to %s", outputFilePath)
			if err := o.fileWriter.SaveToFile(synthesisContent, outputFilePath); err != nil {
				o.logger.Error("Failed to save synthesis output: %v", err)
			} else {
				o.logger.Info("Successfully saved synthesis output to %s", outputFilePath)
			}
		} else {
			o.logger.Warn("No model outputs available for synthesis")
		}
	}

	return nil
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
func (o *Orchestrator) logRateLimitingConfiguration() {
	if o.config.MaxConcurrentRequests > 0 {
		o.logger.Info("Concurrency limited to %d simultaneous requests", o.config.MaxConcurrentRequests)
	} else {
		o.logger.Info("No concurrency limit applied")
	}

	if o.config.RateLimitRequestsPerMinute > 0 {
		o.logger.Info("Rate limited to %d requests per minute per model", o.config.RateLimitRequestsPerMinute)
	} else {
		o.logger.Info("No rate limit applied")
	}

	o.logger.Info("Processing %d models concurrently...", len(o.config.ModelNames))
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
// Returns:
// - A map of model names to their generated content (empty for models that failed)
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

	for result := range resultChan {
		// Always store the output (which may be empty if there was an error)
		modelOutputs[result.modelName] = result.content

		// Collect errors separately
		if result.err != nil {
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

	// Acquire rate limiting permission
	o.logger.Debug("Attempting to acquire rate limiter for model %s...", modelName)
	acquireStart := time.Now()
	if err := o.rateLimiter.Acquire(ctx, modelName); err != nil {
		o.logger.Error("Rate limiting error for model %s: %v", modelName, err)
		resultChan <- modelResult{
			modelName: modelName,
			content:   "",
			err:       fmt.Errorf("model %s rate limit: %w", modelName, err),
		}
		return
	}
	acquireDuration := time.Since(acquireStart)
	o.logger.Debug("Rate limiter acquired for model %s (waited %v)", modelName, acquireDuration)

	// Release rate limiter when done
	defer func() {
		o.logger.Debug("Releasing rate limiter for model %s", modelName)
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
		o.logger.Error("Processing model %s failed: %v", modelName, err)
		resultChan <- modelResult{
			modelName: modelName,
			content:   "",
			err:       fmt.Errorf("model %s: %w", modelName, err),
		}
		return
	}

	// Send successful result
	o.logger.Debug("Processing model %s completed successfully", modelName)
	resultChan <- modelResult{
		modelName: modelName,
		content:   content,
		err:       nil,
	}
}

// aggregateAndFormatErrors combines multiple errors into a single, user-friendly error message.
// This method consolidates errors from multiple model processing operations into
// a coherent error message for the user. It specially handles rate limit errors
// by providing additional guidance on how to adjust configuration parameters
// to avoid these errors in the future. This approach ensures users receive
// actionable feedback when errors occur.
func (o *Orchestrator) aggregateAndFormatErrors(modelErrors []error) error {
	// If there are no errors, return nil
	if len(modelErrors) == 0 {
		return nil
	}

	// Count rate limit errors
	var rateLimitErrors []error
	for _, err := range modelErrors {
		if strings.Contains(err.Error(), "rate limit") {
			rateLimitErrors = append(rateLimitErrors, err)
		}
	}

	// Build the error message
	errMsg := "errors occurred during model processing:"
	for _, e := range modelErrors {
		errMsg += "\n  - " + e.Error()
	}

	// Add rate limit guidance if applicable
	if len(rateLimitErrors) > 0 {
		errMsg += "\n\nTip: If you're encountering rate limit errors, consider adjusting the --max-concurrent and --rate-limit flags to prevent overwhelming the API."
	}

	return errors.New(errMsg)
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
	return nil, fmt.Errorf("GetModelDefinition not supported by the underlying APIService implementation")
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
	return 0, 0, fmt.Errorf("GetModelTokenLimits not supported by the underlying APIService implementation")
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

// sanitizeFilename replaces characters that are not valid in filenames
// with safe alternatives to ensure filenames are valid across different operating systems.
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
		" ", "_", // Replace spaces with underscores for better readability
	)
	return replacer.Replace(filename)
}

// synthesizeResults processes multiple model outputs through a synthesis model.
// It builds a prompt that includes the original instructions and all model outputs,
// then sends this to the synthesis model to generate a consolidated result.
//
// This is a key component of the synthesis feature, which allows combining outputs
// from multiple models into a single coherent response. The method:
// 1. Creates a specially formatted synthesis prompt using StitchSynthesisPrompt
// 2. Initializes a client for the synthesis model
// 3. Calls the synthesis model API with the combined prompt
// 4. Processes and returns the synthesized result
//
// The method includes comprehensive audit logging at each step of the synthesis process,
// allowing for detailed tracking and debugging.
//
// Parameters:
//   - ctx: The context for API client communication
//   - originalInstructions: The user's original instructions
//   - modelOutputs: A map of model names to their generated content
//
// Returns:
//   - A string containing the synthesized result
//   - An error if any step of the synthesis process fails
func (o *Orchestrator) synthesizeResults(ctx context.Context, originalInstructions string, modelOutputs map[string]string) (string, error) {
	startTime := time.Now()

	// Log synthesis process start with audit logger
	o.logAuditEvent(auditlog.AuditEntry{
		Operation: "SynthesisStart",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"synthesis_model": o.config.SynthesisModel,
			"model_count":     len(modelOutputs),
			"model_names":     getMapKeys(modelOutputs),
		},
		Message: fmt.Sprintf("Starting synthesis with model %s, processing %d outputs",
			o.config.SynthesisModel, len(modelOutputs)),
	})

	// Build synthesis prompt using the dedicated prompt function
	o.logger.Debug("Building synthesis prompt")
	synthesisPrompt := prompt.StitchSynthesisPrompt(originalInstructions, modelOutputs)
	o.logger.Debug("Synthesis prompt built, length: %d characters", len(synthesisPrompt))

	// Log prompt building completed
	o.logAuditEvent(auditlog.AuditEntry{
		Operation: "SynthesisPromptCreated",
		Status:    "Success",
		Inputs: map[string]interface{}{
			"synthesis_model": o.config.SynthesisModel,
			"prompt_length":   len(synthesisPrompt),
		},
		Message: "Synthesis prompt built successfully",
	})

	// Get model parameters
	o.logger.Debug("Getting model parameters for synthesis model: %s", o.config.SynthesisModel)
	modelParams, err := o.apiService.GetModelParameters(o.config.SynthesisModel)
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		durationMs := duration

		// Log error with audit logger
		o.logAuditEvent(auditlog.AuditEntry{
			Operation:  "SynthesisModelParameters",
			Status:     "Failure",
			DurationMs: &durationMs,
			Inputs: map[string]interface{}{
				"synthesis_model": o.config.SynthesisModel,
			},
			Error: &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "ModelParameterError",
			},
			Message: fmt.Sprintf("Failed to get parameters for synthesis model %s", o.config.SynthesisModel),
		})

		return "", fmt.Errorf("failed to get model parameters for synthesis model: %w", err)
	}

	// Log successful parameter retrieval
	o.logAuditEvent(auditlog.AuditEntry{
		Operation: "SynthesisModelParameters",
		Status:    "Success",
		Inputs: map[string]interface{}{
			"synthesis_model": o.config.SynthesisModel,
			"param_count":     len(modelParams),
		},
		Message: "Successfully retrieved synthesis model parameters",
	})

	// Get client for synthesis model
	o.logger.Debug("Initializing client for synthesis model: %s", o.config.SynthesisModel)
	client, err := o.apiService.InitLLMClient(ctx, "", o.config.SynthesisModel, "")
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		durationMs := duration

		// Log error with audit logger
		o.logAuditEvent(auditlog.AuditEntry{
			Operation:  "SynthesisClientInit",
			Status:     "Failure",
			DurationMs: &durationMs,
			Inputs: map[string]interface{}{
				"synthesis_model": o.config.SynthesisModel,
			},
			Error: &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "ClientInitializationError",
			},
			Message: fmt.Sprintf("Failed to initialize client for synthesis model %s", o.config.SynthesisModel),
		})

		return "", fmt.Errorf("failed to initialize synthesis model client: %w", err)
	}

	// Log successful client initialization
	o.logAuditEvent(auditlog.AuditEntry{
		Operation: "SynthesisClientInit",
		Status:    "Success",
		Inputs: map[string]interface{}{
			"synthesis_model": o.config.SynthesisModel,
		},
		Message: "Successfully initialized synthesis model client",
	})

	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			o.logger.Warn("Error closing synthesis model client: %v", closeErr)

			// Log client close error
			o.logAuditEvent(auditlog.AuditEntry{
				Operation: "SynthesisClientClose",
				Status:    "Failure",
				Inputs: map[string]interface{}{
					"synthesis_model": o.config.SynthesisModel,
				},
				Error: &auditlog.ErrorInfo{
					Message: closeErr.Error(),
					Type:    "ClientCloseError",
				},
				Message: "Error closing synthesis model client",
			})
		} else {
			// Log successful client close
			o.logAuditEvent(auditlog.AuditEntry{
				Operation: "SynthesisClientClose",
				Status:    "Success",
				Inputs: map[string]interface{}{
					"synthesis_model": o.config.SynthesisModel,
				},
				Message: "Successfully closed synthesis model client",
			})
		}
	}()

	// Log API call start
	apiCallStartTime := time.Now()
	o.logAuditEvent(auditlog.AuditEntry{
		Operation: "SynthesisAPICall",
		Status:    "InProgress",
		Inputs: map[string]interface{}{
			"synthesis_model": o.config.SynthesisModel,
			"prompt_length":   len(synthesisPrompt),
		},
		Message: fmt.Sprintf("Calling synthesis model API: %s", o.config.SynthesisModel),
	})

	// Call model API
	o.logger.Info("Calling synthesis model API: %s", o.config.SynthesisModel)
	result, err := client.GenerateContent(ctx, synthesisPrompt, modelParams)

	// Calculate API call duration
	apiCallDurationMs := time.Since(apiCallStartTime).Milliseconds()

	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		durationMs := duration

		// Log API call failure
		o.logAuditEvent(auditlog.AuditEntry{
			Operation:  "SynthesisAPICall",
			Status:     "Failure",
			DurationMs: &apiCallDurationMs,
			Inputs: map[string]interface{}{
				"synthesis_model": o.config.SynthesisModel,
				"prompt_length":   len(synthesisPrompt),
			},
			Error: &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "APICallError",
			},
			Message: fmt.Sprintf("Synthesis model API call failed for model %s", o.config.SynthesisModel),
		})

		// Log overall synthesis failure
		o.logAuditEvent(auditlog.AuditEntry{
			Operation:  "SynthesisEnd",
			Status:     "Failure",
			DurationMs: &durationMs,
			Inputs: map[string]interface{}{
				"synthesis_model": o.config.SynthesisModel,
			},
			Error: &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "SynthesisError",
			},
			Message: fmt.Sprintf("Synthesis process failed with model %s", o.config.SynthesisModel),
		})

		return "", fmt.Errorf("synthesis model API call failed: %w", err)
	}

	// Log successful API call
	o.logAuditEvent(auditlog.AuditEntry{
		Operation:  "SynthesisAPICall",
		Status:     "Success",
		DurationMs: &apiCallDurationMs,
		Inputs: map[string]interface{}{
			"synthesis_model": o.config.SynthesisModel,
		},
		Outputs: map[string]interface{}{
			"result_received": result != nil,
		},
		Message: "Synthesis model API call completed successfully",
	})

	// Process response
	responseStartTime := time.Now()
	o.logger.Debug("Processing synthesis model response")

	synthesisOutput, err := o.apiService.ProcessLLMResponse(result)

	// Calculate response processing duration
	responseDurationMs := time.Since(responseStartTime).Milliseconds()

	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		durationMs := duration

		// Log response processing failure
		o.logAuditEvent(auditlog.AuditEntry{
			Operation:  "SynthesisResponseProcessing",
			Status:     "Failure",
			DurationMs: &responseDurationMs,
			Inputs: map[string]interface{}{
				"synthesis_model": o.config.SynthesisModel,
			},
			Error: &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "ResponseProcessingError",
			},
			Message: "Failed to process synthesis model response",
		})

		// Log overall synthesis failure
		o.logAuditEvent(auditlog.AuditEntry{
			Operation:  "SynthesisEnd",
			Status:     "Failure",
			DurationMs: &durationMs,
			Inputs: map[string]interface{}{
				"synthesis_model": o.config.SynthesisModel,
			},
			Error: &auditlog.ErrorInfo{
				Message: err.Error(),
				Type:    "SynthesisError",
			},
			Message: fmt.Sprintf("Synthesis process failed with model %s", o.config.SynthesisModel),
		})

		return "", fmt.Errorf("failed to process synthesis model response: %w", err)
	}

	// Log successful response processing
	o.logAuditEvent(auditlog.AuditEntry{
		Operation:  "SynthesisResponseProcessing",
		Status:     "Success",
		DurationMs: &responseDurationMs,
		Inputs: map[string]interface{}{
			"synthesis_model": o.config.SynthesisModel,
		},
		Outputs: map[string]interface{}{
			"output_length": len(synthesisOutput),
		},
		Message: "Successfully processed synthesis model response",
	})

	// Calculate total duration and log overall success
	duration := time.Since(startTime).Milliseconds()
	durationMs := duration

	o.logAuditEvent(auditlog.AuditEntry{
		Operation:  "SynthesisEnd",
		Status:     "Success",
		DurationMs: &durationMs,
		Inputs: map[string]interface{}{
			"synthesis_model": o.config.SynthesisModel,
			"model_count":     len(modelOutputs),
		},
		Outputs: map[string]interface{}{
			"output_length": len(synthesisOutput),
		},
		Message: fmt.Sprintf("Synthesis completed successfully with model %s", o.config.SynthesisModel),
	})

	return synthesisOutput, nil
}

// handleSynthesisError processes a synthesis error and generates a user-friendly error message.
// It categorizes the error based on its type and adds helpful guidance for the user.
//
// This method is a key part of the synthesis feature's error handling, providing
// detailed, actionable feedback when synthesis fails. It analyzes the error type
// (e.g., rate limiting, content filtering, connectivity, authentication) and
// provides specific guidance for each case. For example, rate limit errors include
// tips on waiting and retrying, while content filtering errors suggest reviewing prompts.
//
// This ensures that synthesis errors are presented in a clear, actionable format,
// improving the user experience when problems occur.
func (o *Orchestrator) handleSynthesisError(err error) error {
	// Log the detailed error for debugging purposes
	o.logger.Error("Synthesis error occurred: %v", err)

	var errMsg string

	// Check for specific error types
	switch {
	case strings.Contains(err.Error(), "rate limit"):
		// Rate limiting error
		errMsg = fmt.Sprintf("Synthesis model '%s' encountered rate limiting: %v", o.config.SynthesisModel, err)
		errMsg += "\n\nTip: If you're encountering rate limit errors, consider waiting a moment before retrying or using a different synthesis model."

	case strings.Contains(err.Error(), "safety") || strings.Contains(err.Error(), "content filter") || o.apiService.IsSafetyBlockedError(err):
		// Content filtering/safety error
		errMsg = fmt.Sprintf("Synthesis was blocked by content safety filters: %v", err)
		errMsg += "\n\nTip: Your input or the model outputs may contain content that triggered safety filters. Consider reviewing your prompts and instructions."

	case strings.Contains(err.Error(), "connect") || strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline"):
		// Connectivity error
		errMsg = fmt.Sprintf("Connectivity issue when calling synthesis model '%s': %v", o.config.SynthesisModel, err)
		errMsg += "\n\nTip: Check your internet connection and try again. If the issue persists, the service might be experiencing temporary problems."

	case strings.Contains(err.Error(), "auth") || strings.Contains(err.Error(), "key") || strings.Contains(err.Error(), "credential"):
		// Authentication error
		errMsg = fmt.Sprintf("Authentication failed for synthesis model '%s': %v", o.config.SynthesisModel, err)
		errMsg += "\n\nTip: Check your API key and ensure it has proper permissions for this model."

	case o.apiService.IsEmptyResponseError(err):
		// Empty response error
		errMsg = fmt.Sprintf("Synthesis model '%s' returned an empty response: %v", o.config.SynthesisModel, err)
		errMsg += "\n\nTip: The model might be having trouble with your instructions or the outputs from other models. Try simplifying the instructions."

	default:
		// Generic error with model details
		errMsg = fmt.Sprintf("Error synthesizing results with model '%s': %v", o.config.SynthesisModel, err)

		// Add any additional details from API service
		if details := o.apiService.GetErrorDetails(err); details != "" {
			errMsg += "\nDetails: " + details
		}

		errMsg += "\n\nTip: If this error persists, try using a different synthesis model or check the model's documentation for limitations."
	}

	// Return a new error with the formatted message
	return fmt.Errorf("synthesis failure: %s", errMsg)
}

// NOTE: Previous versions used a TokenManagerAdapter between interfaces.TokenManager
// and modelproc.TokenManager. This has been replaced by direct creation of TokenManager
// instances in ModelProcessor.Process with model-specific LLMClient instances.

// getMapKeys extracts and returns all keys from a map as a sorted string slice.
// This helper function is used for audit logging to capture all model names.
func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// logAuditEvent writes an audit log entry and logs any errors that occur.
// This helper ensures proper error handling for all audit log operations.
func (o *Orchestrator) logAuditEvent(entry auditlog.AuditEntry) {
	if err := o.auditLogger.Log(entry); err != nil {
		o.logger.Warn("Failed to write audit log: %v", err)
	}
}
