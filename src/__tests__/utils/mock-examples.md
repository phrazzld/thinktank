# Test Mocking Examples

This document provides real-world examples of how to use the mock utilities in the thinktank project. It focuses on common test scenarios and how to implement them effectively.

## Complete Test Examples

### Testing File Reader Functions

```typescript
import { readFileContent } from '../../utils/fileReader';
import { 
  resetMockFs, 
  setupMockFs, 
  mockAccess, 
  mockReadFile, 
  createFsError,
  mockedFs
} from '../utils/mockFsUtils';

// Mock fs module at the test level
jest.mock('fs/promises');

describe('File Reader Functions', () => {
  const testFilePath = '/path/to/test-file.txt';
  const testContent = 'This is test content';
  
  beforeEach(() => {
    // Reset and setup mocks before each test
    resetMockFs();
    setupMockFs();
  });
  
  it('should read file content successfully', async () => {
    // Configure mocks for the happy path
    mockAccess(testFilePath, true);
    mockReadFile(testFilePath, testContent);
    
    // Call the function under test
    const result = await readFileContent(testFilePath);
    
    // Verify the result
    expect(result).toBe(testContent);
    expect(mockedFs.access).toHaveBeenCalledWith(testFilePath, expect.any(Number));
    expect(mockedFs.readFile).toHaveBeenCalledWith(testFilePath, 'utf-8');
  });
  
  it('should handle file not found errors', async () => {
    // Configure mocks for the error scenario
    mockAccess(testFilePath, false, {
      errorCode: 'ENOENT',
      errorMessage: 'File not found'
    });
    
    // Call function and expect it to throw
    await expect(readFileContent(testFilePath)).rejects.toThrow(/File not found/);
    expect(mockedFs.access).toHaveBeenCalledWith(testFilePath, expect.any(Number));
    expect(mockedFs.readFile).not.toHaveBeenCalled();
  });
  
  it('should handle permission denied errors', async () => {
    // Configure mocks for permission errors
    mockAccess(testFilePath, false, {
      errorCode: 'EACCES',
      errorMessage: 'Permission denied'
    });
    
    // Call function and expect it to throw
    await expect(readFileContent(testFilePath)).rejects.toThrow(/Permission denied/);
  });
});
```

### Testing Directory Reader with Gitignore

```typescript
import path from 'path';
import { readDirectoryContents } from '../../utils/fileReader';
import { 
  resetMockFs, 
  setupMockFs, 
  mockStat, 
  mockReaddir,
  mockReadFile,
  mockAccess
} from '../utils/mockFsUtils';
import {
  resetMockGitignore,
  setupMockGitignore,
  mockShouldIgnorePath
} from '../utils/mockGitignoreUtils';

// Mock required modules
jest.mock('fs/promises');
jest.mock('../../utils/gitignoreUtils');

describe('Directory Reader with Gitignore', () => {
  const testDir = '/path/to/test-dir';
  
  beforeEach(() => {
    // Reset and setup both filesystem and gitignore mocks
    resetMockFs();
    setupMockFs();
    resetMockGitignore();
    setupMockGitignore();
    
    // Setup basic directory structure
    mockReaddir(testDir, [
      'file1.txt',
      'file2.log',
      'important.config',
      'node_modules',
      'src'
    ]);
    
    // Setup subdirectory
    mockReaddir(path.join(testDir, 'src'), ['app.js', 'helper.js']);
    
    // Setup file stats
    const fileStats = {
      isFile: () => true,
      isDirectory: () => false,
      size: 1024
    };
    
    const dirStats = {
      isFile: () => false,
      isDirectory: () => true,
      size: 4096
    };
    
    // Mock stats for each path
    mockStat(testDir, dirStats);
    mockStat(path.join(testDir, 'file1.txt'), fileStats);
    mockStat(path.join(testDir, 'file2.log'), fileStats);
    mockStat(path.join(testDir, 'important.config'), fileStats);
    mockStat(path.join(testDir, 'node_modules'), dirStats);
    mockStat(path.join(testDir, 'src'), dirStats);
    mockStat(path.join(testDir, 'src', 'app.js'), fileStats);
    mockStat(path.join(testDir, 'src', 'helper.js'), fileStats);
    
    // Mock file access
    mockAccess(testDir, true);
    mockAccess(path.join(testDir, 'file1.txt'), true);
    mockAccess(path.join(testDir, 'file2.log'), true);
    mockAccess(path.join(testDir, 'important.config'), true);
    mockAccess(path.join(testDir, 'node_modules'), true);
    mockAccess(path.join(testDir, 'src'), true);
    mockAccess(path.join(testDir, 'src', 'app.js'), true);
    mockAccess(path.join(testDir, 'src', 'helper.js'), true);
    
    // Mock file contents
    mockReadFile(path.join(testDir, 'file1.txt'), 'Content of file1');
    mockReadFile(path.join(testDir, 'file2.log'), 'Content of file2.log');
    mockReadFile(path.join(testDir, 'important.config'), 'Important config content');
    mockReadFile(path.join(testDir, 'src', 'app.js'), 'console.log("Hello");');
    mockReadFile(path.join(testDir, 'src', 'helper.js'), 'function helper() {}');
    
    // Setup gitignore filtering to ignore log files and node_modules
    mockShouldIgnorePath(/\.log$/, true);
    mockShouldIgnorePath(/node_modules/, true);
    mockShouldIgnorePath(/important\.config/, false); // Never ignore this file
  });
  
  it('should read directory contents with gitignore filtering', async () => {
    const results = await readDirectoryContents(testDir);
    
    // Should have 4 results (file1.txt, important.config, src/app.js, src/helper.js)
    // The .log file and node_modules directory should be ignored
    expect(results.length).toBe(4);
    
    // Verify each expected file is found
    expect(results.some(r => r.path.endsWith('file1.txt'))).toBe(true);
    expect(results.some(r => r.path.endsWith('important.config'))).toBe(true);
    expect(results.some(r => r.path.endsWith('app.js'))).toBe(true);
    expect(results.some(r => r.path.endsWith('helper.js'))).toBe(true);
    
    // Verify ignored files are not included
    expect(results.some(r => r.path.endsWith('file2.log'))).toBe(false);
    expect(results.some(r => r.path.includes('node_modules'))).toBe(false);
    
    // Verify content of specific file
    const configFile = results.find(r => r.path.endsWith('important.config'));
    expect(configFile?.content).toBe('Important config content');
  });
});
```

