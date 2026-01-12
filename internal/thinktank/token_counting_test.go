package thinktank

import (
	"context"
	"fmt"
	"testing"

	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/phrazzld/thinktank/internal/thinktank/interfaces"
	"github.com/phrazzld/thinktank/internal/thinktank/tokenizers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewTokenCountingServiceWithLoggerForTest creates a service with a specific logger for testing
func NewTokenCountingServiceWithLoggerForTest(logger logutil.LoggerInterface) interfaces.TokenCountingService {
	return &tokenCountingServiceImpl{
		tokenizerManager: nil, // Will be lazy-loaded
		logger:           logger,
	}
}

// Basic smoke test to ensure the service can be created and used
func TestTokenCountingService_BasicSmoke(t *testing.T) {
	t.Parallel()

	service := NewTokenCountingService()
	require.NotNil(t, service, "Should create token counting service")

	// Test basic functionality
	result, err := service.CountTokens(context.Background(), TokenCountingRequest{
		Instructions: "test",
		Files:        []FileContent{},
	})

	require.NoError(t, err)
	assert.NotNil(t, result)
}

// Test service creation with custom manager
func TestTokenCountingService_WithCustomManager(t *testing.T) {
	t.Parallel()

	mockLogger := &testutil.MockLogger{}
	mockManager := &MockTokenizerManager{}

	service := NewTokenCountingServiceWithManagerAndLogger(mockManager, mockLogger)
	require.NotNil(t, service, "Should create service with custom manager")

	// Test that it works
	result, err := service.CountTokens(context.Background(), TokenCountingRequest{
		Instructions: "test",
		Files:        []FileContent{},
	})

	require.NoError(t, err)
	assert.NotNil(t, result)
}

// Mock implementations shared across test files

type MockTokenizerManager struct{}

func (m *MockTokenizerManager) GetTokenizer(provider string) (tokenizers.AccurateTokenCounter, error) {
	return &MockTokenizer{}, nil
}

func (m *MockTokenizerManager) SupportsProvider(provider string) bool {
	return provider == "openai" || provider == "gemini"
}

func (m *MockTokenizerManager) ClearCache() {
	// No-op for mock
}

type MockTokenizer struct {
	ShouldFail bool
}

func (m *MockTokenizer) CountTokens(ctx context.Context, text string, modelName string) (int, error) {
	if m.ShouldFail {
		return 0, fmt.Errorf("mock tokenizer failure")
	}
	// Simple approximation for testing
	return len(text), nil
}

func (m *MockTokenizer) SupportsModel(modelName string) bool {
	return !m.ShouldFail
}

func (m *MockTokenizer) GetEncoding(modelName string) (string, error) {
	if m.ShouldFail {
		return "", fmt.Errorf("mock encoding failure")
	}
	return "mock-encoding", nil
}

// BenchmarkCountFileTokens benchmarks the file token counting with various file counts.
// This benchmark verifies O(n) performance after fixing O(n²) string concatenation.
func BenchmarkCountFileTokens(b *testing.B) {
	service := NewTokenCountingService().(*tokenCountingServiceImpl)

	// Generate test files with realistic content sizes
	generateFiles := func(count int) []interfaces.FileContent {
		files := make([]interfaces.FileContent, count)
		// Average file content ~1KB each
		content := "package main\n\nfunc example() {\n\t// Sample code\n\treturn nil\n}\n"
		for i := range files {
			files[i] = interfaces.FileContent{
				Path:    fmt.Sprintf("file%d.go", i),
				Content: content,
			}
		}
		return files
	}

	benchmarks := []struct {
		name      string
		fileCount int
	}{
		{"10_files", 10},
		{"100_files", 100},
		{"500_files", 500},
		{"1000_files", 1000},
	}

	for _, bm := range benchmarks {
		files := generateFiles(bm.fileCount)
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = service.countFileTokens(files)
			}
		})
	}
}
