package architect

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// MockLogger for testing
type mockContextLogger struct {
	logutil.LoggerInterface
	debugMessages []string
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func (m *mockContextLogger) Debug(format string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, format)
}

func (m *mockContextLogger) Info(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, format)
}

func (m *mockContextLogger) Warn(format string, args ...interface{}) {
	m.warnMessages = append(m.warnMessages, format)
}

func (m *mockContextLogger) Error(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, format)
}

func (m *mockContextLogger) Fatal(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, "FATAL: "+format)
}

func (m *mockContextLogger) Printf(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, format)
}

func (m *mockContextLogger) Println(v ...interface{}) {
	m.infoMessages = append(m.infoMessages, fmt.Sprint(v...))
}

// MockTokenManager for testing
type mockTokenManager struct {
	getTokenInfoFunc          func(ctx context.Context, client gemini.Client, prompt string) (*TokenResult, error)
	checkTokenLimitFunc       func(ctx context.Context, client gemini.Client, prompt string) error
	promptForConfirmationFunc func(tokenCount int32, threshold int) bool
}

func (m *mockTokenManager) GetTokenInfo(ctx context.Context, client gemini.Client, prompt string) (*TokenResult, error) {
	if m.getTokenInfoFunc != nil {
		return m.getTokenInfoFunc(ctx, client, prompt)
	}
	return &TokenResult{TokenCount: 100, InputLimit: 1000, Percentage: 10.0}, nil
}

func (m *mockTokenManager) CheckTokenLimit(ctx context.Context, client gemini.Client, prompt string) error {
	if m.checkTokenLimitFunc != nil {
		return m.checkTokenLimitFunc(ctx, client, prompt)
	}
	return nil
}

func (m *mockTokenManager) PromptForConfirmation(tokenCount int32, threshold int) bool {
	if m.promptForConfirmationFunc != nil {
		return m.promptForConfirmationFunc(tokenCount, threshold)
	}
	return true
}

// MockGeminiClient for testing
type mockGeminiClient struct {
	countTokensFunc     func(ctx context.Context, prompt string) (*gemini.TokenCount, error)
	generateContentFunc func(ctx context.Context, prompt string) (*gemini.GenerationResult, error)
	getModelInfoFunc    func(ctx context.Context) (*gemini.ModelInfo, error)
	closeFunc           func() error
}

func (m *mockGeminiClient) CountTokens(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
	if m.countTokensFunc != nil {
		return m.countTokensFunc(ctx, prompt)
	}
	return &gemini.TokenCount{Total: 100}, nil
}

func (m *mockGeminiClient) GenerateContent(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
	if m.generateContentFunc != nil {
		return m.generateContentFunc(ctx, prompt)
	}
	return &gemini.GenerationResult{Content: "test content"}, nil
}

func (m *mockGeminiClient) GetModelInfo(ctx context.Context) (*gemini.ModelInfo, error) {
	if m.getModelInfoFunc != nil {
		return m.getModelInfoFunc(ctx)
	}
	return &gemini.ModelInfo{
		Name:             "test-model",
		InputTokenLimit:  1000,
		OutputTokenLimit: 500,
	}, nil
}

func (m *mockGeminiClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// Helper function to create a temporary directory with test files
func createTestDirectory(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "context-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create a few test files
	files := map[string]string{
		"file1.go":          "package main\n\nfunc main() {\n}\n",
		"file2.txt":         "This is a text file\nWith multiple lines\n",
		"subdir/file3.md":   "# Markdown file\n\nWith some content\n",
		".hidden/file4.txt": "Hidden file content\n",
	}

	for path, content := range files {
		fullPath := filepath.Join(tempDir, path)
		dirPath := filepath.Dir(fullPath)

		// Create directory if it doesn't exist
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}

		// Write file
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to write test file %s: %v", fullPath, err)
		}
	}

	return tempDir, func() {
		os.RemoveAll(tempDir)
	}
}

// TestNewContextGatherer tests the constructor
func TestNewContextGatherer(t *testing.T) {
	logger := &mockContextLogger{}
	tokenManager := &mockTokenManager{}

	gatherer := NewContextGatherer(logger, true, tokenManager)
	if gatherer == nil {
		t.Error("Expected non-nil ContextGatherer, got nil")
	}
}

