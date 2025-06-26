// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/phrazzld/thinktank/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSelectOptimalModels tests the new token-based model filtering logic
func TestSelectOptimalModels(t *testing.T) {
	tests := []struct {
		name         string
		inputTokens  int64
		setupEnv     func()
		cleanupEnv   func()
		expectEmpty  bool
		expectModels []string // specific models we expect to see
		minModels    int      // minimum number of models expected
	}{
		{
			name:        "very large input - no models can handle",
			inputTokens: 2000000, // 2M tokens - exceeds all model context windows
			setupEnv: func() {
				// Keep existing API keys for this test - we want to verify context window filtering
			},
			cleanupEnv: func() {
				// No cleanup needed
			},
			expectEmpty: true,
		},
		{
			name:        "small input - all models available",
			inputTokens: 1000, // Very small input
			setupEnv: func() {
				_ = os.Setenv("OPENAI_API_KEY", "sk-test123456789")
				_ = os.Setenv("GEMINI_API_KEY", "gemini-test-key-123456")
				_ = os.Setenv("OPENROUTER_API_KEY", "openrouter-test-key-123456")
			},
			cleanupEnv: func() {
				_ = os.Unsetenv("OPENAI_API_KEY")
				_ = os.Unsetenv("GEMINI_API_KEY")
				_ = os.Unsetenv("OPENROUTER_API_KEY")
			},
			minModels: 5, // Should have multiple models available
		},
		{
			name:        "medium input - most models available",
			inputTokens: 50000, // 50k tokens
			setupEnv: func() {
				_ = os.Setenv("OPENAI_API_KEY", "sk-test123456789")
				_ = os.Setenv("GEMINI_API_KEY", "gemini-test-key-123456")
				_ = os.Setenv("OPENROUTER_API_KEY", "openrouter-test-key-123456")
			},
			cleanupEnv: func() {
				_ = os.Unsetenv("OPENAI_API_KEY")
				_ = os.Unsetenv("GEMINI_API_KEY")
				_ = os.Unsetenv("OPENROUTER_API_KEY")
			},
			expectModels: []string{"gpt-4.1", "gemini-2.5-pro", "gemini-2.5-flash"}, // High-capacity models
			minModels:    3,
		},
		{
			name:        "no API keys - fallback behavior",
			inputTokens: 10000,
			setupEnv: func() {
				// Clear ALL possible API keys
				_ = os.Unsetenv("OPENAI_API_KEY")
				_ = os.Unsetenv("GEMINI_API_KEY")
				_ = os.Unsetenv("OPENROUTER_API_KEY")
				_ = os.Unsetenv("ANTHROPIC_API_KEY")
			},
			cleanupEnv: func() {
				// Restore API keys for other tests
				_ = os.Setenv("OPENAI_API_KEY", "sk-test123456789")
				_ = os.Setenv("GEMINI_API_KEY", "gemini-test-key-123456")
				_ = os.Setenv("OPENROUTER_API_KEY", "openrouter-test-key-123456")
			},
			expectEmpty: true, // Should return empty when no providers available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ClearProviderCache()
			tt.setupEnv()
			defer tt.cleanupEnv()

			result := SelectOptimalModels(tt.inputTokens)

			if tt.expectEmpty {
				assert.Empty(t, result, "Expected no models for input tokens %d", tt.inputTokens)
			} else {
				assert.NotEmpty(t, result, "Expected models for input tokens %d", tt.inputTokens)

				if tt.minModels > 0 {
					assert.GreaterOrEqual(t, len(result), tt.minModels, "Expected at least %d models", tt.minModels)
				}

				for _, expectedModel := range tt.expectModels {
					assert.Contains(t, result, expectedModel, "Expected model %s in result", expectedModel)
				}

				// Verify all returned models actually exist in the models package
				for _, modelName := range result {
					_, err := models.GetModelInfo(modelName)
					assert.NoError(t, err, "Model %s should exist in models package", modelName)
				}

				// Verify all returned models have sufficient context window
				for _, modelName := range result {
					modelInfo, err := models.GetModelInfo(modelName)
					require.NoError(t, err)
					maxInputTokens := int64(float64(modelInfo.ContextWindow) * 0.8)
					assert.Greater(t, maxInputTokens, tt.inputTokens,
						"Model %s context window (%d * 0.8 = %d) should be greater than input tokens %d",
						modelName, modelInfo.ContextWindow, maxInputTokens, tt.inputTokens)
				}
			}
		})
	}
}

