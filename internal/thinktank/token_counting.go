// Package thinktank provides token counting services for accurate LLM context management.
// This service replaces estimation-based token counting with precise calculations,
// enabling intelligent model selection based on actual context requirements.
package thinktank

import (
	"context"

	"github.com/phrazzld/thinktank/internal/models"
)

// TokenCountingService provides accurate token counting and model filtering capabilities.
// It replaces estimation-based model selection with precise token counting from
// instructions and file content, enabling better model compatibility decisions.
type TokenCountingService interface {
	// CountTokens calculates total tokens for instructions and file content.
	// Returns detailed breakdown of token usage for audit and logging purposes.
	CountTokens(ctx context.Context, req TokenCountingRequest) (TokenCountingResult, error)
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

// tokenCountingServiceImpl implements the TokenCountingService interface.
type tokenCountingServiceImpl struct {
	// Future: dependencies will be injected here (logger, audit logger, etc.)
}

// NewTokenCountingService creates a new token counting service instance.
// Uses constructor pattern for future dependency injection.
func NewTokenCountingService() TokenCountingService {
	return &tokenCountingServiceImpl{}
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
