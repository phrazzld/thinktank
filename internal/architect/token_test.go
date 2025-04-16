package architect

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/registry"
	"github.com/stretchr/testify/assert"
)

// mockLogger implements logutil.LoggerInterface for testing
type mockLogger struct{}

func (m *mockLogger) Debug(format string, args ...interface{})              {}
func (m *mockLogger) Info(format string, args ...interface{})               {}
func (m *mockLogger) Warn(format string, args ...interface{})               {}
func (m *mockLogger) Error(format string, args ...interface{})              {}
func (m *mockLogger) Format(logutil.LogLevel, string, []interface{}) string { return "" }

// mockAuditLogger implements auditlog.AuditLogger for testing
type mockAuditLogger struct{}

func (m *mockAuditLogger) Log(entry auditlog.AuditEntry) error          { return nil }
func (m *mockAuditLogger) Close() error                                 { return nil }
func (m *mockAuditLogger) SetConfig(config auditlog.LoggerConfig) error { return nil }
func (m *mockAuditLogger) Last() *auditlog.AuditEntry                   { return nil }

// mockLLMClient implements llm.LLMClient for testing
type mockLLMClient struct {
	modelName          string
	tokenCount         int32
	inputTokenLimit    int32
	outputTokenLimit   int32
	generateContentErr error
	countTokensErr     error
	getModelInfoErr    error
}

func (m *mockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	if m.generateContentErr != nil {
		return nil, m.generateContentErr
	}
	return &llm.ProviderResult{
		Content:    "mock response",
		TokenCount: m.tokenCount,
	}, nil
}

func (m *mockLLMClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	if m.countTokensErr != nil {
		return nil, m.countTokensErr
	}
	return &llm.ProviderTokenCount{Total: m.tokenCount}, nil
}

func (m *mockLLMClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
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
	return m.modelName
}

func (m *mockLLMClient) Close() error {
	return nil
}

// mockRegistry implements a simple registry for testing
type mockRegistry struct {
	models map[string]registry.ModelDefinition
}

func newMockRegistry() *mockRegistry {
	return &mockRegistry{
		models: make(map[string]registry.ModelDefinition),
	}
}

func (r *mockRegistry) GetModel(name string) (*registry.ModelDefinition, error) {
	model, ok := r.models[name]
	if !ok {
		return nil, errors.New("model not found")
	}
	return &model, nil
}

func (r *mockRegistry) AddModel(name string, contextWindow, maxOutputTokens int32) {
	r.models[name] = registry.ModelDefinition{
		Name:            name,
		ContextWindow:   contextWindow,
		MaxOutputTokens: maxOutputTokens,
	}
}

// TestTokenManager_GetTokenInfo_Registry tests token checking using registry info
func TestTokenManager_GetTokenInfo_Registry(t *testing.T) {
	// Initialize test dependencies
	logger := &mockLogger{}
	auditLogger := &mockAuditLogger{}

	// Create a mock registry with test model
	mockReg := newMockRegistry()
	mockReg.AddModel("test-model", 8000, 1000) // Context window 8000, max output 1000

	// Create a mock LLM client
	client := &mockLLMClient{
		modelName:       "test-model",
		tokenCount:      5000,
		inputTokenLimit: 4000, // This should be ignored in favor of registry value
	}

	// Create a token manager with registry
	originalNewTokenManagerWithClient := NewTokenManagerWithClient
	defer func() { NewTokenManagerWithClient = originalNewTokenManagerWithClient }()

	NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) TokenManager {
		return &tokenManager{
			logger:      logger,
			auditLogger: auditLogger,
			client:      client,
			registry:    reg,
		}
	}

	// Create the token manager
	tokenManager, err := NewTokenManager(logger, auditLogger, client, mockReg)
	assert.NoError(t, err)

	// Test token checking
	ctx := context.Background()
	tokenInfo, err := tokenManager.GetTokenInfo(ctx, "test prompt")

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, int32(5000), tokenInfo.TokenCount)
	assert.Equal(t, int32(8000), tokenInfo.InputLimit) // Should use registry limit
	assert.False(t, tokenInfo.ExceedsLimit)
	assert.Equal(t, float64(5000)/float64(8000)*100, tokenInfo.Percentage)
}

// TestTokenManager_GetTokenInfo_Registry_Exceeds tests token limit exceeding with registry info
func TestTokenManager_GetTokenInfo_Registry_Exceeds(t *testing.T) {
	// Initialize test dependencies
	logger := &mockLogger{}
	auditLogger := &mockAuditLogger{}

	// Create a mock registry with test model that has a low token limit
	mockReg := newMockRegistry()
	mockReg.AddModel("test-model", 3000, 1000) // Context window 3000, max output 1000

	// Create a mock LLM client with token count exceeding the limit
	client := &mockLLMClient{
		modelName:       "test-model",
		tokenCount:      5000, // Exceeds the 3000 limit from registry
		inputTokenLimit: 8000, // This should be ignored in favor of registry value
	}

	// Create a token manager with registry
	originalNewTokenManagerWithClient := NewTokenManagerWithClient
	defer func() { NewTokenManagerWithClient = originalNewTokenManagerWithClient }()

	NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) TokenManager {
		return &tokenManager{
			logger:      logger,
			auditLogger: auditLogger,
			client:      client,
			registry:    reg,
		}
	}

	// Create the token manager
	tokenManager, err := NewTokenManager(logger, auditLogger, client, mockReg)
	assert.NoError(t, err)

	// Test token checking
	ctx := context.Background()
	tokenInfo, err := tokenManager.GetTokenInfo(ctx, "test prompt")

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, int32(5000), tokenInfo.TokenCount)
	assert.Equal(t, int32(3000), tokenInfo.InputLimit) // Should use registry limit
	assert.True(t, tokenInfo.ExceedsLimit)
	assert.Contains(t, tokenInfo.LimitError, "prompt exceeds token limit")
}

