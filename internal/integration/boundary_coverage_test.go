// Package integration provides comprehensive coverage tests for boundary operations
// Following TDD principles to target 0% coverage functions with highest business value
package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockFilesystemIOContextMethods tests context-aware boundary operations
// Targets: ReadFileWithContext, WriteFileWithContext, MkdirAllWithContext,
// RemoveAllWithContext, StatWithContext (all 0% coverage)
func TestMockFilesystemIOContextMethods(t *testing.T) {
	t.Run("ReadFileWithContext with canceled context", func(t *testing.T) {
		fs := NewMockFilesystemIO()
		fs.FileContents["/test/file"] = []byte("test content")

		// Test context cancellation
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Custom function that respects context cancellation
		fs.ReadFileWithContextFunc = func(ctx context.Context, path string) ([]byte, error) {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return fs.ReadFileFunc(path)
		}

		_, err := fs.ReadFileWithContext(ctx, "/test/file")
		assert.Equal(t, context.Canceled, err, "Should return context cancellation error")
	})

	t.Run("ReadFileWithContext with valid context", func(t *testing.T) {
		fs := NewMockFilesystemIO()
		fs.FileContents["/test/file"] = []byte("test content")

		content, err := fs.ReadFileWithContext(context.Background(), "/test/file")
		require.NoError(t, err)
		assert.Equal(t, []byte("test content"), content)
	})

	t.Run("WriteFileWithContext with canceled context", func(t *testing.T) {
		fs := NewMockFilesystemIO()
		fs.CreatedDirs["/test"] = true

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Custom function that respects context cancellation
		fs.WriteFileWithContextFunc = func(ctx context.Context, path string, data []byte, perm int) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fs.WriteFileFunc(path, data, perm)
		}

		err := fs.WriteFileWithContext(ctx, "/test/file", []byte("data"), 0644)
		assert.Equal(t, context.Canceled, err, "Should return context cancellation error")
	})

	t.Run("WriteFileWithContext with valid context", func(t *testing.T) {
		fs := NewMockFilesystemIO()
		fs.CreatedDirs["/test"] = true

		err := fs.WriteFileWithContext(context.Background(), "/test/file", []byte("data"), 0644)
		require.NoError(t, err)
		assert.Equal(t, []byte("data"), fs.FileContents["/test/file"])
	})

	t.Run("MkdirAllWithContext with canceled context", func(t *testing.T) {
		fs := NewMockFilesystemIO()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Custom function that respects context cancellation
		fs.MkdirAllWithContextFunc = func(ctx context.Context, path string, perm int) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fs.MkdirAllFunc(path, perm)
		}

		err := fs.MkdirAllWithContext(ctx, "/test/dir", 0755)
		assert.Equal(t, context.Canceled, err, "Should return context cancellation error")
	})

	t.Run("MkdirAllWithContext with valid context", func(t *testing.T) {
		fs := NewMockFilesystemIO()

		err := fs.MkdirAllWithContext(context.Background(), "/test/dir", 0755)
		require.NoError(t, err)
		assert.True(t, fs.CreatedDirs["/test/dir"])
	})

	t.Run("RemoveAllWithContext with canceled context", func(t *testing.T) {
		fs := NewMockFilesystemIO()
		fs.CreatedDirs["/test/dir"] = true
		fs.FileContents["/test/dir/file"] = []byte("content")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Custom function that respects context cancellation
		fs.RemoveAllWithContextFunc = func(ctx context.Context, path string) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fs.RemoveAllFunc(path)
		}

		err := fs.RemoveAllWithContext(ctx, "/test/dir")
		assert.Equal(t, context.Canceled, err, "Should return context cancellation error")
	})

	t.Run("RemoveAllWithContext with valid context", func(t *testing.T) {
		fs := NewMockFilesystemIO()
		fs.CreatedDirs["/test/dir"] = true
		fs.FileContents["/test/dir/file"] = []byte("content")

		err := fs.RemoveAllWithContext(context.Background(), "/test/dir")
		require.NoError(t, err)
		assert.False(t, fs.CreatedDirs["/test/dir"])
		assert.NotContains(t, fs.FileContents, "/test/dir/file")
	})

	t.Run("StatWithContext with canceled context", func(t *testing.T) {
		fs := NewMockFilesystemIO()
		fs.FileContents["/test/file"] = []byte("content")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Custom function that respects context cancellation
		fs.StatWithContextFunc = func(ctx context.Context, path string) (bool, error) {
			if ctx.Err() != nil {
				return false, ctx.Err()
			}
			return fs.StatFunc(path)
		}

		_, err := fs.StatWithContext(ctx, "/test/file")
		assert.Equal(t, context.Canceled, err, "Should return context cancellation error")
	})

	t.Run("StatWithContext with valid context", func(t *testing.T) {
		fs := NewMockFilesystemIO()
		fs.FileContents["/test/file"] = []byte("content")

		exists, err := fs.StatWithContext(context.Background(), "/test/file")
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

// TestMockEnvironmentProviderLookupEnv tests uncovered environment provider method
// Targets: LookupEnv (0% coverage)
func TestMockEnvironmentProviderLookupEnv(t *testing.T) {
	t.Run("LookupEnv with existing variable", func(t *testing.T) {
		env := NewMockEnvironmentProvider()
		env.EnvVars["TEST_VAR"] = "test_value"

		value, exists := env.LookupEnv("TEST_VAR")
		assert.True(t, exists)
		assert.Equal(t, "test_value", value)
	})

	t.Run("LookupEnv with non-existing variable", func(t *testing.T) {
		env := NewMockEnvironmentProvider()

		value, exists := env.LookupEnv("NONEXISTENT_VAR")
		assert.False(t, exists)
		assert.Equal(t, "", value)
	})
}

// TestMockTimeProviderMethods tests uncovered time provider methods
// Targets: Now, Sleep (0% coverage)
func TestMockTimeProviderMethods(t *testing.T) {
	t.Run("Now returns current time", func(t *testing.T) {
		timeProvider := NewMockTimeProvider()
		expectedTime := time.Date(2025, 4, 1, 12, 0, 0, 0, time.UTC)

		actualTime := timeProvider.Now()
		assert.Equal(t, expectedTime, actualTime)
	})

	t.Run("Sleep advances time", func(t *testing.T) {
		timeProvider := NewMockTimeProvider()
		initialTime := timeProvider.Now()

		timeProvider.Sleep(5 * time.Minute)

		finalTime := timeProvider.Now()
		expectedTime := initialTime.Add(5 * time.Minute)
		assert.Equal(t, expectedTime, finalTime)
	})

	t.Run("Sleep with zero duration", func(t *testing.T) {
		timeProvider := NewMockTimeProvider()
		initialTime := timeProvider.Now()

		timeProvider.Sleep(0)

		finalTime := timeProvider.Now()
		assert.Equal(t, initialTime, finalTime)
	})
}

// TestMockFileErrorImplementation tests the error interface implementation
// Targets: Error (0% coverage)
func TestMockFileErrorImplementation(t *testing.T) {
	t.Run("mockFileError implements error interface", func(t *testing.T) {
		err := &mockFileError{msg: "test error message"}
		assert.Equal(t, "test error message", err.Error())
	})

	t.Run("mockFileError with empty message", func(t *testing.T) {
		err := &mockFileError{msg: ""}
		assert.Equal(t, "", err.Error())
	})
}
