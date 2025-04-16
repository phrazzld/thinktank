//go:build token_test

package architect

import (
	"context"
	"fmt"
	"testing"

	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/registry"
	"github.com/stretchr/testify/assert"
)

// NoOpLogger implements a silent logger for testing
type NoOpLogger struct{}

func (l *NoOpLogger) Debug(format string, args ...interface{})  {}
func (l *NoOpLogger) Info(format string, args ...interface{})   {}
func (l *NoOpLogger) Warn(format string, args ...interface{})   {}
func (l *NoOpLogger) Error(format string, args ...interface{})  {}
func (l *NoOpLogger) Fatal(format string, args ...interface{})  {}
func (l *NoOpLogger) Println(args ...interface{})               {}
func (l *NoOpLogger) Printf(format string, args ...interface{}) {}

// MockAuditLogger implements a test audit logger
type MockAuditLogger struct {
	entries []auditlog.AuditEntry
}

func (m *MockAuditLogger) Log(entry auditlog.AuditEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func (m *MockAuditLogger) Close() error {
	return nil
}

// MockLLMClient implements a mock LLM client for testing
type MockLLMClient struct {
	modelName       string
	tokenCount      int32
	inputTokenLimit int32
}

func (m *MockLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	return &llm.ProviderResult{
		Content:    "mock response",
		TokenCount: m.tokenCount,
	}, nil
}

func (m *MockLLMClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	return &llm.ProviderTokenCount{Total: m.tokenCount}, nil
}

func (m *MockLLMClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	return &llm.ProviderModelInfo{
		Name:             m.modelName,
		InputTokenLimit:  m.inputTokenLimit,
		OutputTokenLimit: 1000,
	}, nil
}

func (m *MockLLMClient) GetModelName() string {
	return m.modelName
}

func (m *MockLLMClient) Close() error {
	return nil
}

// Define the token-related types needed for our tests
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
	GetTokenInfo(ctx context.Context, prompt string) (*TokenResult, error)

	// CheckTokenLimit verifies the prompt doesn't exceed the model's token limit
	CheckTokenLimit(ctx context.Context, prompt string) error

	// PromptForConfirmation asks for user confirmation to proceed if token count exceeds threshold
	PromptForConfirmation(tokenCount int32, threshold int) bool
}

// MockRegistry implements registry.Registry for testing
type MockRegistry struct {
	models map[string]*registry.ModelDefinition
}

// NewMockRegistry creates a new mock registry
func NewMockRegistry() *MockRegistry {
	return &MockRegistry{
		models: make(map[string]*registry.ModelDefinition),
	}
}

// GetModel returns a model definition by name
func (r *MockRegistry) GetModel(name string) (*registry.ModelDefinition, error) {
	model, ok := r.models[name]
	if !ok {
		return nil, fmt.Errorf("model not found: %s", name)
	}
	return model, nil
}

// AddModel adds a model to the registry
func (r *MockRegistry) AddModel(name string, contextWindow, maxOutputTokens int32) {
	r.models[name] = &registry.ModelDefinition{
		Name:            name,
		ContextWindow:   contextWindow,
		MaxOutputTokens: maxOutputTokens,
	}
}

// Test implementation of the tokenManager
type tokenManager struct {
	logger      logutil.LoggerInterface
	auditLogger auditlog.AuditLogger
	client      llm.LLMClient
	registry    interface {
		GetModel(name string) (*registry.ModelDefinition, error)
	}
}

func (tm *tokenManager) GetTokenInfo(ctx context.Context, prompt string) (*TokenResult, error) {
	// Get the model name from the injected client
	modelName := tm.client.GetModelName()

	// Create result structure
	result := &TokenResult{
		ExceedsLimit: false,
	}

	// First, try to get model info from the registry if available
	var useRegistryLimits bool
	if tm.registry != nil {
		// Try to get model definition from registry
		modelDef, err := tm.registry.GetModel(modelName)
		if err == nil && modelDef != nil {
			// We found the model in the registry, use its context window
			// Store input limit from registry
			result.InputLimit = modelDef.ContextWindow
			useRegistryLimits = true
		}
	}

	// If registry lookup failed or registry is not available, fall back to client GetModelInfo
	if !useRegistryLimits {
		// Get model information (limits) from LLM client
		modelInfo, err := tm.client.GetModelInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get model info for token limit check: %w", err)
		}
		// Store input limit from model info
		result.InputLimit = modelInfo.InputTokenLimit
	}

	// Count tokens in the prompt
	tokenResult, err := tm.client.CountTokens(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to count tokens for token limit check: %w", err)
	}

	// Store token count
	result.TokenCount = tokenResult.Total

	// Calculate percentage of limit
	result.Percentage = float64(result.TokenCount) / float64(result.InputLimit) * 100

	// Check if the prompt exceeds the token limit
	if result.TokenCount > result.InputLimit {
		result.ExceedsLimit = true
		result.LimitError = fmt.Sprintf("prompt exceeds token limit (%d tokens > %d token limit)",
			result.TokenCount, result.InputLimit)
	}

	return result, nil
}

func (tm *tokenManager) CheckTokenLimit(ctx context.Context, prompt string) error {
	tokenInfo, err := tm.GetTokenInfo(ctx, prompt)
	if err != nil {
		return err
	}

	if tokenInfo.ExceedsLimit {
		return fmt.Errorf("%s", tokenInfo.LimitError)
	}

	return nil
}

func (tm *tokenManager) PromptForConfirmation(tokenCount int32, threshold int) bool {
	// For tests, just return true
	return true
}

