# Refactor src/cli/__tests__/run-command.test.ts

## Goal
Ensure the run-command.test.ts file fully utilizes the virtualFsUtils (memfs) approach for filesystem mocking, removing any usage of the legacy mockFsUtils approach.

## Current State Analysis
After examining the file, I've found that this test file is actually already using virtualFsUtils correctly:

1. It properly imports from virtualFsUtils at the top of the file
2. It sets up Jest mocks for fs modules before importing them
3. It uses resetVirtualFs(), getVirtualFs(), and createFsError() from virtualFsUtils
4. It properly initializes the virtual filesystem with test files

The file is listed in the TODO.md and in jest.config.js as needing migration, but it appears this work has already been started or completed. The file already follows the patterns recommended in PLAN_PHASE1.md.

## Implementation Approach
Since the file is already using virtualFsUtils correctly, the main tasks are:

1. **Review Error Testing**: Make sure any error simulation is using the recommended approach with `createFsError` and `jest.spyOn()` rather than direct mocking.
   - Currently the file is using a correct approach with `jest.spyOn(fs, 'access')` and `createFsError()`

2. **Review Virtual Filesystem Setup**: Ensure the test directory structure is set up with createVirtualFs() rather than individual calls to the virtual filesystem
   - Currently the file is using direct virtualFs.mkdirSync() and virtualFs.writeFileSync() calls, which work but could be simplified using createVirtualFs()

3. **Update Imports**: Verify that no mockFsUtils imports remain
   - The file is not importing from mockFsUtils at all

4. **Update jest.config.js**: Remove this file from the testPathIgnorePatterns since it's already properly using virtualFsUtils

## Recommended Changes
1. Refactor the beforeEach() setup to use createVirtualFs() instead of multiple direct calls to virtualFs
2. Update the TODO.md to mark this task as complete
3. Update jest.config.js to remove this file from testPathIgnorePatterns

## Reasoning for This Approach
The file is already mostly following the recommended patterns for using virtualFsUtils. The suggested changes will:

1. Simplify the test setup by using the more declarative createVirtualFs() approach
2. Ensure consistency with other refactored test files
3. Update documentation to reflect the true state of the codebase

This approach will require minimal changes while ensuring the file fully conforms to the new testing patterns.