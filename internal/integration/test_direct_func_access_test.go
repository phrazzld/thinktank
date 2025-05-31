package integration

import (
	"fmt"
	"sync"
	"testing"
)

// TestMockFilesystemIO_DirectFuncAccess tests what happens when funcs are called directly
func TestMockFilesystemIO_DirectFuncAccess(t *testing.T) {
	// This test demonstrates the potential race condition if someone
	// bypasses the wrapper methods and calls the Func fields directly

	fs := NewMockFilesystemIO()

	// Pre-populate
	fs.MkdirAll("/test", 0755)

	var wg sync.WaitGroup

	// Multiple goroutines calling the func directly (bypassing mutex)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				// DIRECT access to WriteFileFunc - bypasses the mutex!
				path := fmt.Sprintf("/test/file%d_%d.txt", id, j)
				fs.WriteFileFunc(path, []byte("data"), 0644)
			}
		}(i)
	}

	// Reader also bypassing mutex
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				// DIRECT access to ReadFileFunc - bypasses the mutex!
				path := fmt.Sprintf("/test/file%d_%d.txt", id, j)
				fs.ReadFileFunc(path)
			}
		}(i)
	}

	// This SHOULD trigger a race condition
	wg.Wait()
}
