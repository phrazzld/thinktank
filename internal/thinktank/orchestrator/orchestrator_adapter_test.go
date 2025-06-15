package orchestrator

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/registry"
)

// ConfigurableAPIService provides a configurable mock implementation for testing APIServiceAdapter
type ConfigurableAPIService struct {
	isEmptyResponseError  bool
	isSafetyBlockedError  bool
	errorDetails          string
	modelParameters       map[string]interface{}
	modelDefinition       *registry.ModelDefinition
	contextWindow         int32
	maxOutputTokens       int32
	tokenLimitsError      error
	paramValidationResult bool
	paramValidationError  error
}

func (m *ConfigurableAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	return nil, nil
}

func (m *ConfigurableAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	return "", nil
}

func (m *ConfigurableAPIService) IsEmptyResponseError(err error) bool {
	return m.isEmptyResponseError
}

func (m *ConfigurableAPIService) IsSafetyBlockedError(err error) bool {
	return m.isSafetyBlockedError
}

func (m *ConfigurableAPIService) GetErrorDetails(err error) string {
	return m.errorDetails
}

func (m *ConfigurableAPIService) GetModelParameters(ctx context.Context, modelName string) (map[string]interface{}, error) {
	return m.modelParameters, nil
}

func (m *ConfigurableAPIService) GetModelDefinition(ctx context.Context, modelName string) (*registry.ModelDefinition, error) {
	return m.modelDefinition, nil
}

func (m *ConfigurableAPIService) GetModelTokenLimits(ctx context.Context, modelName string) (contextWindow, maxOutputTokens int32, err error) {
	return m.contextWindow, m.maxOutputTokens, m.tokenLimitsError
}

func (m *ConfigurableAPIService) ValidateModelParameter(ctx context.Context, modelName, paramName string, value interface{}) (bool, error) {
	return m.paramValidationResult, m.paramValidationError
}

// TestOutputWriter provides a configurable mock implementation for testing LegacyOutputWriterAdapter
type TestOutputWriter struct {
	saveIndividualCount int
	saveIndividualPaths map[string]string
	saveIndividualError error
	saveSynthesisPath   string
	saveSynthesisError  error
}

func (m *TestOutputWriter) SaveIndividualOutputs(ctx context.Context, modelOutputs map[string]string, outputDir string) (int, map[string]string, error) {
	return m.saveIndividualCount, m.saveIndividualPaths, m.saveIndividualError
}

func (m *TestOutputWriter) SaveSynthesisOutput(ctx context.Context, content string, modelName string, outputDir string) (string, error) {
	return m.saveSynthesisPath, m.saveSynthesisError
}

