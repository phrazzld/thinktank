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

	"github.com/phrazzld/architect/internal/architect/interfaces"
	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/architect/prompt"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/ratelimit"
)

// Orchestrator coordinates the main application logic.
// It depends on various services to perform tasks like interacting with the API,
// gathering context, managing tokens, writing files, logging audits, and handling rate limits.
type Orchestrator struct {
	apiService      interfaces.APIService
	contextGatherer interfaces.ContextGatherer
	tokenManager    interfaces.TokenManager
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
	tokenManager interfaces.TokenManager,
	fileWriter interfaces.FileWriter,
	auditLogger auditlog.AuditLogger,
	rateLimiter *ratelimit.RateLimiter,
	config *config.CliConfig,
	logger logutil.LoggerInterface,
) *Orchestrator {
	return &Orchestrator{
		apiService:      apiService,
		contextGatherer: contextGatherer,
		tokenManager:    tokenManager,
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
	modelErrors := o.processModels(ctx, stitchedPrompt)

	// STEP 5: Handle any errors from model processing
	if len(modelErrors) > 0 {
		return o.aggregateAndFormatErrors(modelErrors)
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
// goroutines, each handling a different model, and collects any errors that
// occur during processing. This approach significantly improves throughput
// when multiple models are specified.
//
// Returns a slice of errors encountered during processing, which will be empty
// if all models were processed successfully.
func (o *Orchestrator) processModels(ctx context.Context, stitchedPrompt string) []error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(o.config.ModelNames))

	// Launch a goroutine for each model
	for _, modelName := range o.config.ModelNames {
		wg.Add(1)
		go o.processModelWithRateLimit(ctx, modelName, stitchedPrompt, &wg, errChan)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Collect errors from the channel
	var modelErrors []error
	for err := range errChan {
		modelErrors = append(modelErrors, err)
	}

	return modelErrors
}

// processModelWithRateLimit processes a single model with rate limiting.
func (o *Orchestrator) processModelWithRateLimit(
	ctx context.Context,
	modelName string,
	stitchedPrompt string,
	wg *sync.WaitGroup,
	errChan chan<- error,
) {
	defer wg.Done()

	// Acquire rate limiting permission
	o.logger.Debug("Attempting to acquire rate limiter for model %s...", modelName)
	acquireStart := time.Now()
	if err := o.rateLimiter.Acquire(ctx, modelName); err != nil {
		o.logger.Error("Rate limiting error for model %s: %v", modelName, err)
		errChan <- fmt.Errorf("model %s rate limit: %w", modelName, err)
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
		nil, // tokenManager is created inside the Process method
		o.fileWriter,
		o.auditLogger,
		o.logger,
		o.config,
	)

	// Process the model
	err := processor.Process(ctx, modelName, stitchedPrompt)
	if err != nil {
		o.logger.Error("Processing model %s failed: %v", modelName, err)
		errChan <- fmt.Errorf("model %s: %w", modelName, err)
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

// InitClient initializes and returns a Gemini client for the specified model.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) InitClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
	return a.APIService.InitClient(ctx, apiKey, modelName, apiEndpoint)
}

// ProcessResponse extracts content from the API response.
// It delegates to the underlying APIService implementation.
func (a *APIServiceAdapter) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	return a.APIService.ProcessResponse(result)
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

// TokenManagerAdapter is a deprecated adapter that was used to adapt interfaces.TokenManager
// to modelproc.TokenManager. It is no longer needed since TokenManagers are now created directly
// in the ModelProcessor.Process method with model-specific clients.
// This documentation is kept as a historical note, as the adapter code has been removed.
// The current design eliminates the need for this adapter by having ModelProcessor create
// its own TokenManager instances tailored to each specific model, improving encapsulation.
