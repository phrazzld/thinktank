package cli

import (
	"context"
	"testing"

	"github.com/misty-step/thinktank/internal/auditlog"
	"github.com/misty-step/thinktank/internal/logutil"
	"github.com/misty-step/thinktank/internal/thinktank"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnhancedAuditLogging tests the audit logging enhancements for Phase 7.2
func TestEnhancedAuditLogging(t *testing.T) {
	tests := []struct {
		name                 string
		compatibleModels     []thinktank.ModelCompatibility
		expectedOperation    string
		expectedStatus       string
		expectedInputFields  []string
		expectedOutputFields []string
		expectedTokenCounts  bool
	}{
		{
			name: "successful model selection with multiple compatible models",
			compatibleModels: []thinktank.ModelCompatibility{
				{
					ModelName:     "gpt-4",
					IsCompatible:  true,
					TokenCount:    50000,
					ContextWindow: 128000,
					UsableContext: 102400,
					Provider:      "openai",
					TokenizerUsed: "tiktoken",
					IsAccurate:    true,
				},
				{
					ModelName:     "gemini-1.5-flash",
					IsCompatible:  true,
					TokenCount:    50000,
					ContextWindow: 1000000,
					UsableContext: 800000,
					Provider:      "gemini",
					TokenizerUsed: "sentencepiece",
					IsAccurate:    true,
				},
				{
					ModelName:     "claude-3-sonnet",
					IsCompatible:  false,
					TokenCount:    50000,
					ContextWindow: 200000,
					UsableContext: 160000,
					Provider:      "openrouter",
					TokenizerUsed: "estimation",
					IsAccurate:    false,
					Reason:        "requires 50000 tokens but model only has 40000 usable tokens",
				},
			},
			expectedOperation:    "model_selection",
			expectedStatus:       "Success",
			expectedInputFields:  []string{"correlation_id", "instruction_tokens", "file_tokens", "total_tokens", "provider_count"},
			expectedOutputFields: []string{"selected_models", "skipped_models", "tokenizer_method", "compatible_count", "accurate_count"},
			expectedTokenCounts:  true,
		},
		{
			name: "no compatible models found",
			compatibleModels: []thinktank.ModelCompatibility{
				{
					ModelName:     "gpt-4",
					IsCompatible:  false,
					TokenCount:    200000,
					ContextWindow: 128000,
					UsableContext: 102400,
					Provider:      "openai",
					TokenizerUsed: "tiktoken",
					IsAccurate:    true,
					Reason:        "requires 200000 tokens but model only has 102400 usable tokens",
				},
			},
			expectedOperation:    "model_selection",
			expectedStatus:       "Failure",
			expectedInputFields:  []string{"correlation_id", "instruction_tokens", "file_tokens", "total_tokens", "provider_count"},
			expectedOutputFields: []string{"selected_models", "skipped_models", "tokenizer_method", "compatible_count", "accurate_count"},
			expectedTokenCounts:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock audit logger
			mockAuditLogger := &MockAuditLogger{}

			// Create context with correlation ID
			ctx := logutil.WithCorrelationID(context.Background())

			// Create token counting request for test
			tokenReq := thinktank.TokenCountingRequest{
				Instructions:        "Test instructions",
				Files:               []thinktank.FileContent{},
				SafetyMarginPercent: 20,
			}

			// Call the function under test
			err := LogModelSelectionAudit(ctx, mockAuditLogger, tokenReq, tt.compatibleModels, nil)

			// Verify no error
			require.NoError(t, err)

			// Verify audit entry was logged
			require.Len(t, mockAuditLogger.LoggedEntries, 1)
			entry := mockAuditLogger.LoggedEntries[0]

			// Verify basic audit entry fields
			assert.Equal(t, tt.expectedOperation, entry.Operation)
			assert.Equal(t, tt.expectedStatus, entry.Status)

			// Verify correlation ID is present
			correlationID := logutil.GetCorrelationID(ctx)
			assert.Equal(t, correlationID, entry.Inputs["correlation_id"])

			// Verify expected input fields are present
			for _, field := range tt.expectedInputFields {
				assert.Contains(t, entry.Inputs, field, "Missing input field: %s", field)
			}

			// Verify expected output fields are present
			for _, field := range tt.expectedOutputFields {
				assert.Contains(t, entry.Outputs, field, "Missing output field: %s", field)
			}

			// Verify token counts if expected
			if tt.expectedTokenCounts {
				assert.NotNil(t, entry.TokenCounts)
				assert.Greater(t, entry.TokenCounts.TotalTokens, int32(0))
			}

			// Verify model-specific fields based on test case
			selectedModels := entry.Outputs["selected_models"].([]string)
			skippedModels := entry.Outputs["skipped_models"].([]string)

			compatibleCount := 0
			for _, model := range tt.compatibleModels {
				if model.IsCompatible {
					compatibleCount++
					assert.Contains(t, selectedModels, model.ModelName)
				} else {
					assert.Contains(t, skippedModels, model.ModelName)
				}
			}

			assert.Equal(t, compatibleCount, entry.Outputs["compatible_count"])
		})
	}
}

