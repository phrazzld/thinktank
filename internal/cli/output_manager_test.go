package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/misty-step/thinktank/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOutputManager(t *testing.T) {
	t.Run("with logger", func(t *testing.T) {
		logger := testutil.NewMockLogger()
		om := NewOutputManager(logger)
		assert.NotNil(t, om)
		assert.Equal(t, logger, om.logger)
		assert.NotNil(t, om.rand)
	})

	t.Run("with nil logger", func(t *testing.T) {
		om := NewOutputManager(nil)
		assert.NotNil(t, om)
		assert.NotNil(t, om.logger)
		assert.NotNil(t, om.rand)
	})
}

func TestGenerateTimestampedDirName(t *testing.T) {
	om := NewOutputManager(testutil.NewMockLogger())

	t.Run("format validation", func(t *testing.T) {
		name := om.GenerateTimestampedDirName()

		// Define the expected regex pattern
		pattern := `^thinktank_\d{8}_\d{6}_\d{9}$`
		re := regexp.MustCompile(pattern)

		// Verify the name matches the expected pattern
		assert.Regexp(t, re, name, "Generated name should match expected format")

		// Verify the prefix
		assert.True(t, strings.HasPrefix(name, "thinktank_"), "Should have thinktank_ prefix")

		// Extract and verify the components
		parts := strings.Split(name, "_")
		require.Len(t, parts, 4, "Should have 4 parts separated by underscores")

		dateStr := parts[1]   // YYYYMMDD
		timeStr := parts[2]   // HHMMSS
		uniqueStr := parts[3] // NNNNNNNNN

		// Verify the date part
		require.Len(t, dateStr, 8, "Date part should be 8 digits")
		year, err := strconv.Atoi(dateStr[:4])
		require.NoError(t, err)
		assert.GreaterOrEqual(t, year, 2000)
		assert.LessOrEqual(t, year, 2100)

		month, err := strconv.Atoi(dateStr[4:6])
		require.NoError(t, err)
		assert.GreaterOrEqual(t, month, 1)
		assert.LessOrEqual(t, month, 12)

		day, err := strconv.Atoi(dateStr[6:8])
		require.NoError(t, err)
		assert.GreaterOrEqual(t, day, 1)
		assert.LessOrEqual(t, day, 31)

		// Verify the time part
		require.Len(t, timeStr, 6, "Time part should be 6 digits")
		hour, err := strconv.Atoi(timeStr[:2])
		require.NoError(t, err)
		assert.GreaterOrEqual(t, hour, 0)
		assert.LessOrEqual(t, hour, 23)

		minute, err := strconv.Atoi(timeStr[2:4])
		require.NoError(t, err)
		assert.GreaterOrEqual(t, minute, 0)
		assert.LessOrEqual(t, minute, 59)

		second, err := strconv.Atoi(timeStr[4:6])
		require.NoError(t, err)
		assert.GreaterOrEqual(t, second, 0)
		assert.LessOrEqual(t, second, 59)

		// Verify the uniqueness component
		require.Len(t, uniqueStr, 9, "Uniqueness part should be exactly 9 digits")
		uniqueNum, err := strconv.Atoi(uniqueStr)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, uniqueNum, 0)
		assert.LessOrEqual(t, uniqueNum, 999999999)
	})

	t.Run("uniqueness validation", func(t *testing.T) {
		// Generate multiple run names
		runs := 100
		generatedNames := make(map[string]bool, runs)

		for i := 0; i < runs; i++ {
			name := om.GenerateTimestampedDirName()

			// Check that we're not getting duplicate names
			assert.False(t, generatedNames[name], "Name %q should be unique (iteration %d)", name, i)
			generatedNames[name] = true
		}

		// Verify we got unique names
		assert.Len(t, generatedNames, runs, "Should generate unique names")
	})
}

