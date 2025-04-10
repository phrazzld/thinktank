// Package architect provides the command-line interface for the architect tool
package architect

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/prompt"
)

// outputTokenManager implements the TokenManager interface for testing in output_test.go
type outputTokenManager struct {
	getTokenInfoFunc          func(ctx context.Context, client gemini.Client, prompt string) (*TokenResult, error)
	checkTokenLimitFunc       func(ctx context.Context, client gemini.Client, prompt string) error
	promptForConfirmationFunc func(tokenCount int32, confirmTokens int) bool
}

func newOutputTokenManager() *outputTokenManager {
	return &outputTokenManager{
		getTokenInfoFunc: func(ctx context.Context, client gemini.Client, prompt string) (*TokenResult, error) {
			return &TokenResult{
				TokenCount:   100,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   10.0,
			}, nil
		},
		checkTokenLimitFunc: func(ctx context.Context, client gemini.Client, prompt string) error {
			return nil
		},
		promptForConfirmationFunc: func(tokenCount int32, confirmTokens int) bool {
			return true
		},
	}
}

func (m *outputTokenManager) GetTokenInfo(ctx context.Context, client gemini.Client, prompt string) (*TokenResult, error) {
	return m.getTokenInfoFunc(ctx, client, prompt)
}

func (m *outputTokenManager) CheckTokenLimit(ctx context.Context, client gemini.Client, prompt string) error {
	return m.checkTokenLimitFunc(ctx, client, prompt)
}

func (m *outputTokenManager) PromptForConfirmation(tokenCount int32, confirmTokens int) bool {
	return m.promptForConfirmationFunc(tokenCount, confirmTokens)
}

// mockPromptManager implements the prompt.ManagerInterface for testing
type mockPromptManager struct {
	buildPromptFunc          func(templateName string, data *prompt.TemplateData) (string, error)
	listExampleTemplatesFunc func() ([]string, error)
	getExampleTemplateFunc   func(name string) (string, error)
}

func newMockPromptManager() *mockPromptManager {
	return &mockPromptManager{
		buildPromptFunc: func(templateName string, data *prompt.TemplateData) (string, error) {
			return "mock prompt", nil
		},
		listExampleTemplatesFunc: func() ([]string, error) {
			return []string{"example.tmpl"}, nil
		},
		getExampleTemplateFunc: func(name string) (string, error) {
			return "example template content", nil
		},
	}
}

func (m *mockPromptManager) BuildPrompt(templateName string, data *prompt.TemplateData) (string, error) {
	return m.buildPromptFunc(templateName, data)
}

func (m *mockPromptManager) ListExampleTemplates() ([]string, error) {
	return m.listExampleTemplatesFunc()
}

func (m *mockPromptManager) GetExampleTemplate(name string) (string, error) {
	return m.getExampleTemplateFunc(name)
}

// outputGeminiClient implements a simplified gemini.Client for testing
type outputGeminiClient struct {
	generateContentFunc func(ctx context.Context, prompt string) (*gemini.GenerationResult, error)
	getModelInfoFunc    func(ctx context.Context) (*gemini.ModelInfo, error)
	countTokensFunc     func(ctx context.Context, prompt string) (*gemini.TokenCount, error)
	closeFunc           func()
}

func newOutputGeminiClient() *outputGeminiClient {
	return &outputGeminiClient{
		generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
			return &gemini.GenerationResult{
				Content:    "generated content",
				TokenCount: 50,
			}, nil
		},
		getModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
			return &gemini.ModelInfo{
				InputTokenLimit: 1000,
			}, nil
		},
		countTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
			return &gemini.TokenCount{Total: 100}, nil
		},
		closeFunc: func() {},
	}
}

func (m *outputGeminiClient) GenerateContent(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
	return m.generateContentFunc(ctx, prompt)
}

func (m *outputGeminiClient) GetModelInfo(ctx context.Context) (*gemini.ModelInfo, error) {
	return m.getModelInfoFunc(ctx)
}

