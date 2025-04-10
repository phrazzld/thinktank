// Package architect provides the command-line interface for the architect tool
package architect

import (
	"context"
	"fmt"
	"os"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// TokenResult holds information about token counts and limits
type TokenResult struct {
	TokenCount   int32
	InputLimit   int32
	ExceedsLimit bool
	LimitError   string
	Percentage   float64
}

// TokenManager defines the interface for token counting and management
type TokenManager interface {
	// GetTokenInfo retrieves token count information and checks limits
	GetTokenInfo(ctx context.Context, client gemini.Client, prompt string) (*TokenResult, error)

	// CheckTokenLimit verifies the prompt doesn't exceed the model's token limit
	CheckTokenLimit(ctx context.Context, client gemini.Client, prompt string) error

	// PromptForConfirmation asks for user confirmation to proceed if token count exceeds threshold
	PromptForConfirmation(tokenCount int32, threshold int) bool
}

// tokenManager implements the TokenManager interface
type tokenManager struct {
	logger logutil.LoggerInterface
}

// NewTokenManager creates a new TokenManager instance
func NewTokenManager(logger logutil.LoggerInterface) TokenManager {
	return &tokenManager{
		logger: logger,
	}
}

// GetTokenInfo retrieves token count information and checks limits
func (tm *tokenManager) GetTokenInfo(ctx context.Context, client gemini.Client, prompt string) (*TokenResult, error) {
	// Stub implementation - will be replaced with actual code from main.go
	return nil, fmt.Errorf("not implemented yet")
}

// CheckTokenLimit verifies the prompt doesn't exceed the model's token limit
func (tm *tokenManager) CheckTokenLimit(ctx context.Context, client gemini.Client, prompt string) error {
	// Stub implementation - will be replaced with actual code from main.go
	return fmt.Errorf("not implemented yet")
}

// PromptForConfirmation asks for user confirmation to proceed
func (tm *tokenManager) PromptForConfirmation(tokenCount int32, threshold int) bool {
	// Stub implementation - will be replaced with actual code from main.go
	tm.logger.Info("Token count (%d) exceeds confirmation threshold (%d).", tokenCount, threshold)
	tm.logger.Info("Do you want to proceed with the API call? [y/N]: ")

	// Return false by default for the stub implementation
	return false
}
