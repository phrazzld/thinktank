/**
 * Tests for the mockFsUtils utility
 */
// We'll use the mockedFs instead of importing fs directly
// Stats is only needed for typechecking, not used directly
import { jest } from '@jest/globals';
import { 
  mockedFs, 
  setupMockFs, 
  resetMockFs, 
  FsMockConfig,
  mockAccess,
  mockReadFile,
  mockStat,
  mockReaddir,
  createFsError,
  MockedStats
} from '../mockFsUtils';

// We don't mock fs/promises here because mockFsUtils already does that
// Instead, we test that mockFsUtils correctly configures the mocks

describe('mockFsUtils core functions', () => {
  beforeEach(() => {
    // Clear all mocks before each test to start fresh
    jest.clearAllMocks();
  });

  describe('resetMockFs', () => {
    it('should reset all fs mock functions', () => {
      // Setup some mocks first
      mockedFs.access.mockRejectedValueOnce(new Error('Access denied'));
      mockedFs.readFile.mockResolvedValueOnce('Some content');
      
      // Initial mocks should be configured
      expect(mockedFs.access.mock.calls.length).toBe(0);
      expect(mockedFs.readFile.mock.calls.length).toBe(0);
      
      // Reset the mocks
      resetMockFs();
      
      // All mock implementations should be cleared
      expect(mockedFs.access.mock.calls).toHaveLength(0);
      expect(mockedFs.access.mock.instances).toHaveLength(0);
      expect(mockedFs.readFile.mock.calls).toHaveLength(0);
      expect(mockedFs.readFile.mock.instances).toHaveLength(0);
      expect(mockedFs.writeFile.mock.calls).toHaveLength(0);
      expect(mockedFs.writeFile.mock.instances).toHaveLength(0);
      expect(mockedFs.stat.mock.calls).toHaveLength(0);
      expect(mockedFs.stat.mock.instances).toHaveLength(0);
      expect(mockedFs.readdir.mock.calls).toHaveLength(0);
      expect(mockedFs.readdir.mock.instances).toHaveLength(0);
      expect(mockedFs.mkdir.mock.calls).toHaveLength(0);
      expect(mockedFs.mkdir.mock.instances).toHaveLength(0);
    });
  });

  describe('setupMockFs', () => {
    beforeEach(() => {
      resetMockFs();
    });

    it('should configure default fs.access behavior to succeed', async () => {
      // Setup with defaults
      setupMockFs();
      
      // Should succeed by default
      await expect(mockedFs.access('/any/path')).resolves.toBeUndefined();
    });

    it('should configure custom fs.access behavior based on config', async () => {
      // Setup with custom default behavior (deny access)
      const config: FsMockConfig = {
        defaultAccessBehavior: false,
        defaultAccessErrorCode: 'ENOENT'
      };
      setupMockFs(config);
      
      // Should fail with the specified error
      await expect(mockedFs.access('/any/path')).rejects.toMatchObject({
        code: 'ENOENT'
      });
    });

    it('should configure default fs.readFile behavior', async () => {
      // Setup with defaults
      setupMockFs();
      
      // Should return empty string by default
      await expect(mockedFs.readFile('/any/path', 'utf-8')).resolves.toBe('');
    });

    it('should configure custom fs.readFile behavior based on config', async () => {
      // Setup with custom default content
      const config: FsMockConfig = {
        defaultFileContent: 'Default file content'
      };
      setupMockFs(config);
      
      // Should return the specified content
      await expect(mockedFs.readFile('/any/path', 'utf-8')).resolves.toBe('Default file content');
    });

    it('should configure default fs.stat behavior', async () => {
      // Setup with defaults
      setupMockFs();
      
      // Should return default stats (file, not directory)
      const stats = await mockedFs.stat('/any/path');
      expect(stats.isFile()).toBe(true);
      expect(stats.isDirectory()).toBe(false);
    });

    it('should configure custom fs.stat behavior based on config', async () => {
      // Setup with custom default stats (directory, not file)
      const config: FsMockConfig = {
        defaultStats: {
          isFile: () => false,
          isDirectory: () => true,
          size: 1024
        }
      };
      setupMockFs(config);
      
      // Should return the specified stats
      const stats = await mockedFs.stat('/any/path');
      expect(stats.isFile()).toBe(false);
      expect(stats.isDirectory()).toBe(true);
      expect(stats.size).toBe(1024);
    });
  });

  describe('mockAccess', () => {
    beforeEach(() => {
      resetMockFs();
      setupMockFs(); // Start with default configuration
    });

    it('should allow access to specific paths when configured to allow', async () => {
      // Configure specific path to allow access
      mockAccess('/allowed/path', true);
      
      // The configured path should be accessible
      await expect(mockedFs.access('/allowed/path')).resolves.toBeUndefined();
    });

    it('should deny access to specific paths when configured to deny', async () => {
      // Configure specific path to deny access with default error code
      mockAccess('/denied/path', false);
      
      // The configured path should not be accessible
      await expect(mockedFs.access('/denied/path')).rejects.toMatchObject({
        code: 'ENOENT', // Default error code
        path: '/denied/path'
      });
    });

    it('should deny access with custom error codes and messages', async () => {
      // Configure specific path to deny access with custom error
      mockAccess('/permission/denied/path', false, {
        errorCode: 'EACCES',
        errorMessage: 'Permission denied'
      });
      
      // The configured path should not be accessible with custom error
      await expect(mockedFs.access('/permission/denied/path')).rejects.toMatchObject({
        code: 'EACCES',
        message: expect.stringContaining('Permission denied'),
        path: '/permission/denied/path'
      });
    });

    it('should support pattern matching with regular expressions', async () => {
      // Configure regex pattern to deny access
      mockAccess(/^\/sensitive\/.*$/, false);
      
      // All paths matching the pattern should not be accessible
      await expect(mockedFs.access('/sensitive/file1')).rejects.toHaveProperty('code', 'ENOENT');
      await expect(mockedFs.access('/sensitive/folder/file2')).rejects.toHaveProperty('code', 'ENOENT');
    });

    it('should prioritize specific path matches over default behavior', async () => {
      // Setup with default access denied
      setupMockFs({ defaultAccessBehavior: false });
      
      // But allow specific path
      mockAccess('/special/allowed/path', true);
      
      // Default behavior is to deny
      await expect(mockedFs.access('/random/path')).rejects.toHaveProperty('code');
      
      // But specific path should be allowed
      await expect(mockedFs.access('/special/allowed/path')).resolves.toBeUndefined();
    });

    it('should allow overriding existing path configurations', async () => {
      // Initially deny access
      mockAccess('/config/path', false);
      await expect(mockedFs.access('/config/path')).rejects.toHaveProperty('code');
      
      // Then override to allow access
      mockAccess('/config/path', true);
      await expect(mockedFs.access('/config/path')).resolves.toBeUndefined();
    });

    it('should handle multiple path patterns in correct order', async () => {
      // Setup with conflicting patterns
      mockAccess(/^\/multi\/.*$/, false); // Deny all in /multi/
      mockAccess('/multi/special', true); // But allow one special file
      
      // The more specific path should take precedence
      await expect(mockedFs.access('/multi/regular')).rejects.toHaveProperty('code');
      await expect(mockedFs.access('/multi/special')).resolves.toBeUndefined();
    });
  });

  describe('mockReadFile', () => {
    beforeEach(() => {
      resetMockFs();
      setupMockFs(); // Start with default configuration
    });

    it('should return specific content for exact path matches', async () => {
      // Configure specific file to return custom content
      mockReadFile('/path/to/file.txt', 'Custom file content');
      
      // The configured file should return the specified content
      await expect(mockedFs.readFile('/path/to/file.txt', 'utf-8')).resolves.toBe('Custom file content');
    });

    it('should support pattern matching with regular expressions', async () => {
      // Configure all JSON files to return a specific JSON content
      mockReadFile(/\.json$/, '{"test": true}');
      
      // All paths matching the pattern should return the specified content
      await expect(mockedFs.readFile('/path/to/config.json', 'utf-8')).resolves.toBe('{"test": true}');
      await expect(mockedFs.readFile('/another/file.json', 'utf-8')).resolves.toBe('{"test": true}');
    });

    it('should reject with specific errors when configured', async () => {
      // Configure specific file to throw an error
      const error = createFsError('EBUSY', 'File is locked', 'readFile', '/locked/file.txt');
      mockReadFile('/locked/file.txt', error);
      
      // Reading the file should reject with the specified error
      await expect(mockedFs.readFile('/locked/file.txt', 'utf-8')).rejects.toMatchObject({
        message: expect.stringContaining('File is locked'),
        code: 'EBUSY'
      });
    });

    it('should support pattern matching for error cases', async () => {
      // Configure all files in a specific directory to throw permission errors
      const error = createFsError('EACCES', 'Permission denied', 'readFile', '/secure/path');
      mockReadFile(/^\/secure\/.*$/, error);
      
      // Reading any file in the directory should reject with the specified error
      await expect(mockedFs.readFile('/secure/file1.txt', 'utf-8')).rejects.toMatchObject({
        code: 'EACCES'
      });
      await expect(mockedFs.readFile('/secure/nested/file2.txt', 'utf-8')).rejects.toMatchObject({
        code: 'EACCES'
      });
    });

    it('should fall back to default behavior for non-matching paths', async () => {
      // Setup specific file with content
      mockReadFile('/specific/file.txt', 'Specific content');
      
      // Setup default content
      setupMockFs({ defaultFileContent: 'Default content' });
      
      // The specific file should return its configured content
      await expect(mockedFs.readFile('/specific/file.txt', 'utf-8')).resolves.toBe('Specific content');
      
      // Other files should return the default content
      await expect(mockedFs.readFile('/other/file.txt', 'utf-8')).resolves.toBe('Default content');
    });

    it('should support Buffer return values', async () => {
      // Configure a binary file to return a string that will be treated as Buffer
      mockReadFile('/binary/file.bin', 'binary-content');
      
      // The file should return a Buffer when no encoding is specified
      const result = await mockedFs.readFile('/binary/file.bin');
      expect(Buffer.isBuffer(result)).toBe(true);
      expect(result.toString()).toBe('binary-content');
    });

    it('should allow overriding previously configured behavior', async () => {
      // Initially configure a file with content
      mockReadFile('/config/file.json', '{"version": 1}');
      await expect(mockedFs.readFile('/config/file.json', 'utf-8')).resolves.toBe('{"version": 1}');
      
      // Then override with new content
      mockReadFile('/config/file.json', '{"version": 2}');
      await expect(mockedFs.readFile('/config/file.json', 'utf-8')).resolves.toBe('{"version": 2}');
      
      // Then override with an error
      const error = createFsError('ENOENT', 'File not found', 'readFile', '/config/file.json');
      mockReadFile('/config/file.json', error);
      await expect(mockedFs.readFile('/config/file.json', 'utf-8')).rejects.toHaveProperty('code', 'ENOENT');
    });

    it('should handle the encoding option correctly', async () => {
      // Configure a file with content
      mockReadFile('/test/file.txt', 'File content');
      
      // Test with different encoding options
      await expect(mockedFs.readFile('/test/file.txt', 'utf-8')).resolves.toBe('File content');
      await expect(mockedFs.readFile('/test/file.txt', { encoding: 'utf-8' })).resolves.toBe('File content');
      
      // Without encoding, it should return a Buffer
      const result = await mockedFs.readFile('/test/file.txt');
      expect(Buffer.isBuffer(result)).toBe(true);
      expect(result.toString()).toBe('File content');
    });
  });

  describe('mockStat', () => {
    beforeEach(() => {
      resetMockFs();
      setupMockFs(); // Start with default configuration
    });

    it('should return specific stats for exact path matches', async () => {
      // Configure a file with custom stats
      const customStats: MockedStats = {
        isFile: () => true,
        isDirectory: () => false,
        size: 12345,
        birthtime: new Date('2023-01-01'),
        mtime: new Date('2023-01-02')
      };
      mockStat('/path/to/file.txt', customStats);
      
      // The stat call should return the specified stats
      const stats = await mockedFs.stat('/path/to/file.txt');
      expect(stats.isFile()).toBe(true);
      expect(stats.isDirectory()).toBe(false);
      expect(stats.size).toBe(12345);
      expect(stats.birthtime).toEqual(new Date('2023-01-01'));
      expect(stats.mtime).toEqual(new Date('2023-01-02'));
    });

    it('should return directory stats when configured', async () => {
      // Configure a directory
      mockStat('/path/to/directory', {
        isFile: () => false,
        isDirectory: () => true
      });
      
      // The stat call should return directory stats
      const stats = await mockedFs.stat('/path/to/directory');
      expect(stats.isFile()).toBe(false);
      expect(stats.isDirectory()).toBe(true);
    });

    it('should support pattern matching with regular expressions', async () => {
      // Configure all JS files to return the same stats
      mockStat(/\.js$/, {
        isFile: () => true,
        isDirectory: () => false,
        size: 500
      });
      
      // All paths matching the pattern should return the specified stats
      const stats1 = await mockedFs.stat('/path/to/file.js');
      expect(stats1.isFile()).toBe(true);
      expect(stats1.size).toBe(500);
      
      const stats2 = await mockedFs.stat('/another/script.js');
      expect(stats2.isFile()).toBe(true);
      expect(stats2.size).toBe(500);
    });

    it('should reject with specific errors when configured', async () => {
      // Configure specific file to throw an error
      const error = createFsError('ENOENT', 'File not found', 'stat', '/nonexistent/file.txt');
      mockStat('/nonexistent/file.txt', error);
      
      // The stat call should reject with the specified error
      await expect(mockedFs.stat('/nonexistent/file.txt')).rejects.toMatchObject({
        code: 'ENOENT',
        message: expect.stringContaining('File not found')
      });
    });

    it('should support pattern matching for error cases', async () => {
      // Configure all files in a specific directory to throw errors
      const error = createFsError('EACCES', 'Permission denied', 'stat', '/secure/path');
      mockStat(/^\/secure\/.*$/, error);
      
      // Stat calls for matching paths should reject with the specified error
      await expect(mockedFs.stat('/secure/file1.txt')).rejects.toMatchObject({
        code: 'EACCES'
      });
      await expect(mockedFs.stat('/secure/nested/file2.txt')).rejects.toMatchObject({
        code: 'EACCES'
      });
    });

    it('should fall back to default behavior for non-matching paths', async () => {
      // Setup specific file with custom stats
      mockStat('/specific/file.txt', {
        isFile: () => true,
        size: 1000
      });
      
      // Setup custom default stats
      setupMockFs({
        defaultStats: {
          isFile: () => false,
          isDirectory: () => true
        }
      });
      
      // The specific file should return its configured stats
      const specificStats = await mockedFs.stat('/specific/file.txt');
      expect(specificStats.isFile()).toBe(true);
      expect(specificStats.size).toBe(1000);
      
      // Other paths should return the default stats
      const defaultStats = await mockedFs.stat('/other/path');
      expect(defaultStats.isFile()).toBe(false);
      expect(defaultStats.isDirectory()).toBe(true);
    });

    it('should allow overriding previously configured behavior', async () => {
      // Initially configure a file to be a regular file
      mockStat('/config/path', {
        isFile: () => true,
        isDirectory: () => false
      });
      
      const initialStats = await mockedFs.stat('/config/path');
      expect(initialStats.isFile()).toBe(true);
      expect(initialStats.isDirectory()).toBe(false);
      
      // Then override to be a directory
      mockStat('/config/path', {
        isFile: () => false,
        isDirectory: () => true
      });
      
      const updatedStats = await mockedFs.stat('/config/path');
      expect(updatedStats.isFile()).toBe(false);
      expect(updatedStats.isDirectory()).toBe(true);
      
      // Then override with an error
      const error = createFsError('ENOENT', 'File not found', 'stat', '/config/path');
      mockStat('/config/path', error);
      
      await expect(mockedFs.stat('/config/path')).rejects.toHaveProperty('code', 'ENOENT');
    });
  });

  describe('mockReaddir', () => {
    beforeEach(() => {
      resetMockFs();
      setupMockFs(); // Start with default configuration
    });

    it('should return specific file lists for exact path matches', async () => {
      // Configure a directory with specific files
      const fileList = ['file1.txt', 'file2.js', 'subdirectory'];
      mockReaddir('/path/to/directory', fileList);
      
      // The readdir call should return the specified file list
      const files = await mockedFs.readdir('/path/to/directory');
      expect(files).toEqual(fileList);
    });

    it('should support pattern matching with regular expressions', async () => {
      // Configure all directories in a specific path to return the same content
      mockReaddir(/^\/config\/.*$/, ['settings.json', 'environment.json']);
      
      // All paths matching the pattern should return the specified content
      const files1 = await mockedFs.readdir('/config/app');
      expect(files1).toEqual(['settings.json', 'environment.json']);
      
      const files2 = await mockedFs.readdir('/config/user');
      expect(files2).toEqual(['settings.json', 'environment.json']);
    });

    it('should reject with specific errors when configured', async () => {
      // Configure specific directory to throw an error
      const error = createFsError('ENOENT', 'Directory not found', 'readdir', '/nonexistent/directory');
      mockReaddir('/nonexistent/directory', error);
      
      // The readdir call should reject with the specified error
      await expect(mockedFs.readdir('/nonexistent/directory')).rejects.toMatchObject({
        code: 'ENOENT',
        message: expect.stringContaining('Directory not found')
      });
    });

    it('should support pattern matching for error cases', async () => {
      // Configure all directories in a specific path to throw permission errors
      const error = createFsError('EACCES', 'Permission denied', 'readdir', '/secure/path');
      mockReaddir(/^\/secure\/.*$/, error);
      
      // Readdir calls for matching paths should reject with the specified error
      await expect(mockedFs.readdir('/secure/folder1')).rejects.toMatchObject({
        code: 'EACCES'
      });
      await expect(mockedFs.readdir('/secure/nested/folder2')).rejects.toMatchObject({
        code: 'EACCES'
      });
    });

    it('should fall back to default behavior for non-matching paths', async () => {
      // Setup specific directory with content
      mockReaddir('/specific/directory', ['file1.txt', 'file2.txt']);
      
      // Default behavior is an empty directory
      const specificFiles = await mockedFs.readdir('/specific/directory');
      expect(specificFiles).toEqual(['file1.txt', 'file2.txt']);
      
      // Other directories should return the default (empty array)
      const defaultFiles = await mockedFs.readdir('/other/directory');
      expect(defaultFiles).toEqual([]);
    });

    it('should allow overriding previously configured behavior', async () => {
      // Initially configure a directory with files
      mockReaddir('/config/directory', ['initial1.txt', 'initial2.txt']);
      const initialFiles = await mockedFs.readdir('/config/directory');
      expect(initialFiles).toEqual(['initial1.txt', 'initial2.txt']);
      
      // Then override with new content
      mockReaddir('/config/directory', ['updated1.txt', 'updated2.txt']);
      const updatedFiles = await mockedFs.readdir('/config/directory');
      expect(updatedFiles).toEqual(['updated1.txt', 'updated2.txt']);
      
      // Then override with an error
      const error = createFsError('ENOENT', 'Directory not found', 'readdir', '/config/directory');
      mockReaddir('/config/directory', error);
      await expect(mockedFs.readdir('/config/directory')).rejects.toHaveProperty('code', 'ENOENT');
    });

    it('should handle the withFileTypes option correctly', async () => {
      // Mock a directory with files and subdirectories
      mockReaddir('/mixed/directory', ['file1.txt', 'file2.js', 'subdirectory']);
      
      // Without withFileTypes option, it should return string array
      const files = await mockedFs.readdir('/mixed/directory');
      expect(Array.isArray(files)).toBe(true);
      expect(files).toEqual(['file1.txt', 'file2.js', 'subdirectory']);
      
      // With withFileTypes option, it should still work but return simple strings
      // Note: Full Dirent objects aren't supported in this implementation
      const filesWithTypes = await mockedFs.readdir('/mixed/directory', { withFileTypes: true });
      expect(Array.isArray(filesWithTypes)).toBe(true);
      expect(filesWithTypes).toEqual(['file1.txt', 'file2.js', 'subdirectory']);
    });
  });
});