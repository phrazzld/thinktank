# Thinktank Testing Guide

This document provides a comprehensive guide for testing the Thinktank application, with a focus on filesystem testing using the virtual filesystem approach.

## Table of Contents
- [Testing Philosophy](#testing-philosophy)
- [Test Suite Organization](#test-suite-organization)
- [Running Tests](#running-tests)
- [Virtual Filesystem Testing](#virtual-filesystem-testing)
- [Testing Error Conditions](#testing-error-conditions)
- [Platform-Specific Testing](#platform-specific-testing)
- [Testing Large Directory Structures](#testing-large-directory-structures)
- [Test Utilities](#test-utilities)
- [Troubleshooting](#troubleshooting)

## Testing Philosophy

Thinktank follows test-driven development (TDD) principles:

1. **Write tests first**: Before implementing a feature or fixing a bug, write tests that define the expected behavior.
2. **Focus on behavior, not implementation**: Test what the code does, not how it does it.
3. **Test edge cases and error conditions**: Don't just test the happy path; test error handling and edge cases.
4. **Isolation**: Tests should be independent of each other and the external environment.

## Test Suite Organization

Tests are organized parallel to the source code structure:

```
src/
├── core/
│   ├── __tests__/       # Tests for core functionality
│   ├── errors/
│   │   ├── __tests__/   # Tests for error handling
├── utils/
│   ├── __tests__/       # Tests for utility functions
├── __tests__/           # Shared test utilities
    ├── utils/
        ├── virtualFsUtils.ts    # Virtual filesystem utilities
        ├── mockGitignoreUtils.ts # Gitignore mocking utilities
```

## Running Tests

The following npm scripts are available for running tests:

```bash
# Run all tests
npm test

# Run tests with coverage report
npm run test:cov

# Run tests with specific debug options
npm run test:debug

# Run a specific test file or test name
npm test -- -t "test name"
```

## Virtual Filesystem Testing

Thinktank uses the `memfs` library to create an in-memory filesystem for tests, allowing for consistent and isolated testing of filesystem operations without affecting the real filesystem.

### Key Benefits

- **Isolation**: Tests don't affect the real filesystem
- **Performance**: Faster than real filesystem operations
- **Consistency**: Tests run the same way in any environment
- **Simplicity**: Easy to set up and tear down test fixtures

### Basic Usage

```typescript
import { 
  createVirtualFs, 
  resetVirtualFs, 
  mockFsModules 
} from '../../__tests__/utils/virtualFsUtils';

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Import fs modules after mocking
import fs from 'fs';
import fsPromises from 'fs/promises';

describe('My test suite', () => {
  beforeEach(() => {
    // Reset the virtual filesystem before each test
    resetVirtualFs();
    
    // Create a test filesystem structure
    createVirtualFs({
      'path/to/file.txt': 'File content',
      'path/to/config.json': '{"key": "value"}',
      'path/to/empty-dir/': '',  // Creates an empty directory
    });
  });
  
  it('should read a file', async () => {
    const content = await fsPromises.readFile('path/to/file.txt', 'utf-8');
    expect(content).toBe('File content');
  });
});
```

### Important Considerations

1. **Path Format**: memfs requires paths without leading slashes for compatibility. Use relative paths like `path/to/file.txt` instead of `/path/to/file.txt`.

2. **Mocking Order**: Always place the `jest.mock()` calls before importing any filesystem modules.

3. **Test Isolation**: Always reset the virtual filesystem in `beforeEach()` to avoid test interference.

4. **Error Simulation**: For error conditions, use spies to inject errors rather than trying to create actual error conditions.

For more detailed examples and patterns, see the [utils/README.md](src/__tests__/utils/README.md) file.

## Testing Error Conditions

Error handling is a critical aspect of filesystem operations. Here's how to effectively test error conditions:

### Using Error Injection

```typescript
it('should handle file not found error', async () => {
  // Setup filesystem without the target file
  resetVirtualFs();
  createVirtualFs({});
  
  // Test the function's error handling
  await expect(readFileContent('non-existent.txt'))
    .rejects.toThrow(/File not found/);
});

it('should handle permission errors', async () => {
  // Create the file first
  createVirtualFs({
    'protected.txt': 'Secret content'
  });
  
  // Inject a permission error
  const accessSpy = jest.spyOn(fsPromises, 'access');
  accessSpy.mockRejectedValueOnce(
    createFsError('EACCES', 'Permission denied', 'access', 'protected.txt')
  );
  
  // Test error handling
  await expect(readFileContent('protected.txt'))
    .rejects.toThrow(/Permission denied/);
  
  // Restore the original function
  accessSpy.mockRestore();
});
```

### Common Error Scenarios to Test

- **File not found**: ENOENT errors
- **Permission denied**: EACCES errors
- **Path is a directory**: Testing when a file operation is performed on a directory
- **Directory not empty**: When trying to remove a non-empty directory
- **Path too long**: Testing very long file paths
- **Disk full**: ENOSPC errors for write operations
- **Too many open files**: EMFILE errors

## Platform-Specific Testing

Thinktank runs on multiple platforms, and filesystem behavior varies between them. Here's how to test platform-specific code:

### Conditional Testing

```typescript
describe('Platform-specific behavior', () => {
  it('should handle Windows paths correctly', () => {
    // Only run this test on Windows or when simulating Windows
    if (process.platform === 'win32') {
      // Windows-specific test
    } else {
      // Skip with a note in non-Windows environments
      console.log('Skipping Windows-specific test on non-Windows platform');
    }
  });
  
  it('should handle Linux and macOS paths', () => {
    if (process.platform !== 'win32') {
      // Unix-like platform test
    } else {
      console.log('Skipping Unix test on Windows platform');
    }
  });
});
```

### Mocking Platform Detection

To test platform-specific code regardless of the actual platform:

```typescript
describe('Windows-specific error handling', () => {
  const originalPlatform = process.platform;
  
  beforeEach(() => {
    // Mock process.platform to simulate Windows
    Object.defineProperty(process, 'platform', {
      value: 'win32'
    });
  });
  
  afterEach(() => {
    // Restore the original platform
    Object.defineProperty(process, 'platform', {
      value: originalPlatform
    });
  });
  
  it('should handle Windows-specific errors', async () => {
    // Now the code will behave as if running on Windows
    // ...test Windows-specific logic
  });
});
```

### Key Platform Differences to Test

- **Path separators**: Windows uses backslashes, Unix uses forward slashes
- **Root directory structure**: Windows has drive letters (C:, D:), Unix has a single root (/)
- **Special directories**: AppData on Windows vs. ~/.config on Unix
- **File permissions**: More complex on Unix systems
- **Case sensitivity**: Case-insensitive on Windows, case-sensitive on Unix (usually)
- **Maximum path length**: Different limits on different platforms

## Testing Large Directory Structures

Testing with large directory structures can cause performance issues or Jest worker crashes. Here are strategies to test effectively:

### Targeted Structure Creation

Instead of creating an entire large directory structure, create only the portions needed for specific tests:

```typescript
it('should handle nested directories efficiently', async () => {
  // Create just the directory paths needed for this test
  const paths = {};
  
  // Create a reasonably sized test structure
  for (let i = 0; i < 5; i++) {
    const dirPath = `level-${i}/`;
    paths[dirPath] = '';
    
    for (let j = 0; j < 3; j++) {
      paths[`${dirPath}file-${j}.txt`] = `Content ${i}-${j}`;
    }
  }
  
  createVirtualFs(paths);
  
  // Test with this structure
  const results = await readDirectoryContents('level-0');
  expect(results.length).toBe(8); // 3 files + 5 subdirectories
});
```

### Mocking Recursive Functions

For testing functions that process large directory structures, consider mocking the recursive part:

```typescript
it('should handle very large directory structures', async () => {
  // Set up a minimal structure
  createVirtualFs({
    'root/': '',
    'root/file1.txt': 'content'
  });
  
  // Mock readDirectoryContents to return a simulated large result
  const readDirSpy = jest.spyOn(fileReader, 'readDirectoryContents');
  readDirSpy.mockImplementationOnce(() => {
    // Return a simulated large result set
    return Promise.resolve([
      // Generate many simulated file results
      ...Array(1000).fill(0).map((_, i) => ({
        path: `root/file${i}.txt`,
        content: `Content ${i}`,
        error: null
      }))
    ]);
  });
  
  // Test the function that would use readDirectoryContents
  const result = await processDirectory('root');
  
  // Verify correct handling of the large result set
  expect(result.fileCount).toBe(1000);
  
  // Restore the original function
  readDirSpy.mockRestore();
});
```

## Test Utilities

Thinktank provides several test utilities to simplify testing:

### Virtual Filesystem Utilities

- `createVirtualFs(structure)` - Create a virtual filesystem structure
- `resetVirtualFs()` - Reset the virtual filesystem
- `getVirtualFs()` - Get direct access to the virtual filesystem
- `mockFsModules()` - Get mock implementations for fs modules
- `createFsError()` - Create standardized filesystem error objects

### Gitignore Utilities

- `setupMockGitignore()` - Set up gitignore mocking
- `resetMockGitignore()` - Reset gitignore mocks
- `mockShouldIgnorePath()` - Configure path-specific ignore behavior
- `mockCreateIgnoreFilter()` - Configure directory-specific ignore filtering

For detailed information on these utilities, see the [utils/README.md](src/__tests__/utils/README.md) file.

## Troubleshooting

### Common Test Issues

1. **Tests interfering with each other**
   - Ensure you're resetting the virtual filesystem in `beforeEach()`
   - Check for global state modifications

2. **Jest worker crashes**
   - Reduce the size of test fixtures
   - Check for memory leaks in mocks
   - Try running with `--runInBand` flag

3. **Unexpected mock behavior**
   - Verify mock setup order (jest.mock before imports)
   - Check if mocks are being restored properly
   - Use `mockImplementationOnce` for one-time mocks

4. **Path-related failures**
   - Remember memfs requires relative paths without leading slashes
   - Use path.join() for cross-platform path handling
   - Check for issues with directory vs file paths

5. **Test timeouts**
   - Look for infinite recursion
   - Check for unresolved promises
   - Consider increasing the timeout for complex tests

For specific help with testing issues, consult the detailed documentation in [src/__tests__/utils/README.md](src/__tests__/utils/README.md) or open an issue on GitHub.