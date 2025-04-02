package fileutil

import (
	"context"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// BenchmarkEstimateTokenCount benchmarks the estimateTokenCount function
func BenchmarkEstimateTokenCount(b *testing.B) {
	benchmarks := []struct {
		name string
		text string
	}{
		{
			name: "Small",
			text: strings.Repeat("Hello world. ", 10), // About 120 chars
		},
		{
			name: "Medium",
			text: strings.Repeat("Hello world. This is a benchmark test for token counting. ", 50), // ~2000 chars
		},
		{
			name: "Large",
			text: strings.Repeat("The quick brown fox jumps over the lazy dog. ", 1000), // ~40,000 chars
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = estimateTokenCount(bm.text)
			}
		})
	}
}

// BenchmarkShouldProcess benchmarks the shouldProcess function
func BenchmarkShouldProcess(b *testing.B) {
	benchmarks := []struct {
		name   string
		path   string
		config *Config
	}{
		{
			name: "Simple Path No Filters",
			path: "test.txt",
			config: &Config{
				Logger: logutil.NewLogger(logutil.WarnLevel, nil, "", false), // Minimal logging
			},
		},
		{
			name: "With Include Filters",
			path: "src/main.go",
			config: &Config{
				Logger:      logutil.NewLogger(logutil.WarnLevel, nil, "", false),
				IncludeExts: []string{".go", ".md", ".txt"},
			},
		},
		{
			name: "With Exclude Filters",
			path: "dist/bundle.js",
			config: &Config{
				Logger:       logutil.NewLogger(logutil.WarnLevel, nil, "", false),
				ExcludeExts:  []string{".exe", ".bin", ".obj"},
				ExcludeNames: []string{"node_modules", "dist", "build"},
			},
		},
		{
			name: "With All Filters",
			path: "src/components/App.tsx",
			config: &Config{
				Logger:       logutil.NewLogger(logutil.WarnLevel, nil, "", false),
				IncludeExts:  []string{".go", ".md", ".ts", ".tsx"},
				ExcludeExts:  []string{".exe", ".bin", ".obj"},
				ExcludeNames: []string{"node_modules", "dist", "build"},
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = shouldProcess(bm.path, bm.config)
			}
		})
	}
}

// BenchmarkCalculateStatisticsWithTokenCounting benchmarks the CalculateStatisticsWithTokenCounting function
func BenchmarkCalculateStatisticsWithTokenCounting(b *testing.B) {
	text := strings.Repeat("This is a benchmark test for token counting statistics calculation. ", 100)
	ctx := context.Background()

	// Mock client that returns a constant token count
	mockClient := &gemini.MockClient{
		CountTokensFunc: func(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
			// Simulate processing time
			return &gemini.TokenCount{Total: int32(len(prompt) / 5)}, nil
		},
	}

	logger := logutil.NewLogger(logutil.WarnLevel, nil, "", false)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = CalculateStatisticsWithTokenCounting(ctx, mockClient, text, logger)
	}
}

// BenchmarkIsBinaryFile benchmarks the isBinaryFile function
func BenchmarkIsBinaryFile(b *testing.B) {
	benchmarks := []struct {
		name    string
		content []byte
	}{
		{
			name:    "Small Text",
			content: []byte(strings.Repeat("Hello world. ", 10)),
		},
		{
			name:    "Medium Text",
			content: []byte(strings.Repeat("Hello world. ", 100)),
		},
		{
			name:    "Small Binary",
			content: append([]byte{0x00, 0x01, 0x02}, []byte(strings.Repeat("A", 500))...),
		},
		{
			name:    "Medium Binary",
			content: append([]byte{0x00, 0x01, 0x02}, []byte(strings.Repeat("B", 2000))...),
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = isBinaryFile(bm.content)
			}
		})
	}
}
