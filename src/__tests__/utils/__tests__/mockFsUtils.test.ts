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
  FsMockConfig 
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
});