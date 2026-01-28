package fileutil

import (
	"os"
	"strings"
	"testing"

	"github.com/misty-step/thinktank/internal/testutil/perftest"
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

// BenchmarkGitChecker benchmarks the GitChecker caching behavior.
//
// This benchmark demonstrates the value of caching for git repository detection:
//
//   - "Subprocess": Raw cost of spawning git process (~5ms per call)
//   - "CacheHit": Cost of sync.Map lookup (~70ns per call)
//   - "ColdCache_UniqueDirectories": Simulates directory walk where each dir is new
//   - "WarmCache_SameDirectory": Simulates checking many files in same directory
//
// Real-world impact:
//   - Without caching: 1000 files in 100 dirs = 100 subprocess calls = ~500ms
//   - With caching: 1000 files in 100 dirs = 100 subprocess calls + 900 cache hits = ~500ms + ~0.07ms
//   - The cache saves time when multiple files share directories (common in codebases)
func BenchmarkGitChecker(b *testing.B) {
	currentDir, err := os.Getwd()
	if err != nil {
		b.Skip("Could not get current directory")
	}

	b.Run("Subprocess", func(b *testing.B) {
		// Measures raw subprocess cost - this is the baseline we're optimizing
		perftest.RunBenchmark(b, "CheckGitRepo_Subprocess", func(b *testing.B) {
			perftest.ReportAllocs(b)
			for i := 0; i < b.N; i++ {
				_ = CheckGitRepo(currentDir)
			}
		})
	})

	b.Run("CacheHit", func(b *testing.B) {
		// Measures cache lookup cost after initial population
		gc := NewGitChecker()
		_ = gc.IsRepo(currentDir) // Prime cache

		perftest.RunBenchmark(b, "GitChecker_CacheHit", func(b *testing.B) {
			perftest.ReportAllocs(b)
			for i := 0; i < b.N; i++ {
				_ = gc.IsRepo(currentDir)
			}
		})
	})

	b.Run("ColdCache_UniqueDirectories", func(b *testing.B) {
		// Simulates worst case: every directory is unique (no cache benefit)
		// This is the cost of walking a directory tree for the first time
		tempDirs := make([]string, b.N)
		for i := 0; i < b.N; i++ {
			tempDirs[i] = b.TempDir()
		}

		gc := NewGitChecker()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = gc.IsRepo(tempDirs[i])
		}
	})

	b.Run("WarmCache_SameDirectory", func(b *testing.B) {
		// Simulates best case: many files in same directory (typical codebase)
		// First call pays subprocess cost, subsequent calls are free
		gc := NewGitChecker()

		perftest.RunBenchmark(b, "GitChecker_WarmCache", func(b *testing.B) {
			perftest.ReportAllocs(b)
			for i := 0; i < b.N; i++ {
				_ = gc.IsRepo(currentDir)
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
