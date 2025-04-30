// Package modelproc provides model processing functionality.
package modelproc

import (
	"context"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/registry"
)

// TokenResult represents a simplified token counting result structure
// This is kept for backward compatibility with tests
type TokenResult struct {
	TokenCount   int32
	InputLimit   int32
	ExceedsLimit bool
	LimitError   string
	Percentage   float64
}

// TokenManager defines the interface for token-related operations
// This is kept for backward compatibility with tests
type TokenManager interface {
	CheckTokenLimit(ctx context.Context, prompt string) error
	GetTokenInfo(ctx context.Context, prompt string) (*TokenResult, error)
	PromptForConfirmation(tokenCount int32, threshold int) bool
}

// NewTokenManagerWithClient is a simplified stub function for backward compatibility
// Returns a no-op token manager for tests
var NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) TokenManager {
	return &noopTokenManager{}
}

// noopTokenManager provides a no-op implementation of TokenManager
type noopTokenManager struct{}

func (t *noopTokenManager) CheckTokenLimit(ctx context.Context, prompt string) error {
	return nil
}

func (t *noopTokenManager) GetTokenInfo(ctx context.Context, prompt string) (*TokenResult, error) {
	return &TokenResult{
		TokenCount:   100,
		InputLimit:   1000,
		ExceedsLimit: false,
	}, nil
}

func (t *noopTokenManager) PromptForConfirmation(tokenCount int32, threshold int) bool {
	return true
}
