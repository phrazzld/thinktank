// Package testutil provides testing utilities for the entire codebase
package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RealFS implements the FilesystemIO interface using the actual filesystem
type RealFS struct{}

// NewRealFS creates a new RealFS that operates on the actual filesystem
func NewRealFS() *RealFS {
	return &RealFS{}
}

// ReadFile reads the entire file at the specified path
func (fs *RealFS) ReadFile(path string) ([]byte, error) {
	return fs.ReadFileWithContext(context.Background(), path)
}

// ReadFileWithContext reads the entire file at the specified path with context
func (fs *RealFS) ReadFileWithContext(ctx context.Context, path string) ([]byte, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	//nolint:gosec // G304: Test filesystem wrapper implementing interface
	return os.ReadFile(path)
}

// WriteFile writes data to the file at the specified path
func (fs *RealFS) WriteFile(path string, data []byte, perm int) error {
	return fs.WriteFileWithContext(context.Background(), path, data, perm)
}

// WriteFileWithContext writes data to the file at the specified path with context
func (fs *RealFS) WriteFileWithContext(ctx context.Context, path string, data []byte, perm int) error {
	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}
	// Validate permission value is non-negative and within uint32 range
	if perm < 0 || perm > 0777 {
		return fmt.Errorf("invalid file permission: %d", perm)
	}

	return os.WriteFile(path, data, os.FileMode(perm))
}

// MkdirAll creates a directory named path, along with any necessary parents
func (fs *RealFS) MkdirAll(path string, perm int) error {
	return fs.MkdirAllWithContext(context.Background(), path, perm)
}

// MkdirAllWithContext creates a directory named path, along with any necessary parents with context
func (fs *RealFS) MkdirAllWithContext(ctx context.Context, path string, perm int) error {
	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}
	// Validate permission value is non-negative and within uint32 range
	if perm < 0 || perm > 0777 {
		return fmt.Errorf("invalid directory permission: %d", perm)
	}

	return os.MkdirAll(path, os.FileMode(perm))
}

// RemoveAll removes path and any children it contains
func (fs *RealFS) RemoveAll(path string) error {
	return fs.RemoveAllWithContext(context.Background(), path)
}

// RemoveAllWithContext removes path and any children it contains with context
func (fs *RealFS) RemoveAllWithContext(ctx context.Context, path string) error {
	// Check for context cancellation
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return os.RemoveAll(path)
}

// Stat returns a FileInfo describing the named file
func (fs *RealFS) Stat(path string) (bool, error) {
	return fs.StatWithContext(context.Background(), path)
}

// StatWithContext returns a FileInfo describing the named file with context
func (fs *RealFS) StatWithContext(ctx context.Context, path string) (bool, error) {
	// Check for context cancellation
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// List returns a list of files and directories in the given path
func (fs *RealFS) List(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry.Name())
	}
	sort.Strings(result)
	return result, nil
}

// FileExists checks if a file exists and is not a directory
func (fs *RealFS) FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a directory exists
func (fs *RealFS) DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Path operations for filesystem-agnostic code

// Join joins path elements into a single path
func (fs *RealFS) Join(elem ...string) string {
	return filepath.Join(elem...)
}

// Base returns the last element of path
func (fs *RealFS) Base(path string) string {
	return filepath.Base(path)
}

// Dir returns all but the last element of path
func (fs *RealFS) Dir(path string) string {
	return filepath.Dir(path)
}

// SearchFiles searches for files matching the given pattern
func (fs *RealFS) SearchFiles(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// FindInFiles searches for content in files
func (fs *RealFS) FindInFiles(content string, filePatterns ...string) (map[string][]int, error) {
	results := make(map[string][]int)

	// Process all files or only those matching patterns
	var filesToSearch []string
	var err error

	if len(filePatterns) == 0 {
		// This would search the entire filesystem, which is not practical
		// Instead, search only in the current directory
		filesToSearch, err = filepath.Glob("*")
		if err != nil {
			return nil, err
		}
	} else {
		// Search only files matching patterns
		for _, pattern := range filePatterns {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, err
			}
			filesToSearch = append(filesToSearch, matches...)
		}
	}

	// Search for content in the selected files
	for _, path := range filesToSearch {
		// Skip directories
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}

		// Read file
		//nolint:gosec // G304: File search function using glob-validated paths
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		// Find matches
		lines := strings.Split(string(data), "\n")
		var lineMatches []int

		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), strings.ToLower(content)) {
				lineMatches = append(lineMatches, i+1) // 1-based line numbers
			}
		}

		if len(lineMatches) > 0 {
			results[path] = lineMatches
		}
	}

	return results, nil
}
