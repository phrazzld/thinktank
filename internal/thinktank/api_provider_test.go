package thinktank

import (
	"context"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestProviderTypeDetection tests the DetectProviderFromModel function
func TestProviderTypeDetection(t *testing.T) {
	tests := []struct {
		name         string
		modelName    string
		expectedType ProviderType
	}{
		{
			name:         "Gemini Model",
			modelName:    "gemini-pro",
			expectedType: ProviderGemini,
		},
		{
			name:         "Gemini 1.5 Model",
			modelName:    "gemini-1.5-pro",
			expectedType: ProviderGemini,
		},
		{
			name:         "OpenAI GPT-4 Model",
			modelName:    "gpt-4",
			expectedType: ProviderOpenAI,
		},
		{
			name:         "OpenAI GPT-3.5 Model",
			modelName:    "gpt-3.5-turbo",
			expectedType: ProviderOpenAI,
		},
		{
			name:         "Empty Model Name",
			modelName:    "",
			expectedType: ProviderUnknown,
		},
		{
			name:         "Unknown Provider Model",
			modelName:    "unknown-model",
			expectedType: ProviderUnknown,
		},
		{
			name:         "OpenAI Text Davinci Model",
			modelName:    "text-davinci-003",
			expectedType: ProviderOpenAI,
		},
		{
			name:         "OpenAI Ada Model",
			modelName:    "ada",
			expectedType: ProviderOpenAI,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			providerType := DetectProviderFromModel(tc.modelName)
			if providerType != tc.expectedType {
				t.Errorf("Expected provider type %v for model %s, got %v",
					tc.expectedType, tc.modelName, providerType)
			}
		})
	}
}

// TestModelSupport tests IsModelSupported helper function
func TestModelSupport(t *testing.T) {
	tests := []struct {
		name          string
		modelName     string
		expectSupport bool
	}{
		{
			name:          "Supported Gemini Model",
			modelName:     "gemini-pro",
			expectSupport: true,
		},
		{
			name:          "Supported OpenAI Model",
			modelName:     "gpt-4",
			expectSupport: true,
		},
		{
			name:          "Empty Model Name",
			modelName:     "",
			expectSupport: false,
		},
		{
			name:          "Unsupported Model",
			modelName:     "unknown-model",
			expectSupport: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isSupported := DetectProviderFromModel(tc.modelName) != ProviderUnknown
			if isSupported != tc.expectSupport {
				t.Errorf("Expected model %s to be supported: %v, got: %v",
					tc.modelName, tc.expectSupport, isSupported)
			}
		})
	}
}

// TestProcessLLMResponseMultiProvider tests processing of different provider results
func TestProcessLLMResponseMultiProvider(t *testing.T) {
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "test")
	apiService := NewAPIService(logger)

	tests := []struct {
		name          string
		result        *llm.ProviderResult
		expectError   bool
		errorContains string
		expectedText  string
	}{
		{
			name:          "Nil Result",
			result:        nil,
			expectError:   true,
			errorContains: "empty response",
			expectedText:  "",
		},
		{
			name:          "Valid Content",
			result:        &llm.ProviderResult{Content: "This is valid content"},
			expectError:   false,
			errorContains: "",
			expectedText:  "This is valid content",
		},
		{
			name: "Safety Blocked",
			result: &llm.ProviderResult{
				Content: "",
				SafetyInfo: []llm.Safety{
					{Category: "HARM_CATEGORY_DANGEROUS", Blocked: true},
				},
				FinishReason: "safety",
			},
			expectError:   true,
			errorContains: "safety filters",
			expectedText:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			content, err := apiService.ProcessLLMResponse(tc.result)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got: '%s'", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
					return
				}
			}

			if content != tc.expectedText {
				t.Errorf("Expected content '%s', got '%s'", tc.expectedText, content)
			}
		})
	}
}

// TestInitLLMClientProviderDetection tests that the InitLLMClient method correctly detects providers from model names
func TestInitLLMClientProviderDetection(t *testing.T) {
	// Create API service with custom client wrappers
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "test")
	apiService := &apiService{
		logger:              logger,
		newGeminiClientFunc: newGeminiClientWrapperForTest,
		newOpenAIClientFunc: newOpenAIClientWrapperForTest,
	}

	tests := []struct {
		name        string
		modelName   string
		expectError bool
		provider    string
	}{
		{
			name:        "Gemini Model",
			modelName:   "gemini-pro",
			expectError: false,
			provider:    "Gemini",
		},
		{
			name:        "Gemini 1.5 Model",
			modelName:   "gemini-1.5-pro",
			expectError: false,
			provider:    "Gemini",
		},
		{
			name:        "OpenAI Model",
			modelName:   "gpt-4",
			expectError: false,
			provider:    "OpenAI",
		},
		{
			name:        "OpenAI GPT-3.5 Model",
			modelName:   "gpt-3.5-turbo",
			expectError: false,
			provider:    "OpenAI",
		},
		{
			name:        "Unknown Provider",
			modelName:   "unknown-model",
			expectError: true,
			provider:    "",
		},
		{
			name:        "OpenAI Text Davinci Model",
			modelName:   "text-davinci-003",
			expectError: false,
			provider:    "OpenAI",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := apiService.InitLLMClient(context.Background(), "test-key", tc.modelName, "")

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
					return
				}

				// Verify model name
				if client.GetModelName() != tc.modelName {
					t.Errorf("Expected model name %s, got %s", tc.modelName, client.GetModelName())
				}

				// Verify provider-specific debug logs
				// This isn't a direct test of the functionality but rather a check that
				// the provider detection is logging the right information
				detectResult := DetectProviderFromModel(tc.modelName)
				if tc.provider == "Gemini" && detectResult != ProviderGemini {
					t.Errorf("Expected Gemini provider type for %s, got %v", tc.modelName, detectResult)
				} else if tc.provider == "OpenAI" && detectResult != ProviderOpenAI {
					t.Errorf("Expected OpenAI provider type for %s, got %v", tc.modelName, detectResult)
				}
			}
		})
	}
}