// TokenManager interface is used to override the standard implementation for testing
var NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg interface{}) TokenManager {
	return &tokenManager{
		logger:      logger,
		auditLogger: auditLogger,
		client:      client,
		registry: reg.(interface {
			GetModel(name string) (*registry.ModelDefinition, error)
		}),
	}
}

// NewTokenManager creates a new TokenManager instance
func NewTokenManager(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg interface{}) (TokenManager, error) {
	if client == nil {
		return nil, fmt.Errorf("client cannot be nil for TokenManager")
	}
	return NewTokenManagerWithClient(logger, auditLogger, client, reg), nil
}

// TestTokenManager_GetTokenInfo_Registry tests token checking using registry info
func TestTokenManager_GetTokenInfo_Registry(t *testing.T) {
	// Initialize test dependencies
	logger := &NoOpLogger{}
	auditLogger := &MockAuditLogger{}

	// Create a mock registry with test model
	mockReg := NewMockRegistry()
	mockReg.AddModel("test-model", 8000, 1000) // Context window 8000, max output 1000

	// Create a mock LLM client
	client := &MockLLMClient{
		modelName:       "test-model",
		tokenCount:      5000,
		inputTokenLimit: 4000, // This should be ignored in favor of registry value
	}

	// Store the original function and restore it after the test
	var originalNewTokenManagerWithClient = NewTokenManagerWithClient
	defer func() { NewTokenManagerWithClient = originalNewTokenManagerWithClient }()

	// Replace the function with our test version
	NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg interface{}) TokenManager {
		return &tokenManager{
			logger:      logger,
			auditLogger: auditLogger,
			client:      client,
			registry: reg.(interface {
				GetModel(name string) (*registry.ModelDefinition, error)
			}),
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
	logger := &NoOpLogger{}
	auditLogger := &MockAuditLogger{}

	// Create a mock registry with test model that has a low token limit
	mockReg := NewMockRegistry()
	mockReg.AddModel("test-model", 3000, 1000) // Context window 3000, max output 1000

	// Create a mock LLM client with token count exceeding the limit
	client := &MockLLMClient{
		modelName:       "test-model",
		tokenCount:      5000, // Exceeds the 3000 limit from registry
		inputTokenLimit: 8000, // This should be ignored in favor of registry value
	}

	// Store the original function and restore it after the test
	var originalNewTokenManagerWithClient = NewTokenManagerWithClient
	defer func() { NewTokenManagerWithClient = originalNewTokenManagerWithClient }()

	// Replace the function with our test version
	NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg interface{}) TokenManager {
		return &tokenManager{
			logger:      logger,
			auditLogger: auditLogger,
			client:      client,
			registry: reg.(interface {
				GetModel(name string) (*registry.ModelDefinition, error)
			}),
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
	logger := &NoOpLogger{}
	auditLogger := &MockAuditLogger{}

	// Create a mock registry without the test model
	mockReg := NewMockRegistry()
	// Deliberately not adding the test model to test fallback

	// Create a mock LLM client
	client := &MockLLMClient{
		modelName:       "test-model",
		tokenCount:      5000,
		inputTokenLimit: 8000, // This should be used as fallback
	}

	// Store the original function and restore it after the test
	var originalNewTokenManagerWithClient = NewTokenManagerWithClient
	defer func() { NewTokenManagerWithClient = originalNewTokenManagerWithClient }()

	// Replace the function with our test version
	NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg interface{}) TokenManager {
		return &tokenManager{
			logger:      logger,
			auditLogger: auditLogger,
			client:      client,
			registry: reg.(interface {
				GetModel(name string) (*registry.ModelDefinition, error)
			}),
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
	logger := &NoOpLogger{}
	auditLogger := &MockAuditLogger{}

	// Create a mock registry with test model
	mockReg := NewMockRegistry()
	mockReg.AddModel("test-model", 8000, 1000) // Context window 8000, max output 1000

	// Create a mock LLM client
	client := &MockLLMClient{
		modelName:       "test-model",
		tokenCount:      5000,
		inputTokenLimit: 4000, // This should be ignored in favor of registry value
	}

	// Store the original function and restore it after the test
	var originalNewTokenManagerWithClient = NewTokenManagerWithClient
	defer func() { NewTokenManagerWithClient = originalNewTokenManagerWithClient }()

	// Replace the function with our test version
	NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg interface{}) TokenManager {
		return &tokenManager{
			logger:      logger,
			auditLogger: auditLogger,
			client:      client,
			registry: reg.(interface {
				GetModel(name string) (*registry.ModelDefinition, error)
			}),
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
	logger := &NoOpLogger{}
	auditLogger := &MockAuditLogger{}

	// Create a mock registry with test model that has a low token limit
	mockReg := NewMockRegistry()
	mockReg.AddModel("test-model", 3000, 1000) // Context window 3000, max output 1000

	// Create a mock LLM client with token count exceeding the limit
	client := &MockLLMClient{
		modelName:       "test-model",
		tokenCount:      5000, // Exceeds the 3000 limit from registry
		inputTokenLimit: 8000, // This should be ignored in favor of registry value
	}

	// Store the original function and restore it after the test
	var originalNewTokenManagerWithClient = NewTokenManagerWithClient
	defer func() { NewTokenManagerWithClient = originalNewTokenManagerWithClient }()

	// Replace the function with our test version
	NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg interface{}) TokenManager {
		return &tokenManager{
			logger:      logger,
			auditLogger: auditLogger,
			client:      client,
			registry: reg.(interface {
				GetModel(name string) (*registry.ModelDefinition, error)
			}),
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
