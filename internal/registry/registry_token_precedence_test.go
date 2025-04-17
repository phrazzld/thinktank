package registry_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/registry"
	"github.com/stretchr/testify/assert"
)

// TokenResult holds information about token counts and limits
type TokenResult struct {
	TokenCount   int32
	InputLimit   int32
	ExceedsLimit bool
	LimitError   string
	Percentage   float64
}

// mockTokenManager is a simplified version of the tokenManager from the architect package
type mockTokenManager struct {
	logger      logutil.LoggerInterface
	auditLogger auditlog.AuditLogger
	client      llm.LLMClient
	registry    interface {
		GetModel(name string) (*registry.ModelDefinition, error)
	}
}

// GetTokenInfo simulates the original GetTokenInfo method
func (tm *mockTokenManager) GetTokenInfo(ctx context.Context, prompt string) (*TokenResult, error) {
	// Get the model name from the injected client
	modelName := tm.client.GetModelName()

	// Create result structure
	result := &TokenResult{
		ExceedsLimit: false,
	}

	// First, try to get model info from the registry if available
	var tokenSource = "client"
	// We track registryAttempted in the original code for logging purposes
	// but for tests we can simplify

	if tm.registry != nil {
		// Try to get model definition from registry
		modelDef, err := tm.registry.GetModel(modelName)
		if err == nil && modelDef != nil && modelDef.ContextWindow > 0 {
			// We found the model in the registry with a valid context window
			tm.logger.Info("Using token limits from registry for model %s: %d tokens (registry values take precedence)",
				modelName, modelDef.ContextWindow)

			// Store input limit from registry
			result.InputLimit = modelDef.ContextWindow
			tokenSource = "registry"
		} else {
			if err != nil {
				tm.logger.Debug("Model %s lookup in registry failed: %v", modelName, err)
			} else if modelDef == nil {
				tm.logger.Debug("Model %s found in registry but model definition is nil", modelName)
			} else if modelDef.ContextWindow <= 0 {
				tm.logger.Debug("Model %s found in registry but has invalid context window: %d",
					modelName, modelDef.ContextWindow)
			}
			tm.logger.Info("Model %s not properly configured in registry, falling back to client-provided token limits",
				modelName)
		}
	}

	// If registry lookup failed or registry is not available, fall back to client GetModelInfo
	if tokenSource != "registry" {
		// Get model information (limits) from LLM client
		modelInfo, err := tm.client.GetModelInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get model info for token limit check: %w", err)
		}

		// Store input limit from model info
		result.InputLimit = modelInfo.InputTokenLimit
		tm.logger.Info("Using token limits from client for model %s: %d tokens",
			modelName, modelInfo.InputTokenLimit)
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

// CheckTokenLimit checks if the prompt exceeds the token limit
func (tm *mockTokenManager) CheckTokenLimit(ctx context.Context, prompt string) error {
	tokenInfo, err := tm.GetTokenInfo(ctx, prompt)
	if err != nil {
		return err
	}

	if tokenInfo.ExceedsLimit {
		return fmt.Errorf("%s", tokenInfo.LimitError)
	}

	return nil
}

// MockRegistryWithTokenLimits provides a configurable mock registry for token limit testing
type MockRegistryWithTokenLimits struct {
	models               map[string]*registry.ModelDefinition
	modelNotFound        bool
	modelNil             bool
	invalidContextWindow bool
}

// NewMockRegistryWithTokenLimits creates a new mock registry with configurable behavior
func NewMockRegistryWithTokenLimits() *MockRegistryWithTokenLimits {
	return &MockRegistryWithTokenLimits{
		models: make(map[string]*registry.ModelDefinition),
	}
}

// GetModel returns a model definition by name with configurable behavior for testing
func (r *MockRegistryWithTokenLimits) GetModel(name string) (*registry.ModelDefinition, error) {
	// Simulate model not found in registry
	if r.modelNotFound {
		return nil, fmt.Errorf("model '%s' not found in registry", name)
	}

	// Simulate model found but nil definition
	if r.modelNil {
		return nil, nil
	}

	model, ok := r.models[name]
	if !ok {
		return nil, fmt.Errorf("model not found: %s", name)
	}

	// Simulate invalid context window
	if r.invalidContextWindow {
		model.ContextWindow = 0
	}

	return model, nil
}

// Add a model to the mock registry
func (r *MockRegistryWithTokenLimits) AddModel(name string, contextWindow, maxOutputTokens int32) {
	r.models[name] = &registry.ModelDefinition{
		Name:            name,
		ContextWindow:   contextWindow,
		MaxOutputTokens: maxOutputTokens,
	}
}

// CaptureLogger logs to an in-memory buffer for testing assertions
type CaptureLogger struct {
	InfoLogs  []string
	DebugLogs []string
	WarnLogs  []string
	ErrorLogs []string
}

// NewCaptureLogger creates a new logger that captures logs in memory
func NewCaptureLogger() *CaptureLogger {
	return &CaptureLogger{
		InfoLogs:  make([]string, 0),
		DebugLogs: make([]string, 0),
		WarnLogs:  make([]string, 0),
		ErrorLogs: make([]string, 0),
	}
}

// Debug implements logutil.LoggerInterface
func (l *CaptureLogger) Debug(format string, args ...interface{}) {
	l.DebugLogs = append(l.DebugLogs, fmt.Sprintf(format, args...))
}

// Info implements logutil.LoggerInterface
func (l *CaptureLogger) Info(format string, args ...interface{}) {
	l.InfoLogs = append(l.InfoLogs, fmt.Sprintf(format, args...))
}

// Warn implements logutil.LoggerInterface
func (l *CaptureLogger) Warn(format string, args ...interface{}) {
	l.WarnLogs = append(l.WarnLogs, fmt.Sprintf(format, args...))
}

// Error implements logutil.LoggerInterface
func (l *CaptureLogger) Error(format string, args ...interface{}) {
	l.ErrorLogs = append(l.ErrorLogs, fmt.Sprintf(format, args...))
}

// Implement Println and Printf for the CaptureLogger
func (l *CaptureLogger) Println(args ...interface{}) {
	l.InfoLogs = append(l.InfoLogs, fmt.Sprint(args...))
}

func (l *CaptureLogger) Printf(format string, args ...interface{}) {
	l.InfoLogs = append(l.InfoLogs, fmt.Sprintf(format, args...))
}

// Fatal implements logutil.LoggerInterface
func (l *CaptureLogger) Fatal(format string, args ...interface{}) {
	l.ErrorLogs = append(l.ErrorLogs, fmt.Sprintf(format, args...))
	// In tests, we don't actually exit
}

// MockLLMClientWithTokenLimits implements llm.LLMClient for token limit testing
type MockLLMClientWithTokenLimits struct {
	modelName              string
	tokenCount             int32
	inputTokenLimit        int32
	getModelInfoError      error
	countTokensError       error
	getModelInfoCalledWith context.Context
	countTokensCalledWith  string
	countTokensCalled      bool
	getModelInfoCalled     bool
}

// GenerateContent implements llm.LLMClient
func (m *MockLLMClientWithTokenLimits) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	return &llm.ProviderResult{
		Content:    "mock response",
		TokenCount: m.tokenCount,
	}, nil
}

