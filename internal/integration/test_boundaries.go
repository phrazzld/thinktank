// Package integration provides integration tests for the thinktank package
package integration

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/phrazzld/thinktank/internal/llm"
)

// ExternalAPICaller represents the external boundary for API calls to LLM providers
// This interface focuses only on the actual HTTP communication with external services
type ExternalAPICaller interface {
	// CallLLMAPI makes an HTTP request to an LLM API with the given data
	// This abstracts the actual HTTP call, allowing it to be mocked for tests
	CallLLMAPI(ctx context.Context, modelName, prompt string, params map[string]interface{}) (*llm.ProviderResult, error)
}

// MockExternalAPICaller implements ExternalAPICaller for testing
type MockExternalAPICaller struct {
	// CallLLMAPIFunc allows customizing the behavior of CallLLMAPI in tests
	CallLLMAPIFunc func(ctx context.Context, modelName, prompt string, params map[string]interface{}) (*llm.ProviderResult, error)
}

// CallLLMAPI implements the ExternalAPICaller interface
func (m *MockExternalAPICaller) CallLLMAPI(ctx context.Context, modelName, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	if m.CallLLMAPIFunc != nil {
		return m.CallLLMAPIFunc(ctx, modelName, prompt, params)
	}
	// Default implementation returns a simple response
	return &llm.ProviderResult{
		Content:      "Mock response for model: " + modelName,
		FinishReason: "stop",
	}, nil
}

// FilesystemIO represents the external boundary for filesystem operations
// This interface focuses only on the actual file I/O operations
type FilesystemIO interface {
	// ReadFile reads the entire file at the specified path
	ReadFile(path string) ([]byte, error)

	// ReadFileWithContext reads the entire file at the specified path with context
	ReadFileWithContext(ctx context.Context, path string) ([]byte, error)

	// WriteFile writes data to the file at the specified path
	WriteFile(path string, data []byte, perm int) error

	// WriteFileWithContext writes data to the file at the specified path with context
	WriteFileWithContext(ctx context.Context, path string, data []byte, perm int) error

	// MkdirAll creates a directory named path, along with any necessary parents
	MkdirAll(path string, perm int) error

	// MkdirAllWithContext creates a directory named path, along with any necessary parents with context
	MkdirAllWithContext(ctx context.Context, path string, perm int) error

	// RemoveAll removes path and any children it contains
	RemoveAll(path string) error

	// RemoveAllWithContext removes path and any children it contains with context
	RemoveAllWithContext(ctx context.Context, path string) error

	// Stat returns a FileInfo describing the named file
	Stat(path string) (bool, error)

	// StatWithContext returns a FileInfo describing the named file with context
	StatWithContext(ctx context.Context, path string) (bool, error)
}

// MockFilesystemIO implements FilesystemIO for testing
type MockFilesystemIO struct {
	// ReadFileFunc allows customizing the behavior of ReadFile in tests
	ReadFileFunc func(path string) ([]byte, error)

	// ReadFileWithContextFunc allows customizing the behavior of ReadFileWithContext in tests
	ReadFileWithContextFunc func(ctx context.Context, path string) ([]byte, error)

	// WriteFileFunc allows customizing the behavior of WriteFile in tests
	WriteFileFunc func(path string, data []byte, perm int) error

	// WriteFileWithContextFunc allows customizing the behavior of WriteFileWithContext in tests
	WriteFileWithContextFunc func(ctx context.Context, path string, data []byte, perm int) error

	// MkdirAllFunc allows customizing the behavior of MkdirAll in tests
	MkdirAllFunc func(path string, perm int) error

	// MkdirAllWithContextFunc allows customizing the behavior of MkdirAllWithContext in tests
	MkdirAllWithContextFunc func(ctx context.Context, path string, perm int) error

	// RemoveAllFunc allows customizing the behavior of RemoveAll in tests
	RemoveAllFunc func(path string) error

	// RemoveAllWithContextFunc allows customizing the behavior of RemoveAllWithContext in tests
	RemoveAllWithContextFunc func(ctx context.Context, path string) error

	// StatFunc allows customizing the behavior of Stat in tests
	StatFunc func(path string) (bool, error)

	// StatWithContextFunc allows customizing the behavior of StatWithContext in tests
	StatWithContextFunc func(ctx context.Context, path string) (bool, error)

	// fileContents stores file contents for in-memory simulation
	// Access must be synchronized via mutex
	fileContents map[string][]byte

	// createdDirs tracks directories that have been created
	// Access must be synchronized via mutex
	createdDirs map[string]bool

	// mutex for thread-safe access to maps
	// Use RWMutex for better read performance
	mutex sync.RWMutex
}

