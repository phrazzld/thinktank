// Package cli provides context analysis capabilities for the thinktank tool
package cli

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/phrazzld/thinktank/internal/logutil"
)

// ComplexityLevel represents the complexity category of a project
type ComplexityLevel int

const (
	ComplexitySimple ComplexityLevel = iota
	ComplexityMedium
	ComplexityLarge
	ComplexityXLarge
)

// String returns the string representation of complexity level
func (c ComplexityLevel) String() string {
	switch c {
	case ComplexitySimple:
		return "Simple"
	case ComplexityMedium:
		return "Medium"
	case ComplexityLarge:
		return "Large"
	case ComplexityXLarge:
		return "XLarge"
	default:
		return "Unknown"
	}
}

// Complexity thresholds in tokens
const (
	simpleThreshold = 10_000  // < 10k tokens
	mediumThreshold = 50_000  // < 50k tokens
	largeThreshold  = 200_000 // < 200k tokens
	// XLarge is >= 200k tokens
)

// Token estimation constant - Claude uses approximately 4 characters per token
const avgCharsPerToken = 4

// AnalysisResult contains the results of complexity analysis
type AnalysisResult struct {
	TotalFiles      int64           `json:"total_files"`
	TotalLines      int64           `json:"total_lines"`
	TotalChars      int64           `json:"total_chars"`
	EstimatedTokens int64           `json:"estimated_tokens"`
	Complexity      ComplexityLevel `json:"complexity"`
	AnalysisTime    time.Duration   `json:"analysis_time_ns"`
	CacheHit        bool            `json:"cache_hit"`
}

// CacheEntry represents a cached analysis result with metadata
type CacheEntry struct {
	Result       AnalysisResult   `json:"result"`
	Timestamp    time.Time        `json:"timestamp"`
	PathHash     string           `json:"path_hash"`
	FileModTimes map[string]int64 `json:"file_mod_times"` // file path -> unix timestamp
}

// ContextAnalyzer provides complexity analysis with caching
type ContextAnalyzer struct {
	cache  map[string]*CacheEntry
	mu     sync.RWMutex
	logger logutil.LoggerInterface
}

// NewContextAnalyzer creates a new context analyzer instance
func NewContextAnalyzer(logger logutil.LoggerInterface) *ContextAnalyzer {
	if logger == nil {
		logger = logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)
	}

	return &ContextAnalyzer{
		cache:  make(map[string]*CacheEntry),
		logger: logger,
	}
}

// analyzeTaskComplexity analyzes the complexity of a target path
func (ca *ContextAnalyzer) analyzeTaskComplexity(targetPath string) (int64, error) {
	result, err := ca.AnalyzeComplexity(targetPath)
	if err != nil {
		return 0, err
	}
	return result.EstimatedTokens, nil
}

// AnalyzeComplexity performs comprehensive complexity analysis on the target path
func (ca *ContextAnalyzer) AnalyzeComplexity(targetPath string) (*AnalysisResult, error) {
	startTime := time.Now()

	// Convert to absolute path for consistent caching
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", targetPath, err)
	}

	// Check cache first
	if cached := ca.getCachedResult(absPath); cached != nil {
		cached.CacheHit = true
		cached.AnalysisTime = time.Since(startTime)
		return cached, nil
	}

	// Perform fresh analysis
	result, fileModTimes, err := ca.performAnalysis(absPath)
	if err != nil {
		return nil, err
	}

	result.AnalysisTime = time.Since(startTime)
	result.CacheHit = false

	// Cache the result with file modification times
	ca.cacheResultWithModTimes(absPath, result, fileModTimes)

	return result, nil
}

// getCachedResult retrieves cached analysis if valid
func (ca *ContextAnalyzer) getCachedResult(targetPath string) *AnalysisResult {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	cacheKey := ca.generateCacheKey(targetPath)
	entry, exists := ca.cache[cacheKey]
	if !exists {
		return nil
	}

	// Validate cache by checking file modification times
	if !ca.isCacheValid(targetPath, entry) {
		// Cache is stale, will be cleaned up later
		return nil
	}

	// Return a copy to avoid race conditions
	result := entry.Result
	return &result
}

