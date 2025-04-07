/**
 * Integration tests for gitignore filtering within directory traversal
 */
import { mockFsModules, resetVirtualFs, createVirtualFs } from '../../__tests__/utils/virtualFsUtils';

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Now import fs and other modules
import path from 'path';
import { readDirectoryContents } from '../fileReader';
import {
  resetMockGitignore,
  setupMockGitignore,
  mockedGitignoreUtils
} from '../../__tests__/utils/mockGitignoreUtils';

// Mock dependencies
jest.mock('../gitignoreUtils');

describe('Gitignore Filtering Integration', () => {
  const testDirPath = '/path/to/test/directory';
  
  beforeEach(() => {
    // Reset and setup mocks
    resetVirtualFs();
    resetMockGitignore();
    setupMockGitignore();
  });
  
  it('should filter files based on gitignore patterns', async () => {
    // Set up virtual filesystem with test files
    createVirtualFs({
      [path.join(testDirPath, 'file1.txt')]: 'Content of file1.txt',
      [path.join(testDirPath, 'file2.md')]: 'Content of file2.md',
      [path.join(testDirPath, 'ignored.log')]: 'Content of ignored.log',
      [path.join(testDirPath, '.gitignore')]: '*.log'
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
    // Set up virtual filesystem with nested structure
    const dirStructure: Record<string, string> = {};
    
    // Add test files and directories
    dirStructure[path.join(testDirPath, 'file1.txt')] = 'Content of file1.txt';
    dirStructure[path.join(testDirPath, '.gitignore')] = '*.log';
    dirStructure[path.join(testDirPath, 'subdir', 'subfile.txt')] = 'Content of subfile.txt';
    dirStructure[path.join(testDirPath, 'subdir', 'ignored.spec.js')] = 'Content of ignored.spec.js';
    dirStructure[path.join(testDirPath, 'subdir', '.gitignore')] = '*.spec.js';
    
    // Create the virtual filesystem
    createVirtualFs(dirStructure);
    
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
    // Set up virtual filesystem
    createVirtualFs({
      [path.join(testDirPath, 'regular.txt')]: 'Content of regular.txt',
      [path.join(testDirPath, 'ignored.log')]: 'Content of ignored.log',
      [path.join(testDirPath, 'important.log')]: 'Content of important.log',
      [path.join(testDirPath, '.gitignore')]: '*.log\n!important.log'
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