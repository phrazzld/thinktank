package models

import (
	"strings"
	"testing"
)

func TestGetModelInfo(t *testing.T) {
	tests := []struct {
		name           string
		modelName      string
		wantProvider   string
		wantAPIModelID string
		wantError      bool
		errorContains  string
	}{
		// Valid models - OpenAI
		{
			name:           "gpt-4.1 valid model",
			modelName:      "gpt-4.1",
			wantProvider:   "openai",
			wantAPIModelID: "gpt-4.1",
			wantError:      false,
		},
		{
			name:           "o4-mini valid model",
			modelName:      "o4-mini",
			wantProvider:   "openai",
			wantAPIModelID: "o4-mini",
			wantError:      false,
		},
		// Valid models - Gemini
		{
			name:           "gemini-2.5-pro valid model",
			modelName:      "gemini-2.5-pro",
			wantProvider:   "gemini",
			wantAPIModelID: "gemini-2.5-pro",
			wantError:      false,
		},
		{
			name:           "gemini-2.5-flash valid model",
			modelName:      "gemini-2.5-flash",
			wantProvider:   "gemini",
			wantAPIModelID: "gemini-2.5-flash",
			wantError:      false,
		},
		// Valid models - OpenRouter
		{
			name:           "deepseek-chat-v3 valid model",
			modelName:      "openrouter/deepseek/deepseek-chat-v3-0324",
			wantProvider:   "openrouter",
			wantAPIModelID: "deepseek/deepseek-chat-v3-0324",
			wantError:      false,
		},
		{
			name:           "deepseek-r1-0528 valid model",
			modelName:      "openrouter/deepseek/deepseek-r1-0528",
			wantProvider:   "openrouter",
			wantAPIModelID: "deepseek/deepseek-r1-0528",
			wantError:      false,
		},
		{
			name:           "grok-3-beta valid model",
			modelName:      "openrouter/x-ai/grok-3-beta",
			wantProvider:   "openrouter",
			wantAPIModelID: "x-ai/grok-3-beta",
			wantError:      false,
		},
		// Invalid models
		{
			name:          "unknown model",
			modelName:     "unknown-model",
			wantError:     true,
			errorContains: "unknown model: unknown-model",
		},
		{
			name:          "empty model name",
			modelName:     "",
			wantError:     true,
			errorContains: "unknown model:",
		},
		{
			name:          "similar but incorrect model name",
			modelName:     "gpt-5",
			wantError:     true,
			errorContains: "unknown model: gpt-5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := GetModelInfo(tt.modelName)

			// Check error expectations
			if tt.wantError {
				if err == nil {
					t.Errorf("GetModelInfo(%q) expected error, got nil", tt.modelName)
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("GetModelInfo(%q) error = %v, want error containing %q", tt.modelName, err, tt.errorContains)
				}
				return
			}

			// Check for unexpected errors
			if err != nil {
				t.Errorf("GetModelInfo(%q) unexpected error = %v", tt.modelName, err)
				return
			}

			// Validate returned ModelInfo
			if info.Provider != tt.wantProvider {
				t.Errorf("GetModelInfo(%q).Provider = %v, want %v", tt.modelName, info.Provider, tt.wantProvider)
			}
			if info.APIModelID != tt.wantAPIModelID {
				t.Errorf("GetModelInfo(%q).APIModelID = %v, want %v", tt.modelName, info.APIModelID, tt.wantAPIModelID)
			}

			// Validate that required fields are not zero values
			if info.ContextWindow <= 0 {
				t.Errorf("GetModelInfo(%q).ContextWindow = %v, want > 0", tt.modelName, info.ContextWindow)
			}
			if info.MaxOutputTokens <= 0 {
				t.Errorf("GetModelInfo(%q).MaxOutputTokens = %v, want > 0", tt.modelName, info.MaxOutputTokens)
			}
			if info.DefaultParams == nil {
				t.Errorf("GetModelInfo(%q).DefaultParams = nil, want non-nil map", tt.modelName)
			}
		})
	}
}