// NewMockFilesystemIO creates a new MockFilesystemIO with default implementations
func NewMockFilesystemIO() *MockFilesystemIO {
	fileContents := make(map[string][]byte)
	createdDirs := make(map[string]bool)

	m := &MockFilesystemIO{
		fileContents: fileContents,
		createdDirs:  createdDirs,
	}

	// Default implementation of ReadFile
	m.ReadFileFunc = func(path string) ([]byte, error) {
		// Normalize path
		path = filepath.Clean(path)

		// Lock for reading
		m.mutex.RLock()
		content, ok := m.fileContents[path]
		m.mutex.RUnlock()

		if ok {
			return content, nil
		}
		return nil, &mockFileError{msg: "file not found: " + path}
	}

	// Default implementation of ReadFileWithContext (calls the non-context version)
	m.ReadFileWithContextFunc = func(ctx context.Context, path string) ([]byte, error) {
		return m.ReadFileFunc(path)
	}

	// Default implementation of WriteFile
	m.WriteFileFunc = func(path string, data []byte, perm int) error {
		// Normalize path
		path = filepath.Clean(path)

		// Check if parent directory exists (need read lock)
		dir := filepath.Dir(path)
		if dir != "." {
			m.mutex.RLock()
			dirExists := m.createdDirs[dir]
			m.mutex.RUnlock()

			if !dirExists {
				return &mockFileError{msg: "directory does not exist: " + dir}
			}
		}

		// Store file content (need write lock)
		m.mutex.Lock()
		m.fileContents[path] = data
		m.mutex.Unlock()

		return nil
	}

	// Default implementation of WriteFileWithContext (calls the non-context version)
	m.WriteFileWithContextFunc = func(ctx context.Context, path string, data []byte, perm int) error {
		return m.WriteFileFunc(path, data, perm)
	}

	// Default implementation of MkdirAll
	m.MkdirAllFunc = func(path string, perm int) error {
		// Normalize path
		path = filepath.Clean(path)

		// Mark directory as created (need write lock)
		m.mutex.Lock()
		m.createdDirs[path] = true
		m.mutex.Unlock()

		return nil
	}

	// Default implementation of MkdirAllWithContext (calls the non-context version)
	m.MkdirAllWithContextFunc = func(ctx context.Context, path string, perm int) error {
		return m.MkdirAllFunc(path, perm)
	}

	// Default implementation of RemoveAll
	m.RemoveAllFunc = func(path string) error {
		// Normalize path
		path = filepath.Clean(path)

		// Need write lock for all operations
		m.mutex.Lock()
		defer m.mutex.Unlock()

		// Remove directory
		delete(m.createdDirs, path)

		// Remove all files in this directory
		// Create a list of files to delete to avoid modifying map during iteration
		var toDelete []string
		for filePath := range m.fileContents {
			if filepath.Dir(filePath) == path {
				toDelete = append(toDelete, filePath)
			}
		}

		// Delete the files
		for _, filePath := range toDelete {
			delete(m.fileContents, filePath)
		}

		return nil
	}

	// Default implementation of RemoveAllWithContext (calls the non-context version)
	m.RemoveAllWithContextFunc = func(ctx context.Context, path string) error {
		return m.RemoveAllFunc(path)
	}

	// Default implementation of Stat
	m.StatFunc = func(path string) (bool, error) {
		// Normalize path
		path = filepath.Clean(path)

		// Check if file or directory exists (need read lock)
		m.mutex.RLock()
		_, fileExists := m.fileContents[path]
		dirExists := m.createdDirs[path]
		m.mutex.RUnlock()

		if fileExists || dirExists {
			return true, nil
		}

		return false, &mockFileError{msg: "file or directory not found: " + path}
	}

	// Default implementation of StatWithContext (calls the non-context version)
	m.StatWithContextFunc = func(ctx context.Context, path string) (bool, error) {
		return m.StatFunc(path)
	}

	return m
}

// ReadFile implements the FilesystemIO interface
func (m *MockFilesystemIO) ReadFile(path string) ([]byte, error) {
	// No locking here - the func handles its own locking
	return m.ReadFileFunc(path)
}

// ReadFileWithContext implements the FilesystemIO interface
func (m *MockFilesystemIO) ReadFileWithContext(ctx context.Context, path string) ([]byte, error) {
	// No locking here - the func handles its own locking
	return m.ReadFileWithContextFunc(ctx, path)
}

// WriteFile implements the FilesystemIO interface
func (m *MockFilesystemIO) WriteFile(path string, data []byte, perm int) error {
	// No locking here - the func handles its own locking
	return m.WriteFileFunc(path, data, perm)
}

// WriteFileWithContext implements the FilesystemIO interface
func (m *MockFilesystemIO) WriteFileWithContext(ctx context.Context, path string, data []byte, perm int) error {
	// No locking here - the func handles its own locking
	return m.WriteFileWithContextFunc(ctx, path, data, perm)
}

// MkdirAll implements the FilesystemIO interface
func (m *MockFilesystemIO) MkdirAll(path string, perm int) error {
	// No locking here - the func handles its own locking
	return m.MkdirAllFunc(path, perm)
}

// MkdirAllWithContext implements the FilesystemIO interface
func (m *MockFilesystemIO) MkdirAllWithContext(ctx context.Context, path string, perm int) error {
	// No locking here - the func handles its own locking
	return m.MkdirAllWithContextFunc(ctx, path, perm)
}

