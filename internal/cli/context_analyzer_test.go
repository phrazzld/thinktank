package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplexityLevel_String(t *testing.T) {
	tests := []struct {
		level    ComplexityLevel
		expected string
	}{
		{ComplexitySimple, "Simple"},
		{ComplexityMedium, "Medium"},
		{ComplexityLarge, "Large"},
		{ComplexityXLarge, "XLarge"},
		{ComplexityLevel(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

func TestNewContextAnalyzer(t *testing.T) {
	t.Run("with logger", func(t *testing.T) {
		logger := testutil.NewMockLogger()
		analyzer := NewContextAnalyzer(logger)
		assert.NotNil(t, analyzer)
		assert.NotNil(t, analyzer.cache)
		assert.Equal(t, logger, analyzer.logger)
	})

	t.Run("with nil logger", func(t *testing.T) {
		analyzer := NewContextAnalyzer(nil)
		assert.NotNil(t, analyzer)
		assert.NotNil(t, analyzer.logger)
		assert.NotNil(t, analyzer.cache)
	})
}

func TestContextAnalyzer_estimateTokens(t *testing.T) {
	analyzer := NewContextAnalyzer(testutil.NewMockLogger())

	tests := []struct {
		name     string
		chars    int64
		expected int64
	}{
		{"zero chars", 0, 0},
		{"small text", 100, 25},
		{"medium text", 4000, 1000},
		{"large text", 200000, 50000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.estimateTokens(tt.chars)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContextAnalyzer_categorizeComplexity(t *testing.T) {
	analyzer := NewContextAnalyzer(testutil.NewMockLogger())

	tests := []struct {
		name     string
		tokens   int64
		expected ComplexityLevel
	}{
		{"simple project", 5000, ComplexitySimple},
		{"medium project", 25000, ComplexityMedium},
		{"large project", 100000, ComplexityLarge},
		{"xlarge project", 500000, ComplexityXLarge},
		{"boundary simple", 9999, ComplexitySimple},
		{"boundary medium", 10000, ComplexityMedium},
		{"boundary large", 50000, ComplexityLarge},
		{"boundary xlarge", 200000, ComplexityXLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.categorizeComplexity(tt.tokens)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContextAnalyzer_shouldAnalyzeFile(t *testing.T) {
	analyzer := NewContextAnalyzer(testutil.NewMockLogger())

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"go file", "/path/to/file.go", true},
		{"python file", "/path/to/script.py", true},
		{"text file", "/path/to/readme.txt", true},
		{"hidden file", "/path/to/.hidden", false},
		{"hidden directory", "/path/.git/config", false},
		{"binary executable", "/path/to/app.exe", false},
		{"image file", "/path/to/image.jpg", false},
		{"archive file", "/path/to/data.zip", false},
		{"object file", "/path/to/lib.o", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.shouldAnalyzeFile(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContextAnalyzer_isLikelyBinary(t *testing.T) {
	analyzer := NewContextAnalyzer(testutil.NewMockLogger())

	// Create temporary files for testing
	tempDir := t.TempDir()

	// Create a small text file
	textFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(textFile, []byte("hello world"), 0644)
	require.NoError(t, err)

	textInfo, err := os.Stat(textFile)
	require.NoError(t, err)

	tests := []struct {
		name     string
		path     string
		info     os.FileInfo
		expected bool
	}{
		{"text file", textFile, textInfo, false},
		{"exe file", "/path/to/app.exe", textInfo, true},
		{"image file", "/path/to/image.png", textInfo, true},
		{"archive file", "/path/to/data.tar.gz", textInfo, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.isLikelyBinary(tt.path, tt.info)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContextAnalyzer_generateCacheKey(t *testing.T) {
	analyzer := NewContextAnalyzer(testutil.NewMockLogger())

	// Same paths should generate same keys
	key1 := analyzer.generateCacheKey("/path/to/project")
	key2 := analyzer.generateCacheKey("/path/to/project")
	assert.Equal(t, key1, key2)

	// Different paths should generate different keys
	key3 := analyzer.generateCacheKey("/different/path")
	assert.NotEqual(t, key1, key3)

	// Keys should be hex strings
	assert.Regexp(t, "^[a-f0-9]{32}$", key1)
}

func TestContextAnalyzer_CacheOperations(t *testing.T) {
	analyzer := NewContextAnalyzer(testutil.NewMockLogger())

	// Initially empty cache
	stats := analyzer.GetCacheStats()
	assert.Equal(t, 0, stats["entries"])

	// Test cache operations
	testPath := "/test/path"
	testResult := &AnalysisResult{
		TotalFiles:      5,
		TotalLines:      100,
		TotalChars:      2000,
		EstimatedTokens: 500,
		Complexity:      ComplexitySimple,
	}

	// Cache result
	analyzer.cacheResult(testPath, testResult)

	// Verify cache has entry
	stats = analyzer.GetCacheStats()
	assert.Equal(t, 1, stats["entries"])

	// Clear cache
	analyzer.ClearCache()
	stats = analyzer.GetCacheStats()
	assert.Equal(t, 0, stats["entries"])
}

func TestContextAnalyzer_SerializationRoundTrip(t *testing.T) {
	analyzer := NewContextAnalyzer(testutil.NewMockLogger())

	// Add some test data to cache
	testResult := &AnalysisResult{
		TotalFiles:      10,
		TotalLines:      200,
		TotalChars:      4000,
		EstimatedTokens: 1000,
		Complexity:      ComplexityMedium,
	}
	analyzer.cacheResult("/test/path", testResult)

	// Serialize cache
	data, err := analyzer.SerializeCache()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Clear cache and deserialize
	analyzer.ClearCache()
	stats := analyzer.GetCacheStats()
	assert.Equal(t, 0, stats["entries"])

	err = analyzer.DeserializeCache(data)
	require.NoError(t, err)

	// Verify cache was restored
	stats = analyzer.GetCacheStats()
	assert.Equal(t, 1, stats["entries"])
}

func TestContextAnalyzer_IntegrationWithTempFiles(t *testing.T) {
	analyzer := NewContextAnalyzer(testutil.NewMockLogger())
	tempDir := t.TempDir()

	// Create test files with known content
	files := map[string]string{
		"main.go":    "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n",
		"utils.go":   "package main\n\nfunc helper() string {\n\treturn \"test\"\n}\n",
		"README.md":  "# Test Project\n\nThis is a test.\n",
		".hidden":    "hidden file content",
		"binary.exe": "\x00\x01\x02\x03", // Binary content
	}

	for filename, content := range files {
		path := filepath.Join(tempDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Analyze the temporary directory
	result, err := analyzer.AnalyzeComplexity(tempDir)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify results
	assert.Greater(t, result.TotalFiles, int64(0))
	assert.Greater(t, result.TotalLines, int64(0))
	assert.Greater(t, result.TotalChars, int64(0))
	assert.Greater(t, result.EstimatedTokens, int64(0))
	assert.False(t, result.CacheHit) // First analysis should not be a cache hit

	// Second analysis should hit cache
	result2, err := analyzer.AnalyzeComplexity(tempDir)
	require.NoError(t, err)
	assert.Equal(t, result.TotalFiles, result2.TotalFiles)
	assert.Equal(t, result.TotalLines, result2.TotalLines)
	assert.True(t, result2.CacheHit) // Should be a cache hit
}

func TestContextAnalyzer_analyzeTaskComplexity(t *testing.T) {
	analyzer := NewContextAnalyzer(testutil.NewMockLogger())
	tempDir := t.TempDir()

	// Create a simple test file
	testFile := filepath.Join(tempDir, "test.go")
	content := "package main\n\nfunc main() {\n\tprintln(\"hello world\")\n}\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Test the analyzeTaskComplexity function
	tokenCount, err := analyzer.analyzeTaskComplexity(tempDir)
	require.NoError(t, err)
	assert.Greater(t, tokenCount, int64(0))

	// Token count should be reasonable for the content
	expectedTokens := int64(len(content)) / avgCharsPerToken
	assert.Equal(t, expectedTokens, tokenCount)
}

func TestContextAnalyzer_ErrorHandling(t *testing.T) {
	analyzer := NewContextAnalyzer(testutil.NewMockLogger())

	t.Run("non-existent path", func(t *testing.T) {
		_, err := analyzer.AnalyzeComplexity("/non/existent/path")
		assert.Error(t, err)
	})

	t.Run("invalid serialization data", func(t *testing.T) {
		err := analyzer.DeserializeCache([]byte("invalid json"))
		assert.Error(t, err)
	})
}

func TestContextAnalyzer_PerformanceCharacteristics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	analyzer := NewContextAnalyzer(testutil.NewMockLogger())
	tempDir := t.TempDir()

	// Create multiple files to test performance
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("file%d.go", i))
		content := fmt.Sprintf("package main\n\n// File %d\nfunc func%d() {\n\treturn\n}\n", i, i)
		err := os.WriteFile(filename, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Time the analysis
	start := time.Now()
	result, err := analyzer.AnalyzeComplexity(tempDir)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Analysis should complete reasonably quickly (< 1 second for 100 small files)
	assert.Less(t, duration, time.Second, "Analysis took too long: %v", duration)

	// Verify we processed the expected number of files
	assert.Equal(t, int64(100), result.TotalFiles)

	// Cache hit should be much faster
	start = time.Now()
	result2, err := analyzer.AnalyzeComplexity(tempDir)
	cacheHitDuration := time.Since(start)

	require.NoError(t, err)
	assert.True(t, result2.CacheHit)

	// Cache hit should be faster (timing can be variable in tests)
	assert.Less(t, cacheHitDuration, duration, "Cache hit should be faster than initial analysis")
}

func TestAnalyzeTaskComplexityForModelSelection(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.go")
	content := "package main\n\nfunc main() {\n\tprintln(\"hello world\")\n}\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	// Test the convenience function
	tokenCount, err := AnalyzeTaskComplexityForModelSelection(tempDir)
	require.NoError(t, err)
	assert.Greater(t, tokenCount, int64(0))

	// Token count should be reasonable for the content
	expectedTokens := int64(len(content)) / avgCharsPerToken
	assert.Equal(t, expectedTokens, tokenCount)
}

func TestGetComplexityAnalysis(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple test files to get a complexity result
	files := map[string]string{
		"main.go":   "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n",
		"utils.go":  "package main\n\nfunc helper() string {\n\treturn \"test\"\n}\n",
		"readme.md": "# Test Project\n\nThis is a test.\n",
	}

	for filename, content := range files {
		path := filepath.Join(tempDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Test the convenience function
	result, err := GetComplexityAnalysis(tempDir)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify analysis results
	assert.Equal(t, int64(3), result.TotalFiles) // Should find 3 files
	assert.Greater(t, result.TotalLines, int64(0))
	assert.Greater(t, result.TotalChars, int64(0))
	assert.Greater(t, result.EstimatedTokens, int64(0))
	assert.Equal(t, ComplexitySimple, result.Complexity)
	assert.False(t, result.CacheHit) // First analysis should not be a cache hit
}

func TestContextAnalyzer_IntegrationWithModelSelector(t *testing.T) {
	tempDir := t.TempDir()

	// Create files of different complexities to test model selection
	complexityTests := []struct {
		name           string
		fileCount      int
		contentPerFile string
		expectedModel  string // We'll verify this matches expectations
	}{
		{
			name:           "simple project",
			fileCount:      2,
			contentPerFile: "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n",
			expectedModel:  "", // We'll check this is a valid model
		},
		{
			name:           "medium project",
			fileCount:      20,
			contentPerFile: strings.Repeat("// This is a comment\nfunc example() {\n\treturn nil\n}\n", 50),
			expectedModel:  "",
		},
	}

	for _, tt := range complexityTests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test directory for this complexity level
			testDir := filepath.Join(tempDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			require.NoError(t, err)

			// Create files
			for i := 0; i < tt.fileCount; i++ {
				filename := filepath.Join(testDir, fmt.Sprintf("file%d.go", i))
				err := os.WriteFile(filename, []byte(tt.contentPerFile), 0644)
				require.NoError(t, err)
			}

			// Analyze complexity
			tokenCount, err := AnalyzeTaskComplexityForModelSelection(testDir)
			require.NoError(t, err)
			assert.Greater(t, tokenCount, int64(0))

			// Get detailed analysis
			result, err := GetComplexityAnalysis(testDir)
			require.NoError(t, err)
			assert.Equal(t, int64(tt.fileCount), result.TotalFiles)

			// Test integration with model selector
			// Note: We're testing the integration point, not the specific model selection logic
			availableProviders := []string{"openai", "gemini"} // Mock available providers
			selectedModel := SelectOptimalModel(availableProviders, tokenCount)
			assert.NotEmpty(t, selectedModel, "Should select a model based on complexity")

			// Log the results for inspection
			t.Logf("Complexity: %s, Token Count: %d, Selected Model: %s",
				result.Complexity.String(), tokenCount, selectedModel)
		})
	}
}
