// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/models"
)

// ProviderModelInfo represents information about a model and its provider availability
type ProviderModelInfo struct {
	Name      string
	Provider  string
	Available bool
}

// ProviderCache caches provider availability to avoid repeated env var lookups
type ProviderCache struct {
	mu         sync.RWMutex
	cache      map[string]bool
	lastUpdate time.Time
	ttl        time.Duration
}

// newProviderCache creates a new provider cache with 5-minute TTL
func newProviderCache() *ProviderCache {
	return &ProviderCache{
		cache: make(map[string]bool),
		ttl:   5 * time.Minute,
	}
}

// Global cache instance
var providerCache = newProviderCache()

// getAvailableModelsWithContext returns all models that have sufficient context window for the input
// and whose providers have valid API keys. This replaces the old ranking system with simple filtering.
func getAvailableModelsWithContext(inputTokens int64) []string {
	allModels := models.ListAllModels()
	var availableModels []string

	for _, modelName := range allModels {
		modelInfo, err := models.GetModelInfo(modelName)
		if err != nil {
			continue // Skip unknown models
		}

		// Check if model has sufficient context window
		// Reserve some tokens for output (use 80% of context window for input)
		maxInputTokens := int64(float64(modelInfo.ContextWindow) * 0.8)
		if inputTokens > maxInputTokens {
			continue // Skip models with insufficient context
		}

		// Check if provider has valid API key
		if providerCache.isProviderAvailable(modelInfo.Provider) {
			availableModels = append(availableModels, modelName)
		}
	}

	// Sort for deterministic results
	sort.Strings(availableModels)
	return availableModels
}

// isProviderAvailable checks if a provider has valid API key in environment
func (pc *ProviderCache) isProviderAvailable(provider string) bool {
	pc.mu.RLock()

	// Check if cache is still valid
	if time.Since(pc.lastUpdate) < pc.ttl {
		if available, exists := pc.cache[provider]; exists {
			pc.mu.RUnlock()
			return available
		}
	}
	pc.mu.RUnlock()

	// Cache miss or expired, update cache
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Double-check after acquiring write lock
	if time.Since(pc.lastUpdate) < pc.ttl {
		if available, exists := pc.cache[provider]; exists {
			return available
		}
	}

	// Update cache
	pc.cache = make(map[string]bool)
	pc.lastUpdate = time.Now()

	// Check all providers by iterating through all models to find unique providers
	providers := make(map[string]bool)
	allModels := models.ListAllModels()
	for _, modelName := range allModels {
		if modelInfo, err := models.GetModelInfo(modelName); err == nil {
			providers[modelInfo.Provider] = true
		}
	}

	// Check each provider for API key
	for providerName := range providers {
		envVar := models.GetAPIKeyEnvVar(providerName)
		if envVar != "" {
			apiKey := strings.TrimSpace(os.Getenv(envVar))
			pc.cache[providerName] = apiKey != "" && len(apiKey) > 10 // Basic validation
		}
	}

	available, exists := pc.cache[provider]
	return exists && available
}

// GetAvailableProviders returns list of providers with valid API keys
func GetAvailableProviders() []string {
	providers := make(map[string]bool)
	allModels := models.ListAllModels()

	// Find all unique providers
	for _, modelName := range allModels {
		if modelInfo, err := models.GetModelInfo(modelName); err == nil {
			providers[modelInfo.Provider] = true
		}
	}

	// Check which providers are available
	var available []string
	for provider := range providers {
		if providerCache.isProviderAvailable(provider) {
			available = append(available, provider)
		}
	}
	sort.Strings(available) // Ensure deterministic ordering
	return available
}

// SelectOptimalModels selects ALL available models that can handle the input token count.
// This implements the user's requirement: "send 100% of the models that have a context window greater than the input token count"
func SelectOptimalModels(inputTokens int64) []string {
	availableModels := getAvailableModelsWithContext(inputTokens)

	// For token-based filtering, we return exactly what the filter produces
	// No fallback to default model if context windows are insufficient
	return availableModels
}

// SelectBestModels is the main entry point for multi-model selection
// It returns all models that can handle the input token count
func SelectBestModels(inputTokens int64) []string {
	return SelectOptimalModels(inputTokens)
}

// SelectBestModel provides backward compatibility - returns the first available model
// This function should be deprecated in favor of SelectBestModels
func SelectBestModel(inputTokens int64) string {
	models := SelectBestModels(inputTokens)
	if len(models) > 0 {
		return models[0]
	}
	// Fallback to default model only for backward compatibility
	return config.DefaultModel
}

// ValidateModelAvailability checks if a specific model is available
func ValidateModelAvailability(modelName string) bool {
	// Use the models package to get model info
	modelInfo, err := models.GetModelInfo(modelName)
	if err != nil {
		// For unknown models, check if any provider is available
		availableProviders := GetAvailableProviders()
		return len(availableProviders) > 0
	}

	return providerCache.isProviderAvailable(modelInfo.Provider)
}

// GetModelProvider returns the provider for a given model name
func GetModelProvider(modelName string) string {
	// Use the models package to get model info
	modelInfo, err := models.GetModelInfo(modelName)
	if err != nil {
		// For unknown models, try to infer provider from name patterns
		if strings.Contains(modelName, "gpt") || strings.Contains(modelName, "openai") {
			return "openai"
		}
		if strings.Contains(modelName, "gemini") || strings.Contains(modelName, "google") {
			return "gemini"
		}
		if strings.Contains(modelName, "claude") || strings.Contains(modelName, "anthropic") || strings.Contains(modelName, "openrouter") {
			return "openrouter"
		}
		return "unknown"
	}

	return modelInfo.Provider
}

// ClearProviderCache forces a refresh of the provider availability cache
// Useful for testing or when environment changes
func ClearProviderCache() {
	providerCache.mu.Lock()
	defer providerCache.mu.Unlock()
	providerCache.cache = make(map[string]bool)
	providerCache.lastUpdate = time.Time{}
}

// GetAllSupportedModels returns all supported models for testing
func GetAllSupportedModels() []string {
	return models.ListAllModels()
}

// GetModelsForProvider returns models for a specific provider for testing
func GetModelsForProvider(provider string) []string {
	return models.ListModelsForProvider(provider)
}
