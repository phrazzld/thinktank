# Mock Audit Report: Race Condition Analysis

## Summary

Audit of all test mocks in the codebase for potential race conditions. The race detector currently reports no issues, but some mocks lack proper synchronization.

## Current Status

### ✅ Properly Protected Mocks

1. **internal/fileutil/mock_logger.go**
   - Has `sync.Mutex` protection
   - All methods properly lock/unlock
   - Safe for concurrent access

2. **internal/testutil/mocklogger.go**
   - Has `sync.Mutex` protection
   - All methods properly lock/unlock
   - Safe for concurrent access

3. **internal/testutil/mockregistry.go**
   - Has `sync.RWMutex` protection
   - Proper read/write locking
   - Safe for concurrent access

4. **internal/thinktank/orchestrator/mocks_test.go**
   - `MockAuditLogger` has mutex protection
   - `MockFileWriter` has mutex protection
   - Safe for concurrent access

### ⚠️ Mocks Needing Protection

1. **internal/thinktank/filewriter_test.go - mockAuditLogger**
   - **Issue**: Appends to slice without synchronization
   - **Risk**: Race condition if used in parallel tests
   - **Fix Required**: Add mutex protection

2. **internal/thinktank/orchestrator/output_writer_mock.go - BaseMockOutputWriter**
   - **Issue**: Modifies fields without synchronization
   - **Risk**: Low - appears to be used in sequential tests only
   - **Fix Recommended**: Add mutex for future-proofing

3. **internal/thinktank/orchestrator/orchestrator_individual_output_test.go - MockOutputWriter**
   - **Issue**: Modifies fields without synchronization
   - **Risk**: Low - not used in parallel tests currently
   - **Fix Recommended**: Add mutex for safety

### ✅ Mocks Without State (No Protection Needed)

1. **internal/gemini/mock_client.go - MockClient**
   - Only contains function pointers, no mutable state
   - Safe as-is

2. **internal/thinktank/orchestrator/mocks_test.go - MockLogger**
   - No state, empty implementations
   - Safe as-is

## Implementation Plan

### Priority 1: Fix Critical Issues

#### Fix mockAuditLogger in filewriter_test.go

```go
// mockAuditLogger for testing FileWriter
type mockAuditLogger struct {
    mu      sync.Mutex
    entries []auditlog.AuditEntry
}

func (m *mockAuditLogger) Log(ctx context.Context, entry auditlog.AuditEntry) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.entries = append(m.entries, entry)
    return nil
}

func (m *mockAuditLogger) LogOp(ctx context.Context, operation, status string, inputs map[string]interface{}, outputs map[string]interface{}, err error) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    entry := auditlog.AuditEntry{
        Operation: operation,
        Status:    status,
        Inputs:    inputs,
        Outputs:   outputs,
    }
    m.entries = append(m.entries, entry)
    return nil
}

// Add getter method for safe access
func (m *mockAuditLogger) GetEntries() []auditlog.AuditEntry {
    m.mu.Lock()
    defer m.mu.Unlock()
    result := make([]auditlog.AuditEntry, len(m.entries))
    copy(result, m.entries)
    return result
}
```

### Priority 2: Future-Proof Other Mocks

#### Fix MockOutputWriter variants

```go
type MockOutputWriter struct {
    mu                   sync.Mutex
    savedCount           int
    saveError            error
    capturedModelOutputs map[string]string
    capturedOutputDir    string
}

func (m *MockOutputWriter) SaveIndividualOutputs(ctx context.Context, modelOutputs map[string]string, outputDir string) (int, map[string]string, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.capturedModelOutputs = modelOutputs
    m.capturedOutputDir = outputDir

    filePaths := make(map[string]string)
    for modelName := range modelOutputs {
        filePaths[modelName] = outputDir + "/" + modelName + ".md"
    }

    return m.savedCount, filePaths, m.saveError
}

// Add getter methods for safe access
func (m *MockOutputWriter) GetCapturedOutputs() (map[string]string, string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    outputs := make(map[string]string)
    for k, v := range m.capturedModelOutputs {
        outputs[k] = v
    }
    return outputs, m.capturedOutputDir
}
```

## Best Practices for Thread-Safe Mocks

1. **Always initialize mocks with constructors**
   ```go
   func NewMockXXX() *MockXXX {
       return &MockXXX{
           field: make(map[string]string),
           // Initialize all fields
       }
   }
   ```

2. **Use mutex for any mutable state**
   - Slices that are appended to
   - Maps that are modified
   - Any fields that are written after construction

3. **Provide safe getter methods**
   - Return copies, not references to internal state
   - Lock during the entire copy operation

4. **Use RWMutex for read-heavy mocks**
   - When reads significantly outnumber writes
   - Use RLock() for read operations

5. **Document concurrency safety**
   ```go
   // MockXXX is a thread-safe mock implementation of XXX
   ```

## Verification Steps

1. Run race detector on entire codebase:
   ```bash
   go test -race ./...
   ```

2. Run race detector on specific packages with mocks:
   ```bash
   go test -race ./internal/thinktank/...
   go test -race ./internal/testutil/...
   ```

3. Enable parallel tests where appropriate to catch issues:
   ```go
   t.Parallel() // Add to test functions using mocks
   ```

## Conclusion

While the race detector currently reports no issues, several mocks lack proper synchronization. The most critical fix is for `mockAuditLogger` in `filewriter_test.go` as it modifies a slice without protection. Other mocks should be updated for future-proofing and to follow Go concurrency best practices.
