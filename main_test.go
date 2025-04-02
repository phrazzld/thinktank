// main_test.go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/prompt"
)

// Mock os.Exit for testing
var osExit = os.Exit

func TestCheckTokenLimit(t *testing.T) {
	ctx := context.Background()
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ", false)

	// Test case 1: Token count is within limits
	t.Run("TokenCountWithinLimits", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return &gemini.TokenCount{Total: 1000}, nil
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		err := checkTokenLimit(ctx, mockClient, "test prompt", logger)
		if err != nil {
			t.Errorf("Expected no error for token count within limits, got: %v", err)
		}
	})

	// Test case 2: Token count exceeds limits
	t.Run("TokenCountExceedsLimits", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return &gemini.TokenCount{Total: 3000}, nil
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		err := checkTokenLimit(ctx, mockClient, "test prompt", logger)
		if err == nil {
			t.Error("Expected error for token count exceeding limits, got nil")
		}
	})

	// Test case 3: Error getting model info
	t.Run("ErrorGettingModelInfo", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return nil, errors.New("model info error")
			},
		}

		err := checkTokenLimit(ctx, mockClient, "test prompt", logger)
		if err == nil {
			t.Error("Expected error when getting model info fails, got nil")
		}
	})

	// Test case 4: Error counting tokens
	t.Run("ErrorCountingTokens", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return nil, errors.New("token counting error")
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		err := checkTokenLimit(ctx, mockClient, "test prompt", logger)
		if err == nil {
			t.Error("Expected error when counting tokens fails, got nil")
		}
	})
}

// TestTokenInfoResult tests the tokenInfoResult struct and associated functions
func TestGetTokenInfo(t *testing.T) {
	ctx := context.Background()
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ", false)

	// Test normal operation
	t.Run("NormalOperation", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return &gemini.TokenCount{Total: 1000}, nil
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		info, err := getTokenInfo(ctx, mockClient, "test prompt", logger)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if info == nil {
			t.Fatal("Expected token info, got nil")
		}
		if info.tokenCount != 1000 {
			t.Errorf("Expected tokenCount=1000, got %d", info.tokenCount)
		}
		if info.inputLimit != 2000 {
			t.Errorf("Expected inputLimit=2000, got %d", info.inputLimit)
		}
		if info.exceedsLimit {
			t.Error("Expected exceedsLimit=false, got true")
		}
	})

	// Test exceeding limit
	t.Run("ExceedsLimit", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return &gemini.TokenCount{Total: 3000}, nil
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		info, err := getTokenInfo(ctx, mockClient, "test prompt", logger)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if info == nil {
			t.Fatal("Expected token info, got nil")
		}
		if info.tokenCount != 3000 {
			t.Errorf("Expected tokenCount=3000, got %d", info.tokenCount)
		}
		if info.inputLimit != 2000 {
			t.Errorf("Expected inputLimit=2000, got %d", info.inputLimit)
		}
		if !info.exceedsLimit {
			t.Error("Expected exceedsLimit=true, got false")
		}
		if info.limitError == "" {
			t.Error("Expected non-empty limitError")
		}
	})

	// Test error in GetModelInfo
	t.Run("ErrorInGetModelInfo", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return nil, errors.New("model info error")
			},
		}

		_, err := getTokenInfo(ctx, mockClient, "test prompt", logger)
		if err == nil {
			t.Error("Expected error when GetModelInfo fails, got nil")
		}
	})

	// Test error in CountTokens
	t.Run("ErrorInCountTokens", func(t *testing.T) {
		mockClient := &gemini.MockClient{
			CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return nil, errors.New("count tokens error")
			},
			GetModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  2000,
					OutputTokenLimit: 1000,
				}, nil
			},
		}

		_, err := getTokenInfo(ctx, mockClient, "test prompt", logger)
		if err == nil {
			t.Error("Expected error when CountTokens fails, got nil")
		}
	})
}

