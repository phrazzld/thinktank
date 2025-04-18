package architect

import (
	"testing"
)

// TestProviderDetectionLogic tests the provider detection logic
func TestProviderDetectionLogic(t *testing.T) {
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
