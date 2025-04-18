package integration

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/openai"
	"github.com/phrazzld/architect/internal/providers/openrouter"
	"github.com/phrazzld/architect/internal/ratelimit"
	"github.com/phrazzld/architect/internal/registry"
)

// Define a key type for context values
type contextKey string

const modelNameKey contextKey = "modelName"

// TestRateLimiting tests that rate limits are enforced across provider implementations
func TestRateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rate limit test in short mode")
	}

	// Setup test environment
	env := newTestEnvironment(t)
	defer env.Cleanup()

	// Create rate limiter
	rateLimiter := ratelimit.NewRateLimiter(2, 1*time.Second) // 2 requests per second

	// Define test cases
	testCases := []struct {
		name        string
		models      []string
		concurrency int
		duration    time.Duration
		expectRates map[string]float64 // expected max request rate per model (reqs/sec)
	}{
		{
			name:        "Single model with rate limit",
			models:      []string{"gpt-3.5-turbo"},
			concurrency: 5,
			duration:    3 * time.Second,
			expectRates: map[string]float64{
				"gpt-3.5-turbo": 2.5, // Should be around 2 but allow some margin
			},
		},
		{
			name:        "Multiple models with shared rate limit",
			models:      []string{"gpt-3.5-turbo", "gpt-4", "gemini-pro"},
			concurrency: 10,
			duration:    4 * time.Second,
			expectRates: map[string]float64{
				"gpt-3.5-turbo": 2.5,
				"gpt-4":         2.5,
				"gemini-pro":    2.5,
			},
		},
	}

	// Execute test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset environment for each test
			env.Reset()

			// Track request counts and timestamps per model
			var requestMu sync.Mutex
			requestCounts := make(map[string]int)
			requestTimes := make(map[string][]time.Time)

			// Mock OpenAI API responses
			env.MockOpenAI.ChatCompletionFunc = func(ctx context.Context, request openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
				// Extract model name from request
				modelName := request.Model

				// Track request and time
				requestMu.Lock()
				requestCounts[modelName]++
				requestTimes[modelName] = append(requestTimes[modelName], time.Now())
				count := requestCounts[modelName]
				requestMu.Unlock()

				// Simulate rate limit error based on conditions
				if modelName == "rate-limited-model" || count > 20 {
					// Return a rate limit error with backoff signal
					return nil, openai.FormatAPIError(errors.New("rate limit exceeded"), 429, fmt.Sprintf("Rate limit exceeded for model %s", modelName))
				}

				// Normal response
				return &openai.ChatCompletionResponse{
					ID:      "test-id",
					Object:  "chat.completion",
					Created: time.Now().Unix(),
					Choices: []openai.ChatCompletionChoice{
						{
							Index: 0,
							Message: openai.ChatMessage{
								Role:    "assistant",
								Content: "Test response for rate limit testing",
							},
							FinishReason: "stop",
						},
					},
					Usage: openai.Usage{
						PromptTokens:     10,
						CompletionTokens: 10,
						TotalTokens:      20,
					},
				}, nil
			}

			// Mock Gemini API responses
			env.MockClient.GenerateContentFunc = func(ctx context.Context, prompt string, params map[string]interface{}) (*gemini.GenerationResult, error) {
				// Extract model name from context
				modelName := ctx.Value(modelNameKey).(string)

				// Track request count for this model
				requestMu.Lock()
				requestCounts[modelName]++
				requestTimes[modelName] = append(requestTimes[modelName], time.Now())
				count := requestCounts[modelName]
				requestMu.Unlock()

				// Simulate rate limit error for specific models or conditions
				if modelName == "rate-limited-model" || count > 2 {
					// Return a rate limit error
					return nil, gemini.CreateAPIError(
						llm.CategoryRateLimit,
						"Rate limit exceeded",
						errors.New("request rate limit exceeded"),
						"Try again later",
					)
				}

				// Normal response
				return &gemini.GenerationResult{
					Content:      "Test response for rate limit testing with Gemini",
					FinishReason: "STOP",
					TokenCount:   20,
				}, nil
			}

			// Mock OpenRouter API responses
			env.MockOpenRouterClient.CompletionFunc = func(ctx context.Context, params map[string]interface{}) (*openrouter.CompletionResponse, error) {
				// Extract model name from params
				modelName, _ := params["model"].(string)

				// Track request count for this model
				requestMu.Lock()
				requestCounts[modelName]++
				requestTimes[modelName] = append(requestTimes[modelName], time.Now())
				count := requestCounts[modelName]
				requestMu.Unlock()

				// Simulate rate limit error for specific models or conditions
				if modelName == "rate-limited-model" || count > 20 {
					// Return a rate limit error
					return nil, openrouter.FormatAPIError(errors.New("rate limit exceeded"), 429, fmt.Sprintf("Rate limit exceeded for model %s", modelName))
				}

				// Normal response
				return &openrouter.CompletionResponse{
					ID:      "test-id",
					Object:  "chat.completion",
					Created: time.Now().Unix(),
					Choices: []openrouter.CompletionChoice{
						{
							Index: 0,
							Message: openrouter.Message{
								Role:    "assistant",
								Content: "Test response for rate limit testing with OpenRouter",
							},
							FinishReason: "stop",
						},
					},
					Usage: openrouter.Usage{
						PromptTokens:     10,
						CompletionTokens: 10,
						TotalTokens:      20,
					},
				}, nil
			}

			// Configure registry for the test
			regMgr := registry.NewManager(nil, env.Logger)
			for _, model := range tc.models {
				providerName := "openai" // Default provider
				if model == "gemini-pro" {
					providerName = "gemini"
				} else if model == "openrouter/model" {
					providerName = "openrouter"
				}

				// Add model to registry
				err := regMgr.AddModel(registry.ModelDefinition{
					ID:       model,
					Provider: providerName,
					Settings: map[string]interface{}{
						"temperature":       0.7,
						"top_p":             1.0,
						"context_window":    4000,
						"max_output_tokens": 1000,
					},
				})
				if err != nil {
					t.Fatalf("Failed to add model to registry: %v", err)
				}
			}

			// Create an app with rate limiting
			app, err := architect.NewApp(
				architect.WithRateLimiter(rateLimiter),
				architect.WithOpenAIClient(env.OpenAIClient),
				architect.WithGeminiClient(env.GeminiClient),
				architect.WithOpenRouterClient(env.OpenRouterClient),
				architect.WithRegistry(regMgr),
				architect.WithLogger(env.Logger),
			)
			if err != nil {
				t.Fatalf("Failed to create app: %v", err)
			}

			// Start concurrent workers to generate tasks
			startTime := time.Now()
			endTime := startTime.Add(tc.duration)
			var wg sync.WaitGroup

			for i := 0; i < tc.concurrency; i++ {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()

					for time.Now().Before(endTime) {
						// Pick a model from the list
						model := tc.models[workerID%len(tc.models)]

						// Create a context with the model name for Gemini
						ctx := context.WithValue(context.Background(), modelNameKey, model)

						// Prepare instructions
						instructions := fmt.Sprintf("Test prompt for rate limiting with model %s", model)

						// Generate content
						_, err := app.GenerateContent(ctx, instructions, []string{"test"}, architect.GenerateOptions{
							Model: model,
						})

						// We expect some rate limit errors, so don't fail the test
						if err != nil {
							env.Logger.Debugf("Worker %d got error: %v", workerID, err)
							// Sleep a bit before retrying
							time.Sleep(100 * time.Millisecond)
						}

						// Small pause between requests to avoid overwhelming the system
						time.Sleep(50 * time.Millisecond)
					}
				}(i)
			}

			// Wait for all workers to finish
			wg.Wait()

			// Analyze the request rates per model
			for model, times := range requestTimes {
				if len(times) < 2 {
					t.Logf("Not enough requests for model %s to calculate rate", model)
					continue
				}

				// Calculate rate: requests per second
				duration := times[len(times)-1].Sub(times[0]).Seconds()
				if duration == 0 {
					t.Logf("All requests for model %s happened at the same time", model)
					continue
				}

				rate := float64(len(times)) / duration
				t.Logf("Model %s: %d requests in %.2fs = %.2f req/s", model, len(times), duration, rate)

				// Check if rate exceeds expected
				if expectedMax, ok := tc.expectRates[model]; ok {
					if rate > expectedMax {
						t.Errorf("Rate for model %s (%.2f req/s) exceeds expected maximum (%.2f req/s)",
							model, rate, expectedMax)
					}
				}
			}
		})
	}
}

