// Package models provides model configuration and selection functionality
package models

import (
	"testing"
)

func TestGetModelsWithMinContextWindow(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		minTokens      int
		expectedModels []string
		verifyOrder    bool
	}{
		{
			name:      "very small threshold includes all models",
			minTokens: 1000,
			expectedModels: []string{
				"openrouter/meta-llama/llama-4-maverick",
				"openrouter/meta-llama/llama-4-scout",
				"gemini-2.5-pro",
				"gpt-4.1",
				"gemini-2.5-flash",
				"o3",
				"o4-mini",
				"openrouter/x-ai/grok-3-mini-beta",
				"openrouter/meta-llama/llama-3.3-70b-instruct",
				"openrouter/x-ai/grok-3-beta",
				"openrouter/deepseek/deepseek-r1-0528",
				"openrouter/deepseek/deepseek-chat-v3-0324:free",
				"openrouter/deepseek/deepseek-chat-v3-0324",
				"openrouter/google/gemma-3-27b-it",
			},
			verifyOrder: true,
		},
		{
			name:      "medium threshold filters some models",
			minTokens: 100000,
			expectedModels: []string{
				"openrouter/meta-llama/llama-4-maverick",
				"openrouter/meta-llama/llama-4-scout",
				"gemini-2.5-pro",
				"gpt-4.1",
				"gemini-2.5-flash",
				"o3",
				"o4-mini",
				"openrouter/x-ai/grok-3-mini-beta",
				"openrouter/meta-llama/llama-3.3-70b-instruct",
				"openrouter/x-ai/grok-3-beta",
				"openrouter/deepseek/deepseek-r1-0528",
				"openrouter/deepseek/deepseek-chat-v3-0324:free",
				"openrouter/deepseek/deepseek-chat-v3-0324",
			},
			verifyOrder: true,
		},
		{
			name:      "high threshold only largest models",
			minTokens: 500000,
			expectedModels: []string{
				"openrouter/meta-llama/llama-4-maverick",
				"openrouter/meta-llama/llama-4-scout",
				"gemini-2.5-pro",
				"gpt-4.1",
				"gemini-2.5-flash",
			},
			verifyOrder: true,
		},
		{
			name:      "very high threshold",
			minTokens: 900000,
			expectedModels: []string{
				"openrouter/meta-llama/llama-4-maverick",
				"openrouter/meta-llama/llama-4-scout",
				"gemini-2.5-pro",
				"gpt-4.1",
				"gemini-2.5-flash",
			},
			verifyOrder: true,
		},
		{
			name:      "threshold above most models",
			minTokens: 300000,
			expectedModels: []string{
				"openrouter/meta-llama/llama-4-maverick",
				"openrouter/meta-llama/llama-4-scout",
				"gemini-2.5-pro",
				"gpt-4.1",
				"gemini-2.5-flash",
				"o3",
				"o4-mini",
				"openrouter/x-ai/grok-3-mini-beta",
				"openrouter/meta-llama/llama-3.3-70b-instruct",
				"openrouter/x-ai/grok-3-beta",
				"openrouter/deepseek/deepseek-r1-0528",
			},
			verifyOrder: true,
		},
		{
			name:      "exact threshold boundary",
			minTokens: 1000000,
			expectedModels: []string{
				"openrouter/meta-llama/llama-4-maverick",
				"openrouter/meta-llama/llama-4-scout",
				"gemini-2.5-pro",
				"gpt-4.1",
				"gemini-2.5-flash",
			},
			verifyOrder: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetModelsWithMinContextWindow(tt.minTokens)

			// Verify correct number of models
			if len(result) != len(tt.expectedModels) {
				t.Errorf("GetModelsWithMinContextWindow(%d) returned %d models, want %d\nGot: %v\nWant: %v",
					tt.minTokens, len(result), len(tt.expectedModels), result, tt.expectedModels)
			}

			// Create set of expected models for easy lookup
			expectedSet := make(map[string]bool)
			for _, model := range tt.expectedModels {
				expectedSet[model] = true
			}

			// Verify all returned models are expected
			for _, modelName := range result {
				if !expectedSet[modelName] {
					t.Errorf("GetModelsWithMinContextWindow(%d) returned unexpected model: %s",
						tt.minTokens, modelName)
				}
			}

			// Verify all expected models are present
			resultSet := make(map[string]bool)
			for _, model := range result {
				resultSet[model] = true
			}
			for _, expected := range tt.expectedModels {
				if !resultSet[expected] {
					t.Errorf("GetModelsWithMinContextWindow(%d) missing expected model: %s",
						tt.minTokens, expected)
				}
			}

			// Verify descending order by context window (largest first)
			if tt.verifyOrder && len(result) > 1 {
				for i := 0; i < len(result)-1; i++ {
					info1, err1 := GetModelInfo(result[i])
					info2, err2 := GetModelInfo(result[i+1])
					if err1 != nil || err2 != nil {
						t.Errorf("GetModelsWithMinContextWindow(%d) returned invalid model names: %v, %v",
							tt.minTokens, err1, err2)
						continue
					}
					if info1.ContextWindow < info2.ContextWindow {
						t.Errorf("GetModelsWithMinContextWindow(%d) not properly sorted: "+
							"models[%d]=%s (%d tokens) < models[%d]=%s (%d tokens)",
							tt.minTokens, i, result[i], info1.ContextWindow, i+1, result[i+1], info2.ContextWindow)
					}
				}
			}

			// Verify all returned models meet the minimum threshold
			for _, modelName := range result {
				info, err := GetModelInfo(modelName)
				if err != nil {
					t.Errorf("GetModelsWithMinContextWindow(%d) returned invalid model: %s",
						tt.minTokens, modelName)
					continue
				}
				if info.ContextWindow < tt.minTokens {
					t.Errorf("GetModelsWithMinContextWindow(%d) returned model %s with context window %d < %d",
						tt.minTokens, modelName, info.ContextWindow, tt.minTokens)
				}
			}
		})
	}
}

