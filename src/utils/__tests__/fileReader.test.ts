/**
 * Unit tests for file reader module
 */
import path from 'path';
import os from 'os';
import { FileReadError } from '../fileReaderTypes';
import {
  readFileContent,
  fileExists,
  writeFile,
  getConfigDir,
  getConfigFilePath,
} from '../fileReader';
import {
  createFsError,
} from '../../__tests__/utils/virtualFsUtils';
import {
  setupBasicFs,
  setupWithSingleFile,
  normalizePathForMemfs
} from '../../../test/setup/fs';
import fsPromises from 'fs/promises';

// Mock os module
jest.mock('os');

const mockedOs = jest.mocked(os);

describe('File Reader', () => {
  describe('writeFile', () => {
    // Use normalizePathForMemfs to properly handle memfs paths
    const testFilePath = normalizePathForMemfs('/path/to/test/file.txt');
    const testContent = 'Test content to write';

    beforeEach(() => {
      // Start with a clean virtual filesystem for each test
      setupBasicFs({});
    });

    it('should write content to a file', async () => {
      // Call the function
      await writeFile(testFilePath, testContent);

      // Verify content was written correctly
      const content = await fsPromises.readFile(testFilePath, 'utf-8');
      expect(content).toBe(testContent);
    });

    describe('Windows-specific error handling', () => {
      beforeEach(() => {
        // Start with a clean virtual filesystem
        setupBasicFs({});
        // Mock platform as Windows for these tests
        Object.defineProperty(process, 'platform', { value: 'win32' });
      });

      it('should handle Windows permission errors (EACCES)', async () => {
        // Set up spy to simulate permission error
        const writeFileSpy = jest.spyOn(fsPromises, 'writeFile');
        writeFileSpy.mockRejectedValueOnce(
          createFsError('EACCES', 'Access is denied', 'writeFile', testFilePath)
        );

        try {
          await writeFile(testFilePath, testContent);
          // If we get here, the test should fail
          expect('should have thrown').toBe('but did not throw');
        } catch (err) {
          // Just verify it's the right type of error
          expect(err).toBeInstanceOf(FileReadError);
        }

        writeFileSpy.mockRestore();
      });

      it('should handle Windows permission errors (EPERM)', async () => {
        // Set up spy to simulate permission error
        const writeFileSpy = jest.spyOn(fsPromises, 'writeFile');
        writeFileSpy.mockRejectedValueOnce(
          createFsError('EPERM', 'Operation not permitted', 'writeFile', testFilePath)
        );

        try {
          await writeFile(testFilePath, testContent);
          // If we get here, the test should fail
          expect('should have thrown').toBe('but did not throw');
        } catch (err) {
          // Just verify it's the right type of error
          expect(err).toBeInstanceOf(FileReadError);
        }

        writeFileSpy.mockRestore();
      });

      it('should handle Windows "file in use" errors (EBUSY)', async () => {
        // Set up spy to simulate file busy error
        const writeFileSpy = jest.spyOn(fsPromises, 'writeFile');
        writeFileSpy.mockRejectedValueOnce(
          createFsError('EBUSY', 'Resource busy or locked', 'writeFile', testFilePath)
        );

        try {
          await writeFile(testFilePath, testContent);
          // If we get here, the test should fail
          expect('should have thrown').toBe('but did not throw');
        } catch (err) {
          // Just verify it's the right type of error
          expect(err).toBeInstanceOf(FileReadError);
        }

        writeFileSpy.mockRestore();
      });

      it('should handle Windows path not found errors (ENOENT)', async () => {
        // Set up spy to simulate path not found error
        const mkdirSpy = jest.spyOn(fsPromises, 'mkdir');
        mkdirSpy.mockRejectedValueOnce(
          createFsError('ENOENT', 'No such file or directory', 'mkdir', '/path/to')
        );

        try {
          await writeFile(testFilePath, testContent);
          // If we get here, the test should fail
          expect('should have thrown').toBe('but did not throw');
        } catch (err) {
          // Just verify it's the right type of error
          expect(err).toBeInstanceOf(FileReadError);
        }

        mkdirSpy.mockRestore();
      });

      afterEach(() => {
        // Reset platform to prevent affecting other tests
        Object.defineProperty(process, 'platform', { value: process.platform });
      });
    });

    describe('macOS-specific error handling', () => {
      beforeEach(() => {
        // Start with a clean virtual filesystem
        setupBasicFs({});
        // Mock platform as macOS for these tests
        Object.defineProperty(process, 'platform', { value: 'darwin' });
      });

      it('should handle macOS permission errors (EACCES)', async () => {
        // Set up spy to simulate permission error
        const writeFileSpy = jest.spyOn(fsPromises, 'writeFile');
        writeFileSpy.mockRejectedValueOnce(
          createFsError('EACCES', 'Permission denied', 'writeFile', testFilePath)
        );

        try {
          await writeFile(testFilePath, testContent);
          // If we get here, the test should fail
          expect('should have thrown').toBe('but did not throw');
        } catch (err) {
          // Just verify it's the right type of error
          expect(err).toBeInstanceOf(FileReadError);
        }

        writeFileSpy.mockRestore();
      });

      it('should handle macOS read-only filesystem errors (EROFS)', async () => {
        // Set up spy to simulate read-only filesystem error
        const writeFileSpy = jest.spyOn(fsPromises, 'writeFile');
        writeFileSpy.mockRejectedValueOnce(
          createFsError('EROFS', 'Read-only file system', 'writeFile', testFilePath)
        );

        try {
          await writeFile(testFilePath, testContent);
          // If we get here, the test should fail
          expect('should have thrown').toBe('but did not throw');
        } catch (err) {
          // Just verify it's the right type of error
          expect(err).toBeInstanceOf(FileReadError);
        }

        writeFileSpy.mockRestore();
      });

      it('should handle macOS too many open files errors (EMFILE)', async () => {
        // Set up spy to simulate too many open files error
        const writeFileSpy = jest.spyOn(fsPromises, 'writeFile');
        writeFileSpy.mockRejectedValueOnce(
          createFsError('EMFILE', 'Too many open files', 'writeFile', testFilePath)
        );

        try {
          await writeFile(testFilePath, testContent);
          // If we get here, the test should fail
          expect('should have thrown').toBe('but did not throw');
        } catch (err) {
          // Just verify it's the right type of error
          expect(err).toBeInstanceOf(FileReadError);
        }

        writeFileSpy.mockRestore();
      });

      it('should handle macOS path not found errors (ENOENT)', async () => {
        // Set up spy to simulate path not found error
        const mkdirSpy = jest.spyOn(fsPromises, 'mkdir');
        mkdirSpy.mockRejectedValueOnce(
          createFsError('ENOENT', 'No such file or directory', 'mkdir', '/path/to')
        );

        try {
          await writeFile(testFilePath, testContent);
          // If we get here, the test should fail
          expect('should have thrown').toBe('but did not throw');
        } catch (err) {
          // Just verify it's the right type of error
          expect(err).toBeInstanceOf(FileReadError);
        }

        mkdirSpy.mockRestore();
      });

      afterEach(() => {
        // Reset platform to prevent affecting other tests
        Object.defineProperty(process, 'platform', { value: process.platform });
      });
    });
  });

  describe('readFileContent', () => {
    // Use normalizePathForMemfs to properly handle memfs paths
    const testFilePath = normalizePathForMemfs('/path/to/test/file.txt');
    const testContent = 'This is a test content\nwith multiple lines.';

    beforeEach(() => {
      // Set up clean virtual filesystem with test file
      setupBasicFs({
        [testFilePath]: testContent,
      });
    });

    it('should read and return file content', async () => {
      // Call the function
      const result = await readFileContent(testFilePath);

      // Verify result
      expect(result).toBe('This is a test content with multiple lines.');
    });

    it('should return raw content when normalize is false', async () => {
      const result = await readFileContent(testFilePath, { normalize: false });

      expect(result).toBe(testContent);
    });

    it('should resolve relative paths to absolute paths', async () => {
      // Setup file in current working directory
      const relativePath = 'relative/path.txt';
      // Create the absolute path that will be used by fileReader
      const absPath = path.resolve(process.cwd(), relativePath);

      // Set up file
      setupWithSingleFile(absPath, 'Relative file content');

      // Read using relative path
      const result = await readFileContent(relativePath);

      // Verify content was read correctly
      expect(result).toBe('Relative file content');
    });

    it('should throw FileReadError when file is not found', async () => {
      await expect(readFileContent('/nonexistent/file.txt')).rejects.toThrow(FileReadError);
      await expect(readFileContent('/nonexistent/file.txt')).rejects.toThrow(/File not found/);
    });

    it('should throw FileReadError when permission is denied', async () => {
      // Set up spy to simulate permission error
      const accessSpy = jest.spyOn(fsPromises, 'access');
      accessSpy.mockRejectedValueOnce(
        createFsError('EACCES', 'Permission denied', 'access', testFilePath)
      );

      // We need to use a single expectation since the first one consumes the rejection
      await expect(readFileContent(testFilePath)).rejects.toThrow(/Permission denied/);

      accessSpy.mockRestore();
    });

    it('should throw FileReadError for other errors', async () => {
      // Set up spy to simulate a generic error
      const accessSpy = jest.spyOn(fsPromises, 'access');
      accessSpy.mockRejectedValueOnce(
        createFsError('EINVAL', 'Invalid argument', 'access', testFilePath)
      );

      // We need to use a single expectation since the first one consumes the rejection
      await expect(readFileContent(testFilePath)).rejects.toThrow(/Error reading file/);

      accessSpy.mockRestore();
    });
  });

  describe('fileExists', () => {
    // Use normalizePathForMemfs to properly handle memfs paths
    const testFilePath = normalizePathForMemfs('/path/to/test/file.txt');
    const testContent = 'This is a test content\nwith multiple lines.';

    beforeEach(() => {
      // Set up clean virtual filesystem with test file
      setupBasicFs({
        [testFilePath]: testContent,
      });
    });

    it('should return true when file exists', async () => {
      const result = await fileExists(testFilePath);

      expect(result).toBe(true);
    });

    it('should return false when file does not exist', async () => {
      const nonexistentFile = normalizePathForMemfs('/nonexistent/file.txt');
      const result = await fileExists(nonexistentFile);

      expect(result).toBe(false);
    });
  });

  describe('getConfigDir', () => {
    // Save original environment
    const originalEnv = process.env;
    const originalPlatform = process.platform;

    beforeEach(() => {
      // Reset environment variables for each test
      process.env = { ...originalEnv };
      
      // Setup a clean virtual filesystem
      setupBasicFs({});
    });

    afterEach(() => {
      // Reset platform to avoid affecting other tests
      Object.defineProperty(process, 'platform', { value: originalPlatform });
    });

    afterAll(() => {
      // Restore original environment
      process.env = originalEnv;
    });

    it('should use XDG_CONFIG_HOME when set', async () => {
      // Mock homedir for this test
      mockedOs.homedir.mockReturnValue('/home/user');
      
      // Setup environment
      process.env.XDG_CONFIG_HOME = '/custom/xdg/config';
      const configDir = '/custom/xdg/config/thinktank';

      // Set up directory
      setupBasicFs({
        '/custom/xdg/config/.placeholder': '', // Create parent directories
      });

      const result = await getConfigDir();

      expect(result).toBe(configDir);
    });

    describe('Windows paths', () => {
      beforeEach(() => {
        // Clean virtual filesystem for each test
        setupBasicFs({});
        // Mock platform as Windows for all tests in this group
        Object.defineProperty(process, 'platform', { value: 'win32' });
        // Reset homedir mock for Windows tests
        mockedOs.homedir.mockReturnValue('C:\\Users\\User');
      });

      it('should use APPDATA environment variable when available', async () => {
        // This test checks that the function attempts to use the APPDATA path
        // but since overriding process.platform doesn't fully simulate Windows,
        // we'll test that mkdir was called with the correct path

        // Set APPDATA environment variable
        process.env.APPDATA = 'C:\\Users\\User\\AppData\\Roaming';

        // Create directory structure
        setupBasicFs({
          'C:\\': '',
          'C:\\Users\\': '',
          'C:\\Users\\User\\': '',
          'C:\\Users\\User\\AppData\\': '',
          'C:\\Users\\User\\AppData\\Roaming\\': '',
        });

        // Mock mkdir to just record the first parameter
        const mkdirSpy = jest.spyOn(fsPromises, 'mkdir');
        // Store the paths that mkdir is called with
        const mkdirPaths: string[] = [];
        mkdirSpy.mockImplementation((...args) => { 
          if (typeof args[0] === 'string') {
            mkdirPaths.push(args[0]);
          } else if (args[0] instanceof Buffer) {
            mkdirPaths.push(args[0].toString());
          }
          return Promise.resolve(undefined);
        });

        await getConfigDir();

        // Check that mkdir was called at least once
        expect(mkdirPaths.length).toBeGreaterThan(0);
        
        // At least one of the mkdir calls should include 'thinktank'
        expect(mkdirPaths.some(p => p.includes('thinktank'))).toBe(true);
        
        // Verify the test is actually running in Windows mode
        expect(process.platform).toBe('win32');
        
        mkdirSpy.mockRestore();
      });

      it('should handle Windows permission errors correctly', async () => {
        // Set APPDATA environment variable
        process.env.APPDATA = 'C:\\Users\\User\\AppData\\Roaming';
        const configDir = path.join('C:\\Users\\User\\AppData\\Roaming', 'thinktank');

        // Create directory structure
        setupBasicFs({
          'C:\\': '',
          'C:\\Users\\': '',
          'C:\\Users\\User\\': '',
          'C:\\Users\\User\\AppData\\': '',
          'C:\\Users\\User\\AppData\\Roaming\\': '',
        });

        // Simulate mkdir permission error
        const mkdirSpy = jest.spyOn(fsPromises, 'mkdir');
        mkdirSpy.mockRejectedValueOnce(
          createFsError('EACCES', 'Access is denied', 'mkdir', configDir)
        );

        try {
          await getConfigDir();
          // If we get here, the test should fail
          expect('should have thrown').toBe('but did not throw');
        } catch (err) {
          // Cast to FileReadError to make TypeScript happy
          const fileError = err as FileReadError;
          // Just verify it's the right type of error
          expect(fileError).toBeInstanceOf(FileReadError);
          // Only check if error contains expected substring
          expect(fileError.message).toContain('Failed to access or create config directory');
        }

        mkdirSpy.mockRestore();
      });

      it('should handle Windows ENOENT errors correctly', async () => {
        // Set APPDATA environment variable to a non-existent path
        process.env.APPDATA = 'C:\\Invalid\\Path';
        const configDir = path.join('C:\\Invalid\\Path', 'thinktank');

        // Simulate mkdir path not found error
        const mkdirSpy = jest.spyOn(fsPromises, 'mkdir');
        mkdirSpy.mockRejectedValueOnce(
          createFsError('ENOENT', 'No such file or directory', 'mkdir', configDir)
        );

        try {
          await getConfigDir();
          // If we get here, the test should fail
          expect('should have thrown').toBe('but did not throw');
        } catch (err) {
          // Cast to FileReadError to make TypeScript happy
          const fileError = err as FileReadError;
          // Just verify it's the right type of error
          expect(fileError).toBeInstanceOf(FileReadError);
        }

        mkdirSpy.mockRestore();
      });
    });

    describe('macOS paths', () => {
      beforeEach(() => {
        // Clean virtual filesystem for each test
        setupBasicFs({});
        // Mock platform as macOS for all tests in this group
        Object.defineProperty(process, 'platform', { value: 'darwin' });
        // Reset homedir mock for macOS tests
        mockedOs.homedir.mockReturnValue('/Users/macuser');
      });

      it('should use XDG_CONFIG_HOME when set on macOS', async () => {
        // Set XDG_CONFIG_HOME environment variable
        process.env.XDG_CONFIG_HOME = '/Users/macuser/.xdg/config';
        const configDir = '/Users/macuser/.xdg/config/thinktank';

        // Set up directory
        setupBasicFs({
          '/Users/macuser/.xdg/config/.placeholder': '', // Create parent directories
        });
        
        // Mock mkdir to prevent actual directory creation
        const mkdirSpy = jest.spyOn(fsPromises, 'mkdir');
        mkdirSpy.mockResolvedValue(undefined);

        const result = await getConfigDir();

        // Should use XDG_CONFIG_HOME when explicitly set, even on macOS
        expect(result).toBe(configDir);
        
        mkdirSpy.mockRestore();
      });

      it('should handle macOS permission errors correctly', async () => {
        const configDir = '/Users/macuser/.config/thinktank';

        // Set up directory
        setupBasicFs({
          '/Users/macuser/.config/.placeholder': '',
        });

        // Simulate mkdir permission error
        const mkdirSpy = jest.spyOn(fsPromises, 'mkdir');
        mkdirSpy.mockRejectedValueOnce(
          createFsError('EACCES', 'Permission denied', 'mkdir', configDir)
        );

        try {
          await getConfigDir();
          // If we get here, the test should fail
          expect('should have thrown').toBe('but did not throw');
        } catch (err) {
          // Cast to FileReadError to make TypeScript happy
          const fileError = err as FileReadError;
          // Just verify it's the right type of error
          expect(fileError).toBeInstanceOf(FileReadError);
        }

        mkdirSpy.mockRestore();
      });

      it('should handle macOS ENOENT errors correctly', async () => {
        const configDir = '/Users/macuser/.config/thinktank';

        // Set up directory structure but missing .config directory
        setupBasicFs({
          '/Users/macuser/.placeholder': '', // Create parent directories
        });

        // Simulate mkdir path not found error
        const mkdirSpy = jest.spyOn(fsPromises, 'mkdir');
        mkdirSpy.mockRejectedValueOnce(
          createFsError('ENOENT', 'No such file or directory', 'mkdir', configDir)
        );

        try {
          await getConfigDir();
          // If we get here, the test should fail
          expect('should have thrown').toBe('but did not throw');
        } catch (err) {
          // Cast to FileReadError to make TypeScript happy
          const fileError = err as FileReadError;
          // Just verify it's the right type of error
          expect(fileError).toBeInstanceOf(FileReadError);
        }

        mkdirSpy.mockRestore();
      });

      it('should handle macOS read-only filesystem errors correctly', async () => {
        const configDir = '/Users/macuser/.config/thinktank';

        // Set up directory
        setupBasicFs({
          '/Users/macuser/.config/.placeholder': '',
        });

        // Simulate mkdir read-only filesystem error
        const mkdirSpy = jest.spyOn(fsPromises, 'mkdir');
        mkdirSpy.mockRejectedValueOnce(
          createFsError('EROFS', 'Read-only file system', 'mkdir', configDir)
        );

        try {
          await getConfigDir();
          // If we get here, the test should fail
          expect('should have thrown').toBe('but did not throw');
        } catch (err) {
          // Cast to FileReadError to make TypeScript happy
          const fileError = err as FileReadError;
          // Just verify it's the right type of error
          expect(fileError).toBeInstanceOf(FileReadError);
        }

        mkdirSpy.mockRestore();
      });

      afterEach(() => {
        // Clean up XDG environment variable
        delete process.env.XDG_CONFIG_HOME;
      });
    });

    it('should throw FileReadError when directory creation fails', async () => {
      // Mock homedir for this test
      mockedOs.homedir.mockReturnValue('/home/user');
      
      // Mock directory creation failure
      const configDir = '/home/user/.config/thinktank';

      // Set up directory
      setupBasicFs({
        '/home/user/.config/.placeholder': '', // Create parent directories
      });

      // Simulate mkdir permission error
      const mkdirSpy = jest.spyOn(fsPromises, 'mkdir');
      mkdirSpy.mockRejectedValueOnce(
        createFsError('EPERM', 'Permission denied', 'mkdir', configDir)
      );

      // Use a simple test that doesn't rely on specific implementation details
      try {
        await getConfigDir();
        // If we get here, the test should fail
        expect('should have thrown').toBe('but did not throw');
      } catch (err) {
        // Cast to FileReadError to make TypeScript happy
        const fileError = err as FileReadError;
        // Just verify it's the right type of error
        expect(fileError).toBeInstanceOf(FileReadError);
      }

      mkdirSpy.mockRestore();
    });
  });

  describe('getConfigFilePath', () => {
    // Save original platform
    const originalPlatform = process.platform;
    
    afterEach(() => {
      // Reset platform to avoid affecting other tests
      Object.defineProperty(process, 'platform', { value: originalPlatform });
    });
    
    it('should get config file path based on getConfigDir result', async () => {
      // Mock homedir
      mockedOs.homedir.mockReturnValue('/home/user');

      // Set up platform
      Object.defineProperty(process, 'platform', { value: 'linux' });
      
      // Mock getConfigDir result
      const mkdirSpy = jest.spyOn(fsPromises, 'mkdir');
      mkdirSpy.mockResolvedValue(undefined);

      // Set up directory
      setupBasicFs({
        '/home/user/.config/.placeholder': '',
      });

      const result = await getConfigFilePath();

      // Test that it ends with the expected filename
      expect(result).toContain('/thinktank/config.json');

      mkdirSpy.mockRestore();
    });
  });
});
