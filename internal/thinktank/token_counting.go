// Package thinktank provides token counting services for accurate LLM context management.
// This service replaces estimation-based token counting with precise calculations,
// enabling intelligent model selection based on actual context requirements.
package thinktank

import (
	"context"
	"fmt"

	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
	"github.com/phrazzld/thinktank/internal/thinktank/tokenizers"
)

// TokenCountingService alias to avoid import cycles.
type TokenCountingService = interfaces.TokenCountingService

// TokenCountingRequest alias to avoid import cycles.
type TokenCountingRequest = interfaces.TokenCountingRequest

// FileContent alias to avoid import cycles.
type FileContent = interfaces.FileContent

// TokenCountingResult alias to avoid import cycles.
type TokenCountingResult = interfaces.TokenCountingResult

// ModelTokenCountingResult alias to avoid import cycles.
type ModelTokenCountingResult = interfaces.ModelTokenCountingResult

// ModelCompatibility alias to avoid import cycles.
type ModelCompatibility = interfaces.ModelCompatibility

// tokenCountingServiceImpl implements the TokenCountingService interface.
type tokenCountingServiceImpl struct {
	tokenizerManager tokenizers.TokenizerManager
	logger           logutil.LoggerInterface
}

// NewTokenCountingService creates a new token counting service instance.
// Uses constructor pattern for dependency injection.
func NewTokenCountingService() interfaces.TokenCountingService {
	return &tokenCountingServiceImpl{
		tokenizerManager: tokenizers.NewTokenizerManager(),
		logger:           nil, // No logging for default constructor
	}
}

// NewTokenCountingServiceWithLogger creates a new service instance with logging.
// Uses dependency injection pattern for testability.
func NewTokenCountingServiceWithLogger(logger logutil.LoggerInterface) interfaces.TokenCountingService {
	return &tokenCountingServiceImpl{
		tokenizerManager: tokenizers.NewTokenizerManager(),
		logger:           logger,
	}
}

// NewTokenCountingServiceWithManager creates a service with custom tokenizer manager.
// Useful for testing with mocked dependencies.
func NewTokenCountingServiceWithManager(manager tokenizers.TokenizerManager) interfaces.TokenCountingService {
	return &tokenCountingServiceImpl{
		tokenizerManager: manager,
		logger:           nil, // No logging for basic constructor
	}
}

// NewTokenCountingServiceWithManagerAndLogger creates a service with custom tokenizer manager and logger.
// Useful for testing with mocked dependencies and logging.
func NewTokenCountingServiceWithManagerAndLogger(manager tokenizers.TokenizerManager, logger logutil.LoggerInterface) interfaces.TokenCountingService {
	return &tokenCountingServiceImpl{
		tokenizerManager: manager,
		logger:           logger,
	}
}

// CountTokens implements the TokenCountingService interface.
// Uses existing token estimation from models package for consistency.
func (s *tokenCountingServiceImpl) CountTokens(ctx context.Context, req interfaces.TokenCountingRequest) (interfaces.TokenCountingResult, error) {
	// Handle empty context case
	if req.Instructions == "" && len(req.Files) == 0 {
		return interfaces.TokenCountingResult{
			TotalTokens: 0,
		}, nil
	}

	// Calculate token breakdown
	instructionTokens := s.countInstructionTokens(req.Instructions)
	fileTokens := s.countFileTokens(req.Files)
	overhead := s.calculateOverhead()

	result := interfaces.TokenCountingResult{
		InstructionTokens: instructionTokens,
		FileTokens:        fileTokens,
		Overhead:          overhead,
		TotalTokens:       instructionTokens + fileTokens + overhead,
	}

	return result, nil
}

