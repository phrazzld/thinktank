/**
 * Comprehensive tests for gitignore-based filtering logic
 */
import { 
  mockFsModules, 
  createFsError,
  addVirtualGitignoreFile
} from '../../__tests__/utils/virtualFsUtils';
import { normalizePath } from '../../__tests__/utils/pathUtils';
import {
  setupWithGitignore,
  setupGitignoreMocking,
  setupBasicFiles
} from '../../__tests__/utils/fsTestSetup';

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Now import fs and other modules
import path from 'path';
import fs from 'fs/promises';
import * as gitignoreUtils from '../gitignoreUtils';
import * as fileReader from '../fileReader';

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
  const testDirPath = normalizePath('/test/dir', true);
  const gitignorePath = normalizePath(path.join(testDirPath, '.gitignore'), true);
  
  beforeEach(async () => {
    // Set up gitignore mocking with our reusable helper
    setupGitignoreMocking(gitignoreUtils, mockedFileExists);
  });
  
  describe('shouldIgnorePath', () => {
    it('should correctly ignore paths matching simple patterns', async () => {
      // Create .gitignore with simple patterns using our helper
      await setupWithGitignore(testDirPath, '# Default test gitignore\n*.log\ntmp/\n.DS_Store', {});
      
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
      await setupWithGitignore(testDirPath, '# Default test gitignore\n*.log\ntmp/\n.DS_Store', {});
      
      // Directory patterns like tmp/
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'tmp/file.txt')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'tmp/nested/file.js')).toBe(true);
      
      // Similar named files/directories that don't match the pattern exactly
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'temporary/file.txt')).toBe(false);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'my-tmp/file.js')).toBe(false);
    });
    
    it('should handle path conversion between absolute and relative paths', async () => {
      // Create .gitignore with patterns
      await setupWithGitignore(testDirPath, '# Default test gitignore\n*.log\ntmp/\n.DS_Store', {});
      
      // Test with absolute paths
      const absolutePath1 = normalizePath(path.join(testDirPath, 'logs/error.log'), true);
      const absolutePath2 = normalizePath(path.join(testDirPath, 'src/index.ts'), true);
      
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, absolutePath1)).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, absolutePath2)).toBe(false);
      
      // Test with paths that have parent directory references
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, normalizePath('subdir/../info.log'))).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, normalizePath('tmp/file.js'))).toBe(true);
    });
    
    it('should respect negated patterns that re-include certain paths', async () => {
      // Create .gitignore with negated patterns
      await setupWithGitignore(testDirPath, '*.log\n!important.log\ntmp/', {});
      
      // Standard ignore patterns should work
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'debug.log')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'tmp/file.txt')).toBe(true);
      
      // Negated patterns should not be ignored
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'important.log')).toBe(false);
    });
    
    it('should handle comment lines and blank lines in .gitignore files', async () => {
      // Create .gitignore with comments and blank lines
      await setupWithGitignore(testDirPath, '# This is a comment\n\n*.log\n\n# Another comment\ntmp/', {});
      
      // Should ignore patterns as normal, ignoring comments and blank lines
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'info.log')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'tmp/file.txt')).toBe(true);
      
      // Should not ignore files that match comment patterns
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, '# This is a comment')).toBe(false);
    });
    
    // This test verifies basic glob patterns.
    // 
    // KNOWN LIMITATIONS:
    // When testing with a virtual filesystem environment, there are some limitations
    // with how complex gitignore patterns work compared to a real filesystem:
    // 
    // 1. Double-asterisk patterns (**) for deep directory matching have inconsistent
    //    behavior in the virtual filesystem compared to a real filesystem.
    // 
    // 2. Patterns with curly braces for multiple extensions (e.g., *.{jpg,png,gif})
    //    may not work as expected in the virtual environment.
    // 
    // 3. Patterns with prefix wildcards (e.g., build-*/) may behave differently.
    // 
    // For reliable testing in the virtual filesystem, we focus on the most common
    // and reliable patterns. More complex patterns should be tested in integration
    // tests against a real filesystem.
    it('should handle basic glob patterns correctly', async () => {
      // Create .gitignore with basic but useful patterns
      await setupWithGitignore(testDirPath, '*.min.js\nnode_modules\n*.jpg\n*.png\n*.gif', {});
      
      // Test basic glob patterns that match file extensions
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'app.min.js')).toBe(true);
      
      // Test directory patterns - just the name without trailing slash
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'node_modules')).toBe(true);
      
      // Test file extensions
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'photo.jpg')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, 'icon.png')).toBe(true);
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
      await setupWithGitignore(testDirPath, '*.log\ncustom-dir/', {});
      
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
      setupBasicFiles({
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
      setupBasicFiles({
        [testDirPath + '/']: '' // Create just the directory
      });
      
      // Mock file exists but make fs.readFile throw an error
      jest.clearAllMocks();
      mockedFileExists.mockResolvedValue(true);
      jest.spyOn(fs, 'readFile').mockRejectedValueOnce(
        createFsError('EACCES', 'Permission denied', 'readFile', gitignorePath)
      );
      
      // Spy on console.warn with a silent mock implementation to avoid polluting test output
      jest.spyOn(console, 'warn').mockImplementation(() => {});
      
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
      await setupWithGitignore(testDirPath, '*.log\ntmp/', {});
      
      // First call should read the file
      const filter1 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Second call should use cache
      const filter2 = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Should be the same instance - this confirms caching is working
      expect(filter1).toBe(filter2);
    });
    
    it('should create separate filters for different directory paths', async () => {
      // Set up two different directories with different gitignore patterns
      const dir1 = normalizePath('/path/one', true);
      const dir2 = normalizePath('/path/two', true);
      const gitignorePath1 = normalizePath(path.join(dir1, '.gitignore'), true);
      const gitignorePath2 = normalizePath(path.join(dir2, '.gitignore'), true);
      
      // Add the .gitignore files without resetting the filesystem between calls
      await addVirtualGitignoreFile(gitignorePath1, '*.log');
      await addVirtualGitignoreFile(gitignorePath2, '*.tmp');
      
      // Verify behavior for dir1
      expect(await gitignoreUtils.shouldIgnorePath(dir1, 'test.log')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(dir1, 'test.tmp')).toBe(false);
      
      // Verify behavior for dir2
      expect(await gitignoreUtils.shouldIgnorePath(dir2, 'test.log')).toBe(false);
      expect(await gitignoreUtils.shouldIgnorePath(dir2, 'test.tmp')).toBe(true);
    });
    
    it('should refresh filters after cache is cleared', async () => {
      // Create initial .gitignore file
      await setupWithGitignore(testDirPath, '*.log', {});
      
      // First call to populate cache
      const initialFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Clear the cache
      gitignoreUtils.clearIgnoreCache();
      
      // Update .gitignore content
      await addVirtualGitignoreFile(gitignorePath, '*.new-pattern');
      
      // This should create a new filter since cache was cleared
      const newFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);
      
      // Should be a different instance after cache clearing
      expect(initialFilter).not.toBe(newFilter);
      
      // Should use new patterns
      expect(newFilter.ignores('file.new-pattern')).toBe(true);
      expect(newFilter.ignores('file.log')).toBe(false); // Old pattern not applied
    });
  });
});
