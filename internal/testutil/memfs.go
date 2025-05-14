// Package testutil provides testing utilities for the entire codebase
package testutil

import (
	"bytes"
	"context"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemFile represents a file in the in-memory filesystem
type MemFile struct {
	Data    []byte
	ModTime time.Time
	Mode    fs.FileMode
}

// MemDirectory represents a directory in the in-memory filesystem
type MemDirectory struct {
	ModTime time.Time
	Mode    fs.FileMode
}

// MemFS implements the FilesystemIO interface using an in-memory representation
type MemFS struct {
	mutex       sync.RWMutex
	files       map[string]*MemFile
	directories map[string]*MemDirectory
}

// fileError is a simple error type for filesystem operations
type fileError struct {
	msg string
}

// Error implements the error interface
func (e *fileError) Error() string {
	return e.msg
}

// NewMemFS creates a new in-memory filesystem
func NewMemFS() *MemFS {
	return &MemFS{
		files:       make(map[string]*MemFile),
		directories: make(map[string]*MemDirectory),
	}
}

// ReadFile reads the entire file at the specified path
func (m *MemFS) ReadFile(path string) ([]byte, error) {
	return m.ReadFileWithContext(context.Background(), path)
}

// ReadFileWithContext reads the entire file at the specified path with context
func (m *MemFS) ReadFileWithContext(ctx context.Context, path string) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Normalize path
	path = filepath.Clean(path)

	// Check if file exists
	file, ok := m.files[path]
	if !ok {
		return nil, &fileError{msg: "file not found: " + path}
	}

	// Return a copy of the data to prevent modification
	result := make([]byte, len(file.Data))
	copy(result, file.Data)
	return result, nil
}

// WriteFile writes data to the file at the specified path
func (m *MemFS) WriteFile(path string, data []byte, perm int) error {
	return m.WriteFileWithContext(context.Background(), path, data, perm)
}

// WriteFileWithContext writes data to the file at the specified path with context
func (m *MemFS) WriteFileWithContext(ctx context.Context, path string, data []byte, perm int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Normalize path
	path = filepath.Clean(path)

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		if _, ok := m.directories[dir]; !ok {
			return &fileError{msg: "directory does not exist: " + dir}
		}
	}

	// Create or update file
	m.files[path] = &MemFile{
		Data:    append([]byte{}, data...), // Make a copy of the data
		ModTime: time.Now(),
		Mode:    fs.FileMode(perm),
	}

	return nil
}

// MkdirAll creates a directory named path, along with any necessary parents
func (m *MemFS) MkdirAll(path string, perm int) error {
	return m.MkdirAllWithContext(context.Background(), path, perm)
}

// MkdirAllWithContext creates a directory named path, along with any necessary parents with context
func (m *MemFS) MkdirAllWithContext(ctx context.Context, path string, perm int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Normalize path
	path = filepath.Clean(path)
	if path == "." || path == "/" {
		// Root directory always exists
		return nil
	}

	// Create all parent directories
	segments := strings.Split(path, string(filepath.Separator))
	currentPath := ""

	for _, segment := range segments {
		if segment == "" {
			// Skip empty segments (can happen with leading or trailing slashes)
			continue
		}

		// Build up path incrementally
		if currentPath == "" {
			// Handle Windows absolute paths that start with a drive letter
			if strings.HasSuffix(segment, ":") {
				currentPath = segment + string(filepath.Separator)
				continue
			}
			currentPath = segment
		} else {
			currentPath = filepath.Join(currentPath, segment)
		}

		// Create directory if it doesn't exist
		if _, ok := m.directories[currentPath]; !ok {
			m.directories[currentPath] = &MemDirectory{
				ModTime: time.Now(),
				Mode:    fs.FileMode(perm),
			}
		}
	}

	return nil
}

// RemoveAll removes path and any children it contains
func (m *MemFS) RemoveAll(path string) error {
	return m.RemoveAllWithContext(context.Background(), path)
}

// RemoveAllWithContext removes path and any children it contains with context
func (m *MemFS) RemoveAllWithContext(ctx context.Context, path string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Normalize path
	path = filepath.Clean(path)

	// Remove directory itself
	delete(m.directories, path)

	// Remove all files and subdirectories
	for filePath := range m.files {
		if isChildPath(filePath, path) {
			delete(m.files, filePath)
		}
	}

	for dirPath := range m.directories {
		if isChildPath(dirPath, path) {
			delete(m.directories, dirPath)
		}
	}

	return nil
}

// Stat returns a FileInfo describing the named file
func (m *MemFS) Stat(path string) (bool, error) {
	return m.StatWithContext(context.Background(), path)
}