// MockAuditLogger implements auditlog.AuditLogger for testing
type MockAuditLogger struct {
	LoggedEntries []auditlog.AuditEntry
}

func (m *MockAuditLogger) Log(ctx context.Context, entry auditlog.AuditEntry) error {
	m.LoggedEntries = append(m.LoggedEntries, entry)
	return nil
}

func (m *MockAuditLogger) LogLegacy(entry auditlog.AuditEntry) error {
	return m.Log(context.Background(), entry)
}

func (m *MockAuditLogger) LogOp(ctx context.Context, operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	entry := auditlog.AuditEntry{
		Operation: operation,
		Status:    status,
		Inputs:    inputs,
		Outputs:   outputs,
	}
	return m.Log(ctx, entry)
}

func (m *MockAuditLogger) LogOpLegacy(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
	return m.LogOp(context.Background(), operation, status, inputs, outputs, err)
}

func (m *MockAuditLogger) Close() error {
	return nil
}

// TestStructuredLogging tests the structured logging functionality for model selection
func TestStructuredLogging(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer []byte
	mockWriter := &MockWriter{Buffer: &logBuffer}

	// Create logger that writes to our mock writer
	logger := logutil.NewSlogLoggerFromLogLevel(mockWriter, logutil.InfoLevel)

	// Create context with correlation ID
	ctx := logutil.WithCorrelationID(context.Background())

	// Create test data
	tokenReq := thinktank.TokenCountingRequest{
		Instructions:        "Test instructions for structured logging",
		Files:               []thinktank.FileContent{},
		SafetyMarginPercent: 20,
	}

	compatibleModels := []thinktank.ModelCompatibility{
		{
			ModelName:     "gpt-4",
			IsCompatible:  true,
			TokenCount:    25000,
			ContextWindow: 128000,
			Provider:      "openai",
			TokenizerUsed: "tiktoken",
			IsAccurate:    true,
		},
		{
			ModelName:     "gemini-1.5-flash",
			IsCompatible:  false,
			TokenCount:    25000,
			ContextWindow: 50000,
			Provider:      "gemini",
			TokenizerUsed: "sentencepiece",
			IsAccurate:    true,
			Reason:        "insufficient context window",
		},
	}

	// Call the structured logging function
	err := LogModelSelectionStructured(ctx, logger, tokenReq, compatibleModels, nil)
	require.NoError(t, err)

	// Convert buffer to string for verification
	logOutput := string(logBuffer)

	// Verify that the log contains the correlation ID
	correlationID := logutil.GetCorrelationID(ctx)
	assert.Contains(t, logOutput, correlationID, "Log should contain correlation ID")

	// Verify that the log contains required fields from Phase 7.2
	assert.Contains(t, logOutput, "Token counting summary", "Log should contain summary message")
	assert.Contains(t, logOutput, "input_tokens", "Log should contain input_tokens field")
	assert.Contains(t, logOutput, "tiktoken", "Log should contain tokenizer_method")
	assert.Contains(t, logOutput, "selected_models", "Log should contain selected_models field")
	assert.Contains(t, logOutput, "skipped_models", "Log should contain skipped_models field")
	assert.Contains(t, logOutput, "gpt-4", "Log should contain selected model name")
	assert.Contains(t, logOutput, "gemini-1.5-flash", "Log should contain skipped model name")
}

// MockWriter implements io.Writer for testing
type MockWriter struct {
	Buffer *[]byte
}

func (mw *MockWriter) Write(p []byte) (n int, err error) {
	*mw.Buffer = append(*mw.Buffer, p...)
	return len(p), nil
}
