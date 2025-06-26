package cli

import (
	"os"
	"sync"
)

// envCache provides a thread-safe cache for environment variable lookups
// This optimizes startup performance by avoiding repeated os.Getenv() calls
type envCache struct {
	cache sync.Map // string (key) -> string (value)
	once  sync.Once
}

// Global instance for environment variable caching
var globalEnvCache = &envCache{}

// GetEnv returns the environment variable value, using cache for repeated lookups
// This provides significant startup performance improvement as environment variables
// are immutable during process lifetime and frequently accessed during CLI startup
func GetEnv(key string) string {
	// Try cache first (fastest path)
	if value, exists := globalEnvCache.cache.Load(key); exists {
		return value.(string)
	}

	// Cache miss: get from environment and store in cache
	value := os.Getenv(key)
	globalEnvCache.cache.Store(key, value)

	return value
}

// InvalidateEnvCache invalidates a specific key in the environment cache
// This is useful when environment variables are changed during testing
func InvalidateEnvCache(key string) {
	globalEnvCache.cache.Delete(key)
}

// PrewarmEnvCache loads common environment variables into cache during startup
// This is called once during CLI initialization to populate the cache with
// the most frequently accessed environment variables
func PrewarmEnvCache() {
	globalEnvCache.once.Do(func() {
		// Pre-cache common API key environment variables
		commonEnvVars := []string{
			"OPENAI_API_KEY",
			"GEMINI_API_KEY",
			"OPENROUTER_API_KEY",
			"GEMINI_API_URL",
			// Add other commonly accessed environment variables
			"THINKTANK_LOG_LEVEL",
			"THINKTANK_CONFIG_DIR",
			"HOME",
			"USER",
		}

		for _, key := range commonEnvVars {
			value := os.Getenv(key)
			globalEnvCache.cache.Store(key, value)
		}
	})
}

// ClearEnvCache clears the environment variable cache
// This is primarily used for testing to ensure clean state
func ClearEnvCache() {
	globalEnvCache.cache.Range(func(key, value interface{}) bool {
		globalEnvCache.cache.Delete(key)
		return true
	})
	// Reset the once so PrewarmEnvCache can be called again in tests
	globalEnvCache.once = sync.Once{}
}

// GetEnvWithDefault returns the environment variable value or a default if not set
// Uses the same caching mechanism as GetEnv for consistent performance
func GetEnvWithDefault(key, defaultValue string) string {
	value := GetEnv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
