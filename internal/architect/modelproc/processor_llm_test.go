package modelproc_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/llm"
)

// TestModelProcessor_Process_WithLLMClient tests the ModelProcessor.Process method
// with a provider-agnostic LLMClient.
func TestModelProcessor_Process_WithLLMClient(t *testing.T) {
	// Setup mocks
	mockAPI := &mockLLMAPIService{
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return &mockLLMClient{
				getModelNameFunc: func() string { return modelName },
				generateContentFunc: func(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
					return &llm.ProviderResult{
						Content:    "Generated LLM content",
						TokenCount: 50,
					}, nil
				},
			}, nil
		},
		processLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			return result.Content, nil
		},
	}

	// Track if SaveToFile was called
	saveToFileCalled := false
	mockWriter := &mockFileWriter{
		saveToFileFunc: func(content, outputFile string) error {
			saveToFileCalled = true
			// Verify the content and output path are what we expect
			if content != "Generated LLM content" {
				t.Errorf("Expected content 'Generated LLM content', got '%s'", content)
			}
			if filepath.Base(outputFile) != "test-llm-model.md" {
				t.Errorf("Expected output file 'test-llm-model.md', got '%s'", filepath.Base(outputFile))
			}
			return nil
		},
	}

	// Track the audit log entries
	auditEntries := make([]auditlog.AuditEntry, 0)
	mockAudit := &mockAuditLogger{
		logFunc: func(entry auditlog.AuditEntry) error {
			auditEntries = append(auditEntries, entry)
			return nil
		},
	}

	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-llm-model",
		"Test prompt",
	)

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !saveToFileCalled {
		t.Errorf("Expected SaveToFile to be called, but it wasn't")
	}

	// Verify audit log entries
	expectedOperations := []string{
		"CheckTokensStart",
		"CheckTokens",
		"GenerateContentStart",
		"GenerateContentEnd",
		"SaveOutputStart",
		"SaveOutputEnd",
	}

	if len(auditEntries) != len(expectedOperations) {
		t.Errorf("Expected %d audit entries, got %d", len(expectedOperations), len(auditEntries))
	} else {
		for i, operation := range expectedOperations {
			if auditEntries[i].Operation != operation {
				t.Errorf("Expected audit operation '%s', got '%s'", operation, auditEntries[i].Operation)
			}
		}
	}
}

// TestModelProcessor_Process_UseTokenManagerWithLLMClient specifically tests that
// token management in the processor uses the LLMClient interface correctly.
func TestModelProcessor_Process_UseTokenManagerWithLLMClient(t *testing.T) {
	// Create a mock LLM client that counts calls to different methods
	getModelInfoCalled := false
	countTokensCalled := false
	mockLLM := &mockLLMClient{
		getModelNameFunc: func() string { return "test-model" },
		getModelInfoFunc: func(ctx context.Context) (*llm.ProviderModelInfo, error) {
			getModelInfoCalled = true
			return &llm.ProviderModelInfo{
				Name:             "test-model",
				InputTokenLimit:  1000,
				OutputTokenLimit: 500,
			}, nil
		},
		countTokensFunc: func(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
			countTokensCalled = true
			return &llm.ProviderTokenCount{
				Total: 100,
			}, nil
		},
		generateContentFunc: func(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
			return &llm.ProviderResult{
				Content:    "Test content",
				TokenCount: 50,
			}, nil
		},
	}

	mockAPI := &mockLLMAPIService{
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return mockLLM, nil
		},
		processLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			return result.Content, nil
		},
	}

	// Configure other mocks
	mockWriter := &mockFileWriter{
		saveToFileFunc: func(content, outputFile string) error {
			return nil
		},
	}
	mockAudit := &mockAuditLogger{
		logFunc: func(entry auditlog.AuditEntry) error {
			return nil
		},
	}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify that token management methods were called
	if !getModelInfoCalled {
		t.Errorf("Expected GetModelInfo to be called, but it wasn't")
	}

	if !countTokensCalled {
		t.Errorf("Expected CountTokens to be called, but it wasn't")
	}
}

