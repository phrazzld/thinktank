// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// MockFileSystem implements the FileSystem interface for testing
type MockFileSystem struct {
	// Files maps file paths to their content
	Files map[string][]byte

	// FilePermissions tracks permissions for created files
	FilePermissions map[string]os.FileMode

	// Dirs tracks created directories
	Dirs map[string]os.FileMode

	// TempFiles tracks temporary files created
	TempFiles []*MockFile

	// Errors maps operations to errors to simulate failures
	Errors map[string]error

	// CallLog tracks all method calls for verification
	CallLog []string
}

// MockFile represents a mock file for testing
type MockFile struct {
	Name    string
	Content []byte
	Closed  bool
}

// Write implements io.Writer for MockFile
func (mf *MockFile) Write(p []byte) (n int, err error) {
	mf.Content = append(mf.Content, p...)
	return len(p), nil
}

// Close implements io.Closer for MockFile
func (mf *MockFile) Close() error {
	mf.Closed = true
	return nil
}

// NewMockFileSystem creates a new MockFileSystem for testing
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files:           make(map[string][]byte),
		FilePermissions: make(map[string]os.FileMode),
		Dirs:            make(map[string]os.FileMode),
		TempFiles:       make([]*MockFile, 0),
		Errors:          make(map[string]error),
		CallLog:         make([]string, 0),
	}
}

// CreateTemp creates a mock temporary file
func (mfs *MockFileSystem) CreateTemp(dir, pattern string) (*os.File, error) {
	mfs.CallLog = append(mfs.CallLog, fmt.Sprintf("CreateTemp(%s, %s)", dir, pattern))

	if err, exists := mfs.Errors["CreateTemp"]; exists {
		return nil, err
	}

	// Create a mock file
	mockFile := &MockFile{
		Name:    fmt.Sprintf("%s/temp_%d", dir, len(mfs.TempFiles)),
		Content: make([]byte, 0),
		Closed:  false,
	}
	mfs.TempFiles = append(mfs.TempFiles, mockFile)

	// Return a real *os.File - this is tricky, we'll need to create a real temp file
	// For now, return nil and let tests handle this differently
	return nil, fmt.Errorf("MockFileSystem.CreateTemp not fully implemented - use WriteFile for testing")
}

// WriteFile writes data to the mock file system
func (mfs *MockFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	mfs.CallLog = append(mfs.CallLog, fmt.Sprintf("WriteFile(%s, %d bytes, %v)", filename, len(data), perm))

	if err, exists := mfs.Errors["WriteFile"]; exists {
		return err
	}

	mfs.Files[filename] = make([]byte, len(data))
	copy(mfs.Files[filename], data)
	mfs.FilePermissions[filename] = perm
	return nil
}

// ReadFile reads data from the mock file system
func (mfs *MockFileSystem) ReadFile(filename string) ([]byte, error) {
	mfs.CallLog = append(mfs.CallLog, fmt.Sprintf("ReadFile(%s)", filename))

	if err, exists := mfs.Errors["ReadFile"]; exists {
		return nil, err
	}

	if data, exists := mfs.Files[filename]; exists {
		result := make([]byte, len(data))
		copy(result, data)
		return result, nil
	}

	return nil, fmt.Errorf("file not found: %s", filename)
}

// Remove removes a file from the mock file system
func (mfs *MockFileSystem) Remove(name string) error {
	mfs.CallLog = append(mfs.CallLog, fmt.Sprintf("Remove(%s)", name))

	if err, exists := mfs.Errors["Remove"]; exists {
		return err
	}

	delete(mfs.Files, name)
	delete(mfs.FilePermissions, name)
	delete(mfs.Dirs, name)
	return nil
}

// MkdirAll creates directories in the mock file system
func (mfs *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	mfs.CallLog = append(mfs.CallLog, fmt.Sprintf("MkdirAll(%s, %v)", path, perm))

	if err, exists := mfs.Errors["MkdirAll"]; exists {
		return err
	}

	mfs.Dirs[path] = perm
	return nil
}

// OpenFile opens a file in the mock file system
func (mfs *MockFileSystem) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	mfs.CallLog = append(mfs.CallLog, fmt.Sprintf("OpenFile(%s, %d, %v)", name, flag, perm))

	if err, exists := mfs.Errors["OpenFile"]; exists {
		return nil, err
	}

	// For mock purposes, we can't return a real *os.File easily
	// Tests should use WriteFile/ReadFile instead
	return nil, fmt.Errorf("MockFileSystem.OpenFile not fully implemented - use WriteFile/ReadFile for testing")
}

// MockExitHandler implements the ExitHandler interface for testing
type MockExitHandler struct {
	// ExitCodes tracks all exit codes that would have been called
	ExitCodes []int

	// Errors tracks all errors that were handled
	Errors []error

	// CallLog tracks all method calls for verification
	CallLog []string
}

// NewMockExitHandler creates a new MockExitHandler for testing
func NewMockExitHandler() *MockExitHandler {
	return &MockExitHandler{
		ExitCodes: make([]int, 0),
		Errors:    make([]error, 0),
		CallLog:   make([]string, 0),
	}
}

// Exit records the exit code instead of actually exiting
func (meh *MockExitHandler) Exit(code int) {
	meh.CallLog = append(meh.CallLog, fmt.Sprintf("Exit(%d)", code))
	meh.ExitCodes = append(meh.ExitCodes, code)
	// Don't actually exit in tests!
}

// HandleError records the error and determines exit code without exiting
func (meh *MockExitHandler) HandleError(ctx context.Context, err error, logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, operation string) {
	meh.CallLog = append(meh.CallLog, fmt.Sprintf("HandleError(%v, %s)", err, operation))
	meh.Errors = append(meh.Errors, err)

	// Determine exit code like the real handler but don't exit
	exitCode := getExitCodeFromError(err)
	meh.ExitCodes = append(meh.ExitCodes, exitCode)
}

// LastExitCode returns the most recent exit code, or -1 if none
func (meh *MockExitHandler) LastExitCode() int {
	if len(meh.ExitCodes) == 0 {
		return -1
	}
	return meh.ExitCodes[len(meh.ExitCodes)-1]
}

// WasCalled returns true if any method was called
func (meh *MockExitHandler) WasCalled() bool {
	return len(meh.CallLog) > 0
}

// GetFilePermission returns the permission for a specific file
func (mfs *MockFileSystem) GetFilePermission(filename string) (os.FileMode, bool) {
	perm, exists := mfs.FilePermissions[filename]
	return perm, exists
}

// GetDirPermission returns the permission for a specific directory
func (mfs *MockFileSystem) GetDirPermission(dirname string) (os.FileMode, bool) {
	perm, exists := mfs.Dirs[dirname]
	return perm, exists
}

// GetAllFilePermissions returns a copy of all file permissions for testing
func (mfs *MockFileSystem) GetAllFilePermissions() map[string]os.FileMode {
	result := make(map[string]os.FileMode)
	for path, perm := range mfs.FilePermissions {
		result[path] = perm
	}
	return result
}

// GetAllDirPermissions returns a copy of all directory permissions for testing
func (mfs *MockFileSystem) GetAllDirPermissions() map[string]os.FileMode {
	result := make(map[string]os.FileMode)
	for path, perm := range mfs.Dirs {
		result[path] = perm
	}
	return result
}
