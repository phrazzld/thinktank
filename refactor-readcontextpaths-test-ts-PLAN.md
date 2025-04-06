# Refactor readContextPaths.test.ts

## Task Goal
Replace direct fs mocks in the readContextPaths.test.ts file with calls to the new utility functions from mockFsUtils.ts.

## Implementation Approach
1. Analyze the existing readContextPaths.test.ts file to identify all instances where fs module mocks are directly set up
2. Determine which mockFsUtils functions need to be used (access, stat, readdir, readFile)
3. Replace each direct mock with the corresponding utility function call
4. Maintain all existing test cases and behavior
5. Ensure proper cleanup is implemented between tests using resetMockFs()

## Reasoning
This approach maintains consistency with other recently refactored test files while ensuring all tests continue to function correctly. By using the shared mock utilities, we reduce code duplication and improve maintainability. The refactoring will make the tests more readable and focused on the behavior being tested rather than the mock setup.