// TestAPIServiceAdapter_IsEmptyResponseError tests the IsEmptyResponseError delegation
func TestAPIServiceAdapter_IsEmptyResponseError(t *testing.T) {
	tests := []struct {
		name           string
		mockResult     bool
		expectedResult bool
	}{
		{
			name:           "empty response error detected",
			mockResult:     true,
			expectedResult: true,
		},
		{
			name:           "not an empty response error",
			mockResult:     false,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &ConfigurableAPIService{
				isEmptyResponseError: tt.mockResult,
			}
			adapter := &APIServiceAdapter{APIService: mockAPI}

			testErr := errors.New("test error")
			result := adapter.IsEmptyResponseError(testErr)

			if result != tt.expectedResult {
				t.Errorf("IsEmptyResponseError() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// TestAPIServiceAdapter_IsSafetyBlockedError tests the IsSafetyBlockedError delegation
func TestAPIServiceAdapter_IsSafetyBlockedError(t *testing.T) {
	tests := []struct {
		name           string
		mockResult     bool
		expectedResult bool
	}{
		{
			name:           "safety blocked error detected",
			mockResult:     true,
			expectedResult: true,
		},
		{
			name:           "not a safety blocked error",
			mockResult:     false,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := &ConfigurableAPIService{
				isSafetyBlockedError: tt.mockResult,
			}
			adapter := &APIServiceAdapter{APIService: mockAPI}

			testErr := errors.New("test error")
			result := adapter.IsSafetyBlockedError(testErr)

			if result != tt.expectedResult {
				t.Errorf("IsSafetyBlockedError() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// TestAPIServiceAdapter_GetModelDefinition tests the GetModelDefinition delegation
func TestAPIServiceAdapter_GetModelDefinition(t *testing.T) {
	ctx := context.Background()
	modelName := "test-model"

	expectedDefinition := &registry.ModelDefinition{
		Name:     modelName,
		Provider: "test-provider",
	}

	mockAPI := &ConfigurableAPIService{
		modelDefinition: expectedDefinition,
	}
	adapter := &APIServiceAdapter{APIService: mockAPI}

	result, err := adapter.GetModelDefinition(ctx, modelName)

	if err != nil {
		t.Errorf("GetModelDefinition() error = %v, want nil", err)
	}

	if result != expectedDefinition {
		t.Errorf("GetModelDefinition() = %v, want %v", result, expectedDefinition)
	}
}

// TestAPIServiceAdapter_GetModelTokenLimits tests the GetModelTokenLimits delegation
func TestAPIServiceAdapter_GetModelTokenLimits(t *testing.T) {
	tests := []struct {
		name            string
		contextWindow   int32
		maxOutputTokens int32
		expectedError   error
		expectError     bool
	}{
		{
			name:            "successful token limits retrieval",
			contextWindow:   4096,
			maxOutputTokens: 1024,
			expectedError:   nil,
			expectError:     false,
		},
		{
			name:            "error retrieving token limits",
			contextWindow:   0,
			maxOutputTokens: 0,
			expectedError:   errors.New("model not found"),
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			modelName := "test-model"

			mockAPI := &ConfigurableAPIService{
				contextWindow:    tt.contextWindow,
				maxOutputTokens:  tt.maxOutputTokens,
				tokenLimitsError: tt.expectedError,
			}
			adapter := &APIServiceAdapter{APIService: mockAPI}

			contextWindow, maxOutputTokens, err := adapter.GetModelTokenLimits(ctx, modelName)

			if tt.expectError {
				if err == nil {
					t.Errorf("GetModelTokenLimits() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("GetModelTokenLimits() error = %v, want nil", err)
				}
				if contextWindow != tt.contextWindow {
					t.Errorf("GetModelTokenLimits() contextWindow = %v, want %v", contextWindow, tt.contextWindow)
				}
				if maxOutputTokens != tt.maxOutputTokens {
					t.Errorf("GetModelTokenLimits() maxOutputTokens = %v, want %v", maxOutputTokens, tt.maxOutputTokens)
				}
			}
		})
	}
}

// TestAPIServiceAdapter_ValidateModelParameter tests the ValidateModelParameter delegation
func TestAPIServiceAdapter_ValidateModelParameter(t *testing.T) {
	tests := []struct {
		name          string
		validResult   bool
		expectedError error
		expectError   bool
	}{
		{
			name:          "parameter validation successful",
			validResult:   true,
			expectedError: nil,
			expectError:   false,
		},
		{
			name:          "parameter validation failed",
			validResult:   false,
			expectedError: nil,
			expectError:   false,
		},
		{
			name:          "parameter validation error",
			validResult:   false,
			expectedError: errors.New("validation error"),
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			modelName := "test-model"
			paramName := "temperature"
			value := 0.7

			mockAPI := &ConfigurableAPIService{
				paramValidationResult: tt.validResult,
				paramValidationError:  tt.expectedError,
			}
			adapter := &APIServiceAdapter{APIService: mockAPI}

			valid, err := adapter.ValidateModelParameter(ctx, modelName, paramName, value)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateModelParameter() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateModelParameter() error = %v, want nil", err)
				}
				if valid != tt.validResult {
					t.Errorf("ValidateModelParameter() = %v, want %v", valid, tt.validResult)
				}
			}
		})
	}
}

// TestLegacyOutputWriterAdapter_SaveIndividualOutputs tests the legacy adapter
func TestLegacyOutputWriterAdapter_SaveIndividualOutputs(t *testing.T) {
	tests := []struct {
		name          string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "successful save",
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "save with error",
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			modelOutputs := map[string]string{
				"model1": "output1",
				"model2": "output2",
			}
			outputDir := "/test/output"

			mockOutputWriter := &TestOutputWriter{
				saveIndividualCount: tt.expectedCount,
				saveIndividualPaths: map[string]string{
					"model1": "/test/output/model1.txt",
					"model2": "/test/output/model2.txt",
				},
			}

			if tt.expectError {
				mockOutputWriter.saveIndividualError = errors.New("save error")
			}

			adapter := &LegacyOutputWriterAdapter{outputWriter: mockOutputWriter}

			count, err := adapter.SaveIndividualOutputs(ctx, modelOutputs, outputDir)

			if tt.expectError {
				if err == nil {
					t.Errorf("SaveIndividualOutputs() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("SaveIndividualOutputs() error = %v, want nil", err)
				}
				if count != tt.expectedCount {
					t.Errorf("SaveIndividualOutputs() count = %v, want %v", count, tt.expectedCount)
				}
			}
		})
	}
}

// TestLegacyOutputWriterAdapter_SaveSynthesisOutput tests the legacy adapter
func TestLegacyOutputWriterAdapter_SaveSynthesisOutput(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "successful save",
			expectError: false,
		},
		{
			name:        "save with error",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			content := "synthesis content"
			modelName := "synthesis-model"
			outputDir := "/test/output"

			mockOutputWriter := &TestOutputWriter{
				saveSynthesisPath: "/test/output/synthesis.txt",
			}

			if tt.expectError {
				mockOutputWriter.saveSynthesisError = errors.New("save error")
			}

			adapter := &LegacyOutputWriterAdapter{outputWriter: mockOutputWriter}

			err := adapter.SaveSynthesisOutput(ctx, content, modelName, outputDir)

			if tt.expectError {
				if err == nil {
					t.Errorf("SaveSynthesisOutput() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("SaveSynthesisOutput() error = %v, want nil", err)
				}
			}
		})
	}
}

// TestCategorizeOrchestratorError tests the error categorization function
func TestCategorizeOrchestratorError(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		expectedCategory llm.ErrorCategory
	}{
		{
			name:             "ErrInvalidSynthesisModel",
			err:              ErrInvalidSynthesisModel,
			expectedCategory: llm.CategoryInvalidRequest,
		},
		{
			name:             "ErrNoValidModels",
			err:              ErrNoValidModels,
			expectedCategory: llm.CategoryInvalidRequest,
		},
		{
			name:             "ErrPartialProcessingFailure",
			err:              ErrPartialProcessingFailure,
			expectedCategory: llm.CategoryServer,
		},
		{
			name:             "ErrAllProcessingFailed",
			err:              ErrAllProcessingFailed,
			expectedCategory: llm.CategoryServer,
		},
		{
			name:             "ErrSynthesisFailed",
			err:              ErrSynthesisFailed,
			expectedCategory: llm.CategoryServer,
		},
		{
			name:             "ErrOutputFileSaveFailed",
			err:              ErrOutputFileSaveFailed,
			expectedCategory: llm.CategoryServer,
		},
		{
			name:             "ErrModelProcessingCancelled",
			err:              ErrModelProcessingCancelled,
			expectedCategory: llm.CategoryCancelled,
		},
		{
			name:             "wrapped sentinel error",
			err:              errors.New("wrapper: " + ErrInvalidSynthesisModel.Error()),
			expectedCategory: llm.CategoryUnknown, // wrapped errors won't match errors.Is()
		},
		{
			name:             "LLM error with category",
			err:              llm.New("test", "CODE", 401, "message", "req123", errors.New("cause"), llm.CategoryAuth),
			expectedCategory: llm.CategoryAuth,
		},
		{
			name:             "generic error",
			err:              errors.New("generic error"),
			expectedCategory: llm.CategoryUnknown,
		},
		{
			name:             "nil error",
			err:              nil,
			expectedCategory: llm.CategoryUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := CategorizeOrchestratorError(tt.err)
			if category != tt.expectedCategory {
				t.Errorf("CategorizeOrchestratorError(%v) = %v, want %v", tt.err, category, tt.expectedCategory)
			}
		})
	}
}

// TestWrapOrchestratorError tests the error wrapping function
func TestWrapOrchestratorError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		message       string
		expectNil     bool
		expectWrapped bool
	}{
		{
			name:          "nil error",
			err:           nil,
			message:       "test message",
			expectNil:     true,
			expectWrapped: false,
		},
		{
			name:          "wrap sentinel error",
			err:           ErrInvalidSynthesisModel,
			message:       "test context",
			expectNil:     false,
			expectWrapped: true,
		},
		{
			name:          "wrap generic error",
			err:           errors.New("generic error"),
			message:       "test context",
			expectNil:     false,
			expectWrapped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapOrchestratorError(tt.err, tt.message)

			if tt.expectNil {
				if result != nil {
					t.Errorf("WrapOrchestratorError() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Errorf("WrapOrchestratorError() = nil, want non-nil error")
				return
			}

			if tt.expectWrapped {
				// Verify that the original error is wrapped
				if !errors.Is(result, tt.err) {
					t.Errorf("WrapOrchestratorError() result should wrap original error %v", tt.err)
				}

				// Verify that the message is included
				if !strings.Contains(result.Error(), tt.message) {
					t.Errorf("WrapOrchestratorError() result should contain message %q, got %q", tt.message, result.Error())
				}
			}
		})
	}
}