// CountTokens implements llm.LLMClient
func (m *MockLLMClientWithTokenLimits) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	m.countTokensCalled = true
	m.countTokensCalledWith = prompt

	if m.countTokensError != nil {
		return nil, m.countTokensError
	}

	return &llm.ProviderTokenCount{Total: m.tokenCount}, nil
}

// GetModelInfo implements llm.LLMClient
func (m *MockLLMClientWithTokenLimits) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	m.getModelInfoCalled = true
	m.getModelInfoCalledWith = ctx

	if m.getModelInfoError != nil {
		return nil, m.getModelInfoError
	}

	return &llm.ProviderModelInfo{
		Name:             m.modelName,
		InputTokenLimit:  m.inputTokenLimit,
		OutputTokenLimit: 1000,
	}, nil
}

// GetModelName implements llm.LLMClient
func (m *MockLLMClientWithTokenLimits) GetModelName() string {
	return m.modelName
}

// Close implements llm.LLMClient
func (m *MockLLMClientWithTokenLimits) Close() error {
	return nil
}

// MockAuditLoggerNoOp implements a dummy audit logger for testing
type MockAuditLoggerNoOp struct{}

func (m *MockAuditLoggerNoOp) Log(entry auditlog.AuditEntry) error {
	return nil
}

func (m *MockAuditLoggerNoOp) Close() error {
	return nil
}

