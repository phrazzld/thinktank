package compat_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/architect/compat"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
)

// mockLogger implements the logutil.LoggerInterface for testing
type mockLogger struct{}

var _ logutil.LoggerInterface = (*mockLogger)(nil) // Static check for interface implementation

func (m *mockLogger) Debug(format string, args ...interface{})  {}
func (m *mockLogger) Info(format string, args ...interface{})   {}
func (m *mockLogger) Warn(format string, args ...interface{})   {}
func (m *mockLogger) Error(format string, args ...interface{})  {}
func (m *mockLogger) Fatal(format string, args ...interface{})  {}
func (m *mockLogger) Println(args ...interface{})               {}
func (m *mockLogger) Printf(format string, args ...interface{}) {}

// mockLLMClient mocks llm.LLMClient for testing
type mockLLMClient struct {
	modelName   string
	genResponse *llm.ProviderResult
	genErr      error
	countResult *llm.ProviderTokenCount
	countErr    error
	modelInfo   *llm.ProviderModelInfo
	modelErr    error
	closeErr    error
}

func (m *mockLLMClient) GenerateContent(ctx context.Context, prompt string) (*llm.ProviderResult, error) {
	if m.genErr != nil {
		return nil, m.genErr
	}
	if m.genResponse != nil {
		return m.genResponse, nil
	}
	return &llm.ProviderResult{Content: "mock content for " + prompt}, nil
}

func (m *mockLLMClient) CountTokens(ctx context.Context, prompt string) (*llm.ProviderTokenCount, error) {
	if m.countErr != nil {
		return nil, m.countErr
	}
	if m.countResult != nil {
		return m.countResult, nil
	}
	return &llm.ProviderTokenCount{Total: int32(len(prompt))}, nil
}

func (m *mockLLMClient) GetModelInfo(ctx context.Context) (*llm.ProviderModelInfo, error) {
	if m.modelErr != nil {
		return nil, m.modelErr
	}
	if m.modelInfo != nil {
		return m.modelInfo, nil
	}
	return &llm.ProviderModelInfo{
		Name:             m.modelName,
		InputTokenLimit:  1000,
		OutputTokenLimit: 500,
	}, nil
}

func (m *mockLLMClient) GetModelName() string {
	return m.modelName
}

func (m *mockLLMClient) Close() error {
	return m.closeErr
}

// TestInitClient tests the InitClient compatibility function
func TestInitClient(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	// Define test cases
	tests := []struct {
		name           string
		apiKey         string
		modelName      string
		apiEndpoint    string
		createClientFn func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error)
		wantErr        bool
		errContains    string
	}{
		{
			name:      "Basic successful case",
			apiKey:    "test-key",
			modelName: "gemini-1.5-pro",
			createClientFn: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
				return &mockLLMClient{modelName: modelName}, nil
			},
			wantErr: false,
		},
		{
			name:      "Non-gemini model should fail",
			apiKey:    "test-key",
			modelName: "gpt-4",
			createClientFn: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
				return &mockLLMClient{modelName: modelName}, nil
			},
			wantErr:     true,
			errContains: "only supports Gemini models",
		},
		{
			name:      "LLMClient creation error",
			apiKey:    "test-key",
			modelName: "gemini-1.5-pro",
			createClientFn: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
				return nil, errors.New("client creation failed")
			},
			wantErr:     true,
			errContains: "client creation failed",
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := compat.InitClient(ctx, tc.apiKey, tc.modelName, tc.apiEndpoint, tc.createClientFn, logger)

			// Check error case
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Expected error but got nil")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("Error doesn't contain expected text. Want: %q, Got: %q", tc.errContains, err.Error())
				}
				return
			}

			// Check successful case
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if client == nil {
				t.Errorf("Expected non-nil client but got nil")
			}
		})
	}
}

// TestProcessResponse tests the ProcessResponse compatibility function
func TestProcessResponse(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		result      *gemini.GenerationResult
		processFn   func(result *llm.ProviderResult) (string, error)
		wantContent string
		wantErr     bool
		errContains string
	}{
		{
			name: "Basic successful case",
			result: &gemini.GenerationResult{
				Content: "Test content",
			},
			processFn: func(result *llm.ProviderResult) (string, error) {
				return result.Content, nil
			},
			wantContent: "Test content",
		},
		{
			name:   "Nil result",
			result: nil,
			processFn: func(result *llm.ProviderResult) (string, error) {
				return "", errors.New("should not be called")
			},
			wantErr:     true,
			errContains: "result is nil",
		},
		{
			name: "Processing error",
			result: &gemini.GenerationResult{
				Content: "Test content",
			},
			processFn: func(result *llm.ProviderResult) (string, error) {
				return "", errors.New("processing failed")
			},
			wantErr:     true,
			errContains: "processing failed",
		},
		{
			name: "Safety ratings conversion",
			result: &gemini.GenerationResult{
				Content: "Test content",
				SafetyRatings: []gemini.SafetyRating{
					{Category: "test-category", Blocked: true, Score: 0.9},
				},
			},
			processFn: func(result *llm.ProviderResult) (string, error) {
				// Verify safety info was correctly converted
				if len(result.SafetyInfo) != 1 {
					t.Errorf("Expected 1 safety info, got %d", len(result.SafetyInfo))
				}
				if result.SafetyInfo[0].Category != "test-category" || !result.SafetyInfo[0].Blocked {
					t.Errorf("Safety info not correctly converted")
				}
				return result.Content, nil
			},
			wantContent: "Test content",
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			content, err := compat.ProcessResponse(tc.result, tc.processFn)

			// Check error case
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Expected error but got nil")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("Error doesn't contain expected text. Want: %q, Got: %q", tc.errContains, err.Error())
				}
				return
			}

			// Check successful case
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if content != tc.wantContent {
				t.Errorf("Content mismatch. Want: %q, Got: %q", tc.wantContent, content)
			}
		})
	}
}

