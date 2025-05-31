//go:build race
// +build race

package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestMockFilesystemIO_StressConcurrentMapAccess aggressively tests for race conditions
func TestMockFilesystemIO_StressConcurrentMapAccess(t *testing.T) {
	// This test is designed to trigger race conditions in map access

	for iteration := 0; iteration < 10; iteration++ {
		fs := NewMockFilesystemIO()

		// Pre-populate to ensure maps are initialized
		fs.MkdirAll("/test", 0755)
		fs.WriteFile("/test/seed.txt", []byte("seed"), 0644)

		var wg sync.WaitGroup

		// Start multiple readers
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				// Tight loop to increase chance of collision
				for j := 0; j < 1000; j++ {
					// Direct map access through ReadFile
					fs.ReadFile("/test/seed.txt")
					// Also trigger map iteration
					for k := 0; k < 10; k++ {
						fs.ReadFile(fmt.Sprintf("/test/file%d.txt", k))
					}
				}
			}(i)
		}

		// Start multiple writers
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 1000; j++ {
					path := fmt.Sprintf("/test/writer%d_%d.txt", id, j)
					// This directly writes to the map
					fs.WriteFile(path, []byte(fmt.Sprintf("data%d", j)), 0644)
				}
			}(i)
		}

		// Start a goroutine that does RemoveAll (map iteration + deletion)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				// This iterates over the map while others are modifying it
				fs.RemoveAll("/test")
				fs.MkdirAll("/test", 0755)
				time.Sleep(time.Microsecond) // Small delay to allow other operations
			}
		}()

		wg.Wait()
	}
}

// TestMockFilesystemIO_ThreadSafety verifies the mock's thread-safe implementation
// This test replaces the previous DirectMapManipulation test which demonstrated
// unsafe direct map access (now prevented by making fields unexported)
func TestMockFilesystemIO_ThreadSafety(t *testing.T) {
	fs := NewMockFilesystemIO()

	// Create base directory first
	if err := fs.MkdirAll("/direct", 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	var wg sync.WaitGroup

	// Writer using the thread-safe WriteFile method
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			path := fmt.Sprintf("/direct/file%d.txt", i)
			if err := fs.WriteFile(path, []byte(fmt.Sprintf("content%d", i)), 0644); err != nil {
				// Directory might have been removed by RemoveAll, which is expected
				continue
			}
		}
	}()

	// Reader using the thread-safe ReadFile method
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			path := fmt.Sprintf("/direct/file%d.txt", i)
			if _, err := fs.ReadFile(path); err != nil {
				// File might not exist yet or been removed, which is expected
				continue
			}
		}
	}()

	// Cleaner using the thread-safe RemoveAll method
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			// RemoveAll is now thread-safe with proper locking
			fs.RemoveAll("/direct")
			// Recreate the directory for continued testing
			fs.MkdirAll("/direct", 0755)
			time.Sleep(time.Microsecond)
		}
	}()

	wg.Wait()

	// If we get here without a race condition, the implementation is thread-safe
	t.Log("MockFilesystemIO implementation is thread-safe")
}