// isCacheValid checks if cached entry is still valid
// It checks both TTL and file modification times for accuracy
func (ca *ContextAnalyzer) isCacheValid(targetPath string, entry *CacheEntry) bool {
	// Check TTL first for quick exit
	if time.Since(entry.Timestamp) >= 30*time.Second {
		return false
	}

	// Check if any tracked files have been modified
	for filePath, cachedModTime := range entry.FileModTimes {
		info, err := os.Stat(filePath)
		if err != nil {
			// File deleted or inaccessible - cache is invalid
			return false
		}

		if info.ModTime().Unix() != cachedModTime {
			// File has been modified - cache is invalid
			return false
		}
	}

	return true
}

// shouldAnalyzeFile determines if a file should be included in analysis
// This mirrors the logic in fileutil but focuses on performance
func (ca *ContextAnalyzer) shouldAnalyzeFile(path string) bool {
	base := filepath.Base(path)

	// Skip hidden files and directories
	if len(base) > 0 && base[0] == '.' {
		return false
	}

	// Skip .git directory contents
	if strings.Contains(path, ".git") {
		return false
	}

	// Skip common binary/non-code files by extension
	ext := filepath.Ext(path)
	skipExts := map[string]bool{
		".exe": true, ".bin": true, ".so": true, ".dll": true,
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".pdf": true, ".zip": true, ".tar": true, ".gz": true,
		".mp4": true, ".avi": true, ".mov": true,
		".o": true, ".obj": true, ".a": true, ".lib": true,
	}

	return !skipExts[ext]
}

// performAnalysis conducts the actual complexity analysis
func (ca *ContextAnalyzer) performAnalysis(targetPath string) (*AnalysisResult, map[string]int64, error) {
	// Verify target path exists
	if _, err := os.Stat(targetPath); err != nil {
		return nil, nil, fmt.Errorf("target path does not exist: %w", err)
	}

	var totalFiles, totalLines, totalChars int64
	fileModTimes := make(map[string]int64)

	err := filepath.WalkDir(targetPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			ca.logger.Printf("Warning: Error accessing path %s: %v", path, err)
			return nil // Continue walking despite errors
		}

		if d.IsDir() {
			return nil
		}

		if !ca.shouldAnalyzeFile(path) {
			return nil
		}

		// Get file info for modification time
		info, err := d.Info()
		if err != nil {
			ca.logger.Printf("Warning: Cannot get file info for %s: %v", path, err)
			return nil
		}

		// Store modification time for cache validation
		fileModTimes[path] = info.ModTime().Unix()

		// Read and analyze file efficiently
		if err := ca.analyzeFile(path, info, &totalFiles, &totalLines, &totalChars); err != nil {
			ca.logger.Printf("Warning: Cannot analyze file %s: %v", path, err)
			return nil // Continue despite individual file errors
		}

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to walk directory %s: %w", targetPath, err)
	}

	// Calculate token estimate and complexity
	estimatedTokens := ca.estimateTokens(totalChars)
	complexity := ca.categorizeComplexity(estimatedTokens)

	result := &AnalysisResult{
		TotalFiles:      totalFiles,
		TotalLines:      totalLines,
		TotalChars:      totalChars,
		EstimatedTokens: estimatedTokens,
		Complexity:      complexity,
	}

	return result, fileModTimes, nil
}

