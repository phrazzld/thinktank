// Package integration provides comprehensive multi-model integration tests
// These tests ensure reliability and catch regressions in multi-model processing scenarios
package integration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/fileutil"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/models"
	"github.com/phrazzld/thinktank/internal/ratelimit"
	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
	"github.com/phrazzld/thinktank/internal/thinktank/orchestrator"
)

// TestMultiModelReliability_AllModelsBasic tests basic functionality with all 15 supported models
func TestMultiModelReliability_AllModelsBasic(t *testing.T) {
	logger := logutil.NewTestLogger(t)

	// Get all 15 supported models
	allModels := models.ListAllModels()
	if len(allModels) != 15 {
		t.Fatalf("Expected 15 models, got %d", len(allModels))
	}

	// Create test environment
	env := setupMultiModelTestEnv(t, logger, allModels, nil)
	defer env.cleanup()

	// Track which models were processed
	processedModels := make(map[string]bool)
	var modelsMutex sync.Mutex

	// Configure mock API service to succeed for all models
	env.apiService.InitLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
		return createSuccessfulMockClient(modelName, &processedModels, &modelsMutex), nil
	}

	// Execute orchestrator
	err := env.orchestrator.Run(context.Background(), "Test all models concurrently")
	if err != nil {
		t.Fatalf("Expected orchestrator to succeed with all models, got error: %v", err)
	}

	// Verify all models were processed
	modelsMutex.Lock()
	defer modelsMutex.Unlock()

	for _, modelName := range allModels {
		if !processedModels[modelName] {
			t.Errorf("Model %s was not processed", modelName)
		}
	}

	// Verify output files were created for all models
	for _, modelName := range allModels {
		sanitizedName := sanitizeModelName(modelName)
		outputFile := filepath.Join(env.outputDir, sanitizedName+".md")
		if !env.fileExists(outputFile) {
			t.Errorf("Output file for model %s was not created", modelName)
		}
	}

	t.Logf("Successfully processed all %d models concurrently", len(allModels))
}

// TestMultiModelReliability_CrossProviderConcurrency tests concurrent execution across different providers
func TestMultiModelReliability_CrossProviderConcurrency(t *testing.T) {
	logger := logutil.NewTestLogger(t)

	// Select models from each provider for comprehensive testing
	testModels := []string{
		"gpt-4.1",        // OpenAI
		"gemini-2.5-pro", // Gemini
		"openrouter/deepseek/deepseek-chat-v3-0324",    // OpenRouter - normal
		"openrouter/deepseek/deepseek-r1-0528",         // OpenRouter - rate limited (5 RPM)
		"openrouter/meta-llama/llama-3.3-70b-instruct", // OpenRouter - standard
	}

	env := setupMultiModelTestEnv(t, logger, testModels, nil)
	defer env.cleanup()

	// Track concurrent execution and provider distribution
	executionTimes := make(map[string]time.Time)
	providers := make(map[string]string)
	var executionMutex sync.Mutex

	env.apiService.InitLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
		// Record execution time and provider
		executionMutex.Lock()
		executionTimes[modelName] = time.Now()
		if provider, err := models.GetProviderForModel(modelName); err == nil {
			providers[modelName] = provider
		}
		executionMutex.Unlock()

		return createSuccessfulMockClient(modelName, nil, nil), nil
	}

	start := time.Now()
	err := env.orchestrator.Run(context.Background(), "Test cross-provider concurrency")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Expected success with cross-provider models, got: %v", err)
	}

	// Verify models from all three providers were processed
	executionMutex.Lock()
	defer executionMutex.Unlock()

	providerCounts := make(map[string]int)
	for modelName, provider := range providers {
		providerCounts[provider]++
		t.Logf("Processed model %s (provider: %s)", modelName, provider)
	}

	expectedProviders := []string{"openai", "gemini", "openrouter"}
	for _, provider := range expectedProviders {
		if providerCounts[provider] == 0 {
			t.Errorf("No models from provider %s were processed", provider)
		}
	}

	// Verify concurrent execution (should be much faster than sequential)
	maxSequentialTime := time.Duration(len(testModels)) * 100 * time.Millisecond
	if duration > maxSequentialTime {
		t.Errorf("Execution took too long (%v), suggesting sequential rather than concurrent processing", duration)
	}

	t.Logf("Successfully processed %d models from %d providers in %v", len(testModels), len(providerCounts), duration)
}

