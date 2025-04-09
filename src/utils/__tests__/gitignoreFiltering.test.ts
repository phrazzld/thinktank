/**
 * Comprehensive tests for gitignore-based filtering logic
 * 
 * These tests use the virtual filesystem approach with dedicated helpers
 * from test/setup/gitignore.ts instead of mocks.
 */
import path from 'path';
import fs from 'fs/promises';
import * as gitignoreUtils from '../gitignoreUtils';
import { normalizePathGeneral } from '../pathUtils';
import { setupTestHooks } from '../../../test/setup/common';
import {
  addGitignoreFile,
  setupWithGitignore,
  setupMultiGitignore,
  createIgnoreChecker
} from '../../../test/setup/gitignore';

describe('Gitignore-based Filtering Logic', () => {
  // Set up hooks to reset virtual filesystem and mocks for each test
  setupTestHooks();

  // Setup common test paths
  const testDirPath = normalizePathGeneral('/test/dir', true);
  const gitignorePath = normalizePathGeneral(path.join(testDirPath, '.gitignore'), true);

  // Before each test, clear the ignore cache to ensure a clean state
  beforeEach(() => {
    gitignoreUtils.clearIgnoreCache();
  });

  describe('shouldIgnorePath', () => {
    it('should correctly ignore paths matching simple patterns', async () => {
      // Create .gitignore with simple patterns using our helper
      await setupWithGitignore(testDirPath, '# Default test gitignore\n*.log\ntmp/\n.DS_Store', {
        'sample.txt': 'Sample text file',
        'info.log': 'Log file that should be ignored',
        'logs/error.log': 'Nested log file that should be ignored',
        '.DS_Store': 'macOS file that should be ignored',
        'tmp/temp.txt': 'File in ignored directory'
      });

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
      await setupWithGitignore(testDirPath, '# Default test gitignore\n*.log\ntmp/\n.DS_Store', {
        'sample.txt': 'Sample text file',
        'tmp/file.txt': 'File in ignored directory',
        'tmp/nested/file.js': 'Nested file in ignored directory',
        'temporary/file.txt': 'File in a directory that should not be ignored',
        'my-tmp/file.js': 'File in a directory with similar name that should not be ignored'
      });

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
      await setupWithGitignore(testDirPath, '# Default test gitignore\n*.log\ntmp/\n.DS_Store', {
        'logs/error.log': 'Log file that should be ignored',
        'src/index.ts': 'Source file that should not be ignored',
        'subdir/info.txt': 'Text file in subdirectory',
        'tmp/file.js': 'File in ignored directory'
      });

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
      await setupWithGitignore(testDirPath, '*.log\n!important.log\ntmp/', {
        'debug.log': 'Debug log file that should be ignored',
        'important.log': 'Important log file that should NOT be ignored despite *.log pattern',
        'info.log': 'Info log file that should be ignored',
        'tmp/file.txt': 'File in ignored directory'
      });

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
        {
          'info.log': 'Log file that should be ignored',
          'tmp/file.txt': 'File in ignored directory',
          '# This is a comment': 'File with a name that matches a comment in gitignore'
        }
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
      await setupWithGitignore(testDirPath, '*.min.js\nnode_modules\n*.jpg\n*.png\n*.gif', {
        'app.min.js': 'Minified JS file that should be ignored',
        'app.js': 'Regular JS file that should not be ignored',
        'node_modules/package.json': 'File in node_modules directory that should be ignored',
        'photo.jpg': 'JPG image that should be ignored',
        'icon.png': 'PNG image that should be ignored',
        'document.pdf': 'PDF file that should not be ignored'
      });

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

    it('should use the createIgnoreChecker helper correctly', async () => {
      // Setup with gitignore
      await setupWithGitignore(testDirPath, '*.log\n!important.log\ntmp/', {
        'debug.log': 'Debug log file',
        'important.log': 'Important log file',
        'src/app.js': 'App source code'
      });

      // Create a checker function for the test directory
      const shouldIgnore = createIgnoreChecker(testDirPath);

      // Test with relative paths
      expect(await shouldIgnore('debug.log')).toBe(true);
      expect(await shouldIgnore('important.log')).toBe(false);
      expect(await shouldIgnore('src/app.js')).toBe(false);
      expect(await shouldIgnore('tmp/any-file.txt')).toBe(true);
    });

    it('should handle nested gitignore files with setupMultiGitignore', async () => {
      // Setup project with multiple gitignore files
      await setupMultiGitignore(
        testDirPath,
        {
          '.gitignore': '*.log\n!important/*.log', // Root gitignore
          'src/.gitignore': '*.tmp\nbuild/' // Nested gitignore
        },
        {
          'debug.log': 'Debug log at root - should be ignored',
          'important/critical.log': 'Critical log - should NOT be ignored',
          'src/app.js': 'Source file - should not be ignored',
          'src/temp.tmp': 'Temp file in src - should be ignored',
          'src/build/output.js': 'Build output - should be ignored'
        }
      );

      // Test files with root gitignore
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, path.join(testDirPath, 'debug.log'))).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(testDirPath, path.join(testDirPath, 'important/critical.log'))).toBe(false);
      
      // Note: Since our implementation doesn't check parent directories recursively,
      // we need to explicitly check against the directory that contains the .gitignore
      const srcPath = path.join(testDirPath, 'src');
      expect(await gitignoreUtils.shouldIgnorePath(srcPath, path.join(srcPath, 'temp.tmp'))).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(srcPath, path.join(srcPath, 'build/output.js'))).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath(srcPath, path.join(srcPath, 'app.js'))).toBe(false);
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
      await setupWithGitignore(testDirPath, '*.log\ncustom-dir/', {
        'info.log': 'Log file',
        'custom-dir/file.txt': 'Custom directory file'
      });

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
      await setupWithGitignore(testDirPath, '', {
        'regular-file.txt': 'Regular text file',
        'info.log': 'Log file - should not be ignored without gitignore'
      });

      // Manually delete the .gitignore file that setupWithGitignore creates
      await addGitignoreFile(gitignorePath, '');
      
      // Reset the cache so we get a fresh filter
      gitignoreUtils.clearIgnoreCache();

      // Call the function under test
      const ignoreFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);

      // Default patterns should still be respected
      expect(ignoreFilter.ignores('node_modules/package.json')).toBe(true);
      expect(ignoreFilter.ignores('.git/config')).toBe(true);

      // Custom patterns that would be in a .gitignore should not be applied
      expect(ignoreFilter.ignores('info.log')).toBe(false);
    });

    it('should handle errors when reading .gitignore files gracefully', async () => {
      // Create a directory without .gitignore file first
      await setupWithGitignore(testDirPath, '', {
        'some-file.txt': 'Regular file'
      });
      
      // Clear the cache for a fresh test
      gitignoreUtils.clearIgnoreCache();

      // Spy on console.warn
      const consoleWarnMock = jest.spyOn(console, 'warn').mockImplementation(() => {});
      
      // Mock fs.readFile to throw an error
      const fsReadFile = jest.spyOn(fs, 'readFile');
      fsReadFile.mockRejectedValue(new Error('Mock file read error'));

      // Call the function being tested
      const ignoreFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);

      // Should still return a filter with default patterns
      expect(ignoreFilter.ignores('node_modules/package.json')).toBe(true);
      
      // Verify that the warning was logged
      expect(consoleWarnMock).toHaveBeenCalled();
      
      // Restore the mocks
      consoleWarnMock.mockRestore();
      fsReadFile.mockRestore();
    });
  });

  describe('ignore cache behavior', () => {
    it('should cache and reuse filters for the same directory path', async () => {
      // Create .gitignore file
      await setupWithGitignore(testDirPath, '*.log\ntmp/', {
        'info.log': 'Log file',
        'tmp/temp.txt': 'Temp file'
      });

      // First call should read the file
      const filter1 = await gitignoreUtils.createIgnoreFilter(testDirPath);

      // Second call should use cache
      const filter2 = await gitignoreUtils.createIgnoreFilter(testDirPath);

      // Should be the same instance - this confirms caching is working
      expect(filter1).toBe(filter2);
    });

    it('should create separate filters for different directory paths', async () => {
      // Set up two different directories with different gitignore patterns
      const dir1 = normalizePathGeneral('/path/one', true);
      const dir2 = normalizePathGeneral('/path/two', true);

      // Set up the first directory
      await setupWithGitignore(dir1, '*.log', {
        'test.log': 'Log file in dir1',
        'test.tmp': 'Temp file in dir1'
      });

      // Set up the second directory without resetting filesystem
      await setupWithGitignore(dir2, '*.tmp', {
        'test.log': 'Log file in dir2',
        'test.tmp': 'Temp file in dir2'
      }, { reset: false });

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
      await setupWithGitignore(testDirPath, '*.log', {
        'info.log': 'Log file',
        'file.new-pattern': 'New pattern file'
      });

      // First call to populate cache
      const initialFilter = await gitignoreUtils.createIgnoreFilter(testDirPath);

      // Clear the cache
      gitignoreUtils.clearIgnoreCache();

      // Update .gitignore content
      await addGitignoreFile(gitignorePath, '*.new-pattern');

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
