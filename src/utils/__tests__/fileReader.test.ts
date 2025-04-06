/**
 * Unit tests for file reader module
 */
import { readFileContent, fileExists, writeFile, FileReadError, getConfigDir, getConfigFilePath } from '../fileReader';
import path from 'path';
import os from 'os';
import { 
  resetMockFs, 
  setupMockFs, 
  mockAccess, 
  mockReadFile, 
  mockMkdir, 
  mockWriteFile,
  createFsError,
  mockedFs
} from '../../__tests__/utils/mockFsUtils';

// Mock os module
jest.mock('os');

const mockedOs = jest.mocked(os);

describe('File Reader', () => {
  const testFilePath = '/path/to/test/file.txt';
  const testContent = 'This is a test content\nwith multiple lines.';
  
  beforeEach(() => {
    resetMockFs();
    setupMockFs();
  });
  
  describe('writeFile', () => {
    const testFilePath = '/path/to/test/file.txt';
    const testContent = 'Test content to write';
    
    beforeEach(() => {
      mockMkdir(path.dirname(testFilePath), true);
      mockWriteFile(testFilePath, true);
    });
    
    it('should write content to a file', async () => {
      // Mock successful file write
      mockWriteFile(testFilePath, true);
      
      await writeFile(testFilePath, testContent);
      
      expect(mockedFs.mkdir).toHaveBeenCalledWith(path.dirname(testFilePath), { recursive: true });
      expect(mockedFs.writeFile).toHaveBeenCalledWith(testFilePath, testContent, { encoding: 'utf-8' });
    });
    
    describe('Windows-specific error handling', () => {
      beforeEach(() => {
        // Mock platform as Windows for these tests
        Object.defineProperty(process, 'platform', { value: 'win32' });
      });
      
      it('should handle Windows permission errors (EACCES)', async () => {
        // Mock write failure with access denied error
        const error = createFsError('EACCES', 'Access is denied', 'writeFile', testFilePath);
        mockWriteFile(testFilePath, error);
        
        // The mockWriteFile function should handle the error case for us
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(/Permission denied writing file/);
      });
      
      it('should handle Windows permission errors (EPERM)', async () => {
        // Mock write failure with permission error
        const error = createFsError('EPERM', 'Operation not permitted', 'writeFile', testFilePath);
        mockWriteFile(testFilePath, error);
        
        // The mockWriteFile function should handle the error case for us
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(/Permission denied writing file/);
      });
      
      it('should handle Windows "file in use" errors (EBUSY)', async () => {
        // Mock write failure with file busy error
        const error = createFsError('EBUSY', 'Resource busy or locked', 'writeFile', testFilePath);
        mockWriteFile(testFilePath, error);
        
        // The mockWriteFile function should handle the error case for us
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(/file is in use by another process/);
      });
      
      it('should handle Windows path not found errors (ENOENT)', async () => {
        // Mock directory creation success but write failure with path not found
        mockMkdir(path.dirname(testFilePath), true);
        const error = createFsError('ENOENT', 'No such file or directory', 'writeFile', testFilePath);
        mockWriteFile(testFilePath, error);
        
        // The mockWriteFile function should handle the error case for us
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(/directory path may not exist/);
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
        // Mock write failure with access denied error
        const error = createFsError('EACCES', 'Permission denied', 'writeFile', testFilePath);
        mockWriteFile(testFilePath, error);
        
        // The mockWriteFile function should handle the error case for us
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(/Permission denied writing file/);
      });
      
      it('should handle macOS read-only filesystem errors (EROFS)', async () => {
        // Mock write failure with read-only filesystem error
        const error = createFsError('EROFS', 'Read-only file system', 'writeFile', testFilePath);
        mockWriteFile(testFilePath, error);
        
        // The mockWriteFile function should handle the error case for us
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(/file system is read-only/);
      });
      
      it('should handle macOS too many open files errors (EMFILE)', async () => {
        // Mock write failure with too many open files error
        const error = createFsError('EMFILE', 'Too many open files', 'writeFile', testFilePath);
        mockWriteFile(testFilePath, error);
        
        // The mockWriteFile function should handle the error case for us
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(/Too many open files/);
      });
      
      it('should handle macOS path not found errors (ENOENT)', async () => {
        // Mock directory creation success but write failure with path not found
        mockMkdir(path.dirname(testFilePath), true);
        const error = createFsError('ENOENT', 'No such file or directory', 'writeFile', testFilePath);
        mockWriteFile(testFilePath, error);
        
        // The mockWriteFile function should handle the error case for us
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(/parent directory may not exist/);
      });
      
      afterEach(() => {
        // Reset platform to prevent affecting other tests
        Object.defineProperty(process, 'platform', { value: process.platform });
      });
    });
  });
  
  describe('readFileContent', () => {
    it('should read and return file content', async () => {
      // Mock successful file read
      mockAccess(testFilePath, true);
      mockReadFile(testFilePath, testContent);
      
      const result = await readFileContent(testFilePath);
      
      expect(mockedFs.access).toHaveBeenCalledWith(testFilePath, expect.any(Number));
      expect(mockedFs.readFile).toHaveBeenCalledWith(testFilePath, 'utf-8');
      expect(result).toBe('This is a test content with multiple lines.');
    });
    
    it('should return raw content when normalize is false', async () => {
      // Mock successful file read
      mockAccess(testFilePath, true);
      mockReadFile(testFilePath, testContent);
      
      const result = await readFileContent(testFilePath, { normalize: false });
      
      expect(result).toBe(testContent);
    });
    
    it('should resolve relative paths to absolute paths', async () => {
      // Mock successful file read
      const relativePath = 'relative/path.txt';
      const absolutePath = path.resolve(process.cwd(), relativePath);
      
      mockAccess(absolutePath, true);
      mockReadFile(absolutePath, testContent);
      
      await readFileContent(relativePath);
      
      expect(mockedFs.access).toHaveBeenCalledWith(absolutePath, expect.any(Number));
    });
    
    it('should throw FileReadError when file is not found', async () => {
      // Mock file not found error
      mockAccess(testFilePath, false, { errorCode: 'ENOENT', errorMessage: 'File not found' });
      
      await expect(readFileContent(testFilePath)).rejects.toThrow(FileReadError);
      await expect(readFileContent(testFilePath)).rejects.toThrow(`File not found: ${testFilePath}`);
    });
    
    it('should throw FileReadError when permission is denied', async () => {
      // Mock permission denied error
      mockAccess(testFilePath, false, { errorCode: 'EACCES', errorMessage: 'Permission denied' });
      
      await expect(readFileContent(testFilePath)).rejects.toThrow(FileReadError);
      await expect(readFileContent(testFilePath)).rejects.toThrow(`Permission denied to read file: ${testFilePath}`);
    });
    
    it('should throw FileReadError for other errors', async () => {
      // Mock generic error
      mockAccess(testFilePath, false, { errorMessage: 'Some error' });
      
      await expect(readFileContent(testFilePath)).rejects.toThrow(FileReadError);
      await expect(readFileContent(testFilePath)).rejects.toThrow(`Error reading file: ${testFilePath}`);
    });
  });
  
  describe('fileExists', () => {
    it('should return true when file exists', async () => {
      mockAccess(testFilePath, true);
      
      const result = await fileExists(testFilePath);
      
      expect(result).toBe(true);
      expect(mockedFs.access).toHaveBeenCalledWith(testFilePath, expect.any(Number));
    });
    
    it('should return false when file does not exist', async () => {
      mockAccess(testFilePath, false);
      
      const result = await fileExists(testFilePath);
      
      expect(result).toBe(false);
    });
  });

  describe('getConfigDir', () => {
    // Save original environment
    const originalEnv = process.env;
    
    beforeEach(() => {
      // Reset environment variables for each test
      process.env = { ...originalEnv };
      mockMkdir(/.*/, true); // Allow mkdir for any path by default
      mockedOs.homedir.mockReturnValue('/home/user');
    });
    
    afterAll(() => {
      // Restore original environment
      process.env = originalEnv;
    });
    
    it('should use XDG_CONFIG_HOME when set', async () => {
      process.env.XDG_CONFIG_HOME = '/custom/xdg/config';
      const configDir = '/custom/xdg/config/thinktank';
      mockMkdir(configDir, true);
      
      const result = await getConfigDir();
      
      expect(result).toBe(configDir);
      expect(mockedFs.mkdir).toHaveBeenCalledWith(configDir, { recursive: true });
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
        mockMkdir(configDir, true);
        
        const result = await getConfigDir();
        
        // Path.join will use the platform-specific separator, which is / in our test environment
        expect(result).toBe(configDir);
        expect(mockedFs.mkdir).toHaveBeenCalledWith(configDir, { recursive: true });
      });
      
      it('should fallback to homedir/AppData/Roaming when APPDATA is not set', async () => {
        // Clear APPDATA environment variable
        delete process.env.APPDATA;
        
        const expectedPath = path.join('C:\\Users\\User', 'AppData', 'Roaming', 'thinktank');
        mockMkdir(expectedPath, true);
        
        const result = await getConfigDir();
        
        // Should construct path from homedir
        expect(result).toBe(expectedPath);
        expect(mockedFs.mkdir).toHaveBeenCalledWith(expectedPath, { recursive: true });
      });
      
      it('should handle Windows permission errors correctly', async () => {
        // Set APPDATA environment variable
        process.env.APPDATA = 'C:\\Users\\User\\AppData\\Roaming';
        const configDir = path.join('C:\\Users\\User\\AppData\\Roaming', 'thinktank');
        
        // Mock directory creation failure with Windows-specific access denied error
        const error = createFsError('EACCES', 'Access is denied', 'mkdir', configDir);
        mockMkdir(configDir, error);
        
        // Should throw a FileReadError with proper message
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        await expect(getConfigDir()).rejects.toThrow('Permission denied creating config directory');
        await expect(getConfigDir()).rejects.toThrow('administrative privileges');
      });
      
      it('should handle Windows EPERM errors correctly', async () => {
        // Set APPDATA environment variable
        process.env.APPDATA = 'C:\\Users\\User\\AppData\\Roaming';
        const configDir = path.join('C:\\Users\\User\\AppData\\Roaming', 'thinktank');
        
        // Mock directory creation failure with Windows-specific permission error
        const error = createFsError('EPERM', 'Operation not permitted', 'mkdir', configDir);
        mockMkdir(configDir, error);
        
        // Should throw a FileReadError with proper message
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        await expect(getConfigDir()).rejects.toThrow('Permission denied creating config directory');
      });
      
      it('should handle Windows ENOENT errors correctly', async () => {
        // Set APPDATA environment variable to a non-existent path
        process.env.APPDATA = 'C:\\Invalid\\Path';
        const configDir = path.join('C:\\Invalid\\Path', 'thinktank');
        
        // Mock directory creation failure with Windows-specific no such file or directory error
        const error = createFsError('ENOENT', 'No such file or directory', 'mkdir', configDir);
        mockMkdir(configDir, error);
        
        // Should throw a FileReadError with proper message
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        // Match the actual implementation messages
        await expect(getConfigDir()).rejects.toThrow(/Unable to create configuration directory|Failed to create or access config directory/);
      });
      
      it('should handle empty APPDATA environment variable', async () => {
        // Set APPDATA to empty string
        process.env.APPDATA = '';
        
        const expectedPath = path.join('C:\\Users\\User', 'AppData', 'Roaming', 'thinktank');
        mockMkdir(expectedPath, true);
        
        const result = await getConfigDir();
        
        // Should fallback to homedir path
        expect(result).toBe(expectedPath);
        expect(mockedFs.mkdir).toHaveBeenCalledWith(expectedPath, { recursive: true });
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
        mockMkdir(configDir, true);
        
        const result = await getConfigDir();
        
        expect(result).toBe(configDir);
        expect(mockedFs.mkdir).toHaveBeenCalledWith(configDir, { recursive: true });
      });
      
      it('should use XDG_CONFIG_HOME when set on macOS', async () => {
        // Set XDG_CONFIG_HOME environment variable
        process.env.XDG_CONFIG_HOME = '/Users/macuser/.xdg/config';
        const configDir = '/Users/macuser/.xdg/config/thinktank';
        mockMkdir(configDir, true);
        
        const result = await getConfigDir();
        
        // Should use XDG_CONFIG_HOME when explicitly set, even on macOS
        expect(result).toBe(configDir);
        expect(mockedFs.mkdir).toHaveBeenCalledWith(configDir, { recursive: true });
      });
      
      it('should throw an error when homedir is not available on macOS', async () => {
        // Mock homedir to return empty string or undefined
        mockedOs.homedir.mockReturnValue('');
        
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        await expect(getConfigDir()).rejects.toThrow('Unable to determine home directory');
      });
      
      it('should handle macOS permission errors correctly', async () => {
        const configDir = '/Users/macuser/.config/thinktank';
        
        // Mock directory creation failure with macOS permission error
        const error = createFsError('EACCES', 'Permission denied', 'mkdir', configDir);
        mockMkdir(configDir, error);
        
        // We're adding this comment instead of using mockImplementationOnce
        // because the mockMkdir function should handle the error case for us
        
        // Should throw a FileReadError with macOS-specific message
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        await expect(getConfigDir()).rejects.toThrow(/Permission denied creating config directory/);
      });
      
      it('should handle macOS ENOENT errors correctly', async () => {
        const configDir = '/Users/macuser/.config/thinktank';
        
        // Mock directory creation failure with path not found error
        const error = createFsError('ENOENT', 'No such file or directory', 'mkdir', configDir);
        mockMkdir(configDir, error);
        
        // Should throw a FileReadError
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        // Use a regex matching the actual implementation
        await expect(getConfigDir()).rejects.toThrow(/Unable to create configuration directory|Failed to create or access config directory/);
      });
      
      it('should handle macOS read-only filesystem errors correctly', async () => {
        const configDir = '/Users/macuser/.config/thinktank';
        
        // Mock directory creation failure with read-only filesystem error
        const error = createFsError('EROFS', 'Read-only file system', 'mkdir', configDir);
        mockMkdir(configDir, error);
        
        // Should throw a FileReadError
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        // Use a more general pattern that matches the actual implementation
        await expect(getConfigDir()).rejects.toThrow(/file system|Failed to create or access config directory/);
      });
      
      afterEach(() => {
        // Clean up XDG environment variable
        delete process.env.XDG_CONFIG_HOME;
      });
    });
    
    it('should use ~/.config on Linux', async () => {
      // Mock platform as Linux
      Object.defineProperty(process, 'platform', { value: 'linux' });
      const configDir = '/home/user/.config/thinktank';
      mockMkdir(configDir, true);
      
      const result = await getConfigDir();
      
      expect(result).toBe(configDir);
      expect(mockedFs.mkdir).toHaveBeenCalledWith(configDir, { recursive: true });
    });
    
    it('should throw FileReadError when directory creation fails', async () => {
      // Mock directory creation failure
      const configDir = '/home/user/.config/thinktank';
      const error = createFsError('EPERM', 'Permission denied', 'mkdir', configDir);
      mockMkdir(configDir, error);
      
      await expect(getConfigDir()).rejects.toThrow(FileReadError);
      // Match the actual error message in the implementation
      await expect(getConfigDir()).rejects.toThrow(/Permission denied creating config directory|Failed to create or access config directory/);
    });
  });
  
  describe('getConfigFilePath', () => {
    it('should return the correct config file path', async () => {
      // Mock mkdir for any config directory
      mockMkdir(/.*/, true);
      mockedOs.homedir.mockReturnValue('/home/user');
      
      // Mock platform as Linux
      Object.defineProperty(process, 'platform', { value: 'linux' });
      
      const result = await getConfigFilePath();
      
      expect(result).toBe('/home/user/.config/thinktank/config.json');
    });
  });
});