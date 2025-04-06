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

describe('gitignore filtering in directory traversal', () => {
  const testDirPath = '/path/to/test/directory';
  
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Mock directory structure
    mockedFs.access.mockResolvedValue(undefined);
    
    // Mock directory stats
    mockedFs.stat.mockImplementation(async (filePath) => {
      const pathStr = String(filePath);
      
      if (
        pathStr.endsWith('.git') || 
        pathStr.endsWith('node_modules') || 
        pathStr.endsWith('subdir') || 
        pathStr === testDirPath
      ) {
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
    
    // Mock directory read
    mockedFs.readdir.mockImplementation(async (dirPath) => {
      const pathStr = String(dirPath);
      
      if (pathStr.includes('subdir')) {
        return ['nested.txt', 'nested-ignored.tmp', '.gitignore'] as any;
      }
      
      return [
        'file1.txt',
        'file2.md',
        'subdir',
        'node_modules',
        '.git',
        '.gitignore',
        'ignored-by-gitignore.log'
      ] as any;
    });
    
    // Mock successful file read
    mockedFs.readFile.mockImplementation(async (filePath) => {
      const pathStr = String(filePath);
      
      if (pathStr.includes('.gitignore')) {
        if (pathStr.includes('subdir')) {
          return '*.tmp\n';
        }
        return '*.log\n';
      }
      
      return 'Content of ' + path.basename(pathStr);
    });
    
    // Mock gitignore utils
    mockedGitignoreUtils.shouldIgnorePath.mockImplementation(async (_basePath, filePath) => {
      const filename = path.basename(filePath);
      
      // Simulate ignoring specific patterns
      if (filename.endsWith('.log')) {
        return true;
      } else if (filePath.includes('subdir') && filename.endsWith('.tmp')) {
        return true;
      }
      
      // node_modules and .git are already handled by DEFAULT_IGNORED_DIRECTORIES
      // so we don't need to handle them in the gitignore mock
      
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
  });
});