func (m *outputGeminiClient) CountTokens(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
	return m.countTokensFunc(ctx, prompt)
}

func (m *outputGeminiClient) Close() {
	if m.closeFunc != nil {
		m.closeFunc()
	}
}

// mockConfigManager implements a simplified config.ManagerInterface for testing
type mockConfigManager struct {
	loadFromFilesFunc    func() error
	ensureConfigDirsFunc func() error
	getConfigFunc        func() *config.AppConfig
	mergeWithFlagsFunc   func(flags map[string]interface{}) error
	getConfigDirsFunc    func() config.ConfigDirectories
}

func newMockConfigManager() *mockConfigManager {
	return &mockConfigManager{
		loadFromFilesFunc: func() error {
			return nil
		},
		ensureConfigDirsFunc: func() error {
			return nil
		},
		getConfigFunc: func() *config.AppConfig {
			return &config.AppConfig{}
		},
		mergeWithFlagsFunc: func(flags map[string]interface{}) error {
			return nil
		},
		getConfigDirsFunc: func() config.ConfigDirectories {
			return config.ConfigDirectories{}
		},
	}
}

func (m *mockConfigManager) LoadFromFiles() error {
	return m.loadFromFilesFunc()
}

func (m *mockConfigManager) EnsureConfigDirs() error {
	return m.ensureConfigDirsFunc()
}

func (m *mockConfigManager) GetConfig() *config.AppConfig {
	return m.getConfigFunc()
}

func (m *mockConfigManager) MergeWithFlags(flags map[string]interface{}) error {
	return m.mergeWithFlagsFunc(flags)
}

func (m *mockConfigManager) GetConfigDirs() config.ConfigDirectories {
	return m.getConfigDirsFunc()
}

// mockAPIService implements a simplified APIService for testing
type mockAPIService struct {
	processResponseFunc      func(result *gemini.GenerationResult) (string, error)
	getErrorDetailsFunc      func(err error) string
	isEmptyResponseErrorFunc func(err error) bool
	isSafetyBlockedErrorFunc func(err error) bool
}

func newMockAPIService() *mockAPIService {
	return &mockAPIService{
		processResponseFunc: func(result *gemini.GenerationResult) (string, error) {
			if result == nil {
				return "", errors.New("empty response")
			}
			return result.Content, nil
		},
		getErrorDetailsFunc: func(err error) string {
			return "mock error details"
		},
		isEmptyResponseErrorFunc: func(err error) bool {
			return false
		},
		isSafetyBlockedErrorFunc: func(err error) bool {
			return false
		},
	}
}

// Override the default NewAPIService function for testing
var originalNewAPIService = NewAPIService

func (m *mockAPIService) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	return m.processResponseFunc(result)
}

func (m *mockAPIService) GetErrorDetails(err error) string {
	return m.getErrorDetailsFunc(err)
}

func (m *mockAPIService) IsEmptyResponseError(err error) bool {
	return m.isEmptyResponseErrorFunc(err)
}

func (m *mockAPIService) IsSafetyBlockedError(err error) bool {
	return m.isSafetyBlockedErrorFunc(err)
}