// TestRegistryTokenPrecedence tests that registry token limits take precedence over client limits
func TestRegistryTokenPrecedence(t *testing.T) {
	// Initialize test dependencies
	captureLogger := &CaptureLogger{}
	auditLogger := &MockAuditLoggerNoOp{}

	// Test cases for token limit precedence
	tests := []struct {
		name                   string
		modelName              string
		registryContextWindow  int32
		clientInputTokenLimit  int32
		tokenCount             int32
		expectedLimit          int32
		expectedExceedsLimit   bool
		configureRegistry      func(*MockRegistryWithTokenLimits)
		expectedRegistrySource bool
		expectedLogContains    string
	}{
		{
			name:                   "Registry values take precedence when available",
			modelName:              "test-model-1",
			registryContextWindow:  8000,
			clientInputTokenLimit:  4000, // This should be ignored in favor of registry value
			tokenCount:             5000,
			expectedLimit:          8000, // Should use registry limit
			expectedExceedsLimit:   false,
			configureRegistry:      func(r *MockRegistryWithTokenLimits) {},
			expectedRegistrySource: true,
			expectedLogContains:    "Using token limits from registry for model test-model-1",
		},
		{
			name:                  "Client values used when model not found in registry",
			modelName:             "test-model-2",
			registryContextWindow: 8000,
			clientInputTokenLimit: 4000,
			tokenCount:            3000,
			expectedLimit:         4000, // Should use client limit
			expectedExceedsLimit:  false,
			configureRegistry: func(r *MockRegistryWithTokenLimits) {
				r.modelNotFound = true
			},
			expectedRegistrySource: false,
			expectedLogContains:    "not properly configured in registry",
		},
		{
			name:                  "Client values used when registry return nil model",
			modelName:             "test-model-3",
			registryContextWindow: 8000,
			clientInputTokenLimit: 4000,
			tokenCount:            3000,
			expectedLimit:         4000, // Should use client limit
			expectedExceedsLimit:  false,
			configureRegistry: func(r *MockRegistryWithTokenLimits) {
				r.modelNil = true
			},
			expectedRegistrySource: false,
			expectedLogContains:    "found in registry but model definition is nil",
		},
		{
			name:                  "Client values used when registry has invalid context window",
			modelName:             "test-model-4",
			registryContextWindow: 8000,
			clientInputTokenLimit: 4000,
			tokenCount:            3000,
			expectedLimit:         4000, // Should use client limit
			expectedExceedsLimit:  false,
			configureRegistry: func(r *MockRegistryWithTokenLimits) {
				r.invalidContextWindow = true
			},
			expectedRegistrySource: false,
			expectedLogContains:    "found in registry but has invalid context window",
		},
		{
			name:                   "Registry values cause token limit exceeded when appropriate",
			modelName:              "test-model-5",
			registryContextWindow:  3000, // Lower than token count
			clientInputTokenLimit:  8000, // This should be ignored in favor of registry value
			tokenCount:             5000,
			expectedLimit:          3000, // Should use registry limit
			expectedExceedsLimit:   true, // Should exceed limit
			configureRegistry:      func(r *MockRegistryWithTokenLimits) {},
			expectedRegistrySource: true,
			expectedLogContains:    "Using token limits from registry for model test-model-5",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize test dependencies
			captureLogger.InfoLogs = nil
			captureLogger.DebugLogs = nil
			captureLogger.WarnLogs = nil
			captureLogger.ErrorLogs = nil

			// Create a mock registry with test model
			mockReg := NewMockRegistryWithTokenLimits()
			mockReg.AddModel(tc.modelName, tc.registryContextWindow, 1000)

			// Apply any registry configurations
			tc.configureRegistry(mockReg)

			// Create a mock LLM client
			client := &MockLLMClientWithTokenLimits{
				modelName:       tc.modelName,
				tokenCount:      tc.tokenCount,
				inputTokenLimit: tc.clientInputTokenLimit,
			}

			// Create token manager test implementation
			tm := &mockTokenManager{
				logger:      captureLogger,
				auditLogger: auditLogger,
				client:      client,
				registry:    mockReg,
			}

			// Test token checking
			ctx := context.Background()
			tokenInfo, err := tm.GetTokenInfo(ctx, "test prompt")

			// Assert no error
			assert.NoError(t, err)

			// Verify token count and limit
			assert.Equal(t, tc.tokenCount, tokenInfo.TokenCount)
			assert.Equal(t, tc.expectedLimit, tokenInfo.InputLimit)
			assert.Equal(t, tc.expectedExceedsLimit, tokenInfo.ExceedsLimit)

			// Verify percentage is calculated correctly
			expectedPercentage := float64(tc.tokenCount) / float64(tc.expectedLimit) * 100
			assert.Equal(t, expectedPercentage, tokenInfo.Percentage)

			// Verify that the client method was called
			assert.True(t, client.countTokensCalled)

			// Check if we used registry or client token limits based on the expected source
			if tc.expectedRegistrySource {
				// Check that we didn't call GetModelInfo() on the client
				assert.False(t, client.getModelInfoCalled, "GetModelInfo should not be called when registry values are used")

				// Check logs to verify registry was used
				foundRegistry := false
				for _, logMsg := range captureLogger.InfoLogs {
					if strings.Contains(logMsg, "Using token limits from registry for model") {
						foundRegistry = true
						break
					}
				}
				assert.True(t, foundRegistry, "Expected log message about using registry not found")
			} else {
				// Check that we called GetModelInfo() on the client for fallback
				assert.True(t, client.getModelInfoCalled, "GetModelInfo should be called when registry values are not used")

				// Check logs to verify client was used
				clientLogFound := false
				for _, logMsg := range captureLogger.InfoLogs {
					if strings.Contains(logMsg, "Using token limits from client for model") {
						clientLogFound = true
						break
					}
				}
				assert.True(t, clientLogFound, "Expected log message about using client not found")
			}
		})
	}
}