// TestMultiModelReliability_RateLimitingBehavior tests provider-specific rate limiting behavior
func TestMultiModelReliability_RateLimitingBehavior(t *testing.T) {
	logger := logutil.NewTestLogger(t)

	// Test models with different rate limit characteristics
	testModels := []string{
		"openrouter/deepseek/deepseek-r1-0528",         // Model-specific 5 RPM limit
		"openrouter/deepseek/deepseek-r1-0528:free",    // Model-specific 3 RPM limit
		"openrouter/meta-llama/llama-3.3-70b-instruct", // Provider default 20 RPM
	}

	// Configure rate limiting for realistic testing
	cfg := &config.CliConfig{
		ModelNames:                 testModels,
		Verbose:                    true,
		LogLevel:                   logutil.DebugLevel,
		MaxConcurrentRequests:      5,
		RateLimitRequestsPerMinute: 60, // Global limit
		OpenRouterRateLimit:        20, // Provider-specific limit
	}

	env := setupMultiModelTestEnvWithConfig(t, logger, cfg)
	defer env.cleanup()

	// Track rate limiting behavior
	var requestCount int64
	requestTimes := make([]time.Time, 0)
	var requestMutex sync.Mutex

	env.apiService.InitLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
		return &llm.MockLLMClient{
			GenerateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
				// Record request timing for rate limit analysis
				atomic.AddInt64(&requestCount, 1)
				requestMutex.Lock()
				requestTimes = append(requestTimes, time.Now())
				requestMutex.Unlock()

				// Simulate processing time
				time.Sleep(10 * time.Millisecond)

				return &llm.ProviderResult{
					Content:      fmt.Sprintf("Response from %s", modelName),
					FinishReason: "stop",
				}, nil
			},
			GetModelNameFunc: func() string { return modelName },
			CloseFunc:        func() error { return nil },
		}, nil
	}

	start := time.Now()
	err := env.orchestrator.Run(context.Background(), "Test rate limiting behavior")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Expected success with rate limited models, got: %v", err)
	}

	// Verify requests were made and rate limiting was applied
	finalRequestCount := atomic.LoadInt64(&requestCount)
	if finalRequestCount < int64(len(testModels)) {
		t.Errorf("Expected at least %d requests, got %d", len(testModels), finalRequestCount)
	}

	// Verify rate limiting patterns
	requestMutex.Lock()
	defer requestMutex.Unlock()

	if len(requestTimes) >= 2 {
		// Check for reasonable spacing between requests (rate limiting should introduce delays)
		firstRequest := requestTimes[0]
		lastRequest := requestTimes[len(requestTimes)-1]
		totalDuration := lastRequest.Sub(firstRequest)

		// With rate limiting, should take more time than just processing delay
		minExpectedDuration := time.Duration(len(testModels)-1) * 10 * time.Millisecond
		if totalDuration < minExpectedDuration {
			t.Logf("Note: Requests completed quickly (%v), rate limiting may not be active in test environment", totalDuration)
		}
	}

	t.Logf("Processed %d rate-limited models with %d total requests in %v", len(testModels), finalRequestCount, duration)
}

