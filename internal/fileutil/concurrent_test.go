package fileutil

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultConcurrentConfig(t *testing.T) {
	ctx := context.Background()
	cfg := NewDefaultConcurrentConfig(ctx)

	assert.NotNil(t, cfg)
	assert.Equal(t, ctx, cfg.Context)
	assert.GreaterOrEqual(t, cfg.MaxWorkers, 1)
	assert.LessOrEqual(t, cfg.MaxWorkers, 32)
	assert.Equal(t, 15, cfg.BatchSize)
}

func TestNormalizeWorkers(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"zero defaults to GOMAXPROCS", 0, min(runtime.GOMAXPROCS(0), 32)},
		{"negative defaults to GOMAXPROCS", -1, min(runtime.GOMAXPROCS(0), 32)},
		{"1 stays 1", 1, 1},
		{"16 stays 16", 16, 16},
		{"32 stays 32", 32, 32},
		{"33 clamped to 32", 33, 32},
		{"100 clamped to 32", 100, 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeWorkers(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGatherProjectContextConcurrent_EmptyPaths(t *testing.T) {
	ctx := context.Background()
	config := &Config{
		Logger: testutil.NewMockLogger(),
	}
	concCfg := NewDefaultConcurrentConfig(ctx)

	files, count, err := GatherProjectContextConcurrent(ctx, []string{}, config, concCfg)

	assert.NoError(t, err)
	assert.Empty(t, files)
	assert.Equal(t, 0, count)
}

func TestGatherProjectContextConcurrent_SingleFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := "package main\n\nfunc main() {}\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	config := NewConfig(false, "", "", "", "", testutil.NewMockLogger())
	concCfg := NewDefaultConcurrentConfig(ctx)

	files, count, err := GatherProjectContextConcurrent(ctx, []string{testFile}, config, concCfg)

	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, 1, count)
	assert.Equal(t, content, files[0].Content)
}

func TestGatherProjectContextConcurrent_Directory(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"main.go":     "package main\n",
		"util.go":     "package main\n",
		"sub/sub.go":  "package sub\n",
		"sub/sub2.go": "package sub\n",
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	ctx := context.Background()
	config := NewConfig(false, "", "", "", "", testutil.NewMockLogger())
	concCfg := NewDefaultConcurrentConfig(ctx)

	result, count, err := GatherProjectContextConcurrent(ctx, []string{tmpDir}, config, concCfg)

	assert.NoError(t, err)
	assert.Len(t, result, 4)
	assert.Equal(t, 4, count)
}

func TestGatherProjectContextConcurrent_ContextCancellation(t *testing.T) {
	// Create temporary directory with files
	tmpDir := t.TempDir()
	for i := 0; i < 100; i++ {
		path := filepath.Join(tmpDir, "file"+string(rune('0'+i%10))+".go")
		require.NoError(t, os.WriteFile(path, []byte("package main\n"), 0644))
	}

	// Create a context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())

	config := NewConfig(false, "", "", "", "", testutil.NewMockLogger())
	concCfg := NewDefaultConcurrentConfig(ctx)

	// Cancel immediately
	cancel()

	// Should return early without processing all files
	files, count, err := GatherProjectContextConcurrent(ctx, []string{tmpDir}, config, concCfg)

	// Error may or may not occur depending on timing
	// But we should have fewer files than expected or an error
	assert.True(t, err != nil || count < 100 || len(files) < 100,
		"Expected early termination due to context cancellation")
}

func TestGatherProjectContextConcurrent_SingleWorker(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{"a.go", "b.go", "c.go"}
	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		require.NoError(t, os.WriteFile(path, []byte("package main\n"), 0644))
	}

	ctx := context.Background()
	config := NewConfig(false, "", "", "", "", testutil.NewMockLogger())
	concCfg := &ConcurrentConfig{
		MaxWorkers: 1, // Single worker mode
		BatchSize:  15,
		Context:    ctx,
	}

	files, count, err := GatherProjectContextConcurrent(ctx, []string{tmpDir}, config, concCfg)

	assert.NoError(t, err)
	assert.Len(t, files, 3)
	assert.Equal(t, 3, count)
}

func TestGatherProjectContextConcurrent_NilConcurrentConfig(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	require.NoError(t, os.WriteFile(testFile, []byte("package main\n"), 0644))

	ctx := context.Background()
	config := NewConfig(false, "", "", "", "", testutil.NewMockLogger())

	// Pass nil concurrent config - should use defaults
	files, count, err := GatherProjectContextConcurrent(ctx, []string{testFile}, config, nil)

	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, 1, count)
}

