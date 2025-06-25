package cli

import (
	"os"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/testutil"
)

func TestNewAPIKeyDetector(t *testing.T) {
	logger := &testutil.MockLogger{}
	detector := NewAPIKeyDetector(logger)

	if detector == nil {
		t.Fatal("NewAPIKeyDetector returned nil")
	}

	if detector.logger != logger {
		t.Error("Logger not properly assigned")
	}

	if detector.cache == nil {
		t.Error("Cache not initialized")
	}

	expectedTTL := 5 * time.Minute
	if detector.cache.ttl != expectedTTL {
		t.Errorf("Expected TTL %v, got %v", expectedTTL, detector.cache.ttl)
	}
}

func TestDetectAndValidate_EmptyEnvironment(t *testing.T) {
	// Clean environment
	cleanupEnv := setupCleanEnvironment(t)
	defer cleanupEnv()

	logger := &testutil.MockLogger{}
	detector := NewAPIKeyDetector(logger)

	result, err := detector.DetectAndValidate()
	if err != nil {
		t.Fatalf("DetectAndValidate failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// Check no providers detected
	for provider, detected := range result.DetectedProviders {
		if detected {
			t.Errorf("Provider %s should not be detected in clean environment", provider)
		}
	}

	// Check no valid keys
	for provider, valid := range result.ValidKeys {
		if valid {
			t.Errorf("Provider %s should not have valid key in clean environment", provider)
		}
	}

	// Check capabilities are populated even without keys
	expectedProviders := []string{"openai", "gemini", "openrouter"}
	for _, provider := range expectedProviders {
		capability, exists := result.ProviderCapabilities[provider]
		if !exists {
			t.Errorf("Capability missing for provider %s", provider)
			continue
		}

		if capability.Provider != provider {
			t.Errorf("Wrong provider name in capability: expected %s, got %s", provider, capability.Provider)
		}

		if capability.HasValidKey {
			t.Errorf("Provider %s should not have valid key", provider)
		}

		if capability.DefaultRateLimit <= 0 {
			t.Errorf("Provider %s should have positive rate limit, got %d", provider, capability.DefaultRateLimit)
		}

		if capability.ModelCount != len(capability.AvailableModels) {
			t.Errorf("Model count mismatch for %s: count=%d, models=%d", provider, capability.ModelCount, len(capability.AvailableModels))
		}
	}
}

func TestDetectAndValidate_ValidKeys(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		envVar   string
		keyValue string
		valid    bool
	}{
		{
			name:     "valid_openai_key",
			provider: "openai",
			envVar:   "OPENAI_API_KEY",
			keyValue: "sk-1234567890abcdef1234567890abcdef",
			valid:    true,
		},
		{
			name:     "valid_openai_project_key",
			provider: "openai",
			envVar:   "OPENAI_API_KEY",
			keyValue: "sk-proj-1234567890abcdef1234567890abcdef",
			valid:    true,
		},
		{
			name:     "invalid_openai_key_too_short",
			provider: "openai",
			envVar:   "OPENAI_API_KEY",
			keyValue: "sk-short",
			valid:    false,
		},
		{
			name:     "invalid_openai_key_wrong_prefix",
			provider: "openai",
			envVar:   "OPENAI_API_KEY",
			keyValue: "ak-1234567890abcdef1234567890abcdef",
			valid:    false,
		},
		{
			name:     "valid_gemini_key",
			provider: "gemini",
			envVar:   "GEMINI_API_KEY",
			keyValue: "AIzaSyDdI0hCZtE6vySjMm-WEfRq3CPzqKqqsHI", // 39 chars
			valid:    true,
		},
		{
			name:     "invalid_gemini_key_wrong_length",
			provider: "gemini",
			envVar:   "GEMINI_API_KEY",
			keyValue: "AIzaSyDdI0hCZtE6vySjMm-WEfRq3CPzqKqqsHIX", // 40 chars
			valid:    false,
		},
		{
			name:     "valid_openrouter_key",
			provider: "openrouter",
			envVar:   "OPENROUTER_API_KEY",
			keyValue: "sk-or-v1-1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			valid:    true,
		},
		{
			name:     "invalid_openrouter_key_wrong_prefix",
			provider: "openrouter",
			envVar:   "OPENROUTER_API_KEY",
			keyValue: "sk-1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment and set specific key
			cleanupEnv := setupCleanEnvironment(t)
			defer cleanupEnv()

			_ = os.Setenv(tt.envVar, tt.keyValue)

			logger := &testutil.MockLogger{}
			detector := NewAPIKeyDetector(logger)

			result, err := detector.DetectAndValidate()
			if err != nil {
				t.Fatalf("DetectAndValidate failed: %v", err)
			}

			// Check detection
			detected := result.DetectedProviders[tt.provider]
			if !detected {
				t.Errorf("Provider %s should be detected", tt.provider)
			}

			// Check validation
			valid := result.ValidKeys[tt.provider]
			if valid != tt.valid {
				t.Errorf("Provider %s validation: expected %v, got %v", tt.provider, tt.valid, valid)
			}

			// Check capability
			capability := result.ProviderCapabilities[tt.provider]
			if capability.HasValidKey != tt.valid {
				t.Errorf("Capability HasValidKey: expected %v, got %v", tt.valid, capability.HasValidKey)
			}
		})
	}
}

