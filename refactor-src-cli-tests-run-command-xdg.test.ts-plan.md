# Refactor src/cli/__tests__/run-command-xdg.test.ts

## Goal
Optimize the run-command-xdg.test.ts file to fully utilize the virtualFsUtils (memfs) approach for filesystem mocking, ensuring it properly follows the patterns established in PLAN_PHASE1.md.

## Current State Analysis
After examining the file, I've found that the test file is already using virtualFsUtils for its filesystem mocking:

1. It imports the necessary functions from virtualFsUtils (mockFsModules, resetVirtualFs, getVirtualFs, createFsError)
2. It properly sets up Jest mocks before importing fs modules
3. It uses createFsError and resetVirtualFs() correctly

However, there are some improvements that can be made:

1. It doesn't import and use createVirtualFs(), which is the recommended way to set up the virtual filesystem
2. It uses direct virtualFs.mkdirSync() and virtualFs.writeFileSync() calls instead of the more declarative createVirtualFs() approach
3. It doesn't mock CLI index and logger modules, which can lead to test leakage issues, as was found in run-command.test.ts
4. The MockedFunction typing is used inconsistently

## Implementation Approach
Based on the analysis and following the patterns from the successfully refactored run-command.test.ts, I'll:

1. **Update Imports**: Add createVirtualFs to the imports from virtualFsUtils
2. **Replace Direct FS Calls**: Replace direct virtualFs.mkdirSync() and virtualFs.writeFileSync() calls with createVirtualFs() for a more declarative setup
3. **Add Missing Mocks**: Mock the CLI index and logger modules to prevent test leakage
4. **Consistent Error Testing**: Ensure error testing follows the recommended pattern with createFsError and spies
5. **Maintain Test Logic**: Keep the test assertions and mocking logic the same while only improving the filesystem setup approach

## Reasoning for This Approach
The test file is already using virtualFsUtils for filesystem mocking, so this refactoring is mostly an optimization rather than a wholesale migration. The changes will bring consistency to how filesystem mocking is done across all test files:

1. Using createVirtualFs() provides a more declarative and maintainable way to set up test fixtures
2. Mocking CLI index and logger prevents test leakage and makes tests more robust
3. Following the patterns from other successfully refactored tests ensures consistency across the codebase

These improvements will help ensure maintainability and consistency without changing the underlying test logic or assertions, which are already correctly testing the code's behavior.