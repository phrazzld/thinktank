package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/phrazzld/architect/internal/logutil"
)

func TestResolvePath(t *testing.T) {
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ", false)

	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "architect-path-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original environment variables and restore them after the test
	origCacheHome := xdg.CacheHome
	origConfigHome := xdg.ConfigHome
	defer func() {
		xdg.CacheHome = origCacheHome
		xdg.ConfigHome = origConfigHome
	}()

	// Set XDG directories to our test directory
	xdgCacheDir := filepath.Join(tmpDir, "cache")
	xdgConfigDir := filepath.Join(tmpDir, "config")
	xdg.CacheHome = xdgCacheDir
	xdg.ConfigHome = xdgConfigDir

	// Test 1: Absolute paths should remain unchanged
	t.Run("AbsolutePath", func(t *testing.T) {
		absPath := filepath.Join(tmpDir, "somefile.txt")
		resolved, err := resolvePath(absPath, "log", logger)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if resolved != absPath {
			t.Errorf("Expected absolute path '%s' to remain unchanged, got '%s'", absPath, resolved)
		}
	})

	// Test 2: Relative log paths should go to XDG_CACHE_HOME/architect
	t.Run("RelativeLogPath", func(t *testing.T) {
		relPath := "audit.log"
		expected := filepath.Join(xdgCacheDir, "architect", relPath)
		resolved, err := resolvePath(relPath, "log", logger)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if resolved != expected {
			t.Errorf("Expected log path to be '%s', got '%s'", expected, resolved)
		}
	})

	// Test 3: Relative config paths should go to XDG_CONFIG_HOME/architect
	t.Run("RelativeConfigPath", func(t *testing.T) {
		relPath := "config.toml"
		expected := filepath.Join(xdgConfigDir, "architect", relPath)
		resolved, err := resolvePath(relPath, "config", logger)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if resolved != expected {
			t.Errorf("Expected config path to be '%s', got '%s'", expected, resolved)
		}
	})

	// Test 4: Relative output paths should go to current working directory
	t.Run("RelativeOutputPath", func(t *testing.T) {
		relPath := "output.md"
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current working directory: %v", err)
		}
		expected := filepath.Join(cwd, relPath)
		resolved, err := resolvePath(relPath, "output", logger)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if resolved != expected {
			t.Errorf("Expected output path to be '%s', got '%s'", expected, resolved)
		}
	})

	// Test 5: Invalid path type should return an error
	t.Run("InvalidPathType", func(t *testing.T) {
		relPath := "somefile.txt"
		_, err := resolvePath(relPath, "invalid", logger)
		if err == nil {
			t.Error("Expected error for invalid path type, got nil")
		}
	})

	// Test 6: Empty path should return an error
	t.Run("EmptyPath", func(t *testing.T) {
		_, err := resolvePath("", "log", logger)
		if err == nil {
			t.Error("Expected error for empty path, got nil")
		}
	})
}
