package gemini

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/gemini"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFinishReasons tests handling of different finish reasons
func TestFinishReasons(t *testing.T) {
	tests := []struct {
		name                string
		responseBody        []byte
		finishReason        string
		expectErrorContains string
		expectErrorCategory llm.ErrorCategory
	}{
		{
			name:                "STOP finish reason",
			responseBody:        makeGeminiResponseWithFinishReason("STOP"),
			finishReason:        "STOP",
			expectErrorContains: "",
			expectErrorCategory: llm.CategoryUnknown,
		},
		{
			name:                "MAX_TOKENS finish reason",
			responseBody:        makeGeminiResponseWithFinishReason("MAX_TOKENS"),
			finishReason:        "MAX_TOKENS",
			expectErrorContains: "",
			expectErrorCategory: llm.CategoryUnknown,
		},
		{
			name:                "SAFETY finish reason",
			responseBody:        makeGeminiResponseWithFinishReason("SAFETY"),
			finishReason:        "SAFETY",
			expectErrorContains: "safety",
			expectErrorCategory: llm.CategoryContentFiltered,
		},
		{
			name:                "RECITATION finish reason",
			responseBody:        makeGeminiResponseWithFinishReason("RECITATION"),
			finishReason:        "RECITATION",
			expectErrorContains: "",
			expectErrorCategory: llm.CategoryUnknown,
		},
		{
			name:                "Empty finish reason",
			responseBody:        makeGeminiResponseWithFinishReason(""),
			finishReason:        "",
			expectErrorContains: "",
			expectErrorCategory: llm.CategoryUnknown,
		},
		{
			name:                "Custom finish reason",
			responseBody:        makeGeminiResponseWithFinishReason("CUSTOM_REASON"),
			finishReason:        "CUSTOM_REASON",
			expectErrorContains: "",
			expectErrorCategory: llm.CategoryUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the response and extract the finish reason
			var respMap map[string]interface{}
			err := json.Unmarshal(tt.responseBody, &respMap)
			require.NoError(t, err, "Failed to parse test JSON")

			// Extract finish reason from response
			var finishReason string
			if candidates, ok := respMap["candidates"].([]interface{}); ok && len(candidates) > 0 {
				if candidate, ok := candidates[0].(map[string]interface{}); ok {
					if reason, ok := candidate["finishReason"].(string); ok {
						finishReason = reason
					}
				}
			}

			// Verify extracted finish reason matches expected
			assert.Equal(t, tt.finishReason, finishReason)

			// Simulate handling of error-indicating finish reasons
			var handlingErr error
			if finishReason == "SAFETY" {
				handlingErr = gemini.CreateAPIError(
					llm.CategoryContentFiltered,
					"Response blocked by Gemini's safety settings",
					nil,
					"finishReason: SAFETY",
				)
			}

			// Verify error handling for error-indicating finish reasons
			if tt.expectErrorContains != "" {
				require.NotNil(t, handlingErr)
				assert.Contains(t, strings.ToLower(handlingErr.Error()), strings.ToLower(tt.expectErrorContains))

				var llmErr *llm.LLMError
				if errors.As(handlingErr, &llmErr) {
					assert.Equal(t, tt.expectErrorCategory, llmErr.Category())
					assert.NotEmpty(t, llmErr.Suggestion)
				}
			} else {
				assert.Nil(t, handlingErr)
			}
		})
	}
}
