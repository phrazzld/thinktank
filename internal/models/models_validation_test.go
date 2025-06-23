package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateParameter(t *testing.T) {
	tests := []struct {
		name          string
		modelName     string
		paramName     string
		value         interface{}
		wantError     bool
		errorContains string
	}{
		// Valid temperature values across all models
		{
			name:      "temperature valid lower bound",
			modelName: "gpt-4.1",
			paramName: "temperature",
			value:     0.0,
			wantError: false,
		},
		{
			name:      "temperature valid middle value",
			modelName: "gemini-2.5-pro",
			paramName: "temperature",
			value:     1.0,
			wantError: false,
		},
		{
			name:      "temperature valid upper bound",
			modelName: "openrouter/deepseek/deepseek-r1-0528",
			paramName: "temperature",
			value:     2.0,
			wantError: false,
		},
		{
			name:      "temperature as int (converted to float)",
			modelName: "gpt-4.1",
			paramName: "temperature",
			value:     1,
			wantError: false,
		},
		{
			name:      "temperature as float32",
			modelName: "gemini-2.5-flash",
			paramName: "temperature",
			value:     float32(0.7),
			wantError: false,
		},

		// Invalid temperature values
		{
			name:          "temperature below minimum",
			modelName:     "gpt-4.1",
			paramName:     "temperature",
			value:         -0.1,
			wantError:     true,
			errorContains: "temperature",
		},
		{
			name:          "temperature above maximum",
			modelName:     "gemini-2.5-pro",
			paramName:     "temperature",
			value:         2.1,
			wantError:     true,
			errorContains: "temperature",
		},
		{
			name:          "temperature wrong type",
			modelName:     "gpt-4.1",
			paramName:     "temperature",
			value:         "0.7",
			wantError:     true,
			errorContains: "must be a numeric value",
		},

		// Valid top_p values
		{
			name:      "top_p valid lower bound",
			modelName: "gpt-4.1",
			paramName: "top_p",
			value:     0.0,
			wantError: false,
		},
		{
			name:      "top_p valid middle value",
			modelName: "gemini-2.5-pro",
			paramName: "top_p",
			value:     0.95,
			wantError: false,
		},
		{
			name:      "top_p valid upper bound",
			modelName: "openrouter/deepseek/deepseek-r1-0528",
			paramName: "top_p",
			value:     1.0,
			wantError: false,
		},

		// Invalid top_p values
		{
			name:          "top_p below minimum",
			modelName:     "gpt-4.1",
			paramName:     "top_p",
			value:         -0.1,
			wantError:     true,
			errorContains: "top_p",
		},
		{
			name:          "top_p above maximum",
			modelName:     "gemini-2.5-pro",
			paramName:     "top_p",
			value:         1.1,
			wantError:     true,
			errorContains: "top_p",
		},

		// Valid max_tokens/max_output_tokens
		{
			name:      "max_tokens valid",
			modelName: "gpt-4.1",
			paramName: "max_tokens",
			value:     1000,
			wantError: false,
		},
		{
			name:      "max_output_tokens valid",
			modelName: "gemini-2.5-pro",
			paramName: "max_output_tokens",
			value:     5000,
			wantError: false,
		},
		{
			name:      "max_tokens as float64 whole number",
			modelName: "gpt-4.1",
			paramName: "max_tokens",
			value:     float64(2048),
			wantError: false,
		},

		// Invalid max_tokens/max_output_tokens
		{
			name:          "max_tokens zero",
			modelName:     "gpt-4.1",
			paramName:     "max_tokens",
			value:         0,
			wantError:     true,
			errorContains: "max_tokens",
		},
		{
			name:          "max_tokens negative",
			modelName:     "gpt-4.1",
			paramName:     "max_tokens",
			value:         -100,
			wantError:     true,
			errorContains: "max_tokens",
		},
		{
			name:          "max_tokens as non-integer float",
			modelName:     "gpt-4.1",
			paramName:     "max_tokens",
			value:         1000.5,
			wantError:     true,
			errorContains: "must be an integer",
		},

		// Valid top_k (Gemini models)
		{
			name:      "top_k valid",
			modelName: "gemini-2.5-pro",
			paramName: "top_k",
			value:     40,
			wantError: false,
		},
		{
			name:      "top_k minimum value",
			modelName: "gemini-2.5-flash",
			paramName: "top_k",
			value:     1,
			wantError: false,
		},

		// Invalid top_k
		{
			name:          "top_k zero",
			modelName:     "gemini-2.5-pro",
			paramName:     "top_k",
			value:         0,
			wantError:     true,
			errorContains: "top_k",
		},
		{
			name:          "top_k above maximum",
			modelName:     "gemini-2.5-pro",
			paramName:     "top_k",
			value:         101,
			wantError:     true,
			errorContains: "top_k",
		},

		// Valid frequency_penalty and presence_penalty (OpenAI models)
		{
			name:      "frequency_penalty valid",
			modelName: "gpt-4.1",
			paramName: "frequency_penalty",
			value:     0.5,
			wantError: false,
		},
		{
			name:      "presence_penalty valid",
			modelName: "o4-mini",
			paramName: "presence_penalty",
			value:     -1.0,
			wantError: false,
		},
		{
			name:      "frequency_penalty upper bound",
			modelName: "gpt-4.1",
			paramName: "frequency_penalty",
			value:     2.0,
			wantError: false,
		},
		{
			name:      "presence_penalty lower bound",
			modelName: "o4-mini",
			paramName: "presence_penalty",
			value:     -2.0,
			wantError: false,
		},

		// Invalid frequency_penalty and presence_penalty
		{
			name:          "frequency_penalty below minimum",
			modelName:     "gpt-4.1",
			paramName:     "frequency_penalty",
			value:         -2.1,
			wantError:     true,
			errorContains: "frequency_penalty",
		},
		{
			name:          "presence_penalty above maximum",
			modelName:     "o4-mini",
			paramName:     "presence_penalty",
			value:         2.1,
			wantError:     true,
			errorContains: "presence_penalty",
		},

		// Parameters not defined in constraints (should be accepted)
		{
			name:      "undefined parameter accepted",
			modelName: "gpt-4.1",
			paramName: "custom_param",
			value:     "any_value",
			wantError: false,
		},
		{
			name:      "reasoning parameter accepted (nested object)",
			modelName: "o4-mini",
			paramName: "reasoning",
			value:     map[string]interface{}{"effort": "high"},
			wantError: false,
		},

		// Model validation errors
		{
			name:          "unknown model",
			modelName:     "unknown-model",
			paramName:     "temperature",
			value:         0.7,
			wantError:     true,
			errorContains: "model 'unknown-model' not supported",
		},

		// Edge cases for different numeric types
		{
			name:      "int32 parameter",
			modelName: "gpt-4.1",
			paramName: "max_tokens",
			value:     int32(1024),
			wantError: false,
		},
		{
			name:      "int64 parameter",
			modelName: "gpt-4.1",
			paramName: "max_tokens",
			value:     int64(2048),
			wantError: false,
		},
		{
			name:      "float32 temperature",
			modelName: "gpt-4.1",
			paramName: "temperature",
			value:     float32(0.8),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateParameter(tt.modelName, tt.paramName, tt.value)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateParameter(%q, %q, %v) expected error, got nil", tt.modelName, tt.paramName, tt.value)
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ValidateParameter(%q, %q, %v) error = %v, want error containing %q",
						tt.modelName, tt.paramName, tt.value, err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateParameter(%q, %q, %v) unexpected error = %v", tt.modelName, tt.paramName, tt.value, err)
				}
			}
		})
	}
}

