// internal/fileutil/fileutil.go
package fileutil

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
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
		stdLogger := log.New(os.Stderr, "[fileutil] ", log.LstdFlags)
		logger = logutil.NewStdLoggerAdapter(stdLogger)
	}

	cfg := &Config{
		Verbose:      verbose,
		Format:       format,
		Logger:       logger,
		GitAvailable: gitAvailable,
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
	// Always ignore .git directory contents implicitly
	if base == ".git" || strings.Contains(path, string(filepath.Separator)+".git"+string(filepath.Separator)) {
		return true
	}

	// Use git check-ignore if available
	if config.GitAvailable {
		dir := filepath.Dir(path)
		// Check if the directory is actually a git repo first
		gitRepoCheck := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
		if gitRepoCheck.Run() == nil { // If it is a git repo
			cmd := exec.Command("git", "-C", dir, "check-ignore", "-q", base)
			err := cmd.Run()
			if err == nil { // Exit code 0: file IS ignored
				config.Logger.Printf("Verbose: Git ignored: %s\n", path)
				return true
			}
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				// Exit code 1: file is NOT ignored
				// Continue to other checks like hidden files
			} else if err != nil {
				// Other errors running check-ignore, log it but fall back
				config.Logger.Printf("Verbose: Error running git check-ignore for %s: %v. Falling back.\n", path, err)
			}
		}
		// If not a git repo or check-ignore failed non-fatally, proceed to hidden check
	}

	// Fallback or additional check: Check if the file/directory itself is hidden
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

	content, err := os.ReadFile(path)
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
	config.Logger.Printf("Verbose: Processing file (%d/%d): %s\n", config.processedFiles, config.totalFiles, path)

	// If a file collector is set, call it
	if config.fileCollector != nil {
		config.fileCollector(path)
	}

	// Create a FileMeta and add it to the slice
	*files = append(*files, FileMeta{
		Path:    path,
		Content: string(content),
	})
}

// GatherProjectContext walks paths and gathers files into a slice of FileMeta.
func GatherProjectContext(paths []string, config *Config) ([]FileMeta, int, error) {
	var files []FileMeta

	config.processedFiles = 0
	config.totalFiles = 0

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			config.Logger.Printf("Warning: Cannot stat path %s: %v. Skipping.\n", p, err)
			continue
		}

		if info.IsDir() {
			// Walk the directory
			err := filepath.WalkDir(p, func(subPath string, d os.DirEntry, err error) error {
				if err != nil {
					config.Logger.Printf("Warning: Error accessing path %s during walk: %v\n", subPath, err)
					return err // Report error up
				}

				// Check if the directory itself should be skipped (e.g., .git, node_modules)
				if d.IsDir() {
					if isGitIgnored(subPath, config) || slices.Contains(config.ExcludeNames, d.Name()) {
						config.Logger.Printf("Verbose: Skipping directory: %s\n", subPath)
						return filepath.SkipDir // Skip this whole directory
					}
					return nil // Continue walking into directory
				}

				// It's a file, process it
				if !d.IsDir() {
					processFile(subPath, &files, config)
				}

				return nil // Continue walking
			})
			if err != nil {
				config.Logger.Printf("Error walking directory %s: %v\n", p, err)
				// Continue with other paths if possible
			}
		} else {
			// It's a single file
			processFile(p, &files, config)
		}
	}

	return files, config.processedFiles, nil
}

// CalculateStatistics calculates basic string stats.
func CalculateStatistics(content string) (charCount, lineCount, tokenCount int) {
	charCount = len(content)
	lineCount = strings.Count(content, "\n") + 1
	tokenCount = estimateTokenCount(content) // Fallback estimation
	return charCount, lineCount, tokenCount
}

// CalculateStatisticsWithTokenCounting calculates accurate statistics using Gemini's token counter.
func CalculateStatisticsWithTokenCounting(ctx context.Context, geminiClient gemini.Client, content string, logger logutil.LoggerInterface) (charCount, lineCount, tokenCount int) {
	charCount = len(content)
	lineCount = strings.Count(content, "\n") + 1

	// Use the Gemini API for accurate token counting
	if geminiClient != nil {
		tokenResult, err := geminiClient.CountTokens(ctx, content)
		if err != nil {
			// Log the error and fall back to estimation
			if logger != nil {
				logger.Warn("Failed to count tokens accurately: %v. Using estimation instead.", err)
			}
			tokenCount = estimateTokenCount(content)
		} else {
			tokenCount = int(tokenResult.Total)
			if logger != nil {
				logger.Debug("Accurate token count: %d tokens", tokenCount)
			}
		}
	} else {
		// Fall back to estimation if no client provided
		tokenCount = estimateTokenCount(content)
		if logger != nil {
			logger.Debug("Using estimated token count: %d tokens", tokenCount)
		}
	}

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
