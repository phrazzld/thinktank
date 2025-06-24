// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// ModelInfo represents information about a model and its provider
type ModelInfo struct {
	Name      string
	Provider  string
	Score     int // Higher score = higher priority
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

// Model rankings based on performance and capabilities
var modelRankings = []ModelInfo{
	{Name: "gemini-2.5-pro", Provider: "gemini", Score: 100},
	{Name: "gpt-4o", Provider: "openai", Score: 95},
	{Name: "gpt-4", Provider: "openai", Score: 90},
	{Name: "claude-3-opus", Provider: "openrouter", Score: 85},
	{Name: "claude-3-sonnet", Provider: "openrouter", Score: 80},
	{Name: "gemini-1.5-pro", Provider: "gemini", Score: 75},
	{Name: "openrouter/anthropic/claude-3-opus", Provider: "openrouter", Score: 70},
}

// Provider environment variable mapping
var providerEnvVars = map[string]string{
	"openai":     "OPENAI_API_KEY",
	"gemini":     "GEMINI_API_KEY",
	"openrouter": "OPENROUTER_API_KEY",
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

	// Check all providers
	for providerName, envVar := range providerEnvVars {
		apiKey := strings.TrimSpace(os.Getenv(envVar))
		pc.cache[providerName] = apiKey != "" && len(apiKey) > 10 // Basic validation
	}

	available, exists := pc.cache[provider]
	return exists && available
}

// GetAvailableProviders returns list of providers with valid API keys
func GetAvailableProviders() []string {
	var available []string
	for provider := range providerEnvVars {
		if providerCache.isProviderAvailable(provider) {
			available = append(available, provider)
		}
	}
	sort.Strings(available) // Ensure deterministic ordering
	return available
}

// SelectOptimalModel selects the best available model based on provider availability and task complexity
func SelectOptimalModel(availableProviders []string, taskSize int64) string {
	if len(availableProviders) == 0 {
		// Fallback to default even without API key for testing/dry-run scenarios
		return DefaultModel
	}

	// Create set of available providers for O(1) lookup
	availableSet := make(map[string]bool)
	for _, provider := range availableProviders {
		availableSet[provider] = true
	}

	// Find highest-scoring available model
	for _, model := range modelRankings {
		if availableSet[model.Provider] {
			// For large tasks (>100k tokens), prefer more capable models
			if taskSize > 100000 && model.Score >= 90 {
				return model.Name
			}
			// For medium tasks (10k-100k tokens), prefer balanced models
			if taskSize > 10000 && model.Score >= 80 {
				return model.Name
			}
			// For small tasks, any available model is fine
			if model.Score >= 70 {
				return model.Name
			}
		}
	}

	// Fallback to default if no ranked models available
	return DefaultModel
}

// SelectBestModel is the main entry point for model selection
// It combines provider detection with optimal model selection
func SelectBestModel(taskSize int64) string {
	availableProviders := GetAvailableProviders()
	return SelectOptimalModel(availableProviders, taskSize)
}

// ValidateModelAvailability checks if a specific model is available
func ValidateModelAvailability(modelName string) bool {
	// Find the model in our rankings
	for _, model := range modelRankings {
		if model.Name == modelName {
			return providerCache.isProviderAvailable(model.Provider)
		}
	}

	// For unknown models, check if any provider is available
	// This allows for custom models not in our rankings
	availableProviders := GetAvailableProviders()
	return len(availableProviders) > 0
}

// GetModelProvider returns the provider for a given model name
func GetModelProvider(modelName string) string {
	for _, model := range modelRankings {
		if model.Name == modelName {
			return model.Provider
		}
	}

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

// ClearProviderCache forces a refresh of the provider availability cache
// Useful for testing or when environment changes
func ClearProviderCache() {
	providerCache.mu.Lock()
	defer providerCache.mu.Unlock()
	providerCache.cache = make(map[string]bool)
	providerCache.lastUpdate = time.Time{}
}

// GetModelRankings returns the model rankings for testing
func GetModelRankings() []ModelInfo {
	return modelRankings
}

// GetProviderEnvVars returns the provider environment variable mapping for testing
func GetProviderEnvVars() map[string]string {
	return providerEnvVars
}
