// internal/gemini/client_test.go
package gemini

import (
	"context"
	"testing"
)

func TestDefaultModelConfig(t *testing.T) {
	config := DefaultModelConfig()

	// Check default values
	if config.MaxOutputTokens != 8192 {
		t.Errorf("Expected MaxOutputTokens to be 8192, got %d", config.MaxOutputTokens)
	}

	if config.Temperature != 0.3 {
		t.Errorf("Expected Temperature to be 0.3, got %f", config.Temperature)
	}

	if config.TopP != 0.9 {
		t.Errorf("Expected TopP to be 0.9, got %f", config.TopP)
	}
}

func TestMockClient_GetModelInfo(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// Test with default mock implementation
	info, err := client.GetModelInfo(ctx)
	if err != nil {
		t.Fatalf("GetModelInfo returned unexpected error: %v", err)
	}

	if info == nil {
		t.Fatal("GetModelInfo returned nil info")
	}

	if info.Name != "mock-model" {
		t.Errorf("Expected model name %q, got %q", "mock-model", info.Name)
	}

	if info.InputTokenLimit != 32000 {
		t.Errorf("Expected input token limit %d, got %d", 32000, info.InputTokenLimit)
	}

	if info.OutputTokenLimit != 8192 {
		t.Errorf("Expected output token limit %d, got %d", 8192, info.OutputTokenLimit)
	}

	// Test with custom mock implementation
	customInfo := &ModelInfo{
		Name:             "custom-model",
		InputTokenLimit:  10000,
		OutputTokenLimit: 5000,
	}

	client.GetModelInfoFunc = func(ctx context.Context) (*ModelInfo, error) {
		return customInfo, nil
	}

	info, err = client.GetModelInfo(ctx)
	if err != nil {
		t.Fatalf("GetModelInfo with custom mock returned unexpected error: %v", err)
	}

	if info != customInfo {
		t.Errorf("Expected custom model info, got different instance")
	}
}

// Note: More extensive tests for the actual geminiClient would require mocking the HTTP client,
// which we'll implement in the integration tests (T3) task later