// RemoveAll implements the FilesystemIO interface
func (m *MockFilesystemIO) RemoveAll(path string) error {
	// No locking here - the func handles its own locking
	return m.RemoveAllFunc(path)
}

// RemoveAllWithContext implements the FilesystemIO interface
func (m *MockFilesystemIO) RemoveAllWithContext(ctx context.Context, path string) error {
	// No locking here - the func handles its own locking
	return m.RemoveAllWithContextFunc(ctx, path)
}

// Stat implements the FilesystemIO interface
func (m *MockFilesystemIO) Stat(path string) (bool, error) {
	// No locking here - the func handles its own locking
	return m.StatFunc(path)
}

// StatWithContext implements the FilesystemIO interface
func (m *MockFilesystemIO) StatWithContext(ctx context.Context, path string) (bool, error) {
	// No locking here - the func handles its own locking
	return m.StatWithContextFunc(ctx, path)
}

// mockFileError is a simple error type for filesystem operations
type mockFileError struct {
	msg string
}

// Error implements the error interface
func (e *mockFileError) Error() string {
	return e.msg
}

// EnvironmentProvider represents the external boundary for environment variables
type EnvironmentProvider interface {
	// GetEnv gets an environment variable
	GetEnv(key string) string

	// LookupEnv looks up an environment variable
	LookupEnv(key string) (string, bool)
}

// MockEnvironmentProvider implements EnvironmentProvider for testing
type MockEnvironmentProvider struct {
	// GetEnvFunc allows customizing the behavior of GetEnv in tests
	GetEnvFunc func(key string) string

	// LookupEnvFunc allows customizing the behavior of LookupEnv in tests
	LookupEnvFunc func(key string) (string, bool)

	// EnvVars stores environment variables for in-memory simulation
	// Access must be synchronized via mutex
	EnvVars map[string]string

	// mutex for thread-safe access to EnvVars
	mutex sync.RWMutex
}

// NewMockEnvironmentProvider creates a new MockEnvironmentProvider with default implementations
func NewMockEnvironmentProvider() *MockEnvironmentProvider {
	envVars := make(map[string]string)

	m := &MockEnvironmentProvider{
		EnvVars: envVars,
	}

	// Default implementation of GetEnv
	m.GetEnvFunc = func(key string) string {
		m.mutex.RLock()
		value := m.EnvVars[key]
		m.mutex.RUnlock()
		return value
	}

	// Default implementation of LookupEnv
	m.LookupEnvFunc = func(key string) (string, bool) {
		m.mutex.RLock()
		value, exists := m.EnvVars[key]
		m.mutex.RUnlock()
		return value, exists
	}

	return m
}

// GetEnv implements the EnvironmentProvider interface
func (m *MockEnvironmentProvider) GetEnv(key string) string {
	return m.GetEnvFunc(key)
}

// LookupEnv implements the EnvironmentProvider interface
func (m *MockEnvironmentProvider) LookupEnv(key string) (string, bool) {
	return m.LookupEnvFunc(key)
}

// SetEnv sets an environment variable in a thread-safe manner
func (m *MockEnvironmentProvider) SetEnv(key, value string) {
	m.mutex.Lock()
	m.EnvVars[key] = value
	m.mutex.Unlock()
}

// TimeProvider represents the external boundary for time operations
type TimeProvider interface {
	// Now returns the current time
	Now() time.Time

	// Sleep pauses execution for the specified duration
	Sleep(d time.Duration)
}

// MockTimeProvider implements TimeProvider for testing
type MockTimeProvider struct {
	// NowFunc allows customizing the behavior of Now in tests
	NowFunc func() time.Time

	// SleepFunc allows customizing the behavior of Sleep in tests
	SleepFunc func(d time.Duration)

	// CurrentTime stores the current time for in-memory simulation
	// Access must be synchronized via mutex
	CurrentTime time.Time

	// mutex for thread-safe access to CurrentTime
	mutex sync.RWMutex
}

// NewMockTimeProvider creates a new MockTimeProvider with default implementations
func NewMockTimeProvider() *MockTimeProvider {
	m := &MockTimeProvider{
		CurrentTime: time.Date(2025, 4, 1, 12, 0, 0, 0, time.UTC),
	}

	// Default implementation of Now
	m.NowFunc = func() time.Time {
		m.mutex.RLock()
		t := m.CurrentTime
		m.mutex.RUnlock()
		return t
	}

	// Default implementation of Sleep
	m.SleepFunc = func(d time.Duration) {
		// Simulate time passing
		m.mutex.Lock()
		m.CurrentTime = m.CurrentTime.Add(d)
		m.mutex.Unlock()
	}

	return m
}

// Now implements the TimeProvider interface
func (m *MockTimeProvider) Now() time.Time {
	return m.NowFunc()
}

// Sleep implements the TimeProvider interface
func (m *MockTimeProvider) Sleep(d time.Duration) {
	m.SleepFunc(d)
}

// SetTime sets the current time in a thread-safe manner
func (m *MockTimeProvider) SetTime(t time.Time) {
	m.mutex.Lock()
	m.CurrentTime = t
	m.mutex.Unlock()
}
