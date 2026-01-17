// internal/fileutil/fileutil.go
package fileutil

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"github.com/phrazzld/thinktank/internal/logutil"
)

// FileMeta represents a file with its path and content.
type FileMeta struct {
	Path    string
	Content string
}

// Config holds file processing configuration
type Config struct {
	Verbose        bool
	IncludeExts    []string
	ExcludeExts    []string
	ExcludeNames   []string
	Format         string
	Logger         logutil.LoggerInterface
	GitAvailable   bool
	GitChecker     *GitChecker // Cached git operations (created automatically if nil)
	processedFiles int
	totalFiles     int               // For verbose logging
	fileCollector  func(path string) // Optional callback to collect processed file paths
}

// NewConfig creates a configuration with defaults.
func NewConfig(verbose bool, include, exclude, excludeNames, format string, logger logutil.LoggerInterface) *Config {
	// Check if git is available
	_, gitErr := exec.LookPath("git")
	gitAvailable := gitErr == nil

	if logger == nil {
		// Use the slog-based logger instead of the standard library logger
		logger = logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)
	}

	cfg := &Config{
		Verbose:      verbose,
		Format:       format,
		Logger:       logger,
		GitAvailable: gitAvailable,
		GitChecker:   NewGitChecker(),
	}

	// Process include/exclude extensions
	if include != "" {
		cfg.IncludeExts = strings.Split(include, ",")
		for i, ext := range cfg.IncludeExts {
			ext = strings.TrimSpace(ext)
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			cfg.IncludeExts[i] = strings.ToLower(ext)
		}
	}
	if exclude != "" {
		cfg.ExcludeExts = strings.Split(exclude, ",")
		for i, ext := range cfg.ExcludeExts {
			ext = strings.TrimSpace(ext)
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			cfg.ExcludeExts[i] = strings.ToLower(ext)
		}
	}
	// Process exclude names
	if excludeNames != "" {
		cfg.ExcludeNames = strings.Split(excludeNames, ",")
		for i, name := range cfg.ExcludeNames {
			cfg.ExcludeNames[i] = strings.TrimSpace(name)
		}
	}

	return cfg
}

// SetFileCollector sets a callback function that will be called for each processed file
func (c *Config) SetFileCollector(collector func(path string)) {
	c.fileCollector = collector
}

// isGitIgnored checks if a file is likely ignored by git or is hidden.
func isGitIgnored(path string, config *Config) bool {
	base := filepath.Base(path)

	// Always ignore .git directory contents
	if base == ".git" || strings.Contains(path, string(filepath.Separator)+".git"+string(filepath.Separator)) {
		return true
	}

	// Check git ignore status if git is available
	if config.GitAvailable {
		// Lazy init GitChecker if not set (preserves behavior for manual Config creation)
		if config.GitChecker == nil {
			config.GitChecker = NewGitChecker()
		}
		dir := filepath.Dir(path)
		isIgnored, err := config.GitChecker.IsIgnored(dir, base)
		if err != nil {
			config.Logger.Printf("Verbose: Error running git check-ignore for %s: %v. Falling back.\n", path, err)
		} else if isIgnored {
			config.Logger.Printf("Verbose: Git ignored: %s\n", path)
			return true
		}
	}

	// Check if hidden file/directory (starts with dot)
	if strings.HasPrefix(base, ".") && base != "." && base != ".." {
		config.Logger.Printf("Verbose: Hidden file/dir ignored: %s\n", path)
		return true
	}

	return false
}

// Constants for binary file detection
const (
	binarySampleSize            = 512
	binaryNonPrintableThreshold = 0.3
)

// isBinaryFile checks if content is likely binary.
func isBinaryFile(content []byte) bool {
	if len(content) > 0 && bytes.IndexByte(content, 0) != -1 {
		return true // Contains null byte
	}
	nonPrintable := 0
	sampleSize := min(len(content), binarySampleSize)
	for i := 0; i < sampleSize; i++ {
		if content[i] < 32 && !isWhitespace(content[i]) {
			nonPrintable++
		}
	}
	return float64(nonPrintable) > float64(sampleSize)*binaryNonPrintableThreshold
}

func isWhitespace(b byte) bool {
	return b == '\n' || b == '\r' || b == '\t' || b == ' '
}

