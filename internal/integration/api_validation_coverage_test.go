// Package integration provides comprehensive coverage tests for API validation and error classification
// Following TDD principles to target high-value business logic with 0% coverage
package integration

import (
	"context"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBoundaryAPIServiceValidation tests uncovered API validation methods
// Targets: ValidateModelParameter, GetModelDefinition, GetModelTokenLimits (all 0% coverage)
func TestBoundaryAPIServiceValidation(t *testing.T) {
	apiCaller := &MockExternalAPICaller{}
	envProvider := NewMockEnvironmentProvider()
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "test")
	service := NewBoundaryAPIService(apiCaller, envProvider, logger)

	t.Run("ValidateModelParameter with valid temperature", func(t *testing.T) {
		valid, err := service.ValidateModelParameter(context.Background(), "gpt-4", "temperature", 0.7)
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("ValidateModelParameter with invalid temperature", func(t *testing.T) {
		valid, err := service.ValidateModelParameter(context.Background(), "gpt-4", "temperature", 2.5)
		assert.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "temperature must be")
	})

	t.Run("ValidateModelParameter with valid max_tokens", func(t *testing.T) {
		valid, err := service.ValidateModelParameter(context.Background(), "gpt-4", "max_tokens", 1000)
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("ValidateModelParameter with invalid max_tokens", func(t *testing.T) {
		valid, err := service.ValidateModelParameter(context.Background(), "gpt-4", "max_tokens", -100)
		assert.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "max_tokens must be")
	})

	t.Run("ValidateModelParameter with unknown parameter", func(t *testing.T) {
		valid, err := service.ValidateModelParameter(context.Background(), "gpt-4", "unknown_param", "value")
		require.NoError(t, err)
		assert.True(t, valid) // Unknown parameters are allowed
	})

	t.Run("GetModelDefinition with gpt model", func(t *testing.T) {
		definition, err := service.GetModelDefinition(context.Background(), "gpt-4")
		require.NoError(t, err)
		assert.NotNil(t, definition)
		assert.Equal(t, "gpt-4", definition.APIModelID)
		assert.Equal(t, "openai", definition.Provider)
	})

	t.Run("GetModelDefinition with gemini model", func(t *testing.T) {
		definition, err := service.GetModelDefinition(context.Background(), "gemini-pro")
		require.NoError(t, err)
		assert.NotNil(t, definition)
		assert.Equal(t, "gemini-pro", definition.APIModelID)
		assert.Equal(t, "gemini", definition.Provider)
	})

	t.Run("GetModelDefinition with unknown model", func(t *testing.T) {
		definition, err := service.GetModelDefinition(context.Background(), "unknown-model")
		require.NoError(t, err) // Function always succeeds, returns unknown provider
		assert.NotNil(t, definition)
		assert.Equal(t, "unknown-model", definition.APIModelID)
		assert.Equal(t, "unknown", definition.Provider)
	})

	t.Run("GetModelTokenLimits with gpt model", func(t *testing.T) {
		contextWindow, maxOutput, err := service.GetModelTokenLimits(context.Background(), "gpt-4")
		require.NoError(t, err)
		assert.Greater(t, contextWindow, int32(0))
		assert.Greater(t, maxOutput, int32(0))
	})

	t.Run("GetModelTokenLimits with gemini model", func(t *testing.T) {
		contextWindow, maxOutput, err := service.GetModelTokenLimits(context.Background(), "gemini-pro")
		require.NoError(t, err)
		assert.Greater(t, contextWindow, int32(0))
		assert.Greater(t, maxOutput, int32(0))
	})

	t.Run("GetModelTokenLimits with unknown model", func(t *testing.T) {
		contextWindow, maxOutput, err := service.GetModelTokenLimits(context.Background(), "unknown-model")
		require.NoError(t, err) // Function provides defaults
		assert.Greater(t, contextWindow, int32(0))
		assert.Greater(t, maxOutput, int32(0))
	})
}

// TestBoundaryAPIServiceErrorClassification tests uncovered error classification methods
// Targets: IsEmptyResponseError, IsSafetyBlockedError (both 0% coverage)
func TestBoundaryAPIServiceErrorClassification(t *testing.T) {
	apiCaller := &MockExternalAPICaller{}
	envProvider := NewMockEnvironmentProvider()
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "test")
	service := NewBoundaryAPIService(apiCaller, envProvider, logger)

	t.Run("IsEmptyResponseError with empty response error", func(t *testing.T) {
		err := &EmptyResponseError{Message: "response was empty"}
		isEmptyError := service.IsEmptyResponseError(err)
		assert.True(t, isEmptyError)
	})

	t.Run("IsEmptyResponseError with non-empty response error", func(t *testing.T) {
		err := &GenericAPIError{Message: "some other error"}
		isEmptyError := service.IsEmptyResponseError(err)
		assert.False(t, isEmptyError)
	})

	t.Run("IsEmptyResponseError with nil error", func(t *testing.T) {
		isEmptyError := service.IsEmptyResponseError(nil)
		assert.False(t, isEmptyError)
	})

	t.Run("IsSafetyBlockedError with safety blocked error", func(t *testing.T) {
		err := &SafetyBlockedError{Message: "content was blocked by safety filters"}
		isSafetyError := service.IsSafetyBlockedError(err)
		assert.True(t, isSafetyError)
	})

	t.Run("IsSafetyBlockedError with non-safety error", func(t *testing.T) {
		err := &GenericAPIError{Message: "some other error"}
		isSafetyError := service.IsSafetyBlockedError(err)
		assert.False(t, isSafetyError)
	})

	t.Run("IsSafetyBlockedError with nil error", func(t *testing.T) {
		isSafetyError := service.IsSafetyBlockedError(nil)
		assert.False(t, isSafetyError)
	})
}

// TestBoundaryLLMClientGetModelName tests uncovered model name getter
// Targets: GetModelName (0% coverage)
func TestBoundaryLLMClientGetModelName(t *testing.T) {
	client := &BoundaryLLMClient{
		apiCaller: &MockExternalAPICaller{},
		modelName: "test-model-name",
		apiKey:    "test-key",
		endpoint:  "test-endpoint",
		logger:    logutil.NewLogger(logutil.InfoLevel, nil, "test"),
	}

	t.Run("GetModelName returns configured model name", func(t *testing.T) {
		modelName := client.GetModelName()
		assert.Equal(t, "test-model-name", modelName)
	})

	t.Run("GetModelName with empty model name", func(t *testing.T) {
		client := &BoundaryLLMClient{
			apiCaller: &MockExternalAPICaller{},
			modelName: "",
			apiKey:    "test-key",
			endpoint:  "test-endpoint",
			logger:    logutil.NewLogger(logutil.InfoLevel, nil, "test"),
		}
		modelName := client.GetModelName()
		assert.Equal(t, "", modelName)
	})
}

// Error types for testing error classification
type EmptyResponseError struct {
	Message string
}

func (e *EmptyResponseError) Error() string {
	return e.Message
}

type SafetyBlockedError struct {
	Message string
}

func (e *SafetyBlockedError) Error() string {
	return e.Message
}

type GenericAPIError struct {
	Message string
}

func (e *GenericAPIError) Error() string {
	return e.Message
}