func TestGetProviderForModel(t *testing.T) {
	tests := []struct {
		name          string
		modelName     string
		wantProvider  string
		wantError     bool
		errorContains string
	}{
		// Valid models - OpenAI
		{
			name:         "gpt-4.1 provider",
			modelName:    "gpt-4.1",
			wantProvider: "openai",
			wantError:    false,
		},
		{
			name:         "o4-mini provider",
			modelName:    "o4-mini",
			wantProvider: "openai",
			wantError:    false,
		},
		// Valid models - Gemini
		{
			name:         "gemini-2.5-pro provider",
			modelName:    "gemini-2.5-pro",
			wantProvider: "gemini",
			wantError:    false,
		},
		{
			name:         "gemini-2.5-flash provider",
			modelName:    "gemini-2.5-flash",
			wantProvider: "gemini",
			wantError:    false,
		},
		// Valid models - OpenRouter
		{
			name:         "deepseek-chat-v3 provider",
			modelName:    "openrouter/deepseek/deepseek-chat-v3-0324",
			wantProvider: "openrouter",
			wantError:    false,
		},
		{
			name:         "deepseek-r1-0528 provider",
			modelName:    "openrouter/deepseek/deepseek-r1-0528",
			wantProvider: "openrouter",
			wantError:    false,
		},
		{
			name:         "grok-3-beta provider",
			modelName:    "openrouter/x-ai/grok-3-beta",
			wantProvider: "openrouter",
			wantError:    false,
		},
		// Invalid models
		{
			name:          "unknown model",
			modelName:     "unknown-model",
			wantError:     true,
			errorContains: "unknown model: unknown-model",
		},
		{
			name:          "empty model name",
			modelName:     "",
			wantError:     true,
			errorContains: "unknown model:",
		},
		{
			name:          "similar but incorrect model name",
			modelName:     "gpt-5",
			wantError:     true,
			errorContains: "unknown model: gpt-5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := GetProviderForModel(tt.modelName)

			// Check error expectations
			if tt.wantError {
				if err == nil {
					t.Errorf("GetProviderForModel(%q) expected error, got nil", tt.modelName)
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("GetProviderForModel(%q) error = %v, want error containing %q", tt.modelName, err, tt.errorContains)
				}
				return
			}

			// Check for unexpected errors
			if err != nil {
				t.Errorf("GetProviderForModel(%q) unexpected error = %v", tt.modelName, err)
				return
			}

			// Validate returned provider
			if provider != tt.wantProvider {
				t.Errorf("GetProviderForModel(%q) = %v, want %v", tt.modelName, provider, tt.wantProvider)
			}
		})
	}
}

func TestListAllModels(t *testing.T) {
	models := ListAllModels()

	// Define expected models in alphabetical order
	expectedModels := []string{
		"gemini-2.5-flash",
		"gemini-2.5-pro",
		"gpt-4.1",
		"o3",
		"o4-mini",
		"openrouter/deepseek/deepseek-chat-v3-0324",
		"openrouter/deepseek/deepseek-chat-v3-0324:free",
		"openrouter/deepseek/deepseek-r1-0528",
		"openrouter/deepseek/deepseek-r1-0528:free",
		"openrouter/google/gemma-3-27b-it",
		"openrouter/meta-llama/llama-3.3-70b-instruct",
		"openrouter/meta-llama/llama-4-maverick",
		"openrouter/meta-llama/llama-4-scout",
		"openrouter/x-ai/grok-3-beta",
		"openrouter/x-ai/grok-3-mini-beta",
	}

	// Verify correct number of models
	if len(models) != len(expectedModels) {
		t.Errorf("ListAllModels() returned %d models, want %d", len(models), len(expectedModels))
	}

	// Verify all expected models are present and in correct order
	for i, expected := range expectedModels {
		if i >= len(models) {
			t.Errorf("ListAllModels() missing model at index %d: %s", i, expected)
			continue
		}
		if models[i] != expected {
			t.Errorf("ListAllModels()[%d] = %s, want %s", i, models[i], expected)
		}
	}

	// Verify no unexpected models are present
	if len(models) > len(expectedModels) {
		for i := len(expectedModels); i < len(models); i++ {
			t.Errorf("ListAllModels() contains unexpected model at index %d: %s", i, models[i])
		}
	}

	// Verify the slice is sorted
	for i := 1; i < len(models); i++ {
		if models[i-1] >= models[i] {
			t.Errorf("ListAllModels() not properly sorted: models[%d]=%s >= models[%d]=%s",
				i-1, models[i-1], i, models[i])
		}
	}
}

