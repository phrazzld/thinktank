// internal/fileutil/fileutil_test.go
// This file has been refactored into multiple test files for better organization:
// - core_utils_test.go: Basic utility tests (token counting, statistics, binary file detection)
// - config_test.go: Configuration-related tests
// - git_operations_test.go: Git-related functionality tests (isGitIgnored)
// - filtering_test.go: Tests for shouldProcess and file filtering functionality
// - file_processing_test.go: Tests for file processing and error handling
// - project_context_test.go: Tests for GatherProjectContext

package fileutil

import (
	"path/filepath"
)

// Keep the isWindows helper function here since it's used by multiple test files
// This avoids duplicating it across files
func isWindows() bool {
	return filepath.Separator == '\\' && filepath.ListSeparator == ';'
}