// TestAdapterMethods tests the methods of the llmToGeminiClientAdapter
func TestAdapterMethods(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	t.Run("GenerateContent", func(t *testing.T) {
		// Setup
		mockClient := &mockLLMClient{
			modelName: "gemini-test",
			genResponse: &llm.ProviderResult{
				Content:      "Generated content",
				FinishReason: "FINISHED",
				TokenCount:   42,
				Truncated:    false,
			},
		}
		adapter := compat.NewLLMToGeminiClientAdapter(mockClient, logger)

		// Execute
		result, err := adapter.GenerateContent(ctx, "test prompt")

		// Verify
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		expected := "Generated content"
		if result.Content != expected {
			t.Errorf("Content mismatch. Want: %q, Got: %q", expected, result.Content)
		}
		if result.TokenCount != 42 {
			t.Errorf("TokenCount mismatch. Want: 42, Got: %d", result.TokenCount)
		}
		if result.FinishReason != "FINISHED" {
			t.Errorf("FinishReason mismatch. Want: FINISHED, Got: %s", result.FinishReason)
		}
	})

	t.Run("GenerateContent Error", func(t *testing.T) {
		// Setup
		mockClient := &mockLLMClient{
			modelName: "gemini-test",
			genErr:    errors.New("generation failed"),
		}
		adapter := compat.NewLLMToGeminiClientAdapter(mockClient, logger)

		// Execute
		_, err := adapter.GenerateContent(ctx, "test prompt")

		// Verify
		if err == nil {
			t.Fatalf("Expected error but got nil")
		}
		if !strings.Contains(err.Error(), "generation failed") {
			t.Errorf("Error doesn't contain expected text. Want: 'generation failed', Got: %q", err.Error())
		}
	})

	t.Run("Default methods", func(t *testing.T) {
		// These methods just return default values and log that they were called
		mockClient := &mockLLMClient{modelName: "gemini-test"}
		adapter := compat.NewLLMToGeminiClientAdapter(mockClient, logger)

		// Exercise methods
		temp := adapter.GetTemperature()
		if temp != 0.7 {
			t.Errorf("Expected default temperature 0.7, got %f", temp)
		}

		maxTokens := adapter.GetMaxOutputTokens()
		if maxTokens != 1024 {
			t.Errorf("Expected default max tokens 1024, got %d", maxTokens)
		}

		topP := adapter.GetTopP()
		if topP != 0.95 {
			t.Errorf("Expected default topP 0.95, got %f", topP)
		}

		// Logging cannot be tested with the simple mock logger, but the methods should not crash
	})

	t.Run("CountTokens", func(t *testing.T) {
		// Setup
		mockClient := &mockLLMClient{
			modelName:   "gemini-test",
			countResult: &llm.ProviderTokenCount{Total: 100},
		}
		adapter := compat.NewLLMToGeminiClientAdapter(mockClient, logger)

		// Execute
		result, err := adapter.CountTokens(ctx, "test prompt")

		// Verify
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Total != 100 {
			t.Errorf("Total tokens mismatch. Want: 100, Got: %d", result.Total)
		}
	})

	t.Run("GetModelInfo", func(t *testing.T) {
		// Setup
		mockClient := &mockLLMClient{
			modelName: "gemini-test",
			modelInfo: &llm.ProviderModelInfo{
				Name:             "gemini-test",
				InputTokenLimit:  4000,
				OutputTokenLimit: 2000,
			},
		}
		adapter := compat.NewLLMToGeminiClientAdapter(mockClient, logger)

		// Execute
		info, err := adapter.GetModelInfo(ctx)

		// Verify
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if info.Name != "gemini-test" {
			t.Errorf("Model name mismatch. Want: gemini-test, Got: %s", info.Name)
		}
		if info.InputTokenLimit != 4000 {
			t.Errorf("InputTokenLimit mismatch. Want: 4000, Got: %d", info.InputTokenLimit)
		}
		if info.OutputTokenLimit != 2000 {
			t.Errorf("OutputTokenLimit mismatch. Want: 2000, Got: %d", info.OutputTokenLimit)
		}
	})

	t.Run("GetModelName", func(t *testing.T) {
		// Setup
		mockClient := &mockLLMClient{modelName: "gemini-test"}
		adapter := compat.NewLLMToGeminiClientAdapter(mockClient, logger)

		// Execute & Verify
		if adapter.GetModelName() != "gemini-test" {
			t.Errorf("Model name mismatch. Want: gemini-test, Got: %s", adapter.GetModelName())
		}
	})

	t.Run("Close", func(t *testing.T) {
		// Setup
		mockClient := &mockLLMClient{closeErr: nil}
		adapter := compat.NewLLMToGeminiClientAdapter(mockClient, logger)

		// Execute & Verify
		if err := adapter.Close(); err != nil {
			t.Errorf("Unexpected error on Close: %v", err)
		}
	})

	t.Run("Close with error", func(t *testing.T) {
		// Setup
		mockClient := &mockLLMClient{closeErr: errors.New("close failed")}
		adapter := compat.NewLLMToGeminiClientAdapter(mockClient, logger)

		// Execute & Verify
		if err := adapter.Close(); err == nil {
			t.Error("Expected error on Close but got nil")
		} else if !strings.Contains(err.Error(), "close failed") {
			t.Errorf("Error doesn't contain expected text. Want: 'close failed', Got: %q", err.Error())
		}
	})
}
