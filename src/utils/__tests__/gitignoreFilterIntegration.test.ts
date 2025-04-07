/**
 * Integration tests for gitignore filtering within directory traversal
 */
import path from 'path';
import { 
  resetVirtualFs, 
  getVirtualFs, 
  mockFsModules 
} from '../../__tests__/utils/virtualFsUtils';
import {
  resetMockGitignore,
  setupMockGitignore,
  mockedGitignoreUtils
} from '../../__tests__/utils/mockGitignoreUtils';

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);
jest.mock('../gitignoreUtils');

// Import modules after mocking
import { readDirectoryContents } from '../fileReader';

describe('gitignore filtering in directory traversal', () => {
  const testDirPath = '/path/to/test/directory';
  
  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();
    resetVirtualFs();
    resetMockGitignore();
    setupMockGitignore();
    
    // Get virtual filesystem reference
    const virtualFs = getVirtualFs();
    
    // Create test directory structure
    virtualFs.mkdirSync(testDirPath, { recursive: true });
    virtualFs.mkdirSync(path.join(testDirPath, 'subdir'), { recursive: true });
    virtualFs.mkdirSync(path.join(testDirPath, 'node_modules'), { recursive: true });
    virtualFs.mkdirSync(path.join(testDirPath, '.git'), { recursive: true });
    
    // Create files with content
    virtualFs.writeFileSync(path.join(testDirPath, 'file1.txt'), 'Content of file1.txt');
    virtualFs.writeFileSync(path.join(testDirPath, 'file2.md'), 'Content of file2.md');
    virtualFs.writeFileSync(path.join(testDirPath, '.gitignore'), '*.log\n');
    virtualFs.writeFileSync(path.join(testDirPath, 'ignored-by-gitignore.log'), 'Content of ignored-by-gitignore.log');
    virtualFs.writeFileSync(path.join(testDirPath, 'subdir/nested.txt'), 'Content of nested.txt');
    virtualFs.writeFileSync(path.join(testDirPath, 'subdir/nested-ignored.tmp'), 'Content of nested-ignored.tmp');
    virtualFs.writeFileSync(path.join(testDirPath, 'subdir/.gitignore'), '*.tmp\n');
    
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