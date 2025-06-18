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
			name:           "deepseek-r1 valid model",
			modelName:      "openrouter/deepseek/deepseek-r1",
			wantProvider:   "openrouter",
			wantAPIModelID: "deepseek/deepseek-r1",
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
			name:         "deepseek-r1 provider",
			modelName:    "openrouter/deepseek/deepseek-r1",
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
		"o4-mini",
		"openrouter/deepseek/deepseek-chat-v3-0324",
		"openrouter/deepseek/deepseek-r1",
		"openrouter/x-ai/grok-3-beta",
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
				"openrouter/deepseek/deepseek-r1",
				"openrouter/x-ai/grok-3-beta",
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
			name:        "deepseek-r1 supported",
			modelName:   "openrouter/deepseek/deepseek-r1",
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

// TODO: Additional tests for other functions will be implemented in subsequent tasks
