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

// Run executes the main application workflow.
// It manages context gathering, prompt creation, and model processing.
func (o *Orchestrator) Run(ctx context.Context, instructions string) error {
	// 1. Create gather config
	gatherConfig := interfaces.GatherConfig{
		Paths:        o.config.Paths,
		Include:      o.config.Include,
		Exclude:      o.config.Exclude,
		ExcludeNames: o.config.ExcludeNames,
		Format:       o.config.Format,
		Verbose:      o.config.Verbose,
		LogLevel:     o.config.LogLevel,
	}

	// Note: The contextGatherer now has its own client injected and handles its own audit logging
	// so we don't need to create a client or perform audit logging here

	// 3. Gather context files (model-independent step)
	contextFiles, contextStats, err := o.contextGatherer.GatherContext(ctx, gatherConfig)
	if err != nil {
		return fmt.Errorf("failed during project context gathering: %w", err)
	}

	// 4. Handle dry run mode
	if o.config.DryRun {
		err = o.contextGatherer.DisplayDryRunInfo(ctx, contextStats)
		if err != nil {
			o.logger.Error("Error displaying dry run information: %v", err)
			return fmt.Errorf("error displaying dry run information: %w", err)
		}
		return nil
	}

	// 5. Stitch prompt (model-independent step)
	stitchedPrompt := prompt.StitchPrompt(instructions, contextFiles)
	o.logger.Info("Prompt constructed successfully")
	o.logger.Debug("Stitched prompt length: %d characters", len(stitchedPrompt))

	// 6. Process each model concurrently (with rate limiting)
	var wg sync.WaitGroup
	// Create a buffered error channel to collect errors from goroutines
	errChan := make(chan error, len(o.config.ModelNames))

	// Log rate limiting configuration
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

	// Launch a goroutine for each model
	for _, name := range o.config.ModelNames {
		// Capture the loop variable to avoid data race
		modelName := name

		// Add to wait group before launching goroutine
		wg.Add(1)

		// Launch goroutine to process this model
		go func() {
			// Ensure we signal completion when goroutine exits
			defer wg.Done()

			// Acquire rate limiting permission with context
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

			// Create adapters for interfaces
			apiServiceAdapter := &APIServiceAdapter{APIService: o.apiService}

			// We no longer need the TokenManagerAdapter since ModelProcessor creates its own TokenManager
			// with the specific client. Passing nil here since it won't be used.
			processor := modelproc.NewProcessor(
				apiServiceAdapter,
				nil, // tokenManager is now created inside the Process method
				o.fileWriter,
				o.auditLogger,
				o.logger,
				o.config,
			)

			// Process the model
			err := processor.Process(ctx, modelName, stitchedPrompt)

			// If there was an error, send it to the error channel
			if err != nil {
				o.logger.Error("Processing model %s failed: %v", modelName, err)
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

// APIServiceAdapter adapts interfaces.APIService to modelproc.APIService
type APIServiceAdapter struct {
	APIService interfaces.APIService
}

func (a *APIServiceAdapter) InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
	return a.APIService.InitClient(ctx, apiKey, modelName)
}

func (a *APIServiceAdapter) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	return a.APIService.ProcessResponse(result)
}

func (a *APIServiceAdapter) IsEmptyResponseError(err error) bool {
	return a.APIService.IsEmptyResponseError(err)
}

func (a *APIServiceAdapter) IsSafetyBlockedError(err error) bool {
	return a.APIService.IsSafetyBlockedError(err)
}

func (a *APIServiceAdapter) GetErrorDetails(err error) string {
	return a.APIService.GetErrorDetails(err)
}

// TokenManagerAdapter adapts interfaces.TokenManager to modelproc.TokenManager
// TokenManagerAdapter is no longer needed since we create TokenManagers in the Process method
// The ModelProcessor now creates its own TokenManager directly with the specific client
// This adapter remains as a placeholder but is not used, and will be removed in future refactoring
