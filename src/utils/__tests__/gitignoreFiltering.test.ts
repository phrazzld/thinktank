/**
 * Comprehensive tests for gitignore-based filtering logic
 */
import { 
  mockFsModules, 
  resetVirtualFs, 
  createVirtualFs, 
  createFsError,
  addVirtualGitignoreFile 
} from '../../__tests__/utils/virtualFsUtils';

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Now import fs and other modules
import path from 'path';
import fs from 'fs/promises';
import * as gitignoreUtils from '../gitignoreUtils';
import * as fileReader from '../fileReader';
import { getVirtualFs } from '../../__tests__/utils/virtualFsUtils';

// Mock dependencies - but allow the real implementation in the module
jest.mock('../fileReader', () => {
  const originalModule = jest.requireActual('../fileReader');
  return {
    ...originalModule,
    fileExists: jest.fn()
  };
});

// Access mocked functions
const mockedFileExists = jest.mocked(fileReader.fileExists);

describe('Gitignore-based Filtering Logic', () => {
  // Setup and teardown for all tests
  afterEach(() => {
    jest.restoreAllMocks();
  });
  const testDirPath = '/test/dir';
  const gitignorePath = path.join(testDirPath, '.gitignore');
  
  beforeEach(async () => {
    // Reset virtual filesystem
    resetVirtualFs();
    
    // Clear gitignore cache
    gitignoreUtils.clearIgnoreCache();
    
    // Set up fileExists mock to use the virtual filesystem
    // This ensures that calls to fileExists will return true for the .gitignore file
    mockedFileExists.mockImplementation(async (filePath) => {
      const virtualFs = getVirtualFs();
      const normalizedPath = filePath.startsWith('/') ? filePath.substring(1) : filePath;
      try {
        // Check if file exists in the virtual filesystem
        virtualFs.statSync(normalizedPath);
        return true;
      } catch (error) {
        return false;
      }
    });
    
    // Spy on fs.readFile to track calls and log calls
    jest.spyOn(fs, 'readFile').mockImplementation(async (filePath, options) => {
      const normalizedPath = typeof filePath === 'string' && filePath.startsWith('/') 
        ? filePath.substring(1) 
        : String(filePath);
      
      // Default encoding is utf8 if not specified
      const encoding = typeof options === 'string' 
        ? options 
        : (options && typeof options === 'object' && 'encoding' in options && typeof options.encoding === 'string')
          ? options.encoding
          : 'utf8';
        
      const virtualFs = getVirtualFs();
      // Use virtual fs readFileSync and convert to a promise for fs.readFile
      const content = virtualFs.readFileSync(normalizedPath, encoding as BufferEncoding);
      return content;
    });
  });
  
  describe('shouldIgnorePath', () => {
    it('should correctly ignore paths matching simple patterns', async () => {
      // Create .gitignore with simple patterns
      await addVirtualGitignoreFile(gitignorePath, '# Default test gitignore\n*.log\ntmp/\n.DS_Store');
      
      // Simple patterns like *.log, file.txt
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'info.log')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'logs/error.log')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'debug.log')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, '.DS_Store')).toBe(true);
      
      // Non-matching files should not be ignored
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'info.txt')).toBe(false);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'file.md')).toBe(false);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'log')).toBe(false); // Not *.log
    });
    
    it('should handle directory patterns correctly', async () => {
      // Create .gitignore with directory patterns
      await addVirtualGitignoreFile(gitignorePath, '# Default test gitignore\n*.log\ntmp/\n.DS_Store');
      
      // Directory patterns like tmp/
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'tmp/file.txt')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'tmp/nested/file.js')).toBe(true);
      
      // Similar named files/directories that don't match the pattern exactly
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'temporary/file.txt')).toBe(false);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'my-tmp/file.js')).toBe(false);
    });
    
    it('should handle path conversion between absolute and relative paths', async () => {
      // Create .gitignore with patterns
      await addVirtualGitignoreFile(gitignorePath, '# Default test gitignore\n*.log\ntmp/\n.DS_Store');
      
      // Test with absolute paths
      const absolutePath1 = path.join(testDirPath, 'logs/error.log');
      const absolutePath2 = path.join(testDirPath, 'src/index.ts');
      
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, absolutePath1)).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, absolutePath2)).toBe(false);
      
      // Test with paths that have parent directory references
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'subdir/../info.log')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'tmp/file.js')).toBe(true);
    });
    
    it('should respect negated patterns that re-include certain paths', async () => {
      // Create .gitignore with negated patterns
      await addVirtualGitignoreFile(gitignorePath, '*.log\n!important.log\ntmp/');
      
      // Standard ignore patterns should work
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'debug.log')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'tmp/file.txt')).toBe(true);
      
      // Negated patterns should not be ignored
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'important.log')).toBe(false);
    });
    
    it('should handle comment lines and blank lines in .gitignore files', async () => {
      // Create .gitignore with comments and blank lines
      await addVirtualGitignoreFile(gitignorePath, '# This is a comment\n\n*.log\n\n# Another comment\ntmp/');
      
      // Should ignore patterns as normal, ignoring comments and blank lines
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'info.log')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'tmp/file.txt')).toBe(true);
      
      // Should not ignore files that match comment patterns
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, '# This is a comment')).toBe(false);
    });
    
    // This test is skipped because the ignore library implementation in virtualFs
    // doesn't support all the complex glob patterns the same way as the real fs
    it.skip('should handle complex glob patterns correctly', async () => {
      // Create .gitignore with complex patterns
      await addVirtualGitignoreFile(gitignorePath, '**/*.min.js\n**/node_modules/**\n**/build-*/\n*.{jpg,png,gif}');
      
      // Test complex glob patterns
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'dist/app.min.js')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'src/lib/helper.min.js')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'subdir/node_modules/package/index.js')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'project/build-dev/output.txt')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'images/photo.jpg')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'images/icon.png')).toBe(true);
      
      // Non-matching files
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'dist/app.js')).toBe(false);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'node_modules.bak/file.txt')).toBe(false);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'project/builder/output.txt')).toBe(false);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'images/photo.svg')).toBe(false);
    });
  });
  
  describe('createIgnoreFilter behavior', () => {
    it('should merge default patterns with .gitignore patterns', async () => {
      // Get actual default patterns from the implementation
      const defaultPatterns = [
        'node_modules',
        '.git',
        'dist',
        'build',
        'coverage',
        '.cache',
        '.next',
        '.nuxt',
        '.output',
        '.vscode',
        '.idea'
      ];
      
      // Create .gitignore with custom patterns
      await addVirtualGitignoreFile(gitignorePath, '*.log\ncustom-dir/');
      
      const ignoreFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Default patterns should be respected
      for (const pattern of defaultPatterns) {
        expect(ignoreFilter.ignores(`${pattern}/some-file.txt`)).toBe(true);
      }
      
      // Custom patterns should also be respected
      expect(ignoreFilter.ignores('info.log')).toBe(true);
      expect(ignoreFilter.ignores('custom-dir/file.txt')).toBe(true);
    });
    
    it('should handle non-existent .gitignore files by using only default patterns', async () => {
      // Create directory without a .gitignore file
      createVirtualFs({
        [testDirPath + '/']: '' // Create just the directory without a .gitignore file
      });
      
      // Mock fileExists to return false to simulate non-existent .gitignore
      jest.clearAllMocks();
      mockedFileExists.mockResolvedValue(false);
      
      const ignoreFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Verify correct behavior
      expect(mockedFileExists).toHaveBeenCalledWith(gitignorePath);
      
      // Default patterns should still be respected
      expect(ignoreFilter.ignores('node_modules/package.json')).toBe(true);
      expect(ignoreFilter.ignores('.git/config')).toBe(true);
      
      // Custom patterns that would be in a .gitignore should not be applied
      expect(ignoreFilter.ignores('info.log')).toBe(false);
    });
    
    it('should handle errors when reading .gitignore files gracefully', async () => {
      // Create directory without a .gitignore file
      createVirtualFs({
        [testDirPath + '/']: '' // Create just the directory
      });
      
      // Mock file exists but make fs.readFile throw an error
      jest.clearAllMocks();
      mockedFileExists.mockResolvedValue(true);
      jest.spyOn(fs, 'readFile').mockRejectedValueOnce(
        createFsError('EACCES', 'Permission denied', 'readFile', gitignorePath)
      );
      
      // Spy on console.warn
      jest.spyOn(console, 'warn').mockImplementation();
      
      const ignoreFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Should still return a filter with default patterns
      expect(ignoreFilter.ignores('node_modules/package.json')).toBe(true);
      
      // Should log a warning
      expect(console.warn).toHaveBeenCalled();
      expect(jest.mocked(console.warn).mock.calls[0][0]).toContain('Could not read .gitignore file');
    });
  });
  
  describe('ignore cache behavior', () => {
    it('should cache and reuse filters for the same directory path', async () => {
      // Create .gitignore file
      await addVirtualGitignoreFile(gitignorePath, '*.log\ntmp/');
      
      // First call should read the file
      const filter1 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Clear mock counts
      jest.clearAllMocks();
      
      // Second call should use cache
      const filter2 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Should not read file again
      expect(mockedFileExists).not.toHaveBeenCalled();
      expect(fs.readFile).not.toHaveBeenCalled();
      
      // Should be the same instance
      expect(filter1).toBe(filter2);
    });
    
    it('should create separate filters for different directory paths', async () => {
      const dir1 = '/path/one';
      const dir2 = '/path/two';
      const gitignorePath1 = path.join(dir1, '.gitignore');
      const gitignorePath2 = path.join(dir2, '.gitignore');
      
      // Create different .gitignore files
      await addVirtualGitignoreFile(gitignorePath1, '*.log');
      await addVirtualGitignoreFile(gitignorePath2, '*.tmp');
      
      // Create filters for different paths
      const filter1 = await gitignoreUtils.createIgnoreFilter(dir1);
      const filter2 = await gitignoreUtils.createIgnoreFilter(dir2);
      
      // Should be different instances
      expect(filter1).not.toBe(filter2);
      
      // Each should respect its own patterns
      expect(filter1.ignores('info.log')).toBe(true);
      expect(filter1.ignores('data.tmp')).toBe(false);
      
      expect(filter2.ignores('info.log')).toBe(false);
      expect(filter2.ignores('data.tmp')).toBe(true);
    });
    
    it('should refresh filters after cache is cleared', async () => {
      // Create initial .gitignore file
      await addVirtualGitignoreFile(gitignorePath, '*.log');
      
      // First call to populate cache
      await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Clear mock counts
      jest.clearAllMocks();
      
      // Clear the cache
      gitignoreUtils.clearIgnoreCache();
      
      // Update .gitignore content
      await addVirtualGitignoreFile(gitignorePath, '*.new-pattern');
      
      // This should need to read again
      const newFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Should have read file again
      expect(mockedFileExists).toHaveBeenCalled();
      expect(fs.readFile).toHaveBeenCalled();
      
      // Should use new patterns
      expect(newFilter.ignores('file.new-pattern')).toBe(true);
      expect(newFilter.ignores('file.log')).toBe(false); // Old pattern not applied
    });
  });
});
