package models

import (
	"strings"
	"testing"
)

func TestGetModelInfo(t *testing.T) {
	t.Parallel()
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
			name:           "openrouter model",
			modelName:      "openrouter/meta-llama/llama-4-maverick",
			wantProvider:   "openrouter",
			wantAPIModelID: "meta-llama/llama-4-maverick",
			wantError:      false,
		},
		// Error cases
		{
			name:          "empty model name",
			modelName:     "",
			wantError:     true,
			errorContains: "unknown model",
		},
		{
			name:          "unknown model",
			modelName:     "unknown-model",
			wantError:     true,
			errorContains: "unknown model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := GetModelInfo(tt.modelName)

			if tt.wantError {
				if err == nil {
					t.Errorf("GetModelInfo(%q) expected error, got nil", tt.modelName)
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("GetModelInfo(%q) error = %q, want error containing %q", tt.modelName, err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("GetModelInfo(%q) unexpected error: %v", tt.modelName, err)
				return
			}

			if info.Provider != tt.wantProvider {
				t.Errorf("GetModelInfo(%q).Provider = %q, want %q", tt.modelName, info.Provider, tt.wantProvider)
			}

			if info.APIModelID != tt.wantAPIModelID {
				t.Errorf("GetModelInfo(%q).APIModelID = %q, want %q", tt.modelName, info.APIModelID, tt.wantAPIModelID)
			}

			// Verify context window is positive
			if info.ContextWindow <= 0 {
				t.Errorf("GetModelInfo(%q).ContextWindow = %d, want > 0", tt.modelName, info.ContextWindow)
			}

			// Verify max output tokens is positive
			if info.MaxOutputTokens <= 0 {
				t.Errorf("GetModelInfo(%q).MaxOutputTokens = %d, want > 0", tt.modelName, info.MaxOutputTokens)
			}
		})
	}
}

func TestGetProviderDefaultRateLimitCore(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		provider string
		expected int
	}{
		{
			name:     "openai provider",
			provider: "openai",
			expected: 3000,
		},
		{
			name:     "gemini provider",
			provider: "gemini",
			expected: 60,
		},
		{
			name:     "openrouter provider",
			provider: "openrouter",
			expected: 20,
		},
		{
			name:     "unknown provider",
			provider: "unknown",
			expected: 0,
		},
		{
			name:     "empty provider",
			provider: "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetProviderDefaultRateLimit(tt.provider)
			if result != tt.expected {
				t.Errorf("GetProviderDefaultRateLimit(%q) = %d, want %d", tt.provider, result, tt.expected)
			}
		})
	}
}

func TestListModelsForProvider(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		provider string
		expected []string
	}{
		{
			name:     "openai provider",
			provider: "openai",
			expected: []string{"gpt-4.1", "o3", "o4-mini"},
		},
		{
			name:     "gemini provider",
			provider: "gemini",
			expected: []string{"gemini-2.5-pro", "gemini-2.5-flash"},
		},
		{
			name:     "openrouter provider",
			provider: "openrouter",
			expected: []string{
				"openrouter/meta-llama/llama-4-maverick",
				"openrouter/meta-llama/llama-4-scout",
				"openrouter/x-ai/grok-3-mini-beta",
				"openrouter/meta-llama/llama-3.3-70b-instruct",
				"openrouter/x-ai/grok-3-beta",
				"openrouter/deepseek/deepseek-r1-0528",
				"openrouter/deepseek/deepseek-chat-v3-0324:free",
				"openrouter/deepseek/deepseek-chat-v3-0324",
				"openrouter/google/gemma-3-27b-it",
			},
		},
		{
			name:     "unknown provider",
			provider: "unknown",
			expected: []string{},
		},
		{
			name:     "empty provider",
			provider: "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ListModelsForProvider(tt.provider)

			if len(result) != len(tt.expected) {
				t.Errorf("ListModelsForProvider(%q) returned %d models, want %d\nGot: %v\nWant: %v",
					tt.provider, len(result), len(tt.expected), result, tt.expected)
				return
			}

			// Create sets for comparison (order might vary)
			resultSet := make(map[string]bool)
			for _, model := range result {
				resultSet[model] = true
			}

			expectedSet := make(map[string]bool)
			for _, model := range tt.expected {
				expectedSet[model] = true
			}

			// Verify all expected models are present
			for _, expected := range tt.expected {
				if !resultSet[expected] {
					t.Errorf("ListModelsForProvider(%q) missing expected model: %s", tt.provider, expected)
				}
			}

			// Verify no unexpected models are present
			for _, actual := range result {
				if !expectedSet[actual] {
					t.Errorf("ListModelsForProvider(%q) returned unexpected model: %s", tt.provider, actual)
				}
			}
		})
	}
}

// Test benchmarks to verify performance
func BenchmarkGetModelInfo(b *testing.B) {
	modelName := "gpt-4.1"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetModelInfo(modelName)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetProviderDefaultRateLimit(b *testing.B) {
	provider := "openai"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetProviderDefaultRateLimit(provider)
	}
}

func BenchmarkListModelsForProvider(b *testing.B) {
	provider := "openai"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ListModelsForProvider(provider)
	}
}
