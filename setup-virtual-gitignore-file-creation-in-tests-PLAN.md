# Setup virtual .gitignore file creation in tests

## Goal
Implement beforeEach hooks in the test files to create virtual filesystem with .gitignore files for each test, following the pattern in the example from PLAN_PHASE2.md.

## Implementation Approaches

### Approach 1: Enable all tests at once
This approach would:
1. Update all test files simultaneously to use virtual .gitignore files
2. Remove all skipped tests and make them active
3. Fix any issues that arise until all tests pass

### Approach 2: Incremental file-by-file approach
This approach would:
1. Target one test file at a time
2. Implement the beforeEach hooks and fix the tests for that file
3. Ensure those tests pass before moving to the next file
4. Repeat until all files are updated

### Approach 3: Selective enabling based on complexity
This approach would:
1. Categorize the test files by complexity
2. Start with the simplest files first
3. Implement the beforeEach hooks for each file
4. Selectively unskip tests that are ready to run with the new implementation
5. Leave more complex tests skipped until they can be properly addressed

## Selected Approach: Incremental file-by-file approach with focus on gitignoreUtils.test.ts

I'll implement a staged approach that:
1. Focuses first on gitignoreUtils.test.ts since:
   - It directly tests the gitignore functionality
   - It has the most comprehensive tests of gitignore behavior
   - It's the foundation for the other tests
2. Updates the beforeEach hooks to create virtual .gitignore files using addVirtualGitignoreFile
3. Unskips the tests in this file and ensures they pass
4. Creates a reusable pattern that can be applied to the other test files

## Reasoning
This approach is preferred because:

1. **Focused Effort**: By targeting the core gitignoreUtils.test.ts file first, we ensure the fundamental functionality works correctly
2. **Incremental Validation**: Each step builds on the previous, ensuring we don't introduce regressions
3. **Pattern Establishment**: The implementation for gitignoreUtils.test.ts will serve as a template for the other files
4. **Complexity Management**: It handles the most important functionality first, then applies lessons learned to other tests
5. **Test Stability**: It avoids introducing a large number of test failures all at once
6. **Following the Example**: The implementation will closely follow the pattern provided in PLAN_PHASE2.md