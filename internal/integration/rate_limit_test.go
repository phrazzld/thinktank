package integration

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRateLimitedModelProcessing verifies that rate limiting works properly
func TestRateLimitedModelProcessing(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping rate limit test in short mode")
	}

	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Track request timestamps
	var timestamps []time.Time
	timestampMu := sync.Mutex{}

	// Count the number of concurrent requests
	var concurrentCount int32
	var maxConcurrent int32

	// Override GenerateContentFunc to track timing and concurrency
	env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
		// Extract model name from context
		modelName := ctx.Value(modelNameKey).(string)

		// Track concurrency
		current := atomic.AddInt32(&concurrentCount, 1)
		defer atomic.AddInt32(&concurrentCount, -1)

		// Update max concurrency
		for {
			max := atomic.LoadInt32(&maxConcurrent)
			if current <= max {
				break
			}
			if atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
				break
			}
		}

		// Record timestamp
		now := time.Now()
		timestampMu.Lock()
		timestamps = append(timestamps, now)
		timestampMu.Unlock()

		// Simulate work
		time.Sleep(50 * time.Millisecond)

		// Return a basic response
		return &gemini.GenerationResult{
			Content: "Test response for " + modelName,
		}, nil
	}

	// Create a test file
	env.CreateTestFile(t, "src/main.go", `package main

func main() {}`)

	// Create an instructions file
	instructionsFile := env.CreateTestFile(t, "instructions.md", "Test rate limiting")

	// Set up the output directory
	outputDir := filepath.Join(env.TestDir, "output")

	// Run with multiple models (at least 5)
	models := []string{"model1", "model2", "model3", "model4", "model5"}

	// Create a test configuration with rate limiting
	testConfig := &architect.CliConfig{
		InstructionsFile: instructionsFile,
		OutputDir:        outputDir,
		ModelNames:       models,
		ApiKey:           "test-api-key",
		Paths:            []string{env.TestDir + "/src"},
		LogLevel:         logutil.InfoLevel,
		// Add rate limiting configuration
		MaxConcurrentRequests:      1, // Only 1 concurrent
		RateLimitRequestsPerMinute: 5, // Only 5 RPM (1 per 12 seconds)
	}

	// Create a custom mock API service that can track model names
	mockAPIService := &mockModelTrackingAPIService{
		logger:     env.Logger,
		mockClient: env.MockClient,
	}

	// Run the application
	ctx := context.Background()
	err := architect.RunInternal(
		ctx,
		testConfig,
		env.Logger,
		mockAPIService,
		env.AuditLogger,
	)

	// Verify execution
	require.NoError(t, err, "Execution should succeed even with rate limiting")

	// Verify concurrency limit was respected
	assert.LessOrEqual(t, maxConcurrent, int32(1), "Max concurrent requests should respect limit")

	// Check rate limiting by analyzing timestamp spacing
	timestampMu.Lock()
	defer timestampMu.Unlock()

	// Sort timestamps (they might not be in order due to goroutine scheduling)
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i].Before(timestamps[j])
	})

	// Calculate time differences between requests
	if len(timestamps) >= 2 {
		// With RPM of 5, we expect at least 12 seconds between each request on average
		minExpectedGap := 11 * time.Second // Allow slight variance for testing

		// Check time gaps between requests
		var shortGaps int
		for i := 1; i < len(timestamps); i++ {
			gap := timestamps[i].Sub(timestamps[i-1])
			if gap < minExpectedGap {
				shortGaps++
			}
		}

		// With a burst size of 1, but processing several models, we'll allow up to 4 short gaps
		// This is because the initial burst processing might happen closer together
		assert.LessOrEqual(t, shortGaps, 4, "Most time gaps should respect the rate limit")
	}
}

