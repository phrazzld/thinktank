// Package orchestrator contains tests for the orchestrator package.
// Tests have been refactored into separate files to improve organization:
// - orchestrator_init_test.go: Tests for orchestrator initialization
// - orchestrator_run_test.go: Tests for basic run workflows
// - orchestrator_error_test.go: Tests for error handling
// - orchestrator_integration_test.go: Integration tests
// - orchestrator_helpers_test.go: Shared test helpers and mock implementations
package orchestrator

import (
	"sync"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/testutil"
)

// TestGetRateLimiterForModel tests the getRateLimiterForModel function comprehensively
func TestGetRateLimiterForModel(t *testing.T) {
	// Setup test orchestrator
	logger := testutil.NewMockLogger()

	// Create a test config with provider rate limits
	cfg := &config.CliConfig{
		OpenRouterRateLimit: 100,
	}

	// Create global rate limiter
	globalRateLimiter := ratelimit.NewRateLimiter(10, 60)

	// Create orchestrator with basic dependencies
	orchestrator := &Orchestrator{
		logger:            logger,
		config:            cfg,
		rateLimiter:       globalRateLimiter,
		modelRateLimiters: make(map[string]*ratelimit.RateLimiter),
		rateLimiterMutex:  sync.RWMutex{},
	}

	tests := []struct {
		name           string
		modelName      string
		expectedGlobal bool // If true, expects global rate limiter
	}{
		{
			name:           "Unknown model uses global rate limiter",
			modelName:      "unknown-model",
			expectedGlobal: true,
		},
		{
			name:           "Known model with no MaxConcurrentRequests uses global rate limiter",
			modelName:      "gpt-4.1", // Use actual model from the registry
			expectedGlobal: true,
		},
		{
			name:           "Known model uses appropriate rate limiter",
			modelName:      "gemini-2.5-pro", // Use actual model from the registry
			expectedGlobal: true,             // Will be true unless the model has MaxConcurrentRequests set
		},
		{
			name:           "Model with MaxConcurrentRequests creates model-specific rate limiter",
			modelName:      "openrouter/deepseek/deepseek-r1-0528", // Model with MaxConcurrentRequests: &[]int{1}[0]
			expectedGlobal: false,                                  // Should create model-specific rate limiter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call getRateLimiterForModel
			rateLimiter := orchestrator.getRateLimiterForModel(tt.modelName)

			// Verify expectations
			if tt.expectedGlobal {
				if rateLimiter != globalRateLimiter {
					t.Errorf("Expected global rate limiter, got different limiter")
				}
			} else {
				if rateLimiter == globalRateLimiter {
					t.Errorf("Expected model-specific rate limiter, got global limiter")
				}
				// Verify the limiter was stored in the map
				if stored, exists := orchestrator.modelRateLimiters[tt.modelName]; !exists {
					t.Errorf("Model-specific rate limiter not stored in map")
				} else if stored != rateLimiter {
					t.Errorf("Stored rate limiter doesn't match returned limiter")
				}
			}
		})
	}
}

// TestGetRateLimiterForModelConcurrency tests concurrent access to getRateLimiterForModel
func TestGetRateLimiterForModelConcurrency(t *testing.T) {
	// Setup test orchestrator
	logger := testutil.NewMockLogger()
	cfg := &config.CliConfig{}
	globalRateLimiter := ratelimit.NewRateLimiter(10, 60)

	orchestrator := &Orchestrator{
		logger:            logger,
		config:            cfg,
		rateLimiter:       globalRateLimiter,
		modelRateLimiters: make(map[string]*ratelimit.RateLimiter),
		rateLimiterMutex:  sync.RWMutex{},
	}

	// Use a real model from the registry that has MaxConcurrentRequests set
	testModel := "openrouter/deepseek/deepseek-r1-0528"

	// Test concurrent access
	const numGoroutines = 10
	limiters := make([]*ratelimit.RateLimiter, numGoroutines)
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			limiters[index] = orchestrator.getRateLimiterForModel(testModel)
		}(i)
	}

	wg.Wait()

	// Verify all goroutines got the same rate limiter instance
	firstLimiter := limiters[0]
	for i := 1; i < numGoroutines; i++ {
		if limiters[i] != firstLimiter {
			t.Errorf("Concurrent access returned different rate limiter instances")
		}
	}

	// Since most real models don't have MaxConcurrentRequests, they'll all use the global rate limiter
	// So we just verify that concurrent access is safe
	if firstLimiter != globalRateLimiter {
		// If this model has MaxConcurrentRequests, verify only one was created
		if len(orchestrator.modelRateLimiters) != 1 {
			t.Errorf("Expected 1 rate limiter in map, got %d", len(orchestrator.modelRateLimiters))
		}
	}
}

// TestGetRateLimiterForModelCaching tests that rate limiters are properly cached
func TestGetRateLimiterForModelCaching(t *testing.T) {
	// Setup test orchestrator
	logger := testutil.NewMockLogger()
	cfg := &config.CliConfig{}
	globalRateLimiter := ratelimit.NewRateLimiter(10, 60)

	orchestrator := &Orchestrator{
		logger:            logger,
		config:            cfg,
		rateLimiter:       globalRateLimiter,
		modelRateLimiters: make(map[string]*ratelimit.RateLimiter),
		rateLimiterMutex:  sync.RWMutex{},
	}

	// Use a real model from the registry that has MaxConcurrentRequests set
	testModel := "openrouter/deepseek/deepseek-r1-0528"

	// Get rate limiter first time
	limiter1 := orchestrator.getRateLimiterForModel(testModel)

	// Get rate limiter second time
	limiter2 := orchestrator.getRateLimiterForModel(testModel)

	// Verify they are the same instance (cached)
	if limiter1 != limiter2 {
		t.Errorf("Expected same rate limiter instance from cache, got different instances")
	}

	// For models without MaxConcurrentRequests, they should both be the global rate limiter
	// For models with MaxConcurrentRequests, verify only one was created and stored
	if limiter1 != globalRateLimiter {
		// This model has MaxConcurrentRequests, verify only one was created
		if len(orchestrator.modelRateLimiters) != 1 {
			t.Errorf("Expected 1 rate limiter in map, got %d", len(orchestrator.modelRateLimiters))
		}
	}
}

// This file intentionally left mostly empty after refactoring tests into multiple files
// for improved organization and maintainability.
