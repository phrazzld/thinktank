# Create E2E tests for CLI usage with context files

## Goal
Create end-to-end tests that verify the CLI tool correctly handles context files in real-world usage scenarios by running the actual CLI command.

## Analysis of Current Testing
The project already has extensive unit and integration testing for the context paths feature, including:
1. Unit tests for file reading utilities
2. Unit tests for CLI command parameter parsing
3. Integration tests for runThinktank workflow with context paths

However, it's missing actual E2E tests that launch the CLI as a real executable and verify its behavior with actual context files. These tests are crucial to ensure the feature works correctly in the way users will experience it.

## Potential Approaches

### Approach 1: Create CLI process via execa/child_process
- Use execa or Node's child_process to spawn actual CLI processes
- Create temporary test files on disk for testing
- Capture stdout/stderr to validate outputs
- **Pros:** Closest to real usage, tests the complete execution path
- **Cons:** More complex setup, potential for flakiness with process handling

### Approach 2: Use CLI testing library like commander-test
- Use a specialized library designed for testing Commander.js applications
- Configure the testing environment to mimic CLI execution
- Verify outputs programmatically
- **Pros:** Simpler setup, more deterministic
- **Cons:** Not testing the complete execution chain, may miss process-related issues

### Approach 3: Manual integration testing
- Create a script that runs the CLI with various inputs
- Manually inspect results
- Document findings
- **Pros:** Simple to set up
- **Cons:** Not automated, not suitable for CI/CD pipelines

## Selected Approach
**Approach 1: Create CLI process via execa/child_process**

This approach is most appropriate because:

1. **Genuine E2E Testing**: It actually tests the CLI program as it would be run by users
2. **Complete Coverage**: Tests the entire execution path from command parsing to file output
3. **Alignment with Standards**: Provides true E2E tests, not just extended integration tests
4. **Future-proofing**: Most resilient to future changes in implementation details

## Implementation Strategy

I'll implement this approach with the following methodology:

1. Create a dedicated E2E test file in a suitable location
2. Set up temporary test directory and context files for testing
3. Use execa or child_process to run the actual thinktank CLI command
4. Test scenarios:
   - Simple single context file
   - Context file with content that affects LLM output
   - Various file formats (text, markdown, code files)
   - Files with different content sizes
5. Verify results by checking:
   - Command exit code (should be 0 for success)
   - Output files are created with expected content
   - Console output contains expected information
   - Any error cases are handled correctly

This strategy provides comprehensive testing for how the CLI handles context files in real-world usage.