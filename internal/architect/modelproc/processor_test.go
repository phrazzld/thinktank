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
	"github.com/phrazzld/architect/internal/gemini"
)

// Import necessary dependencies
// Note: Most mock definitions have been moved to mocks_test.go

func TestModelProcessor_Process_Success(t *testing.T) {
	// Setup mocks
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					return &gemini.GenerationResult{
						Content:    "Generated content",
						TokenCount: 50,
					}, nil
				},
			}, nil
		},
		processResponseFunc: func(result *gemini.GenerationResult) (string, error) {
			return result.Content, nil
		},
	}

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

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
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