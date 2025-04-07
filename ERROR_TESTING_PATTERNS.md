# Error Testing Patterns

This document outlines the recommended patterns for testing error scenarios in the Thinktank project, focusing on filesystem errors. Following these patterns ensures consistent, realistic, and maintainable tests.

## Recommended Pattern

### 1. Import Required Utilities

```typescript
import { mockFsModules, resetVirtualFs, createVirtualFs, getVirtualFs, createFsError } from '../../__tests__/utils/virtualFsUtils';

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Now import fs after mocking
import fs from 'fs';
import fsPromises from 'fs/promises';
```

### 2. Setup Test Environment

```typescript
beforeEach(() => {
  jest.clearAllMocks();
  resetVirtualFs();
  
  // Create basic filesystem structure needed for tests
  createVirtualFs({
    '/path/to/file.txt': 'file content',
    '/path/to/directory/': '',  // Empty string creates a directory
  });
});
```

### 3. Testing File System Errors with Spies

```typescript
it('should handle permission denied errors', async () => {
  // Spy on the specific fs function you want to mock
  const writeFileSpy = jest.spyOn(fsPromises, 'writeFile');
  
  // Mock rejection with a realistic error using createFsError
  writeFileSpy.mockRejectedValueOnce(
    createFsError('EACCES', 'Permission denied', 'writeFile', '/path/to/file.txt')
  );
  
  // Test the function that would encounter this error
  await expect(yourFunction('/path/to/file.txt', 'new content'))
    .rejects.toThrow(/Permission denied/);
  
  // Clean up the spy to avoid affecting other tests
  writeFileSpy.mockRestore();
});
```

### 4. Common Error Codes to Test

Test a variety of error scenarios relevant to your functionality:

```typescript
// File not found
createFsError('ENOENT', 'No such file or directory', 'readFile', '/path/to/nonexistent.txt')

// Permission denied
createFsError('EACCES', 'Permission denied', 'writeFile', '/path/to/file.txt')

// Operation not permitted
createFsError('EPERM', 'Operation not permitted', 'unlink', '/path/to/file.txt')

// Read-only file system
createFsError('EROFS', 'Read-only file system', 'mkdir', '/path/to/directory')

// Directory not empty
createFsError('ENOTEMPTY', 'Directory not empty', 'rmdir', '/path/to/directory')

// File or directory busy
createFsError('EBUSY', 'Resource busy or locked', 'rmdir', '/path/to/directory')

// Disk full
createFsError('ENOSPC', 'No space left on device', 'writeFile', '/path/to/file.txt')
```

### 5. Testing Expected Error Types

If your application wraps native errors in custom error types, test that error translation happens correctly:

```typescript
it('should wrap permission errors in FileAccessError', async () => {
  const statSpy = jest.spyOn(fsPromises, 'stat');
  statSpy.mockRejectedValueOnce(
    createFsError('EACCES', 'Permission denied', 'stat', '/path/to/file.txt')
  );
  
  try {
    await yourFunction('/path/to/file.txt');
    fail('Expected function to throw');
  } catch (error) {
    expect(error).toBeInstanceOf(FileAccessError);
    expect(error.code).toBe('EACCES');
    expect(error.message).toContain('Permission denied');
    expect(error.path).toBe('/path/to/file.txt');
  }
  
  statSpy.mockRestore();
});
```

## Alternative Pattern for Complex Scenarios

For more complex scenarios like the output-directory.test.ts file, where many filesystem operations are involved, it can be simpler and more maintainable to mock higher-level functions:

```typescript
// Mock a higher-level module
jest.mock('../myModule', () => {
  const originalModule = jest.requireActual('../myModule');
  return {
    ...originalModule,
    writeToDirectory: jest.fn().mockImplementation(async (options) => {
      // Simulate normal behavior
      if (options.path === '/error/path') {
        throw new Error('Simulated error');
      }
      return 'Success';
    })
  };
});

// Then in your test
it('should handle errors from the module', async () => {
  await expect(yourFunctionThatUsesTheModule('/error/path'))
    .rejects.toThrow('Simulated error');
});
```

## Best Practices

1. **Clean up spies**: Always call `mockRestore()` on your spies after using them
2. **Test specific error types**: Verify your application handles each error type appropriately
3. **Use realistic error codes and messages**: Match real-world scenarios for better testing
4. **Isolate tests**: Reset the virtual filesystem between tests with `resetVirtualFs()`
5. **Mock at the right level**: Choose between low-level fs function mocking or higher-level module mocking based on test complexity
6. **Verify error propagation**: Ensure errors are correctly translated to your application's error types
7. **Test platform-specific behavior**: For functions that behave differently on Windows/macOS/Linux, test each platform by mocking `process.platform`