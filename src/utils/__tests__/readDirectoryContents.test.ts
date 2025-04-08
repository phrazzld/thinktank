/**
 * Tests for the directory reader utility
 */
import path from 'path';
import type { PathLike } from 'fs';
import { 
  resetVirtualFs, 
  getVirtualFs, 
  createFsError,
  mockFsModules,
  createVirtualFs,
  createMockStats
} from '../../__tests__/utils/virtualFsUtils';

// Create partial fileReader mock for readContextFile only
const readContextFileMock = jest.fn().mockResolvedValue({
  path: '',
  content: 'Mocked content',
  error: null
});

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);
jest.mock('../fileReader', () => {
  // Import actual fileReader module
  const actual = jest.requireActual('../fileReader');
  
  // Return a modified version with mocked readContextFile
  return {
    ...actual,
    readContextFile: readContextFileMock
  };
});

// Import actual gitignoreUtils
import * as gitignoreUtils from '../gitignoreUtils';

// Import modules after mocking
import fsPromises from 'fs/promises';
import * as fileReader from '../fileReader';

// Get the functions to use in tests
const { readDirectoryContents } = fileReader;

describe('readDirectoryContents', () => {
  // Use a relative path without leading slash for memfs compatibility
  const testDirPath = 'path/to/test/directory';
  
  // Mock the implementation of readDirectoryContents to use our test mocks
  // This is necessary because the actual implementation uses the real file system
  // even though we've mocked fs/promises modules
  const originalReadDirectoryContents = fileReader.readDirectoryContents;
  beforeEach(() => {
    // For each test, we'll customize the mocked readDirectoryContents implementation
    // This is a hybrid approach - we're keeping the jest.spyOn for the readDirectoryContents function
    // but using virtual filesystem for setup and state
    jest.spyOn(fileReader, 'readDirectoryContents').mockImplementation(originalReadDirectoryContents);
  });
  
  afterAll(() => {
    jest.restoreAllMocks();
  });
  
  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();
    resetVirtualFs();
    
    // Reset readContextFile mock to its default value
    readContextFileMock.mockReset();
    readContextFileMock.mockResolvedValue({
      path: '',
      content: 'Mocked content',
      error: null
    });
    
    // Clear gitignore cache
    gitignoreUtils.clearIgnoreCache();
    
    // Setup a default mock for gitignoreUtils.shouldIgnorePath
    jest.spyOn(gitignoreUtils, 'shouldIgnorePath').mockImplementation((_, filePath) => {
      // Default implementation to ignore node_modules and .git files by path
      return Promise.resolve(
        filePath.includes('node_modules') || 
        filePath.includes('.git')
      );
    });
  });

  describe('Basic Directory Traversal', () => {
    beforeEach(() => {
      const virtualFs = getVirtualFs();
      
      // Create directory structure
      virtualFs.mkdirSync(testDirPath, { recursive: true });
      virtualFs.mkdirSync(path.join(testDirPath, 'subdir'), { recursive: true });
      virtualFs.mkdirSync(path.join(testDirPath, 'node_modules'), { recursive: true });
      virtualFs.mkdirSync(path.join(testDirPath, '.git'), { recursive: true });
      
      // Create files with content (using relative paths)
      virtualFs.writeFileSync(path.join(testDirPath, 'file1.txt'), 'Content of file1.txt');
      virtualFs.writeFileSync(path.join(testDirPath, 'file2.md'), 'Content of file2.md');
      virtualFs.writeFileSync(path.join(testDirPath, 'subdir/nested.txt'), 'Content of nested.txt');
    });
    
    it('should read all files in a directory and return their contents', async () => {
      // Use the virtual file system directly, already set up in beforeEach
      
      // Reset the readContextFile mock with appropriate implementations
      readContextFileMock.mockImplementation((filePath: string) => {
        const pathStr = String(filePath);
        if (pathStr.endsWith('file1.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of file1.txt',
            error: null
          });
        } else if (pathStr.endsWith('file2.md')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of file2.md',
            error: null
          });
        }
        return Promise.resolve({
          path: pathStr,
          content: null,
          error: { code: 'READ_ERROR', message: 'Error reading file' }
        });
      });
      
      const results = await readDirectoryContents(testDirPath);
      
      // Expect to have at least 2 files (file1.txt, file2.md)
      expect(results.length).toBeGreaterThanOrEqual(2);
      
      // But we may also get nested.txt since we're now using the actual recursive traversal function
      // Let's confirm it's finding our specific test files
      
      // Check if files were processed correctly
      const file1Result = results.find(r => r.path.endsWith('file1.txt'));
      const file2Result = results.find(r => r.path.endsWith('file2.md'));
      
      expect(file1Result).toBeDefined();
      expect(file1Result?.content).toBe('Content of file1.txt');
      expect(file1Result?.error).toBeNull();
      
      expect(file2Result).toBeDefined();
      expect(file2Result?.content).toBe('Content of file2.md');
      expect(file2Result?.error).toBeNull();
    });
    
    it('should recursively traverse subdirectories', async () => {
      // Use the virtual file system already setup in beforeEach
      // The directory structure is already configured with subdirectories
      
      // Reset the readContextFile mock with appropriate implementations
      readContextFileMock.mockImplementation((filePath: string) => {
        const pathStr = String(filePath);
        if (pathStr.endsWith('file1.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of file1.txt',
            error: null
          });
        } else if (pathStr.endsWith('nested.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of nested.txt',
            error: null
          });
        }
        return Promise.resolve({
          path: pathStr,
          content: null,
          error: { code: 'READ_ERROR', message: 'Error reading file' }
        });
      });
      
      const results = await readDirectoryContents(testDirPath);
      
      // Should include files from subdirectories
      const nestedFileResult = results.find(r => 
        r.path.includes('nested.txt')
      );
      
      expect(nestedFileResult).toBeDefined();
      expect(nestedFileResult?.content).toBe('Content of nested.txt');
      expect(nestedFileResult?.error).toBeNull();
    });
    
    it('should skip common directories like node_modules and .git', async () => {
      // The test setup already includes node_modules and .git directories
      
      // Create a test file in node_modules to verify it's skipped
      const virtualFs = getVirtualFs();
      virtualFs.writeFileSync(path.join(testDirPath, 'node_modules/package.json'), '{"name": "test"}');
      virtualFs.writeFileSync(path.join(testDirPath, '.git/HEAD'), 'ref: refs/heads/main');
      
      // Setup a spy on fsPromises.readdir to verify it's not called on node_modules or .git
      const readdirSpy = jest.spyOn(fsPromises, 'readdir');
      
      // Reset the readContextFile mock implementation
      readContextFileMock.mockImplementation((filePath: string) => {
        return Promise.resolve({
          path: filePath,
          content: 'Mocked content',
          error: null
        });
      });
      
      const results = await readDirectoryContents(testDirPath);
      
      // Check that files from node_modules and .git are not included
      const nodeModulesFile = results.find(r => r.path.includes('node_modules'));
      const gitFile = results.find(r => r.path.includes('.git'));
      
      expect(nodeModulesFile).toBeUndefined();
      expect(gitFile).toBeUndefined();
      
      // Verify readdir was called (checking the call count is unreliable)
      expect(readdirSpy).toHaveBeenCalled();
      
      // Verify that readdir wasn't called on node_modules or .git
      const nodeModulesCall = readdirSpy.mock.calls.find(
        call => String(call[0]).includes('node_modules')
      );
      const gitCall = readdirSpy.mock.calls.find(
        call => String(call[0]).includes('.git')
      );
      
      expect(nodeModulesCall).toBeUndefined();
      expect(gitCall).toBeUndefined();
      
      // Clean up
      readdirSpy.mockRestore();
    });
  });

  describe('Path Handling', () => {
    it('should handle relative paths by resolving them to absolute paths', async () => {
      // Reset and create files
      resetVirtualFs();
      
      // Create a directory with a relative path
      const relativeTestPath = 'relative/test/path';
      
      // Create the virtual filesystem structure
      createVirtualFs({
        [relativeTestPath + '/']: '',
        [relativeTestPath + '/file.txt']: 'Content of file.txt'
      });
      
      // Reset the readContextFile mock with appropriate implementation
      readContextFileMock.mockImplementation((filePath: string) => {
        const pathStr = String(filePath);
        if (pathStr.includes('file.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of file.txt',
            error: null
          });
        }
        return Promise.resolve({
          path: pathStr,
          content: null,
          error: { code: 'READ_ERROR', message: 'Error reading file' }
        });
      });
      
      // Call the function with the relative path
      const results = await readDirectoryContents(relativeTestPath);
      
      // Verify the results
      expect(results.length).toBe(1);
      expect(results[0].path).toContain('file.txt');
      expect(results[0].content).toBe('Content of file.txt');
    });

    it('should handle path with special characters', async () => {
      // Use path without leading slash for memfs compatibility
      const specialPath = 'path/with spaces and #special characters!';
      
      // Reset and create virtual filesystem with special characters
      resetVirtualFs();
      createVirtualFs({
        [specialPath + '/']: '',
        [specialPath + '/file.txt']: 'File content'
      });
      
      // Setup readContextFile mock
      readContextFileMock.mockImplementation((filePath: string) => {
        const pathStr = String(filePath);
        if (pathStr.includes('file.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: 'File content',
            error: null
          });
        }
        return Promise.resolve({
          path: pathStr,
          content: null,
          error: { code: 'READ_ERROR', message: 'Error reading file' }
        });
      });
      
      const results = await readDirectoryContents(specialPath);
      
      // Should handle the special characters correctly
      expect(results).toHaveLength(1);
      expect(results[0].path.includes('file.txt')).toBe(true);
      expect(results[0].content).toBe('File content');
    });

    it('should handle Windows-style paths', async () => {
      // For memfs, we need to use a normalized path format
      const windowsPath = 'Users/user/Documents/test';
      
      // Reset and create virtual filesystem with Windows-like path
      resetVirtualFs();
      createVirtualFs({
        [windowsPath + '/']: '',
        [windowsPath + '/file.txt']: 'File content'
      });
      
      // Setup readContextFile mock
      readContextFileMock.mockImplementation((filePath: string) => {
        const pathStr = String(filePath);
        if (pathStr.includes('file.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: 'File content',
            error: null
          });
        }
        return Promise.resolve({
          path: pathStr,
          content: null,
          error: { code: 'NOT_FOUND', message: 'File not found' }
        });
      });
      
      const results = await readDirectoryContents(windowsPath);
      
      // Should handle the path correctly
      expect(results).toHaveLength(1);
      expect(results[0].path.includes('file.txt')).toBe(true);
      expect(results[0].content).toBe('File content');
    });
  });

  describe('Error Handling', () => {
    it('should handle directory access errors gracefully', async () => {
      resetVirtualFs();

      // Setup spy to simulate access error
      const accessSpy = jest.spyOn(fsPromises, 'access');
      accessSpy.mockRejectedValueOnce(
        createFsError('EACCES', 'Permission denied', 'access', testDirPath)
      );
      
      const results = await readDirectoryContents(testDirPath);
      
      // Should return error for the directory
      expect(results).toHaveLength(1);
      expect(results[0].path).toBe(testDirPath);
      expect(results[0].content).toBeNull();
      expect(results[0].error).toBeDefined();
      expect(results[0].error?.code).toBe('READ_ERROR');
      expect(results[0].error?.message).toContain('Error reading directory');
      
      // Clean up
      accessSpy.mockRestore();
    });
    
    it('should handle file read errors within directories', async () => {
      // Reset and create a basic directory with files
      resetVirtualFs();
      createVirtualFs({
        [testDirPath + '/']: '',
        [testDirPath + '/file1.txt']: 'Content of file1.txt',
        [testDirPath + '/file2.md']: 'Content of file2.md'
      });
      
      // Setup readContextFile to simulate failure for one file
      readContextFileMock.mockImplementation((filePath: string) => {
        const pathStr = String(filePath);
        if (pathStr.includes('file1.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: null,
            error: { code: 'READ_ERROR', message: 'Failed to read file' }
          });
        } else if (pathStr.includes('file2.md')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of file2.md',
            error: null
          });
        }
        return Promise.resolve({
          path: pathStr,
          content: null,
          error: { code: 'UNKNOWN', message: 'Unexpected file in test' }
        });
      });
      
      const results = await readDirectoryContents(testDirPath);
      
      // Should include both files, but one with error
      expect(results.length).toBeGreaterThanOrEqual(2);
      
      const file1Result = results.find(r => r.path.includes('file1.txt'));
      const file2Result = results.find(r => r.path.includes('file2.md'));
      
      expect(file1Result).toBeDefined();
      expect(file1Result?.content).toBeNull();
      expect(file1Result?.error).toBeDefined();
      expect(file1Result?.error?.code).toBe('READ_ERROR');
      
      expect(file2Result).toBeDefined();
      expect(file2Result?.content).toBe('Content of file2.md');
      expect(file2Result?.error).toBeNull();
    });

    it('should handle directory read errors', async () => {
      // Reset and create a basic directory
      resetVirtualFs();
      createVirtualFs({
        [testDirPath + '/']: ''
      });
      
      // Simulate readdir failure
      const readdirSpy = jest.spyOn(fsPromises, 'readdir');
      readdirSpy.mockRejectedValueOnce(
        createFsError('EIO', 'Failed to read directory', 'readdir', testDirPath)
      );
      
      const results = await readDirectoryContents(testDirPath);
      
      // Should return an error result
      expect(results).toHaveLength(1);
      expect(results[0].path).toBe(testDirPath);
      expect(results[0].content).toBeNull();
      expect(results[0].error).toBeDefined();
      expect(results[0].error?.code).toBe('READ_ERROR');
      
      // Restore the original implementations
      readdirSpy.mockRestore();
    });

    it('should handle non-Error objects in exceptions', async () => {
      // Reset and create a basic directory
      resetVirtualFs();
      createVirtualFs({
        [testDirPath + '/']: ''
      });
      
      // Simulate readdir throwing a non-Error object
      const readdirSpy = jest.spyOn(fsPromises, 'readdir');
      readdirSpy.mockImplementationOnce(() => {
        throw 'Not an error object';
      });
      
      const results = await readDirectoryContents(testDirPath);
      
      // Should still return a structured error result
      expect(results).toHaveLength(1);
      expect(results[0].path).toBe(testDirPath);
      expect(results[0].content).toBeNull();
      expect(results[0].error).toBeDefined();
      expect(results[0].error?.code).toBe('UNKNOWN');
      
      // Restore the original implementations
      readdirSpy.mockRestore();
    });

    it('should handle stat errors for directory entries', async () => {
      // Reset and create a basic directory with files
      resetVirtualFs();
      createVirtualFs({
        [testDirPath + '/']: '',
        [testDirPath + '/file1.txt']: 'Content of file1.txt',
        [testDirPath + '/file2.md']: 'Content of file2.md'
      });
      
      // Mock stat to fail for file1.txt
      const statSpy = jest.spyOn(fsPromises, 'stat');
      statSpy.mockImplementation((pathLike: PathLike) => {
        const pathStr = String(pathLike);
        // First call for the main directory
        if (pathStr === testDirPath) {
          return Promise.resolve(createMockStats(false, 4096)); // directory
        }
        // Fail for file1.txt
        if (pathStr.includes('file1.txt')) {
          return Promise.reject(createFsError('EIO', 'Failed to stat file', 'stat', pathStr));
        }
        // Succeed for file2.md
        if (pathStr.includes('file2.md')) {
          return Promise.resolve(createMockStats(true, 1024)); // file
        }
        
        return Promise.reject(createFsError('ENOENT', 'Unexpected path', 'stat', pathStr));
      });
      
      // Setup readContextFile mock to handle file2.md
      readContextFileMock.mockImplementation((filePath: string) => {
        const pathStr = String(filePath);
        if (pathStr.includes('file2.md')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of file2.md',
            error: null
          });
        }
        return Promise.resolve({
          path: pathStr,
          content: null,
          error: { code: 'READ_ERROR', message: 'Error reading file' }
        });
      });
      
      const results = await readDirectoryContents(testDirPath);
      
      // We should have at least one result (file2.md)
      expect(results.length).toBeGreaterThanOrEqual(1);
      
      // Verify file2.md was read successfully
      const successEntry = results.find(r => r.path.includes('file2.md'));
      expect(successEntry).toBeDefined();
      expect(successEntry?.content).toBe('Content of file2.md');
      expect(successEntry?.error).toBeNull();
      
      // Restore mocks
      statSpy.mockRestore();
    });
  });

  describe('Special Cases', () => {
    it('should handle empty directories', async () => {
      // Reset and create an empty directory
      resetVirtualFs();
      createVirtualFs({
        [testDirPath + '/']: '' // Create empty directory
      });
      
      const results = await readDirectoryContents(testDirPath);
      
      // Should return empty array (no error)
      expect(results).toBeInstanceOf(Array);
      expect(results).toHaveLength(0);
    });

    it('should handle when path is a file, not a directory', async () => {
      // Reset and set up a file instead of directory at testDirPath
      resetVirtualFs();
      createVirtualFs({
        [testDirPath]: 'File content'  // Create file at testDirPath
      });
      
      // Mock stat to indicate this is a file 
      const statSpy = jest.spyOn(fsPromises, 'stat');
      statSpy.mockResolvedValueOnce(createMockStats(true, 12)); // true = file, 12 = size
      
      // Setup readContextFile mock 
      readContextFileMock.mockResolvedValueOnce({
        path: testDirPath,
        content: 'File content',
        error: null
      });
      
      // Run the function under test
      const results = await readDirectoryContents(testDirPath);
      
      // Should return the file content directly
      expect(results).toHaveLength(1);
      expect(results[0].path).toBe(testDirPath);
      expect(results[0].content).toBe('File content');
      expect(results[0].error).toBeNull();
      
      // Clean up
      statSpy.mockRestore();
    });

    it('should handle various file types and extensions', async () => {
      // Reset and setup virtual filesystem with various file types
      resetVirtualFs();
      
      // Create the directory structure with different file types
      createVirtualFs({
        [`${testDirPath}/`]: '',
        [`${testDirPath}/script.js`]: 'console.log("hello");',
        [`${testDirPath}/style.css`]: 'body { color: red; }',
        [`${testDirPath}/data.json`]: '{"key": "value"}',
        [`${testDirPath}/document.md`]: '# Heading'
      });
      
      // Setup mock for readContextFile to return file contents
      readContextFileMock.mockImplementation((filePath: string) => {
        const pathStr = String(filePath);
        if (pathStr.includes('script.js')) {
          return Promise.resolve({
            path: pathStr,
            content: 'console.log("hello");',
            error: null
          });
        } else if (pathStr.includes('style.css')) {
          return Promise.resolve({
            path: pathStr,
            content: 'body { color: red; }',
            error: null
          });
        } else if (pathStr.includes('data.json')) {
          return Promise.resolve({
            path: pathStr,
            content: '{"key": "value"}',
            error: null
          });
        } else if (pathStr.includes('document.md')) {
          return Promise.resolve({
            path: pathStr,
            content: '# Heading',
            error: null
          });
        }
        return Promise.resolve({
          path: pathStr,
          content: null,
          error: { code: 'READ_ERROR', message: 'Error reading file' }
        });
      });
      
      // Run the function being tested
      const results = await readDirectoryContents(testDirPath);
      
      // Verify results
      expect(results.length).toBe(4); // Should find all 4 files
      
      // Verify each file was processed correctly
      const jsFile = results.find(r => r.path.includes('script.js'));
      const cssFile = results.find(r => r.path.includes('style.css'));
      const jsonFile = results.find(r => r.path.includes('data.json'));
      const mdFile = results.find(r => r.path.includes('document.md'));
      
      expect(jsFile).toBeDefined();
      expect(jsFile?.content).toBe('console.log("hello");');
      
      expect(cssFile).toBeDefined();
      expect(cssFile?.content).toBe('body { color: red; }');
      
      expect(jsonFile).toBeDefined();
      expect(jsonFile?.content).toBe('{"key": "value"}');
      
      expect(mdFile).toBeDefined();
      expect(mdFile?.content).toBe('# Heading');
    });
  });

  describe('Integration with Other Features', () => {
    it('should integrate with gitignore-based filtering', async () => {
      // Reset virtual filesystem and create test structure
      resetVirtualFs();
      
      // Create directory structure with test files
      createVirtualFs({
        [`${testDirPath}/`]: '',
        [`${testDirPath}/file1.txt`]: 'Content of file1.txt',
        [`${testDirPath}/file2.md`]: 'Content of file2.md',
        [`${testDirPath}/subdir/`]: '',
        [`${testDirPath}/subdir/nested.txt`]: 'Content of nested.txt',
        [`${testDirPath}/.gitignore`]: 'file1.txt'
      });
      
      // Setup readContextFile mock
      readContextFileMock.mockImplementation((filePath: string) => {
        const pathStr = String(filePath);
        if (pathStr.includes('file2.md')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of file2.md',
            error: null
          });
        } else if (pathStr.includes('.gitignore')) {
          return Promise.resolve({
            path: pathStr,
            content: 'file1.txt',
            error: null
          });
        } else if (pathStr.includes('nested.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of nested.txt',
            error: null
          });
        } else if (pathStr.includes('file1.txt')) {
          return Promise.resolve({
            path: pathStr, 
            content: 'Content of file1.txt',
            error: null
          });
        }
        return Promise.resolve({
          path: pathStr,
          content: null,
          error: { code: 'READ_ERROR', message: 'Error reading file' }
        });
      });
      
      // Mock the gitignore functions for this test
      const shouldIgnorePathSpy = jest.spyOn(gitignoreUtils, 'shouldIgnorePath')
        .mockImplementation((_, filePath: string) => {
          // Simple implementation that checks if the file matches the pattern
          const fileName = path.basename(filePath);
          return Promise.resolve(fileName === 'file1.txt');
        });
      
      // Run the directory traversal
      const results = await readDirectoryContents(testDirPath);
      
      // Test for the presence/absence of specific files
      const file1 = results.find(r => r.path.includes('file1.txt'));
      const file2 = results.find(r => r.path.includes('file2.md'));
      const nested = results.find(r => r.path.includes('nested.txt'));
      const gitignoreFile = results.find(r => r.path.includes('.gitignore'));
      
      // file1.txt should be ignored
      expect(file1).toBeUndefined();
      
      // .gitignore is usually included (not excluded by default)
      expect(gitignoreFile).toBeDefined();
      
      // file2.md and nested.txt should be included
      expect(file2).toBeDefined();
      expect(file2?.content).toBe('Content of file2.md');
      expect(nested).toBeDefined();
      expect(nested?.content).toBe('Content of nested.txt');
      
      // Verify shouldIgnorePath was called with appropriate arguments
      expect(shouldIgnorePathSpy).toHaveBeenCalled();
      
      // Clean up
      shouldIgnorePathSpy.mockRestore();
    });

    it('should detect and handle binary files correctly', async () => {
      // Reset and create directory with binary file
      resetVirtualFs();
      
      // Create the directory structure
      createVirtualFs({
        [`${testDirPath}/`]: '',
        [`${testDirPath}/binary.bin`]: 'Content with \0 null bytes',
        [`${testDirPath}/text.txt`]: 'Normal text content'
      });
      
      // Mock readContextFile to detect binary content
      readContextFileMock.mockImplementation((filePath: string) => {
        const pathStr = String(filePath);
        if (pathStr.includes('binary.bin')) {
          return Promise.resolve({
            path: pathStr,
            content: null,
            error: { code: 'BINARY_FILE', message: 'Binary file detected' }
          });
        } else if (pathStr.includes('text.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Normal text content',
            error: null
          });
        }
        return Promise.resolve({
          path: pathStr,
          content: null,
          error: { code: 'READ_ERROR', message: 'Error reading file' }
        });
      });
      
      // Call the function being tested
      const results = await readDirectoryContents(testDirPath);
      
      // We expect two results: one binary and one text
      expect(results.length).toBe(2);
      
      // Check binary file is properly identified
      const binaryFile = results.find(r => r.path.includes('binary.bin'));
      expect(binaryFile).toBeDefined();
      expect(binaryFile?.content).toBeNull();
      expect(binaryFile?.error).toBeDefined();
      expect(binaryFile?.error?.code).toBe('BINARY_FILE');
      
      // Check text file is properly read
      const textFile = results.find(r => r.path.includes('text.txt'));
      expect(textFile).toBeDefined();
      expect(textFile?.content).toBe('Normal text content');
      expect(textFile?.error).toBeNull();
    });

    it('should handle deep nested directory structures', async () => {
      // Create a deep nested directory structure (5 levels)
      const maxDepth = 5;
      resetVirtualFs();
      
      // Create structure object
      const structure: Record<string, string> = {
        [`${testDirPath}/`]: '',
        [`${testDirPath}/root.txt`]: 'Content of root.txt'
      };
      
      // Create nested directories and files
      let currentPath = testDirPath;
      for (let i = 1; i <= maxDepth; i++) {
        currentPath = `${currentPath}/level${i}`;
        structure[`${currentPath}/`] = ''; // Directory marker
        
        if (i < maxDepth) {
          structure[`${currentPath}/file${i}.txt`] = `Content of file${i}.txt`;
        } else {
          structure[`${currentPath}/finalfile.txt`] = 'Content of finalfile.txt';
        }
      }
      
      // Create virtual filesystem
      createVirtualFs(structure);
      
      // Mock readContextFile to return file content
      readContextFileMock.mockImplementation((filePath: string) => {
        const pathStr = String(filePath);
        
        if (pathStr.includes('root.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of root.txt',
            error: null
          });
        } else if (pathStr.includes('file1.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of file1.txt',
            error: null
          });
        } else if (pathStr.includes('file2.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of file2.txt',
            error: null
          });
        } else if (pathStr.includes('file3.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of file3.txt',
            error: null
          });
        } else if (pathStr.includes('file4.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of file4.txt',
            error: null
          });
        } else if (pathStr.includes('finalfile.txt')) {
          return Promise.resolve({
            path: pathStr,
            content: 'Content of finalfile.txt',
            error: null
          });
        }
        
        return Promise.resolve({
          path: pathStr,
          content: null,
          error: { code: 'READ_ERROR', message: 'Error reading file' }
        });
      });
      
      // Call the function
      const results = await readDirectoryContents(testDirPath);
      
      // We expect 1 + maxDepth files (root file + one file at each level)
      expect(results.length).toBe(1 + maxDepth);
      
      // Check root file
      const rootFile = results.find(r => r.path.includes('root.txt'));
      expect(rootFile).toBeDefined();
      expect(rootFile?.content).toBe('Content of root.txt');
      
      // Check nested files at each level
      for (let i = 1; i < maxDepth; i++) {
        const levelFile = results.find(r => r.path.includes(`file${i}.txt`));
        expect(levelFile).toBeDefined();
        expect(levelFile?.content).toBe(`Content of file${i}.txt`);
      }
      
      // Check the deepest file
      const finalFile = results.find(r => r.path.includes('finalfile.txt'));
      expect(finalFile).toBeDefined();
      expect(finalFile?.content).toBe('Content of finalfile.txt');
    });
  });
});