### Testing Complex Error Conditions

```typescript
import { createConfigDirectory } from '../../utils/configHelper';
import { 
  resetMockFs, 
  setupMockFs, 
  mockMkdir, 
  createFsError
} from '../utils/mockFsUtils';

jest.mock('fs/promises');
jest.mock('os', () => ({
  homedir: jest.fn().mockReturnValue('/home/user')
}));

describe('Config Directory Creation', () => {
  beforeEach(() => {
    resetMockFs();
    setupMockFs();
  });
  
  describe('Platform-specific Errors', () => {
    describe('Windows Errors', () => {
      beforeEach(() => {
        // Mock platform as Windows
        Object.defineProperty(process, 'platform', { value: 'win32' });
      });
      
      it('should handle Windows access denied errors', async () => {
        // Mock access denied error with Windows-specific error code
        const configDir = 'C:\\Users\\User\\AppData\\Roaming\\thinktank';
        const winError = createFsError(
          'EACCES',
          'Access is denied',
          'mkdir',
          configDir
        );
        
        mockMkdir(configDir, winError);
        
        // Test error handling with Windows-specific message
        await expect(createConfigDirectory()).rejects.toThrow(/Access is denied/);
        await expect(createConfigDirectory()).rejects.toThrow(/administrative privileges/);
      });
      
      afterEach(() => {
        // Reset platform back to original
        Object.defineProperty(process, 'platform', { value: process.platform });
      });
    });
    
    describe('Unix/Linux Errors', () => {
      beforeEach(() => {
        // Mock platform as Linux
        Object.defineProperty(process, 'platform', { value: 'linux' });
      });
      
      it('should handle Linux permission errors', async () => {
        // Mock permission denied error with Linux-specific patterns
        const configDir = '/home/user/.config/thinktank';
        const linuxError = createFsError(
          'EACCES',
          'Permission denied',
          'mkdir',
          configDir
        );
        
        mockMkdir(configDir, linuxError);
        
        // Test error handling with Linux-specific messaging
        await expect(createConfigDirectory()).rejects.toThrow(/Permission denied/);
      });
      
      afterEach(() => {
        // Reset platform back to original
        Object.defineProperty(process, 'platform', { value: process.platform });
      });
    });
  });
});
```

## Testing Multiple File Operations

