/**
 * Integration tests for gitignore filtering within directory traversal
 */
import fs from 'fs/promises';
import path from 'path';
import { Stats } from 'fs';
import { readDirectoryContents } from '../fileReader';
import * as gitignoreUtils from '../gitignoreUtils';

// Mock dependencies
jest.mock('fs/promises');
jest.mock('../gitignoreUtils');

// Access mocked functions
const mockedFs = jest.mocked(fs);
const mockedGitignoreUtils = jest.mocked(gitignoreUtils);

describe('Gitignore Filtering Integration', () => {
  const testDirPath = '/path/to/test/directory';
  
  beforeEach(() => {
    jest.clearAllMocks();
  });
  
  it('should filter files based on gitignore patterns', async () => {
    // Set up mock file system
    mockedFs.access.mockResolvedValue(undefined);
    
    // Set up mock directory structure
    mockedFs.readdir.mockImplementation(async (dirPath) => {
      const pathStr = String(dirPath);
      
      if (pathStr === testDirPath) {
        return [
          'file1.txt',
          'file2.md',
          'ignored.log',
          '.gitignore'
        ] as any;
      }
      
      return [] as any;
    });
    
    // Set up stat mock for files
    mockedFs.stat.mockImplementation(async (filePath) => {
      const filePathStr = String(filePath);
      
      if (filePathStr === testDirPath) {
        return {
          isFile: () => false,
          isDirectory: () => true,
          size: 4096
        } as Stats;
      }
      
      return {
        isFile: () => true,
        isDirectory: () => false,
        size: 1024
      } as Stats;
    });
    
    // Set up file read mock
    mockedFs.readFile.mockImplementation(async (filePath) => {
      const filePathStr = String(filePath);
      
      if (filePathStr.endsWith('.gitignore')) {
        return '*.log';
      }
      
      return `Content of ${path.basename(filePathStr)}`;
    });
    
    // Set up gitignore filtering mock
    mockedGitignoreUtils.shouldIgnorePath.mockImplementation(async (_basePath, filePath) => {
      const filePathStr = String(filePath);
      
      // Only ignore .log files based on our mock .gitignore
      return filePathStr.endsWith('.log') && !filePathStr.endsWith('.gitignore');
    });
    
    // Run the directory traversal
    const results = await readDirectoryContents(testDirPath);
    
    // Should include non-ignored files
    expect(results.some(r => r.path.includes('file1.txt'))).toBe(true);
    expect(results.some(r => r.path.includes('file2.md'))).toBe(true);
    
    // Should include .gitignore itself
    expect(results.some(r => r.path.endsWith('.gitignore'))).toBe(true);
    
    // Should exclude ignored files
    expect(results.some(r => r.path.includes('ignored.log'))).toBe(false);
    
    // Verify gitignore filtering was called
    expect(mockedGitignoreUtils.shouldIgnorePath).toHaveBeenCalled();
  });
  
  it('should handle nested .gitignore files with different patterns', async () => {
    // Set up mock directory structure
    mockedFs.readdir.mockImplementation(async (dirPath) => {
      const pathStr = String(dirPath);
      
      if (pathStr === testDirPath) {
        return [
          'file1.txt',
          'subdir',
          '.gitignore'
        ] as any;
      } else if (pathStr === path.join(testDirPath, 'subdir')) {
        return [
          'subfile.txt',
          'ignored.spec.js',
          '.gitignore'
        ] as any;
      }
      
      return [] as any;
    });
    
    // Set up stat mock for files and directories
    mockedFs.stat.mockImplementation(async (filePath) => {
      const filePathStr = String(filePath);
      
      if (filePathStr === testDirPath || filePathStr.endsWith('/subdir')) {
        return {
          isFile: () => false,
          isDirectory: () => true,
          size: 4096
        } as Stats;
      }
      
      return {
        isFile: () => true,
        isDirectory: () => false,
        size: 1024
      } as Stats;
    });
    
    // Set up file read mock for different .gitignore files
    mockedFs.readFile.mockImplementation(async (filePath) => {
      const filePathStr = String(filePath);
      
      if (filePathStr === path.join(testDirPath, '.gitignore')) {
        return '*.log';
      } else if (filePathStr === path.join(testDirPath, 'subdir', '.gitignore')) {
        return '*.spec.js';
      }
      
      return `Content of ${path.basename(filePathStr)}`;
    });
    
    // Mock shouldIgnorePath with different behaviors for root vs subdir
    mockedGitignoreUtils.shouldIgnorePath.mockImplementation(async (basePath, filePath) => {
      const filePathStr = String(filePath);
      const basePathStr = String(basePath);
      
      // For the subdir context
      if (basePathStr.endsWith('/subdir') || filePathStr.includes('/subdir/')) {
        // Ignore spec files in subdir
        return filePathStr.endsWith('.spec.js');
      }
      
      // For root context
      return filePathStr.endsWith('.log');
    });
    
    // Run the directory traversal
    const results = await readDirectoryContents(testDirPath);
    
    // Should include non-ignored files
    expect(results.some(r => r.path.includes('file1.txt'))).toBe(true);
    expect(results.some(r => r.path.includes('subdir/subfile.txt'))).toBe(true);
    
    // Should include both .gitignore files
    expect(results.some(r => r.path === path.join(testDirPath, '.gitignore'))).toBe(true);
    expect(results.some(r => r.path === path.join(testDirPath, 'subdir', '.gitignore'))).toBe(true);
    
    // Should exclude ignored files based on the correct .gitignore
    expect(results.some(r => r.path.includes('ignored.spec.js'))).toBe(false);
    
    // Verify gitignore filtering was called with the correct base paths
    const shouldIgnorePathCalls = mockedGitignoreUtils.shouldIgnorePath.mock.calls;
    
    // Should be called with root directory and also with subdir
    expect(shouldIgnorePathCalls.some(call => 
      String(call[0]) === testDirPath && String(call[1]).includes('file1.txt')
    )).toBe(true);
    
    expect(shouldIgnorePathCalls.some(call => 
      String(call[0]).includes('/subdir') && String(call[1]).includes('ignored.spec.js')
    )).toBe(true);
  });
  
  it('should respect negated patterns', async () => {
    // Set up mock directory structure
    mockedFs.readdir.mockImplementation(async (dirPath) => {
      const pathStr = String(dirPath);
      
      if (pathStr === testDirPath) {
        return [
          'regular.txt',
          'ignored.log',
          'important.log', // Should be kept despite pattern
          '.gitignore'
        ] as any;
      }
      
      return [] as any;
    });
    
    // Set up stat mock for files
    mockedFs.stat.mockImplementation(async (filePath) => {
      const filePathStr = String(filePath);
      
      if (filePathStr === testDirPath) {
        return {
          isFile: () => false,
          isDirectory: () => true,
          size: 4096
        } as Stats;
      }
      
      return {
        isFile: () => true,
        isDirectory: () => false,
        size: 1024
      } as Stats;
    });
    
    // Set up file read mock
    mockedFs.readFile.mockImplementation(async (filePath) => {
      const filePathStr = String(filePath);
      
      if (filePathStr.endsWith('.gitignore')) {
        // Pattern with negation
        return '*.log\n!important.log';
      }
      
      return `Content of ${path.basename(filePathStr)}`;
    });
    
    // Set up gitignore filtering mock that respects negation
    mockedGitignoreUtils.shouldIgnorePath.mockImplementation(async (_basePath, filePath) => {
      const filePathStr = String(filePath);
      const filename = path.basename(filePathStr);
      
      // Simulate ignoring pattern with negation
      if (filename === 'important.log') {
        return false; // Not ignored due to negation
      }
      
      return filePathStr.endsWith('.log') && !filePathStr.endsWith('.gitignore');
    });
    
    // Run the directory traversal
    const results = await readDirectoryContents(testDirPath);
    
    // Should include regular files and non-ignored files
    expect(results.some(r => r.path.includes('regular.txt'))).toBe(true);
    expect(results.some(r => r.path.includes('important.log'))).toBe(true); // Not ignored due to negation
    
    // Should include .gitignore itself
    expect(results.some(r => r.path.endsWith('.gitignore'))).toBe(true);
    
    // Should exclude ignored files
    expect(results.some(r => r.path.includes('ignored.log'))).toBe(false);
  });
});