// Package integration provides integration tests for the thinktank package
package integration

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/misty-step/thinktank/internal/llm"
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

	// FileContents stores file contents for in-memory simulation
	FileContents map[string][]byte

	// CreatedDirs tracks directories that have been created
	CreatedDirs map[string]bool

	// mutex for thread-safe access to maps
	mutex sync.Mutex
}

// NewMockFilesystemIO creates a new MockFilesystemIO with default implementations
func NewMockFilesystemIO() *MockFilesystemIO {
	fileContents := make(map[string][]byte)
	createdDirs := make(map[string]bool)

	m := &MockFilesystemIO{
		FileContents: fileContents,
		CreatedDirs:  createdDirs,
	}

	// Default implementation of ReadFile
	m.ReadFileFunc = func(path string) ([]byte, error) {
		// Normalize path
		path = filepath.Clean(path)

		if content, ok := m.FileContents[path]; ok {
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

		// Check if parent directory exists
		dir := filepath.Dir(path)
		if dir != "." && !m.CreatedDirs[dir] {
			return &mockFileError{msg: "directory does not exist: " + dir}
		}

		// Store file content
		m.FileContents[path] = data
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

		// Mark directory as created
		m.CreatedDirs[path] = true
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

		// Remove directory
		delete(m.CreatedDirs, path)

		// Remove all files in this directory
		for filePath := range m.FileContents {
			if filepath.Dir(filePath) == path {
				delete(m.FileContents, filePath)
			}
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

		// Check if file exists
		_, fileExists := m.FileContents[path]

		// Check if directory exists
		dirExists := m.CreatedDirs[path]

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
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.ReadFileFunc(path)
}

// ReadFileWithContext implements the FilesystemIO interface
func (m *MockFilesystemIO) ReadFileWithContext(ctx context.Context, path string) ([]byte, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.ReadFileWithContextFunc(ctx, path)
}

// WriteFile implements the FilesystemIO interface
func (m *MockFilesystemIO) WriteFile(path string, data []byte, perm int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.WriteFileFunc(path, data, perm)
}

// WriteFileWithContext implements the FilesystemIO interface
func (m *MockFilesystemIO) WriteFileWithContext(ctx context.Context, path string, data []byte, perm int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.WriteFileWithContextFunc(ctx, path, data, perm)
}

// MkdirAll implements the FilesystemIO interface
func (m *MockFilesystemIO) MkdirAll(path string, perm int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.MkdirAllFunc(path, perm)
}

// MkdirAllWithContext implements the FilesystemIO interface
func (m *MockFilesystemIO) MkdirAllWithContext(ctx context.Context, path string, perm int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.MkdirAllWithContextFunc(ctx, path, perm)
}

// RemoveAll implements the FilesystemIO interface
func (m *MockFilesystemIO) RemoveAll(path string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.RemoveAllFunc(path)
}

// RemoveAllWithContext implements the FilesystemIO interface
func (m *MockFilesystemIO) RemoveAllWithContext(ctx context.Context, path string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.RemoveAllWithContextFunc(ctx, path)
}

// Stat implements the FilesystemIO interface
func (m *MockFilesystemIO) Stat(path string) (bool, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.StatFunc(path)
}

// StatWithContext implements the FilesystemIO interface
func (m *MockFilesystemIO) StatWithContext(ctx context.Context, path string) (bool, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
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
	EnvVars map[string]string
}

// NewMockEnvironmentProvider creates a new MockEnvironmentProvider with default implementations
func NewMockEnvironmentProvider() *MockEnvironmentProvider {
	envVars := make(map[string]string)

	m := &MockEnvironmentProvider{
		EnvVars: envVars,
	}

	// Default implementation of GetEnv
	m.GetEnvFunc = func(key string) string {
		return m.EnvVars[key]
	}

	// Default implementation of LookupEnv
	m.LookupEnvFunc = func(key string) (string, bool) {
		value, exists := m.EnvVars[key]
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
	CurrentTime time.Time
}

// NewMockTimeProvider creates a new MockTimeProvider with default implementations
func NewMockTimeProvider() *MockTimeProvider {
	m := &MockTimeProvider{
		CurrentTime: time.Date(2025, 4, 1, 12, 0, 0, 0, time.UTC),
	}

	// Default implementation of Now
	m.NowFunc = func() time.Time {
		return m.CurrentTime
	}

	// Default implementation of Sleep
	m.SleepFunc = func(d time.Duration) {
		// Simulate time passing
		m.CurrentTime = m.CurrentTime.Add(d)
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