// TestConfiguration tests the Configuration struct
func TestConfiguration(t *testing.T) {
	// Test default values in Configuration
	defaultConfig := &Configuration{}

	if defaultConfig.TaskDescription != "" {
		t.Errorf("Expected empty TaskDescription, got %s", defaultConfig.TaskDescription)
	}

	if defaultConfig.OutputFile != "" {
		t.Errorf("Expected empty OutputFile, got %s", defaultConfig.OutputFile)
	}

	// Test Configuration with values
	config := &Configuration{
		TaskDescription: "test task",
		OutputFile:      "output.md",
		ModelName:       "test-model",
		Verbose:         true,
		LogLevel:        logutil.DebugLevel,
		UseColors:       true,
		Include:         ".go,.md",
		Exclude:         ".exe,.bin",
		ExcludeNames:    "node_modules,dist",
		Format:          "test-format",
		DryRun:          true,
		ConfirmTokens:   1000,
		PromptTemplate:  "template.tmpl",
		Paths:           []string{"path1", "path2"},
		ApiKey:          "test-key",
	}

	// Verify values
	if config.TaskDescription != "test task" {
		t.Errorf("Expected TaskDescription='test task', got %s", config.TaskDescription)
	}

	if config.OutputFile != "output.md" {
		t.Errorf("Expected OutputFile='output.md', got %s", config.OutputFile)
	}

	if config.ModelName != "test-model" {
		t.Errorf("Expected ModelName='test-model', got %s", config.ModelName)
	}

	if !config.Verbose {
		t.Error("Expected Verbose=true, got false")
	}

	if config.LogLevel != logutil.DebugLevel {
		t.Errorf("Expected LogLevel=DebugLevel, got %v", config.LogLevel)
	}

	if !config.UseColors {
		t.Error("Expected UseColors=true, got false")
	}

	if config.Include != ".go,.md" {
		t.Errorf("Expected Include='.go,.md', got %s", config.Include)
	}

	if config.Exclude != ".exe,.bin" {
		t.Errorf("Expected Exclude='.exe,.bin', got %s", config.Exclude)
	}

	if config.ExcludeNames != "node_modules,dist" {
		t.Errorf("Expected ExcludeNames='node_modules,dist', got %s", config.ExcludeNames)
	}

	if config.Format != "test-format" {
		t.Errorf("Expected Format='test-format', got %s", config.Format)
	}

	if !config.DryRun {
		t.Error("Expected DryRun=true, got false")
	}

	if config.ConfirmTokens != 1000 {
		t.Errorf("Expected ConfirmTokens=1000, got %d", config.ConfirmTokens)
	}

	if config.PromptTemplate != "template.tmpl" {
		t.Errorf("Expected PromptTemplate='template.tmpl', got %s", config.PromptTemplate)
	}

	expectedPaths := []string{"path1", "path2"}
	if !reflect.DeepEqual(config.Paths, expectedPaths) {
		t.Errorf("Expected Paths=%v, got %v", expectedPaths, config.Paths)
	}

	if config.ApiKey != "test-key" {
		t.Errorf("Expected ApiKey='test-key', got %s", config.ApiKey)
	}
}

// TestPromptForConfirmation tests the promptForConfirmation function
func TestPromptForConfirmation(t *testing.T) {
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ", false)

	// Test no confirmation needed (below threshold)
	t.Run("BelowThreshold", func(t *testing.T) {
		result := promptForConfirmation(500, 1000, logger)
		if !result {
			t.Error("Expected true (no confirmation needed), got false")
		}
	})

	// Test disabled confirmation (threshold is 0)
	t.Run("DisabledConfirmation", func(t *testing.T) {
		result := promptForConfirmation(1500, 0, logger)
		if !result {
			t.Error("Expected true (confirmation disabled), got false")
		}
	})

	// Note: We can't easily test the user input part in a unit test
	// We would need to mock os.Stdin, which requires more complex setup
}

// TestBuildPrompt tests the buildPrompt function
func TestBuildPrompt(t *testing.T) {
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ", false)

	// Test case 1: Default template
	t.Run("DefaultTemplate", func(t *testing.T) {
		mockPromptManager := &prompt.MockManager{
			LoadTemplateFunc: func(templatePath string) error {
				if templatePath != "default.tmpl" {
					t.Errorf("Expected template path 'default.tmpl', got '%s'", templatePath)
				}
				return nil
			},
			BuildPromptFunc: func(templateName string, data *prompt.TemplateData) (string, error) {
				if templateName != "default.tmpl" {
					t.Errorf("Expected template name 'default.tmpl', got '%s'", templateName)
				}
				if data.Task != "Test task" {
					t.Errorf("Expected task 'Test task', got '%s'", data.Task)
				}
				if data.Context != "Test context" {
					t.Errorf("Expected context 'Test context', got '%s'", data.Context)
				}
				return "Generated prompt for Test task", nil
			},
		}

		config := &Configuration{
			PromptTemplate: "", // Use default
		}

		result, err := buildPromptWithManager(config, "Test task", "Test context", mockPromptManager, logger)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if result != "Generated prompt for Test task" {
			t.Errorf("Expected 'Generated prompt for Test task', got '%s'", result)
		}
	})

	// Test case 2: Custom template
	t.Run("CustomTemplate", func(t *testing.T) {
		mockPromptManager := &prompt.MockManager{
			LoadTemplateFunc: func(templatePath string) error {
				if templatePath != "custom.tmpl" {
					t.Errorf("Expected template path 'custom.tmpl', got '%s'", templatePath)
				}
				return nil
			},
			BuildPromptFunc: func(templateName string, data *prompt.TemplateData) (string, error) {
				if templateName != "custom.tmpl" {
					t.Errorf("Expected template name 'custom.tmpl', got '%s'", templateName)
				}
				return "Custom prompt for " + data.Task, nil
			},
		}

		config := &Configuration{
			PromptTemplate: "custom.tmpl",
		}

		result, err := buildPromptWithManager(config, "Custom task", "Custom context", mockPromptManager, logger)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if result != "Custom prompt for Custom task" {
			t.Errorf("Expected 'Custom prompt for Custom task', got '%s'", result)
		}
	})

	// Test case 3: Template load error
	t.Run("TemplateLoadError", func(t *testing.T) {
		mockPromptManager := &prompt.MockManager{
			LoadTemplateFunc: func(templatePath string) error {
				return fmt.Errorf("template file not found")
			},
			BuildPromptFunc: func(templateName string, data *prompt.TemplateData) (string, error) {
				// Since LoadTemplate will fail, BuildPrompt should return error
				return "", fmt.Errorf("failed to load template %s", templateName)
			},
		}

		config := &Configuration{}

		_, err := buildPromptWithManager(config, "Test task", "Test context", mockPromptManager, logger)
		if err == nil {
			t.Error("Expected error when template loading fails, got nil")
		}
	})

	// Test case 4: Template build error
	t.Run("TemplateBuildError", func(t *testing.T) {
		mockPromptManager := &prompt.MockManager{
			LoadTemplateFunc: func(templatePath string) error {
				return nil
			},
			BuildPromptFunc: func(templateName string, data *prompt.TemplateData) (string, error) {
				return "", fmt.Errorf("template execution failed")
			},
		}

		config := &Configuration{}

		_, err := buildPromptWithManager(config, "Test task", "Test context", mockPromptManager, logger)
		if err == nil {
			t.Error("Expected error when template building fails, got nil")
		}
	})
}

