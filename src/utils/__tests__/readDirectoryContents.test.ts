/**
 * Tests for the directory reader utility
 */
import path from 'path';
import fs from 'fs/promises';
import { Stats } from 'fs';
import { readDirectoryContents } from '../fileReader';

// Mock fs.promises module
jest.mock('fs/promises');

const mockedFs = jest.mocked(fs);

describe('readDirectoryContents', () => {
  const testDirPath = '/path/to/test/directory';
  
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Mock directory entries for readdir
    mockedFs.readdir.mockResolvedValue([
      'file1.txt',
      'file2.md',
      'subdir',
      'node_modules',
      '.git'
    ] as any);
    
    // Mock file stats for different types of entries
    const fileStats = {
      isFile: () => true,
      isDirectory: () => false,
      size: 1024
    } as Stats;
    
    const dirStats = {
      isFile: () => false,
      isDirectory: () => true,
      size: 4096
    } as Stats;
    
    // Setup stat mock to return different results based on path
    mockedFs.stat.mockImplementation(async (filePath) => {
      // Safe cast for the test mock - we know we're only passing strings in our tests
      const pathStr = String(filePath);
      if (pathStr.includes('file1.txt') || pathStr.includes('file2.md') || 
          pathStr.includes('subdir/nested.txt')) {
        return fileStats;
      }
      return dirStats;
    });
    
    // Mock successful file read for file1.txt
    mockedFs.readFile.mockImplementation(async (filePath) => {
      // Safe cast for the test mock - we know we're only passing strings in our tests
      const pathStr = String(filePath);
      if (pathStr.includes('file1.txt')) {
        return 'Content of file1.txt';
      } else if (pathStr.includes('file2.md')) {
        return 'Content of file2.md';
      } else if (pathStr.includes('nested.txt')) {
        return 'Content of nested.txt';
      }
      throw new Error('Unexpected file');
    });
    
    // Mock successful access
    mockedFs.access.mockResolvedValue(undefined);
    
    // For recursive test, mock a nested structure in the subdirectory
    mockedFs.readdir.mockImplementation(async (dirPath) => {
      // Safe cast for the test mock - we know we're only passing strings in our tests
      const pathStr = String(dirPath);
      if (pathStr.includes('/subdir')) {
        return ['nested.txt'] as any;
      }
      return [
        'file1.txt',
        'file2.md',
        'subdir',
        'node_modules',
        '.git'
      ] as any;
    });
  });
  
  it('should read all files in a directory and return their contents', async () => {
    const results = await readDirectoryContents(testDirPath);
    
    // Should find both files in the directory (excluding ignored dirs)
    expect(results).toHaveLength(3); // file1.txt, file2.md, and subdir/nested.txt
    
    // Check if files were processed correctly
    const file1Result = results.find(r => r.path === path.join(testDirPath, 'file1.txt'));
    const file2Result = results.find(r => r.path === path.join(testDirPath, 'file2.md'));
    
    expect(file1Result).toBeDefined();
    expect(file1Result?.content).toBe('Content of file1.txt');
    expect(file1Result?.error).toBeNull();
    
    expect(file2Result).toBeDefined();
    expect(file2Result?.content).toBe('Content of file2.md');
    expect(file2Result?.error).toBeNull();
  });
  
  it('should recursively traverse subdirectories', async () => {
    const results = await readDirectoryContents(testDirPath);
    
    // Should include files from subdirectories
    const nestedFileResult = results.find(r => 
      r.path === path.join(testDirPath, 'subdir', 'nested.txt')
    );
    
    expect(nestedFileResult).toBeDefined();
    expect(nestedFileResult?.content).toBe('Content of nested.txt');
    expect(nestedFileResult?.error).toBeNull();
  });
  
  it('should skip common directories like node_modules and .git', async () => {
    await readDirectoryContents(testDirPath);
    
    // Check if readdir was called on the main directory but not on ignored directories
    expect(mockedFs.readdir).toHaveBeenCalledWith(testDirPath);
    expect(mockedFs.readdir).toHaveBeenCalledWith(path.join(testDirPath, 'subdir'));
    expect(mockedFs.readdir).not.toHaveBeenCalledWith(path.join(testDirPath, 'node_modules'));
    expect(mockedFs.readdir).not.toHaveBeenCalledWith(path.join(testDirPath, '.git'));
  });
  
  it('should handle directory access errors gracefully', async () => {
    // Mock directory access error
    mockedFs.access.mockRejectedValueOnce(new Error('Permission denied') as NodeJS.ErrnoException);
    
    const accessResults = await readDirectoryContents(testDirPath);
    
    // Should return error for the directory
    expect(accessResults).toHaveLength(1);
    expect(accessResults[0].path).toBe(testDirPath);
    expect(accessResults[0].content).toBeNull();
    expect(accessResults[0].error).toBeDefined();
    expect(accessResults[0].error?.code).toBe('READ_ERROR');
    expect(accessResults[0].error?.message).toContain('Error reading directory');
  });
  
  it('should handle file read errors within directories', async () => {
    // Mock readdir success but readFile failure for one file
    mockedFs.readFile.mockImplementation(async (filePath) => {
      // Safe cast for the test mock - we know we're only passing strings in our tests
      const pathStr = String(filePath);
      if (pathStr.includes('file1.txt')) {
        throw new Error('Failed to read file');
      } else if (pathStr.includes('file2.md')) {
        return 'Content of file2.md';
      } else if (pathStr.includes('nested.txt')) {
        return 'Content of nested.txt';
      }
      throw new Error('Unexpected file');
    });
    
    const results = await readDirectoryContents(testDirPath);
    
    // Should include both files, but one with error
    expect(results).toHaveLength(3);
    
    const file1Result = results.find(r => r.path === path.join(testDirPath, 'file1.txt'));
    const file2Result = results.find(r => r.path === path.join(testDirPath, 'file2.md'));
    
    expect(file1Result).toBeDefined();
    expect(file1Result?.content).toBeNull();
    expect(file1Result?.error).toBeDefined();
    expect(file1Result?.error?.code).toBe('READ_ERROR');
    
    expect(file2Result).toBeDefined();
    expect(file2Result?.content).toBe('Content of file2.md');
    expect(file2Result?.error).toBeNull();
  });
});