// TestProviderCache tests the caching mechanism
func TestProviderCache(t *testing.T) {
	// Test with the global cache - we'll use environment manipulation

	t.Run("cache hit after provider check", func(t *testing.T) {
		// Clear cache first
		ClearProviderCache()

		// Set environment variable
		require.NoError(t, os.Setenv("OPENAI_API_KEY", "test-key-12345"))
		defer func() { require.NoError(t, os.Unsetenv("OPENAI_API_KEY")) }()

		// First call should populate cache
		providers := GetAvailableProviders()
		assert.Contains(t, providers, "openai")

		// Second call should use cache (we can't directly test this but it should work)
		providers2 := GetAvailableProviders()
		assert.Equal(t, providers, providers2)
	})

	t.Run("cache invalidation works", func(t *testing.T) {
		// Set up environment
		require.NoError(t, os.Setenv("GEMINI_API_KEY", "gemini-key-12345"))
		defer func() { require.NoError(t, os.Unsetenv("GEMINI_API_KEY")) }()

		// Get initial providers
		providers := GetAvailableProviders()
		assert.Contains(t, providers, "gemini")

		// Clear cache and remove env var
		ClearProviderCache()
		require.NoError(t, os.Unsetenv("GEMINI_API_KEY"))

		// Should now return empty
		providers = GetAvailableProviders()
		assert.NotContains(t, providers, "gemini")
	})
}

// TestGetAvailableProviders tests provider detection
func TestGetAvailableProviders(t *testing.T) {
	// Clear environment
	allModels := models.ListAllModels()
	for _, modelName := range allModels {
		if modelInfo, err := models.GetModelInfo(modelName); err == nil {
			envVar := models.GetAPIKeyEnvVar(modelInfo.Provider)
			if envVar != "" {
				require.NoError(t, os.Unsetenv(envVar))
			}
		}
	}
	ClearProviderCache()

	t.Run("no providers available", func(t *testing.T) {
		providers := GetAvailableProviders()
		assert.Empty(t, providers)
	})

	t.Run("single provider available", func(t *testing.T) {
		require.NoError(t, os.Setenv("OPENAI_API_KEY", "sk-test123456789"))
		defer func() { require.NoError(t, os.Unsetenv("OPENAI_API_KEY")) }()
		ClearProviderCache()

		providers := GetAvailableProviders()
		assert.Contains(t, providers, "openai")
	})

	t.Run("multiple providers available", func(t *testing.T) {
		require.NoError(t, os.Setenv("OPENAI_API_KEY", "sk-test123456789"))
		require.NoError(t, os.Setenv("GEMINI_API_KEY", "gemini-key-123456"))
		defer func() {
			require.NoError(t, os.Unsetenv("OPENAI_API_KEY"))
			require.NoError(t, os.Unsetenv("GEMINI_API_KEY"))
		}()
		ClearProviderCache()

		providers := GetAvailableProviders()
		assert.Len(t, providers, 2)
		assert.Contains(t, providers, "openai")
		assert.Contains(t, providers, "gemini")
	})

	t.Run("invalid api key ignored", func(t *testing.T) {
		require.NoError(t, os.Setenv("OPENAI_API_KEY", "short")) // Too short
		defer func() { require.NoError(t, os.Unsetenv("OPENAI_API_KEY")) }()
		ClearProviderCache()

		providers := GetAvailableProviders()
		assert.Empty(t, providers)
	})
}

// TestSelectBestModels tests the main entry point for multi-model selection
func TestSelectBestModels(t *testing.T) {
	// Clear environment first
	allModels := models.ListAllModels()
	for _, modelName := range allModels {
		if modelInfo, err := models.GetModelInfo(modelName); err == nil {
			envVar := models.GetAPIKeyEnvVar(modelInfo.Provider)
			if envVar != "" {
				require.NoError(t, os.Unsetenv(envVar))
			}
		}
	}
	ClearProviderCache()

	t.Run("no api keys returns empty", func(t *testing.T) {
		models := SelectBestModels(10000)
		assert.Empty(t, models)
	})

	t.Run("with api key returns filtered models", func(t *testing.T) {
		require.NoError(t, os.Setenv("GEMINI_API_KEY", "gemini-test-key-123456"))
		defer func() { require.NoError(t, os.Unsetenv("GEMINI_API_KEY")) }()
		ClearProviderCache()

		models := SelectBestModels(50000)
		assert.NotEmpty(t, models, "Should return models when API key available")
		// All returned models should be Gemini models since only Gemini API key is set
		for _, modelName := range models {
			provider := GetModelProvider(modelName)
			assert.Equal(t, "gemini", provider, "All models should be from gemini provider")
		}
	})

	t.Run("backward compatibility - SelectBestModel", func(t *testing.T) {
		require.NoError(t, os.Setenv("OPENAI_API_KEY", "sk-test123456789"))
		defer func() { require.NoError(t, os.Unsetenv("OPENAI_API_KEY")) }()
		ClearProviderCache()

		model := SelectBestModel(10000)
		assert.NotEmpty(t, model, "SelectBestModel should return first available model")
		provider := GetModelProvider(model)
		assert.Equal(t, "openai", provider, "Returned model should be from openai provider")
	})
}