// TestCheckTokenLimitRegistryPrecedence tests that the CheckTokenLimit method
// respects registry precedence for token limits
func TestCheckTokenLimitRegistryPrecedence(t *testing.T) {
	// Initialize test dependencies
	captureLogger := &CaptureLogger{}
	auditLogger := &MockAuditLoggerNoOp{}

	// Test cases
	tests := []struct {
		name                  string
		modelName             string
		registryContextWindow int32
		clientInputTokenLimit int32
		tokenCount            int32
		expectError           bool
		configureRegistry     func(*MockRegistryWithTokenLimits)
	}{
		{
			name:                  "No error when registry limit not exceeded",
			modelName:             "test-model-1",
			registryContextWindow: 8000,
			clientInputTokenLimit: 4000, // Should be ignored
			tokenCount:            5000,
			expectError:           false,
			configureRegistry:     func(r *MockRegistryWithTokenLimits) {},
		},
		{
			name:                  "Error when registry limit exceeded",
			modelName:             "test-model-2",
			registryContextWindow: 3000, // Lower than token count
			clientInputTokenLimit: 8000, // Should be ignored
			tokenCount:            5000,
			expectError:           true,
			configureRegistry:     func(r *MockRegistryWithTokenLimits) {},
		},
		{
			name:                  "No error when client limit not exceeded (registry not found)",
			modelName:             "test-model-3",
			registryContextWindow: 2000, // Not used
			clientInputTokenLimit: 6000,
			tokenCount:            5000,
			expectError:           false,
			configureRegistry: func(r *MockRegistryWithTokenLimits) {
				r.modelNotFound = true
			},
		},
		{
			name:                  "Error when client limit exceeded (registry not found)",
			modelName:             "test-model-4",
			registryContextWindow: 8000, // Not used
			clientInputTokenLimit: 4000,
			tokenCount:            5000,
			expectError:           true,
			configureRegistry: func(r *MockRegistryWithTokenLimits) {
				r.modelNotFound = true
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock registry with test model
			mockReg := NewMockRegistryWithTokenLimits()
			mockReg.AddModel(tc.modelName, tc.registryContextWindow, 1000)

			// Apply any registry configurations
			tc.configureRegistry(mockReg)

			// Create a mock LLM client
			client := &MockLLMClientWithTokenLimits{
				modelName:       tc.modelName,
				tokenCount:      tc.tokenCount,
				inputTokenLimit: tc.clientInputTokenLimit,
			}

			// Create token manager test implementation
			tm := &mockTokenManager{
				logger:      captureLogger,
				auditLogger: auditLogger,
				client:      client,
				registry:    mockReg,
			}

			// Test token limit checking
			ctx := context.Background()
			err := tm.CheckTokenLimit(ctx, "test prompt")

			// Verify error expectation
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "prompt exceeds token limit")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRegistryTokenEdgeCases tests edge cases in registry token limit handling
func TestRegistryTokenEdgeCases(t *testing.T) {
	// Initialize test dependencies
	captureLogger := &CaptureLogger{}
	auditLogger := &MockAuditLoggerNoOp{}

	// Test cases
	tests := []struct {
		name                  string
		modelName             string
		registryContextWindow int32
		clientInputTokenLimit int32
		tokenCount            int32
		expectedLimit         int32
		configureRegistry     func(*MockRegistryWithTokenLimits)
		configureClient       func(*MockLLMClientWithTokenLimits)
		expectError           bool
		expectedErrorContains string
	}{
		{
			name:                  "GetModelInfo error results in error return",
			modelName:             "error-model",
			registryContextWindow: 0, // Not used
			clientInputTokenLimit: 0, // Not used
			tokenCount:            1000,
			expectedLimit:         0, // Not used
			configureRegistry: func(r *MockRegistryWithTokenLimits) {
				r.modelNotFound = true
			},
			configureClient: func(c *MockLLMClientWithTokenLimits) {
				c.getModelInfoError = fmt.Errorf("API error getting model info")
			},
			expectError:           true,
			expectedErrorContains: "failed to get model info for token limit check",
		},
		{
			name:                  "CountTokens error results in error return",
			modelName:             "token-error-model",
			registryContextWindow: 8000,
			clientInputTokenLimit: 4000,
			tokenCount:            0,    // Not used
			expectedLimit:         8000, // Registry value, but not used due to error
			configureRegistry:     func(r *MockRegistryWithTokenLimits) {},
			configureClient: func(c *MockLLMClientWithTokenLimits) {
				c.countTokensError = fmt.Errorf("API error counting tokens")
			},
			expectError:           true,
			expectedErrorContains: "failed to count tokens for token limit check",
		},
		{
			name:                  "Zero token count doesn't exceed limit",
			modelName:             "zero-tokens",
			registryContextWindow: 1000,
			clientInputTokenLimit: 1000,
			tokenCount:            0,
			expectedLimit:         1000,
			configureRegistry:     func(r *MockRegistryWithTokenLimits) {},
			configureClient:       func(c *MockLLMClientWithTokenLimits) {},
			expectError:           false,
		},
		{
			name:                  "Equal token count to limit doesn't exceed limit",
			modelName:             "exact-limit",
			registryContextWindow: 5000,
			clientInputTokenLimit: 4000,
			tokenCount:            5000, // Exactly the registry limit
			expectedLimit:         5000,
			configureRegistry:     func(r *MockRegistryWithTokenLimits) {},
			configureClient:       func(c *MockLLMClientWithTokenLimits) {},
			expectError:           false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock registry with test model
			mockReg := NewMockRegistryWithTokenLimits()
			if tc.registryContextWindow > 0 {
				mockReg.AddModel(tc.modelName, tc.registryContextWindow, 1000)
			}

			// Apply any registry configurations
			tc.configureRegistry(mockReg)

			// Create a mock LLM client
			client := &MockLLMClientWithTokenLimits{
				modelName:       tc.modelName,
				tokenCount:      tc.tokenCount,
				inputTokenLimit: tc.clientInputTokenLimit,
			}

			// Apply any client configurations
			tc.configureClient(client)

			// Create token manager test implementation
			tm := &mockTokenManager{
				logger:      captureLogger,
				auditLogger: auditLogger,
				client:      client,
				registry:    mockReg,
			}

			// Test token checking
			ctx := context.Background()
			result, err := tm.GetTokenInfo(ctx, "test prompt")

			// Verify error expectation
			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedErrorContains != "" {
					assert.Contains(t, err.Error(), tc.expectedErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.tokenCount, result.TokenCount)
				assert.Equal(t, tc.expectedLimit, result.InputLimit)
			}
		})
	}
}