// TestGatherContext tests the GatherContext method
func TestGatherContext(t *testing.T) {
	// Create a temporary directory with test files
	tempDir, cleanup := createTestDirectory(t)
	defer cleanup()

	// Basic success test
	t.Run("BasicGathering", func(t *testing.T) {
		logger := &mockContextLogger{}
		tokenManager := &mockTokenManager{}
		client := &mockGeminiClient{
			countTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
				return &gemini.TokenCount{Total: 100}, nil
			},
		}

		gatherer := NewContextGatherer(logger, false, tokenManager)
		ctx := context.Background()

		config := GatherConfig{
			Paths:        []string{tempDir},
			Include:      "",
			Exclude:      "",
			ExcludeNames: "",
			Format:       "{path}\n{content}\n",
			Verbose:      true,
			LogLevel:     logutil.DebugLevel,
		}

		projectContext, stats, err := gatherer.GatherContext(ctx, client, config)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(projectContext) == 0 {
			t.Error("Expected non-empty project context, got empty slice")
		}

		if stats == nil {
			t.Error("Expected non-nil stats, got nil")
		} else {
			// Should have processed at least 3 files (excluding hidden)
			if stats.ProcessedFilesCount < 3 {
				t.Errorf("Expected at least 3 processed files, got %d", stats.ProcessedFilesCount)
			}
		}
	})

	// Test with file filtering
	t.Run("FileFiltering", func(t *testing.T) {
		logger := &mockContextLogger{}
		tokenManager := &mockTokenManager{}
		client := &mockGeminiClient{}

		gatherer := NewContextGatherer(logger, false, tokenManager)
		ctx := context.Background()

		config := GatherConfig{
			Paths:        []string{tempDir},
			Include:      ".go", // Only include Go files
			Exclude:      "",
			ExcludeNames: "",
			Format:       "{path}\n{content}\n",
			Verbose:      true,
			LogLevel:     logutil.DebugLevel,
		}

		_, stats, err := gatherer.GatherContext(ctx, client, config)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Should have processed exactly 1 Go file
		if stats.ProcessedFilesCount != 1 {
			t.Errorf("Expected 1 processed file, got %d", stats.ProcessedFilesCount)
		}
	})

	// Test handling of non-existent paths
	t.Run("FileAccessError", func(t *testing.T) {
		logger := &mockContextLogger{}
		tokenManager := &mockTokenManager{}
		client := &mockGeminiClient{}

		gatherer := NewContextGatherer(logger, false, tokenManager)
		ctx := context.Background()

		config := GatherConfig{
			Paths:        []string{"/this/path/does/not/exist"},
			Include:      "",
			Exclude:      "",
			ExcludeNames: "",
			Format:       "{path}\n{content}\n",
			Verbose:      true,
			LogLevel:     logutil.DebugLevel,
		}

		projectContext, stats, err := gatherer.GatherContext(ctx, client, config)

		// The refactored implementation returns an empty result with no error
		// when path doesn't exist (which is an acceptable behavior)
		// Verify that no files were processed
		if err != nil {
			// If it returns an error, that's also acceptable
			return
		}

		// With the new FileMeta return type, we should get an empty slice
		// when no files are processed
		if len(projectContext) > 0 {
			t.Errorf("Expected empty context files for non-existent path, got %d files", len(projectContext))
		}

		if stats != nil && stats.ProcessedFilesCount > 0 {
			t.Errorf("Expected zero processed files for non-existent path, got %d", stats.ProcessedFilesCount)
		}
	})

	// Test dry run mode
	t.Run("DryRunMode", func(t *testing.T) {
		logger := &mockContextLogger{}
		tokenManager := &mockTokenManager{}
		client := &mockGeminiClient{}

		gatherer := NewContextGatherer(logger, true, tokenManager)
		ctx := context.Background()

		config := GatherConfig{
			Paths:        []string{tempDir},
			Include:      "",
			Exclude:      "",
			ExcludeNames: "",
			Format:       "{path}\n{content}\n",
			Verbose:      true,
			LogLevel:     logutil.DebugLevel,
		}

		_, stats, err := gatherer.GatherContext(ctx, client, config)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// In dry run mode, ProcessedFiles slice should be populated
		if len(stats.ProcessedFiles) == 0 {
			t.Error("Expected non-empty ProcessedFiles slice in dry run mode")
		}

		// Verify logging contains dry run message
		found := false
		for _, msg := range logger.infoMessages {
			if msg == "Dry run mode: gathering files that would be included in context..." {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected dry run mode message in logs")
		}
	})
}

