# Filesystem Testing

This guide covers the recommended patterns for testing filesystem operations in the Thinktank project.

## Virtual Filesystem Approach

Thinktank uses the `memfs` library through `virtualFsUtils` to create an in-memory filesystem for tests, allowing for consistent and isolated testing of filesystem operations without affecting the real filesystem.

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

### Common Error Codes to Test

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

### Testing Expected Error Types

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

For more complex scenarios where many filesystem operations are involved, it can be simpler and more maintainable to mock higher-level functions:

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

## Best Practices

1. **Clean up spies**: Always call `mockRestore()` on your spies after using them
2. **Test specific error types**: Verify your application handles each error type appropriately
3. **Use realistic error codes and messages**: Match real-world scenarios for better testing
4. **Isolate tests**: Reset the virtual filesystem between tests with `resetVirtualFs()`
5. **Mock at the right level**: Choose between low-level fs function mocking or higher-level module mocking based on test complexity
6. **Verify error propagation**: Ensure errors are correctly translated to your application's error types
7. **Test platform-specific behavior**: For functions that behave differently on Windows/macOS/Linux, test each platform by mocking `process.platform`

## Important Considerations

1. **Path Format**: memfs requires paths without leading slashes for compatibility. Use relative paths like `path/to/file.txt` instead of `/path/to/file.txt`.

2. **Mocking Order**: Always place the `jest.mock()` calls before importing any filesystem modules.

3. **Test Isolation**: Always reset the virtual filesystem in `beforeEach()` to avoid test interference.

4. **Error Simulation**: For error conditions, use spies to inject errors rather than trying to create actual error conditions.

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