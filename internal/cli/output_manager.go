// Package cli provides output directory management for the thinktank tool
package cli

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/phrazzld/thinktank/internal/logutil"
)

// OutputManager handles intelligent output directory naming and management
type OutputManager struct {
	logger logutil.LoggerInterface
	rand   *rand.Rand
}

// NewOutputManager creates a new output manager instance
func NewOutputManager(logger logutil.LoggerInterface) *OutputManager {
	if logger == nil {
		logger = logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)
	}

	return &OutputManager{
		logger: logger,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Global counter for ensuring uniqueness across concurrent operations
var incrementalCounter uint32

// GenerateTimestampedDirName generates a unique directory name in the format:
// thinktank_YYYYMMDD_HHMMSS_NNNNNNNNN
// where NNNNNNNNN combines nanoseconds, random number, and counter for uniqueness
func (om *OutputManager) GenerateTimestampedDirName() string {
	// Get current time
	now := time.Now()

	// Generate timestamp in format YYYYMMDD_HHMMSS
	timestamp := now.Format("20060102_150405")

	// Use multiple strategies to ensure uniqueness:
	// 1. Nanoseconds from the current time (0-999999999)
	// 2. Random number (0-999)
	// 3. Incremental counter (0-999999)

	// Extract nanoseconds (last 3 digits)
	nanos := now.Nanosecond() % 1000

	// Generate a random number
	randNum := om.rand.Intn(1000)

	// Increment the counter atomically (thread-safe)
	counter := atomic.AddUint32(&incrementalCounter, 1) % 1000

	// Combine all three components for a truly unique value
	// This gives us a billion possibilities (1000 × 1000 × 1000) within the same second
	uniqueNum := (nanos * 1000000) + (randNum * 1000) + int(counter)

	// Combine with prefix and format with leading zeros (9 digits)
	return fmt.Sprintf("thinktank_%s_%09d", timestamp, uniqueNum)
}

// CreateOutputDirectory creates an output directory with collision detection
// If basePath is empty, uses current working directory
func (om *OutputManager) CreateOutputDirectory(basePath string, permissions os.FileMode) (string, error) {
	if basePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			om.logger.Printf("Error getting current working directory: %v", err)
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
		basePath = cwd
	}

	// Generate unique directory name
	dirName := om.GenerateTimestampedDirName()
	fullPath := filepath.Join(basePath, dirName)

	// Handle collision detection with automatic incrementing
	attempts := 0
	maxAttempts := 10
	originalName := dirName

	for attempts < maxAttempts {
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			// Path doesn't exist, we can use it
			break
		}

		// Path exists, generate a new name
		attempts++
		dirName = fmt.Sprintf("%s_retry%d", originalName, attempts)
		fullPath = filepath.Join(basePath, dirName)

		om.logger.Printf("Directory collision detected, trying: %s", dirName)
	}

	if attempts >= maxAttempts {
		return "", fmt.Errorf("failed to find unique directory name after %d attempts", maxAttempts)
	}

	// Create the directory with proper permissions (0755)
	if err := os.MkdirAll(fullPath, permissions); err != nil {
		om.logger.Printf("Error creating output directory %s: %v", fullPath, err)
		return "", fmt.Errorf("failed to create output directory %s: %w", fullPath, err)
	}

	om.logger.Printf("Created output directory: %s", fullPath)
	return fullPath, nil
}