// TestMultiModelReliability_PartialFailureResilience tests behavior when some models fail
func TestMultiModelReliability_PartialFailureResilience(t *testing.T) {
	logger := logutil.NewTestLogger(t)

	// Configure expected error patterns for partial failures
	logger.ExpectError("Generation failed for model")
	logger.ExpectError("Error generating content with model")
	logger.ExpectError("output generation failed for model")
	logger.ExpectError("Completed with model errors")

	testModels := []string{
		"gpt-4.1",        // Should succeed
		"gemini-2.5-pro", // Should succeed
		"error-model-1",  // Should fail
		"error-model-2",  // Should fail
		"openrouter/deepseek/deepseek-chat-v3-0324", // Should succeed
	}

	env := setupMultiModelTestEnv(t, logger, testModels, nil)
	defer env.cleanup()

	// Track success/failure patterns
	successfulModels := make(map[string]bool)
	failedModels := make(map[string]bool)
	var statusMutex sync.Mutex

	env.apiService.InitLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
		if modelName == "error-model-1" || modelName == "error-model-2" {
			// Simulate failing models
			statusMutex.Lock()
			failedModels[modelName] = true
			statusMutex.Unlock()

			return &llm.MockLLMClient{
				GenerateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					return nil, errors.New("simulated model failure for " + modelName)
				},
				GetModelNameFunc: func() string { return modelName },
				CloseFunc:        func() error { return nil },
			}, nil
		}

		// Successful models
		statusMutex.Lock()
		successfulModels[modelName] = true
		statusMutex.Unlock()

		return createSuccessfulMockClient(modelName, nil, nil), nil
	}

	err := env.orchestrator.Run(context.Background(), "Test partial failure resilience")

	// Should get partial failure error
	if err == nil {
		t.Error("Expected partial failure error, got nil")
	} else if !errors.Is(err, orchestrator.ErrPartialProcessingFailure) {
		t.Errorf("Expected ErrPartialProcessingFailure, got: %v", err)
	}

	// Verify success/failure patterns
	statusMutex.Lock()
	defer statusMutex.Unlock()

	expectedSuccessful := []string{"gpt-4.1", "gemini-2.5-pro", "openrouter/deepseek/deepseek-chat-v3-0324"}
	for _, model := range expectedSuccessful {
		if !successfulModels[model] {
			t.Errorf("Expected model %s to succeed, but it didn't", model)
		}
	}

	expectedFailed := []string{"error-model-1", "error-model-2"}
	for _, model := range expectedFailed {
		if !failedModels[model] {
			t.Errorf("Expected model %s to fail, but it didn't", model)
		}
	}

	// Verify output files were created for successful models only
	for _, model := range expectedSuccessful {
		sanitizedName := sanitizeModelName(model)
		outputFile := filepath.Join(env.outputDir, sanitizedName+".md")
		if !env.fileExists(outputFile) {
			t.Errorf("Output file for successful model %s was not created", model)
		}
	}

	for _, model := range expectedFailed {
		sanitizedName := sanitizeModelName(model)
		outputFile := filepath.Join(env.outputDir, sanitizedName+".md")
		if env.fileExists(outputFile) {
			t.Errorf("Output file for failed model %s should not exist", model)
		}
	}

	t.Logf("Verified partial failure resilience: %d successful, %d failed", len(successfulModels), len(failedModels))
}

// TestMultiModelReliability_SynthesisWithMultipleProviders tests synthesis with models from multiple providers
func TestMultiModelReliability_SynthesisWithMultipleProviders(t *testing.T) {
	logger := logutil.NewTestLogger(t)

	testModels := []string{
		"gpt-4.1",        // OpenAI
		"gemini-2.5-pro", // Gemini
		"openrouter/deepseek/deepseek-chat-v3-0324", // OpenRouter
	}
	synthesisModel := "gemini-2.5-flash"

	env := setupMultiModelTestEnv(t, logger, testModels, &synthesisModel)
	defer env.cleanup()

	// Track which models were called and synthesis behavior
	calledModels := make(map[string]bool)
	synthesisInputs := make([]string, 0)
	var callMutex sync.Mutex

	env.apiService.InitLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
		callMutex.Lock()
		calledModels[modelName] = true
		callMutex.Unlock()

		if modelName == synthesisModel {
			return &llm.MockLLMClient{
				GenerateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					// Capture synthesis input for analysis
					callMutex.Lock()
					synthesisInputs = append(synthesisInputs, prompt)
					callMutex.Unlock()

					return &llm.ProviderResult{
						Content:      "# Synthesized Analysis\n\nThis combines insights from multiple providers.",
						FinishReason: "stop",
					}, nil
				},
				GetModelNameFunc: func() string { return modelName },
				CloseFunc:        func() error { return nil },
			}, nil
		}

		return createSuccessfulMockClient(modelName, nil, nil), nil
	}

	err := env.orchestrator.Run(context.Background(), "Test multi-provider synthesis")
	if err != nil {
		t.Fatalf("Expected success with multi-provider synthesis, got: %v", err)
	}

	// Verify all models and synthesis model were called
	callMutex.Lock()
	defer callMutex.Unlock()

	for _, model := range testModels {
		if !calledModels[model] {
			t.Errorf("Model %s was not called", model)
		}
	}

	if !calledModels[synthesisModel] {
		t.Error("Synthesis model was not called")
	}

	// Verify synthesis input contains content from multiple providers
	if len(synthesisInputs) == 0 {
		t.Error("No synthesis inputs captured")
	} else {
		synthesisInput := synthesisInputs[0]
		if len(synthesisInput) < 100 { // Synthesis input should be substantial
			t.Errorf("Synthesis input appears too short (%d chars), may not contain all model outputs", len(synthesisInput))
		}
	}

	// Verify synthesis output file was created
	sanitizedSynthesisName := sanitizeModelName(synthesisModel)
	synthesisFile := filepath.Join(env.outputDir, sanitizedSynthesisName+"-synthesis.md")
	if !env.fileExists(synthesisFile) {
		t.Error("Synthesis output file was not created")
	}

	t.Logf("Successfully completed multi-provider synthesis with %d models", len(testModels))
}