```typescript
import { processFiles } from '../../utils/batchProcessor';
import { 
  resetMockFs, 
  setupMockFs, 
  mockAccess,
  mockReadFile, 
  mockWriteFile,
  mockStat
} from '../utils/mockFsUtils';

jest.mock('fs/promises');

describe('Batch File Processing', () => {
  const testFiles = [
    '/path/to/file1.txt',
    '/path/to/file2.txt',
    '/path/to/file3.txt'
  ];
  
  beforeEach(() => {
    resetMockFs();
    setupMockFs();
    
    // Setup files with different content and properties
    testFiles.forEach((file, index) => {
      // All files exist and are accessible
      mockAccess(file, true);
      
      // Setup basic file stats
      mockStat(file, { 
        isFile: () => true, 
        isDirectory: () => false,
        size: 100 * (index + 1) // Different sizes
      });
      
      // Setup file content
      mockReadFile(file, `Content of file ${index + 1}`);
      
      // Allow writing to all files
      mockWriteFile(file, true);
    });
    
    // Make one file read-only to test error handling
    mockWriteFile('/path/to/file3.txt', createFsError(
      'EACCES',
      'Permission denied',
      'writeFile',
      '/path/to/file3.txt'
    ));
  });
  
  it('should process multiple files and handle errors for individual files', async () => {
    const results = await processFiles(testFiles, content => content.toUpperCase());
    
    // Should have results for all files
    expect(results.length).toBe(3);
    
    // Check successful operations
    expect(results[0].success).toBe(true);
    expect(results[1].success).toBe(true);
    
    // Check failed operation
    expect(results[2].success).toBe(false);
    expect(results[2].error?.code).toBe('EACCES');
  });
});
```

## Advanced Mocking Techniques

### Regex Pattern Matching

```typescript
// Match all files in a specific directory
mockReadFile(/^\/path\/to\/configs\//, 'Default config content');

// Match files with specific extension
mockStat(/\.jpg$/, { 
  isFile: () => true, 
  size: 1024 * 1024 // 1MB
});

// Complex pattern: match temporary files with numeric suffix
mockAccess(/\/tmp\/file\d+\.tmp$/, false, {
  errorCode: 'ENOENT'
});
```

### Using Custom Predicates for Gitignore

```typescript
// Configure ignore filter with a custom predicate function
mockCreateIgnoreFilter('/project', (path) => {
  // Complex ignore logic
  if (path.includes('node_modules')) return true;
  if (path.includes('.git')) return true;
  if (path.endsWith('.log') && !path.includes('/logs/')) return true;
  if (path.includes('/tmp/') && !path.includes('important')) return true;
  return false;
});
```

### Testing Binary File Detection

```typescript
// Mock a text file
mockReadFile('/path/to/text.txt', 'Plain text content');

// Mock a binary file using Buffer with null bytes
mockReadFile('/path/to/binary.bin', Buffer.from([
  0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x00, 0x77, 0x6f, 0x72, 0x6c, 0x64
])); // "hello\0world" with a null byte

// Test binary file detection
it('should detect binary files', async () => {
  const textResult = await readContextFile('/path/to/text.txt');
  const binaryResult = await readContextFile('/path/to/binary.bin');
  
  expect(textResult.content).toBe('Plain text content');
  expect(textResult.error).toBeNull();
  
  expect(binaryResult.content).toBeNull();
  expect(binaryResult.error?.code).toBe('BINARY_FILE');
});
```

## Testing Tips

1. **Mock Configuration Per Test**: Configure mocks specifically for each test case to isolate behaviors.

2. **Test Both Success and Failure Paths**: Always test both happy paths and error conditions.

3. **Verify Mock Calls**: Check that the underlying mocked functions were called as expected:
   ```typescript
   expect(mockedFs.readFile).toHaveBeenCalledWith('/path/to/file.txt', 'utf-8');
   ```

4. **Reset Between Tests**: Always reset mocks in beforeEach to avoid cross-test contamination:
   ```typescript
   beforeEach(() => {
     resetMockFs();
     setupMockFs();
   });
   ```

5. **Handle Platform Differences**: When testing platform-specific code, mock the platform property:
   ```typescript
   // Mock as Windows
   Object.defineProperty(process, 'platform', { value: 'win32' });
   
   // Run tests...
   
   // Reset afterward
   Object.defineProperty(process, 'platform', { value: process.platform });
   ```

6. **Clean Up Global Mocks**: Always clean up any global object property mocks:
   ```typescript
   afterEach(() => {
     jest.spyOn(path, 'isAbsolute').mockRestore();
   });
   ```