// TestValidateModelAvailability tests model validation
func TestValidateModelAvailability(t *testing.T) {
	// Clear environment
	allModels := models.ListAllModels()
	for _, modelName := range allModels {
		if modelInfo, err := models.GetModelInfo(modelName); err == nil {
			envVar := models.GetAPIKeyEnvVar(modelInfo.Provider)
			if envVar != "" {
				require.NoError(t, os.Unsetenv(envVar))
			}
		}
	}
	ClearProviderCache()

	t.Run("actual model without provider", func(t *testing.T) {
		// Use an actual model from the models package
		allModels := models.ListAllModels()
		require.NotEmpty(t, allModels, "Should have at least one model defined")
		modelName := allModels[0]

		available := ValidateModelAvailability(modelName)
		assert.False(t, available)
	})

	t.Run("actual model with provider", func(t *testing.T) {
		// Find an OpenAI model and test with OpenAI API key
		openaiModels := models.ListModelsForProvider("openai")
		if len(openaiModels) > 0 {
			require.NoError(t, os.Setenv("OPENAI_API_KEY", "sk-test123456789"))
			defer func() { require.NoError(t, os.Unsetenv("OPENAI_API_KEY")) }()
			ClearProviderCache()

			available := ValidateModelAvailability(openaiModels[0])
			assert.True(t, available)
		} else {
			t.Skip("No OpenAI models available for testing")
		}
	})

	t.Run("unknown model with any provider", func(t *testing.T) {
		require.NoError(t, os.Setenv("GEMINI_API_KEY", "gemini-test-key-123456"))
		defer func() { require.NoError(t, os.Unsetenv("GEMINI_API_KEY")) }()
		ClearProviderCache()

		available := ValidateModelAvailability("custom-model-name")
		assert.True(t, available) // Should return true if any provider available
	})
}

// TestGetModelProvider tests provider inference using actual models
func TestGetModelProvider(t *testing.T) {
	// Test with actual models from the models package
	allModels := models.ListAllModels()
	require.NotEmpty(t, allModels, "Should have models defined")

	for _, modelName := range allModels {
		t.Run(fmt.Sprintf("actual model %s", modelName), func(t *testing.T) {
			provider := GetModelProvider(modelName)

			// Get expected provider from models package
			modelInfo, err := models.GetModelInfo(modelName)
			require.NoError(t, err)
			expectedProvider := modelInfo.Provider

			assert.Equal(t, expectedProvider, provider, "Provider should match models package")
		})
	}

	// Test inference for unknown models
	tests := []struct {
		name             string
		modelName        string
		expectedProvider string
	}{
		{
			name:             "unknown gpt model",
			modelName:        "gpt-5",
			expectedProvider: "openai",
		},
		{
			name:             "unknown gemini model",
			modelName:        "gemini-ultra",
			expectedProvider: "gemini",
		},
		{
			name:             "unknown claude model",
			modelName:        "claude-4",
			expectedProvider: "openrouter",
		},
		{
			name:             "completely unknown model",
			modelName:        "mystery-model",
			expectedProvider: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := GetModelProvider(tt.modelName)
			assert.Equal(t, tt.expectedProvider, provider)
		})
	}
}

// TestAllSupportedModels tests that we can access all models from the models package
func TestAllSupportedModels(t *testing.T) {
	allModels := GetAllSupportedModels()
	require.NotEmpty(t, allModels, "Should have at least one model")

	// Verify all models exist in the models package
	for _, modelName := range allModels {
		modelInfo, err := models.GetModelInfo(modelName)
		assert.NoError(t, err, "Model %s should exist in models package", modelName)
		assert.NotEmpty(t, modelInfo.Provider, "Model %s should have a provider", modelName)
		assert.Greater(t, modelInfo.ContextWindow, 0, "Model %s should have positive context window", modelName)
	}

	// Verify models are sorted (deterministic ordering)
	sortedModels := make([]string, len(allModels))
	copy(sortedModels, allModels)
	sort.Strings(sortedModels)
	assert.Equal(t, sortedModels, allModels, "Models should be sorted alphabetically")
}

