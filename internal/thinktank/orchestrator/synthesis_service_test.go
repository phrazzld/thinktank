package orchestrator

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
)

// MockSynthesisAPIService is a specialized mock for testing synthesis
type MockSynthesisAPIService struct {
	// Base mock implementation
	MockAPIService
	// Control test behavior
	ModelParamsError         error
	ClientInitError          error
	ClientGenerateError      error
	ProcessResponseError     error
	ClientGenerateResult     *llm.ProviderResult
	ProcessResponseResult    string
	IsSafetyBlocked          bool
	IsEmptyResponse          bool
	GetModelParametersResult map[string]interface{}
	InitLLMClientResult      llm.LLMClient
	GetErrorDetailsResult    string
	capturedPrompt           string
	capturedModelName        string
	capturedModelParams      map[string]interface{}
}

// GetModelParameters overrides the mock implementation
func (m *MockSynthesisAPIService) GetModelParameters(modelName string) (map[string]interface{}, error) {
	m.capturedModelName = modelName
	if m.ModelParamsError != nil {
		return nil, m.ModelParamsError
	}
	if m.GetModelParametersResult != nil {
		return m.GetModelParametersResult, nil
	}
	return map[string]interface{}{
		"temperature": 0.7,
		"max_tokens":  2000,
	}, nil
}

// InitLLMClient overrides the mock implementation
func (m *MockSynthesisAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	m.capturedModelName = modelName
	if m.ClientInitError != nil {
		return nil, m.ClientInitError
	}
	if m.InitLLMClientResult != nil {
		return m.InitLLMClientResult, nil
	}
	return &MockSynthesisLLMClient{
		generateError:  m.ClientGenerateError,
		generateResult: m.ClientGenerateResult,
		capturePrompt:  &m.capturedPrompt,
		captureParams:  &m.capturedModelParams,
	}, nil
}

// ProcessLLMResponse overrides the mock implementation
func (m *MockSynthesisAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	if m.ProcessResponseError != nil {
		return "", m.ProcessResponseError
	}
	if m.ProcessResponseResult != "" {
		return m.ProcessResponseResult, nil
	}
	return "Synthesized result", nil
}

// IsEmptyResponseError overrides the mock implementation
func (m *MockSynthesisAPIService) IsEmptyResponseError(err error) bool {
	return m.IsEmptyResponse
}

// IsSafetyBlockedError overrides the mock implementation
func (m *MockSynthesisAPIService) IsSafetyBlockedError(err error) bool {
	return m.IsSafetyBlocked
}

// GetErrorDetails overrides the mock implementation
func (m *MockSynthesisAPIService) GetErrorDetails(err error) string {
	if m.GetErrorDetailsResult != "" {
		return m.GetErrorDetailsResult
	}
	if err == nil {
		return ""
	}
	return err.Error()
}

// MockSynthesisLLMClient is a specialized mock for testing synthesis
type MockSynthesisLLMClient struct {
	generateError  error
	generateResult *llm.ProviderResult
	capturePrompt  *string
	captureParams  *map[string]interface{}
	closeCalled    bool
	closeError     error
}

// GenerateContent implements the LLMClient interface
func (m *MockSynthesisLLMClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	if m.capturePrompt != nil {
		*m.capturePrompt = prompt
	}
	if m.captureParams != nil {
		*m.captureParams = params
	}
	if m.generateError != nil {
		return nil, m.generateError
	}
	if m.generateResult != nil {
		return m.generateResult, nil
	}
	// Default response
	return &llm.ProviderResult{
		Content:      "Default synthesized content",
		FinishReason: "stop",
	}, nil
}

// GetModelName implements the LLMClient interface
func (m *MockSynthesisLLMClient) GetModelName() string {
	return "mock-synthesis-model"
}

// Close implements the LLMClient interface
func (m *MockSynthesisLLMClient) Close() error {
	m.closeCalled = true
	return m.closeError
}