// TestSaveToFile tests the SaveToFile method
func TestSaveToFile(t *testing.T) {
	// Create a logger for testing
	logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ", false)

	// Create a token manager for testing
	tokenManager := newOutputTokenManager()

	// Create an output writer
	outputWriter := NewOutputWriter(logger, tokenManager)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "output_test")
	if err \!= nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file in a non-writable directory for error case
	readOnlyDir := filepath.Join(tempDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0500); err \!= nil { // 0500 = read-only directory
		t.Fatalf("Failed to create read-only directory: %v", err)
	}

	// Define test cases
	tests := []struct {
		name       string
		content    string
		outputFile string
		setupFunc  func() // Function to run before test
		cleanFunc  func() // Function to run after test
		wantErr    bool
	}{
		{
			name:       "Valid file path - absolute",
			content:    "Test content",
			outputFile: filepath.Join(tempDir, "test_output.md"),
			setupFunc:  func() {},
			cleanFunc:  func() {},
			wantErr:    false,
		},
		{
			name:       "Valid file path - relative",
			content:    "Test content with relative path",
			outputFile: "test_output_relative.md",
			setupFunc:  func() {},
			cleanFunc: func() {
				// Clean up relative path file
				cwd, _ := os.Getwd()
				os.Remove(filepath.Join(cwd, "test_output_relative.md"))
			},
			wantErr: false,
		},
		{
			name:       "Empty content",
			content:    "",
			outputFile: filepath.Join(tempDir, "empty_file.md"),
			setupFunc:  func() {},
			cleanFunc:  func() {},
			wantErr:    false,
		},
		{
			name:       "Long content",
			content:    strings.Repeat("Long content test ", 1000), // ~ 18KB of content
			outputFile: filepath.Join(tempDir, "long_file.md"),
			setupFunc:  func() {},
			cleanFunc:  func() {},
			wantErr:    false,
		},
		{
			name:       "Non-existent directory",
			content:    "Test content",
			outputFile: filepath.Join(tempDir, "non-existent", "test_output.md"),
			setupFunc:  func() {},
			cleanFunc:  func() {},
			wantErr:    true,
		},
	}

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Run setup function
			tc.setupFunc()

			// Save to file
			err := outputWriter.SaveToFile(tc.content, tc.outputFile)

			// Run cleanup function
			defer tc.cleanFunc()

			// Check error
			if (err \!= nil) \!= tc.wantErr {
				t.Errorf("SaveToFile() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// Skip file validation for expected errors
			if tc.wantErr {
				return
			}

			// Determine output path for validation
			outputPath := tc.outputFile
			if \!filepath.IsAbs(outputPath) {
				cwd, _ := os.Getwd()
				outputPath = filepath.Join(cwd, outputPath)
			}

			// Verify file was created and content matches
			content, err := os.ReadFile(outputPath)
			if err \!= nil {
				t.Errorf("Failed to read output file: %v", err)
				return
			}

			if string(content) \!= tc.content {
				t.Errorf("File content = %v, want %v", string(content), tc.content)
			}
		})
	}
}

