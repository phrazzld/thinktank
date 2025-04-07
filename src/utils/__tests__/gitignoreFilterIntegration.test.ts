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

// TODO: Remove mock - importing real module for next tasks
// jest.mock('../gitignoreUtils');
import * as gitignoreUtils from '../gitignoreUtils';

// Import modules after mocking
import { readDirectoryContents } from '../fileReader';

describe.skip('gitignore filtering in directory traversal', () => {
  const testDirPath = path.join('/', 'path', 'to', 'test', 'directory');
  
  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();
    resetVirtualFs();
    
    // Clear gitignore cache
    gitignoreUtils.clearIgnoreCache();
    
    // Create test directory structure using createVirtualFs
    createVirtualFs({
      [path.join(testDirPath, 'file1.txt')]: 'Content of file1.txt',
      [path.join(testDirPath, 'file2.md')]: 'Content of file2.md',
      [path.join(testDirPath, '.gitignore')]: '*.log\n',
      [path.join(testDirPath, 'ignored-by-gitignore.log')]: 'Content of ignored-by-gitignore.log',
      [path.join(testDirPath, 'subdir', 'nested.txt')]: 'Content of nested.txt',
      [path.join(testDirPath, 'subdir', 'nested-ignored.tmp')]: 'Content of nested-ignored.tmp',
      [path.join(testDirPath, 'subdir', '.gitignore')]: '*.tmp\n',
      [path.join(testDirPath, 'node_modules', '.placeholder')]: '', // To ensure directory is created
      [path.join(testDirPath, '.git', '.placeholder')]: ''  // To ensure directory is created
    });
    
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