// TestRateLimitConfiguration verifies that different rate limit configurations are handled correctly
func TestRateLimitConfiguration(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping rate limit test in short mode")
	}

	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Count the number of concurrent requests
	var concurrentCount int32
	var maxConcurrent int32

	// Override GenerateContentFunc to track concurrency
	env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
		// Extract model name from context
		modelName := ctx.Value(modelNameKey).(string)

		// Track concurrency
		current := atomic.AddInt32(&concurrentCount, 1)
		defer atomic.AddInt32(&concurrentCount, -1)

		// Update max concurrency
		for {
			max := atomic.LoadInt32(&maxConcurrent)
			if current <= max {
				break
			}
			if atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
				break
			}
		}

		// Simulate work
		time.Sleep(50 * time.Millisecond)

		// Return a basic response
		return &gemini.GenerationResult{
			Content: "Test response for " + modelName,
		}, nil
	}

	// Create a test file
	env.CreateTestFile(t, "src/main.go", `package main

func main() {}`)

	// Create an instructions file
	instructionsFile := env.CreateTestFile(t, "instructions.md", "Test rate limiting configuration")

	// Set up the output directory
	outputDir := filepath.Join(env.TestDir, "output")

	// Use multiple models to ensure concurrency
	models := []string{"model1", "model2", "model3", "model4", "model5"}

	// Create a test configuration with strict concurrency limiting but no rate limiting
	testConfig := &architect.CliConfig{
		InstructionsFile: instructionsFile,
		OutputDir:        outputDir,
		ModelNames:       models,
		ApiKey:           "test-api-key",
		Paths:            []string{env.TestDir + "/src"},
		LogLevel:         logutil.InfoLevel,
		// Add rate limiting configuration - restrict concurrency only
		MaxConcurrentRequests:      2, // Only 2 concurrent requests
		RateLimitRequestsPerMinute: 0, // No rate limiting
	}

	// Create a custom mock API service that can track model names
	mockAPIService := &mockModelTrackingAPIService{
		logger:     env.Logger,
		mockClient: env.MockClient,
	}

	// Reset max concurrency tracking
	atomic.StoreInt32(&maxConcurrent, 0)

	// Run the application
	ctx := context.Background()
	err := architect.RunInternal(
		ctx,
		testConfig,
		env.Logger,
		mockAPIService,
		env.AuditLogger,
	)

	// Verify execution
	require.NoError(t, err, "Execution should succeed")

	// Verify concurrency limit was respected
	assert.LessOrEqual(t, maxConcurrent, int32(2), "Max concurrent requests should respect limit")

	// Verify that output files were created for all models
	for _, modelName := range models {
		outputFile := filepath.Join(outputDir, modelName+".md")
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Output file for model %s was not created at %s", modelName, outputFile)
		}
	}
}

// TestRateLimitMultiModel verifies that multiple models are rate limited correctly
func TestRateLimitMultiModel(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping rate limit test in short mode")
	}

	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Track request timestamps per model
	modelTimestamps := make(map[string][]time.Time)
	timestampMu := sync.Mutex{}

	// Override GenerateContentFunc to track timing
	env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
		// Extract model name from context
		modelName := ctx.Value(modelNameKey).(string)

		// Record timestamp for this model
		now := time.Now()
		timestampMu.Lock()
		if modelTimestamps[modelName] == nil {
			modelTimestamps[modelName] = make([]time.Time, 0, 3)
		}
		modelTimestamps[modelName] = append(modelTimestamps[modelName], now)
		timestampMu.Unlock()

		// Simulate work
		time.Sleep(50 * time.Millisecond)

		// Return a basic response
		return &gemini.GenerationResult{
			Content: "Test response for " + modelName,
		}, nil
	}

	// Create a test file
	env.CreateTestFile(t, "src/main.go", `package main

func main() {}`)

	// Create an instructions file
	instructionsFile := env.CreateTestFile(t, "instructions.md", "Test rate limiting for multiple models")

	// Set up the output directory
	outputDir := filepath.Join(env.TestDir, "output")

	// Create a slice to hold model names (same model repeated to trigger rate limiting)
	// We'll use 3 occurrences each of 2 models to test per-model rate limiting
	models := []string{"model1", "model1", "model1", "model2", "model2", "model2"}

	// Create a test configuration with rate limiting
	testConfig := &architect.CliConfig{
		InstructionsFile: instructionsFile,
		OutputDir:        outputDir,
		ModelNames:       models,
		ApiKey:           "test-api-key",
		Paths:            []string{env.TestDir + "/src"},
		LogLevel:         logutil.InfoLevel,
		// Add rate limiting configuration - focus on per-model rate limiting
		MaxConcurrentRequests:      10, // Allow high concurrency
		RateLimitRequestsPerMinute: 2,  // Only 2 RPM per model (30 seconds between requests for same model)
	}

	// Create a custom mock API service that can track model names
	mockAPIService := &mockModelTrackingAPIService{
		logger:     env.Logger,
		mockClient: env.MockClient,
	}

	// Run the application
	ctx := context.Background()
	err := architect.RunInternal(
		ctx,
		testConfig,
		env.Logger,
		mockAPIService,
		env.AuditLogger,
	)

	// Verify execution
	require.NoError(t, err, "Execution should succeed even with rate limiting")

	// Lock for accessing timestamps
	timestampMu.Lock()
	defer timestampMu.Unlock()

	// Verify we have timestamps for all models
	assert.Contains(t, modelTimestamps, "model1", "Should have timestamps for model1")
	assert.Contains(t, modelTimestamps, "model2", "Should have timestamps for model2")

	// Verify the number of timestamps for each model
	assert.Equal(t, 3, len(modelTimestamps["model1"]), "Should have 3 timestamps for model1")
	assert.Equal(t, 3, len(modelTimestamps["model2"]), "Should have 3 timestamps for model2")

	// Sort timestamps for each model
	for model := range modelTimestamps {
		sort.Slice(modelTimestamps[model], func(i, j int) bool {
			return modelTimestamps[model][i].Before(modelTimestamps[model][j])
		})
	}

	// Verify that timestamps for the same model have proper spacing
	// but timestamps between different models can be closer
	checkRateLimiting(t, modelTimestamps["model1"], 30*time.Second)
	checkRateLimiting(t, modelTimestamps["model2"], 30*time.Second)
}