// TestRateLimitRecovery tests the ability to recover from rate limit errors
func TestRateLimitRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rate limit recovery test in short mode")
	}

	// Setup test environment
	env := newTestEnvironment(t)
	defer env.Cleanup()

	// Rate limiter with relatively quick recovery
	rateLimiter := ratelimit.NewRateLimiter(3, 1*time.Second)

	// Variables to track rate limit events
	var (
		rateLimitMu sync.Mutex
		rateLimited int
		recovered   int
	)

	// Mock API to trigger rate limit errors followed by successful responses
	env.MockOpenAI.ChatCompletionFunc = func(ctx context.Context, request openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
		rateLimitMu.Lock()
		limitedCount := rateLimited
		recoveredCount := recovered
		rateLimitMu.Unlock()

		// First few requests trigger rate limit (but even numbers succeed to simulate partial limiting)
		if limitedCount < 8 && limitedCount%2 == 1 {
			rateLimitMu.Lock()
			rateLimited++
			rateLimitMu.Unlock()
			return nil, openai.FormatAPIError(errors.New("rate limit exceeded"), 429, "Rate limit exceeded")
		}

		// After enough rate limits, start recovering
		if limitedCount >= 8 {
			rateLimitMu.Lock()
			recovered++
			rateLimitMu.Unlock()
		}

		// Successful response
		return &openai.ChatCompletionResponse{
			ID:      fmt.Sprintf("test-id-%d", recoveredCount),
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Choices: []openai.ChatCompletionChoice{
				{
					Index: 0,
					Message: openai.ChatMessage{
						Role:    "assistant",
						Content: "Test response after rate limit recovery",
					},
					FinishReason: "stop",
				},
			},
			Usage: openai.Usage{
				PromptTokens:     10,
				CompletionTokens: 10,
				TotalTokens:      20,
			},
		}, nil
	}

	// Configure registry
	regMgr := registry.NewManager(nil, env.Logger)
	err := regMgr.AddModel(registry.ModelDefinition{
		ID:       "gpt-3.5-turbo",
		Provider: "openai",
		Settings: map[string]interface{}{
			"temperature":       0.7,
			"top_p":             1.0,
			"context_window":    4000,
			"max_output_tokens": 1000,
		},
	})
	if err != nil {
		t.Fatalf("Failed to add model to registry: %v", err)
	}

	// Create app with rate limiter and exponential backoff
	app, err := architect.NewApp(
		architect.WithRateLimiter(rateLimiter),
		architect.WithOpenAIClient(env.OpenAIClient),
		architect.WithRegistry(regMgr),
		architect.WithLogger(env.Logger),
		architect.WithRetryOptions(architect.RetryOptions{
			MaxAttempts:       5,
			InitialBackoff:    100 * time.Millisecond,
			MaxBackoff:        1 * time.Second,
			BackoffMultiplier: 2.0,
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Make several requests and track their success
	var (
		successMu  sync.Mutex
		successful int
		failed     int
	)

	// Run concurrent requests
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Generate content
			result, err := app.GenerateContent(context.Background(), "Test prompt for rate limit recovery", []string{"test"}, architect.GenerateOptions{
				Model: "gpt-3.5-turbo",
			})

			// Track success or failure
			successMu.Lock()
			if err != nil {
				failed++
				env.Logger.Errorf("Request %d failed: %v", i, err)
			} else {
				successful++
				env.Logger.Infof("Request %d succeeded: %v", i, result.Response)
			}
			successMu.Unlock()
		}(i)
	}

	// Wait for all requests to complete
	wg.Wait()

	// Verify that we recovered from rate limiting
	rateLimitMu.Lock()
	finalRateLimited := rateLimited
	finalRecovered := recovered
	rateLimitMu.Unlock()

	successMu.Lock()
	finalSuccess := successful
	finalFailed := failed
	successMu.Unlock()

	t.Logf("Rate limit events: %d, Recovered: %d", finalRateLimited, finalRecovered)
	t.Logf("Successful requests: %d, Failed requests: %d", finalSuccess, finalFailed)

	if finalRateLimited == 0 {
		t.Errorf("Expected rate limiting to occur, but no rate limit events were triggered")
	}

	if finalRecovered == 0 {
		t.Errorf("Expected recovery from rate limiting, but no recovered requests were recorded")
	}

	if finalSuccess == 0 {
		t.Errorf("Expected some requests to succeed, but none did")
	}

	// Percentage of successful requests should be reasonable
	successRate := float64(finalSuccess) / float64(finalSuccess+finalFailed)
	t.Logf("Success rate: %.2f", successRate)
	if successRate < 0.5 {
		t.Errorf("Success rate too low: %.2f", successRate)
	}
}