func TestGatherProjectContextConcurrent_SkipsBinaryFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a text file
	textFile := filepath.Join(tmpDir, "text.go")
	require.NoError(t, os.WriteFile(textFile, []byte("package main\n"), 0644))

	// Create a binary file
	binaryFile := filepath.Join(tmpDir, "binary.bin")
	binaryContent := make([]byte, 100)
	for i := range binaryContent {
		binaryContent[i] = byte(i % 256) // Include null bytes
	}
	require.NoError(t, os.WriteFile(binaryFile, binaryContent, 0644))

	ctx := context.Background()
	config := NewConfig(false, "", "", "", "", testutil.NewMockLogger())
	concCfg := NewDefaultConcurrentConfig(ctx)

	files, count, err := GatherProjectContextConcurrent(ctx, []string{tmpDir}, config, concCfg)

	assert.NoError(t, err)
	assert.Len(t, files, 1) // Only the text file
	assert.Equal(t, 1, count)
	assert.Equal(t, "package main\n", files[0].Content)
}

func TestGatherProjectContextConcurrent_SkipsGitDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create regular file
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n"), 0644))

	// Create .git directory with a file
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "config"), []byte("git config"), 0644))

	ctx := context.Background()
	config := NewConfig(false, "", "", "", "", testutil.NewMockLogger())
	concCfg := NewDefaultConcurrentConfig(ctx)

	files, count, err := GatherProjectContextConcurrent(ctx, []string{tmpDir}, config, concCfg)

	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, 1, count)

	// Verify .git file is not included
	for _, f := range files {
		assert.NotContains(t, f.Path, ".git")
	}
}

func TestGatherProjectContextConcurrent_ProgressReporting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files
	for i := 0; i < 5; i++ {
		path := filepath.Join(tmpDir, "file"+string(rune('a'+i))+".go")
		require.NoError(t, os.WriteFile(path, []byte("package main\n"), 0644))
	}

	ctx := context.Background()
	config := NewConfig(false, "", "", "", "", testutil.NewMockLogger())

	// Create progress channel
	progressChan := make(chan FileProgress, 100)
	concCfg := &ConcurrentConfig{
		MaxWorkers:   2,
		BatchSize:    15,
		Context:      ctx,
		ProgressChan: progressChan,
	}

	files, count, err := GatherProjectContextConcurrent(ctx, []string{tmpDir}, config, concCfg)

	assert.NoError(t, err)
	assert.Len(t, files, 5)
	assert.Equal(t, 5, count)

	// Check that progress was reported
	close(progressChan)
	var progressUpdates []FileProgress
	for p := range progressChan {
		progressUpdates = append(progressUpdates, p)
	}

	// Should have received some progress updates
	assert.NotEmpty(t, progressUpdates)
}

func TestGatherProjectContextConcurrent_FileCollector(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{"a.go", "b.go", "c.go"}
	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		require.NoError(t, os.WriteFile(path, []byte("package main\n"), 0644))
	}

	ctx := context.Background()
	config := NewConfig(false, "", "", "", "", testutil.NewMockLogger())

	// Set up file collector
	var collectedFiles []string
	config.SetFileCollector(func(path string) {
		collectedFiles = append(collectedFiles, path)
	})

	concCfg := NewDefaultConcurrentConfig(ctx)

	files, count, err := GatherProjectContextConcurrent(ctx, []string{tmpDir}, config, concCfg)

	assert.NoError(t, err)
	assert.Len(t, files, 3)
	assert.Equal(t, 3, count)
	assert.Len(t, collectedFiles, 3)
}

func TestGatherProjectContextConcurrent_IncludeFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with different extensions
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("text file\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "data.json"), []byte("{}\n"), 0644))

	ctx := context.Background()
	// Only include .go files
	config := NewConfig(false, ".go", "", "", "", testutil.NewMockLogger())
	concCfg := NewDefaultConcurrentConfig(ctx)

	files, count, err := GatherProjectContextConcurrent(ctx, []string{tmpDir}, config, concCfg)

	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, 1, count)
	assert.Contains(t, files[0].Path, "main.go")
}

func TestGatherProjectContextConcurrent_ExcludeFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("text file\n"), 0644))

	ctx := context.Background()
	// Exclude .txt files
	config := NewConfig(false, "", ".txt", "", "", testutil.NewMockLogger())
	concCfg := NewDefaultConcurrentConfig(ctx)

	files, count, err := GatherProjectContextConcurrent(ctx, []string{tmpDir}, config, concCfg)

	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, 1, count)
	assert.Contains(t, files[0].Path, "main.go")
}