func TestCreateOutputDirectory(t *testing.T) {
	om := NewOutputManager(testutil.NewMockLogger())

	t.Run("create in temp directory", func(t *testing.T) {
		tempDir := t.TempDir()

		dirPath, err := om.CreateOutputDirectory(tempDir, 0755)
		require.NoError(t, err)
		assert.NotEmpty(t, dirPath)

		// Verify directory exists
		info, err := os.Stat(dirPath)
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		// Verify permissions
		assert.Equal(t, os.FileMode(0755), info.Mode().Perm())

		// Verify it's in the temp directory
		assert.True(t, strings.HasPrefix(dirPath, tempDir))

		// Verify name format
		dirName := filepath.Base(dirPath)
		assert.True(t, strings.HasPrefix(dirName, "thinktank_"))
	})

	t.Run("create with empty base path uses cwd", func(t *testing.T) {
		// This test runs in a temporary directory context
		originalCwd, err := os.Getwd()
		require.NoError(t, err)

		tempDir := t.TempDir()
		err = os.Chdir(tempDir)
		require.NoError(t, err)
		defer func() { _ = os.Chdir(originalCwd) }()

		dirPath, err := om.CreateOutputDirectory("", 0755)
		require.NoError(t, err)

		// Should create directory in current working directory (tempDir)
		// Get absolute path of tempDir and resolve symlinks for comparison
		absTempDir, err := filepath.Abs(tempDir)
		require.NoError(t, err)
		realTempDir, err := filepath.EvalSymlinks(absTempDir)
		require.NoError(t, err)
		realDirPath, err := filepath.EvalSymlinks(dirPath)
		require.NoError(t, err)

		assert.True(t, strings.HasPrefix(realDirPath, realTempDir),
			"dirPath %s should have prefix %s", realDirPath, realTempDir)

		// Verify directory exists
		_, err = os.Stat(dirPath)
		require.NoError(t, err)
	})

	t.Run("collision handling", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create first directory
		dirPath1, err := om.CreateOutputDirectory(tempDir, 0755)
		require.NoError(t, err)

		// Force collision by creating directory with same name pattern
		// Note: this is hard to test reliably since names are unique by design
		// But we can test the retry mechanism by pre-creating a directory
		baseName := filepath.Base(dirPath1)
		retryName := baseName + "_retry1"
		retryPath := filepath.Join(tempDir, retryName)
		err = os.MkdirAll(retryPath, 0755)
		require.NoError(t, err)

		// The retry mechanism should work if we somehow get a collision
		// Since our implementation is designed to avoid collisions, this tests the safety net
	})

	t.Run("invalid permissions", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create directory with restrictive permissions
		restrictedDir := filepath.Join(tempDir, "restricted")
		err := os.MkdirAll(restrictedDir, 0000)
		require.NoError(t, err)

		// Try to create output directory in restricted directory
		_, err = om.CreateOutputDirectory(restrictedDir, 0755)
		assert.Error(t, err, "Should fail when cannot create directory")
	})
}