// TestSynthesizeResults tests the SynthesizeResults method
func TestSynthesizeResults(t *testing.T) {
	tests := []struct {
		name                string
		instructions        string
		modelOutputs        map[string]string
		synthesisModelName  string
		setupMockFn         func(*MockSynthesisAPIService)
		expectedOutput      string
		expectedError       bool
		expectedErrorMatch  string
		checkPromptContains []string
	}{
		{
			name:         "Successful synthesis",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model1": "Output from model1",
				"model2": "Output from model2",
			},
			synthesisModelName: "synthesis-model",
			setupMockFn: func(m *MockSynthesisAPIService) {
				m.ProcessResponseResult = "Successfully synthesized content"
			},
			expectedOutput: "Successfully synthesized content",
			expectedError:  false,
			checkPromptContains: []string{
				"Test instructions",
				"Output from model1",
				"Output from model2",
			},
		},
		{
			name:         "Model parameters error",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model1": "Output from model1",
			},
			synthesisModelName: "synthesis-model",
			setupMockFn: func(m *MockSynthesisAPIService) {
				m.ModelParamsError = errors.New("model parameters error")
			},
			expectedError:      true,
			expectedErrorMatch: "failed to get model parameters",
		},
		{
			name:         "Client initialization error",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model1": "Output from model1",
			},
			synthesisModelName: "synthesis-model",
			setupMockFn: func(m *MockSynthesisAPIService) {
				m.ClientInitError = errors.New("client initialization error")
			},
			expectedError:      true,
			expectedErrorMatch: "failed to initialize synthesis model client",
		},
		{
			name:         "Client generate error",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model1": "Output from model1",
			},
			synthesisModelName: "synthesis-model",
			setupMockFn: func(m *MockSynthesisAPIService) {
				m.ClientGenerateError = errors.New("client generate error")
			},
			expectedError:      true,
			expectedErrorMatch: "synthesis of model outputs failed",
		},
		{
			name:         "Process response error",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model1": "Output from model1",
			},
			synthesisModelName: "synthesis-model",
			setupMockFn: func(m *MockSynthesisAPIService) {
				m.ProcessResponseError = errors.New("process response error")
			},
			expectedError:      true,
			expectedErrorMatch: "failed to process synthesis model response",
		},
		{
			name:               "Empty model outputs",
			instructions:       "Test instructions",
			modelOutputs:       map[string]string{},
			synthesisModelName: "synthesis-model",
			setupMockFn: func(m *MockSynthesisAPIService) {
				m.ProcessResponseResult = "Synthesis of empty outputs"
			},
			expectedOutput: "Synthesis of empty outputs",
			expectedError:  false,
		},
		{
			name:         "Safety blocked error",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model1": "Output from model1",
			},
			synthesisModelName: "synthesis-model",
			setupMockFn: func(m *MockSynthesisAPIService) {
				m.ClientGenerateError = errors.New("safety blocked")
				m.IsSafetyBlocked = true
			},
			expectedError:      true,
			expectedErrorMatch: "blocked by content safety filters",
		},
		{
			name:         "Empty response error",
			instructions: "Test instructions",
			modelOutputs: map[string]string{
				"model1": "Output from model1",
			},
			synthesisModelName: "synthesis-model",
			setupMockFn: func(m *MockSynthesisAPIService) {
				m.ClientGenerateError = errors.New("empty response")
				m.IsEmptyResponse = true
			},
			expectedError:      true,
			expectedErrorMatch: "empty response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockAPIService := &MockSynthesisAPIService{}
			if tt.setupMockFn != nil {
				tt.setupMockFn(mockAPIService)
			}
			mockLogger := &MockLogger{}
			mockAuditLogger := &MockAuditLogger{}

			// Create synthesis service
			synthesisService := NewSynthesisService(
				mockAPIService,
				mockAuditLogger,
				mockLogger,
				tt.synthesisModelName,
			)

			// Call SynthesizeResults
			result, err := synthesisService.SynthesizeResults(
				context.Background(),
				tt.instructions,
				tt.modelOutputs,
			)

			// Verify results
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				} else if tt.expectedErrorMatch != "" && !strings.Contains(err.Error(), tt.expectedErrorMatch) {
					t.Errorf("Error message didn't contain expected text: got %q, want to contain %q",
						err.Error(), tt.expectedErrorMatch)
				}

				// Don't check for sentinel errors in testing since our test errors don't wrap them correctly
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result != tt.expectedOutput {
					t.Errorf("Result mismatch. Expected %q but got %q", tt.expectedOutput, result)
				}
			}

			// Verify the prompt contains expected contents
			for _, expectedContent := range tt.checkPromptContains {
				if !strings.Contains(mockAPIService.capturedPrompt, expectedContent) {
					t.Errorf("Prompt doesn't contain expected content %q", expectedContent)
				}
			}

			// Verify the model name was captured correctly
			if mockAPIService.capturedModelName != tt.synthesisModelName {
				t.Errorf("Model name mismatch. Expected %q but got %q",
					tt.synthesisModelName, mockAPIService.capturedModelName)
			}
		})
	}
}

