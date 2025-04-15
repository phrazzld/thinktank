# Integration Testing Guide

This guide outlines best practices for writing properly isolated integration tests for the Architect project, focusing on techniques that allow tests to run in parallel efficiently.

## Key Principles for Integration Test Isolation

1. **Use `t.TempDir()` for File Operations**
   - Always use `t.TempDir()` to create isolated directories for file outputs
   - Never share output directories between tests
   - Let Go's testing framework handle cleanup automatically

2. **Consistent Mock Setup**
   - Use the `LLMClientAdapter` to ensure consistent behavior between `gemini.Client` and `llm.LLMClient`
   - Configure mock responses in test setup, not during test execution
   - Set up mocks to be concurrency-safe

3. **Enable Parallel Execution**
   - Add `t.Parallel()` to top-level test functions
   - Add `t.Parallel()` to subtests within `t.Run()` blocks
   - Capture loop variables with `tc := tc` to avoid closures referencing shared variables

4. **Avoid Test Dependencies**
   - Each test must set up its own environment
   - Don't rely on global state or side effects from other tests
   - Use fresh environment for each test/subtest

## Example: Properly Isolated Test

```go
func TestFeature(t *testing.T) {
    t.Parallel() // Enable parallelization for this test

    // Define test cases
    tests := []struct{
        name string
        input string
        expected string
    }{
        {"Case1", "input1", "output1"},
        {"Case2", "input2", "output2"},
    }

    // Run each test case in parallel
    for _, tc := range tests {
        tc := tc // Capture range variable
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel() // Enable parallelization for this subtest

            // Set up test environment
            env := NewTestEnv(t)
            defer env.Cleanup()

            // Set up mock client
            env.SetupMockGeminiClient()

            // Create test files in isolated directory
            env.CreateTestFile(t, "src/test.go", "package main")

            // Use t.TempDir() for output isolation
            outputDir := t.TempDir()
            outputFile := filepath.Join(outputDir, "output.md")

            // Create test config
            config := &config.CliConfig{
                OutputDir: outputDir,
                // ... other config
            }

            // Run test
            // ... test execution

            // Make assertions
            // ... verify results
        })
    }
}
```

## Common Mistakes to Avoid

1. **Shared Output Directories**
   - ❌ Using hardcoded paths: `filepath.Join(env.TestDir, "output")`
   - ✅ Using temporary directories: `outputDir := t.TempDir()`

2. **Inconsistent Mock Behavior**
   - ❌ Using separate mocks for different interfaces with different behavior
   - ✅ Using the adapter pattern to ensure consistent responses

3. **Race Conditions**
   - ❌ Modifying shared variables in parallel tests
   - ✅ Using local variables and proper synchronization

4. **Missing Range Variable Capture**
   - ❌ Using the loop variable directly in closure: `for _, tc := range tests { t.Run(tc.name, func(t *testing.T) {`
   - ✅ Capturing the range variable: `for _, tc := range tests { tc := tc; t.Run(tc.name, func(t *testing.T) {`

## Debugging Parallel Test Issues

If parallel tests are failing:

1. Temporarily remove `t.Parallel()` to see if tests pass sequentially
2. Use `t.Logf()` to debug test execution flow
3. Check for shared resources or state that might be causing race conditions
4. Analyze file path usage to ensure proper isolation
5. Verify proper synchronization for concurrent operations

## Further Reading

For additional information on Go testing best practices, refer to:
- [Go Testing Package Documentation](https://golang.org/pkg/testing/)
- The project's `TESTING_STRATEGY.md` document