// CleanupOldDirectories removes output directories older than the specified duration
// It only removes directories that match the thinktank naming pattern
func (om *OutputManager) CleanupOldDirectories(basePath string, olderThan time.Duration) error {
	if basePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		basePath = cwd
	}

	cutoffTime := time.Now().Add(-olderThan)
	om.logger.Printf("Cleaning up directories older than %v (before %v)", olderThan, cutoffTime.Format(time.RFC3339))

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}

	cleanedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Only process directories that match our naming pattern
		if !om.isThinktankOutputDir(entry.Name()) {
			continue
		}

		dirPath := filepath.Join(basePath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			om.logger.Printf("Warning: Cannot get info for directory %s: %v", dirPath, err)
			continue
		}

		// Check if directory is older than cutoff
		if info.ModTime().Before(cutoffTime) {
			if err := os.RemoveAll(dirPath); err != nil {
				om.logger.Printf("Warning: Failed to remove old directory %s: %v", dirPath, err)
				continue
			}

			om.logger.Printf("Removed old output directory: %s (last modified: %v)",
				entry.Name(), info.ModTime().Format(time.RFC3339))
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		om.logger.Printf("Cleanup completed: removed %d old directories", cleanedCount)
	} else {
		om.logger.Printf("Cleanup completed: no old directories found")
	}

	return nil
}

// isThinktankOutputDir checks if a directory name matches the thinktank output pattern
func (om *OutputManager) isThinktankOutputDir(name string) bool {
	// Check if it starts with "thinktank_" and has the expected format
	if !strings.HasPrefix(name, "thinktank_") {
		return false
	}

	// Split by underscores and check structure
	parts := strings.Split(name, "_")
	if len(parts) < 4 {
		return false
	}

	// Verify the date part (YYYYMMDD)
	dateStr := parts[1]
	if len(dateStr) != 8 {
		return false
	}

	// Verify the time part (HHMMSS)
	timeStr := parts[2]
	if len(timeStr) != 6 {
		return false
	}

	// Verify the unique part (NNNNNNNNN or NNNNNNNNN_retryN)
	uniquePart := parts[3]
	if strings.Contains(uniquePart, "_retry") {
		// Handle retry suffix
		retryParts := strings.Split(uniquePart, "_retry")
		if len(retryParts) != 2 || len(retryParts[0]) != 9 {
			return false
		}
	} else if len(uniquePart) != 9 {
		return false
	}

	return true
}

// GetOutputDirStats returns statistics about output directories in the base path
func (om *OutputManager) GetOutputDirStats(basePath string) (map[string]interface{}, error) {
	if basePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		basePath = cwd
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}

	stats := map[string]interface{}{
		"total_dirs":       0,
		"thinktank_dirs":   0,
		"oldest_dir":       "",
		"newest_dir":       "",
		"total_size_bytes": int64(0),
	}

	var oldestTime, newestTime time.Time
	var oldestDir, newestDir string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		stats["total_dirs"] = stats["total_dirs"].(int) + 1

		if om.isThinktankOutputDir(entry.Name()) {
			stats["thinktank_dirs"] = stats["thinktank_dirs"].(int) + 1

			info, err := entry.Info()
			if err != nil {
				continue
			}

			modTime := info.ModTime()
			if oldestTime.IsZero() || modTime.Before(oldestTime) {
				oldestTime = modTime
				oldestDir = entry.Name()
			}
			if newestTime.IsZero() || modTime.After(newestTime) {
				newestTime = modTime
				newestDir = entry.Name()
			}

			// Calculate directory size
			dirPath := filepath.Join(basePath, entry.Name())
			if size, err := om.calculateDirSize(dirPath); err == nil {
				stats["total_size_bytes"] = stats["total_size_bytes"].(int64) + size
			}
		}
	}

	stats["oldest_dir"] = oldestDir
	stats["newest_dir"] = newestDir

	return stats, nil
}

// calculateDirSize calculates the total size of a directory
func (om *OutputManager) calculateDirSize(dirPath string) (int64, error) {
	var size int64

	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors and continue
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return nil // Skip errors and continue
			}
			size += info.Size()
		}

		return nil
	})

	return size, err
}

// CleanupOldDirectoriesWithDefault performs cleanup with default 30-day retention
func (om *OutputManager) CleanupOldDirectoriesWithDefault(basePath string) error {
	return om.CleanupOldDirectories(basePath, 30*24*time.Hour)
}
