/**
 * Integration tests for gitignore filtering within directory traversal
 */
import { 
  mockFsModules, 
  resetVirtualFs, 
  createVirtualFs,
  addVirtualGitignoreFile
} from '../../__tests__/utils/virtualFsUtils';

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Now import fs and other modules
import path from 'path';
import { readDirectoryContents } from '../fileReader';
import * as gitignoreUtils from '../gitignoreUtils';

describe('Gitignore Filtering Integration', () => {
  const testDirPath = path.join('/', 'path', 'to', 'test', 'directory');
  
  beforeEach(async () => {
    // Reset virtual filesystem
    resetVirtualFs();
    
    // Clear gitignore cache
    gitignoreUtils.clearIgnoreCache();
    
    // Setup virtual filesystem with test files
    createVirtualFs({
      [path.join(testDirPath, 'file1.txt')]: 'Content of file1.txt',
      [path.join(testDirPath, 'file2.md')]: 'Content of file2.md',
      [path.join(testDirPath, 'ignored.log')]: 'Content of ignored.log'
    });
    
    // Create .gitignore file using our dedicated function
    await addVirtualGitignoreFile(path.join(testDirPath, '.gitignore'), '*.log');
  });
  
  it('should filter files based on gitignore patterns', async () => {
    // Run the directory traversal
    const results = await readDirectoryContents(testDirPath);
    
    // Should include non-ignored files
    expect(results.some(r => r.path.includes('file1.txt'))).toBe(true);
    expect(results.some(r => r.path.includes('file2.md'))).toBe(true);
    
    // Should include .gitignore itself
    expect(results.some(r => r.path.endsWith('.gitignore'))).toBe(true);
    
    // Should exclude ignored files
    expect(results.some(r => r.path.includes('ignored.log'))).toBe(false);
    
    // Verify the actual gitignoreUtils behavior directly
    expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'ignored.log')).toBe(true);
    expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'file1.txt')).toBe(false);
  });
  
  it('should handle nested .gitignore files with different patterns', async () => {
    // Set up virtual filesystem with nested structure
    resetVirtualFs();
    gitignoreUtils.clearIgnoreCache();
    
    const dirStructure: Record<string, string> = {};
    
    // Add test files and directories
    dirStructure[path.join(testDirPath, 'file1.txt')] = 'Content of file1.txt';
    dirStructure[path.join(testDirPath, 'subdir', 'subfile.txt')] = 'Content of subfile.txt';
    dirStructure[path.join(testDirPath, 'subdir', 'ignored.spec.js')] = 'Content of ignored.spec.js';
    
    // Create the virtual filesystem
    createVirtualFs(dirStructure);
    
    // Add virtual gitignore files with different patterns
    await addVirtualGitignoreFile(path.join(testDirPath, '.gitignore'), '*.log');
    await addVirtualGitignoreFile(path.join(testDirPath, 'subdir', '.gitignore'), '*.spec.js');
    
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
    
    // Verify actual gitignore behavior directly
    expect(await gitignoreUtils.shouldIgnorePath(path.join(testDirPath, 'subdir'), 'ignored.spec.js')).toBe(true);
    expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'file1.txt')).toBe(false);
  });
  
  it('should respect negated patterns', async () => {
    // Set up virtual filesystem
    resetVirtualFs();
    gitignoreUtils.clearIgnoreCache();
    
    createVirtualFs({
      [path.join(testDirPath, 'regular.txt')]: 'Content of regular.txt',
      [path.join(testDirPath, 'ignored.log')]: 'Content of ignored.log',
      [path.join(testDirPath, 'important.log')]: 'Content of important.log'
    });
    
    // Add virtual gitignore file with negated pattern
    await addVirtualGitignoreFile(path.join(testDirPath, '.gitignore'), '*.log\n!important.log');
    
    // Run the directory traversal
    const results = await readDirectoryContents(testDirPath);
    
    // Should include regular files and non-ignored files
    expect(results.some(r => r.path.includes('regular.txt'))).toBe(true);
    expect(results.some(r => r.path.includes('important.log'))).toBe(true); // Not ignored due to negation
    
    // Should include .gitignore itself
    expect(results.some(r => r.path.endsWith('.gitignore'))).toBe(true);
    
    // Should exclude ignored files
    expect(results.some(r => r.path.includes('ignored.log'))).toBe(false);
    
    // Verify actual gitignore behavior directly
    expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'ignored.log')).toBe(true);
    expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'important.log')).toBe(false);
    expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'regular.txt')).toBe(false);
  });
});