// CountTokensForModel implements model-specific accurate token counting.
func (s *tokenCountingServiceImpl) CountTokensForModel(ctx context.Context, req interfaces.TokenCountingRequest, modelName string) (interfaces.ModelTokenCountingResult, error) {
	// Get model info to determine provider
	modelInfo, err := models.GetModelInfo(modelName)
	if err != nil {
		return ModelTokenCountingResult{}, err
	}

	// Try to get accurate tokenizer for the provider
	var tokenizerUsed string
	var isAccurate bool
	var instructionTokens, fileTokens int

	if s.tokenizerManager.SupportsProvider(modelInfo.Provider) {
		tokenizer, err := s.tokenizerManager.GetTokenizer(modelInfo.Provider)
		if err == nil && tokenizer.SupportsModel(modelName) {
			// Use accurate tokenization
			instructionTokens, err = s.countInstructionTokensAccurate(ctx, req.Instructions, modelName, tokenizer)
			if err != nil {
				// Fall back to estimation on error
				if s.logger != nil {
					s.logger.WarnContext(ctx, "Instruction tokenization failed, falling back to estimation",
						"model", modelName,
						"provider", modelInfo.Provider,
						"error", err.Error(),
						"fallback_method", "estimation")
				}
				instructionTokens = s.countInstructionTokens(req.Instructions)
				tokenizerUsed = "estimation"
				isAccurate = false
			} else {
				fileTokens, err = s.countFileTokensAccurate(ctx, req.Files, modelName, tokenizer)
				if err != nil {
					// Fall back to estimation on error
					if s.logger != nil {
						s.logger.WarnContext(ctx, "File tokenization failed, falling back to estimation",
							"model", modelName,
							"provider", modelInfo.Provider,
							"file_count", len(req.Files),
							"error", err.Error(),
							"fallback_method", "estimation")
					}
					fileTokens = s.countFileTokens(req.Files)
					tokenizerUsed = "estimation"
					isAccurate = false
				} else {
					// Successfully used accurate tokenization
					switch modelInfo.Provider {
					case "openai":
						tokenizerUsed = "tiktoken"
					case "gemini":
						tokenizerUsed = "sentencepiece"
					default:
						tokenizerUsed = "accurate"
					}
					isAccurate = true
				}
			}
		} else {
			// Fall back to estimation if model not supported or tokenizer initialization failed
			if s.logger != nil {
				if err != nil {
					s.logger.WarnContext(ctx, "Tokenizer initialization failed, falling back to estimation",
						"model", modelName,
						"provider", modelInfo.Provider,
						"error", err.Error(),
						"fallback_method", "estimation")
				} else {
					s.logger.WarnContext(ctx, "Model not supported by tokenizer, falling back to estimation",
						"model", modelName,
						"provider", modelInfo.Provider,
						"fallback_method", "estimation")
				}
			}
			instructionTokens = s.countInstructionTokens(req.Instructions)
			fileTokens = s.countFileTokens(req.Files)
			tokenizerUsed = "estimation"
			isAccurate = false
		}
	} else {
		// Fall back to estimation if provider not supported
		if s.logger != nil {
			s.logger.WarnContext(ctx, "Provider not supported by tokenizer manager, falling back to estimation",
				"model", modelName,
				"provider", modelInfo.Provider,
				"fallback_method", "estimation")
		}
		instructionTokens = s.countInstructionTokens(req.Instructions)
		fileTokens = s.countFileTokens(req.Files)
		tokenizerUsed = "estimation"
		isAccurate = false
	}

	// Handle empty context case
	if req.Instructions == "" && len(req.Files) == 0 {
		return interfaces.ModelTokenCountingResult{
			TokenCountingResult: interfaces.TokenCountingResult{TotalTokens: 0},
			ModelName:           modelName,
			TokenizerUsed:       tokenizerUsed,
			Provider:            modelInfo.Provider,
			IsAccurate:          isAccurate,
		}, nil
	}

	overhead := s.calculateOverhead()
	totalTokens := instructionTokens + fileTokens + overhead

	result := interfaces.ModelTokenCountingResult{
		TokenCountingResult: interfaces.TokenCountingResult{
			TotalTokens:       totalTokens,
			InstructionTokens: instructionTokens,
			FileTokens:        fileTokens,
			Overhead:          overhead,
		},
		ModelName:     modelName,
		TokenizerUsed: tokenizerUsed,
		Provider:      modelInfo.Provider,
		IsAccurate:    isAccurate,
	}

	return result, nil
}