// TestGenerateAndSavePlan tests the GenerateAndSavePlan method
func TestGenerateAndSavePlan(t *testing.T) {
	// Create a logger for testing
	logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ", false)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "output_test")
	if err \!= nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Define common test parameters
	ctx := context.Background()
	taskDescription := "Test task"
	projectContext := "Test project context"
	outputFile := filepath.Join(tempDir, "test_plan.md")

	// Test cases
	tests := []struct {
		name              string
		promptManagerFunc func() *mockPromptManager
		tokenManagerFunc  func() *outputTokenManager
		geminiClientFunc  func() *outputGeminiClient
		apiServiceFunc    func() *mockAPIService
		expectedContent   string
		wantErr           bool
		errorContains     string
	}{
		{
			name: "Happy path - generates content successfully",
			promptManagerFunc: func() *mockPromptManager {
				pm := newMockPromptManager()
				return pm
			},
			tokenManagerFunc: func() *outputTokenManager {
				tm := newOutputTokenManager()
				return tm
			},
			geminiClientFunc: func() *outputGeminiClient {
				gc := newOutputGeminiClient()
				return gc
			},
			apiServiceFunc: func() *mockAPIService {
				as := newMockAPIService()
				return as
			},
			expectedContent: "generated content",
			wantErr:         false,
		},
		{
			name: "Error - prompt building fails",
			promptManagerFunc: func() *mockPromptManager {
				pm := newMockPromptManager()
				pm.buildPromptFunc = func(templateName string, data *prompt.TemplateData) (string, error) {
					return "", errors.New("failed to build prompt")
				}
				return pm
			},
			tokenManagerFunc: func() *outputTokenManager {
				return newOutputTokenManager()
			},
			geminiClientFunc: func() *outputGeminiClient {
				return newOutputGeminiClient()
			},
			apiServiceFunc: func() *mockAPIService {
				return newMockAPIService()
			},
			wantErr:       true,
			errorContains: "failed to build prompt",
		},
		{
			name: "Error - token check fails",
			promptManagerFunc: func() *mockPromptManager {
				return newMockPromptManager()
			},
			tokenManagerFunc: func() *outputTokenManager {
				tm := newOutputTokenManager()
				tm.getTokenInfoFunc = func(ctx context.Context, client gemini.Client, prompt string) (*TokenResult, error) {
					return nil, errors.New("token count check failed")
				}
				return tm
			},
			geminiClientFunc: func() *outputGeminiClient {
				return newOutputGeminiClient()
			},
			apiServiceFunc: func() *mockAPIService {
				return newMockAPIService()
			},
			wantErr:       true,
			errorContains: "token count check failed",
		},
		{
			name: "Error - token limit exceeded",
			promptManagerFunc: func() *mockPromptManager {
				return newMockPromptManager()
			},
			tokenManagerFunc: func() *outputTokenManager {
				tm := newOutputTokenManager()
				tm.getTokenInfoFunc = func(ctx context.Context, client gemini.Client, prompt string) (*TokenResult, error) {
					return &TokenResult{
						TokenCount:   2000,
						InputLimit:   1000,
						ExceedsLimit: true,
						LimitError:   "token limit exceeded",
						Percentage:   200.0,
					}, nil
				}
				return tm
			},
			geminiClientFunc: func() *outputGeminiClient {
				return newOutputGeminiClient()
			},
			apiServiceFunc: func() *mockAPIService {
				return newMockAPIService()
			},
			wantErr:       true,
			errorContains: "token limit exceeded",
		},
		{
			name: "Error - content generation fails",
			promptManagerFunc: func() *mockPromptManager {
				return newMockPromptManager()
			},
			tokenManagerFunc: func() *outputTokenManager {
				return newOutputTokenManager()
			},
			geminiClientFunc: func() *outputGeminiClient {
				gc := newOutputGeminiClient()
				gc.generateContentFunc = func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					return nil, errors.New("failed to generate content")
				}
				return gc
			},
			apiServiceFunc: func() *mockAPIService {
				return newMockAPIService()
			},
			wantErr:       true,
			errorContains: "plan generation failed",
		},
		{
			name: "Error - API response processing fails",
			promptManagerFunc: func() *mockPromptManager {
				return newMockPromptManager()
			},
			tokenManagerFunc: func() *outputTokenManager {
				return newOutputTokenManager()
			},
			geminiClientFunc: func() *outputGeminiClient {
				return newOutputGeminiClient()
			},
			apiServiceFunc: func() *mockAPIService {
				as := newMockAPIService()
				as.processResponseFunc = func(result *gemini.GenerationResult) (string, error) {
					return "", errors.New("failed to process API response")
				}
				return as
			},
			wantErr:       true,
			errorContains: "failed to process API response",
		},
		{
			name: "Error - empty response",
			promptManagerFunc: func() *mockPromptManager {
				return newMockPromptManager()
			},
			tokenManagerFunc: func() *outputTokenManager {
				return newOutputTokenManager()
			},
			geminiClientFunc: func() *outputGeminiClient {
				return newOutputGeminiClient()
			},
			apiServiceFunc: func() *mockAPIService {
				as := newMockAPIService()
				as.processResponseFunc = func(result *gemini.GenerationResult) (string, error) {
					return "", errors.New("empty response")
				}
				as.isEmptyResponseErrorFunc = func(err error) bool {
					return true
				}
				return as
			},
			wantErr:       true,
			errorContains: "empty content",
		},
		{
			name: "Error - safety blocked",
			promptManagerFunc: func() *mockPromptManager {
				return newMockPromptManager()
			},
			tokenManagerFunc: func() *outputTokenManager {
				return newOutputTokenManager()
			},
			geminiClientFunc: func() *outputGeminiClient {
				return newOutputGeminiClient()
			},
			apiServiceFunc: func() *mockAPIService {
				as := newMockAPIService()
				as.processResponseFunc = func(result *gemini.GenerationResult) (string, error) {
					return "", errors.New("safety blocked")
				}
				as.isSafetyBlockedErrorFunc = func(err error) bool {
					return true
				}
				return as
			},
			wantErr:       true,
			errorContains: "safety restrictions",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create dependencies
			promptManager := tc.promptManagerFunc()
			tokenManager := tc.tokenManagerFunc()
			geminiClient := tc.geminiClientFunc()
			apiService := tc.apiServiceFunc()

			// Replace NewAPIService to return our mock
			NewAPIService = func(logger logutil.LoggerInterface) APIService {
				return apiService
			}
			defer func() {
				NewAPIService = originalNewAPIService
			}()

			// Create output writer
			outputWriter := NewOutputWriter(logger, tokenManager)

			// Create unique output file for this test
			testOutputFile := filepath.Join(tempDir, t.Name()+".md")

			// Call the method being tested
			err = outputWriter.GenerateAndSavePlan(ctx, geminiClient, taskDescription, projectContext, testOutputFile, promptManager)

			// Verify error behavior
			if tc.wantErr {
				if err == nil {
					t.Errorf("GenerateAndSavePlan() error = nil, expected error")
					return
				}
				if tc.errorContains \!= "" && \!strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("GenerateAndSavePlan() error = %v, expected to contain %v", err, tc.errorContains)
					return
				}
				return
			}

			// For non-error cases, verify the result
			if err \!= nil {
				t.Errorf("GenerateAndSavePlan() unexpected error = %v", err)
				return
			}

			// Verify the file was created
			if _, err := os.Stat(testOutputFile); os.IsNotExist(err) {
				t.Errorf("output file not created: %v", err)
				return
			}

			// Read the file content
			content, err := os.ReadFile(testOutputFile)
			if err \!= nil {
				t.Errorf("Failed to read output file: %v", err)
				return
			}

			// Verify content
			if string(content) \!= tc.expectedContent {
				t.Errorf("File content = %v, want %v", string(content), tc.expectedContent)
			}
		})
	}
}

