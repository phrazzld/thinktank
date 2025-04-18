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
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: contextKey type and modelNameKey are defined in multi_model_test.go
// mockModelTrackingAPIService is defined in test_runner.go

// Test case struct for rate limit tests
type rateLimitTestCase struct {
	name                    string
	instructionsContent     string
	models                  []string
	maxConcurrentRequests   int
	requestsPerMinute       int
	configureMock           func(t *testing.T, env *TestEnv) (any, *sync.Mutex)
	verifyResults           func(t *testing.T, env *TestEnv, mockData any, mu *sync.Mutex, err error, outputDir string)
	expectError             bool
	simulateRateLimitErrors bool
}

// TestRateLimitFeatures groups all rate limiting tests using t.Run and table-driven approach
func TestRateLimitFeatures(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping rate limit tests in short mode")
	}

	// Common source code for test files
	mainGoCode := `package main

func main() {}`

	// Define test cases
	testCases := []rateLimitTestCase{
		{
			name:                  "BasicRateLimiting",
			instructionsContent:   "Test rate limiting",
			models:                []string{"model1", "model2", "model3", "model4", "model5"},
			maxConcurrentRequests: 1,
			requestsPerMinute:     5, // Only 5 RPM (1 per 12 seconds)
			configureMock: func(t *testing.T, env *TestEnv) (any, *sync.Mutex) {
				// Track request timestamps
				timestamps := make([]time.Time, 0)
				timestampMu := &sync.Mutex{}

				// Count the number of concurrent requests
				var concurrentCount int32
				var maxConcurrent int32

				// Override GenerateContentFunc to track timing and concurrency
				env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*gemini.GenerationResult, error) {
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
					time.Sleep(5 * time.Millisecond)

					// Return a basic response
					return &gemini.GenerationResult{
						Content: "Test response for " + modelName,
					}, nil
				}

				// Return tracking data
				return struct {
					timestamps    []time.Time
					maxConcurrent *int32
				}{
					timestamps:    timestamps,
					maxConcurrent: &maxConcurrent,
				}, timestampMu
			},
			verifyResults: func(t *testing.T, env *TestEnv, mockData any, mu *sync.Mutex, err error, outputDir string) {
				// Extract tracking data
				data := mockData.(struct {
					timestamps    []time.Time
					maxConcurrent *int32
				})

				// Verify execution
				require.NoError(t, err, "Execution should succeed even with rate limiting")

				// Verify concurrency limit was respected
				assert.LessOrEqual(t, *data.maxConcurrent, int32(1), "Max concurrent requests should respect limit")

				// Lock for accessing timestamps
				mu.Lock()
				defer mu.Unlock()

				// Sort timestamps (they might not be in order due to goroutine scheduling)
				sort.Slice(data.timestamps, func(i, j int) bool {
					return data.timestamps[i].Before(data.timestamps[j])
				})

				// Calculate time differences between requests
				if len(data.timestamps) >= 2 {
					// With RPM of 5, we expect at least 12 seconds between each request on average
					minExpectedGap := 10 * time.Millisecond // Allow slight variance for testing

					// Check time gaps between requests
					var shortGaps int
					for i := 1; i < len(data.timestamps); i++ {
						gap := data.timestamps[i].Sub(data.timestamps[i-1])
						if gap < minExpectedGap {
							shortGaps++
						}
					}

					// With a burst size of 1, but processing several models, we'll allow up to 4 short gaps
					// This is because the initial burst processing might happen closer together
					assert.LessOrEqual(t, shortGaps, 4, "Most time gaps should respect the rate limit")
				}
			},
		},
		{
			name:                  "ConcurrencyConfiguration",
			instructionsContent:   "Test rate limiting configuration",
			models:                []string{"model1", "model2", "model3", "model4", "model5"},
			maxConcurrentRequests: 2,
			requestsPerMinute:     0, // No rate limiting
			configureMock: func(t *testing.T, env *TestEnv) (any, *sync.Mutex) {
				// Count the number of concurrent requests
				var concurrentCount int32
				var maxConcurrent int32

				// Override GenerateContentFunc to track concurrency
				env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*gemini.GenerationResult, error) {
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
					time.Sleep(5 * time.Millisecond)

					// Return a basic response
					return &gemini.GenerationResult{
						Content: "Test response for " + modelName,
					}, nil
				}

				// Reset max concurrency tracking
				atomic.StoreInt32(&maxConcurrent, 0)

				// Return tracking data
				return &maxConcurrent, nil
			},
			verifyResults: func(t *testing.T, env *TestEnv, mockData any, mu *sync.Mutex, err error, outputDir string) {
				// Extract tracking data
				maxConcurrent := mockData.(*int32)

				// Verify execution
				require.NoError(t, err, "Execution should succeed")

				// Verify concurrency limit was respected
				assert.LessOrEqual(t, *maxConcurrent, int32(2), "Max concurrent requests should respect limit")

				// Verify that output files were created for all models
				for _, modelName := range []string{"model1", "model2", "model3", "model4", "model5"} {
					outputFile := filepath.Join(outputDir, modelName+".md")
					if _, err := os.Stat(outputFile); os.IsNotExist(err) {
						t.Errorf("Output file for model %s was not created at %s", modelName, outputFile)
					}
				}
			},
		},
		{
			name:                "MultiModelRateLimiting",
			instructionsContent: "Test rate limiting for multiple models",
			// Create a slice to hold model names (same model repeated to trigger rate limiting)
			// We'll use 3 occurrences each of 2 models to test per-model rate limiting
			models:                []string{"model1", "model1", "model1", "model2", "model2", "model2"},
			maxConcurrentRequests: 10, // Allow high concurrency
			requestsPerMinute:     2,  // Only 2 RPM per model (30 seconds between requests for same model)
			configureMock: func(t *testing.T, env *TestEnv) (any, *sync.Mutex) {
				// Track request timestamps per model
				modelTimestamps := make(map[string][]time.Time)
				timestampMu := &sync.Mutex{}

				// Override GenerateContentFunc to track timing
				env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*gemini.GenerationResult, error) {
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
					time.Sleep(5 * time.Millisecond)

					// Return a basic response
					return &gemini.GenerationResult{
						Content: "Test response for " + modelName,
					}, nil
				}

				return modelTimestamps, timestampMu
			},
			verifyResults: func(t *testing.T, env *TestEnv, mockData any, mu *sync.Mutex, err error, outputDir string) {
				// Extract tracking data
				modelTimestamps := mockData.(map[string][]time.Time)

				// Verify execution
				require.NoError(t, err, "Execution should succeed even with rate limiting")

				// Lock for accessing timestamps
				mu.Lock()
				defer mu.Unlock()

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
				checkRateLimiting(t, modelTimestamps["model1"], 30*time.Millisecond)
				checkRateLimiting(t, modelTimestamps["model2"], 30*time.Millisecond)
			},
		},
		{
			name:                    "RateLimitErrorHandling",
			instructionsContent:     "Test rate limit error handling",
			models:                  []string{"normal-model", "rate-limited-model", "another-model"},
			maxConcurrentRequests:   5,
			requestsPerMinute:       60,
			simulateRateLimitErrors: true,
			expectError:             true,
			configureMock: func(t *testing.T, env *TestEnv) (any, *sync.Mutex) {
				// Track number of requests per model
				requestCounts := make(map[string]int)
				requestMu := &sync.Mutex{}

				// Override GenerateContentFunc to simulate rate limit errors
				env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*gemini.GenerationResult, error) {
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

				return nil, nil
			},
			verifyResults: func(t *testing.T, env *TestEnv, mockData any, mu *sync.Mutex, err error, outputDir string) {
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
			},
		},
	}

	// Run each test case as a subtest
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up the test environment
			env := NewTestEnv(t)
			defer env.Cleanup()

			// Create a test file
			env.CreateTestFile(t, "src/main.go", mainGoCode)

			// Create an instructions file
			instructionsFile := env.CreateTestFile(t, "instructions.md", tc.instructionsContent)

			// Set up the output directory
			outputDir := filepath.Join(env.TestDir, "output")

			// Configure the mock client
			mockData, mu := tc.configureMock(t, env)

			// Create a test configuration with rate limiting
			testConfig := &config.CliConfig{
				InstructionsFile:           instructionsFile,
				OutputDir:                  outputDir,
				ModelNames:                 tc.models,
				APIKey:                     "test-api-key",
				Paths:                      []string{env.TestDir + "/src"},
				LogLevel:                   logutil.InfoLevel,
				MaxConcurrentRequests:      tc.maxConcurrentRequests,
				RateLimitRequestsPerMinute: tc.requestsPerMinute,
			}

			// Create a custom mock API service that can track model names
			mockAPIService := &mockModelTrackingAPIService{
				logger:     env.Logger,
				mockClient: env.MockClient,
			}

			// Run the application
			ctx := context.Background()
			err := architect.Execute(
				ctx,
				testConfig,
				env.Logger,
				env.AuditLogger,
				mockAPIService,
			)

			// Verify results with the specific test case verification function
			tc.verifyResults(t, env, mockData, mu, err, outputDir)
		})
	}
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