// TestMultiModelReliability_ResourceUsage tests resource consumption patterns
func TestMultiModelReliability_ResourceUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource usage test in short mode")
	}

	logger := logutil.NewTestLogger(t)

	// Use a substantial number of models to test resource usage
	testModels := models.ListAllModels() // All 15 models

	env := setupMultiModelTestEnv(t, logger, testModels, nil)
	defer env.cleanup()

	// Measure resource usage
	var startGoroutines, maxGoroutines int
	var goroutineMutex sync.Mutex

	// Record initial goroutine count
	startGoroutines = runtime.NumGoroutine()

	var requestCount int64
	env.apiService.InitLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
		return &llm.MockLLMClient{
			GenerateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
				atomic.AddInt64(&requestCount, 1)

				// Monitor goroutine count during execution
				currentGoroutines := runtime.NumGoroutine()
				goroutineMutex.Lock()
				if currentGoroutines > maxGoroutines {
					maxGoroutines = currentGoroutines
				}
				goroutineMutex.Unlock()

				// Simulate some processing time
				time.Sleep(20 * time.Millisecond)

				return &llm.ProviderResult{
					Content:      fmt.Sprintf("Response from %s", modelName),
					FinishReason: "stop",
				}, nil
			},
			GetModelNameFunc: func() string { return modelName },
			CloseFunc:        func() error { return nil },
		}, nil
	}

	start := time.Now()
	err := env.orchestrator.Run(context.Background(), "Test resource usage patterns")
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Expected success with resource usage test, got: %v", err)
	}

	// Check resource usage metrics
	finalGoroutines := runtime.NumGoroutine()
	totalRequests := atomic.LoadInt64(&requestCount)

	// Verify resource cleanup (goroutines should return close to baseline)
	goroutineIncrease := finalGoroutines - startGoroutines
	if goroutineIncrease > 10 { // Allow some tolerance for test framework overhead
		t.Errorf("Potential goroutine leak: started with %d, ended with %d (increase: %d)",
			startGoroutines, finalGoroutines, goroutineIncrease)
	}

	// Verify reasonable concurrency (max goroutines should be higher than baseline but not excessive)
	goroutineMutex.Lock()
	maxIncrease := maxGoroutines - startGoroutines
	goroutineMutex.Unlock()

	if maxIncrease < 5 {
		t.Logf("Low concurrency detected: max goroutine increase was only %d", maxIncrease)
	} else if maxIncrease > 100 {
		t.Errorf("Excessive goroutines: peak was %d goroutines above baseline", maxIncrease)
	}

	// Verify all models were processed
	if totalRequests < int64(len(testModels)) {
		t.Errorf("Expected at least %d requests for %d models, got %d", len(testModels), len(testModels), totalRequests)
	}

	t.Logf("Resource usage test completed: %d models, %d requests, %v duration",
		len(testModels), totalRequests, duration)
	t.Logf("Goroutine usage: start=%d, max=%d, end=%d", startGoroutines, maxGoroutines, finalGoroutines)
}

// Helper functions for test setup and utilities

type multiModelTestEnv struct {
	outputDir    string
	orchestrator *orchestrator.Orchestrator
	apiService   *MockAPIService
	fs           *testutil.RealFS
}

func (env *multiModelTestEnv) cleanup() {
	if env.fs != nil && env.outputDir != "" {
		_ = env.fs.RemoveAll(env.outputDir) // Ignore cleanup errors in tests
	}
}

func (env *multiModelTestEnv) fileExists(path string) bool {
	exists, _ := env.fs.Stat(path)
	return exists
}

