package architect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetModelTokenLimits(t *testing.T) {
	tests := []struct {
		name            string
		modelName       string
		expectError     bool
		expectedContext int32
		expectedOutput  int32
		expectedErrMsg  string
	}{
		// Gemini models
		{
			name:            "gemini-1.5-pro model",
			modelName:       "gemini-1.5-pro",
			expectError:     false,
			expectedContext: 1000000,
			expectedOutput:  8192,
		},
		{
			name:            "gemini-1.5-flash model",
			modelName:       "gemini-1.5-flash",
			expectError:     false,
			expectedContext: 1000000,
			expectedOutput:  8192,
		},
		{
			name:            "gemini-1.0-pro model",
			modelName:       "gemini-1.0-pro",
			expectError:     false,
			expectedContext: 32768,
			expectedOutput:  8192,
		},
		{
			name:            "gemini-1.0-ultra model",
			modelName:       "gemini-1.0-ultra",
			expectError:     false,
			expectedContext: 32768,
			expectedOutput:  8192,
		},

		// OpenAI models - existing
		{
			name:            "gpt-4-turbo model",
			modelName:       "gpt-4-turbo",
			expectError:     false,
			expectedContext: 128000,
			expectedOutput:  4096,
		},
		{
			name:            "gpt-4 model",
			modelName:       "gpt-4",
			expectError:     false,
			expectedContext: 8192,
			expectedOutput:  4096,
		},
		{
			name:            "gpt-3.5-turbo model",
			modelName:       "gpt-3.5-turbo",
			expectError:     false,
			expectedContext: 16385,
			expectedOutput:  4096,
		},

		// OpenAI models - 1M token models
		{
			name:            "gpt-4.1 model",
			modelName:       "gpt-4.1",
			expectError:     false,
			expectedContext: 1000000,
			expectedOutput:  32768,
		},
		{
			name:            "o4-mini model",
			modelName:       "o4-mini",
			expectError:     false,
			expectedContext: 1000000,
			expectedOutput:  32768,
		},
		{
			name:            "custom model name with o4 prefix",
			modelName:       "o4-preview-experimental",
			expectError:     false,
			expectedContext: 1000000,
			expectedOutput:  32768,
		},

		// OpenAI models - 128k token models
		{
			name:            "gpt-4o model",
			modelName:       "gpt-4o",
			expectError:     false,
			expectedContext: 128000,
			expectedOutput:  4096,
		},

		// Unknown model
		{
			name:           "unknown model",
			modelName:      "unknown-model",
			expectError:    true,
			expectedErrMsg: "token limits not available for model: unknown-model",
		},
	}

	// Create a new adapter with a mock service that delegates to the adapter's implementation
	adapter := &APIServiceAdapter{
		APIService: &MockAPIServiceForAdapter{
			GetModelTokenLimitsFunc: func(modelName string) (int32, int32, error) {
				// Call the adapter's implementation directly to test it
				return (&APIServiceAdapter{}).GetModelTokenLimits(modelName)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			contextWindow, outputTokens, err := adapter.GetModelTokenLimits(tc.modelName)

			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedContext, contextWindow, "Context window should match expected value")
				assert.Equal(t, tc.expectedOutput, outputTokens, "Output tokens should match expected value")
			}
		})
	}
}
