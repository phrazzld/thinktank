// Package pathutil provides utilities for path manipulation and display.
package pathutil

import (
	"os"
	"path/filepath"
	"strings"
)

// SanitizePathForDisplay converts an absolute path to a user-friendly
// relative path for display purposes. This prevents leaking internal
// build paths or developer machine paths to end users.
//
// Behavior:
//   - If path is within cwd, returns relative path (e.g., "./output")
//   - If path equals cwd, returns "."
//   - If path is outside cwd, returns "./" + base directory name
//   - If path is empty, returns empty string
//   - On any error, returns the original path unchanged
func SanitizePathForDisplay(absPath string) string {
	if absPath == "" {
		return ""
	}

	cwd, err := os.Getwd()
	if err != nil {
		return absPath
	}

	return sanitizePathWithCwd(absPath, cwd)
}

// sanitizePathWithCwd is the testable core logic that accepts cwd as parameter.
func sanitizePathWithCwd(absPath, cwd string) string {
	if absPath == "" {
		return ""
	}

	// Clean both paths for consistent comparison
	absPath = filepath.Clean(absPath)
	cwd = filepath.Clean(cwd)

	// If path equals cwd, return "."
	if absPath == cwd {
		return "."
	}

	// Try to make relative to cwd
	relPath, err := filepath.Rel(cwd, absPath)
	if err != nil {
		// Can't make relative, return base name prefixed with ./
		return "./" + filepath.Base(absPath)
	}

	// Check if the relative path escapes cwd (starts with ..)
	if strings.HasPrefix(relPath, "..") {
		// Path is outside cwd, just show base directory name
		return "./" + filepath.Base(absPath)
	}

	// Return relative path, ensuring it starts with ./
	if !strings.HasPrefix(relPath, ".") {
		return "./" + relPath
	}
	return relPath
}
