package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

// TestMockFilesystemIO_ConcurrentAccess verifies thread safety of MockFilesystemIO
func TestMockFilesystemIO_ConcurrentAccess(t *testing.T) {
	fs := NewMockFilesystemIO()

	// Number of concurrent goroutines
	const numGoroutines = 100
	const numOperations = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Channel to collect any errors
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				// Mix of read and write operations
				switch j % 4 {
				case 0:
					// Write file
					path := fmt.Sprintf("/test/file%d_%d.txt", id, j)
					data := []byte(fmt.Sprintf("content %d-%d", id, j))
					if err := fs.WriteFile(path, data, 0644); err != nil {
						// Create directory first
						fs.MkdirAll("/test", 0755)
						if err := fs.WriteFile(path, data, 0644); err != nil {
							errChan <- fmt.Errorf("write failed: %w", err)
							return
						}
					}

				case 1:
					// Read file
					path := fmt.Sprintf("/test/file%d_%d.txt", id, j-1)
					if _, err := fs.ReadFile(path); err != nil {
						// It's ok if file doesn't exist in this test
						continue
					}

				case 2:
					// Create directory
					dir := fmt.Sprintf("/test/dir%d/%d", id, j)
					if err := fs.MkdirAll(dir, 0755); err != nil {
						errChan <- fmt.Errorf("mkdir failed: %w", err)
						return
					}

				case 3:
					// Check file existence
					path := fmt.Sprintf("/test/file%d_%d.txt", id, j-2)
					fs.Stat(path)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		t.Errorf("Concurrent operation failed: %v", err)
	}
}

// TestMockFilesystemIO_ConcurrentMapIteration tests concurrent iteration and modification
func TestMockFilesystemIO_ConcurrentMapIteration(t *testing.T) {
	fs := NewMockFilesystemIO()

	// Pre-populate with some data
	fs.MkdirAll("/base", 0755)
	for i := 0; i < 10; i++ {
		path := fmt.Sprintf("/base/file%d.txt", i)
		fs.WriteFile(path, []byte(fmt.Sprintf("content%d", i)), 0644)
	}

	var wg sync.WaitGroup
	wg.Add(3)

	// Goroutine 1: Continuously write new files
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			path := fmt.Sprintf("/base/new%d.txt", i)
			fs.WriteFile(path, []byte("new content"), 0644)
		}
	}()

	// Goroutine 2: Continuously delete files (via RemoveAll)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			// This will iterate over the FileContents map
			fs.RemoveAll("/base")
			fs.MkdirAll("/base", 0755)
		}
	}()

	// Goroutine 3: Continuously read files
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			// Try to read various files
			for j := 0; j < 10; j++ {
				path := fmt.Sprintf("/base/file%d.txt", j)
				fs.ReadFile(path)
			}
		}
	}()

	// This should complete without panicking due to concurrent map access
	wg.Wait()
}

// TestMockFilesystemIO_ThreadSafetyGuarantees documents expected thread safety
func TestMockFilesystemIO_ThreadSafetyGuarantees(t *testing.T) {
	fs := NewMockFilesystemIO()
	ctx := context.Background()

	// Test that context versions are also thread-safe
	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			fs.WriteFileWithContext(ctx, fmt.Sprintf("/file%d", i), []byte("data"), 0644)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			fs.ReadFileWithContext(ctx, fmt.Sprintf("/file%d", i))
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			fs.MkdirAllWithContext(ctx, fmt.Sprintf("/dir%d", i), 0755)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			fs.StatWithContext(ctx, fmt.Sprintf("/file%d", i))
		}
	}()

	wg.Wait()
}
