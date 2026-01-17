// Package fileutil provides file system utilities for project context gathering.
package fileutil

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
)

// ConcurrentConfig holds configuration for concurrent file processing
type ConcurrentConfig struct {
	// MaxWorkers specifies the number of worker goroutines per stage.
	// Default: runtime.GOMAXPROCS(0). Range: 1-32.
	MaxWorkers int

	// Context for cancellation propagation
	Context context.Context

	// ProgressChan receives optional progress updates during processing.
	// If nil, no progress updates are sent.
	ProgressChan chan<- FileProgress
}

// FileProgress reports progress during file gathering
type FileProgress struct {
	Stage           string // "discovery", "filtering", "reading"
	TotalDiscovered int64
	TotalProcessed  int64
	TotalSkipped    int64
	PercentComplete float64
}

// discoverResult wraps discovered file path with any error
type discoverResult struct {
	path string
	err  error
}

// filterResult wraps filtering decision
type filterResult struct {
	path      string
	shouldAdd bool
}

// readResult wraps file read result
type readResult struct {
	meta FileMeta
	err  error
}

// NewDefaultConcurrentConfig creates a ConcurrentConfig with sensible defaults
func NewDefaultConcurrentConfig(ctx context.Context) *ConcurrentConfig {
	return &ConcurrentConfig{
		MaxWorkers: normalizeWorkers(0),
		Context:    ctx,
	}
}

// normalizeWorkers ensures workers is within valid range [1, 32]
func normalizeWorkers(workers int) int {
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}
	return max(1, min(workers, 32))
}

// GatherProjectContextConcurrent walks paths and gathers files using concurrent processing.
// This provides significant performance improvements on multi-core systems by parallelizing
// file discovery, filtering, and reading.
//
// The pipeline has three stages:
//   - Stage 1 (Discovery): Concurrent directory walking
//   - Stage 2 (Filtering): Parallel filtering and git-ignore checks
//   - Stage 3 (Reading): Concurrent file reading and binary detection
//
// This function maintains backward compatibility - results are identical to the sequential
// version, just faster.
func GatherProjectContextConcurrent(ctx context.Context, paths []string, config *Config, concCfg *ConcurrentConfig) ([]FileMeta, int, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if concCfg == nil {
		concCfg = NewDefaultConcurrentConfig(ctx)
	}

	workers := normalizeWorkers(concCfg.MaxWorkers)
	bufSize := workers * 2
	if bufSize < 10 {
		bufSize = 10
	}

	// Initialize counters for progress tracking
	var totalDiscovered atomic.Int64
	var totalProcessed atomic.Int64
	var totalSkipped atomic.Int64

	// Create pipeline channels with buffering to prevent blocking
	discoverChan := make(chan discoverResult, bufSize)
	filterChan := make(chan filterResult, bufSize)
	readChan := make(chan readResult, bufSize)

	// Start the three-stage pipeline
	go discoverFiles(ctx, paths, config, workers, discoverChan, &totalDiscovered)
	go filterFiles(ctx, discoverChan, config, workers, filterChan, &totalSkipped)
	go readFiles(ctx, filterChan, config, workers, readChan, &totalSkipped)

	// Collect results from the final stage (sequential - no mutex needed)
	var files []FileMeta

	for result := range readChan {
		if result.err != nil {
			config.Logger.Printf("Warning: Error processing file: %v\n", result.err)
			totalSkipped.Add(1)
			continue
		}

		files = append(files, result.meta)

		totalProcessed.Add(1)

		// Call file collector if set
		if config.fileCollector != nil {
			config.fileCollector(result.meta.Path)
		}

		// Send progress update if channel is available
		if concCfg.ProgressChan != nil {
			discovered := totalDiscovered.Load()
			processed := totalProcessed.Load()
			skipped := totalSkipped.Load()

			var percent float64
			if discovered > 0 {
				percent = float64(processed+skipped) / float64(discovered) * 100
			}

			select {
			case concCfg.ProgressChan <- FileProgress{
				Stage:           "reading",
				TotalDiscovered: discovered,
				TotalProcessed:  processed,
				TotalSkipped:    skipped,
				PercentComplete: percent,
			}:
			default:
				// Don't block if channel is full
			}
		}
	}

	// Sort files by path for deterministic output
	// This ensures tests pass and output is predictable regardless of goroutine ordering
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return files, int(totalProcessed.Load()), nil
}

