# CI Failure Analysis

## Summary
The CI pipeline is failing in the Test job due to data race conditions detected by Go's race detector. The initial issue was in the `TestProcessModelsToSynthesis` test in the `github.com/phrazzld/thinktank/internal/thinktank/orchestrator` package, but after fixing that, we found additional data races in the `internal/integration` package.

## Details

### Race Condition 1

The first data race occurred in the `MockAuditLogger.LogOp` method in `internal/thinktank/orchestrator/mocks_test.go`, where multiple goroutines were concurrently accessing and modifying the shared `LogCalls` slice without proper synchronization.

### Race Condition 2

After fixing the first issue, we discovered additional data races in the `MockFilesystemIO` implementation in `internal/integration/test_boundaries.go`. Multiple goroutines were concurrently accessing and modifying the shared `FileContents` and `CreatedDirs` maps without proper synchronization.

### Specific Race Conditions:

1. **Primary Race**:
   - Multiple goroutines simultaneously access and modify the `LogCalls` slice in the `MockAuditLogger.LogOp` method
   - This happens during parallel execution of `orchestrator.(*Orchestrator).processModels`
   - The race detector shows both read and write operations happening concurrently on the same memory location

```go
func (m *MockAuditLogger) LogOp(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
    // Record the call parameters
    m.LogCalls = append(m.LogCalls, LogCall{
        Operation: operation,
        Status:    status,
        Inputs:    inputs,
        Outputs:   outputs,
        Error:     err,
    })
    // Return configured error (nil by default)
    return m.LogError
}
```

### Root Cause:
The `MockAuditLogger` is being used by multiple concurrent goroutines launched by the `Orchestrator.processModels` method, but lacks proper synchronization (e.g., a mutex) to protect the shared `LogCalls` slice. This is a classic data race scenario where multiple threads try to modify a shared data structure without synchronization.

## Recommended Fixes

### Fix for MockAuditLogger

Add mutex synchronization to the `MockAuditLogger` implementation:

1. Add a mutex field to the `MockAuditLogger` struct
2. Lock the mutex before appending to the `LogCalls` slice
3. Unlock the mutex after the operation is complete

```go
type MockAuditLogger struct {
    LogCalls []LogCall
    LogError error // To simulate logging errors for testing error handling
    mutex    sync.Mutex // Add this
}

func (m *MockAuditLogger) LogOp(operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
    // Lock before modifying shared state
    m.mutex.Lock()
    defer m.mutex.Unlock()

    // Record the call parameters
    m.LogCalls = append(m.LogCalls, LogCall{
        Operation: operation,
        Status:    status,
        Inputs:    inputs,
        Outputs:   outputs,
        Error:     err,
    })
    // Return configured error (nil by default)
    return m.LogError
}
```

### Fix for MockFilesystemIO

Add mutex synchronization to the `MockFilesystemIO` implementation:

1. Add a mutex field to the `MockFilesystemIO` struct
2. Add locking in all methods that access shared maps
3. Use `defer` to ensure the mutex is always unlocked

```go
type MockFilesystemIO struct {
    FileContents map[string][]byte
    CreatedDirs map[string]bool
    // Other fields...
    mutex sync.Mutex // Add this
}

// ReadFile implements the FilesystemIO interface
func (m *MockFilesystemIO) ReadFile(path string) ([]byte, error) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    return m.ReadFileFunc(path)
}

// WriteFile implements the FilesystemIO interface
func (m *MockFilesystemIO) WriteFile(path string, data []byte, perm int) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    return m.WriteFileFunc(path, data, perm)
}
// etc. for all methods
```

## Testing Strategy

1. Run the tests locally with race detection to confirm the fixes:
   ```
   go test -race ./internal/thinktank/orchestrator/...
   go test -race ./internal/integration/...
   ```

2. Make sure all other tests still pass with race detection:
   ```
   go test -race ./...
   ```

## Immediate Next Steps

1. Implement the mutex in `MockAuditLogger`
2. Add the sync package import to mocks_test.go
3. Implement the mutex in `MockFilesystemIO`
4. Add the sync package import to test_boundaries.go
5. Run tests locally with race detection to confirm the fixes
6. Commit and push the changes
7. Monitor CI to verify the fixes resolve the issues

## Preventive Measures

1. Consider adding race detection to critical packages in pre-commit hooks
2. Add documentation about the need for thread safety in mock objects used in concurrent tests
3. Review other mock implementations for similar race conditions