// TestModelsForProvider tests provider-specific model listing
func TestModelsForProvider(t *testing.T) {
	// Get all unique providers from models
	allModels := models.ListAllModels()
	providers := make(map[string]bool)
	for _, modelName := range allModels {
		if modelInfo, err := models.GetModelInfo(modelName); err == nil {
			providers[modelInfo.Provider] = true
		}
	}

	for provider := range providers {
		t.Run(fmt.Sprintf("provider %s", provider), func(t *testing.T) {
			providerModels := GetModelsForProvider(provider)
			assert.NotEmpty(t, providerModels, "Provider %s should have models", provider)

			// Verify all returned models are actually from this provider
			for _, modelName := range providerModels {
				modelInfo, err := models.GetModelInfo(modelName)
				require.NoError(t, err)
				assert.Equal(t, provider, modelInfo.Provider, "Model %s should be from provider %s", modelName, provider)
			}

			// Verify models are sorted
			sortedModels := make([]string, len(providerModels))
			copy(sortedModels, providerModels)
			sort.Strings(sortedModels)
			assert.Equal(t, sortedModels, providerModels, "Provider models should be sorted")
		})
	}
}

// TestClearProviderCache tests cache clearing functionality
func TestClearProviderCache(t *testing.T) {
	// Clean slate first
	ClearProviderCache()

	// Populate cache
	require.NoError(t, os.Setenv("OPENAI_API_KEY", "test-key-12345"))
	defer func() { require.NoError(t, os.Unsetenv("OPENAI_API_KEY")) }()

	// First call should populate cache
	providers := GetAvailableProviders()
	assert.Contains(t, providers, "openai")

	// Clear cache - this is the functionality we're testing
	ClearProviderCache()

	// After clearing, env var should still work (cache is repopulated)
	providers = GetAvailableProviders()
	assert.Contains(t, providers, "openai")
}

// BenchmarkSelectOptimalModels benchmarks the new model selection performance
func BenchmarkSelectOptimalModels(b *testing.B) {
	taskSize := int64(50000)

	// Set up environment for realistic benchmark
	_ = os.Setenv("OPENAI_API_KEY", "sk-test123456789")
	_ = os.Setenv("GEMINI_API_KEY", "gemini-test-key-123456")
	defer func() {
		_ = os.Unsetenv("OPENAI_API_KEY")
		_ = os.Unsetenv("GEMINI_API_KEY")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SelectOptimalModels(taskSize)
	}
}

// BenchmarkProviderCacheHit benchmarks cache hit performance
func BenchmarkProviderCacheHit(b *testing.B) {
	// Set up environment for realistic benchmark
	_ = os.Setenv("OPENAI_API_KEY", "test-key")          // nolint:errcheck // Ignore errors in benchmarks
	defer func() { _ = os.Unsetenv("OPENAI_API_KEY") }() // nolint:errcheck

	// Prime the cache
	GetAvailableProviders()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetAvailableProviders()
	}
}

// Property-based test for model selection consistency
func TestModelSelectionConsistency(t *testing.T) {
	// Set up environment
	require.NoError(t, os.Setenv("OPENAI_API_KEY", "sk-test123456789"))
	require.NoError(t, os.Setenv("GEMINI_API_KEY", "gemini-test-key-123456"))
	defer func() {
		_ = os.Unsetenv("OPENAI_API_KEY")
		_ = os.Unsetenv("GEMINI_API_KEY")
	}()

	taskSizes := []int64{1000, 10000, 50000, 100000, 200000}

	// Test that same inputs always produce same outputs
	for _, taskSize := range taskSizes {
		t.Run(fmt.Sprintf("consistency for %d tokens", taskSize), func(t *testing.T) {
			ClearProviderCache()
			models1 := SelectOptimalModels(taskSize)
			models2 := SelectOptimalModels(taskSize)
			assert.Equal(t, models1, models2,
				"Model selection should be deterministic for taskSize %d", taskSize)
		})
	}

	// Test that larger inputs result in fewer available models
	ClearProviderCache()
	smallTaskModels := SelectOptimalModels(1000)
	largeTaskModels := SelectOptimalModels(500000) // Much larger input

	// Large tasks should have fewer or equal available models (due to context window limits)
	assert.LessOrEqual(t, len(largeTaskModels), len(smallTaskModels),
		"Large tasks should have fewer or equal available models due to context window constraints")

	// All models returned for large tasks should also be available for small tasks
	for _, largeModel := range largeTaskModels {
		assert.Contains(t, smallTaskModels, largeModel,
			"Models available for large tasks should also be available for small tasks")
	}
}
