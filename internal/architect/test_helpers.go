// Package architect contains reusable test helpers for testing the architect package.
// This file is explicitly for test use only.
package architect

import (
	"context"

	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/llm"
)

// These test helpers are used in various test files in the architect package.
// nolint:unused // These are used in tests, but the linter doesn't detect it.

// mockAuditLogger implements auditlog.AuditLogger for testing
type mockAuditLogger struct {
	entries   []auditlog.AuditEntry
	lastEntry *auditlog.AuditEntry
}

func (m *mockAuditLogger) Log(entry auditlog.AuditEntry) error {
	m.entries = append(m.entries, entry)
	m.lastEntry = &entry
	return nil
}

func (m *mockAuditLogger) Close() error {
	return nil
}

// SetConfig not needed for the mockAuditLogger

func (m *mockAuditLogger) Last() *auditlog.AuditEntry {
	return m.lastEntry
}

// Note: TokenResultAdapter was removed as part of T032A
// to remove token handling from the application.

// mockLLMClient implements llm.LLMClient for testing
type mockLLMClient struct {
	modelName           string
	tokenCount          int32
	inputTokenLimit     int32
	outputTokenLimit    int32
	generateContentErr  error
	countTokensErr      error
	getModelInfoErr     error
	countTokensFunc     func(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error)
	generateContentFunc func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error)
	getModelInfoFunc    func(ctx context.Context) (*llm.ProviderModelInfo, error)
	getModelNameFunc    func() string
	closeFunc           func() error
}

func (m *mockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	if m.generateContentFunc != nil {
		return m.generateContentFunc(ctx, prompt, params)
	}
	if m.generateContentErr != nil {
		return nil, m.generateContentErr
	}
	return &llm.ProviderResult{
		Content:    "mock response",
		TokenCount: m.tokenCount,
	}, nil
}

func (m *mockLLMClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	if m.countTokensFunc != nil {
		return m.countTokensFunc(ctx, prompt)
	}
	if m.countTokensErr != nil {
		return nil, m.countTokensErr
	}
	return &llm.ProviderTokenCount{Total: m.tokenCount}, nil
}

func (m *mockLLMClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	if m.getModelInfoFunc != nil {
		return m.getModelInfoFunc(ctx)
	}
	if m.getModelInfoErr != nil {
		return nil, m.getModelInfoErr
	}
	return &llm.ProviderModelInfo{
		Name:             m.modelName,
		InputTokenLimit:  m.inputTokenLimit,
		OutputTokenLimit: m.outputTokenLimit,
	}, nil
}

func (m *mockLLMClient) GetModelName() string {
	if m.getModelNameFunc != nil {
		return m.getModelNameFunc()
	}
	return m.modelName
}

func (m *mockLLMClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}
