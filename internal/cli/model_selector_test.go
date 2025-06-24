// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSelectOptimalModel tests the core model selection logic
func TestSelectOptimalModel(t *testing.T) {
	tests := []struct {
		name               string
		availableProviders []string
		taskSize           int64
		expectedModel      string
	}{
		{
			name:               "no providers available",
			availableProviders: []string{},
			taskSize:           1000,
			expectedModel:      DefaultModel,
		},
		{
			name:               "small task with openai",
			availableProviders: []string{"openai"},
			taskSize:           5000,
			expectedModel:      "gpt-4o",
		},
		{
			name:               "medium task with gemini",
			availableProviders: []string{"gemini"},
			taskSize:           50000,
			expectedModel:      "gemini-2.5-pro",
		},
		{
			name:               "large task with multiple providers",
			availableProviders: []string{"openai", "gemini", "openrouter"},
			taskSize:           150000,
			expectedModel:      "gemini-2.5-pro", // Highest score
		},
		{
			name:               "only openrouter available",
			availableProviders: []string{"openrouter"},
			taskSize:           25000,
			expectedModel:      "claude-3-opus",
		},
		{
			name:               "all providers available small task",
			availableProviders: []string{"openai", "gemini", "openrouter"},
			taskSize:           8000,
			expectedModel:      "gemini-2.5-pro", // Best available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SelectOptimalModel(tt.availableProviders, tt.taskSize)
			assert.Equal(t, tt.expectedModel, result)
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
	for _, envVar := range GetProviderEnvVars() {
		require.NoError(t, os.Unsetenv(envVar))
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
		assert.Equal(t, []string{"openai"}, providers)
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

// TestSelectBestModel tests the main entry point
func TestSelectBestModel(t *testing.T) {
	// Clear environment first
	for _, envVar := range GetProviderEnvVars() {
		require.NoError(t, os.Unsetenv(envVar))
	}
	ClearProviderCache()

	t.Run("no api keys returns default", func(t *testing.T) {
		model := SelectBestModel(10000)
		assert.Equal(t, DefaultModel, model)
	})

	t.Run("with api key returns optimal model", func(t *testing.T) {
		require.NoError(t, os.Setenv("GEMINI_API_KEY", "gemini-test-key-123456"))
		defer func() { require.NoError(t, os.Unsetenv("GEMINI_API_KEY")) }()
		ClearProviderCache()

		model := SelectBestModel(50000)
		assert.Equal(t, "gemini-2.5-pro", model)
	})
}

// TestValidateModelAvailability tests model validation
func TestValidateModelAvailability(t *testing.T) {
	// Clear environment
	for _, envVar := range GetProviderEnvVars() {
		require.NoError(t, os.Unsetenv(envVar))
	}
	ClearProviderCache()

	t.Run("known model without provider", func(t *testing.T) {
		available := ValidateModelAvailability("gpt-4o")
		assert.False(t, available)
	})

	t.Run("known model with provider", func(t *testing.T) {
		require.NoError(t, os.Setenv("OPENAI_API_KEY", "sk-test123456789"))
		defer func() { require.NoError(t, os.Unsetenv("OPENAI_API_KEY")) }()
		ClearProviderCache()

		available := ValidateModelAvailability("gpt-4o")
		assert.True(t, available)
	})

	t.Run("unknown model with any provider", func(t *testing.T) {
		require.NoError(t, os.Setenv("GEMINI_API_KEY", "gemini-test-key-123456"))
		defer func() { require.NoError(t, os.Unsetenv("GEMINI_API_KEY")) }()
		ClearProviderCache()

		available := ValidateModelAvailability("custom-model-name")
		assert.True(t, available) // Should return true if any provider available
	})
}

// TestGetModelProvider tests provider inference
func TestGetModelProvider(t *testing.T) {
	tests := []struct {
		name             string
		modelName        string
		expectedProvider string
	}{
		{
			name:             "known gemini model",
			modelName:        "gemini-2.5-pro",
			expectedProvider: "gemini",
		},
		{
			name:             "known openai model",
			modelName:        "gpt-4o",
			expectedProvider: "openai",
		},
		{
			name:             "known openrouter model",
			modelName:        "claude-3-opus",
			expectedProvider: "openrouter",
		},
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

// TestModelRankings tests that model rankings are properly sorted
func TestModelRankings(t *testing.T) {
	require.NotEmpty(t, GetModelRankings())

	// Verify rankings are in descending order of score
	rankings := GetModelRankings()
	for i := 1; i < len(rankings); i++ {
		assert.GreaterOrEqual(t, rankings[i-1].Score, rankings[i].Score,
			"Model rankings should be in descending order of score")
	}

	// Verify all required fields are present
	for _, model := range rankings {
		assert.NotEmpty(t, model.Name, "Model name should not be empty")
		assert.NotEmpty(t, model.Provider, "Model provider should not be empty")
		assert.Greater(t, model.Score, 0, "Model score should be positive")
	}
}

// TestProviderEnvVars tests environment variable mapping
func TestProviderEnvVars(t *testing.T) {
	expectedProviders := []string{"openai", "gemini", "openrouter"}
	envVars := GetProviderEnvVars()

	for _, provider := range expectedProviders {
		envVar, exists := envVars[provider]
		assert.True(t, exists, "Provider %s should have env var mapping", provider)
		assert.NotEmpty(t, envVar, "Env var for %s should not be empty", provider)
		assert.Contains(t, envVar, "API_KEY", "Env var should contain API_KEY")
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

// BenchmarkSelectOptimalModel benchmarks the model selection performance
func BenchmarkSelectOptimalModel(b *testing.B) {
	providers := []string{"openai", "gemini", "openrouter"}
	taskSize := int64(50000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SelectOptimalModel(providers, taskSize)
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
	providers := []string{"openai", "gemini", "openrouter"}
	taskSizes := []int64{1000, 10000, 50000, 100000, 200000}

	// Test that same inputs always produce same outputs
	for _, taskSize := range taskSizes {
		model1 := SelectOptimalModel(providers, taskSize)
		model2 := SelectOptimalModel(providers, taskSize)
		assert.Equal(t, model1, model2,
			"Model selection should be deterministic for taskSize %d", taskSize)
	}

	// Test that larger tasks prefer higher-scored models when available
	smallTask := SelectOptimalModel(providers, 5000)
	largeTask := SelectOptimalModel(providers, 150000)

	// Find scores for selected models
	var smallScore, largeScore int
	rankings := GetModelRankings()
	for _, model := range rankings {
		if model.Name == smallTask {
			smallScore = model.Score
		}
		if model.Name == largeTask {
			largeScore = model.Score
		}
	}

	assert.GreaterOrEqual(t, largeScore, smallScore,
		"Large tasks should prefer higher-scored models")
}
