// Package modelproc provides model processing functionality for the architect tool.
// This file contains the core processor tests for success cases.
package modelproc_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/registry"
)

// Import necessary dependencies
// Note: Most mock definitions have been moved to mocks_test.go

func TestModelProcessor_Process_Success(t *testing.T) {
	// Save original factory function and restore after test
	defer restoreNewTokenManagerWithClient()

	// Setup mocks
	mockAPI := &mockAPIService{
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return &mockLLMClient{
				generateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					return &llm.ProviderResult{
						Content:    "Generated content",
						TokenCount: 50,
					}, nil
				},
			}, nil
		},
		processLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			return result.Content, nil
		},
	}

	// Create mock token manager
	mockToken := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			return &modelproc.TokenResult{
				TokenCount:   100,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   10.0,
			}, nil
		},
	}

	// Mock the factory function
	modelproc.NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) modelproc.TokenManager {
		return mockToken
	}

	// Track if SaveToFile was called
	saveToFileCalled := false
	mockWriter := &mockFileWriter{
		saveToFileFunc: func(content, outputFile string) error {
			saveToFileCalled = true
			// Verify the content and output path are what we expect
			if content != "Generated content" {
				t.Errorf("Expected content 'Generated content', got '%s'", content)
			}
			if filepath.Base(outputFile) != "test-model.md" {
				t.Errorf("Expected output file 'test-model.md', got '%s'", filepath.Base(outputFile))
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

	// Create processor with updated constructor signature
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

	if !saveToFileCalled {
		t.Errorf("Expected SaveToFile to be called, but it wasn't")
	}

	// Verify audit log entries - note that we have 4 entries now due to the provider-agnostic implementation
	expectedOperations := []string{
		"GenerateContentStart",
		"GenerateContentEnd",
		"SaveOutputStart",
		"SaveOutputEnd",
	}

	// Verify we have at least the expected number of operations (some implementation details may change)
	if len(auditEntries) < len(expectedOperations) {
		t.Errorf("Expected at least %d audit entries, got %d", len(expectedOperations), len(auditEntries))
		return
	}

	// Verify all required operations exist, but we allow for additional operations to be present
	operationExists := make(map[string]bool)
	for _, entry := range auditEntries {
		operationExists[entry.Operation] = true
	}

	for _, requiredOperation := range expectedOperations {
		if !operationExists[requiredOperation] {
			t.Errorf("Required audit operation '%s' was not found in log entries", requiredOperation)
		}
	}
}
