package architect_test

import (
	"context"
	"testing"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
)

// mockLLMClient provides a minimal implementation of llm.LLMClient for testing
type mockLLMClient struct {
	modelName     string
	tokenResponse *llm.ProviderTokenCount
	modelInfo     *llm.ProviderModelInfo
}

func (m *mockLLMClient) GenerateContent(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
	return &llm.ProviderResult{
		Content:    "mock content",
		TokenCount: 100,
	}, nil
}

func (m *mockLLMClient) GetModelName() string {
	return m.modelName
}

func (m *mockLLMClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	return m.tokenResponse, nil
}

func (m *mockLLMClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	return m.modelInfo, nil
}

func (m *mockLLMClient) Close() error {
	return nil
}

// mockTokenAuditLogger implements auditlog.AuditLogger for testing token functionality
type mockTokenAuditLogger struct{}

func (m *mockTokenAuditLogger) Log(entry auditlog.AuditEntry) error {
	return nil
}

func (m *mockTokenAuditLogger) Close() error {
	return nil
}

// TestTokenManagerWithLLMClient tests TokenManager with an LLMClient implementation
func TestTokenManagerWithLLMClient(t *testing.T) {
	// Create a mock LLMClient
	client := &mockLLMClient{
		modelName: "test-model",
		tokenResponse: &llm.ProviderTokenCount{
			Total: 100,
		},
		modelInfo: &llm.ProviderModelInfo{
			Name:             "test-model",
			InputTokenLimit:  1000,
			OutputTokenLimit: 2000,
		},
	}

	// Create mock logger and audit logger
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "")

	// Create a simple mock audit logger
	auditLogger := &mockTokenAuditLogger{}

	// Create the TokenManager with the LLMClient
	tokenManager, err := architect.NewTokenManager(logger, auditLogger, client)

	// Validate creation
	if err != nil {
		t.Fatalf("Failed to create TokenManager: %v", err)
	}
	if tokenManager == nil {
		t.Fatal("Expected non-nil TokenManager")
	}

	// Test GetTokenInfo
	ctx := context.Background()
	result, err := tokenManager.GetTokenInfo(ctx, "test prompt")

	// Verify no error
	if err != nil {
		t.Errorf("GetTokenInfo returned error: %v", err)
	}

	// Verify token counts
	if result.TokenCount != 100 {
		t.Errorf("Expected TokenCount=100, got %d", result.TokenCount)
	}
	if result.InputLimit != 1000 {
		t.Errorf("Expected InputLimit=1000, got %d", result.InputLimit)
	}
}