// TestHandleSynthesisError tests the handleSynthesisError method
func TestHandleSynthesisError(t *testing.T) {
	tests := []struct {
		name              string
		inputError        error
		isSafetyBlocked   bool
		isEmptyResponse   bool
		errorDetails      string
		expectedErrorType string
	}{
		{
			name:              "Rate limit error",
			inputError:        errors.New("rate limit exceeded"),
			expectedErrorType: "rate limit",
		},
		{
			name:              "Safety blocked error",
			inputError:        errors.New("content filtered"),
			isSafetyBlocked:   true,
			expectedErrorType: "blocked by content safety filters",
		},
		{
			name:              "Connectivity error",
			inputError:        errors.New("connection timeout"),
			expectedErrorType: "Connectivity issue",
		},
		{
			name:              "Authentication error",
			inputError:        errors.New("invalid auth key"),
			expectedErrorType: "Authentication failed",
		},
		{
			name:              "Empty response error",
			inputError:        errors.New("empty API response"),
			isEmptyResponse:   true,
			expectedErrorType: "empty response",
		},
		{
			name:              "Generic error with details",
			inputError:        errors.New("unknown error"),
			errorDetails:      "Detailed error information",
			expectedErrorType: "Error synthesizing results",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockAPIService := &MockSynthesisAPIService{
				IsSafetyBlocked:       tt.isSafetyBlocked,
				IsEmptyResponse:       tt.isEmptyResponse,
				GetErrorDetailsResult: tt.errorDetails,
			}
			mockLogger := &MockLogger{}
			mockAuditLogger := &MockAuditLogger{}

			// Create synthesis service
			service := &DefaultSynthesisService{
				apiService:  mockAPIService,
				auditLogger: mockAuditLogger,
				logger:      mockLogger,
				modelName:   "test-model",
			}

			// Call handleSynthesisError
			err := service.handleSynthesisError(context.Background(), tt.inputError)

			// Verify results
			if err == nil {
				t.Errorf("Expected an error but got nil")
			} else if !strings.Contains(err.Error(), tt.expectedErrorType) {
				t.Errorf("Error type mismatch. Expected to contain %q but got %q", tt.expectedErrorType, err.Error())
			}

			// Check for helpful tips in the error message
			if !strings.Contains(err.Error(), "Tip:") {
				t.Errorf("Error message doesn't contain user guidance (Tip)")
			}

			// Check for model name in the error message
			if !strings.Contains(err.Error(), "test-model") {
				t.Errorf("Error message doesn't contain the model name")
			}

			// Check for error details if provided
			if tt.errorDetails != "" && !strings.Contains(err.Error(), tt.errorDetails) {
				t.Errorf("Error message doesn't contain the detailed error information")
			}
		})
	}
}