func TestListModelsForProvider(t *testing.T) {
	tests := []struct {
		name           string
		provider       string
		expectedModels []string
	}{
		{
			name:     "openai provider",
			provider: "openai",
			expectedModels: []string{
				"gpt-4.1",
				"o3",
				"o4-mini",
			},
		},
		{
			name:     "gemini provider",
			provider: "gemini",
			expectedModels: []string{
				"gemini-2.5-flash",
				"gemini-2.5-pro",
			},
		},
		{
			name:     "openrouter provider",
			provider: "openrouter",
			expectedModels: []string{
				"openrouter/deepseek/deepseek-chat-v3-0324",
				"openrouter/deepseek/deepseek-chat-v3-0324:free",
				"openrouter/deepseek/deepseek-r1-0528",
				"openrouter/deepseek/deepseek-r1-0528:free",
				"openrouter/google/gemma-3-27b-it",
				"openrouter/meta-llama/llama-3.3-70b-instruct",
				"openrouter/meta-llama/llama-4-maverick",
				"openrouter/meta-llama/llama-4-scout",
				"openrouter/x-ai/grok-3-beta",
				"openrouter/x-ai/grok-3-mini-beta",
			},
		},
		{
			name:           "unknown provider",
			provider:       "unknown-provider",
			expectedModels: []string{},
		},
		{
			name:           "empty provider",
			provider:       "",
			expectedModels: []string{},
		},
		{
			name:           "case sensitive provider",
			provider:       "OpenAI",
			expectedModels: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			models := ListModelsForProvider(tt.provider)

			// Verify correct number of models
			if len(models) != len(tt.expectedModels) {
				t.Errorf("ListModelsForProvider(%q) returned %d models, want %d",
					tt.provider, len(models), len(tt.expectedModels))
			}

			// Verify all expected models are present and in correct order
			for i, expected := range tt.expectedModels {
				if i >= len(models) {
					t.Errorf("ListModelsForProvider(%q) missing model at index %d: %s",
						tt.provider, i, expected)
					continue
				}
				if models[i] != expected {
					t.Errorf("ListModelsForProvider(%q)[%d] = %s, want %s",
						tt.provider, i, models[i], expected)
				}
			}

			// Verify no unexpected models are present
			if len(models) > len(tt.expectedModels) {
				for i := len(tt.expectedModels); i < len(models); i++ {
					t.Errorf("ListModelsForProvider(%q) contains unexpected model at index %d: %s",
						tt.provider, i, models[i])
				}
			}

			// Verify the slice is sorted (only if we have more than one model)
			if len(models) > 1 {
				for i := 1; i < len(models); i++ {
					if models[i-1] >= models[i] {
						t.Errorf("ListModelsForProvider(%q) not properly sorted: models[%d]=%s >= models[%d]=%s",
							tt.provider, i-1, models[i-1], i, models[i])
					}
				}
			}
		})
	}
}

func TestGetAPIKeyEnvVar(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		expectedVar string
	}{
		{
			name:        "openai provider",
			provider:    "openai",
			expectedVar: "OPENAI_API_KEY",
		},
		{
			name:        "gemini provider",
			provider:    "gemini",
			expectedVar: "GEMINI_API_KEY",
		},
		{
			name:        "openrouter provider",
			provider:    "openrouter",
			expectedVar: "OPENROUTER_API_KEY",
		},
		{
			name:        "unknown provider",
			provider:    "unknown-provider",
			expectedVar: "",
		},
		{
			name:        "empty provider",
			provider:    "",
			expectedVar: "",
		},
		{
			name:        "case sensitive provider",
			provider:    "OpenAI",
			expectedVar: "",
		},
		{
			name:        "different case provider",
			provider:    "OPENAI",
			expectedVar: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVar := GetAPIKeyEnvVar(tt.provider)
			if envVar != tt.expectedVar {
				t.Errorf("GetAPIKeyEnvVar(%q) = %q, want %q", tt.provider, envVar, tt.expectedVar)
			}
		})
	}
}

func TestIsModelSupported(t *testing.T) {
	tests := []struct {
		name        string
		modelName   string
		expectedRes bool
	}{
		// Valid models - should return true
		{
			name:        "gpt-4.1 supported",
			modelName:   "gpt-4.1",
			expectedRes: true,
		},
		{
			name:        "o4-mini supported",
			modelName:   "o4-mini",
			expectedRes: true,
		},
		{
			name:        "gemini-2.5-pro supported",
			modelName:   "gemini-2.5-pro",
			expectedRes: true,
		},
		{
			name:        "gemini-2.5-flash supported",
			modelName:   "gemini-2.5-flash",
			expectedRes: true,
		},
		{
			name:        "deepseek-chat-v3 supported",
			modelName:   "openrouter/deepseek/deepseek-chat-v3-0324",
			expectedRes: true,
		},
		{
			name:        "deepseek-r1-0528 supported",
			modelName:   "openrouter/deepseek/deepseek-r1-0528",
			expectedRes: true,
		},
		{
			name:        "grok-3-beta supported",
			modelName:   "openrouter/x-ai/grok-3-beta",
			expectedRes: true,
		},
		// Invalid models - should return false
		{
			name:        "unknown model not supported",
			modelName:   "unknown-model",
			expectedRes: false,
		},
		{
			name:        "empty model name not supported",
			modelName:   "",
			expectedRes: false,
		},
		{
			name:        "similar but incorrect model not supported",
			modelName:   "gpt-5",
			expectedRes: false,
		},
		{
			name:        "case sensitive model not supported",
			modelName:   "GPT-4.1",
			expectedRes: false,
		},
		{
			name:        "partial model name not supported",
			modelName:   "gpt",
			expectedRes: false,
		},
		{
			name:        "model with extra characters not supported",
			modelName:   "gpt-4.1-extra",
			expectedRes: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsModelSupported(tt.modelName)
			if result != tt.expectedRes {
				t.Errorf("IsModelSupported(%q) = %v, want %v", tt.modelName, result, tt.expectedRes)
			}
		})
	}
}

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
