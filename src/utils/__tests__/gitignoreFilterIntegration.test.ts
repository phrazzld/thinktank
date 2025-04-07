/**
 * Integration tests for gitignore filtering within directory traversal
 */
import path from 'path';
import { 
  resetVirtualFs, 
  createVirtualFs,
  mockFsModules,
  addVirtualGitignoreFile
} from '../../__tests__/utils/virtualFsUtils';

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

import * as gitignoreUtils from '../gitignoreUtils';

// Import modules after mocking
import { readDirectoryContents } from '../fileReader';

describe('gitignore filtering in directory traversal', () => {
  const testDirPath = path.join('/', 'path', 'to', 'test', 'directory');
  
  beforeEach(async () => {
    // Reset mocks
    jest.clearAllMocks();
    resetVirtualFs();
    
    // Clear gitignore cache
    gitignoreUtils.clearIgnoreCache();
    
    // Create test directory structure using createVirtualFs, but without .gitignore files
    createVirtualFs({
      [path.join(testDirPath, 'file1.txt')]: 'Content of file1.txt',
      [path.join(testDirPath, 'file2.md')]: 'Content of file2.md',
      [path.join(testDirPath, 'ignored-by-gitignore.log')]: 'Content of ignored-by-gitignore.log',
      [path.join(testDirPath, 'subdir', 'nested.txt')]: 'Content of nested.txt',
      [path.join(testDirPath, 'subdir', 'nested-ignored.tmp')]: 'Content of nested-ignored.tmp',
      [path.join(testDirPath, 'node_modules', '.placeholder')]: '', // To ensure directory is created
      [path.join(testDirPath, '.git', '.placeholder')]: ''  // To ensure directory is created
    });
    
    // Add .gitignore files using our specialized function
    await addVirtualGitignoreFile(path.join(testDirPath, '.gitignore'), '*.log\n');
    await addVirtualGitignoreFile(path.join(testDirPath, 'subdir', '.gitignore'), '*.tmp\n');
  });
  
  it('should filter files based on gitignore rules during directory traversal', async () => {
    const results = await readDirectoryContents(testDirPath);
    
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
    
    // Verify that the actual gitignoreUtils behavior works correctly
    expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'ignored-by-gitignore.log')).toBe(true);
    expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'file1.txt')).toBe(false);
    expect(await gitignoreUtils.shouldIgnorePath(path.join(testDirPath, 'subdir'), 'nested-ignored.tmp')).toBe(true);
  });
  
  it('should check gitignore rules in the correct directories', async () => {
    // Setup a special file structure to test directory-specific rules
    resetVirtualFs();
    gitignoreUtils.clearIgnoreCache();
    
    // Create test directory with a more complex structure to test directory-specific rules
    createVirtualFs({
      [path.join(testDirPath, 'root.txt')]: 'Content of root.txt',
      [path.join(testDirPath, 'test.md')]: 'Content of test.md',
      [path.join(testDirPath, 'test.log')]: 'Ignored by root gitignore',
      [path.join(testDirPath, 'dir1', 'file.txt')]: 'Content of dir1/file.txt',
      [path.join(testDirPath, 'dir1', 'file.md')]: 'Content of dir1/file.md', // Ignored by dir1/.gitignore
      [path.join(testDirPath, 'dir2', 'file.txt')]: 'Content of dir2/file.txt',
      [path.join(testDirPath, 'dir2', 'test.log')]: 'Ignored by root gitignore'
    });
    
    // Add different .gitignore files in different directories
    await addVirtualGitignoreFile(path.join(testDirPath, '.gitignore'), '*.log');
    await addVirtualGitignoreFile(path.join(testDirPath, 'dir1', '.gitignore'), '*.md');
    
    // Run the directory traversal
    const results = await readDirectoryContents(testDirPath);
    
    // Check rules are applied correctly
    // Root .gitignore should affect all subdirectories
    expect(results.some(r => r.path.includes('test.log'))).toBe(false);
    expect(results.some(r => r.path.includes('dir2/test.log'))).toBe(false);
    
    // dir1/.gitignore should only affect dir1
    expect(results.some(r => r.path.includes('test.md'))).toBe(true); // Not affected by dir1/.gitignore
    expect(results.some(r => r.path.includes('dir1/file.md'))).toBe(false); // Affected by dir1/.gitignore
    
    // Verify the actual gitignore implementation directly
    expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'test.log')).toBe(true);
    expect(await gitignoreUtils.shouldIgnorePath(path.join(testDirPath, 'dir2'), 'test.log')).toBe(true);
    expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'test.md')).toBe(false);
    expect(await gitignoreUtils.shouldIgnorePath(path.join(testDirPath, 'dir1'), 'file.md')).toBe(true);
  });
});