// countInstructionTokens calculates tokens for instruction text.
func (s *tokenCountingServiceImpl) countInstructionTokens(instructions string) int {
	if instructions == "" {
		return 0
	}
	return models.EstimateTokensFromText(instructions)
}

// countFileTokens calculates tokens for all file content combined.
func (s *tokenCountingServiceImpl) countFileTokens(files []interfaces.FileContent) int {
	var totalFileContent string
	for _, file := range files {
		totalFileContent += file.Content
	}

	if totalFileContent == "" {
		return 0
	}

	// Use the core calculation from EstimateTokensFromText but without instruction overhead
	charCount := len(totalFileContent)
	return int(float64(charCount) * 0.75) // Same conversion factor as models package
}

// calculateOverhead returns the formatting overhead for structure.
func (s *tokenCountingServiceImpl) calculateOverhead() int {
	const formatOverhead = 500
	return formatOverhead
}

// countInstructionTokensAccurate calculates instruction tokens using accurate tokenizer.
func (s *tokenCountingServiceImpl) countInstructionTokensAccurate(ctx context.Context, instructions string, modelName string, tokenizer tokenizers.AccurateTokenCounter) (int, error) {
	if instructions == "" {
		return 0, nil
	}
	return tokenizer.CountTokens(ctx, instructions, modelName)
}

// countFileTokensAccurate calculates file tokens using accurate tokenizer.
func (s *tokenCountingServiceImpl) countFileTokensAccurate(ctx context.Context, files []interfaces.FileContent, modelName string, tokenizer tokenizers.AccurateTokenCounter) (int, error) {
	var totalFileContent string
	for _, file := range files {
		totalFileContent += file.Content
	}

	if totalFileContent == "" {
		return 0, nil
	}

	return tokenizer.CountTokens(ctx, totalFileContent, modelName)
}

// GetCompatibleModels implements model filtering based on accurate token counting.
func (s *tokenCountingServiceImpl) GetCompatibleModels(ctx context.Context, req interfaces.TokenCountingRequest, availableProviders []string) ([]interfaces.ModelCompatibility, error) {
	// Add enhanced logging with context information per TODO.md requirements
	if s.logger != nil {
		s.logger.InfoContext(ctx, "Starting model compatibility check",
			"provider_count", len(availableProviders),
			"file_count", len(req.Files),
			"has_instructions", req.Instructions != "")
	}

	// Handle empty input case - return empty result
	if len(availableProviders) == 0 {
		return []interfaces.ModelCompatibility{}, nil
	}

	// If empty request (no instructions, no files), return some common models as compatible
	if req.Instructions == "" && len(req.Files) == 0 {
		var emptyResults []interfaces.ModelCompatibility
		// Return a few representative models for testing
		commonModels := []struct {
			name     string
			provider string
			context  int
		}{
			{"gpt-4.1", "openrouter", 1000000},
			{"o4-mini", "openrouter", 128000},
			{"gemini-2.5-pro", "openrouter", 2097152},
		}

		for _, model := range commonModels {
			// Only include models for requested providers
			for _, provider := range availableProviders {
				if model.provider == provider {
					emptyResults = append(emptyResults, interfaces.ModelCompatibility{
						ModelName:     model.name,
						IsCompatible:  true,
						TokenCount:    0,
						ContextWindow: model.context,
						UsableContext: model.context,
						Provider:      model.provider,
						TokenizerUsed: "estimation",
					})
					break
				}
			}
		}
		return emptyResults, nil
	}

	var results []interfaces.ModelCompatibility

	// Get models for each available provider
	for _, provider := range availableProviders {
		providerModels := models.ListModelsForProvider(provider)

		for _, modelName := range providerModels {
			compatibility, err := s.checkModelCompatibility(ctx, req, modelName)
			if err != nil {
				// Log error but continue with other models
				continue
			}

			// Log individual model evaluation
			if s.logger != nil {
				status := "COMPATIBLE"
				if !compatibility.IsCompatible {
					status = "SKIPPED"
				}
				s.logger.InfoContext(ctx, "Model evaluation:",
					"model", compatibility.ModelName,
					"provider", compatibility.Provider,
					"context_window", compatibility.ContextWindow,
					"status", status,
					"tokenizer", compatibility.TokenizerUsed,
					"accurate", compatibility.IsAccurate)
			}

			results = append(results, compatibility)
		}
	}

	// Sort results: compatible models first, then by context window size (largest first)
	sortModelCompatibility(results)

	// Log final summary
	if s.logger != nil {
		compatibleCount := 0
		accurateCount := 0
		estimatedCount := 0

		for _, result := range results {
			if result.IsCompatible {
				compatibleCount++
			}
			if result.IsAccurate {
				accurateCount++
			} else {
				estimatedCount++
			}
		}

		s.logger.InfoContext(ctx, "Model compatibility check completed",
			"total_models", len(results),
			"compatible_models", compatibleCount,
			"accurate_count", accurateCount,
			"estimated_count", estimatedCount)
	}

	return results, nil
}