// TestDisplayDryRunInfo tests the DisplayDryRunInfo method
func TestDisplayDryRunInfo(t *testing.T) {
	logger := &mockContextLogger{}
	tokenManager := &mockTokenManager{}
	gatherer := NewContextGatherer(logger, true, tokenManager)
	ctx := context.Background()

	// Normal case with model info available
	t.Run("NormalCase", func(t *testing.T) {
		client := &mockGeminiClient{
			getModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  1000,
					OutputTokenLimit: 500,
				}, nil
			},
		}

		stats := &ContextStats{
			ProcessedFilesCount: 3,
			CharCount:           1000,
			LineCount:           50,
			TokenCount:          500,
			ProcessedFiles:      []string{"file1.go", "file2.txt", "file3.md"},
		}

		err := gatherer.DisplayDryRunInfo(ctx, client, stats)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify logs contain expected information
		if len(logger.infoMessages) < 5 {
			t.Error("Expected at least 5 info messages")
		}

		// Should show token limit comparison
		tokenLimitMsgFound := false
		for _, msg := range logger.infoMessages {
			if msg == "Token usage: %d / %d (%.1f%% of model's limit)" {
				tokenLimitMsgFound = true
				break
			}
		}
		if !tokenLimitMsgFound {
			t.Error("Expected token limit message in logs")
		}
	})

	// Error getting model info
	t.Run("ModelInfoError", func(t *testing.T) {
		client := &mockGeminiClient{
			getModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return nil, errors.New("model info error")
			},
		}

		stats := &ContextStats{
			ProcessedFilesCount: 3,
			CharCount:           1000,
			LineCount:           50,
			TokenCount:          500,
			ProcessedFiles:      []string{"file1.go", "file2.txt", "file3.md"},
		}

		err := gatherer.DisplayDryRunInfo(ctx, client, stats)
		if err != nil {
			t.Errorf("Expected no error (should handle model info error gracefully), got %v", err)
		}

		// Verify warning message
		modelErrorWarningFound := false
		for _, msg := range logger.warnMessages {
			if msg == "Could not get model information: %v" {
				modelErrorWarningFound = true
				break
			}
		}
		if !modelErrorWarningFound {
			t.Error("Expected model info error warning in logs")
		}
	})

	// Token limit exceeded
	t.Run("TokenLimitExceeded", func(t *testing.T) {
		client := &mockGeminiClient{
			getModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
				return &gemini.ModelInfo{
					Name:             "test-model",
					InputTokenLimit:  400, // Less than token count
					OutputTokenLimit: 200,
				}, nil
			},
		}

		stats := &ContextStats{
			ProcessedFilesCount: 3,
			CharCount:           1000,
			LineCount:           50,
			TokenCount:          500, // Exceeds limit
			ProcessedFiles:      []string{"file1.go", "file2.txt", "file3.md"},
		}

		err := gatherer.DisplayDryRunInfo(ctx, client, stats)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify warning message about token limit
		tokenLimitWarningFound := false
		for _, msg := range logger.errorMessages {
			if msg == "WARNING: Token count exceeds model's limit by %d tokens" {
				tokenLimitWarningFound = true
				break
			}
		}
		if !tokenLimitWarningFound {
			t.Error("Expected token limit exceeded warning in logs")
		}
	})

	// No files processed
	t.Run("NoFilesProcessed", func(t *testing.T) {
		client := &mockGeminiClient{}
		stats := &ContextStats{
			ProcessedFilesCount: 0,
			CharCount:           0,
			LineCount:           0,
			TokenCount:          0,
			ProcessedFiles:      []string{},
		}

		err := gatherer.DisplayDryRunInfo(ctx, client, stats)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify "no files matched" message
		noFilesMessageFound := false
		for _, msg := range logger.infoMessages {
			if msg == "  No files matched the current filters." {
				noFilesMessageFound = true
				break
			}
		}
		if !noFilesMessageFound {
			t.Error("Expected 'no files matched' message in logs")
		}
	})
}