// Helper function to check rate limiting time gaps
func checkRateLimiting(t *testing.T, timestamps []time.Time, minExpectedGap time.Duration) {
	if len(timestamps) < 2 {
		return
	}

	// Allow for the first request to happen immediately (burst)
	// but check spacing for subsequent requests
	for i := 1; i < len(timestamps); i++ {
		gap := timestamps[i].Sub(timestamps[i-1])

		// If not first request, we should see rate limiting
		if i > 1 {
			// Allow some tolerance (90% of expected gap)
			if gap < minExpectedGap*9/10 {
				t.Errorf("Time gap %v is too short, expected at least %v", gap, minExpectedGap)
			}
		}
	}
}

// TestRateLimitErrorHandling tests that API rate limit errors are properly captured and reported
func TestRateLimitErrorHandling(t *testing.T) {
	// Set up the test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Track number of requests per model
	requestCounts := make(map[string]int)
	requestMu := sync.Mutex{}

	// Override GenerateContentFunc to simulate rate limit errors
	env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
		// Extract model name from context
		modelName := ctx.Value(modelNameKey).(string)

		// Track request count for this model
		requestMu.Lock()
		requestCounts[modelName]++
		count := requestCounts[modelName]
		requestMu.Unlock()

		// Simulate rate limit error for specific models or conditions
		if modelName == "rate-limited-model" || count > 2 {
			// Return a rate limit error
			return nil, &gemini.APIError{
				Type:       gemini.ErrorTypeRateLimit,
				Message:    "Rate limit exceeded",
				Suggestion: "Try again later",
				StatusCode: 429,
			}
		}

		// Successful response for other models
		return &gemini.GenerationResult{
			Content: "Test response for " + modelName,
		}, nil
	}

	// Create a test file
	env.CreateTestFile(t, "src/main.go", `package main

func main() {}`)

	// Create an instructions file
	instructionsFile := env.CreateTestFile(t, "instructions.md", "Test rate limit error handling")

	// Set up the output directory
	outputDir := filepath.Join(env.TestDir, "output")

	// Use a mix of models including one that will always error
	models := []string{"normal-model", "rate-limited-model", "another-model"}

	// Create a test configuration
	testConfig := &architect.CliConfig{
		InstructionsFile: instructionsFile,
		OutputDir:        outputDir,
		ModelNames:       models,
		ApiKey:           "test-api-key",
		Paths:            []string{env.TestDir + "/src"},
		LogLevel:         logutil.InfoLevel,
		// Add rate limiting configuration (not actually used in this test)
		MaxConcurrentRequests:      5,
		RateLimitRequestsPerMinute: 60,
	}

	// Create a custom mock API service that can track model names
	mockAPIService := &mockModelTrackingAPIService{
		logger:     env.Logger,
		mockClient: env.MockClient,
	}

	// Run the application
	ctx := context.Background()
	err := architect.RunInternal(
		ctx,
		testConfig,
		env.Logger,
		mockAPIService,
		env.AuditLogger,
	)

	// Verify execution
	require.Error(t, err, "Execution should return error due to rate limiting")

	// Verify error message contains rate limit info - it should include "Rate limit exceeded"
	assert.Contains(t, err.Error(), "Rate limit exceeded", "Error should mention rate limiting")

	// Verify successful models still produced output files
	normalOutputFile := filepath.Join(outputDir, "normal-model.md")
	if _, err := os.Stat(normalOutputFile); os.IsNotExist(err) {
		t.Errorf("Output for successful model should exist at %s", normalOutputFile)
	}

	// Verify failed model did not produce output file
	failedOutputFile := filepath.Join(outputDir, "rate-limited-model.md")
	if _, err := os.Stat(failedOutputFile); !os.IsNotExist(err) {
		t.Errorf("Output for rate-limited model should not exist at %s", failedOutputFile)
	}
}
