/**
 * Unit tests for the filesystem test setup utilities
 */
import {
  mockFsModules,
  resetVirtualFs,
  getVirtualFs,
  createVirtualFs,
  addVirtualGitignoreFile,
} from '../virtualFsUtils';
import { normalizePath } from '../pathUtils';

// Setup mocks for fs modules
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Import the module we're testing
import {
  setupBasicFiles,
  setupProjectStructure,
  setupWithGitignore,
  mockFileExists,
  setupGitignoreMocking,
  setupCacheClearing,
} from '../fsTestSetup';

describe('fsTestSetup', () => {
  beforeEach(() => {
    resetVirtualFs();
  });

  describe('setupBasicFiles', () => {
    it('should reset filesystem and create the specified files', async () => {
      // Arrange
      const structure = {
        '/path/to/file.txt': 'File content',
        '/path/to/another/file.js': 'console.log("Hello");',
      };

      // Act
      setupBasicFiles(structure);

      // Assert
      const virtualFs = getVirtualFs();

      // Check if the files exist and have the right content
      expect(virtualFs.readFileSync('/path/to/file.txt', 'utf8')).toBe('File content');
      expect(virtualFs.readFileSync('/path/to/another/file.js', 'utf8')).toBe(
        'console.log("Hello");'
      );
    });
  });

  describe('setupProjectStructure', () => {
    it('should set up a basic project structure with the given files', async () => {
      // Arrange
      const basePath = '/project';
      const files = {
        'src/index.ts': 'console.log("Hello");',
        'README.md': '# Project',
      };

      // Act
      setupProjectStructure(basePath, files);

      // Assert
      const virtualFs = getVirtualFs();

      // Check that the project structure is correctly set up
      expect(virtualFs.existsSync(normalizePath('/project/src', true))).toBe(true);
      expect(virtualFs.readFileSync(normalizePath('/project/src/index.ts', true), 'utf8')).toBe(
        'console.log("Hello");'
      );
      expect(virtualFs.readFileSync(normalizePath('/project/README.md', true), 'utf8')).toBe(
        '# Project'
      );
    });
  });

  describe('setupWithGitignore', () => {
    it('should set up a directory with files and a gitignore file', async () => {
      // Arrange
      const dirPath = '/test-dir';
      const gitignoreContent = '*.log\nnode_modules/';
      const files = {
        'file.txt': 'Content',
        'file.log': 'Should be ignored',
        'node_modules/package.json': '{}',
      };

      // Act
      await setupWithGitignore(dirPath, gitignoreContent, files);

      // Assert
      const virtualFs = getVirtualFs();

      // Check that all files are created
      expect(virtualFs.readFileSync(normalizePath('/test-dir/.gitignore', true), 'utf8')).toBe(
        gitignoreContent
      );
      expect(virtualFs.readFileSync(normalizePath('/test-dir/file.txt', true), 'utf8')).toBe(
        'Content'
      );
      expect(virtualFs.readFileSync(normalizePath('/test-dir/file.log', true), 'utf8')).toBe(
        'Should be ignored'
      );
      expect(
        virtualFs.readFileSync(normalizePath('/test-dir/node_modules/package.json', true), 'utf8')
      ).toBe('{}');
    });
  });

  describe('mockFileExists', () => {
    it('should mock fileExists to work with the virtual filesystem', async () => {
      // Arrange
      const mockFn = jest.fn();
      const existingPath = normalizePath('/test/file.txt', true);
      const nonExistingPath = normalizePath('/test/missing.txt', true);

      // Create a file in the virtual filesystem
      setupBasicFiles({
        [existingPath]: 'Content',
      });

      // Act
      mockFileExists(mockFn);

      // Assert
      expect(await mockFn(existingPath)).toBe(true);
      expect(await mockFn(nonExistingPath)).toBe(false);
    });
  });

  describe('setupGitignoreMocking', () => {
    it('should set up gitignore mocking with cache clearing and filesystem integration', async () => {
      // Arrange
      // Use a real gitignore utils with clearIgnoreCache method instead of mockGitignoreUtils
      const gitignoreUtils = { clearIgnoreCache: jest.fn() };
      const mockedFileExists = jest.fn();

      // Act
      setupGitignoreMocking(gitignoreUtils, mockedFileExists);

      // Assert
      expect(gitignoreUtils.clearIgnoreCache).toHaveBeenCalled();

      // Create a file to test fileExists mock
      setupBasicFiles({ '/test.txt': 'content' });

      // Verify fileExists works with virtual filesystem
      expect(await mockedFileExists('/test.txt')).toBe(true);
      expect(await mockedFileExists('/nonexistent.txt')).toBe(false);
    });
  });

  describe('setupCacheClearing', () => {
    it('should return a function that clears caches and mocks when called', async () => {
      // Arrange
      const gitignoreUtils = { clearIgnoreCache: jest.fn() };
      const mockedFileExists = jest.fn();

      // Add some files to the filesystem
      setupBasicFiles({ '/test.txt': 'content' });
      const virtualFs = getVirtualFs();
      expect(virtualFs.existsSync('/test.txt')).toBe(true);

      // Create the beforeEach function
      const clearCachesFn = setupCacheClearing(gitignoreUtils, mockedFileExists);

      // Act - call the function as if it were used in beforeEach
      clearCachesFn();

      // Assert
      expect(gitignoreUtils.clearIgnoreCache).toHaveBeenCalled();

      // Verify filesystem was reset
      expect(virtualFs.existsSync('/test.txt')).toBe(false);

      // Create a new file to verify fileExists mock is set up
      setupBasicFiles({ '/new-test.txt': 'new content' });
      expect(await mockedFileExists('/new-test.txt')).toBe(true);
    });
  });

  describe('integration tests', () => {
    it('should demonstrate using createVirtualFs with multiple gitignore files', async () => {
      // This test shows how to set up multiple gitignore setups
      // without resetting the filesystem between them

      // Clear existing filesystem and create initial structure
      resetVirtualFs();

      // First create both directories with their base files
      createVirtualFs({
        '/dir1/file.txt': 'text file',
        '/dir1/log.log': 'log file',
        '/dir2/file.txt': 'another text file',
        '/dir2/config.json': '{"key": "value"}',
      });

      // Then add gitignore files to both directories
      await addVirtualGitignoreFile('/dir1/.gitignore', '*.log');
      await addVirtualGitignoreFile('/dir2/.gitignore', '*.json');

      // Verify both directories have their files and gitignore files
      const virtualFs = getVirtualFs();
      expect(virtualFs.readFileSync('/dir1/.gitignore', 'utf8')).toBe('*.log');
      expect(virtualFs.readFileSync('/dir2/.gitignore', 'utf8')).toBe('*.json');

      // Create a gitignoreUtils implementation to demonstrate how
      // a test would use these files with real behavior
      const gitignoreUtils = {
        clearIgnoreCache: jest.fn(),
        shouldIgnorePath: jest.fn(),
      };

      gitignoreUtils.shouldIgnorePath.mockImplementation((dirPath, filePath) => {
        if (dirPath === '/dir1' && filePath.endsWith('.log')) return Promise.resolve(true);
        if (dirPath === '/dir2' && filePath.endsWith('.json')) return Promise.resolve(true);
        return Promise.resolve(false);
      });

      // Verify the mocked behavior works as expected
      expect(await gitignoreUtils.shouldIgnorePath('/dir1', 'test.log')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath('/dir1', 'test.txt')).toBe(false);
      expect(await gitignoreUtils.shouldIgnorePath('/dir2', 'test.json')).toBe(true);
      expect(await gitignoreUtils.shouldIgnorePath('/dir2', 'test.log')).toBe(false);
    });
  });
});
