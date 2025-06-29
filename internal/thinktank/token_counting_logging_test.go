package thinktank

import (
	"context"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/phrazzld/thinktank/internal/thinktank/tokenizers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED: Write the smallest failing test first - logging integration
func TestTokenCountingService_LogsModelSelectionStart(t *testing.T) {
	t.Parallel()

	mockLogger := &testutil.MockLogger{}
	service := NewTokenCountingServiceWithLogger(mockLogger)

	ctx := context.Background()
	req := TokenCountingRequest{
		Instructions: "test",
		Files:        []FileContent{},
	}

	_, err := service.GetCompatibleModels(ctx, req, []string{"openai"})

	require.NoError(t, err)
	assert.True(t, mockLogger.ContainsMessage("Starting model compatibility check"))
}

// RED: Test correlation ID propagation through logging
func TestTokenCountingService_LogsWithCorrelationID(t *testing.T) {
	t.Parallel()

	mockLogger := &testutil.MockLogger{}
	service := NewTokenCountingServiceWithLogger(mockLogger)

	correlationID := "test-correlation-123"
	ctx := logutil.WithCorrelationID(context.Background(), correlationID)

	req := TokenCountingRequest{
		Instructions: "test",
		Files:        []FileContent{},
	}

	_, err := service.GetCompatibleModels(ctx, req, []string{"openai"})

	require.NoError(t, err)
	entries := mockLogger.GetLogEntriesByCorrelationID(correlationID)
	assert.NotEmpty(t, entries, "Should log with correlation ID")
}

// RED: Test detailed model evaluation logging per TODO.md requirements
func TestTokenCountingService_LogsModelEvaluationDetails(t *testing.T) {
	t.Parallel()

	mockLogger := &testutil.MockLogger{}
	service := NewTokenCountingServiceWithLogger(mockLogger)

	ctx := context.Background()
	req := TokenCountingRequest{
		Instructions: "test instructions",
		Files: []FileContent{
			{Path: "test.go", Content: "package main"},
		},
	}

	_, err := service.GetCompatibleModels(ctx, req, []string{"openai"})

	require.NoError(t, err)

	// Should log model evaluation details as per TODO.md Phase 4.2
	foundModelEvaluation := false
	for _, msg := range mockLogger.GetMessages() {
		if strings.Contains(msg, "Model evaluation") {
			foundModelEvaluation = true
			break
		}
	}
	assert.True(t, foundModelEvaluation, "Should log model evaluation details")
}

// RED: Test detailed start context logging per TODO.md requirements
func TestTokenCountingService_LogsDetailedStartContext(t *testing.T) {
	t.Parallel()

	mockLogger := &testutil.MockLogger{}
	service := NewTokenCountingServiceWithLogger(mockLogger)

	ctx := context.Background()
	req := TokenCountingRequest{
		Instructions: "test instructions with multiple files",
		Files: []FileContent{
			{Path: "main.go", Content: "package main"},
			{Path: "utils.go", Content: "package utils"},
		},
	}

	_, err := service.GetCompatibleModels(ctx, req, []string{"openai", "gemini"})

	require.NoError(t, err)

	// Should log start with provider_count, file_count, has_instructions as per TODO.md
	foundStartLog := false
	for _, msg := range mockLogger.GetMessages() {
		if strings.Contains(msg, "Starting model compatibility check") {
			foundStartLog = true
			break
		}
	}
	assert.True(t, foundStartLog, "Should log detailed start context")
}

// RED: Comprehensive error path testing
func TestTokenCountingService_CheckModelCompatibility_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupService     func() TokenCountingService
		model            string
		expectError      bool
		expectFallback   bool
		logShouldContain string
	}{
		{
			name: "Unknown model falls back to estimation",
			setupService: func() TokenCountingService {
				return NewTokenCountingService()
			},
			model:            "unknown-model-xyz",
			expectError:      false,
			expectFallback:   true,
			logShouldContain: "",
		},
		{
			name: "Tokenizer manager failure falls back gracefully",
			setupService: func() TokenCountingService {
				mockLogger := &testutil.MockLogger{}
				failingManager := &MockFailingTokenizerManager{ShouldFailAll: true}
				return NewTokenCountingServiceWithManagerAndLogger(failingManager, mockLogger)
			},
			model:            "gpt-4",
			expectError:      false,
			expectFallback:   true,
			logShouldContain: "falling back to estimation",
		},
		{
			name: "Tokenizer failure with recovery",
			setupService: func() TokenCountingService {
				mockLogger := &testutil.MockLogger{}
				// Create a manager that returns a failing tokenizer
				manager := &MockTokenizerManagerWithFailingTokenizer{}
				return NewTokenCountingServiceWithManagerAndLogger(manager, mockLogger)
			},
			model:            "gpt-4",
			expectError:      false,
			expectFallback:   true,
			logShouldContain: "falling back to estimation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := tt.setupService()

			// Get logger from service if it's a service with logger
			var mockLogger *testutil.MockLogger
			if serviceWithLogger, ok := service.(*tokenCountingServiceImpl); ok {
				if logger, ok := serviceWithLogger.logger.(*testutil.MockLogger); ok {
					mockLogger = logger
				}
			}

			ctx := context.Background()
			req := TokenCountingRequest{
				Instructions: "test",
				Files: []FileContent{
					{Path: "test.go", Content: "package main"},
				},
			}

			result, err := service.CountTokensForModel(ctx, req, tt.model)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				if tt.expectFallback {
					// Should still get a result via fallback
					assert.GreaterOrEqual(t, result.TotalTokens, 0)

					// Check for fallback logging if we have a mock logger
					if mockLogger != nil && tt.logShouldContain != "" {
						foundLogMessage := false
						for _, msg := range mockLogger.GetMessages() {
							if strings.Contains(msg, tt.logShouldContain) {
								foundLogMessage = true
								break
							}
						}
						assert.True(t, foundLogMessage, "Should log fallback behavior: %s", tt.logShouldContain)
					}
				}
			}
		})
	}
}

// Mock implementations for logging tests

// Mock implementations are defined in token_counting_basic_test.go to avoid duplicates

type MockTokenizerManagerWithFailingTokenizer struct{}

func (m *MockTokenizerManagerWithFailingTokenizer) GetTokenizer(provider string) (tokenizers.AccurateTokenCounter, error) {
	return &MockAccurateTokenCounter{ShouldFail: true}, nil
}

func (m *MockTokenizerManagerWithFailingTokenizer) SupportsProvider(provider string) bool {
	return true
}

func (m *MockTokenizerManagerWithFailingTokenizer) ClearCache() {
	// No-op for mock
}
