// Package thinktank provides token counting services for accurate LLM context management.
// This service replaces estimation-based token counting with precise calculations,
// enabling intelligent model selection based on actual context requirements.
package thinktank

import (
	"context"
	"fmt"

	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
	"github.com/phrazzld/thinktank/internal/thinktank/tokenizers"
)

// TokenCountingService provides accurate token counting and model filtering capabilities.
// It replaces estimation-based model selection with precise token counting from
// instructions and file content, enabling better model compatibility decisions.
type TokenCountingService interface {
	// CountTokens calculates total tokens for instructions and file content.
	// Returns detailed breakdown of token usage for audit and logging purposes.
	CountTokens(ctx context.Context, req TokenCountingRequest) (TokenCountingResult, error)

	// CountTokensForModel calculates tokens using accurate tokenization for the specific model.
	// Falls back to estimation if accurate tokenization is not available for the model.
	CountTokensForModel(ctx context.Context, req TokenCountingRequest, modelName string) (ModelTokenCountingResult, error)

	// GetCompatibleModels returns models that can handle the input with detailed compatibility information.
	// Uses accurate token counting to determine which models can process the given request.
	GetCompatibleModels(ctx context.Context, req TokenCountingRequest, availableProviders []string) ([]ModelCompatibility, error)
}

// TokenCountingRequest contains all inputs needed for token counting.
type TokenCountingRequest struct {
	// Instructions text to be sent to the model
	Instructions string

	// Files contains all file content to be processed
	Files []FileContent
}

// FileContent represents a single file's content for token counting.
type FileContent struct {
	// Path is the file path for identification and logging
	Path string

	// Content is the actual file content to count tokens for
	Content string
}

// TokenCountingResult provides detailed breakdown of token usage.
type TokenCountingResult struct {
	// TotalTokens is the sum of all token counts
	TotalTokens int

	// InstructionTokens is tokens used by instructions
	InstructionTokens int

	// FileTokens is tokens used by file content
	FileTokens int

	// Overhead includes formatting and structural tokens
	Overhead int
}

// ModelTokenCountingResult provides model-specific token counting results.
type ModelTokenCountingResult struct {
	TokenCountingResult

	// ModelName is the model this count was calculated for
	ModelName string

	// TokenizerUsed indicates which tokenization method was used
	TokenizerUsed string // "tiktoken", "sentencepiece", "estimation"

	// Provider is the LLM provider for the model
	Provider string

	// IsAccurate indicates if accurate tokenization was used (vs estimation)
	IsAccurate bool
}

// ModelCompatibility provides detailed compatibility information for a model.
type ModelCompatibility struct {
	// ModelName is the name of the model being evaluated
	ModelName string

	// IsCompatible indicates if the model can handle the input
	IsCompatible bool

	// TokenCount is the actual token count for this model
	TokenCount int

	// ContextWindow is the model's maximum context size
	ContextWindow int

	// UsableContext is the context available after safety margin
	UsableContext int

	// Provider is the LLM provider for this model
	Provider string

	// TokenizerUsed indicates which tokenization method was used
	TokenizerUsed string

	// IsAccurate indicates if accurate tokenization was used
	IsAccurate bool

	// Reason explains why the model is incompatible (if applicable)
	Reason string
}

// tokenCountingServiceImpl implements the TokenCountingService interface.
type tokenCountingServiceImpl struct {
	tokenizerManager tokenizers.TokenizerManager
	logger           logutil.LoggerInterface
}

// NewTokenCountingService creates a new token counting service instance.
// Uses constructor pattern for dependency injection.
func NewTokenCountingService() TokenCountingService {
	return &tokenCountingServiceImpl{
		tokenizerManager: tokenizers.NewTokenizerManager(),
		logger:           nil, // No logging for default constructor
	}
}

// NewTokenCountingServiceWithLogger creates a new service instance with logging.
// Uses dependency injection pattern for testability.
func NewTokenCountingServiceWithLogger(logger logutil.LoggerInterface) TokenCountingService {
	return &tokenCountingServiceImpl{
		tokenizerManager: tokenizers.NewTokenizerManager(),
		logger:           logger,
	}
}

// NewTokenCountingServiceWithManager creates a service with custom tokenizer manager.
// Useful for testing with mocked dependencies.
func NewTokenCountingServiceWithManager(manager tokenizers.TokenizerManager) TokenCountingService {
	return &tokenCountingServiceImpl{
		tokenizerManager: manager,
	}
}

