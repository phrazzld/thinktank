package architect_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/phrazzld/architect/internal/architect/interfaces"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/stretchr/testify/assert"
)

// Mock implementations for integration testing

// integrationContextGatherer implementation for testing
type integrationContextGatherer struct {
	logger      logutil.LoggerInterface
	dryRun      bool
	client      gemini.Client
	auditLogger auditlog.AuditLogger
}

// GatherContext implements interfaces.ContextGatherer
func (cg *integrationContextGatherer) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	// Log start of gathering
	_ = cg.auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now(),
		Operation: "GatherContext",
		Status:    "InProgress",
	})

	// Simple implementation for testing
	// Create FileMeta from files in tempDir
	files := []fileutil.FileMeta{}

	// Use config.Paths, config.Include, config.Exclude to find files
	// For simplicity in tests, we'll just walk the directories
	for _, dir := range config.Paths {
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Apply include/exclude filters
			if config.Include != "" && !strings.Contains(path, config.Include) {
				return nil
			}

			if config.ExcludeNames != "" && strings.Contains(path, config.ExcludeNames) {
				return nil
			}

			// Read file content
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			files = append(files, fileutil.FileMeta{
				Path:    path,
				Content: string(content),
			})

			return nil
		})
	}

	// Calculate stats
	stats := &interfaces.ContextStats{
		ProcessedFilesCount: len(files),
		CharCount:           0,
		LineCount:           0,
	}

	// Calculate character and line counts
	combinedContent := ""
	for _, file := range files {
		combinedContent += file.Content
		stats.CharCount += len(file.Content)
		stats.LineCount += len(strings.Split(file.Content, "\n"))
	}

	// Calculate token count using client
	if cg.client != nil {
		// Call the token counting function from the client
		tokenResult, err := cg.client.CountTokens(ctx, combinedContent)
		if err != nil {
			// Log failure
			_ = cg.auditLogger.Log(auditlog.AuditEntry{
				Timestamp: time.Now(),
				Operation: "GatherContext",
				Status:    "Failure",
				Error: &auditlog.ErrorInfo{
					Message: err.Error(),
				},
			})
			return nil, nil, err
		}

		// Set token count in stats
		stats.TokenCount = tokenResult.Total
	}

	// Log success
	_ = cg.auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now(),
		Operation: "GatherContext",
		Status:    "Success",
	})

	return files, stats, nil
}

// DisplayDryRunInfo implements interfaces.ContextGatherer
func (cg *integrationContextGatherer) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	// Log start of display
	_ = cg.auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now(),
		Operation: "DisplayDryRunInfo",
		Status:    "InProgress",
	})

	// Display stats
	cg.logger.Info("Dry Run Info:")
	cg.logger.Info("Processed %d files", stats.ProcessedFilesCount)
	cg.logger.Info("Character count: %d", stats.CharCount)
	cg.logger.Info("Line count: %d", stats.LineCount)
	cg.logger.Info("Token count: %d tokens", stats.TokenCount)

	// Log success
	_ = cg.auditLogger.Log(auditlog.AuditEntry{
		Timestamp: time.Now(),
		Operation: "DisplayDryRunInfo",
		Status:    "Success",
	})

	return nil
}

// integrationMockLogger for integration testing
type integrationMockLogger struct {
	logutil.LoggerInterface
	debugMessages []string
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func (m *integrationMockLogger) Debug(format string, args ...interface{}) {
	m.debugMessages = append(m.debugMessages, format)
}

func (m *integrationMockLogger) Info(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, format)
}

func (m *integrationMockLogger) Warn(format string, args ...interface{}) {
	m.warnMessages = append(m.warnMessages, format)
}

func (m *integrationMockLogger) Error(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, format)
}

func (m *integrationMockLogger) Fatal(format string, args ...interface{}) {
	m.errorMessages = append(m.errorMessages, "FATAL: "+format)
}