// TestModelProcessor_Process_UsesLLMClientExclusively tests that the ModelProcessor
// only uses the LLMClient interface and doesn't try to use any provider-specific APIs.
func TestModelProcessor_Process_UsesLLMClientExclusively(t *testing.T) {
	// Make a special mock that tracks if any provider-specific conversions are attempted
	mockLLM := &mockLLMClientWithProviderCheck{
		modelName: "test-model",
	}

	mockAPI := &mockLLMAPIService{
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return mockLLM, nil
		},
		processLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			// Ensure we only get llm.ProviderResult and not something that was converted
			if mockLLM.convertedFromProvider {
				t.Error("Processor is converting from provider-specific types instead of using llm.ProviderResult directly")
			}
			return result.Content, nil
		},
	}

	// Other mocks
	mockWriter := &mockFileWriter{
		saveToFileFunc: func(content, outputFile string) error {
			return nil
		},
	}
	mockAudit := &mockAuditLogger{
		logFunc: func(entry auditlog.AuditEntry) error {
			return nil
		},
	}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run the processor
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify success
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that our mock client was used properly
	if mockLLM.convertedFromProvider {
		t.Error("The processor attempted to convert between provider-specific and generic types")
	}
}

// mockLLMClientWithProviderCheck is a special mock LLM client that checks
// if any provider-specific conversions are attempted.
type mockLLMClientWithProviderCheck struct {
	modelName             string
	convertedFromProvider bool
}

func (m *mockLLMClientWithProviderCheck) GenerateContent(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
	return &llm.ProviderResult{
		Content:    "Test content",
		TokenCount: 50,
	}, nil
}

func (m *mockLLMClientWithProviderCheck) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	return &llm.ProviderTokenCount{
		Total: 100,
	}, nil
}

func (m *mockLLMClientWithProviderCheck) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	return &llm.ProviderModelInfo{
		Name:             m.modelName,
		InputTokenLimit:  1000,
		OutputTokenLimit: 500,
	}, nil
}

func (m *mockLLMClientWithProviderCheck) GetModelName() string {
	return m.modelName
}

func (m *mockLLMClientWithProviderCheck) Close() error {
	return nil
}

// mockLLMAPIService mocks the modelproc.APIService interface with LLM support
type mockLLMAPIService struct {
	initLLMClientFunc        func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)
	processLLMResponseFunc   func(result *llm.ProviderResult) (string, error)
	isEmptyResponseErrorFunc func(err error) bool
	isSafetyBlockedErrorFunc func(err error) bool
	getErrorDetailsFunc      func(err error) string
}

func (m *mockLLMAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	return m.initLLMClientFunc(ctx, apiKey, modelName, apiEndpoint)
}

func (m *mockLLMAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	return m.processLLMResponseFunc(result)
}

func (m *mockLLMAPIService) IsEmptyResponseError(err error) bool {
	if m.isEmptyResponseErrorFunc != nil {
		return m.isEmptyResponseErrorFunc(err)
	}
	return false
}

func (m *mockLLMAPIService) IsSafetyBlockedError(err error) bool {
	if m.isSafetyBlockedErrorFunc != nil {
		return m.isSafetyBlockedErrorFunc(err)
	}
	return false
}

func (m *mockLLMAPIService) GetErrorDetails(err error) string {
	if m.getErrorDetailsFunc != nil {
		return m.getErrorDetailsFunc(err)
	}
	return fmt.Sprintf("error: %v", err)
}

// We're using the mockLLMClient from mocks_test.go
// All necessary methods are already defined there

// These are needed for backward compatibility with the old interface
func (m *mockLLMAPIService) InitClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (interface{}, error) {
	// For tests, we just wrap the LLM client
	llmClient, err := m.InitLLMClient(ctx, apiKey, modelName, apiEndpoint)
	if err != nil {
		return nil, err
	}
	return llmClient, nil
}

func (m *mockLLMAPIService) ProcessResponse(result interface{}) (string, error) {
	// For tests, we just convert to LLM result type
	llmResult, ok := result.(*llm.ProviderResult)
	if !ok {
		return "", fmt.Errorf("unexpected result type: %T", result)
	}
	return m.ProcessLLMResponse(llmResult)
}
