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
  mockAccess
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
});