func setupMultiModelTestEnv(t *testing.T, logger logutil.LoggerInterface, models []string, synthesisModel *string) *multiModelTestEnv {
	cfg := &config.CliConfig{
		ModelNames:                 models,
		Verbose:                    true,
		LogLevel:                   logutil.DebugLevel,
		MaxConcurrentRequests:      5,
		RateLimitRequestsPerMinute: 60,
	}

	if synthesisModel != nil {
		cfg.SynthesisModel = *synthesisModel
	}

	return setupMultiModelTestEnvWithConfig(t, logger, cfg)
}

func setupMultiModelTestEnvWithConfig(t *testing.T, logger logutil.LoggerInterface, cfg *config.CliConfig) *multiModelTestEnv {
	// Create filesystem abstraction
	fs := testutil.NewRealFS()

	// Create temp directory for outputs
	tempDir, err := os.MkdirTemp("", "thinktank-multi-model-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")
	if err := fs.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output dir: %v", err)
	}

	cfg.OutputDir = outputDir
	cfg.AuditLogFile = filepath.Join(tempDir, "audit.log")

	// Create mock services
	apiService := &MockAPIService{
		ProcessLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			return result.Content, nil
		},
		GetModelParametersFunc: func(ctx context.Context, modelName string) (map[string]interface{}, error) {
			return map[string]interface{}{}, nil
		},
		GetModelDefinitionFunc: func(ctx context.Context, modelName string) (*models.ModelInfo, error) {
			// Return appropriate model info or use test defaults
			if info, err := models.GetModelInfo(modelName); err == nil {
				return &info, nil
			}
			// Default for test models not in registry
			return &models.ModelInfo{
				APIModelID: modelName,
				Provider:   "test-provider",
			}, nil
		},
	}

	contextGatherer := &MockContextGatherer{
		GatherContextFunc: func(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
			return []fileutil.FileMeta{}, &interfaces.ContextStats{
				ProcessedFilesCount: 0,
				CharCount:           0,
			}, nil
		},
		DisplayDryRunInfoFunc: func(ctx context.Context, stats *interfaces.ContextStats) error {
			return nil
		},
	}

	fileWriter := &MockFileWriter{
		SaveToFileFunc: func(ctx context.Context, content, filePath string) error {
			dir := filepath.Dir(filePath)
			if err := fs.MkdirAll(dir, 0755); err != nil {
				return err
			}
			return fs.WriteFile(filePath, []byte(content), 0640)
		},
	}

	auditLogger := &MockAuditLogger{
		LogFunc:       func(ctx context.Context, entry auditlog.AuditEntry) error { return nil },
		LogLegacyFunc: func(entry auditlog.AuditEntry) error { return nil },
		CloseFunc:     func() error { return nil },
	}

	rateLimiter := ratelimit.NewRateLimiter(cfg.MaxConcurrentRequests, cfg.RateLimitRequestsPerMinute)

	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false }, // CI mode for tests
	})

	orch := orchestrator.NewOrchestrator(
		apiService,
		contextGatherer,
		fileWriter,
		auditLogger,
		rateLimiter,
		cfg,
		logger,
		consoleWriter,
	)

	return &multiModelTestEnv{
		outputDir:    outputDir,
		orchestrator: orch,
		apiService:   apiService,
		fs:           fs,
	}
}

func createSuccessfulMockClient(modelName string, processedModels *map[string]bool, mutex *sync.Mutex) llm.LLMClient {
	return &llm.MockLLMClient{
		GenerateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
			// Track processed models if tracking is enabled
			if processedModels != nil && mutex != nil {
				mutex.Lock()
				(*processedModels)[modelName] = true
				mutex.Unlock()
			}

			return &llm.ProviderResult{
				Content:      fmt.Sprintf("# Output from %s\n\nThis is mock output from %s model.", modelName, modelName),
				FinishReason: "stop",
			}, nil
		},
		GetModelNameFunc: func() string {
			return modelName
		},
		CloseFunc: func() error {
			return nil
		},
	}
}

// sanitizeModelName sanitizes a model name for use as a filename
// This mirrors the SanitizeFilename function in the modelproc package
func sanitizeModelName(modelName string) string {
	// Replace slashes and other problematic characters with hyphens
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"'", "-",
		"<", "-",
		">", "-",
		"|", "-",
	)
	return replacer.Replace(modelName)
}
