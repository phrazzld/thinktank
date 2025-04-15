package orchestrator

import (
	"context"
	"testing"

	"github.com/phrazzld/architect/internal/architect/interfaces"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/ratelimit"
)

// TestOrchestrator_Run_UsesLLMClientExclusively tests that the Orchestrator
// only uses the provider-agnostic LLMClient interface.
func TestOrchestrator_Run_UsesLLMClientExclusively(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "")
	apiService := &mockAPIService{}

	// Track API service calls
	apiService.InitLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
		// Return a mock LLM client
		return &mockLLMClient{
			modelName: modelName,
		}, nil
	}

	// Create dependencies
	contextGatherer := &mockContextGatherer{
		GatherContextFunc: func(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
			// Return empty context for simplicity
			return []fileutil.FileMeta{}, &interfaces.ContextStats{}, nil
		},
	}

	tokenManager := &mockTokenManager{
		CheckTokenLimitFunc: func(ctx context.Context, prompt string) error {
			return nil
		},
	}

	fileWriter := &mockFileWriter{
		SaveToFileFunc: func(content, outputFile string) error {
			return nil
		},
	}

	auditLogger := &mockAuditLogger{}

	// Configure a simple rate limiter and config
	rateLimiter := ratelimit.NewRateLimiter(1, 1)
	cfg := config.NewDefaultCliConfig()
	cfg.ModelNames = []string{"test-model"}
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create the orchestrator
	orch := NewOrchestrator(
		apiService,
		contextGatherer,
		tokenManager,
		fileWriter,
		auditLogger,
		rateLimiter,
		cfg,
		logger,
	)

	// Run the orchestrator
	err := orch.Run(ctx, "test instructions")

	// Check for no errors
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify we only used the LLMClient interface
	if len(apiService.InitLLMClientCalls) != len(cfg.ModelNames) {
		t.Errorf("Expected %d calls to InitLLMClient, got %d", len(cfg.ModelNames), len(apiService.InitLLMClientCalls))
	}

	if len(apiService.InitClientCalls) > 0 {
		t.Error("Unexpected calls to deprecated InitClient method")
	}
}

// TestOrchestrator_APIServiceAdapter_UsesLLMClientExclusively tests that the APIServiceAdapter
// properly delegates to the underlying APIService and only uses LLMClient interface methods.
func TestOrchestrator_APIServiceAdapter_UsesLLMClientExclusively(t *testing.T) {
	// Create a mock APIService
	mockService := &mockAPIService{}

	// Setup the InitLLMClient mock
	expectedClient := &mockLLMClient{modelName: "test-model"}
	mockService.InitLLMClientFunc = func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
		return expectedClient, nil
	}

	// Setup the ProcessLLMResponse mock
	expectedResult := &llm.ProviderResult{
		Content: "Test content",
	}
	mockService.ProcessLLMResponseFunc = func(result *llm.ProviderResult) (string, error) {
		if result != expectedResult {
			t.Errorf("ProcessLLMResponse received unexpected result")
		}
		return result.Content, nil
	}

	// Create the adapter
	adapter := &APIServiceAdapter{APIService: mockService}

	// Test InitLLMClient
	ctx := context.Background()
	client, err := adapter.InitLLMClient(ctx, "apikey", "test-model", "endpoint")

	// Verify the results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if client != expectedClient {
		t.Errorf("Expected client %v, got %v", expectedClient, client)
	}

	// Test ProcessLLMResponse
	content, err := adapter.ProcessLLMResponse(expectedResult)

	// Verify the results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if content != expectedResult.Content {
		t.Errorf("Expected content %q, got %q", expectedResult.Content, content)
	}

	// Verify call counts
	if len(mockService.InitLLMClientCalls) != 1 {
		t.Errorf("Expected 1 call to InitLLMClient, got %d", len(mockService.InitLLMClientCalls))
	}
	if len(mockService.ProcessLLMResponseCalls) != 1 {
		t.Errorf("Expected 1 call to ProcessLLMResponse, got %d", len(mockService.ProcessLLMResponseCalls))
	}
}
