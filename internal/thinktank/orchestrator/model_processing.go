// Package orchestrator provides model processing and error handling functionality.
package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
	"github.com/phrazzld/thinktank/internal/thinktank/modelproc"
)

// processModelsWithErrorHandling processes models and handles any errors that occur.
// It runs the model processing and handles error aggregation and logging.
// Returns the model outputs, any processing errors for later handling, and a critical
// error that should interrupt processing immediately.
func (o *Orchestrator) processModelsWithErrorHandling(ctx context.Context, stitchedPrompt string, contextLogger logutil.LoggerInterface) (map[string]string, error, error) {
	// Start model processing
	contextLogger.InfoContext(ctx, "Beginning model processing")

	// Calculate token metrics for processing summary
	tokenReq := interfaces.TokenCountingRequest{
		Instructions:        stitchedPrompt,
		Files:               []interfaces.FileContent{}, // Files are already included in stitchedPrompt
		SafetyMarginPercent: 20,                         // Default 20% safety margin - TODO: make this configurable via CLI
	}

	tokenResult, err := o.tokenCountingService.CountTokens(ctx, tokenReq)
	if err != nil {
		// Log error but continue with processing - don't fail due to token counting
		contextLogger.WarnContext(ctx, "Failed to calculate token metrics: %v", err)
	} else {
		// Determine accuracy method - we'll check the first model to determine the primary method
		accuracyMethod := "estimation" // Default fallback
		if len(o.config.ModelNames) > 0 {
			modelTokenResult, modelErr := o.tokenCountingService.CountTokensForModel(ctx, tokenReq, o.config.ModelNames[0])
			if modelErr == nil {
				accuracyMethod = modelTokenResult.TokenizerUsed
			}
		}

		// Log processing summary with token context
		contextLogger.InfoContext(ctx, "Analyzing %d models with %d total input tokens (accuracy: %s)",
			len(o.config.ModelNames), tokenResult.TotalTokens, accuracyMethod)

		// Check model compatibility and filter based on context window capacity
		var compatibleModels []string
		var skippedModels []string
		var allModelInfo []ModelCompatibilityInfo

		// Display token analysis header to user
		o.consoleWriter.StatusMessage(fmt.Sprintf("Token analysis: %s tokens detected", formatWithCommas(tokenResult.TotalTokens)))

		// Safety margin calculation
		safetyMarginPercent := 100.0 - float64(tokenReq.SafetyMarginPercent)

		// Collect compatibility data for all models
		for i, modelName := range o.config.ModelNames {
			modelInfo := ModelCompatibilityInfo{
				ModelName: modelName,
			}

			modelTokenResult, modelErr := o.tokenCountingService.CountTokensForModel(ctx, tokenReq, modelName)
			if modelErr == nil {
				// Get model info for context window
				modelDef, infoErr := models.GetModelInfo(modelName)
				if infoErr == nil {
					utilization := float64(modelTokenResult.TotalTokens) / float64(modelDef.ContextWindow) * 100

					modelInfo.ContextWindow = int(modelDef.ContextWindow)
					modelInfo.TokenCount = modelTokenResult.TotalTokens
					modelInfo.Utilization = utilization

					if utilization <= safetyMarginPercent {
						modelInfo.IsCompatible = true
						compatibleModels = append(compatibleModels, modelName)

						contextLogger.InfoContext(ctx, "Model %s (%d/%d) compatible - context: %d tokens, input: %d tokens, utilization: %.1f%%",
							modelName, i+1, len(o.config.ModelNames), modelDef.ContextWindow, modelTokenResult.TotalTokens, utilization)
					} else {
						modelInfo.IsCompatible = false
						modelInfo.FailureReason = "input too large"
						skippedModels = append(skippedModels, modelName)

						contextLogger.InfoContext(ctx, "Skipping model %s (%d/%d) - input too large for context window: %d tokens > %.1f%% of %d tokens",
							modelName, i+1, len(o.config.ModelNames), modelTokenResult.TotalTokens, safetyMarginPercent, modelDef.ContextWindow)
					}
				} else {
					// Can't get model info, assume incompatible for safety
					modelInfo.IsCompatible = false
					modelInfo.FailureReason = "unable to determine context window"
					skippedModels = append(skippedModels, modelName)
					contextLogger.WarnContext(ctx, "Skipping model %s (%d/%d) - unable to get model info: %v",
						modelName, i+1, len(o.config.ModelNames), infoErr)
				}
			} else {
				// Can't count tokens for this model, assume incompatible for safety
				modelInfo.IsCompatible = false
				modelInfo.FailureReason = "unable to count tokens"
				skippedModels = append(skippedModels, modelName)
				contextLogger.WarnContext(ctx, "Skipping model %s (%d/%d) - unable to count tokens: %v",
					modelName, i+1, len(o.config.ModelNames), modelErr)
			}

			allModelInfo = append(allModelInfo, modelInfo)
		}

		// Calculate statistics for display
		analysis := CompatibilityAnalysis{
			TotalModels:      len(o.config.ModelNames),
			CompatibleModels: len(compatibleModels),
			SkippedModels:    len(skippedModels),
			TotalTokens:      tokenResult.TotalTokens,
			SafetyThreshold:  safetyMarginPercent,
			CompatibleList:   compatibleModels,
			SkippedList:      skippedModels,
			AllModels:        allModelInfo,
		}

		// Find min/max utilization and best/worst models
		if len(allModelInfo) > 0 {
			analysis.MinUtilization = allModelInfo[0].Utilization
			analysis.MaxUtilization = allModelInfo[0].Utilization
			analysis.BestModel = allModelInfo[0]
			analysis.WorstModel = allModelInfo[0]

			for _, model := range allModelInfo {
				if model.ContextWindow > 0 { // Only consider models where we could get context info
					if model.Utilization < analysis.MinUtilization {
						analysis.MinUtilization = model.Utilization
						analysis.BestModel = model
					}
					if model.Utilization > analysis.MaxUtilization {
						analysis.MaxUtilization = model.Utilization
						analysis.WorstModel = model
					}
				}
			}
		}

		// Display the compatibility card
		o.displayCompatibilityCard(analysis)

		// Log summary
		if len(skippedModels) > 0 {
			contextLogger.InfoContext(ctx, "Skipped %d incompatible models: %v", len(skippedModels), skippedModels)
		}

		// Check if we have any compatible models
		if len(compatibleModels) == 0 {
			err := fmt.Errorf("no models are compatible with input size of %s tokens", formatWithCommas(tokenResult.TotalTokens))
			contextLogger.ErrorContext(ctx, err.Error())
			o.consoleWriter.StatusMessage("❌ No models are compatible with the input size")
			return nil, nil, err
		}

		// Log the models that will be processed
		contextLogger.InfoContext(ctx, "Processing %d compatible models: %v", len(compatibleModels), compatibleModels)

		// Store counts for later use in error handling and user feedback
		totalModelsRequested := len(o.config.ModelNames)
		skippedModelsCount := len(skippedModels)

		// Temporarily update config to only process compatible models
		originalModelNames := o.config.ModelNames
		o.config.ModelNames = compatibleModels
		defer func() {
			// Restore original model list after processing
			o.config.ModelNames = originalModelNames
		}()

		fmt.Printf("\nProcessing %d compatible models...\n", len(compatibleModels))

		// Store these counts in the context for later access in error handling
		type contextKey string
		ctx = context.WithValue(ctx, contextKey("totalModelsRequested"), totalModelsRequested)
		ctx = context.WithValue(ctx, contextKey("skippedModelsCount"), skippedModelsCount)
	}

	o.logRateLimitingConfiguration(ctx)
	modelOutputs, modelErrors := o.processModels(ctx, stitchedPrompt)

	// Handle model processing errors
	var returnErr error

	// Get stored counts from context (will be 0 if token counting failed)
	type contextKey string
	totalModelsRequested, _ := ctx.Value(contextKey("totalModelsRequested")).(int)
	skippedModelsCount, _ := ctx.Value(contextKey("skippedModelsCount")).(int)

	// If no context values, fall back to current config (for backwards compatibility)
	if totalModelsRequested == 0 {
		totalModelsRequested = len(o.config.ModelNames)
	}

	if len(modelErrors) > 0 {
		// If ALL attempted models failed (no outputs available), fail immediately
		if len(modelOutputs) == 0 {
			returnErr = o.aggregateErrors(modelErrors, len(o.config.ModelNames), 0)
			contextLogger.ErrorContext(ctx, returnErr.Error())

			// Provide user-facing error message for complete failure
			if skippedModelsCount > 0 {
				o.consoleWriter.StatusMessage(fmt.Sprintf("All %d compatible models failed - no outputs generated (%d models were skipped due to input size)",
					len(o.config.ModelNames), skippedModelsCount))
			} else {
				o.consoleWriter.StatusMessage("All models failed - no outputs generated")
			}

			return nil, nil, returnErr
		}

		// Otherwise, log errors but continue with available outputs
		// Get list of successful model names for the log
		var successfulModels []string
		for modelName := range modelOutputs {
			successfulModels = append(successfulModels, modelName)
		}

		// Log a warning with detailed counts and successful model names
		contextLogger.WarnContext(ctx, "Some models failed but continuing with synthesis: %d/%d attempted models successful, %d failed. Successful models: %v",
			len(modelOutputs), len(o.config.ModelNames), len(modelErrors), successfulModels)
		if skippedModelsCount > 0 {
			contextLogger.InfoContext(ctx, "Additionally, %d models were skipped due to input size being too large", skippedModelsCount)
		}

		// Provide user-facing message for partial failures
		if skippedModelsCount > 0 {
			o.consoleWriter.StatusMessage(fmt.Sprintf("● %d/%d models succeeded, continuing with available outputs",
				len(modelOutputs), totalModelsRequested))
		} else {
			o.consoleWriter.StatusMessage(fmt.Sprintf("● %d/%d models succeeded, continuing with available outputs",
				len(modelOutputs), len(o.config.ModelNames)))
		}

		// Log individual error details
		for _, err := range modelErrors {
			contextLogger.ErrorContext(ctx, "%v", err)
		}

		// Create a descriptive error to return after processing is complete
		returnErr = o.aggregateErrors(modelErrors, len(o.config.ModelNames), len(modelOutputs))
	} else if skippedModelsCount > 0 {
		// All attempted models succeeded, but some were skipped
		contextLogger.InfoContext(ctx, "All %d attempted models succeeded (%d models were skipped due to input size)",
			len(modelOutputs), skippedModelsCount)
	}

	return modelOutputs, returnErr, nil
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

	// Start progress tracking
	fmt.Println() // Extra space before processing starts
	o.consoleWriter.StartProcessing(len(o.config.ModelNames))

	// Launch a goroutine for each model, passing the index for progress tracking
	for i, modelName := range o.config.ModelNames {
		wg.Add(1)
		// Pass 1-based index for user-friendly display
		go o.processModelWithRateLimit(ctx, modelName, stitchedPrompt, i+1, &wg, resultChan)
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
	index int,
	wg *sync.WaitGroup,
	resultChan chan<- modelResult,
) {
	defer wg.Done()

	// Get logger with context
	contextLogger := o.logger.WithContext(ctx)

	// Create a local variable to store the result to avoid accessing it from multiple goroutines
	var result modelResult
	result.modelName = modelName

	// Get the appropriate rate limiter for this model (model-specific or global)
	rateLimiter := o.getRateLimiterForModel(modelName)

	// Acquire rate limiting permission
	contextLogger.DebugContext(ctx, "Attempting to acquire rate limiter for model %s...", modelName)
	acquireStart := time.Now()
	if err := rateLimiter.Acquire(ctx, modelName); err != nil {
		contextLogger.ErrorContext(ctx, "Rate limiting error for model %s: %v", modelName, err)
		result.err = llm.Wrap(err, "orchestrator",
			fmt.Sprintf("failed to acquire rate limiter for model %s", modelName),
			llm.CategoryRateLimit)
		resultChan <- result
		return
	}
	acquireDuration := time.Since(acquireStart)
	contextLogger.DebugContext(ctx, "Rate limiter acquired for model %s (waited %v)", modelName, acquireDuration)

	// Report rate limiting delay if significant
	if acquireDuration > 100*time.Millisecond {
		o.consoleWriter.ModelRateLimited(index, len(o.config.ModelNames), modelName, acquireDuration)
	}

	// Release rate limiter when done
	defer func() {
		contextLogger.DebugContext(ctx, "Releasing rate limiter for model %s", modelName)
		rateLimiter.Release()
	}()

	// Report model processing started
	o.consoleWriter.ModelStarted(index, len(o.config.ModelNames), modelName)

	// Create API service adapter and model processor
	apiServiceAdapter := &APIServiceAdapter{APIService: o.apiService}
	processor := modelproc.NewProcessor(
		apiServiceAdapter,
		o.fileWriter,
		o.auditLogger,
		o.logger,
		o.config,
	)

	// Process the model and track timing
	processingStart := time.Now()
	content, err := processor.Process(ctx, modelName, stitchedPrompt)
	processingDuration := time.Since(processingStart)
	if err != nil {
		contextLogger.ErrorContext(ctx, "Processing model %s failed: %v", modelName, err)

		// Preserve the detailed error instead of wrapping with generic message
		result.err = err

		// Show user-friendly error with suggestions if it's an LLMError
		var errorMessage string
		if llmErr, ok := err.(*llm.LLMError); ok {
			// Create enhanced error message with suggestions for certain error types
			switch llmErr.Category() {
			case llm.CategoryAuth:
				errorMessage = fmt.Sprintf("authentication failed\nTry: export %s=your_api_key", getAPIKeyEnvVarName(modelName))
			case llm.CategoryRateLimit:
				errorMessage = "rate limited\nTry: reduce concurrency with --max-concurrent=1"
			case llm.CategoryInvalidRequest:
				if strings.Contains(strings.ToLower(llmErr.Error()), "context") ||
					strings.Contains(strings.ToLower(llmErr.Error()), "token") ||
					strings.Contains(strings.ToLower(llmErr.Error()), "length") {
					errorMessage = "input too large\nTry: --exclude \"docs/,*.md\" or focus on specific files"
				} else {
					errorMessage = "invalid request\nTry: check model name or reduce input size"
				}
			case llm.CategoryInsufficientCredits:
				errorMessage = "insufficient credits\nTry: check your account balance or use a different model"
			default:
				errorMessage = llmErr.UserFacingError()
			}
		} else {
			// Generic error message for non-LLM errors
			errorMessage = err.Error()
		}

		// Report failed model to user
		o.consoleWriter.ModelFailed(index, len(o.config.ModelNames), modelName, errorMessage)

		// Send result to channel
		resultChan <- result
		return
	}

	// Log success
	contextLogger.DebugContext(ctx, "Processing model %s completed successfully in %v", modelName, processingDuration)

	// Store content
	result.content = content

	// Report success
	o.consoleWriter.ModelCompleted(index, len(o.config.ModelNames), modelName, processingDuration)

	// Send result to channel
	resultChan <- result
}

// getAPIKeyEnvVarName returns the environment variable name for the API key for a given model
func getAPIKeyEnvVarName(modelName string) string {
	modelInfo, err := models.GetModelInfo(modelName)
	if err != nil {
		return "API_KEY" // Generic fallback
	}

	switch modelInfo.Provider {
	case "openai":
		return "OPENAI_API_KEY"
	case "gemini":
		return "GEMINI_API_KEY"
	case "openrouter":
		return "OPENROUTER_API_KEY"
	default:
		return "API_KEY"
	}
}
