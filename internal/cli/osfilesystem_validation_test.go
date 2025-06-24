// Package cli provides additional validation tests for OSFileSystem
// These tests specifically validate that OSFileSystem methods are properly covered
// and exercise edge cases to ensure robustness
package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOSFileSystemValidation provides comprehensive validation of OSFileSystem coverage
// This validates that production methods are correctly attributed in coverage reports
func TestOSFileSystemValidation(t *testing.T) {
	fs := &OSFileSystem{}

	t.Run("CreateTemp error handling", func(t *testing.T) {
		// Test CreateTemp with invalid directory to ensure error paths are covered
		_, err := fs.CreateTemp("/nonexistent/directory", "test_*")
		assert.Error(t, err, "CreateTemp should fail with nonexistent directory")
	})

	t.Run("WriteFile error handling", func(t *testing.T) {
		// Test WriteFile with invalid path to ensure error paths are covered
		err := fs.WriteFile("/root/protected/file.txt", []byte("data"), 0644)
		assert.Error(t, err, "WriteFile should fail with protected directory")
	})

	t.Run("ReadFile error handling", func(t *testing.T) {
		// Test ReadFile with nonexistent file to ensure error paths are covered
		_, err := fs.ReadFile("/nonexistent/file.txt")
		assert.Error(t, err, "ReadFile should fail with nonexistent file")
	})

	t.Run("Remove error handling", func(t *testing.T) {
		// Test Remove with nonexistent file to ensure error paths are covered
		err := fs.Remove("/nonexistent/file.txt")
		assert.Error(t, err, "Remove should fail with nonexistent file")
	})

	t.Run("MkdirAll with existing directory", func(t *testing.T) {
		// Create temp directory for test
		tempDir, err := os.MkdirTemp("", "mkdir_validation_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Test MkdirAll with existing directory (should succeed)
		err = fs.MkdirAll(tempDir, 0755)
		assert.NoError(t, err, "MkdirAll should succeed with existing directory")
	})

	t.Run("OpenFile with various flags", func(t *testing.T) {
		// Create temp directory for test
		tempDir, err := os.MkdirTemp("", "open_validation_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		testFile := filepath.Join(tempDir, "test_flags.txt")

		// Test OpenFile with read-only flag on nonexistent file (should fail)
		_, err = fs.OpenFile(testFile, os.O_RDONLY, 0644)
		assert.Error(t, err, "OpenFile should fail when opening nonexistent file for reading")

		// Test OpenFile with create flag (should succeed)
		file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
		assert.NoError(t, err, "OpenFile should succeed with create flag")
		if file != nil {
			_ = file.Close()
		}

		// Test OpenFile with read flag on existing file (should succeed)
		file, err = fs.OpenFile(testFile, os.O_RDONLY, 0644)
		assert.NoError(t, err, "OpenFile should succeed when opening existing file for reading")
		if file != nil {
			_ = file.Close()
		}
	})

	t.Run("Complete workflow validation", func(t *testing.T) {
		// Test complete workflow to ensure all methods work together
		tempDir, err := os.MkdirTemp("", "workflow_validation_*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create nested directory structure
		nestedPath := filepath.Join(tempDir, "nested", "structure")
		err = fs.MkdirAll(nestedPath, 0755)
		require.NoError(t, err)

		// Create temporary file in nested structure
		tempFile, err := fs.CreateTemp(nestedPath, "workflow_*")
		require.NoError(t, err)
		tempFileName := tempFile.Name()
		_ = tempFile.Close()

		// Write data to file
		testData := []byte("validation workflow test data")
		err = fs.WriteFile(tempFileName, testData, 0644)
		require.NoError(t, err)

		// Read data back
		readData, err := fs.ReadFile(tempFileName)
		require.NoError(t, err)
		assert.Equal(t, testData, readData)

		// Open file for appending
		file, err := fs.OpenFile(tempFileName, os.O_APPEND|os.O_WRONLY, 0644)
		require.NoError(t, err)
		_, err = file.WriteString(" - appended")
		require.NoError(t, err)
		_ = file.Close()

		// Verify appended content
		finalData, err := fs.ReadFile(tempFileName)
		require.NoError(t, err)
		expected := append(testData, []byte(" - appended")...)
		assert.Equal(t, expected, finalData)

		// Remove the file
		err = fs.Remove(tempFileName)
		require.NoError(t, err)

		// Verify file is removed
		_, err = fs.ReadFile(tempFileName)
		assert.Error(t, err, "File should be removed")
		assert.True(t, os.IsNotExist(err), "Error should indicate file does not exist")
	})
}

// TestOSFileSystemCoverageAttribution validates that test execution is properly attributed
// to production OSFileSystem methods and not mock methods
func TestOSFileSystemCoverageAttribution(t *testing.T) {
	// This test ensures that when we create an OSFileSystem instance,
	// the methods called are the production implementations in run_implementations.go
	// and not the mock implementations in run_mocks.go

	fs := &OSFileSystem{}

	// Verify the type is correct
	assert.IsType(t, &OSFileSystem{}, fs, "Should be production OSFileSystem type")

	// Create temp file and verify it uses real filesystem
	tempFile, err := fs.CreateTemp("", "attribution_test_*")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tempFile.Name()) }()
	defer func() { _ = tempFile.Close() }()

	// Verify the file actually exists on real filesystem
	_, err = os.Stat(tempFile.Name())
	assert.NoError(t, err, "File should exist on real filesystem, confirming production method usage")

	// Test that WriteFile creates real files
	testFile := tempFile.Name() + "_write_test"
	defer func() { _ = os.Remove(testFile) }()

	err = fs.WriteFile(testFile, []byte("real file test"), 0644)
	require.NoError(t, err)

	// Verify file exists on real filesystem
	_, err = os.Stat(testFile)
	assert.NoError(t, err, "WriteFile should create real files, confirming production method usage")

	// Test that ReadFile reads real files
	data, err := fs.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, []byte("real file test"), data, "ReadFile should read real file content")
}

// TestOSFileSystemInterfaceCompliance validates that OSFileSystem implements FileSystem interface
func TestOSFileSystemInterfaceCompliance(t *testing.T) {
	// Compile-time interface compliance check
	var _ FileSystem = (*OSFileSystem)(nil)

	// Runtime verification that all methods are accessible
	fs := &OSFileSystem{}

	// Verify all interface methods are implemented and callable
	tempDir, err := os.MkdirTemp("", "interface_test_*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Test CreateTemp method
	tempFile, err := fs.CreateTemp(tempDir, "interface_*")
	require.NoError(t, err)
	tempFileName := tempFile.Name()
	_ = tempFile.Close()

	// Test WriteFile method
	err = fs.WriteFile(tempFileName, []byte("interface test"), 0644)
	require.NoError(t, err)

	// Test ReadFile method
	data, err := fs.ReadFile(tempFileName)
	require.NoError(t, err)
	assert.Equal(t, []byte("interface test"), data)

	// Test MkdirAll method
	testDir := filepath.Join(tempDir, "interface_dir")
	err = fs.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	// Test OpenFile method
	testFile := filepath.Join(testDir, "interface_file.txt")
	file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
	require.NoError(t, err)
	_ = file.Close()

	// Test Remove method
	err = fs.Remove(testFile)
	require.NoError(t, err)

	t.Log("✅ OSFileSystem fully implements FileSystem interface with 100% method coverage")
}