// discoverFiles walks directories concurrently using worker pool pattern
func discoverFiles(ctx context.Context, paths []string, config *Config, workers int, results chan<- discoverResult, totalDiscovered *atomic.Int64) {
	defer close(results)

	var wg sync.WaitGroup
	pathChan := make(chan string, workers*2)

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range pathChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				info, err := StatPath(path)
				if err != nil {
					config.Logger.Printf("Warning: Cannot stat path %s: %v. Skipping.\n", path, err)
					continue
				}

				if info.IsDir() {
					walkDirectoryConcurrent(ctx, path, config, results, totalDiscovered)
				} else {
					totalDiscovered.Add(1)
					select {
					case results <- discoverResult{path: path}:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	// Send paths to workers
	go func() {
		for _, p := range paths {
			select {
			case <-ctx.Done():
				close(pathChan)
				return
			case pathChan <- p:
			}
		}
		close(pathChan)
	}()

	wg.Wait()
}

// walkDirectoryConcurrent walks a single directory and sends files to results channel
func walkDirectoryConcurrent(ctx context.Context, root string, config *Config, results chan<- discoverResult, totalDiscovered *atomic.Int64) {
	err := WalkDirectory(root, func(path string, d os.DirEntry, err error) error {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			// Log error in a format that matches the sequential implementation expectation
			// The test expects "Error walking directory" or "Cannot stat path"
			config.Logger.Printf("Error walking directory %s: %v\n", path, err)
			return err // Return error to potentially stop walk on this branch
		}

		// Skip directories that should be excluded
		if d.IsDir() {
			base := d.Name()
			// Skip .git and other excluded directories
			if base == ".git" || isGitIgnored(path, config) {
				config.Logger.Printf("Verbose: Skipping directory: %s\n", path)
				return filepath.SkipDir
			}
			// Check explicit excludes
			for _, name := range config.ExcludeNames {
				if base == name {
					config.Logger.Printf("Verbose: Skipping directory: %s\n", path)
					return filepath.SkipDir
				}
			}
			return nil // Continue into directory
		}

		// It's a file - send to results
		totalDiscovered.Add(1)
		select {
		case results <- discoverResult{path: path}:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	})

	if err != nil && err != context.Canceled {
		config.Logger.Printf("Error walking directory %s: %v\n", root, err)
	}
}

// filterFiles filters discovered paths concurrently
func filterFiles(ctx context.Context, discovered <-chan discoverResult, config *Config, workers int, results chan<- filterResult, totalSkipped *atomic.Int64) {
	defer close(results)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range discovered {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if item.err != nil {
					totalSkipped.Add(1)
					continue
				}

				shouldAdd := shouldProcess(item.path, config)
				if !shouldAdd {
					totalSkipped.Add(1)
				}

				select {
				case results <- filterResult{path: item.path, shouldAdd: shouldAdd}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	wg.Wait()
}

// readFiles reads filtered file contents concurrently
func readFiles(ctx context.Context, filtered <-chan filterResult, config *Config, workers int, results chan<- readResult, totalSkipped *atomic.Int64) {
	defer close(results)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range filtered {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if !item.shouldAdd {
					continue
				}

				content, err := ReadFileContent(item.path)
				if err != nil {
					config.Logger.Printf("Warning: Cannot read file %s: %v\n", item.path, err)
					totalSkipped.Add(1)
					continue
				}

				if isBinaryFile(content) {
					config.Logger.Printf("Verbose: Skipping binary file: %s\n", item.path)
					totalSkipped.Add(1)
					continue
				}

				select {
				case results <- readResult{
					meta: FileMeta{Path: EnsureAbsolutePath(item.path), Content: string(content)},
				}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	wg.Wait()
}