// TestRateLimitDistribution tests that the rate limiter distributes capacity fairly
func TestRateLimitDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rate limit distribution test in short mode")
	}

	// Setup test environment
	env := newTestEnvironment(t)
	defer env.Cleanup()

	// Rate limiter with moderate capacity
	rateLimiter := ratelimit.NewRateLimiter(10, 1*time.Second)

	// Configure registry with several models
	regMgr := registry.NewManager(nil, env.Logger)
	models := []string{"model-a", "model-b", "model-c"}

	for _, model := range models {
		err := regMgr.AddModel(registry.ModelDefinition{
			ID:       model,
			Provider: "openai",
			Settings: map[string]interface{}{
				"temperature":       0.7,
				"context_window":    4000,
				"max_output_tokens": 1000,
			},
		})
		if err != nil {
			t.Fatalf("Failed to add model to registry: %v", err)
		}
	}

	// Track completed requests per model
	var (
		resultsMu         sync.Mutex
		completedRequests = make(map[string]int)
	)

	// Mock API that counts successful requests by model
	env.MockOpenAI.ChatCompletionFunc = func(ctx context.Context, request openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
		modelName := request.Model

		resultsMu.Lock()
		completedRequests[modelName]++
		resultsMu.Unlock()

		return &openai.ChatCompletionResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Choices: []openai.ChatCompletionChoice{
				{
					Index: 0,
					Message: openai.ChatMessage{
						Role:    "assistant",
						Content: "Test response for model " + modelName,
					},
					FinishReason: "stop",
				},
			},
		}, nil
	}

	// Create app with rate limiter
	app, err := architect.NewApp(
		architect.WithRateLimiter(rateLimiter),
		architect.WithOpenAIClient(env.OpenAIClient),
		architect.WithRegistry(regMgr),
		architect.WithLogger(env.Logger),
	)
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Make requests distributed unevenly across models
	var wg sync.WaitGroup
	requestCounts := map[string]int{
		"model-a": 30, // High priority model
		"model-b": 20, // Medium priority
		"model-c": 10, // Low priority
	}

	// Launch all requests concurrently
	for model, count := range requestCounts {
		for i := 0; i < count; i++ {
			wg.Add(1)
			go func(model string, requestID int) {
				defer wg.Done()

				// Generate content
				_, err := app.GenerateContent(context.Background(),
					fmt.Sprintf("Test prompt for model %s, request %d", model, requestID),
					[]string{"test"},
					architect.GenerateOptions{
						Model: model,
					})

				if err != nil {
					// We expect some rate limit errors
					env.Logger.Debugf("Request for %s failed: %v", model, err)
				}

			}(model, i)
		}
	}

	// Wait for all requests to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All requests completed
	case <-time.After(10 * time.Second):
		t.Log("Test timed out waiting for all requests to complete")
	}

	// Analyze distribution of completed requests
	resultsMu.Lock()
	defer resultsMu.Unlock()

	t.Logf("Completed requests by model:")
	totalCompleted := 0
	for _, model := range models {
		completed := completedRequests[model]
		totalCompleted += completed
		t.Logf("  %s: %d/%d (%.2f%%)", model, completed, requestCounts[model],
			float64(completed)/float64(requestCounts[model])*100)
	}

	// Check that each model got a reasonable share of completions
	// We're looking for roughly fair distribution, not perfect equality
	if totalCompleted == 0 {
		t.Errorf("No requests completed successfully")
		return
	}

	minSuccessRate := 0.4 // At least 40% of requests should succeed for each model
	for _, model := range models {
		completed := completedRequests[model]
		rate := float64(completed) / float64(requestCounts[model])
		if rate < minSuccessRate {
			t.Errorf("Model %s has too low success rate: %.2f", model, rate)
		}
	}
}