func (m *integrationMockLogger) Printf(format string, args ...interface{}) {
	m.infoMessages = append(m.infoMessages, format)
}

func (m *integrationMockLogger) Println(v ...interface{}) {
	m.infoMessages = append(m.infoMessages, fmt.Sprint(v...))
}

// integrationMockGeminiClient implements gemini.Client for testing
type integrationMockGeminiClient struct {
	generateContentFunc func(ctx context.Context, prompt string) (*gemini.GenerationResult, error)
	countTokensFunc     func(ctx context.Context, text string) (*gemini.TokenCount, error)
	getModelInfoFunc    func(ctx context.Context) (*gemini.ModelInfo, error)
	closeFunc           func() error
	modelNameFunc       func() string
	temperatureFunc     func() float32
	maxOutputTokensFunc func() int32
	topPFunc            func() float32
}

func (m *integrationMockGeminiClient) GenerateContent(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
	if m.generateContentFunc != nil {
		return m.generateContentFunc(ctx, prompt)
	}
	return &gemini.GenerationResult{
		Content:      "test content",
		TokenCount:   100,
		FinishReason: "STOP",
	}, nil
}

func (m *integrationMockGeminiClient) CountTokens(ctx context.Context, text string) (*gemini.TokenCount, error) {
	if m.countTokensFunc != nil {
		return m.countTokensFunc(ctx, text)
	}
	return &gemini.TokenCount{Total: 100}, nil
}

func (m *integrationMockGeminiClient) GetModelInfo(ctx context.Context) (*gemini.ModelInfo, error) {
	if m.getModelInfoFunc != nil {
		return m.getModelInfoFunc(ctx)
	}
	return &gemini.ModelInfo{
		InputTokenLimit:  1000,
		OutputTokenLimit: 1000,
	}, nil
}

func (m *integrationMockGeminiClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *integrationMockGeminiClient) GetModelName() string {
	if m.modelNameFunc != nil {
		return m.modelNameFunc()
	}
	return "test-model"
}

func (m *integrationMockGeminiClient) GetTemperature() float32 {
	if m.temperatureFunc != nil {
		return m.temperatureFunc()
	}
	return 0.7
}

func (m *integrationMockGeminiClient) GetMaxOutputTokens() int32 {
	if m.maxOutputTokensFunc != nil {
		return m.maxOutputTokensFunc()
	}
	return 1000
}

func (m *integrationMockGeminiClient) GetTopP() float32 {
	if m.topPFunc != nil {
		return m.topPFunc()
	}
	return 0.95
}

// integrationMockAuditLogger implements auditlog.AuditLogger for testing
type integrationMockAuditLogger struct {
	entries []auditlog.AuditEntry
}