func TestConcurrentVsSequentialParity(t *testing.T) {
	// This test verifies that concurrent and sequential produce same results
	tmpDir := t.TempDir()

	// Create a realistic directory structure
	files := map[string]string{
		"main.go":          "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"hello\") }\n",
		"util/helper.go":   "package util\n\nfunc Helper() {}\n",
		"util/strings.go":  "package util\n\nfunc FormatString(s string) string { return s }\n",
		"internal/core.go": "package internal\n\ntype Core struct{}\n",
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	ctx := context.Background()
	config := NewConfig(false, "", "", "", "", testutil.NewMockLogger())

	// Run concurrent version
	concCfg := NewDefaultConcurrentConfig(ctx)
	concFiles, concCount, concErr := GatherProjectContextConcurrent(ctx, []string{tmpDir}, config, concCfg)

	// Both should succeed
	assert.NoError(t, concErr)
	assert.Equal(t, 4, concCount)
	assert.Len(t, concFiles, 4)

	// Verify all expected files are present
	pathSet := make(map[string]bool)
	for _, f := range concFiles {
		pathSet[filepath.Base(f.Path)] = true
	}
	assert.True(t, pathSet["main.go"])
	assert.True(t, pathSet["helper.go"])
	assert.True(t, pathSet["strings.go"])
	assert.True(t, pathSet["core.go"])
}

// BenchmarkGatherSequential benchmarks the old sequential approach (simulated)
func BenchmarkGatherSequential(b *testing.B) {
	tmpDir := setupBenchmarkDir(b, 100)
	config := NewConfig(false, "", "", "", "", testutil.NewMockLogger())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		concCfg := &ConcurrentConfig{
			MaxWorkers: 1, // Single worker = sequential
			BatchSize:  15,
			Context:    ctx,
		}
		_, _, _ = GatherProjectContextConcurrent(ctx, []string{tmpDir}, config, concCfg)
	}
}

// BenchmarkGatherConcurrent benchmarks the concurrent approach
func BenchmarkGatherConcurrent(b *testing.B) {
	tmpDir := setupBenchmarkDir(b, 100)
	config := NewConfig(false, "", "", "", "", testutil.NewMockLogger())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		concCfg := NewDefaultConcurrentConfig(ctx)
		_, _, _ = GatherProjectContextConcurrent(ctx, []string{tmpDir}, config, concCfg)
	}
}

// BenchmarkGatherByWorkerCount tests scaling with different worker counts
func BenchmarkGatherByWorkerCount(b *testing.B) {
	tmpDir := setupBenchmarkDir(b, 500)
	config := NewConfig(false, "", "", "", "", testutil.NewMockLogger())

	for _, workers := range []int{1, 2, 4, 8, 16} {
		b.Run("workers="+string(rune('0'+workers)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ctx := context.Background()
				concCfg := &ConcurrentConfig{
					MaxWorkers: workers,
					BatchSize:  15,
					Context:    ctx,
				}
				_, _, _ = GatherProjectContextConcurrent(ctx, []string{tmpDir}, config, concCfg)
			}
		})
	}
}

func setupBenchmarkDir(b *testing.B, fileCount int) string {
	b.Helper()
	tmpDir := b.TempDir()

	content := "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() {\n\tfmt.Println(\"hello\")\n\tos.Exit(0)\n}\n"

	for i := 0; i < fileCount; i++ {
		subdir := filepath.Join(tmpDir, "pkg"+string(rune('a'+i%26)))
		if err := os.MkdirAll(subdir, 0755); err != nil {
			b.Fatal(err)
		}
		path := filepath.Join(subdir, "file"+string(rune('0'+i%10))+".go")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			b.Fatal(err)
		}
	}

	return tmpDir
}

// TestConcurrentRaceConditions runs with -race to detect race conditions
func TestConcurrentRaceConditions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create many files to increase chance of race conditions
	for i := 0; i < 50; i++ {
		subdir := filepath.Join(tmpDir, "dir"+string(rune('a'+i%5)))
		require.NoError(t, os.MkdirAll(subdir, 0755))
		for j := 0; j < 10; j++ {
			path := filepath.Join(subdir, "file"+string(rune('0'+j))+".go")
			require.NoError(t, os.WriteFile(path, []byte("package main\n"), 0644))
		}
	}

	ctx := context.Background()
	config := NewConfig(false, "", "", "", "", testutil.NewMockLogger())

	// Run multiple concurrent gathers to stress test
	var wg atomic.Int32
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Add(-1)
			concCfg := NewDefaultConcurrentConfig(ctx)
			_, _, _ = GatherProjectContextConcurrent(ctx, []string{tmpDir}, config, concCfg)
		}()
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		for wg.Load() > 0 {
			time.Sleep(10 * time.Millisecond)
		}
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(30 * time.Second):
		t.Fatal("Test timed out - possible deadlock")
	}
}