// CountTokens implements the TokenCountingService interface.
// Uses existing token estimation from models package for consistency.
func (s *tokenCountingServiceImpl) CountTokens(ctx context.Context, req TokenCountingRequest) (TokenCountingResult, error) {
	// Handle empty context case
	if req.Instructions == "" && len(req.Files) == 0 {
		return TokenCountingResult{
			TotalTokens: 0,
		}, nil
	}

	// Calculate token breakdown
	instructionTokens := s.countInstructionTokens(req.Instructions)
	fileTokens := s.countFileTokens(req.Files)
	overhead := s.calculateOverhead()

	result := TokenCountingResult{
		InstructionTokens: instructionTokens,
		FileTokens:        fileTokens,
		Overhead:          overhead,
		TotalTokens:       instructionTokens + fileTokens + overhead,
	}

	return result, nil
}

// CountTokensForModel implements model-specific accurate token counting.
func (s *tokenCountingServiceImpl) CountTokensForModel(ctx context.Context, req TokenCountingRequest, modelName string) (ModelTokenCountingResult, error) {
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
				instructionTokens = s.countInstructionTokens(req.Instructions)
				tokenizerUsed = "estimation"
				isAccurate = false
			} else {
				fileTokens, err = s.countFileTokensAccurate(ctx, req.Files, modelName, tokenizer)
				if err != nil {
					// Fall back to estimation on error
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
			// Fall back to estimation if model not supported
			instructionTokens = s.countInstructionTokens(req.Instructions)
			fileTokens = s.countFileTokens(req.Files)
			tokenizerUsed = "estimation"
			isAccurate = false
		}
	} else {
		// Fall back to estimation if provider not supported
		instructionTokens = s.countInstructionTokens(req.Instructions)
		fileTokens = s.countFileTokens(req.Files)
		tokenizerUsed = "estimation"
		isAccurate = false
	}

	// Handle empty context case
	if req.Instructions == "" && len(req.Files) == 0 {
		return ModelTokenCountingResult{
			TokenCountingResult: TokenCountingResult{TotalTokens: 0},
			ModelName:           modelName,
			TokenizerUsed:       tokenizerUsed,
			Provider:            modelInfo.Provider,
			IsAccurate:          isAccurate,
		}, nil
	}

	overhead := s.calculateOverhead()
	totalTokens := instructionTokens + fileTokens + overhead

	result := ModelTokenCountingResult{
		TokenCountingResult: TokenCountingResult{
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
func (s *tokenCountingServiceImpl) countFileTokens(files []FileContent) int {
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
func (s *tokenCountingServiceImpl) countFileTokensAccurate(ctx context.Context, files []FileContent, modelName string, tokenizer tokenizers.AccurateTokenCounter) (int, error) {
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
func (s *tokenCountingServiceImpl) GetCompatibleModels(ctx context.Context, req TokenCountingRequest, availableProviders []string) ([]ModelCompatibility, error) {
	// Add enhanced logging with context information per TODO.md requirements
	if s.logger != nil {
		s.logger.InfoContext(ctx, "Starting model compatibility check",
			"provider_count", len(availableProviders),
			"file_count", len(req.Files),
			"has_instructions", req.Instructions != "")
	}

	// Handle empty input case - return empty result
	if len(availableProviders) == 0 {
		return []ModelCompatibility{}, nil
	}

	// If empty request (no instructions, no files), return empty result
	if req.Instructions == "" && len(req.Files) == 0 {
		return []ModelCompatibility{}, nil
	}

	var results []ModelCompatibility

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
func (s *tokenCountingServiceImpl) checkModelCompatibility(ctx context.Context, req TokenCountingRequest, modelName string) (ModelCompatibility, error) {
	// Get model info
	modelInfo, err := models.GetModelInfo(modelName)
	if err != nil {
		return ModelCompatibility{}, fmt.Errorf("failed to get model info for %s: %w", modelName, err)
	}

	// Get accurate token count for this model
	tokenResult, err := s.CountTokensForModel(ctx, req, modelName)
	if err != nil {
		return ModelCompatibility{}, fmt.Errorf("failed to count tokens for %s: %w", modelName, err)
	}

	// Calculate safety margin (20% for output buffer)
	safetyMargin := int(float64(modelInfo.ContextWindow) * 0.2)
	usableContext := modelInfo.ContextWindow - safetyMargin

	// Determine compatibility
	isCompatible := tokenResult.TotalTokens <= usableContext
	reason := ""
	if !isCompatible {
		reason = fmt.Sprintf("requires %d tokens but model only has %d usable tokens (%d total - %d safety margin)",
			tokenResult.TotalTokens, usableContext, modelInfo.ContextWindow, safetyMargin)
	}

	return ModelCompatibility{
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
func sortModelCompatibility(results []ModelCompatibility) {
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
func shouldSwap(a, b ModelCompatibility) bool {
	// Compatible models come first
	if a.IsCompatible != b.IsCompatible {
		return !a.IsCompatible // a comes after b if a is not compatible but b is
	}
	// Within same compatibility, larger context windows come first
	return a.ContextWindow < b.ContextWindow
}