// analyzeFile efficiently analyzes a single file's metrics
func (ca *ContextAnalyzer) analyzeFile(path string, info os.FileInfo, totalFiles, totalLines, totalChars *int64) error {
	// Quick binary file detection
	if ca.isLikelyBinary(path, info) {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	// Streaming line and character counting to avoid loading entire file
	buf := make([]byte, 8192) // 8KB buffer for efficient I/O
	var lines, chars int64
	var lastByte byte

	for {
		n, err := file.Read(buf)
		if n > 0 {
			chars += int64(n)

			// Count newlines in buffer
			for i := 0; i < n; i++ {
				if buf[i] == '\n' {
					lines++
				}
			}
			lastByte = buf[n-1]
		}

		if err != nil {
			break
		}
	}

	// Add one more line if file doesn't end with newline
	if chars > 0 && lastByte != '\n' {
		lines++
	}

	*totalFiles++
	*totalLines += lines
	*totalChars += chars

	return nil
}

// isLikelyBinary provides fast binary file detection based on file characteristics
func (ca *ContextAnalyzer) isLikelyBinary(path string, info os.FileInfo) bool {
	// Size-based heuristic: very large files are likely binary
	if info.Size() > 50*1024*1024 { // > 50MB
		return true
	}

	// Extension-based detection
	ext := filepath.Ext(path)
	binaryExts := map[string]bool{
		".exe": true, ".bin": true, ".so": true, ".dll": true,
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".zip": true, ".tar": true, ".gz": true, ".bz2": true, ".rar": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mov": true, ".wav": true,
		".o": true, ".obj": true, ".a": true, ".lib": true,
	}

	return binaryExts[ext]
}

// estimateTokens converts character count to estimated token count
func (ca *ContextAnalyzer) estimateTokens(chars int64) int64 {
	return chars / avgCharsPerToken
}

// categorizeComplexity determines complexity level based on token count
func (ca *ContextAnalyzer) categorizeComplexity(tokens int64) ComplexityLevel {
	switch {
	case tokens < simpleThreshold:
		return ComplexitySimple
	case tokens < mediumThreshold:
		return ComplexityMedium
	case tokens < largeThreshold:
		return ComplexityLarge
	default:
		return ComplexityXLarge
	}
}

// generateCacheKey creates a unique cache key for a target path
func (ca *ContextAnalyzer) generateCacheKey(targetPath string) string {
	hash := md5.Sum([]byte(targetPath))
	return fmt.Sprintf("%x", hash)
}

// cacheResult stores analysis result in cache
func (ca *ContextAnalyzer) cacheResult(targetPath string, result *AnalysisResult) {
	ca.cacheResultWithModTimes(targetPath, result, nil)
}

// cacheResultWithModTimes stores analysis result in cache with file modification times
func (ca *ContextAnalyzer) cacheResultWithModTimes(targetPath string, result *AnalysisResult, fileModTimes map[string]int64) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	cacheKey := ca.generateCacheKey(targetPath)

	entry := &CacheEntry{
		Result:       *result,
		Timestamp:    time.Now(),
		PathHash:     cacheKey,
		FileModTimes: fileModTimes,
	}

	ca.cache[cacheKey] = entry
}

// ClearCache removes all cached entries
func (ca *ContextAnalyzer) ClearCache() {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.cache = make(map[string]*CacheEntry)
}

// GetCacheStats returns cache statistics for monitoring
func (ca *ContextAnalyzer) GetCacheStats() map[string]interface{} {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	return map[string]interface{}{
		"entries":               len(ca.cache),
		"memory_estimate_bytes": len(ca.cache) * 1024, // Rough estimate
	}
}

// SerializeCache exports cache to JSON for persistence
func (ca *ContextAnalyzer) SerializeCache() ([]byte, error) {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	return json.Marshal(ca.cache)
}

// DeserializeCache imports cache from JSON
func (ca *ContextAnalyzer) DeserializeCache(data []byte) error {
	var cache map[string]*CacheEntry
	if err := json.Unmarshal(data, &cache); err != nil {
		return err
	}

	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.cache = cache

	return nil
}

// AnalyzeTaskComplexityForModelSelection provides a convenience method for model selection
// It returns the estimated token count for use with SelectOptimalModel
func AnalyzeTaskComplexityForModelSelection(targetPath string) (int64, error) {
	analyzer := NewContextAnalyzer(nil) // Use default logger
	return analyzer.analyzeTaskComplexity(targetPath)
}

// GetComplexityAnalysis provides detailed complexity analysis for CLI reporting
func GetComplexityAnalysis(targetPath string) (*AnalysisResult, error) {
	analyzer := NewContextAnalyzer(nil) // Use default logger
	return analyzer.AnalyzeComplexity(targetPath)
}
