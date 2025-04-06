/**
 * Unit tests for file reader module
 */
import { createVirtualFs, resetVirtualFs, mockFsModules, createFsError } from '../../__tests__/utils/virtualFsUtils';

// Setup mocks for fs modules
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Import modules after mocking
import fsPromises from 'fs/promises';
import path from 'path';
import os from 'os';
import { readFileContent, fileExists, writeFile, FileReadError, getConfigDir, getConfigFilePath } from '../fileReader';

// Mock os module
jest.mock('os');

const mockedOs = jest.mocked(os);

describe('File Reader', () => {
  const testFilePath = '/path/to/test/file.txt';
  const testContent = 'This is a test content\nwith multiple lines.';
  
  beforeEach(() => {
    resetVirtualFs();
  });
  
  describe('writeFile', () => {
    const testFilePath = '/path/to/test/file.txt';
    const testContent = 'Test content to write';
    
    it('should write content to a file', async () => {
      resetVirtualFs();
      
      // Call the function
      await writeFile(testFilePath, testContent);
      
      // Verify content was written correctly
      const content = await fsPromises.readFile(testFilePath, 'utf-8');
      expect(content).toBe(testContent);
    });
    
    describe('Windows-specific error handling', () => {
      beforeEach(() => {
        // Mock platform as Windows for these tests
        Object.defineProperty(process, 'platform', { value: 'win32' });
      });
      
      it('should handle Windows permission errors (EACCES)', async () => {
        resetVirtualFs();
        
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
        resetVirtualFs();
        
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
        resetVirtualFs();
        
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
        resetVirtualFs();
        
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
        // Mock platform as macOS for these tests
        Object.defineProperty(process, 'platform', { value: 'darwin' });
      });
      
      it('should handle macOS permission errors (EACCES)', async () => {
        resetVirtualFs();
        
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
        resetVirtualFs();
        
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
        resetVirtualFs();
        
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
        resetVirtualFs();
        
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
    it('should read and return file content', async () => {
      // Setup file
      resetVirtualFs();
      createVirtualFs({
        '/path/to/test/file.txt': testContent
      });
      
      // Call the function
      const result = await readFileContent(testFilePath);
      
      // Verify result
      expect(result).toBe('This is a test content with multiple lines.');
    });
    
    it('should return raw content when normalize is false', async () => {
      // Setup file
      resetVirtualFs();
      createVirtualFs({
        '/path/to/test/file.txt': testContent
      });
      
      const result = await readFileContent(testFilePath, { normalize: false });
      
      expect(result).toBe(testContent);
    });
    
    it('should resolve relative paths to absolute paths', async () => {
      // Setup file in current working directory
      const relativePath = 'relative/path.txt';
      // Create the absolute path that will be used by fileReader
      const absPath = path.resolve(process.cwd(), relativePath);
      
      // Set up file
      resetVirtualFs();
      createVirtualFs({
        [absPath]: 'Relative file content'
      });
      
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
      // Create the file
      resetVirtualFs();
      createVirtualFs({
        '/path/to/test/file.txt': testContent
      });
      
      // Set up spy to simulate permission error
      // Important: The fs module has to be mocked *each time* it's called
      const accessSpy = jest.spyOn(fsPromises, 'access');
      accessSpy.mockImplementation(() => {
        throw createFsError('EACCES', 'Permission denied', 'access', testFilePath);
      });
      
      await expect(readFileContent(testFilePath)).rejects.toThrow(FileReadError);
      await expect(readFileContent(testFilePath)).rejects.toThrow(/Permission denied/);
      
      accessSpy.mockRestore();
    });
    
    it('should throw FileReadError for other errors', async () => {
      // Create the file
      resetVirtualFs();
      createVirtualFs({
        '/path/to/test/file.txt': testContent
      });
      
      // Set up spy to simulate a generic error
      // Important: The fs module has to be mocked *each time* it's called
      const accessSpy = jest.spyOn(fsPromises, 'access');
      accessSpy.mockImplementation(() => {
        throw createFsError('EINVAL', 'Invalid argument', 'access', testFilePath);
      });
      
      await expect(readFileContent(testFilePath)).rejects.toThrow(FileReadError);
      await expect(readFileContent(testFilePath)).rejects.toThrow(/Error reading file/);
      
      accessSpy.mockRestore();
    });
  });
  
  describe('fileExists', () => {
    it('should return true when file exists', async () => {
      // Create the file
      resetVirtualFs();
      createVirtualFs({
        '/path/to/test/file.txt': testContent
      });
      
      const result = await fileExists(testFilePath);
      
      expect(result).toBe(true);
    });
    
    it('should return false when file does not exist', async () => {
      const result = await fileExists('/nonexistent/file.txt');
      
      expect(result).toBe(false);
    });
  });

  describe('getConfigDir', () => {
    // Save original environment
    const originalEnv = process.env;
    
    beforeEach(() => {
      // Reset environment variables for each test
      process.env = { ...originalEnv };
      
      // Setup homedir mock
      mockedOs.homedir.mockReturnValue('/home/user');
    });
    
    afterAll(() => {
      // Restore original environment
      process.env = originalEnv;
    });
    
    it('should use XDG_CONFIG_HOME when set', async () => {
      // Setup environment
      process.env.XDG_CONFIG_HOME = '/custom/xdg/config';
      const configDir = '/custom/xdg/config/thinktank';
      
      // Set up directory
      resetVirtualFs();
      createVirtualFs({
        '/custom/xdg/config/.placeholder': '' // Create parent directories
      });
      
      const result = await getConfigDir();
      
      expect(result).toBe(configDir);
    });
    
    describe('Windows paths', () => {
      beforeEach(() => {
        // Mock platform as Windows for all tests in this group
        Object.defineProperty(process, 'platform', { value: 'win32' });
        // Reset homedir mock for Windows tests
        mockedOs.homedir.mockReturnValue('C:\\Users\\User');
      });

      it('should use APPDATA environment variable when available', async () => {
        // Set APPDATA environment variable
        process.env.APPDATA = 'C:\\Users\\User\\AppData\\Roaming';
        const configDir = path.join('C:\\Users\\User\\AppData\\Roaming', 'thinktank');
        
        // Create directory structure with all parent directories
        createVirtualFs({
          'C:\\': '',
          'C:\\Users\\': '',
          'C:\\Users\\User\\': '',
          'C:\\Users\\User\\AppData\\': '',
          'C:\\Users\\User\\AppData\\Roaming\\': ''
        });
        
        const result = await getConfigDir();
        
        // Path.join will use the platform-specific separator
        expect(result).toBe(configDir);
      });
      
      it('should fallback to homedir/AppData/Roaming when APPDATA is not set', async () => {
        // Clear APPDATA environment variable
        delete process.env.APPDATA;
        
        const expectedPath = path.join('C:\\Users\\User', 'AppData', 'Roaming', 'thinktank');
        
        // Create directory structure with all parent directories
        createVirtualFs({
          'C:\\': '',
          'C:\\Users\\': '',
          'C:\\Users\\User\\': '',
          'C:\\Users\\User\\AppData\\': '',
          'C:\\Users\\User\\AppData\\Roaming\\': ''
        });
        
        const result = await getConfigDir();
        
        // Should construct path from homedir
        expect(result).toBe(expectedPath);
      });
      
      it('should handle Windows permission errors correctly', async () => {
        // Set APPDATA environment variable
        process.env.APPDATA = 'C:\\Users\\User\\AppData\\Roaming';
        const configDir = path.join('C:\\Users\\User\\AppData\\Roaming', 'thinktank');
        
        // Create directory structure
        createVirtualFs({
          'C:\\': '',
          'C:\\Users\\': '',
          'C:\\Users\\User\\': '',
          'C:\\Users\\User\\AppData\\': '',
          'C:\\Users\\User\\AppData\\Roaming\\': ''
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
          // Make assertions on the error
          expect(fileError).toBeInstanceOf(FileReadError);
          expect(fileError.message).toContain('Permission denied creating config directory');
          expect(fileError.message).toContain('administrative privileges');
        }
        
        mkdirSpy.mockRestore();
      });
      
      it('should handle Windows EPERM errors correctly', async () => {
        // Set APPDATA environment variable
        process.env.APPDATA = 'C:\\Users\\User\\AppData\\Roaming';
        const configDir = path.join('C:\\Users\\User\\AppData\\Roaming', 'thinktank');
        
        // Create directory structure
        createVirtualFs({
          'C:\\': '',
          'C:\\Users\\': '',
          'C:\\Users\\User\\': '',
          'C:\\Users\\User\\AppData\\': '',
          'C:\\Users\\User\\AppData\\Roaming\\': ''
        });
        
        // Simulate mkdir permission error
        const mkdirSpy = jest.spyOn(fsPromises, 'mkdir');
        mkdirSpy.mockRejectedValueOnce(
          createFsError('EPERM', 'Operation not permitted', 'mkdir', configDir)
        );
        
        try {
          await getConfigDir();
          // If we get here, the test should fail
          expect('should have thrown').toBe('but did not throw');
        } catch (err) {
          // Cast to FileReadError to make TypeScript happy
          const fileError = err as FileReadError;
          // Make assertions on the error
          expect(fileError).toBeInstanceOf(FileReadError);
          expect(fileError.message).toContain('Permission denied creating config directory');
        }
        
        mkdirSpy.mockRestore();
      });
      
      it('should handle Windows ENOENT errors correctly', async () => {
        // Set APPDATA environment variable to a non-existent path
        process.env.APPDATA = 'C:\\Invalid\\Path';
        const configDir = path.join('C:\\Invalid\\Path', 'thinktank');
        
        // Reset filesystem
        resetVirtualFs();
        
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
      
      it('should handle empty APPDATA environment variable', async () => {
        // Set APPDATA to empty string
        process.env.APPDATA = '';
        
        // Use a path that works on both Windows and Unix for testing
        const expectedPath = path.join(os.homedir(), 'AppData', 'Roaming', 'thinktank');
        
        // Create directory structure
        createVirtualFs({
          'C:\\': '',
          'C:\\Users\\': '',
          'C:\\Users\\User\\': '',
          'C:\\Users\\User\\AppData\\': '',
          'C:\\Users\\User\\AppData\\Roaming\\': ''
        });
        
        const result = await getConfigDir();
        
        // Should fallback to homedir path
        expect(result).toBe(expectedPath);
      });
    });
    
    describe('macOS paths', () => {
      beforeEach(() => {
        // Mock platform as macOS for all tests in this group
        Object.defineProperty(process, 'platform', { value: 'darwin' });
        // Reset homedir mock for macOS tests
        mockedOs.homedir.mockReturnValue('/Users/macuser');
      });

      it('should use ~/.config path on macOS for consistency with Linux', async () => {
        const configDir = '/Users/macuser/.config/thinktank';
        
        // Set up directory
        resetVirtualFs();
        createVirtualFs({
          '/Users/macuser/.config/.placeholder': '' // Create parent directories
        });
        
        // Override the platform for this test
        const originalPlatform = process.platform;
        Object.defineProperty(process, 'platform', { value: 'darwin' });
        
        try {
          const result = await getConfigDir();
          
          expect(result).toBe(configDir);
        } finally {
          // Restore the original platform
          Object.defineProperty(process, 'platform', { value: originalPlatform });
        }
      });
      
      it('should use XDG_CONFIG_HOME when set on macOS', async () => {
        // Set XDG_CONFIG_HOME environment variable
        process.env.XDG_CONFIG_HOME = '/Users/macuser/.xdg/config';
        const configDir = '/Users/macuser/.xdg/config/thinktank';
        
        // Set up directory
        resetVirtualFs();
        createVirtualFs({
          '/Users/macuser/.xdg/config/.placeholder': '' // Create parent directories
        });
        
        // Override the platform for this test
        const originalPlatform = process.platform;
        Object.defineProperty(process, 'platform', { value: 'darwin' });
        
        try {
          const result = await getConfigDir();
          
          // Should use XDG_CONFIG_HOME when explicitly set, even on macOS
          expect(result).toBe(configDir);
        } finally {
          // Restore the original platform
          Object.defineProperty(process, 'platform', { value: originalPlatform });
          // Clean up environment variable
          delete process.env.XDG_CONFIG_HOME;
        }
      });
      
      it('should throw an error when homedir is not available on macOS', async () => {
        // Mock homedir to return empty string
        mockedOs.homedir.mockReturnValue('');
        
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        await expect(getConfigDir()).rejects.toThrow('Unable to determine home directory');
      });
      
      it('should handle macOS permission errors correctly', async () => {
        const configDir = '/Users/macuser/.config/thinktank';
        
        // Reset filesystem
        resetVirtualFs();
        createVirtualFs({
          '/Users/macuser/.config/.placeholder': ''
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
        
        // Reset filesystem
        resetVirtualFs();
        createVirtualFs({
          '/Users/macuser/.placeholder': '' // Create parent directories
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
        
        // Reset filesystem
        resetVirtualFs();
        createVirtualFs({
          '/Users/macuser/.config/.placeholder': ''
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
    
    it('should use ~/.config on Linux', async () => {
      // Save original platform
      const originalPlatform = process.platform;
      
      // Mock platform as Linux
      Object.defineProperty(process, 'platform', { value: 'linux' });
      const configDir = '/home/user/.config/thinktank';
      
      // Reset filesystem
      resetVirtualFs();
      createVirtualFs({
        '/home/user/.config/.placeholder': ''
      });
      
      try {
        const result = await getConfigDir();
        
        expect(result).toBe(configDir);
      } finally {
        // Restore original platform
        Object.defineProperty(process, 'platform', { value: originalPlatform });
      }
    });
    
    it('should throw FileReadError when directory creation fails', async () => {
      // Mock directory creation failure
      const configDir = '/home/user/.config/thinktank';
      
      // Set up directory
      resetVirtualFs();
      createVirtualFs({
        '/home/user/.config/.placeholder': '' // Create parent directories
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
    it('should return the correct config file path', async () => {
      // Save original platform
      const originalPlatform = process.platform;
      
      // Mock platform as Linux
      Object.defineProperty(process, 'platform', { value: 'linux' });
      
      // Mock homedir
      mockedOs.homedir.mockReturnValue('/home/user');
      
      // Reset filesystem
      resetVirtualFs();
      createVirtualFs({
        '/home/user/.config/.placeholder': ''
      });
      
      try {
        const result = await getConfigFilePath();
        
        expect(result).toBe('/home/user/.config/thinktank/config.json');
      } finally {
        // Restore original platform
        Object.defineProperty(process, 'platform', { value: originalPlatform });
      }
    });
  });
});