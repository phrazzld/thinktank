package fileutil

import (
	"os"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/testutil/perftest"
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
			perftest.RunBenchmark(b, "EstimateTokenCount_"+bm.name, func(b *testing.B) {
				perftest.ReportAllocs(b)
				for i := 0; i < b.N; i++ {
					_ = estimateTokenCount(bm.text)
				}
			})
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
				Logger: NewMockLogger(), // Use mock logger for testing
			},
		},
		{
			name: "With Include Filters",
			path: "src/main.go",
			config: &Config{
				Logger:      NewMockLogger(),
				IncludeExts: []string{".go", ".md", ".txt"},
			},
		},
		{
			name: "With Exclude Filters",
			path: "dist/bundle.js",
			config: &Config{
				Logger:       NewMockLogger(),
				ExcludeExts:  []string{".exe", ".bin", ".obj"},
				ExcludeNames: []string{"node_modules", "dist", "build"},
			},
		},
		{
			name: "With All Filters",
			path: "src/components/App.tsx",
			config: &Config{
				Logger:       NewMockLogger(),
				IncludeExts:  []string{".go", ".md", ".ts", ".tsx"},
				ExcludeExts:  []string{".exe", ".bin", ".obj"},
				ExcludeNames: []string{"node_modules", "dist", "build"},
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			perftest.RunBenchmark(b, "ShouldProcess_"+bm.name, func(b *testing.B) {
				perftest.ReportAllocs(b)
				for i := 0; i < b.N; i++ {
					_ = shouldProcess(bm.path, bm.config)
				}
			})
		})
	}
}

// BenchmarkGitRepoCaching benchmarks cached vs uncached git repo detection
func BenchmarkGitRepoCaching(b *testing.B) {
	// Use current directory (likely a git repo)
	currentDir, err := os.Getwd()
	if err != nil {
		b.Skip("Could not get current directory")
	}

	b.Run("Uncached", func(b *testing.B) {
		perftest.RunBenchmark(b, "CheckGitRepo_Uncached", func(b *testing.B) {
			perftest.ReportAllocs(b)
			for i := 0; i < b.N; i++ {
				_ = CheckGitRepo(currentDir)
			}
		})
	})

	b.Run("Cached", func(b *testing.B) {
		ClearGitCaches()
		// Prime the cache with one call
		_ = CheckGitRepoCached(currentDir)

		perftest.RunBenchmark(b, "CheckGitRepoCached", func(b *testing.B) {
			perftest.ReportAllocs(b)
			for i := 0; i < b.N; i++ {
				_ = CheckGitRepoCached(currentDir)
			}
		})
	})
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
			perftest.RunBenchmark(b, "IsBinaryFile_"+bm.name, func(b *testing.B) {
				perftest.ReportAllocs(b)
				for i := 0; i < b.N; i++ {
					_ = isBinaryFile(bm.content)
				}
			})
		})
	}
}
