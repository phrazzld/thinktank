# Gitignore Cache Clearing in Tests

## Overview

The thinktank project uses a caching mechanism for gitignore patterns to improve performance. In tests, this cache can cause test interdependencies if not properly cleared. This document outlines the current best practices for clearing the gitignore cache in tests.

## Why Cache Clearing Matters

The gitignore functionality in the codebase uses a cache (`ignoreCache` in `gitignoreUtils.ts`) to avoid repeatedly parsing the same gitignore files. While this improves performance in production, it can cause problems in tests:

- Tests that add or modify gitignore files might get stale results if the cache isn't cleared
- Order-dependent test failures can occur if one test depends on gitignore state from a previous test
- Debugging can be more difficult if the cache state isn't reset between tests

## Standard Approaches

There are three approved ways to handle gitignore cache clearing in tests:

### 1. Using `setupTestHooks()`

This is the **preferred approach** for new tests and refactoring. The `setupTestHooks()` function from `test/setup/common.ts` sets up standard hooks that handle cache clearing:

```typescript
import { setupTestHooks } from '../../../test/setup/common';

describe('My test suite', () => {
  setupTestHooks(); // Sets up hooks including gitignore cache clearing
  
  it('should do something', () => {
    // Test with a clean cache
  });
});
```

### 2. Using `setupGitignoreMocking()` 

For tests that need more control over the gitignore setup or more extensive mocking, use `setupGitignoreMocking()` from `src/__tests__/utils/fsTestSetup.ts`:

```typescript
import * as gitignoreUtils from '../../utils/gitignoreUtils';
import { setupGitignoreMocking } from '../utils/fsTestSetup';

describe('My test suite', () => {
  const mockedFileExists = jest.fn();
  
  beforeEach(() => {
    // Other setup...
    setupGitignoreMocking(gitignoreUtils, mockedFileExists);
  });
});
```

### 3. Directly calling `clearIgnoreCache()`

For more complex tests or those with special requirements, you can call `clearIgnoreCache()` directly:

```typescript
import { clearIgnoreCache } from '../../utils/gitignoreUtils';

describe('My test suite', () => {
  beforeEach(() => {
    // Other setup...
    clearIgnoreCache();
  });
});
```

## Guidelines for When to Use Each Approach

1. **New tests**: Always use `setupTestHooks()` for new tests that interact with gitignore functionality
   
2. **Existing tests**:
   - If the test uses standard mocking: continue using the existing approach
   - If refactoring: consider migrating to `setupTestHooks()`
   
3. **Special cases**:
   - Tests that need custom hooks or have special requirements: use direct `clearIgnoreCache()` calls
   - Tests with complex mocking needs: use `setupGitignoreMocking()`

## Integration with Other Test Helpers

The gitignore cache clearing is integrated with other test setup helpers:

- `setupWithGitignore()` in `test/setup/gitignore.ts` - Sets up a directory with files and a .gitignore file
- `setupMultiGitignore()` in `test/setup/gitignore.ts` - Sets up a project with multiple .gitignore files
- `createIgnoreChecker()` in `test/setup/gitignore.ts` - Creates a helper to check if paths should be ignored

When using these helpers, you should still ensure the cache is cleared in your `beforeEach` hooks, typically by using `setupTestHooks()`.

## Verifying Proper Cache Clearing

To verify a test properly clears the gitignore cache:

1. Check if it uses `setupTestHooks()`
2. If not, check if it uses `setupGitignoreMocking()`
3. If neither, check if it directly calls `clearIgnoreCache()` in a `beforeEach` hook

All tests that interact with gitignore functionality should use one of these approaches.