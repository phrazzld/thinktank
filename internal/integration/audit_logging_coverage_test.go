// Package integration provides comprehensive coverage tests for audit logging and test utilities
// Following TDD principles to target remaining 0% coverage functions
package integration

import (
	"context"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBoundaryAuditLoggerMethods tests uncovered audit logger methods
// Targets: LogLegacy, LogOpLegacy, Close (all 0% coverage)
func TestBoundaryAuditLoggerMethods(t *testing.T) {
	filesystem := NewMockFilesystemIO()
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "test")
	auditLogger := NewBoundaryAuditLogger(filesystem, logger)

	t.Run("LogLegacy with valid entry", func(t *testing.T) {
		entry := auditlog.AuditEntry{
			Operation: "test_operation",
			Status:    "success",
			Message:   "test message",
		}
		err := auditLogger.LogLegacy(entry)
		assert.NoError(t, err, "LogLegacy should handle legacy format successfully")
	})

	t.Run("LogOpLegacy with valid parameters", func(t *testing.T) {
		inputs := map[string]interface{}{
			"param1": "value1",
			"param2": 42,
		}
		outputs := map[string]interface{}{
			"result": "success",
		}

		err := auditLogger.LogOpLegacy("test_operation", "success", inputs, outputs, nil)
		assert.NoError(t, err, "LogOpLegacy should handle legacy operation format successfully")
	})

	t.Run("LogOpLegacy with nil inputs and outputs", func(t *testing.T) {
		err := auditLogger.LogOpLegacy("test_operation", "success", nil, nil, nil)
		assert.NoError(t, err, "LogOpLegacy should handle nil inputs and outputs")
	})

	t.Run("LogOpLegacy with error", func(t *testing.T) {
		testError := assert.AnError
		err := auditLogger.LogOpLegacy("test_operation", "failure", nil, nil, testError)
		assert.NoError(t, err, "LogOpLegacy should handle operations with errors")
	})

	t.Run("Close releases resources", func(t *testing.T) {
		err := auditLogger.Close()
		assert.NoError(t, err, "Close should successfully release resources")
	})

	t.Run("Close is idempotent", func(t *testing.T) {
		filesystem2 := NewMockFilesystemIO()
		logger2 := logutil.NewLogger(logutil.InfoLevel, nil, "test")
		auditLogger2 := NewBoundaryAuditLogger(filesystem2, logger2)

		err := auditLogger2.Close()
		assert.NoError(t, err, "First close should succeed")

		err = auditLogger2.Close()
		assert.NoError(t, err, "Second close should also succeed (idempotent)")
	})
}

// TestTestUtilities tests uncovered test utility functions
// Targets: NewTestEnv, MockResponse (both 0% coverage)
func TestTestUtilities(t *testing.T) {
	t.Run("NewTestEnv creates valid test environment", func(t *testing.T) {
		env := NewTestEnv()

		require.NotNil(t, env, "NewTestEnv should return non-nil environment")
		require.NotNil(t, env.MockClient, "Test environment should have mock client")

		// Test that the mock client works
		result, err := env.MockClient.GenerateContentFunc(context.Background(), "test prompt", nil)
		require.NoError(t, err)
		assert.Equal(t, "Test content", result.Content)
		assert.Equal(t, "stop", result.FinishReason)
		assert.False(t, result.Truncated)

		// Test model name function
		modelName := env.MockClient.GetModelNameFunc()
		assert.Equal(t, "gemini-pro", modelName)
	})

	t.Run("MockResponse creates valid response", func(t *testing.T) {
		content := "test response content"
		response := MockResponse(content)

		require.NotNil(t, response, "MockResponse should return non-nil response")
		assert.Equal(t, content, response.Content)
	})

	t.Run("MockResponse with empty content", func(t *testing.T) {
		response := MockResponse("")

		require.NotNil(t, response, "MockResponse should handle empty content")
		assert.Equal(t, "", response.Content)
	})

	t.Run("MockResponse with special characters", func(t *testing.T) {
		content := "test\nwith\tspecial\rcharacters"
		response := MockResponse(content)

		require.NotNil(t, response, "MockResponse should handle special characters")
		assert.Equal(t, content, response.Content)
	})
}

// TestTestRunner tests uncovered test runner functionality
// Targets: NewTestRunner, RunTest (both 0% coverage)
func TestTestRunner(t *testing.T) {
	t.Run("NewTestRunner creates valid runner", func(t *testing.T) {
		runner := NewTestRunner()

		require.NotNil(t, runner, "NewTestRunner should return non-nil runner")
		// Test that the runner can be used for basic operations
		assert.NotPanics(t, func() {
			_ = runner.RunTest()
		}, "NewTestRunner should create usable runner")
	})

	t.Run("RunTest basic execution", func(t *testing.T) {
		runner := NewTestRunner()

		result := runner.RunTest()

		// Test runner should handle basic execution without panicking
		require.NotNil(t, result, "RunTest should return non-nil result")
		assert.Equal(t, "Test successful", result.Content)
	})
}

// TestLLMClientAdapter tests uncovered adapter methods
// Targets: NewLLMClientAdapter, GenerateContent, GetModelName, Close (all 0% coverage)
func TestLLMClientAdapter(t *testing.T) {
	t.Run("NewLLMClientAdapter creates valid adapter", func(t *testing.T) {
		testEnv := NewTestEnv()
		adapter := NewLLMClientAdapter(testEnv.MockClient, "test-model")

		require.NotNil(t, adapter, "NewLLMClientAdapter should return non-nil adapter")
	})

	t.Run("LLMClientAdapter GenerateContent", func(t *testing.T) {
		testEnv := NewTestEnv()
		adapter := NewLLMClientAdapter(testEnv.MockClient, "test-model")

		result, err := adapter.GenerateContent(context.Background(), "test prompt", map[string]interface{}{
			"temperature": 0.7,
		})

		require.NoError(t, err, "GenerateContent should not error with valid input")
		require.NotNil(t, result, "GenerateContent should return non-nil result")
		assert.NotEmpty(t, result.Content, "Result should have content")
	})

	t.Run("LLMClientAdapter GetModelName", func(t *testing.T) {
		expectedModel := "test-model-name"
		testEnv := NewTestEnv()
		adapter := NewLLMClientAdapter(testEnv.MockClient, expectedModel)

		modelName := adapter.GetModelName()
		assert.Equal(t, expectedModel, modelName)
	})

	t.Run("LLMClientAdapter Close", func(t *testing.T) {
		testEnv := NewTestEnv()
		adapter := NewLLMClientAdapter(testEnv.MockClient, "test-model")

		err := adapter.Close()
		assert.NoError(t, err, "Close should succeed")
	})

	t.Run("LLMClientAdapter Close is idempotent", func(t *testing.T) {
		testEnv := NewTestEnv()
		adapter := NewLLMClientAdapter(testEnv.MockClient, "test-model")

		err := adapter.Close()
		assert.NoError(t, err, "First close should succeed")

		err = adapter.Close()
		assert.NoError(t, err, "Second close should also succeed")
	})
}
