# Test Utilities

This directory contains shared utilities for testing the thinktank application. These utilities help reduce duplication across test files and provide a consistent approach to common testing tasks.

## Available Utilities

### Virtual File System (`virtualFsUtils.ts`) - Recommended

Utilities for creating an in-memory filesystem using `memfs`. This is the recommended approach for all new tests and for migrating existing tests away from the legacy mocking approach.

```typescript
import { 
  createVirtualFs, 
  resetVirtualFs, 
  mockFsModules, 
  getVirtualFs,
  createFsError
} from '../../../__tests__/utils/virtualFsUtils';

// Setup mocks - must come before importing fs modules
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Import fs modules after mocking
import fs from 'fs';
import fsPromises from 'fs/promises';

// Reset filesystem before each test
beforeEach(() => {
  resetVirtualFs();
  
  // Setup virtual filesystem with files and directories
  createVirtualFs({
    '/path/to/file.txt': 'File content',
    '/path/to/config.json': '{"key": "value"}',
    '/path/to/empty-dir/': '',  // Creates an empty directory
    '/path/to/nested/file.js': 'console.log("Hello world");'
  });
});
```

### Legacy File System Mocks (`mockFsUtils.ts`) - Deprecated

Utilities for mocking the Node.js `fs/promises` module in tests. These helpers are being phased out in favor of the new `virtualFsUtils` approach. See the [Migration Guide](#migrating-from-mockfsutils-to-virtualfsutils) below for details on how to update your tests.

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

## Migrating from mockFsUtils to virtualFsUtils

This section provides a step-by-step guide for migrating tests from the legacy `mockFsUtils` approach to the new `virtualFsUtils` approach.

### Why Migrate?

The `virtualFsUtils` approach offers several advantages:

1. **Simpler setup**: Define your entire filesystem structure in one go with `createVirtualFs()`
2. **More reliable**: Avoids Jest worker crashes that sometimes occur with the mocking approach
3. **Closer to real behavior**: Uses an actual in-memory filesystem that implements the entire Node.js fs API
4. **Better type safety**: Provides proper TypeScript types for all filesystem operations
5. **Cleaner tests**: Reduces the amount of boilerplate code needed for test setup

### Step 1: Update Import and Setup

**Before (mockFsUtils):**
```typescript
import { 
  resetMockFs, 
  setupMockFs, 
  mockReadFile,
  // other mock functions...
  createFsError 
} from '../../__tests__/utils/mockFsUtils';

describe('My test suite', () => {
  beforeEach(() => {
    resetMockFs();
    setupMockFs();
    
    // Individual mocks for each file/path
    mockReadFile('/path/to/file.txt', 'File content');
    mockReadFile('/path/to/another.json', '{"key": "value"}');
    mockStat('/path/to/file.txt', { isFile: () => true, size: 123 });
    // ...more mocks
  });
  
  // Tests...
});
```

**After (virtualFsUtils):**
```typescript
import { 
  createVirtualFs, 
  resetVirtualFs, 
  mockFsModules 
} from '../../__tests__/utils/virtualFsUtils';

// Setup mocks - must be outside any describe/test blocks
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Import fs modules after mocking
import fs from 'fs';
import fsPromises from 'fs/promises';

describe('My test suite', () => {
  beforeEach(() => {
    resetVirtualFs();
    
    // Define entire filesystem in one go
    createVirtualFs({
      '/path/to/file.txt': 'File content',
      '/path/to/another.json': '{"key": "value"}'
      // No need for separate stat mocks - files exist in the virtual FS
    });
  });
  
  // Tests...
});
```

### Step 2: Update Test Cases

#### Reading Files

**Before:**
```typescript
it('should read a file', async () => {
  mockReadFile('/path/to/file.txt', 'Hello world');
  
  const content = await readFileContent('/path/to/file.txt');
  expect(content).toBe('Hello world');
});
```

**After:**
```typescript
it('should read a file', async () => {
  // File already defined in beforeEach with createVirtualFs
  // Or create a file specifically for this test:
  createVirtualFs({
    '/path/to/file.txt': 'Hello world'
  });
  
  const content = await readFileContent('/path/to/file.txt');
  expect(content).toBe('Hello world');
});
```

#### Checking File Existence

**Before:**
```typescript
it('should check if file exists', async () => {
  mockAccess('/path/to/existing.txt', true);
  mockAccess('/path/to/missing.txt', false, {
    errorCode: 'ENOENT'
  });
  
  expect(await fileExists('/path/to/existing.txt')).toBe(true);
  expect(await fileExists('/path/to/missing.txt')).toBe(false);
});
```

**After:**
```typescript
it('should check if file exists', async () => {
  createVirtualFs({
    '/path/to/existing.txt': 'exists'
  });
  
  expect(await fileExists('/path/to/existing.txt')).toBe(true);
  expect(await fileExists('/path/to/missing.txt')).toBe(false);
});
```

#### Testing Error Scenarios

**Before:**
```typescript
it('should handle permission errors', async () => {
  const error = createFsError('EACCES', 'Permission denied', 'access', '/protected.txt');
  mockAccess('/protected.txt', false, { 
    errorCode: 'EACCES', 
    errorMessage: 'Permission denied' 
  });
  
  await expect(readFileContent('/protected.txt'))
    .rejects.toThrow(/Permission denied/);
});
```

**After:**
```typescript
it('should handle permission errors', async () => {
  // Two options for permission errors:
  
  // Option 1: Use fsPromises.access with a spy to simulate the error
  const accessSpy = jest.spyOn(fsPromises, 'access');
  const error = createFsError('EACCES', 'Permission denied', 'access', '/protected.txt');
  accessSpy.mockRejectedValueOnce(error);
  
  await expect(readFileContent('/protected.txt'))
    .rejects.toThrow(/Permission denied/);
  
  accessSpy.mockRestore();
  
  // Option 2: For functions that call multiple fs methods, you may need to
  // create the file first, then mock a specific operation to fail
  createVirtualFs({
    '/another-protected.txt': 'secret content'
  });
  
  const readFileSpy = jest.spyOn(fsPromises, 'readFile');
  readFileSpy.mockRejectedValueOnce(
    createFsError('EACCES', 'Permission denied', 'readFile', '/another-protected.txt')
  );
  
  await expect(readFileContent('/another-protected.txt'))
    .rejects.toThrow(/Permission denied/);
  
  readFileSpy.mockRestore();
});
```

#### Working with Directories

**Before:**
```typescript
it('should list directory contents', async () => {
  mockReaddir('/path/to/dir', ['file1.txt', 'file2.txt', 'subdir']);
  mockStat('/path/to/dir/file1.txt', { isFile: () => true });
  mockStat('/path/to/dir/file2.txt', { isFile: () => true });
  mockStat('/path/to/dir/subdir', { isDirectory: () => true });
  
  const result = await listFiles('/path/to/dir');
  expect(result).toEqual(['file1.txt', 'file2.txt']);
});
```

**After:**
```typescript
it('should list directory contents', async () => {
  createVirtualFs({
    '/path/to/dir/file1.txt': 'content 1',
    '/path/to/dir/file2.txt': 'content 2',
    '/path/to/dir/subdir/': '' // Creates an empty directory
  });
  
  const result = await listFiles('/path/to/dir');
  expect(result).toEqual(['file1.txt', 'file2.txt']);
});
```

#### Testing File Writing

**Before:**
```typescript
it('should write content to a file', async () => {
  mockMkdir('/path/to', true);
  mockWriteFile('/path/to/output.txt', true);
  
  await writeFile('/path/to/output.txt', 'New content');
  
  expect(mockedFs.writeFile).toHaveBeenCalledWith(
    '/path/to/output.txt', 
    'New content',
    expect.any(Object)
  );
});
```

**After:**
```typescript
it('should write content to a file', async () => {
  // Create parent directory
  createVirtualFs({
    '/path/to/': ''
  });
  
  await writeFile('/path/to/output.txt', 'New content');
  
  // Verify the file was written by reading it back
  const content = await fsPromises.readFile('/path/to/output.txt', 'utf-8');
  expect(content).toBe('New content');
  
  // Or check using getVirtualFs
  const virtualFs = getVirtualFs();
  expect(virtualFs.readFileSync('/path/to/output.txt', 'utf-8')).toBe('New content');
});
```

### Step 3: Advanced Patterns

#### Testing Complex File Operations

When testing functions that use multiple filesystem operations, the virtual filesystem approach is particularly advantageous because you can set up a complete filesystem structure and then let your code interact with it naturally.

**Before:**
```typescript
it('should copy a directory recursively', async () => {
  // Setup source directory structure
  mockReaddir('/source', ['file1.txt', 'subdir']);
  mockStat('/source/file1.txt', { isFile: () => true });
  mockStat('/source/subdir', { isDirectory: () => true });
  mockReaddir('/source/subdir', ['file2.txt']);
  mockStat('/source/subdir/file2.txt', { isFile: () => true });
  
  // Setup file content
  mockReadFile('/source/file1.txt', 'content 1');
  mockReadFile('/source/subdir/file2.txt', 'content 2');
  
  // Setup target directory checks and creation
  mockAccess('/target', false, { errorCode: 'ENOENT' });
  mockMkdir('/target', true);
  mockMkdir('/target/subdir', true);
  
  // Setup write operations
  mockWriteFile('/target/file1.txt', true);
  mockWriteFile('/target/subdir/file2.txt', true);
  
  // Perform the copy
  await copyDirectory('/source', '/target');
  
  // Verify the expected fs operations were called
  expect(mockedFs.mkdir).toHaveBeenCalledWith('/target', expect.any(Object));
  expect(mockedFs.mkdir).toHaveBeenCalledWith('/target/subdir', expect.any(Object));
  expect(mockedFs.writeFile).toHaveBeenCalledWith('/target/file1.txt', 'content 1', expect.any(Object));
  expect(mockedFs.writeFile).toHaveBeenCalledWith('/target/subdir/file2.txt', 'content 2', expect.any(Object));
});
```

**After:**
```typescript
it('should copy a directory recursively', async () => {
  // Setup source directory structure
  createVirtualFs({
    '/source/file1.txt': 'content 1',
    '/source/subdir/file2.txt': 'content 2'
  });
  
  // Perform the copy
  await copyDirectory('/source', '/target');
  
  // Verify the target directory structure and contents
  const virtualFs = getVirtualFs();
  
  // Check directory structure
  expect(virtualFs.existsSync('/target')).toBe(true);
  expect(virtualFs.existsSync('/target/subdir')).toBe(true);
  
  // Check file contents
  expect(virtualFs.readFileSync('/target/file1.txt', 'utf-8')).toBe('content 1');
  expect(virtualFs.readFileSync('/target/subdir/file2.txt', 'utf-8')).toBe('content 2');
});
```

#### Simulating Specific Error Conditions

In some cases, you may need to simulate specific error conditions that cannot be easily created with just the filesystem structure.

**Before:**
```typescript
it('should handle an unexpected error during file write', async () => {
  // Setup mocks
  mockMkdir('/path/to', true);
  
  // Create a disk full error
  const diskFullError = createFsError('ENOSPC', 'No space left on device', 'writeFile', '/path/to/file.txt');
  mockWriteFile('/path/to/file.txt', diskFullError);
  
  // Test the function's error handling
  await expect(writeFile('/path/to/file.txt', 'content'))
    .rejects.toThrow(/No space left on device/);
});
```

**After:**
```typescript
it('should handle an unexpected error during file write', async () => {
  // Create minimal filesystem structure
  createVirtualFs({
    '/path/to/': ''
  });
  
  // Create a disk full error
  const diskFullError = createFsError('ENOSPC', 'No space left on device', 'writeFile', '/path/to/file.txt');
  
  // Mock writeFile to simulate disk full error
  const writeFileSpy = jest.spyOn(fsPromises, 'writeFile');
  writeFileSpy.mockRejectedValueOnce(diskFullError);
  
  // Test the function's error handling
  await expect(writeFile('/path/to/file.txt', 'content'))
    .rejects.toThrow(/No space left on device/);
  
  // Restore the original function
  writeFileSpy.mockRestore();
});
```

### Best Practices for virtualFsUtils

1. **Import Order Matters**: Always import and set up the mocks before importing the fs modules.

2. **Reset in beforeEach**: Always call `resetVirtualFs()` in your `beforeEach` hook to start with a clean filesystem.

3. **Create Files and Directories**: Use `createVirtualFs({...})` to define your filesystem structure. Empty strings as values create directories.

4. **Verify with Direct Access**: Use `getVirtualFs()` to access the virtual filesystem directly for assertions.

5. **Temporary Spies for Errors**: Use Jest spies for temporary error simulation, but always restore them afterward.

6. **Testing File Operations**: Test the outcome of operations (files created, content written), not just implementation details.

7. **Empty Directory Representation**: To create an empty directory, end the path with a slash and use an empty string as the value.

8. **Handling Binary Data**: For binary files, use Buffers or Uint8Arrays as values in `createVirtualFs()`.

## Legacy Patterns (mockFsUtils)

The following patterns are maintained for reference during the transition period.

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

## General Best Practices

1. **Reset Before Each Test**: Always call `resetVirtualFs()` or `resetMockFs()` in your `beforeEach` hook to prevent test cross-contamination.

2. **Setup After Reset**: If using legacy mocks, call `setupMockFs()` after resetting to configure default behaviors.

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
     createVirtualFs, 
     resetVirtualFs, 
     mockFsModules 
   } from '../../__tests__/utils/virtualFsUtils';
   ```

7. **Test Outcome, Not Implementation**: Focus on testing the behavior and outcome of your functions, not their implementation details.

8. **Accessibility**: Make your tests readable and maintainable so future developers can understand them easily.