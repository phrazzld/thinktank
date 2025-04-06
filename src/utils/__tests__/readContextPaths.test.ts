/**
 * Tests for the master readContextPaths function
 */
import fs from 'fs/promises';
import { Stats } from 'fs';
import path from 'path';
import { readContextPaths, ContextFileResult } from '../fileReader';

// Mock dependencies
jest.mock('fs/promises');
jest.mock('../gitignoreUtils', () => ({
  shouldIgnorePath: jest.fn().mockResolvedValue(false)
}));

// Access mocked functions
const mockedFs = jest.mocked(fs);

describe('readContextPaths function', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Mock file system access
    mockedFs.access.mockResolvedValue(undefined);
  });
  
  it('should process a mix of files and directories', async () => {
    // Test paths to process
    const testPaths = [
      '/path/to/file.txt',
      '/path/to/directory'
    ];
    
    // Mock file stats
    mockedFs.stat.mockImplementation(async (filePath) => {
      if (filePath === '/path/to/file.txt') {
        return {
          isFile: () => true,
          isDirectory: () => false,
          size: 1024
        } as Stats;
      }
      
      if (filePath === '/path/to/directory') {
        return {
          isFile: () => false,
          isDirectory: () => true,
          size: 4096
        } as Stats;
      }
      
      // For subdirectory files
      return {
        isFile: () => true,
        isDirectory: () => false,
        size: 512
      } as Stats;
    });
    
    // Mock readdir for directory content
    mockedFs.readdir.mockResolvedValue(['subfile1.txt', 'subfile2.md'] as any);
    
    // Mock file content
    mockedFs.readFile.mockImplementation(async (filePath) => {
      return `Content of ${path.basename(String(filePath))}`;
    });
    
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
    
    // Mock access to throw for nonexistent file
    mockedFs.access.mockImplementation(async (filePath) => {
      if (String(filePath).includes('nonexistent-file.txt')) {
        throw Object.assign(new Error('File not found'), { code: 'ENOENT' });
      }
      return undefined;
    });
    
    // Mock stat to only return for the valid file
    mockedFs.stat.mockImplementation(async (filePath) => {
      if (String(filePath).includes('valid-file.txt')) {
        return {
          isFile: () => true,
          isDirectory: () => false,
          size: 1024
        } as Stats;
      }
      throw new Error('Should not be called for nonexistent file');
    });
    
    // Mock file content for the valid file
    mockedFs.readFile.mockImplementation(async (filePath) => {
      return `Content of ${path.basename(String(filePath))}`;
    });
    
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