func TestSelectModelsForInput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		estimatedTokens    int
		availableProviders []string
		expectedModels     []string
		verifyOrder        bool
	}{
		{
			name:               "small input, all providers",
			estimatedTokens:    5000,
			availableProviders: []string{"openai", "gemini", "openrouter"},
			expectedModels: []string{
				"openrouter/meta-llama/llama-4-maverick",
				"openrouter/meta-llama/llama-4-scout",
				"gemini-2.5-flash",
				"gpt-4.1",
				"gemini-2.5-pro",
				"o3",
				"o4-mini",
				"openrouter/x-ai/grok-3-mini-beta",
				"openrouter/meta-llama/llama-3.3-70b-instruct",
				"openrouter/x-ai/grok-3-beta",
				"openrouter/deepseek/deepseek-r1-0528",
				"openrouter/deepseek/deepseek-chat-v3-0324:free",
				"openrouter/deepseek/deepseek-chat-v3-0324",
				"openrouter/google/gemma-3-27b-it",
			},
			verifyOrder: true,
		},
		{
			name:               "medium input, gemini only",
			estimatedTokens:    50000,
			availableProviders: []string{"gemini"},
			expectedModels:     []string{"gemini-2.5-flash", "gemini-2.5-pro"},
			verifyOrder:        true,
		},
		{
			name:               "very large input, limited models",
			estimatedTokens:    700000,
			availableProviders: []string{"openai", "gemini", "openrouter"},
			expectedModels: []string{
				"openrouter/meta-llama/llama-4-maverick",
				"openrouter/meta-llama/llama-4-scout",
				"gemini-2.5-pro",
				"gpt-4.1",
				"gemini-2.5-flash",
			},
			verifyOrder: true,
		},
		{
			name:               "openai only",
			estimatedTokens:    10000,
			availableProviders: []string{"openai"},
			expectedModels:     []string{"gpt-4.1", "o3", "o4-mini"},
			verifyOrder:        true,
		},
		{
			name:               "empty providers",
			estimatedTokens:    10000,
			availableProviders: []string{},
			expectedModels:     []string{},
			verifyOrder:        false,
		},
		{
			name:               "unknown provider",
			estimatedTokens:    10000,
			availableProviders: []string{"unknown-provider"},
			expectedModels:     []string{},
			verifyOrder:        false,
		},
		{
			name:               "mixed valid and invalid providers",
			estimatedTokens:    10000,
			availableProviders: []string{"openai", "invalid", "gemini"},
			expectedModels: []string{
				"gemini-2.5-flash",
				"gpt-4.1",
				"gemini-2.5-pro",
				"o3",
				"o4-mini",
			},
			verifyOrder: true,
		},
		{
			name:               "extremely large input, only highest capacity",
			estimatedTokens:    800000,
			availableProviders: []string{"openai", "gemini", "openrouter"},
			expectedModels: []string{
				"openrouter/meta-llama/llama-4-maverick",
				"openrouter/meta-llama/llama-4-scout",
				"gemini-2.5-pro",
				"gpt-4.1",
				"gemini-2.5-flash",
			},
			verifyOrder: true,
		},
		{
			name:               "openrouter only",
			estimatedTokens:    20000,
			availableProviders: []string{"openrouter"},
			expectedModels: []string{
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
			verifyOrder: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SelectModelsForInput(tt.estimatedTokens, tt.availableProviders)

			// Verify correct number of models
			if len(result) != len(tt.expectedModels) {
				t.Errorf("SelectModelsForInput(%d, %v) returned %d models, want %d\nGot: %v\nWant: %v",
					tt.estimatedTokens, tt.availableProviders, len(result), len(tt.expectedModels), result, tt.expectedModels)
			}

			// Create set of expected models for easy lookup
			expectedSet := make(map[string]bool)
			for _, model := range tt.expectedModels {
				expectedSet[model] = true
			}

			// Verify all returned models are expected
			for _, modelName := range result {
				if !expectedSet[modelName] {
					t.Errorf("SelectModelsForInput(%d, %v) returned unexpected model: %s",
						tt.estimatedTokens, tt.availableProviders, modelName)
				}
			}

			// Verify all expected models are present
			resultSet := make(map[string]bool)
			for _, model := range result {
				resultSet[model] = true
			}
			for _, expected := range tt.expectedModels {
				if !resultSet[expected] {
					t.Errorf("SelectModelsForInput(%d, %v) missing expected model: %s",
						tt.estimatedTokens, tt.availableProviders, expected)
				}
			}

			// Verify all returned models are from available providers
			for _, modelName := range result {
				info, err := GetModelInfo(modelName)
				if err != nil {
					t.Errorf("SelectModelsForInput(%d, %v) returned invalid model: %s",
						tt.estimatedTokens, tt.availableProviders, modelName)
					continue
				}

				providerFound := false
				for _, provider := range tt.availableProviders {
					if info.Provider == provider {
						providerFound = true
						break
					}
				}

				if !providerFound {
					t.Errorf("SelectModelsForInput(%d, %v) returned model %s with provider %s not in available providers",
						tt.estimatedTokens, tt.availableProviders, modelName, info.Provider)
				}
			}

			// Verify all models can handle the input with safety margin (1.25x)
			requiredContext := int(float64(tt.estimatedTokens) * 1.25)
			for _, modelName := range result {
				info, err := GetModelInfo(modelName)
				if err != nil {
					continue // Already checked above
				}
				if info.ContextWindow < requiredContext {
					t.Errorf("SelectModelsForInput(%d, %v) returned model %s with insufficient context: %d < %d (required)",
						tt.estimatedTokens, tt.availableProviders, modelName, info.ContextWindow, requiredContext)
				}
			}

			// Verify descending order by context window if requested
			if tt.verifyOrder && len(result) > 1 {
				for i := 0; i < len(result)-1; i++ {
					info1, err1 := GetModelInfo(result[i])
					info2, err2 := GetModelInfo(result[i+1])
					if err1 != nil || err2 != nil {
						continue // Already checked above
					}
					if info1.ContextWindow < info2.ContextWindow {
						t.Errorf("SelectModelsForInput(%d, %v) not properly sorted: "+
							"models[%d]=%s (%d tokens) < models[%d]=%s (%d tokens)",
							tt.estimatedTokens, tt.availableProviders, i, result[i], info1.ContextWindow,
							i+1, result[i+1], info2.ContextWindow)
					}
				}
			}
		})
	}
}