func TestParameterConstraints(t *testing.T) {
	// Test that all models have parameter constraints defined
	allModels := ListAllModels()

	for _, modelName := range allModels {
		t.Run(modelName, func(t *testing.T) {
			info, err := GetModelInfo(modelName)
			if err != nil {
				t.Fatalf("GetModelInfo(%s) failed: %v", modelName, err)
			}

			// All models should have parameter constraints
			if info.ParameterConstraints == nil {
				t.Errorf("Model %s has nil ParameterConstraints", modelName)
				return
			}

			// All models should have temperature constraint
			tempConstraint, exists := info.ParameterConstraints["temperature"]
			if !exists {
				t.Errorf("Model %s missing temperature constraint", modelName)
			} else {
				if tempConstraint.Type != "float" {
					t.Errorf("Model %s temperature constraint type = %s, want float", modelName, tempConstraint.Type)
				}
				if tempConstraint.MinValue == nil || *tempConstraint.MinValue != 0.0 {
					t.Errorf("Model %s temperature min value incorrect", modelName)
				}
				if tempConstraint.MaxValue == nil || *tempConstraint.MaxValue != 2.0 {
					t.Errorf("Model %s temperature max value incorrect", modelName)
				}
			}

			// All models should have top_p constraint
			topPConstraint, exists := info.ParameterConstraints["top_p"]
			if !exists {
				t.Errorf("Model %s missing top_p constraint", modelName)
			} else {
				if topPConstraint.Type != "float" {
					t.Errorf("Model %s top_p constraint type = %s, want float", modelName, topPConstraint.Type)
				}
				if topPConstraint.MinValue == nil || *topPConstraint.MinValue != 0.0 {
					t.Errorf("Model %s top_p min value incorrect", modelName)
				}
				if topPConstraint.MaxValue == nil || *topPConstraint.MaxValue != 1.0 {
					t.Errorf("Model %s top_p max value incorrect", modelName)
				}
			}

			// Provider-specific constraints
			switch info.Provider {
			case "openai":
				// OpenAI models should have frequency_penalty and presence_penalty
				if _, exists := info.ParameterConstraints["frequency_penalty"]; !exists {
					t.Errorf("OpenAI model %s missing frequency_penalty constraint", modelName)
				}
				if _, exists := info.ParameterConstraints["presence_penalty"]; !exists {
					t.Errorf("OpenAI model %s missing presence_penalty constraint", modelName)
				}
				if _, exists := info.ParameterConstraints["max_tokens"]; !exists {
					t.Errorf("OpenAI model %s missing max_tokens constraint", modelName)
				}
			case "gemini":
				// Gemini models should have top_k and max_output_tokens
				if _, exists := info.ParameterConstraints["top_k"]; !exists {
					t.Errorf("Gemini model %s missing top_k constraint", modelName)
				}
				if _, exists := info.ParameterConstraints["max_output_tokens"]; !exists {
					t.Errorf("Gemini model %s missing max_output_tokens constraint", modelName)
				}
			case "openrouter":
				// OpenRouter models should have max_tokens
				if _, exists := info.ParameterConstraints["max_tokens"]; !exists {
					t.Errorf("OpenRouter model %s missing max_tokens constraint", modelName)
				}
			}
		})
	}
}