// TestGenerateAndSavePlanWithConfig tests the GenerateAndSavePlanWithConfig method
func TestGenerateAndSavePlanWithConfig(t *testing.T) {
	// Create a logger for testing
	logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ", false)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "output_test")
	if err \!= nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Define common test parameters
	ctx := context.Background()
	taskDescription := "Test task"
	projectContext := "Test project context"

	// Test cases
	tests := []struct {
		name                   string
		configManagerFunc      func() *mockConfigManager
		tokenManagerFunc       func() *outputTokenManager
		geminiClientFunc       func() *outputGeminiClient
		apiServiceFunc         func() *mockAPIService
		setupPromptManagerFunc func() (prompt.ManagerInterface, error)
		expectedContent        string
		wantErr                bool
		errorContains          string
	}{
		{
			name: "Happy path - generates content successfully",
			configManagerFunc: func() *mockConfigManager {
				return newMockConfigManager()
			},
			tokenManagerFunc: func() *outputTokenManager {
				return newOutputTokenManager()
			},
			geminiClientFunc: func() *outputGeminiClient {
				return newOutputGeminiClient()
			},
			apiServiceFunc: func() *mockAPIService {
				return newMockAPIService()
			},
			setupPromptManagerFunc: func() (prompt.ManagerInterface, error) {
				return newMockPromptManager(), nil
			},
			expectedContent: "generated content",
			wantErr:         false,
		},
		{
			name: "Error - prompt manager setup fails but fallback succeeds",
			configManagerFunc: func() *mockConfigManager {
				return newMockConfigManager()
			},
			tokenManagerFunc: func() *outputTokenManager {
				return newOutputTokenManager()
			},
			geminiClientFunc: func() *outputGeminiClient {
				return newOutputGeminiClient()
			},
			apiServiceFunc: func() *mockAPIService {
				return newMockAPIService()
			},
			setupPromptManagerFunc: func() (prompt.ManagerInterface, error) {
				return nil, errors.New("failed to set up prompt manager")
			},
			expectedContent: "generated content", // Fallback should work
			wantErr:         false,
		},
		{
			name: "Error - content generation fails even with fallback",
			configManagerFunc: func() *mockConfigManager {
				return newMockConfigManager()
			},
			tokenManagerFunc: func() *outputTokenManager {
				return newOutputTokenManager()
			},
			geminiClientFunc: func() *outputGeminiClient {
				gc := newOutputGeminiClient()
				gc.generateContentFunc = func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					return nil, errors.New("generation failed")
				}
				return gc
			},
			apiServiceFunc: func() *mockAPIService {
				return newMockAPIService()
			},
			setupPromptManagerFunc: func() (prompt.ManagerInterface, error) {
				return nil, errors.New("failed to set up prompt manager")
			},
			wantErr:       true,
			errorContains: "generation failed",
		},
	}

	// Original prompt setup function
	originalSetupPromptManagerWithConfig := prompt.SetupPromptManagerWithConfig
	// Original NewManager function
	originalNewManager := prompt.NewManager

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create dependencies
			configManager := tc.configManagerFunc()
			tokenManager := tc.tokenManagerFunc()
			geminiClient := tc.geminiClientFunc()
			apiService := tc.apiServiceFunc()

			// Replace functions for testing
			NewAPIService = func(logger logutil.LoggerInterface) APIService {
				return apiService
			}
			defer func() {
				NewAPIService = originalNewAPIService
			}()

			// Override the prompt setup functions
			prompt.SetupPromptManagerWithConfig = func(logger logutil.LoggerInterface, configManager config.ManagerInterface) (prompt.ManagerInterface, error) {
				return tc.setupPromptManagerFunc()
			}
			defer func() {
				prompt.SetupPromptManagerWithConfig = originalSetupPromptManagerWithConfig
			}()

			// Override the prompt new manager function for fallback
			prompt.NewManager = func(logger logutil.LoggerInterface) prompt.ManagerInterface {
				return newMockPromptManager()
			}
			defer func() {
				prompt.NewManager = originalNewManager
			}()

			// Create output writer
			outputWriter := NewOutputWriter(logger, tokenManager)

			// Create unique output file for this test
			testOutputFile := filepath.Join(tempDir, t.Name()+".md")

			// Call the method being tested
			err = outputWriter.GenerateAndSavePlanWithConfig(ctx, geminiClient, taskDescription, projectContext, testOutputFile, configManager)

			// Verify error behavior
			if tc.wantErr {
				if err == nil {
					t.Errorf("GenerateAndSavePlanWithConfig() error = nil, expected error")
					return
				}
				if tc.errorContains \!= "" && \!strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("GenerateAndSavePlanWithConfig() error = %v, expected to contain %v", err, tc.errorContains)
					return
				}
				return
			}

			// For non-error cases, verify the result
			if err \!= nil {
				t.Errorf("GenerateAndSavePlanWithConfig() unexpected error = %v", err)
				return
			}

			// Verify the file was created
			if _, err := os.Stat(testOutputFile); os.IsNotExist(err) {
				t.Errorf("output file not created: %v", err)
				return
			}

			// Read the file content
			content, err := os.ReadFile(testOutputFile)
			if err \!= nil {
				t.Errorf("Failed to read output file: %v", err)
				return
			}

			// Verify content
			if string(content) \!= tc.expectedContent {
				t.Errorf("File content = %v, want %v", string(content), tc.expectedContent)
			}
		})
	}
}
