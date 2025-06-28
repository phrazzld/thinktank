// Package thinktank provides token counting services for accurate LLM context management.
// This service replaces estimation-based token counting with precise calculations,
// enabling intelligent model selection based on actual context requirements.
package thinktank

import (
	"context"

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

// tokenCountingServiceImpl implements the TokenCountingService interface.
type tokenCountingServiceImpl struct {
	tokenizerManager tokenizers.TokenizerManager
	// Future: dependencies will be injected here (logger, audit logger, etc.)
}

// NewTokenCountingService creates a new token counting service instance.
// Uses constructor pattern for dependency injection.
func NewTokenCountingService() TokenCountingService {
	return &tokenCountingServiceImpl{
		tokenizerManager: tokenizers.NewTokenizerManager(),
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
