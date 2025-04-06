# Test Utilities

This directory contains shared utilities for testing the thinktank application. These utilities help reduce duplication across test files and provide a consistent approach to common testing tasks.

## Available Utilities

### File System Mocks (`mockFsUtils.ts`)

Utilities for mocking the Node.js `fs/promises` module in tests. These helpers simplify setting up file system mocks for operations like `readFile`, `writeFile`, `stat`, and more.

Usage example:
```typescript
import { setupMockFs, mockReadFile, mockStat, resetMockFs } from '../../../__tests__/utils/mockFsUtils';

beforeEach(() => {
  resetMockFs();
  setupMockFs();
  
  // Configure specific mock behaviors
  mockReadFile('/path/to/file.txt', 'File content');
  mockStat('/path/to/file.txt', { isFile: () => true, size: 123 });
});
```

### Gitignore Mocks (`mockGitignoreUtils.ts`)

Utilities for mocking the project's gitignore utilities in tests. These helpers make it easy to mock the behavior of `.gitignore` pattern matching.

Usage example:
```typescript
import { setupMockGitignore, mockShouldIgnorePath, resetMockGitignore } from '../../../__tests__/utils/mockGitignoreUtils';

beforeEach(() => {
  resetMockGitignore();
  setupMockGitignore();
  
  // Configure specific mock behaviors
  mockShouldIgnorePath(async (_base, file) => file.endsWith('.log'));
});
```

## Best Practices

1. Always reset mocks between tests using the provided reset functions
2. Call setup functions before configuring specific mock behaviors
3. For complex mocking needs, you can still access the raw Jest mocks via the exported `mockedFs` and `mockedGitignoreUtils` objects
4. Keep test-specific mock configurations in the test files themselves