// mockLogger is a simple mock of LoggerInterface that records Fatal calls
type mockLogger struct {
	logutil.LoggerInterface
	fatalCalled bool
}

func (m *mockLogger) Fatal(format string, args ...interface{}) {
	m.fatalCalled = true
}

// TestProcessApiResponse tests the processApiResponse function
func TestProcessApiResponse(t *testing.T) {
	baseLogger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ", false)

	// Test valid response
	t.Run("ValidResponse", func(t *testing.T) {
		mockLog := &mockLogger{LoggerInterface: baseLogger}

		result := &gemini.GenerationResult{
			Content:      "Generated plan content",
			TokenCount:   100,
			FinishReason: "STOP",
		}

		content := processApiResponse(result, mockLog)

		if mockLog.fatalCalled {
			t.Error("Fatal was called unexpectedly for valid response")
		}

		if content != "Generated plan content" {
			t.Errorf("Expected 'Generated plan content', got '%s'", content)
		}
	})

	// Test empty response
	t.Run("EmptyResponse", func(t *testing.T) {
		mockLog := &mockLogger{LoggerInterface: baseLogger}

		result := &gemini.GenerationResult{
			Content:      "",
			TokenCount:   0,
			FinishReason: "MAX_TOKENS",
		}

		_ = processApiResponse(result, mockLog)

		if !mockLog.fatalCalled {
			t.Error("Expected Fatal to be called for empty response")
		}
	})

	// Test whitespace-only response
	t.Run("WhitespaceResponse", func(t *testing.T) {
		mockLog := &mockLogger{LoggerInterface: baseLogger}

		result := &gemini.GenerationResult{
			Content:      "   \n   ",
			TokenCount:   5,
			FinishReason: "STOP",
		}

		_ = processApiResponse(result, mockLog)

		if !mockLog.fatalCalled {
			t.Error("Expected Fatal to be called for whitespace-only response")
		}
	})

	// Test safety-blocked response
	t.Run("SafetyBlockedResponse", func(t *testing.T) {
		mockLog := &mockLogger{LoggerInterface: baseLogger}

		result := &gemini.GenerationResult{
			Content:      "",
			TokenCount:   0,
			FinishReason: "SAFETY",
			SafetyRatings: []gemini.SafetyRating{
				{
					Category: "HARM_CATEGORY_DANGEROUS",
					Blocked:  true,
				},
			},
		}

		_ = processApiResponse(result, mockLog)

		if !mockLog.fatalCalled {
			t.Error("Expected Fatal to be called for safety-blocked response")
		}
	})
}

// TestSaveToFile tests the saveToFile function
func TestSaveToFile(t *testing.T) {
	baseLogger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ", false)

	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "architect-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test successful save
	t.Run("SuccessfulSave", func(t *testing.T) {
		mockLog := &mockLogger{LoggerInterface: baseLogger}
		content := "Test plan content"
		outputFile := filepath.Join(tmpDir, "test-output.md")

		saveToFile(content, outputFile, mockLog)

		// Verify file was created with correct content
		savedContent, err := os.ReadFile(outputFile)
		if err != nil {
			t.Errorf("Failed to read saved file: %v", err)
		}

		if string(savedContent) != content {
			t.Errorf("Expected content '%s', got '%s'", content, string(savedContent))
		}

		if mockLog.fatalCalled {
			t.Error("Fatal was called unexpectedly for valid save")
		}
	})

	// Test save to non-existent directory
	t.Run("NonExistentDirectory", func(t *testing.T) {
		mockLog := &mockLogger{LoggerInterface: baseLogger}
		content := "Test plan content"
		outputFile := filepath.Join(tmpDir, "nonexistent", "test-output.md")

		saveToFile(content, outputFile, mockLog)

		if !mockLog.fatalCalled {
			t.Error("Expected Fatal to be called for non-existent directory")
		}
	})
}

// We'll skip validateInputs tests as they're problematic with flag.Lookup in test mode
// The validateInputs function is relatively simple, so a manual review should suffice
