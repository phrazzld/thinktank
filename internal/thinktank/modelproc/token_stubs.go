// Package modelproc provides model processing functionality for the thinktank tool.
package modelproc

import (
	"context"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/registry"
)

// This file formerly contained token-related stubs that are no longer needed
// after the token handling removal refactoring (T036D).
//
// Kept as minimal placeholders to maintain imports in existing tests.
// This file should be removed in a future cleanup once all dependent tests
// are updated or disabled.
//
// NOTE: T036C has already disabled most token-related tests, but some tests
// still have references to token types. A future task may remove these
// references completely.

// TokenResult is a stub structure retained for test compatibility only
type TokenResult struct {
	TokenCount   int32
	InputLimit   int32
	ExceedsLimit bool
	LimitError   string
	Percentage   float64
}

// TokenManager is a stub interface retained for test compatibility only
type TokenManager interface {
	GetTokenInfo(ctx context.Context, prompt string) (*TokenResult, error)
	CheckTokenLimit(ctx context.Context, prompt string) error
	PromptForConfirmation(tokenCount int32, threshold int) bool
}

// NewTokenManagerWithClient is a stub function retained for test compatibility only
var NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) TokenManager {
	return &noOpTokenManager{
		logger:      logger,
		auditLogger: auditLogger,
		client:      client,
	}
}

// noOpTokenManager is a no-op implementation of TokenManager
type noOpTokenManager struct {
	logger      logutil.LoggerInterface
	auditLogger auditlog.AuditLogger
	client      llm.LLMClient
}

func (m *noOpTokenManager) GetTokenInfo(ctx context.Context, prompt string) (*TokenResult, error) {
	return &TokenResult{
		TokenCount:   100,
		InputLimit:   1000,
		ExceedsLimit: false,
		Percentage:   10.0,
	}, nil
}

func (m *noOpTokenManager) CheckTokenLimit(ctx context.Context, prompt string) error {
	return nil
}

func (m *noOpTokenManager) PromptForConfirmation(tokenCount int32, threshold int) bool {
	return true
}