// TestTokenManager_GetTokenInfo_Fallback tests falling back to client model info when registry fails
func TestTokenManager_GetTokenInfo_Fallback(t *testing.T) {
	// Initialize test dependencies
	logger := &mockLogger{}
	auditLogger := &mockAuditLogger{}

	// Create a mock registry without the test model
	mockReg := newMockRegistry()
	// Deliberately not adding the test model to test fallback

	// Create a mock LLM client
	client := &mockLLMClient{
		modelName:       "test-model",
		tokenCount:      5000,
		inputTokenLimit: 8000, // This should be used as fallback
	}

	// Create a token manager with registry
	originalNewTokenManagerWithClient := NewTokenManagerWithClient
	defer func() { NewTokenManagerWithClient = originalNewTokenManagerWithClient }()

	NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) TokenManager {
		return &tokenManager{
			logger:      logger,
			auditLogger: auditLogger,
			client:      client,
			registry:    reg,
		}
	}

	// Create the token manager
	tokenManager, err := NewTokenManager(logger, auditLogger, client, mockReg)
	assert.NoError(t, err)

	// Test token checking
	ctx := context.Background()
	tokenInfo, err := tokenManager.GetTokenInfo(ctx, "test prompt")

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, int32(5000), tokenInfo.TokenCount)
	assert.Equal(t, int32(8000), tokenInfo.InputLimit) // Should fall back to client limit
	assert.False(t, tokenInfo.ExceedsLimit)
}

// TestTokenManager_CheckTokenLimit_Registry tests the CheckTokenLimit method using registry info
func TestTokenManager_CheckTokenLimit_Registry(t *testing.T) {
	// Initialize test dependencies
	logger := &mockLogger{}
	auditLogger := &mockAuditLogger{}

	// Create a mock registry with test model
	mockReg := newMockRegistry()
	mockReg.AddModel("test-model", 8000, 1000) // Context window 8000, max output 1000

	// Create a mock LLM client
	client := &mockLLMClient{
		modelName:       "test-model",
		tokenCount:      5000,
		inputTokenLimit: 4000, // This should be ignored in favor of registry value
	}

	// Create a token manager with registry
	originalNewTokenManagerWithClient := NewTokenManagerWithClient
	defer func() { NewTokenManagerWithClient = originalNewTokenManagerWithClient }()

	NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) TokenManager {
		return &tokenManager{
			logger:      logger,
			auditLogger: auditLogger,
			client:      client,
			registry:    reg,
		}
	}

	// Create the token manager
	tokenManager, err := NewTokenManager(logger, auditLogger, client, mockReg)
	assert.NoError(t, err)

	// Test token checking
	ctx := context.Background()
	err = tokenManager.CheckTokenLimit(ctx, "test prompt")

	// Assert results
	assert.NoError(t, err)
}

// TestTokenManager_CheckTokenLimit_Registry_Exceeds tests the CheckTokenLimit method with registry when exceeding limits
func TestTokenManager_CheckTokenLimit_Registry_Exceeds(t *testing.T) {
	// Initialize test dependencies
	logger := &mockLogger{}
	auditLogger := &mockAuditLogger{}

	// Create a mock registry with test model that has a low token limit
	mockReg := newMockRegistry()
	mockReg.AddModel("test-model", 3000, 1000) // Context window 3000, max output 1000

	// Create a mock LLM client with token count exceeding the limit
	client := &mockLLMClient{
		modelName:       "test-model",
		tokenCount:      5000, // Exceeds the 3000 limit from registry
		inputTokenLimit: 8000, // This should be ignored in favor of registry value
	}

	// Create a token manager with registry
	originalNewTokenManagerWithClient := NewTokenManagerWithClient
	defer func() { NewTokenManagerWithClient = originalNewTokenManagerWithClient }()

	NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) TokenManager {
		return &tokenManager{
			logger:      logger,
			auditLogger: auditLogger,
			client:      client,
			registry:    reg,
		}
	}

	// Create the token manager
	tokenManager, err := NewTokenManager(logger, auditLogger, client, mockReg)
	assert.NoError(t, err)

	// Test token checking
	ctx := context.Background()
	err = tokenManager.CheckTokenLimit(ctx, "test prompt")

	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt exceeds token limit")
}