func TestDetectAndValidate_Caching(t *testing.T) {
	cleanupEnv := setupCleanEnvironment(t)
	defer cleanupEnv()

	logger := &testutil.MockLogger{}
	detector := NewAPIKeyDetector(logger)

	// First call - should perform detection
	start := time.Now()
	result1, err := detector.DetectAndValidate()
	duration1 := time.Since(start)
	if err != nil {
		t.Fatalf("First DetectAndValidate failed: %v", err)
	}

	// Second call - should use cache
	start = time.Now()
	result2, err := detector.DetectAndValidate()
	duration2 := time.Since(start)
	if err != nil {
		t.Fatalf("Second DetectAndValidate failed: %v", err)
	}

	// Results should be identical
	if result1.DetectionTimestamp != result2.DetectionTimestamp {
		t.Error("Cache not working - timestamps differ")
	}

	// Second call should be faster (cache hit)
	if duration2 >= duration1 {
		t.Logf("Cache performance: first=%v, second=%v", duration1, duration2)
		// Note: This might fail in very fast environments, but it's a good indicator
	}

	// Check that cache debug message was logged
	if !logger.ContainsMessage("Using cached API key detection result") {
		t.Error("Cache debug message not found in logs")
	}
}

func TestDetectAndValidate_CacheExpiry(t *testing.T) {
	cleanupEnv := setupCleanEnvironment(t)
	defer cleanupEnv()

	logger := &testutil.MockLogger{}
	detector := NewAPIKeyDetector(logger)

	// Override TTL for testing
	detector.cache.ttl = 10 * time.Millisecond

	// First call
	result1, err := detector.DetectAndValidate()
	if err != nil {
		t.Fatalf("First DetectAndValidate failed: %v", err)
	}

	// Wait for cache to expire
	time.Sleep(20 * time.Millisecond)

	// Second call should perform fresh detection
	result2, err := detector.DetectAndValidate()
	if err != nil {
		t.Fatalf("Second DetectAndValidate failed: %v", err)
	}

	// Timestamps should be different (fresh detection)
	if result1.DetectionTimestamp.Equal(result2.DetectionTimestamp) {
		t.Error("Cache expiry not working - timestamps identical")
	}
}

func TestAPIDetector_GetAvailableProviders(t *testing.T) {
	cleanupEnv := setupCleanEnvironment(t)
	defer cleanupEnv()

	// Set valid keys for two providers
	_ = os.Setenv("OPENAI_API_KEY", "sk-1234567890abcdef1234567890abcdef")
	_ = os.Setenv("GEMINI_API_KEY", "AIzaSyDdI0hCZtE6vySjMm-WEfRq3CPzqKqqsHI")

	logger := &testutil.MockLogger{}
	detector := NewAPIKeyDetector(logger)

	providers, err := detector.GetAvailableProviders()
	if err != nil {
		t.Fatalf("GetAvailableProviders failed: %v", err)
	}

	expectedCount := 2
	if len(providers) != expectedCount {
		t.Errorf("Expected %d providers, got %d: %v", expectedCount, len(providers), providers)
	}

	// Check specific providers
	hasOpenAI := false
	hasGemini := false
	for _, provider := range providers {
		switch provider {
		case "openai":
			hasOpenAI = true
		case "gemini":
			hasGemini = true
		}
	}

	if !hasOpenAI {
		t.Error("OpenAI should be in available providers")
	}
	if !hasGemini {
		t.Error("Gemini should be in available providers")
	}
}