// StatWithContext returns a FileInfo describing the named file with context
func (m *MemFS) StatWithContext(ctx context.Context, path string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Normalize path
	path = filepath.Clean(path)

	// Special case for root directory
	if path == "." || path == "/" {
		return true, nil
	}

	// Check if it's a file
	if _, ok := m.files[path]; ok {
		return true, nil
	}

	// Check if it's a directory
	if _, ok := m.directories[path]; ok {
		return true, nil
	}

	return false, &fileError{msg: "file or directory not found: " + path}
}

// List returns a list of files and directories in the given path
func (m *MemFS) List(path string) ([]string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Normalize path
	path = filepath.Clean(path)
	if path == "." {
		path = ""
	}

	// Check if directory exists (unless it's the root)
	if path != "" {
		if _, ok := m.directories[path]; !ok {
			return nil, &fileError{msg: "directory not found: " + path}
		}
	}

	// Find all direct children
	children := make(map[string]bool)

	// Add files
	for filePath := range m.files {
		dir, file := filepath.Split(filePath)
		dir = filepath.Clean(dir)

		// Check if this file is a direct child of the specified path
		if (path == "" && dir == ".") || dir == path {
			children[file] = true
		}
	}

	// Add directories
	for dirPath := range m.directories {
		// Skip the directory itself
		if dirPath == path {
			continue
		}

		parent := filepath.Dir(dirPath)
		if parent == "." {
			parent = ""
		}

		// Check if this directory is a direct child of the specified path
		if parent == path {
			children[filepath.Base(dirPath)] = true
		}
	}

	// Convert to slice and sort
	result := make([]string, 0, len(children))
	for child := range children {
		result = append(result, child)
	}
	sort.Strings(result)

	return result, nil
}

// GetFileContent returns the content of all files as a map for testing purposes
func (m *MemFS) GetFileContent() map[string][]byte {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string][]byte)
	for path, file := range m.files {
		data := make([]byte, len(file.Data))
		copy(data, file.Data)
		result[path] = data
	}

	return result
}

// GetDirectories returns a list of all directories for testing purposes
func (m *MemFS) GetDirectories() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make([]string, 0, len(m.directories))
	for dir := range m.directories {
		result = append(result, dir)
	}
	sort.Strings(result)

	return result
}

// isChildPath checks if child is a child path of parent
func isChildPath(child, parent string) bool {
	if parent == "." || parent == "/" {
		return true
	}

	// Ensure the parent path has a trailing separator
	if !strings.HasSuffix(parent, string(filepath.Separator)) {
		parent += string(filepath.Separator)
	}

	return strings.HasPrefix(child, parent)
}

// FileExists checks if a file exists
func (m *MemFS) FileExists(path string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Normalize path
	path = filepath.Clean(path)

	_, ok := m.files[path]
	return ok
}

// DirExists checks if a directory exists
func (m *MemFS) DirExists(path string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Normalize path
	path = filepath.Clean(path)

	// Root directory always exists
	if path == "." || path == "/" {
		return true
	}

	_, ok := m.directories[path]
	return ok
}

// Path operations for filesystem-agnostic code

// Join joins path elements into a single path
func (m *MemFS) Join(elem ...string) string {
	return filepath.Join(elem...)
}

// Base returns the last element of path
func (m *MemFS) Base(path string) string {
	return filepath.Base(path)
}

// Dir returns all but the last element of path
func (m *MemFS) Dir(path string) string {
	return filepath.Dir(path)
}

// SearchFiles searches for files matching the given pattern
func (m *MemFS) SearchFiles(pattern string) ([]string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var matches []string

	// Use filepath.Match for simple patterns
	for filePath := range m.files {
		match, err := filepath.Match(pattern, filePath)
		if err != nil {
			return nil, err
		}
		if match {
			matches = append(matches, filePath)
		}
	}

	sort.Strings(matches)
	return matches, nil
}

// FindInFiles searches for content in files
func (m *MemFS) FindInFiles(content string, filePatterns ...string) (map[string][]int, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	results := make(map[string][]int)

	// Process all files or only those matching patterns
	filesToSearch := make(map[string]bool)

	if len(filePatterns) == 0 {
		// Search all files
		for path := range m.files {
			filesToSearch[path] = true
		}
	} else {
		// Search only files matching patterns
		for _, pattern := range filePatterns {
			for path := range m.files {
				match, err := filepath.Match(pattern, filepath.Base(path))
				if err != nil {
					return nil, err
				}
				if match {
					filesToSearch[path] = true
				}
			}
		}
	}

	// Search for content in the selected files
	for path := range filesToSearch {
		file := m.files[path]

		// Find all instances of content
		var lineMatches []int

		lines := bytes.Split(file.Data, []byte("\n"))
		for i, line := range lines {
			// Case-sensitive search for exact content
			if bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(content))) {
				lineMatches = append(lineMatches, i+1) // 1-based line numbers
			}
		}

		if len(lineMatches) > 0 {
			results[path] = lineMatches
		}
	}

	return results, nil
}
