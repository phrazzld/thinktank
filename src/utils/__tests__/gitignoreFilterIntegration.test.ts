/**
 * Integration tests for gitignore filtering within directory traversal
 */
import path from 'path';
import { readDirectoryContents } from '../fileReader';
import { 
  resetMockFs, 
  setupMockFs, 
  mockStat, 
  mockReaddir,
  mockReadFile,
  mockAccess
} from '../../__tests__/utils/mockFsUtils';
import {
  resetMockGitignore,
  setupMockGitignore,
  mockedGitignoreUtils
} from '../../__tests__/utils/mockGitignoreUtils';

// Mock dependencies
jest.mock('fs/promises');
jest.mock('../gitignoreUtils');

describe('gitignore filtering in directory traversal', () => {
  const testDirPath = '/path/to/test/directory';
  
  beforeEach(() => {
    // Reset and setup mocks
    resetMockFs();
    setupMockFs();
    resetMockGitignore();
    setupMockGitignore();
    
    // Mock directory structure and access
    mockAccess(testDirPath, true);
    mockAccess(path.join(testDirPath, 'file1.txt'), true);
    mockAccess(path.join(testDirPath, 'file2.md'), true);
    mockAccess(path.join(testDirPath, 'subdir'), true);
    mockAccess(path.join(testDirPath, 'node_modules'), true);
    mockAccess(path.join(testDirPath, '.git'), true);
    mockAccess(path.join(testDirPath, '.gitignore'), true);
    mockAccess(path.join(testDirPath, 'ignored-by-gitignore.log'), true);
    mockAccess(path.join(testDirPath, 'subdir/nested.txt'), true);
    mockAccess(path.join(testDirPath, 'subdir/nested-ignored.tmp'), true);
    mockAccess(path.join(testDirPath, 'subdir/.gitignore'), true);
    
    // Mock directory stats
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
    
    // Set up stats for directories
    mockStat(testDirPath, dirStats);
    mockStat(path.join(testDirPath, '.git'), dirStats);
    mockStat(path.join(testDirPath, 'node_modules'), dirStats);
    mockStat(path.join(testDirPath, 'subdir'), dirStats);
    
    // Set up stats for files
    mockStat(path.join(testDirPath, 'file1.txt'), fileStats);
    mockStat(path.join(testDirPath, 'file2.md'), fileStats);
    mockStat(path.join(testDirPath, '.gitignore'), fileStats);
    mockStat(path.join(testDirPath, 'ignored-by-gitignore.log'), fileStats);
    mockStat(path.join(testDirPath, 'subdir/nested.txt'), fileStats);
    mockStat(path.join(testDirPath, 'subdir/nested-ignored.tmp'), fileStats);
    mockStat(path.join(testDirPath, 'subdir/.gitignore'), fileStats);
    
    // Mock directory read for root and subdir
    mockReaddir(testDirPath, [
      'file1.txt',
      'file2.md',
      'subdir',
      'node_modules',
      '.git',
      '.gitignore',
      'ignored-by-gitignore.log'
    ]);
    
    mockReaddir(path.join(testDirPath, 'subdir'), [
      'nested.txt',
      'nested-ignored.tmp',
      '.gitignore'
    ]);
    
    // Mock file content
    mockReadFile(path.join(testDirPath, '.gitignore'), '*.log\n');
    mockReadFile(path.join(testDirPath, 'subdir/.gitignore'), '*.tmp\n');
    mockReadFile(path.join(testDirPath, 'file1.txt'), 'Content of file1.txt');
    mockReadFile(path.join(testDirPath, 'file2.md'), 'Content of file2.md');
    mockReadFile(path.join(testDirPath, 'ignored-by-gitignore.log'), 'Content of ignored-by-gitignore.log');
    mockReadFile(path.join(testDirPath, 'subdir/nested.txt'), 'Content of nested.txt');
    mockReadFile(path.join(testDirPath, 'subdir/nested-ignored.tmp'), 'Content of nested-ignored.tmp');
    
    // Mock shouldIgnorePath to simulate gitignore filtering behavior
    // The function is called with (dirPath, entryPath) so we need to be specific about
    // which files should be ignored based on their full paths
    mockedGitignoreUtils.shouldIgnorePath.mockImplementation(async (_basePath, filePath) => {
      // Ignore *.log in the root directory
      if (filePath.endsWith('ignored-by-gitignore.log')) {
        return true;
      }
      
      // Ignore *.tmp in the subdir directory
      if (filePath.endsWith('nested-ignored.tmp')) {
        return true;
      }
      
      // Don't ignore other files
      return false;
    });
  });
  
  it('should filter files based on gitignore rules during directory traversal', async () => {
    const results = await readDirectoryContents(testDirPath);
    
    // For debugging if needed: results.map(r => r.path)
    
    // Should include non-ignored files
    expect(results.some(r => r.path.includes('file1.txt'))).toBe(true);
    expect(results.some(r => r.path.includes('file2.md'))).toBe(true);
    expect(results.some(r => r.path.includes('nested.txt'))).toBe(true);
    
    // Should not include ignored files
    expect(results.some(r => r.path.includes('ignored-by-gitignore.log'))).toBe(false);
    expect(results.some(r => r.path.includes('nested-ignored.tmp'))).toBe(false);
    expect(results.some(r => r.path.includes('node_modules/'))).toBe(false);
    expect(results.some(r => r.path.includes('/.git/'))).toBe(false);
    
    // .gitignore files themselves are included (they're not ignored by default)
    expect(results.some(r => r.path.endsWith('.gitignore'))).toBe(true);
    
    // Make sure gitignore utils are being called
    expect(mockedGitignoreUtils.shouldIgnorePath).toHaveBeenCalled();
  });
  
  it('should check gitignore rules in the correct directories', async () => {
    await readDirectoryContents(testDirPath);
    
    // Verify that shouldIgnorePath is called for various paths
    expect(mockedGitignoreUtils.shouldIgnorePath).toHaveBeenCalled();
    
    // Check that it's called with the root path and the ignored file
    expect(mockedGitignoreUtils.shouldIgnorePath).toHaveBeenCalledWith(
      expect.anything(),
      expect.stringMatching(/ignored-by-gitignore\.log$/)
    );
    
    // Check that it's called with the subdir path and the ignored tmp file
    expect(mockedGitignoreUtils.shouldIgnorePath).toHaveBeenCalledWith(
      expect.anything(),
      expect.stringMatching(/nested-ignored\.tmp$/)
    );
  });
});