func TestGetLargestContextModel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		modelNames    []string
		expectedModel string
	}{
		{
			name:          "empty list returns empty string",
			modelNames:    []string{},
			expectedModel: "",
		},
		{
			name:          "single model returns that model",
			modelNames:    []string{"gpt-4.1"},
			expectedModel: "gpt-4.1",
		},
		{
			name:          "multiple OpenAI models",
			modelNames:    []string{"o4-mini", "gpt-4.1", "o3"},
			expectedModel: "gpt-4.1", // 1M context vs 200k for others
		},
		{
			name: "mixed providers, largest context",
			modelNames: []string{
				"gemini-2.5-flash",
				"openrouter/deepseek/deepseek-chat-v3-0324",
				"gpt-4.1",
				"openrouter/meta-llama/llama-4-maverick",
			},
			expectedModel: "openrouter/meta-llama/llama-4-maverick", // 1048576 context window
		},
		{
			name: "all small context models",
			modelNames: []string{
				"openrouter/google/gemma-3-27b-it",
				"openrouter/deepseek/deepseek-chat-v3-0324",
				"openrouter/deepseek/deepseek-chat-v3-0324:free",
			},
			expectedModel: "openrouter/deepseek/deepseek-chat-v3-0324", // 65536 context (same as :free version, but first in list)
		},
		{
			name: "models with identical context windows",
			modelNames: []string{
				"gemini-2.5-pro",
				"gemini-2.5-flash",
				"gpt-4.1",
			},
			expectedModel: "gemini-2.5-pro", // All have 1M context, first one wins
		},
		{
			name: "include invalid model names",
			modelNames: []string{
				"invalid-model",
				"gpt-4.1",
				"another-invalid",
				"o4-mini",
			},
			expectedModel: "gpt-4.1", // Largest valid model (1M vs 200k)
		},
		{
			name:          "all invalid model names",
			modelNames:    []string{"invalid-1", "invalid-2", "invalid-3"},
			expectedModel: "",
		},
		{
			name: "largest possible context models",
			modelNames: []string{
				"openrouter/meta-llama/llama-4-maverick",
				"openrouter/meta-llama/llama-4-scout",
			},
			expectedModel: "openrouter/meta-llama/llama-4-maverick", // Both have same context, first wins
		},
		{
			name: "comprehensive mix of all model types",
			modelNames: []string{
				"o4-mini",                              // 200k
				"openrouter/google/gemma-3-27b-it",     // 8192
				"gemini-2.5-flash",                     // 1M
				"openrouter/deepseek/deepseek-r1-0528", // 128k
				"openrouter/meta-llama/llama-4-scout",  // 1048576 (largest)
				"openrouter/x-ai/grok-3-beta",          // 131072
			},
			expectedModel: "openrouter/meta-llama/llama-4-scout", // 1048576 context window
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLargestContextModel(tt.modelNames)

			if result != tt.expectedModel {
				t.Errorf("GetLargestContextModel(%v) = %s, want %s", tt.modelNames, result, tt.expectedModel)
			}

			// Additional verification: if result is non-empty, ensure it's actually the largest
			if result != "" && len(tt.modelNames) > 1 {
				resultInfo, err := GetModelInfo(result)
				if err != nil {
					t.Errorf("GetLargestContextModel(%v) returned invalid model: %s", tt.modelNames, result)
				} else {
					// Check that no other valid model has a larger context
					for _, modelName := range tt.modelNames {
						if modelName == result {
							continue
						}
						info, err := GetModelInfo(modelName)
						if err != nil {
							continue // Skip invalid models
						}
						if info.ContextWindow > resultInfo.ContextWindow {
							t.Errorf("GetLargestContextModel(%v) returned %s (%d tokens) but %s has larger context (%d tokens)",
								tt.modelNames, result, resultInfo.ContextWindow, modelName, info.ContextWindow)
						}
					}
				}
			}

			// Verify result is in the input list (if non-empty)
			if result != "" {
				found := false
				for _, modelName := range tt.modelNames {
					if modelName == result {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetLargestContextModel(%v) returned %s which is not in the input list", tt.modelNames, result)
				}
			}
		})
	}
}
