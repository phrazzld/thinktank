# Review and refactor remaining tests with FS dependencies

## Goal
Systematically identify and refactor all remaining test files that use filesystem operations to use the virtualFsUtils approach instead of direct mocking.

## Implementation Approach
After analyzing the codebase, I've identified several test files that still need to be refactored to use the virtualFsUtils approach. The priority order is based on complexity and dependency relationships:

1. `output-directory.test.ts` - Integration tests for output directory functionality
2. `inputHandler.test.ts` - Tests for input handling with file operations
3. `run-command.test.ts` - Command tests for file operations
4. `run-command-xdg.test.ts` - Command tests for XDG configuration paths

For each file, I will:

1. Replace direct fs mocking with `mockFsModules()`
2. Set up proper virtual filesystem initialization in beforeEach hooks
3. Use `resetVirtualFs()` to restore the filesystem between tests
4. Create a realistic directory/file structure in the virtual filesystem
5. Replace spy-based mocking with real filesystem operations
6. Verify operations actually affected the virtual filesystem
7. Use `createFsError()` to generate realistic filesystem errors

## Key Reasoning
I'm focusing on these files first because:

1. They contain direct fs/promises mocking which can be replaced with virtualFsUtils
2. They test critical functionality (file reading, directory creation, config loading)
3. They follow similar patterns to files that have already been successfully refactored
4. The E2E tests are intentionally left for later as they may need a different approach

The chosen implementation approach aligns with the established pattern in previously refactored tests, maintaining consistency throughout the codebase. This approach provides several advantages:
- More realistic testing of filesystem operations
- Better isolation between tests through proper filesystem reset
- Less reliance on spy-based assertions and more focus on actual filesystem state
- Standardized error simulation through the createFsError utility