func TestGetProviderDefaultRateLimit(t *testing.T) {
	tests := []struct {
		name         string
		provider     string
		expectedRate int
	}{
		{
			name:         "OpenAI provider",
			provider:     "openai",
			expectedRate: 3000,
		},
		{
			name:         "Gemini provider",
			provider:     "gemini",
			expectedRate: 60,
		},
		{
			name:         "OpenRouter provider",
			provider:     "openrouter",
			expectedRate: 20,
		},
		{
			name:         "Unknown provider defaults to conservative",
			provider:     "unknown-provider",
			expectedRate: 60,
		},
		{
			name:         "Empty provider defaults to conservative",
			provider:     "",
			expectedRate: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := GetProviderDefaultRateLimit(tt.provider)
			assert.Equal(t, tt.expectedRate, rate)
		})
	}
}

func TestGetModelRateLimit(t *testing.T) {
	tests := []struct {
		name           string
		modelName      string
		expectedRate   int
		expectError    bool
		errorSubstring string
	}{
		{
			name:         "Known model without specific rate limit uses provider default",
			modelName:    "gemini-2.5-pro",
			expectedRate: 60, // Should use gemini provider default
			expectError:  false,
		},
		{
			name:         "Known OpenAI model uses provider default",
			modelName:    "gpt-4.1",
			expectedRate: 3000, // Should use openai provider default
			expectError:  false,
		},
		{
			name:           "Unknown model returns error",
			modelName:      "unknown-invalid-model",
			expectedRate:   0,
			expectError:    true,
			errorSubstring: "unknown model",
		},
		{
			name:           "Empty model name returns error",
			modelName:      "",
			expectedRate:   0,
			expectError:    true,
			errorSubstring: "unknown model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate, err := GetModelRateLimit(tt.modelName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, 0, rate)
				if tt.errorSubstring != "" {
					assert.Contains(t, err.Error(), tt.errorSubstring)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRate, rate)
			}
		})
	}
}