// shouldProcess checks all filters for a given file path.
// This function now uses the pure filtering logic and adds logging.
func shouldProcess(path string, config *Config) bool {
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(path))

	// Check if explicitly excluded by name
	if len(config.ExcludeNames) > 0 && slices.Contains(config.ExcludeNames, base) {
		config.Logger.Printf("Verbose: Skipping excluded name: %s\n", path)
		return false
	}

	// Check if gitignored or hidden (handles .git implicitly)
	if isGitIgnored(path, config) {
		return false // Logging done within isGitIgnored
	}

	// Check include extensions (if specified)
	if len(config.IncludeExts) > 0 {
		included := false
		for _, includeExt := range config.IncludeExts {
			if ext == includeExt {
				included = true
				break
			}
		}
		if !included {
			config.Logger.Printf("Verbose: Skipping non-included extension: %s (%s)\n", path, ext)
			return false
		}
	}

	// Check exclude extensions
	if len(config.ExcludeExts) > 0 {
		for _, excludeExt := range config.ExcludeExts {
			if ext == excludeExt {
				config.Logger.Printf("Verbose: Skipping excluded extension: %s (%s)\n", path, ext)
				return false
			}
		}
	}

	return true
}

// processFile reads, checks, and adds a file to the FileMeta slice.
func processFile(path string, files *[]FileMeta, config *Config) {
	config.totalFiles++ // Increment total count when we attempt to process

	// Run all checks first
	if !shouldProcess(path, config) {
		return // Already logged why it was skipped
	}

	content, err := ReadFileContent(path)
	if err != nil {
		config.Logger.Printf("Warning: Cannot read file %s: %v\n", path, err)
		return
	}

	if isBinaryFile(content) {
		config.Logger.Printf("Verbose: Skipping binary file: %s\n", path)
		return
	}

	// If all checks pass, process it
	config.processedFiles++
	config.Logger.Printf("Verbose: Processing file (%d/%d): %s (size: %d bytes)\n",
		config.processedFiles, config.totalFiles, path, len(content))

	// If a file collector is set, call it
	if config.fileCollector != nil {
		config.fileCollector(path)
	}

	// Convert to absolute path if it's not already
	absPath := path
	if !filepath.IsAbs(path) {
		// If this fails, just use the original path
		if abs, err := GetAbsolutePath(path); err == nil {
			absPath = abs
		} else {
			config.Logger.Printf("Warning: Could not convert %s to absolute path: %v\n", path, err)
		}
	}

	// Create a FileMeta and add it to the slice
	*files = append(*files, FileMeta{
		Path:    absPath,
		Content: string(content),
	})
}

// GatherProjectContextWithContext walks paths and gathers files into a slice of FileMeta.
// This version accepts a context.Context parameter for logging and correlation ID.
//
// This function now delegates to the concurrent implementation for improved performance
// on multi-core systems. The results are identical to the original sequential version.
func GatherProjectContextWithContext(ctx context.Context, paths []string, config *Config) ([]FileMeta, int, error) {
	// If context is nil, create a background context
	if ctx == nil {
		ctx = context.Background()
	}

	// Extract correlation ID from context for logging, if available
	correlationID := logutil.GetCorrelationID(ctx)
	if correlationID != "" {
		config.Logger.DebugContext(ctx, "Starting GatherProjectContext with correlation ID: %s", correlationID)
	}

	// Delegate to concurrent implementation with default configuration
	concCfg := NewDefaultConcurrentConfig(ctx)
	return GatherProjectContextConcurrent(ctx, paths, config, concCfg)
}

// GatherProjectContext is a backward-compatible version that doesn't require a context.
// It creates a background context and calls the context-aware version.
func GatherProjectContext(paths []string, config *Config) ([]FileMeta, int, error) {
	ctx := context.Background()
	return GatherProjectContextWithContext(ctx, paths, config)
}

// FilterProjectFiles applies pure filtering logic to a list of discovered file paths.
// This function demonstrates the use of extracted pure functions and can be used
// when you already have a list of file paths and want to apply filtering rules.
func FilterProjectFiles(paths []string, config *Config) []string {
	filterOpts := CreateFilteringOptions(config)
	return FilterFiles(paths, filterOpts)
}

// AnalyzeFileContent provides detailed statistics about file content using pure functions.
// This demonstrates the enhanced statistics calculation capabilities.
func AnalyzeFileContent(content string) FileStatistics {
	return CalculateFileStatistics(content)
}

// CalculateStatistics calculates basic string stats.
// This function maintains backward compatibility with the original token counting algorithm.
func CalculateStatistics(content string) (charCount, lineCount, tokenCount int) {
	charCount = len(content)
	lineCount = strings.Count(content, "\n") + 1
	tokenCount = estimateTokenCount(content) // Use original algorithm for backward compatibility
	return charCount, lineCount, tokenCount
}

// estimateTokenCount counts tokens simply by whitespace boundaries.
// This is kept as a fallback method in case the API token counting fails.
func estimateTokenCount(text string) int {
	count := 0
	inToken := false
	for _, r := range text {
		if unicode.IsSpace(r) {
			if inToken {
				count++
				inToken = false
			}
		} else {
			inToken = true
		}
	}
	if inToken {
		count++
	}
	return count
}