func (m *integrationMockAuditLogger) Log(entry auditlog.AuditEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func (m *integrationMockAuditLogger) Close() error {
	return nil
}

// hasOperationWithStatus checks if a specific operation with given status exists in the log
func (m *integrationMockAuditLogger) hasOperationWithStatus(operation, status string) bool {
	for _, entry := range m.entries {
		if entry.Operation == operation && entry.Status == status {
			return true
		}
	}
	return false
}

// Integration Tests

// TestIntegration_ContextGatherer_TokenManager tests the interaction between
// ContextGatherer and TokenManager during context gathering and token counting
func TestIntegration_ContextGatherer_TokenManager(t *testing.T) {
	// Create temp directory with test files
	tempDir, err := os.MkdirTemp("", "context_gatherer_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []struct {
		path    string
		content string
	}{
		{filepath.Join(tempDir, "file1.go"), "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}"},
		{filepath.Join(tempDir, "file2.txt"), "This is a test file with some content."},
	}

	for _, tf := range testFiles {
		if err := os.WriteFile(tf.path, []byte(tf.content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.path, err)
		}
	}

	// Mock dependencies
	mockClient := &integrationMockGeminiClient{
		countTokensFunc: func(ctx context.Context, text string) (*gemini.TokenCount, error) {
			// Simulate token counting - return a token count proportional to text length
			return &gemini.TokenCount{
				Total: int32(len(text) / 5), // Simple heuristic: 1 token per 5 chars
			}, nil
		},
	}

	// Create mock logger instead of using mockTokenManager (handled by ContextGatherer)
	mockLogger := &integrationMockLogger{}
	mockAudit := &integrationMockAuditLogger{}

	// Create ContextGatherer
	gatherer := &integrationContextGatherer{
		logger:      mockLogger,
		dryRun:      false,
		client:      mockClient,
		auditLogger: mockAudit,
	}

	// Configure the test scenario
	gatherConfig := interfaces.GatherConfig{
		Paths:  []string{tempDir},
		Format: "false", // Format is a string in the interface
	}

	// Execute the gatherer
	contextFiles, stats, err := gatherer.GatherContext(context.Background(), gatherConfig)

	// Verify results
	assert.NoError(t, err, "GatherContext should not return an error")
	assert.Len(t, contextFiles, 2, "Should have gathered 2 files")
	assert.NotNil(t, stats, "ContextStats should not be nil")
	assert.Equal(t, 2, stats.ProcessedFilesCount, "Should report 2 processed files")
	assert.Greater(t, stats.TokenCount, int32(0), "Token count should be greater than 0")

	// Verify audit logging
	assert.True(t, mockAudit.hasOperationWithStatus("GatherContext", "Success"),
		"Expected successful GatherContext operation in audit log")

	// Verify context stats contain calculated token count
	expectedTokenCount := int32((len(testFiles[0].content) + len(testFiles[1].content)) / 5)
	assert.Equal(t, expectedTokenCount, stats.TokenCount,
		"Token count should match expected calculation")
}

// TestIntegration_ContextGatherer_TokenCountingError tests the behavior when token counting fails
func TestIntegration_ContextGatherer_TokenCountingError(t *testing.T) {
	// Create temp directory with test files
	tempDir, err := os.MkdirTemp("", "context_gatherer_error_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file
	testFilePath := filepath.Join(tempDir, "file1.go")
	if err := os.WriteFile(testFilePath, []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Mock dependencies with error response
	expectedError := errors.New("token counting failed")
	mockClient := &integrationMockGeminiClient{
		countTokensFunc: func(ctx context.Context, text string) (*gemini.TokenCount, error) {
			return nil, expectedError
		},
	}

	mockLogger := &integrationMockLogger{}
	mockAudit := &integrationMockAuditLogger{}

	// Create ContextGatherer
	gatherer := &integrationContextGatherer{
		logger:      mockLogger,
		dryRun:      false,
		client:      mockClient,
		auditLogger: mockAudit,
	}

	// Configure the test scenario
	gatherConfig := interfaces.GatherConfig{
		Paths:  []string{tempDir},
		Format: "false",
	}

	// Execute the gatherer
	_, _, err = gatherer.GatherContext(context.Background(), gatherConfig)

	// Verify error is propagated
	assert.Error(t, err, "GatherContext should return an error when token counting fails")
	assert.ErrorContains(t, err, expectedError.Error(), "Error should contain the original error message")

	// Verify audit logging for failure
	assert.True(t, mockAudit.hasOperationWithStatus("GatherContext", "Failure"),
		"Expected failed GatherContext operation in audit log")
}

// TestIntegration_ContextGatherer_DisplayDryRunInfo tests the dry run mode display
func TestIntegration_ContextGatherer_DisplayDryRunInfo(t *testing.T) {
	// Mock dependencies
	mockClient := &integrationMockGeminiClient{}
	mockLogger := &integrationMockLogger{}
	mockAudit := &integrationMockAuditLogger{}

	// Create ContextGatherer
	gatherer := &integrationContextGatherer{
		logger:      mockLogger,
		dryRun:      false,
		client:      mockClient,
		auditLogger: mockAudit,
	}

	// Create test stats
	stats := &interfaces.ContextStats{
		ProcessedFilesCount: 10,
		CharCount:           1000,
		LineCount:           100,
		TokenCount:          250,
	}

	// Execute DisplayDryRunInfo
	err := gatherer.DisplayDryRunInfo(context.Background(), stats)

	// Verify no error
	assert.NoError(t, err, "DisplayDryRunInfo should not return an error")

	// Verify logger messages
	assert.Greater(t, len(mockLogger.infoMessages), 0, "Expected info messages to be logged")

	// Check for expected stats in log messages
	var foundFileCount, foundTokenCount bool
	for _, msg := range mockLogger.infoMessages {
		if strings.Contains(msg, "10 files") {
			foundFileCount = true
		}
		if strings.Contains(msg, "250 tokens") {
			foundTokenCount = true
		}
	}

	assert.True(t, foundFileCount, "Log should contain file count")
	assert.True(t, foundTokenCount, "Log should contain token count")

	// Verify audit logging
	assert.True(t, mockAudit.hasOperationWithStatus("DisplayDryRunInfo", "Success"),
		"Expected successful DisplayDryRunInfo operation in audit log")
}

// TestIntegration_ContextGatherer_Configuration tests configuration handling
func TestIntegration_ContextGatherer_Configuration(t *testing.T) {
	// Create temp directory with test files
	tempDir, err := os.MkdirTemp("", "context_gatherer_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files in subdirectories
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	testFiles := []struct {
		path    string
		content string
	}{
		{filepath.Join(tempDir, "file1.go"), "package main"},
		{filepath.Join(tempDir, "file2.txt"), "Text content"},
		{filepath.Join(subDir, "file3.go"), "package subpkg"},
		{filepath.Join(subDir, "excluded.tmp"), "Temp file"},
	}

	for _, tf := range testFiles {
		if err := os.WriteFile(tf.path, []byte(tf.content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", tf.path, err)
		}
	}

	// Mock dependencies
	mockClient := &integrationMockGeminiClient{
		countTokensFunc: func(ctx context.Context, text string) (*gemini.TokenCount, error) {
			return &gemini.TokenCount{Total: int32(len(text))}, nil
		},
	}

	mockLogger := &integrationMockLogger{}
	mockAudit := &integrationMockAuditLogger{}

	// Create ContextGatherer
	gatherer := &integrationContextGatherer{
		logger:      mockLogger,
		dryRun:      false,
		client:      mockClient,
		auditLogger: mockAudit,
	}

	// Configure with include/exclude patterns
	gatherConfig := interfaces.GatherConfig{
		Paths:        []string{tempDir},
		Include:      "*.go", // Only include .go files
		Exclude:      "",
		ExcludeNames: "excluded.tmp", // Exclude temp files
		Format:       "false",
	}

	// Execute the gatherer
	contextFiles, stats, err := gatherer.GatherContext(context.Background(), gatherConfig)

	// Verify results
	assert.NoError(t, err, "GatherContext should not return an error")

	// Should only include the two .go files, excluding the .txt and .tmp files
	assert.Len(t, contextFiles, 2, "Should have gathered only 2 files (.go files)")

	// Verify only the expected files were included
	foundFile1 := false
	foundFile3 := false
	for _, file := range contextFiles {
		switch filepath.Base(file.Path) {
		case "file1.go":
			foundFile1 = true
		case "file3.go":
			foundFile3 = true
		case "file2.txt", "excluded.tmp":
			t.Errorf("Unexpected file included: %s", file.Path)
		}
	}

	assert.True(t, foundFile1, "file1.go should be included")
	assert.True(t, foundFile3, "file3.go should be included")

	// Verify stats are calculated correctly
	expectedContent := "package main" + "package subpkg"
	assert.Equal(t, int32(len(expectedContent)), stats.TokenCount,
		"Token count should match expected content length")
}
