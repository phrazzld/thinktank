// Package integration provides integration tests for the thinktank package
package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/misty-step/thinktank/internal/llm"
)

// TestBoundaryMocking verifies that we can properly mock external boundaries
// This test serves as a proof of concept for the T002 task
func TestBoundaryMocking(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "boundary-mock-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to clean up temp dir: %v", err)
		}
	}()

	// Create an external API caller mock
	apiCaller := &MockExternalAPICaller{
		CallLLMAPIFunc: func(ctx context.Context, modelName, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
			return &llm.ProviderResult{
				Content:      "Response from boundary mock for " + modelName,
				FinishReason: "stop",
			}, nil
		},
	}

	// Test calling the mock API
	result, err := apiCaller.CallLLMAPI(context.Background(), "test-model", "test prompt", nil)
	if err != nil {
		t.Fatalf("Failed to call mock API: %v", err)
	}

	expectedOutput := "Response from boundary mock for test-model"
	if result.Content != expectedOutput {
		t.Errorf("Unexpected API response: got %q, want %q", result.Content, expectedOutput)
	}

	// Create a filesystem mock
	filesystem := NewMockFilesystemIO()

	// Initialize the mock filesystem
	if err := filesystem.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("Failed to create directory in mock filesystem: %v", err)
	}
	if err := filesystem.MkdirAll(filepath.Join(tempDir, "output"), 0755); err != nil {
		t.Fatalf("Failed to create output directory in mock filesystem: %v", err)
	}
	instructionsPath := filepath.Join(tempDir, "instructions.txt")
	if err := filesystem.WriteFile(instructionsPath, []byte("Test instructions"), 0640); err != nil {
		t.Fatalf("Failed to write file in mock filesystem: %v", err)
	}

	// Verify file exists in mock filesystem
	exists, err := filesystem.Stat(instructionsPath)
	if err != nil || !exists {
		t.Fatalf("Failed to create file in mock filesystem")
	}

	// Verify the filesystem read works
	content, err := filesystem.ReadFile(instructionsPath)
	if err != nil {
		t.Fatalf("Failed to read from mock filesystem: %v", err)
	}

	if string(content) != "Test instructions" {
		t.Errorf("Incorrect content in mock filesystem: %q", string(content))
	}

	// Test successful - boundary mocking works for external boundaries
	t.Logf("Successfully used external boundary mocks")
}