func TestGetProviderCapability(t *testing.T) {
	cleanupEnv := setupCleanEnvironment(t)
	defer cleanupEnv()

	_ = os.Setenv("OPENAI_API_KEY", "sk-1234567890abcdef1234567890abcdef")

	logger := &testutil.MockLogger{}
	detector := NewAPIKeyDetector(logger)

	// Test existing provider
	capability, err := detector.GetProviderCapability("openai")
	if err != nil {
		t.Fatalf("GetProviderCapability failed: %v", err)
	}

	if capability == nil {
		t.Fatal("Capability is nil for openai")
	}

	if capability.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", capability.Provider)
	}

	if !capability.HasValidKey {
		t.Error("OpenAI should have valid key")
	}

	if capability.DefaultRateLimit <= 0 {
		t.Errorf("Rate limit should be positive, got %d", capability.DefaultRateLimit)
	}

	if capability.ModelCount == 0 {
		t.Error("Model count should be > 0")
	}

	// Test non-existent provider
	capability, err = detector.GetProviderCapability("nonexistent")
	if err != nil {
		t.Fatalf("GetProviderCapability failed for nonexistent: %v", err)
	}

	if capability != nil {
		t.Error("Capability should be nil for nonexistent provider")
	}
}

func TestInvalidateCache(t *testing.T) {
	cleanupEnv := setupCleanEnvironment(t)
	defer cleanupEnv()

	logger := &testutil.MockLogger{}
	detector := NewAPIKeyDetector(logger)

	// First call to populate cache
	_, err := detector.DetectAndValidate()
	if err != nil {
		t.Fatalf("DetectAndValidate failed: %v", err)
	}

	// Verify cache is populated
	detector.cacheMutex.RLock()
	cachePopulated := detector.cache.result != nil
	detector.cacheMutex.RUnlock()

	if !cachePopulated {
		t.Error("Cache should be populated after first call")
	}

	// Invalidate cache
	detector.InvalidateCache()

	// Verify cache is cleared
	detector.cacheMutex.RLock()
	cacheCleared := detector.cache.result == nil
	detector.cacheMutex.RUnlock()

	if !cacheCleared {
		t.Error("Cache should be cleared after invalidation")
	}

	// Check debug message
	if !logger.ContainsMessage("API key detection cache invalidated") {
		t.Error("Cache invalidation debug message not found")
	}
}

func TestDetectionPerformance(t *testing.T) {
	cleanupEnv := setupCleanEnvironment(t)
	defer cleanupEnv()

	// Set all provider keys
	_ = os.Setenv("OPENAI_API_KEY", "sk-1234567890abcdef1234567890abcdef")
	_ = os.Setenv("GEMINI_API_KEY", "AIzaSyDdI0hCZtE6vySjMm-WEfRq3CPzqKqqsHI")
	_ = os.Setenv("OPENROUTER_API_KEY", "sk-or-v1-1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	logger := &testutil.MockLogger{}
	detector := NewAPIKeyDetector(logger)

	// Measure performance
	start := time.Now()
	result, err := detector.DetectAndValidate()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("DetectAndValidate failed: %v", err)
	}

	// Performance assertions
	maxDuration := 10 * time.Millisecond // Should be very fast
	if duration > maxDuration {
		t.Errorf("Detection took too long: %v (max %v)", duration, maxDuration)
	}

	// Verify all providers detected and valid
	for _, provider := range []string{"openai", "gemini", "openrouter"} {
		if !result.DetectedProviders[provider] {
			t.Errorf("Provider %s not detected", provider)
		}
		if !result.ValidKeys[provider] {
			t.Errorf("Provider %s key not valid", provider)
		}
	}

	t.Logf("Detection completed in %v", duration)
}

// setupCleanEnvironment clears all API key environment variables
func setupCleanEnvironment(t *testing.T) func() {
	originalVars := make(map[string]string)
	envVars := []string{"OPENAI_API_KEY", "GEMINI_API_KEY", "OPENROUTER_API_KEY"}

	for _, envVar := range envVars {
		originalVars[envVar] = os.Getenv(envVar)
		_ = os.Unsetenv(envVar)
	}

	return func() {
		for envVar, originalValue := range originalVars {
			if originalValue != "" {
				_ = os.Setenv(envVar, originalValue)
			} else {
				_ = os.Unsetenv(envVar)
			}
		}
	}
}
