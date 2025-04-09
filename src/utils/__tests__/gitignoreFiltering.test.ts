/**
 * Comprehensive tests for gitignore-based filtering logic
 */
import {
  mockFsModules,
  createFsError,
  addVirtualGitignoreFile,
} from '../../__tests__/utils/virtualFsUtils';
import { normalizePath } from '../../__tests__/utils/pathUtils';
import {
  setupWithGitignore,
  setupGitignoreMocking,
  setupBasicFiles,
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
    fileExists: jest.fn(),
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
      // Make all paths absolute to avoid relative path issues
      const infoLogPath = path.join(testDirPath, 'info.log');
      const logsErrorPath = path.join(testDirPath, 'logs/error.log');
      const debugLogPath = path.join(testDirPath, 'debug.log');
      const dsStorePath = path.join(testDirPath, '.DS_Store');

      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, infoLogPath)).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, logsErrorPath)).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, debugLogPath)).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, dsStorePath)).toBe(true);

      // Non-matching files should not be ignored
      const infoTxtPath = path.join(testDirPath, 'info.txt');
      const fileMdPath = path.join(testDirPath, 'file.md');
      const logPath = path.join(testDirPath, 'log'); // Not *.log

      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, infoTxtPath)).toBe(false);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, fileMdPath)).toBe(false);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, logPath)).toBe(false);
    });

    it('should handle directory patterns correctly', async () => {
      // Create .gitignore with directory patterns
      await setupWithGitignore(testDirPath, '# Default test gitignore\n*.log\ntmp/\n.DS_Store', {});

      // Directory patterns like tmp/
      const tmpFilePath = path.join(testDirPath, 'tmp/file.txt');
      const tmpNestedFilePath = path.join(testDirPath, 'tmp/nested/file.js');

      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, tmpFilePath)).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, tmpNestedFilePath)).toBe(true);

      // Similar named files/directories that don't match the pattern exactly
      const tempFilePath = path.join(testDirPath, 'temporary/file.txt');
      const myTmpFilePath = path.join(testDirPath, 'my-tmp/file.js');

      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, tempFilePath)).toBe(false);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, myTmpFilePath)).toBe(false);
    });

    it('should handle path conversion between absolute and relative paths', async () => {
      // Create .gitignore with patterns
      await setupWithGitignore(testDirPath, '# Default test gitignore\n*.log\ntmp/\n.DS_Store', {});

      // Test with absolute paths
      const absolutePath1 = path.join(testDirPath, 'logs/error.log');
      const absolutePath2 = path.join(testDirPath, 'src/index.ts');

      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, absolutePath1)).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, absolutePath2)).toBe(false);

      // Test with paths that have parent directory references
      const subdirParentPath = path.join(testDirPath, 'subdir/../info.log');
      const tmpFilePath = path.join(testDirPath, 'tmp/file.js');

      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, subdirParentPath)).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, tmpFilePath)).toBe(true);
    });

    it('should respect negated patterns that re-include certain paths', async () => {
      // Create .gitignore with negated patterns
      await setupWithGitignore(testDirPath, '*.log\n!important.log\ntmp/', {});

      // Standard ignore patterns should work
      const debugLogPath = path.join(testDirPath, 'debug.log');
      const tmpFilePath = path.join(testDirPath, 'tmp/file.txt');

      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, debugLogPath)).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, tmpFilePath)).toBe(true);

      // Negated patterns should not be ignored
      const importantLogPath = path.join(testDirPath, 'important.log');
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, importantLogPath)).toBe(false);
    });

    it('should handle comment lines and blank lines in .gitignore files', async () => {
      // Create .gitignore with comments and blank lines
      await setupWithGitignore(
        testDirPath,
        '# This is a comment\n\n*.log\n\n# Another comment\ntmp/',
        {}
      );

      // Should ignore patterns as normal, ignoring comments and blank lines
      const infoLogPath = path.join(testDirPath, 'info.log');
      const tmpFilePath = path.join(testDirPath, 'tmp/file.txt');

      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, infoLogPath)).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, tmpFilePath)).toBe(true);

      // Should not ignore files that match comment patterns
      const commentFilePath = path.join(testDirPath, '# This is a comment');
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, commentFilePath)).toBe(false);
    });

    // This test verifies basic glob patterns.
    //
    // NOTE: Complex gitignore patterns are now tested in a dedicated test file:
    // gitignoreComplexPatterns.test.ts, which includes tests for:
    //
    // 1. Double-asterisk patterns (**) for deep directory matching
    // 2. Patterns with curly braces for multiple extensions (e.g., *.{jpg,png,gif})
    // 3. Patterns with prefix wildcards (e.g., build-*/)
    // 4. Negated nested patterns (*.log + !important/*.log)
    // 5. Character range patterns ([0-9]*.js)
    //
    // The complex patterns test suite includes both integration tests with the virtual
    // filesystem and direct tests of the ignore library's pattern matching functionality
    // to ensure comprehensive coverage.
    it('should handle basic glob patterns correctly', async () => {
      // Create .gitignore with basic but useful patterns
      await setupWithGitignore(testDirPath, '*.min.js\nnode_modules\n*.jpg\n*.png\n*.gif', {});

      // Test basic glob patterns that match file extensions
      const minJsPath = path.join(testDirPath, 'app.min.js');
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, minJsPath)).toBe(true);

      // Test directory patterns - just the name without trailing slash
      const nodeModulesPath = path.join(testDirPath, 'node_modules');
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, nodeModulesPath)).toBe(true);

      // Test file extensions
      const jpgPath = path.join(testDirPath, 'photo.jpg');
      const pngPath = path.join(testDirPath, 'icon.png');
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, jpgPath)).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, pngPath)).toBe(true);
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
        '.idea',
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
        [testDirPath + '/']: '', // Create just the directory without a .gitignore file
      });

      // Reset the cache so we get a fresh filter
      gitignoreUtils.clearIgnoreCache();

      // Mock fileExists to return false to simulate non-existent .gitignore
      jest.clearAllMocks();
      mockedFileExists.mockResolvedValue(false);

      // Call the function under test
      const ignoreFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);

      // Verify the mock was called
      expect(mockedFileExists).toHaveBeenCalled();

      // Default patterns should still be respected
      expect(ignoreFilter.ignores('node_modules/package.json')).toBe(true);
      expect(ignoreFilter.ignores('.git/config')).toBe(true);

      // Custom patterns that would be in a .gitignore should not be applied
      expect(ignoreFilter.ignores('info.log')).toBe(false);
    });

    it('should handle errors when reading .gitignore files gracefully', async () => {
      // Create directory without a .gitignore file
      setupBasicFiles({
        [testDirPath + '/']: '', // Create just the directory
      });

      // Clear the cache for a fresh test
      gitignoreUtils.clearIgnoreCache();

      // Mock file exists but make fs.readFile throw an error
      jest.clearAllMocks();
      mockedFileExists.mockResolvedValue(true);

      // Create a fs.readFile mock that throws an error
      const readFileMock = jest.spyOn(fs, 'readFile');
      readFileMock.mockRejectedValue(
        createFsError('EACCES', 'Permission denied', 'readFile', gitignorePath)
      );

      // Spy on console.warn with a silent mock implementation to avoid polluting test output
      const consoleWarnMock = jest.spyOn(console, 'warn').mockImplementation(() => {});

      // Call the function being tested
      const ignoreFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);

      // Should still return a filter with default patterns
      expect(ignoreFilter.ignores('node_modules/package.json')).toBe(true);

      // Should log a warning - wait for async operations to complete
      await new Promise(process.nextTick);

      // Now check if console.warn was called
      expect(consoleWarnMock).toHaveBeenCalled();

      // Restore mocks for next tests
      consoleWarnMock.mockRestore();
      readFileMock.mockRestore();
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
      const testLogPath1 = path.join(dir1, 'test.log');
      const testTmpPath1 = path.join(dir1, 'test.tmp');
      expect(await gitignoreUtils.shouldIgnorePath(dir1, testLogPath1)).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(dir1, testTmpPath1)).toBe(false);

      // Verify behavior for dir2
      const testLogPath2 = path.join(dir2, 'test.log');
      const testTmpPath2 = path.join(dir2, 'test.tmp');
      expect(await gitignoreUtils.shouldIgnorePath(dir2, testLogPath2)).toBe(false);
      expect(await gitignoreUtils.shouldIgnorePath(dir2, testTmpPath2)).toBe(true);
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