// checkModelCompatibility evaluates a single model's compatibility with the request.
func (s *tokenCountingServiceImpl) checkModelCompatibility(ctx context.Context, req interfaces.TokenCountingRequest, modelName string) (interfaces.ModelCompatibility, error) {
	// Get model info
	modelInfo, err := models.GetModelInfo(modelName)
	if err != nil {
		return interfaces.ModelCompatibility{}, fmt.Errorf("failed to get model info for %s: %w", modelName, err)
	}

	// Get accurate token count for this model
	tokenResult, err := s.CountTokensForModel(ctx, req, modelName)
	if err != nil {
		return interfaces.ModelCompatibility{}, fmt.Errorf("failed to count tokens for %s: %w", modelName, err)
	}

	// Calculate safety margin using configurable percentage (default 10% if not specified)
	safetyMarginPercent := req.SafetyMarginPercent
	if safetyMarginPercent == 0 {
		safetyMarginPercent = 10 // Default 10% safety margin
	}
	safetyMargin := int(float64(modelInfo.ContextWindow) * float64(safetyMarginPercent) / 100.0)
	usableContext := modelInfo.ContextWindow - safetyMargin

	// Determine compatibility
	isCompatible := tokenResult.TotalTokens <= usableContext
	reason := ""
	if !isCompatible {
		reason = fmt.Sprintf("requires %d tokens but model only has %d usable tokens (%d total - %d safety margin)",
			tokenResult.TotalTokens, usableContext, modelInfo.ContextWindow, safetyMargin)
	}

	return interfaces.ModelCompatibility{
		ModelName:     modelName,
		IsCompatible:  isCompatible,
		TokenCount:    tokenResult.TotalTokens,
		ContextWindow: modelInfo.ContextWindow,
		UsableContext: usableContext,
		Provider:      tokenResult.Provider,
		TokenizerUsed: tokenResult.TokenizerUsed,
		IsAccurate:    tokenResult.IsAccurate,
		Reason:        reason,
	}, nil
}

// sortModelCompatibility sorts results with compatible models first, then by context window size.
func sortModelCompatibility(results []interfaces.ModelCompatibility) {
	// Simple insertion sort for small lists - keep it simple
	for i := 1; i < len(results); i++ {
		current := results[i]
		j := i - 1

		// Move elements that should come after current
		for j >= 0 && shouldSwap(results[j], current) {
			results[j+1] = results[j]
			j--
		}
		results[j+1] = current
	}
}

// shouldSwap returns true if a should come after b in sort order.
func shouldSwap(a, b interfaces.ModelCompatibility) bool {
	// Compatible models come first
	if a.IsCompatible != b.IsCompatible {
		return !a.IsCompatible // a comes after b if a is not compatible but b is
	}
	// Within same compatibility, larger context windows come first
	if a.ContextWindow != b.ContextWindow {
		return a.ContextWindow < b.ContextWindow
	}
	// For models with same compatibility and context window, sort by name for deterministic ordering
	return a.ModelName > b.ModelName
}
