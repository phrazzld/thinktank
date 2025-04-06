# Test Utilities

This directory contains shared utilities for testing the thinktank application. These utilities help reduce duplication across test files and provide a consistent approach to common testing tasks.

## Available Utilities

### File System Mocks (`mockFsUtils.ts`)

Utilities for mocking the Node.js `fs/promises` module in tests. These helpers simplify setting up file system mocks for operations like `readFile`, `writeFile`, `stat`, `readdir`, `mkdir`, and `access`.

```typescript
import { 
  setupMockFs, 
  resetMockFs, 
  mockReadFile, 
  mockStat, 
  mockAccess,
  mockReaddir,
  mockMkdir,
  mockWriteFile,
  createFsError
} from '../../../__tests__/utils/mockFsUtils';

// Reset and setup mocks before each test
beforeEach(() => {
  resetMockFs();
  setupMockFs();
  
  // Configure specific mock behaviors
  mockReadFile('/path/to/file.txt', 'File content');
  mockStat('/path/to/file.txt', { 
    isFile: () => true, 
    isDirectory: () => false,
    size: 123 
  });
  mockAccess('/path/to/file.txt', true); // File is accessible
  mockAccess('/path/to/missing.txt', false, { 
    errorCode: 'ENOENT', 
    errorMessage: 'File not found' 
  });
});
```

### Gitignore Mocks (`mockGitignoreUtils.ts`)

Utilities for mocking the project's gitignore utilities in tests. These helpers make it easy to mock the behavior of `.gitignore` pattern matching and filtering.

```typescript
import { 
  setupMockGitignore, 
  resetMockGitignore, 
  mockShouldIgnorePath,
  mockCreateIgnoreFilter 
} from '../../../__tests__/utils/mockGitignoreUtils';

// Reset and setup mocks before each test
beforeEach(() => {
  resetMockGitignore();
  setupMockGitignore();
  
  // Configure specific mock behaviors
  mockShouldIgnorePath(/\.log$/, true); // Ignore log files
  mockShouldIgnorePath(/important\.txt$/, false); // Never ignore important.txt
  
  // Configure a directory-specific ignore filter
  mockCreateIgnoreFilter('/home/project', ['node_modules', '*.tmp']);
});
```

## Common Use Cases and Patterns

### Mocking File Access Scenarios

```typescript
// Happy path - file exists and is accessible
mockAccess('/path/to/file.txt', true);

// File not found scenario
mockAccess('/path/to/missing.txt', false, {
  errorCode: 'ENOENT',
  errorMessage: 'File not found'
});

// Permission denied scenario
mockAccess('/path/to/protected.txt', false, {
  errorCode: 'EACCES',
  errorMessage: 'Permission denied'
});
```

### Mocking File Content

```typescript
// Text file content
mockReadFile('/path/to/text.txt', 'Hello, world!');

// Binary file content (using Buffer)
mockReadFile('/path/to/binary.bin', Buffer.from([0x00, 0xFF, 0x42]));

// File read error
mockReadFile('/path/to/error.txt', createFsError(
  'ENOENT',
  'File not found',
  'readFile',
  '/path/to/error.txt'
));
```

### Mocking Directory Listing

```typescript
// Empty directory
mockReaddir('/path/to/empty', []);

// Directory with files
mockReaddir('/path/to/dir', ['file1.txt', 'file2.txt', 'subdir']);

// Directory read error
mockReaddir('/path/to/error', createFsError(
  'EACCES',
  'Permission denied',
  'readdir',
  '/path/to/error'
));
```

### Mocking File/Directory Stats

```typescript
// File stats
mockStat('/path/to/file.txt', {
  isFile: () => true,
  isDirectory: () => false,
  size: 1024
});

// Directory stats
mockStat('/path/to/dir', {
  isFile: () => false,
  isDirectory: () => true,
  size: 4096
});

// Stat error
mockStat('/path/to/error', createFsError(
  'ENOENT',
  'No such file or directory',
  'stat',
  '/path/to/error'
));
```

### Mocking Directory Creation

```typescript
// Successful directory creation
mockMkdir('/path/to/new', true);

// Failed directory creation
mockMkdir('/path/to/error', createFsError(
  'EACCES', 
  'Permission denied', 
  'mkdir', 
  '/path/to/error'
));
```

### Mocking File Writing

```typescript
// Successful file write
mockWriteFile('/path/to/writeable.txt', true);

// Failed file write
mockWriteFile('/path/to/readonly.txt', createFsError(
  'EACCES',
  'Permission denied',
  'writeFile',
  '/path/to/readonly.txt'
));
```

### Mocking Gitignore Patterns

```typescript
// Ignore all log files
mockShouldIgnorePath(/\.log$/, true);

// Ignore files in the node_modules directory
mockShouldIgnorePath(/node_modules\//, true);

// Never ignore important files regardless of location
mockShouldIgnorePath(/important\.config$/, false);

// Use regex pattern for more complex matching
mockShouldIgnorePath(/tmp\/.*\.bak$/, true); // Ignore .bak files in tmp directory
```

### Creating Ignore Filters

```typescript
// Configure an ignore filter for a specific directory
mockCreateIgnoreFilter('/home/project', [
  'node_modules',
  'dist',
  '*.log',
  'tmp'
]);

// Or use a custom function for more complex logic
mockCreateIgnoreFilter('/home/project', (path) => {
  return path.includes('node_modules') || 
         path.endsWith('.log') ||
         path.includes('.git');
});
```

## Best Practices

1. **Reset Before Each Test**: Always call `resetMockFs()` and/or `resetMockGitignore()` in your `beforeEach` hook to prevent test cross-contamination.

2. **Setup After Reset**: Call `setupMockFs()` and/or `setupMockGitignore()` after resetting to configure default behaviors.

3. **Use Error Utility**: Use `createFsError()` consistently for creating filesystem errors with proper codes and messages.

4. **Prefer Regex-Based Testing**: Use regex patterns for error message testing to make tests more resilient to message wording changes.

5. **Properly Clean Up**: If you add any global mocks during tests, clean them up in `afterEach` hooks.

6. **Import Organization**: Follow the project's conventions for import organization:
   ```typescript
   // Node built-in modules
   import path from 'path';
   
   // Project utilities
   import { readFile } from '../fileReader';
   
   // Mock utilities 
   import { 
     resetMockFs, 
     setupMockFs, 
     mockReadFile 
   } from '../../__tests__/utils/mockFsUtils';
   ```

7. **Consistent Pattern Matching**: When mocking multiple functions for the same path, use the same pattern format consistently (string vs regex).

8. **Advanced Usage**: For more complex scenarios, you can still access the raw Jest mocks via the exported `mockedFs` and `mockedGitignoreUtils` objects.