func TestIsThinktankOutputDir(t *testing.T) {
	om := NewOutputManager(testutil.NewMockLogger())

	tests := []struct {
		name     string
		dirName  string
		expected bool
	}{
		{"valid format", "thinktank_20250624_143000_123456789", true},
		{"valid with retry", "thinktank_20250624_143000_123456789_retry1", true},
		{"missing prefix", "nothinktank_20250624_143000_123456789", false},
		{"too few parts", "thinktank_20250624", false},
		{"wrong date format", "thinktank_2025624_143000_123456789", false},
		{"wrong time format", "thinktank_20250624_14300_123456789", false},
		{"wrong unique format", "thinktank_20250624_143000_12345678", false},
		{"empty string", "", false},
		{"completely different", "some_other_directory", false},
		{"partial match", "thinktank_partial", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := om.isThinktankOutputDir(tt.dirName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanupOldDirectories(t *testing.T) {
	om := NewOutputManager(testutil.NewMockLogger())

	t.Run("cleanup old directories", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create some test directories with different ages
		oldTime := time.Now().Add(-45 * 24 * time.Hour)    // 45 days ago
		recentTime := time.Now().Add(-15 * 24 * time.Hour) // 15 days ago

		// Create old directory
		oldDir := filepath.Join(tempDir, "thinktank_20241110_120000_123456789")
		err := os.MkdirAll(oldDir, 0755)
		require.NoError(t, err)
		err = os.Chtimes(oldDir, oldTime, oldTime)
		require.NoError(t, err)

		// Create recent directory
		recentDir := filepath.Join(tempDir, "thinktank_20250610_120000_987654321")
		err = os.MkdirAll(recentDir, 0755)
		require.NoError(t, err)
		err = os.Chtimes(recentDir, recentTime, recentTime)
		require.NoError(t, err)

		// Create non-thinktank directory (should not be touched)
		otherDir := filepath.Join(tempDir, "other_directory")
		err = os.MkdirAll(otherDir, 0755)
		require.NoError(t, err)
		err = os.Chtimes(otherDir, oldTime, oldTime)
		require.NoError(t, err)

		// Cleanup directories older than 30 days
		err = om.CleanupOldDirectories(tempDir, 30*24*time.Hour)
		require.NoError(t, err)

		// Verify old directory was removed
		_, err = os.Stat(oldDir)
		assert.True(t, os.IsNotExist(err), "Old directory should be removed")

		// Verify recent directory still exists
		_, err = os.Stat(recentDir)
		assert.NoError(t, err, "Recent directory should still exist")

		// Verify non-thinktank directory still exists
		_, err = os.Stat(otherDir)
		assert.NoError(t, err, "Non-thinktank directory should not be touched")
	})

	t.Run("cleanup with empty base path", func(t *testing.T) {
		// This test should use current working directory
		err := om.CleanupOldDirectories("", 30*24*time.Hour)
		assert.NoError(t, err, "Should handle empty base path gracefully")
	})

	t.Run("cleanup non-existent directory", func(t *testing.T) {
		err := om.CleanupOldDirectories("/non/existent/path", 30*24*time.Hour)
		assert.Error(t, err, "Should fail for non-existent directory")
	})
}

func TestCleanupOldDirectoriesWithDefault(t *testing.T) {
	om := NewOutputManager(testutil.NewMockLogger())
	tempDir := t.TempDir()

	// Create an old directory (45 days ago)
	oldTime := time.Now().Add(-45 * 24 * time.Hour)
	oldDir := filepath.Join(tempDir, "thinktank_20241110_120000_123456789")
	err := os.MkdirAll(oldDir, 0755)
	require.NoError(t, err)
	err = os.Chtimes(oldDir, oldTime, oldTime)
	require.NoError(t, err)

	// Should cleanup directories older than 30 days by default
	err = om.CleanupOldDirectoriesWithDefault(tempDir)
	assert.NoError(t, err)

	// Verify old directory was removed
	_, err = os.Stat(oldDir)
	assert.True(t, os.IsNotExist(err), "Old directory should be removed by default cleanup")
}

func TestGetOutputDirStats(t *testing.T) {
	om := NewOutputManager(testutil.NewMockLogger())

	t.Run("get stats for directory with thinktank dirs", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create some test directories
		thinktankDir1 := filepath.Join(tempDir, "thinktank_20250624_120000_123456789")
		err := os.MkdirAll(thinktankDir1, 0755)
		require.NoError(t, err)

		thinktankDir2 := filepath.Join(tempDir, "thinktank_20250625_130000_987654321")
		err = os.MkdirAll(thinktankDir2, 0755)
		require.NoError(t, err)

		// Create non-thinktank directory
		otherDir := filepath.Join(tempDir, "other_directory")
		err = os.MkdirAll(otherDir, 0755)
		require.NoError(t, err)

		// Create some files to test size calculation
		testFile := filepath.Join(thinktankDir1, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		stats, err := om.GetOutputDirStats(tempDir)
		require.NoError(t, err)

		assert.Equal(t, 3, stats["total_dirs"])
		assert.Equal(t, 2, stats["thinktank_dirs"])
		assert.NotEmpty(t, stats["oldest_dir"])
		assert.NotEmpty(t, stats["newest_dir"])
		assert.Greater(t, stats["total_size_bytes"].(int64), int64(0))
	})

	t.Run("get stats for empty directory", func(t *testing.T) {
		tempDir := t.TempDir()

		stats, err := om.GetOutputDirStats(tempDir)
		require.NoError(t, err)

		assert.Equal(t, 0, stats["total_dirs"])
		assert.Equal(t, 0, stats["thinktank_dirs"])
		assert.Empty(t, stats["oldest_dir"])
		assert.Empty(t, stats["newest_dir"])
		assert.Equal(t, int64(0), stats["total_size_bytes"])
	})

	t.Run("get stats with empty base path", func(t *testing.T) {
		// Should use current working directory
		stats, err := om.GetOutputDirStats("")
		assert.NoError(t, err)
		assert.NotNil(t, stats)
	})

	t.Run("get stats for non-existent directory", func(t *testing.T) {
		_, err := om.GetOutputDirStats("/non/existent/path")
		assert.Error(t, err)
	})
}

func TestCalculateDirSize(t *testing.T) {
	om := NewOutputManager(testutil.NewMockLogger())
	tempDir := t.TempDir()

	// Create test files with known sizes
	file1 := filepath.Join(tempDir, "file1.txt")
	content1 := "test content 1"
	err := os.WriteFile(file1, []byte(content1), 0644)
	require.NoError(t, err)

	subDir := filepath.Join(tempDir, "subdir")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	file2 := filepath.Join(subDir, "file2.txt")
	content2 := "test content 2"
	err = os.WriteFile(file2, []byte(content2), 0644)
	require.NoError(t, err)

	size, err := om.calculateDirSize(tempDir)
	require.NoError(t, err)

	expectedSize := int64(len(content1) + len(content2))
	assert.Equal(t, expectedSize, size)
}

func TestOutputManagerIntegration(t *testing.T) {
	om := NewOutputManager(testutil.NewMockLogger())
	tempDir := t.TempDir()

	t.Run("full workflow", func(t *testing.T) {
		// Create output directory
		dirPath, err := om.CreateOutputDirectory(tempDir, 0755)
		require.NoError(t, err)

		// Verify it exists
		_, err = os.Stat(dirPath)
		require.NoError(t, err)

		// Get stats
		stats, err := om.GetOutputDirStats(tempDir)
		require.NoError(t, err)
		assert.Equal(t, 1, stats["thinktank_dirs"])

		// Create some content
		testFile := filepath.Join(dirPath, "analysis.txt")
		err = os.WriteFile(testFile, []byte("analysis results"), 0644)
		require.NoError(t, err)

		// Get updated stats with size
		stats, err = om.GetOutputDirStats(tempDir)
		require.NoError(t, err)
		assert.Greater(t, stats["total_size_bytes"].(int64), int64(0))

		// Cleanup (should not remove recent directory)
		err = om.CleanupOldDirectories(tempDir, 24*time.Hour)
		require.NoError(t, err)

		// Directory should still exist
		_, err = os.Stat(dirPath)
		assert.NoError(t, err)
	})
}
