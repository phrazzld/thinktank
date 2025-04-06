/**
 * Tests for the master readContextPaths function
 */
import { readContextPaths, ContextFileResult } from '../fileReader';
import { 
  resetMockFs, 
  setupMockFs, 
  mockAccess, 
  mockStat, 
  mockReaddir,
  mockReadFile,
  mockedFs
} from '../../__tests__/utils/mockFsUtils';

// Mock dependencies
jest.mock('fs/promises');
jest.mock('../gitignoreUtils', () => ({
  shouldIgnorePath: jest.fn().mockResolvedValue(false)
}));

describe('readContextPaths function', () => {
  beforeEach(() => {
    // Reset and setup mocks before each test
    resetMockFs();
    setupMockFs();
  });
  
  it('should process a mix of files and directories', async () => {
    // Test paths to process
    const testPaths = [
      '/path/to/file.txt',
      '/path/to/directory'
    ];
    
    // Set up access for all test paths
    mockAccess('/path/to/file.txt', true);
    mockAccess('/path/to/directory', true);
    mockAccess('/path/to/directory/subfile1.txt', true);
    mockAccess('/path/to/directory/subfile2.md', true);
    
    // Mock file stats
    mockStat('/path/to/file.txt', {
      isFile: () => true,
      isDirectory: () => false,
      size: 1024
    });
    
    mockStat('/path/to/directory', {
      isFile: () => false,
      isDirectory: () => true,
      size: 4096
    });
    
    mockStat('/path/to/directory/subfile1.txt', {
      isFile: () => true,
      isDirectory: () => false,
      size: 512
    });
    
    mockStat('/path/to/directory/subfile2.md', {
      isFile: () => true,
      isDirectory: () => false,
      size: 512
    });
    
    // Mock readdir for directory content
    mockReaddir('/path/to/directory', ['subfile1.txt', 'subfile2.md']);
    
    // Mock file content
    mockReadFile('/path/to/file.txt', 'Content of file.txt');
    mockReadFile('/path/to/directory/subfile1.txt', 'Content of subfile1.txt');
    mockReadFile('/path/to/directory/subfile2.md', 'Content of subfile2.md');
    
    // Call the function
    const results = await readContextPaths(testPaths);
    
    // Should have 3 results (1 file + 2 directory files)
    expect(results.length).toBe(3);
    
    // Verify individual file was processed
    expect(results.some((r: ContextFileResult) => r.path === '/path/to/file.txt')).toBe(true);
    
    // Verify directory files were processed
    expect(results.some((r: ContextFileResult) => r.path.includes('subfile1.txt'))).toBe(true);
    expect(results.some((r: ContextFileResult) => r.path.includes('subfile2.md'))).toBe(true);
    
    // Verify content was read
    expect(results.find((r: ContextFileResult) => r.path === '/path/to/file.txt')?.content).toBe('Content of file.txt');
  });
  
  it('should handle empty paths array', async () => {
    const results = await readContextPaths([]);
    expect(results).toEqual([]);
    expect(mockedFs.stat).not.toHaveBeenCalled();
  });
  
  it('should handle errors for individual paths', async () => {
    // Test paths with one invalid path
    const testPaths = [
      '/path/to/valid-file.txt',
      '/path/to/nonexistent-file.txt'
    ];
    
    // Mock access to succeed for valid file and fail for nonexistent file
    mockAccess('/path/to/valid-file.txt', true);
    mockAccess('/path/to/nonexistent-file.txt', false, {
      errorCode: 'ENOENT',
      errorMessage: 'File not found'
    });
    
    // Mock stat for the valid file
    mockStat('/path/to/valid-file.txt', {
      isFile: () => true,
      isDirectory: () => false,
      size: 1024
    });
    
    // Mock file content for the valid file
    mockReadFile('/path/to/valid-file.txt', 'Content of valid-file.txt');
    
    const results = await readContextPaths(testPaths);
    
    // Should still have 2 results, but one with an error
    expect(results.length).toBe(2);
    
    // Verify valid file was processed correctly
    const validResult = results.find((r: ContextFileResult) => r.path === '/path/to/valid-file.txt');
    expect(validResult?.content).toBe('Content of valid-file.txt');
    expect(validResult?.error).toBeNull();
    
    // Verify error file has appropriate error info
    const errorResult = results.find((r: ContextFileResult) => r.path === '/path/to/nonexistent-file.txt');
    expect(errorResult?.content).toBeNull();
    expect(errorResult?.error?.code).toBe('ENOENT');
  });
});