// Package integration provides integration tests for the thinktank package
package integration

import (
	"context"
	"os"
	"time"

	"github.com/misty-step/thinktank/internal/llm"
)

// RealFilesystemIO implements the FilesystemIO interface with real OS operations
type RealFilesystemIO struct{}

// ReadFile implements the FilesystemIO interface
func (r *RealFilesystemIO) ReadFile(path string) ([]byte, error) {
	return r.ReadFileWithContext(context.Background(), path)
}

// ReadFileWithContext implements the FilesystemIO interface
func (r *RealFilesystemIO) ReadFileWithContext(ctx context.Context, path string) ([]byte, error) {
	// Check if context is canceled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return os.ReadFile(path)
}

// WriteFile implements the FilesystemIO interface
func (r *RealFilesystemIO) WriteFile(path string, data []byte, perm int) error {
	return r.WriteFileWithContext(context.Background(), path, data, perm)
}

// WriteFileWithContext implements the FilesystemIO interface
func (r *RealFilesystemIO) WriteFileWithContext(ctx context.Context, path string, data []byte, perm int) error {
	// Check if context is canceled
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return os.WriteFile(path, data, os.FileMode(perm))
}

// MkdirAll implements the FilesystemIO interface
func (r *RealFilesystemIO) MkdirAll(path string, perm int) error {
	return r.MkdirAllWithContext(context.Background(), path, perm)
}

// MkdirAllWithContext implements the FilesystemIO interface
func (r *RealFilesystemIO) MkdirAllWithContext(ctx context.Context, path string, perm int) error {
	// Check if context is canceled
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return os.MkdirAll(path, os.FileMode(perm))
}

// RemoveAll implements the FilesystemIO interface
func (r *RealFilesystemIO) RemoveAll(path string) error {
	return r.RemoveAllWithContext(context.Background(), path)
}

// RemoveAllWithContext implements the FilesystemIO interface
func (r *RealFilesystemIO) RemoveAllWithContext(ctx context.Context, path string) error {
	// Check if context is canceled
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return os.RemoveAll(path)
}

// Stat implements the FilesystemIO interface
func (r *RealFilesystemIO) Stat(path string) (bool, error) {
	return r.StatWithContext(context.Background(), path)
}

// StatWithContext implements the FilesystemIO interface
func (r *RealFilesystemIO) StatWithContext(ctx context.Context, path string) (bool, error) {
	// Check if context is canceled
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

// RealEnvironmentProvider implements the EnvironmentProvider interface with real OS operations
type RealEnvironmentProvider struct{}

// GetEnv implements the EnvironmentProvider interface
func (r *RealEnvironmentProvider) GetEnv(key string) string {
	return os.Getenv(key)
}

// LookupEnv implements the EnvironmentProvider interface
func (r *RealEnvironmentProvider) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

// RealTimeProvider implements the TimeProvider interface with real time operations
type RealTimeProvider struct{}

// Now implements the TimeProvider interface
func (r *RealTimeProvider) Now() time.Time {
	return time.Now()
}

// Sleep implements the TimeProvider interface
func (r *RealTimeProvider) Sleep(d time.Duration) {
	time.Sleep(d)
}

// RealExternalAPICaller uses the actual LLM client to make API calls
// This is a thin wrapper that would be replaced with mocks in tests
type RealExternalAPICaller struct {
	// The actual client could be injected here
	// For this example, we're abstracting the real implementation
}

// CallLLMAPI makes a real API call to an LLM provider
// In a real implementation, this would use the appropriate client for each model
func (r *RealExternalAPICaller) CallLLMAPI(ctx context.Context, modelName, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	// In a real implementation, this would initialize the appropriate client
	// and make the actual API call

	// For this example, we're returning an error because real API calls
	// should be avoided in tests
	return nil, &RealAPICallError{
		Message: "real API calls should be mocked in tests",
	}
}

// RealAPICallError represents an error from making a real API call
type RealAPICallError struct {
	Message string
}

// Error implements the error interface
func (e *RealAPICallError) Error() string {
	return e.Message
}
