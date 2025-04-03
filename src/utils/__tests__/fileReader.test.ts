/**
 * Unit tests for file reader module
 */
import { readFileContent, fileExists, writeFile, FileReadError, getConfigDir, getConfigFilePath } from '../fileReader';
import fs from 'fs/promises';
import path from 'path';
import os from 'os';

// Mock fs.promises module and os
jest.mock('fs/promises');
jest.mock('os');

const mockedFs = jest.mocked(fs);
const mockedOs = jest.mocked(os);

describe('File Reader', () => {
  const testFilePath = '/path/to/test/file.txt';
  const testContent = 'This is a test content\nwith multiple lines.';
  
  beforeEach(() => {
    jest.clearAllMocks();
  });
  
  describe('writeFile', () => {
    const testFilePath = '/path/to/test/file.txt';
    const testContent = 'Test content to write';
    
    beforeEach(() => {
      mockedFs.mkdir.mockResolvedValue(undefined);
      mockedFs.writeFile.mockResolvedValue(undefined);
    });
    
    it('should write content to a file', async () => {
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
        const error = new Error('Access is denied') as NodeJS.ErrnoException;
        error.code = 'EACCES';
        mockedFs.writeFile.mockRejectedValue(error);
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow('Permission denied writing file');
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow('read-only or in use by another process');
      });
      
      it('should handle Windows permission errors (EPERM)', async () => {
        // Mock write failure with permission error
        const error = new Error('Operation not permitted') as NodeJS.ErrnoException;
        error.code = 'EPERM';
        mockedFs.writeFile.mockRejectedValue(error);
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow('Permission denied writing file');
      });
      
      it('should handle Windows "file in use" errors (EBUSY)', async () => {
        // Mock write failure with file busy error
        const error = new Error('Resource busy or locked') as NodeJS.ErrnoException;
        error.code = 'EBUSY';
        mockedFs.writeFile.mockRejectedValue(error);
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow('file is in use by another process');
      });
      
      it('should handle Windows path not found errors (ENOENT)', async () => {
        // Mock directory creation success but write failure with path not found
        mockedFs.mkdir.mockResolvedValue(undefined);
        const error = new Error('No such file or directory') as NodeJS.ErrnoException;
        error.code = 'ENOENT';
        mockedFs.writeFile.mockRejectedValue(error);
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow('directory path may not exist');
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
        const error = new Error('Permission denied') as NodeJS.ErrnoException;
        error.code = 'EACCES';
        mockedFs.writeFile.mockRejectedValue(error);
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow('Permission denied writing file');
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow('System Integrity Protection');
      });
      
      it('should handle macOS read-only filesystem errors (EROFS)', async () => {
        // Mock write failure with read-only filesystem error
        const error = new Error('Read-only file system') as NodeJS.ErrnoException;
        error.code = 'EROFS';
        mockedFs.writeFile.mockRejectedValue(error);
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow('file system is read-only');
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow('mounted with write permissions');
      });
      
      it('should handle macOS too many open files errors (EMFILE)', async () => {
        // Mock write failure with too many open files error
        const error = new Error('Too many open files') as NodeJS.ErrnoException;
        error.code = 'EMFILE';
        mockedFs.writeFile.mockRejectedValue(error);
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow('Too many open files');
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow('increase the open file limit');
      });
      
      it('should handle macOS path not found errors (ENOENT)', async () => {
        // Mock directory creation success but write failure with path not found
        mockedFs.mkdir.mockResolvedValue(undefined);
        const error = new Error('No such file or directory') as NodeJS.ErrnoException;
        error.code = 'ENOENT';
        mockedFs.writeFile.mockRejectedValue(error);
        
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow(FileReadError);
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow('parent directory may not exist');
        await expect(writeFile(testFilePath, testContent)).rejects.toThrow('on macOS');
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
      mockedFs.access.mockResolvedValue(undefined);
      mockedFs.readFile.mockResolvedValue(testContent);
      
      const result = await readFileContent(testFilePath);
      
      expect(mockedFs.access).toHaveBeenCalledWith(testFilePath, expect.any(Number));
      expect(mockedFs.readFile).toHaveBeenCalledWith(testFilePath, 'utf-8');
      expect(result).toBe('This is a test content with multiple lines.');
    });
    
    it('should return raw content when normalize is false', async () => {
      // Mock successful file read
      mockedFs.access.mockResolvedValue(undefined);
      mockedFs.readFile.mockResolvedValue(testContent);
      
      const result = await readFileContent(testFilePath, { normalize: false });
      
      expect(result).toBe(testContent);
    });
    
    it('should resolve relative paths to absolute paths', async () => {
      // Mock successful file read
      mockedFs.access.mockResolvedValue(undefined);
      mockedFs.readFile.mockResolvedValue(testContent);
      
      const relativePath = 'relative/path.txt';
      const absolutePath = path.resolve(process.cwd(), relativePath);
      
      await readFileContent(relativePath);
      
      expect(mockedFs.access).toHaveBeenCalledWith(absolutePath, expect.any(Number));
    });
    
    it('should throw FileReadError when file is not found', async () => {
      // Mock file not found error
      const error = new Error('File not found') as NodeJS.ErrnoException;
      error.code = 'ENOENT';
      mockedFs.access.mockRejectedValue(error);
      
      await expect(readFileContent(testFilePath)).rejects.toThrow(FileReadError);
      await expect(readFileContent(testFilePath)).rejects.toThrow(`File not found: ${testFilePath}`);
    });
    
    it('should throw FileReadError when permission is denied', async () => {
      // Mock permission denied error
      const error = new Error('Permission denied') as NodeJS.ErrnoException;
      error.code = 'EACCES';
      mockedFs.access.mockRejectedValue(error);
      
      await expect(readFileContent(testFilePath)).rejects.toThrow(FileReadError);
      await expect(readFileContent(testFilePath)).rejects.toThrow(`Permission denied to read file: ${testFilePath}`);
    });
    
    it('should throw FileReadError for other errors', async () => {
      // Mock generic error
      const error = new Error('Some error');
      mockedFs.access.mockRejectedValue(error);
      
      await expect(readFileContent(testFilePath)).rejects.toThrow(FileReadError);
      await expect(readFileContent(testFilePath)).rejects.toThrow(`Error reading file: ${testFilePath}`);
    });
  });
  
  describe('fileExists', () => {
    it('should return true when file exists', async () => {
      mockedFs.access.mockResolvedValue(undefined);
      
      const result = await fileExists(testFilePath);
      
      expect(result).toBe(true);
      expect(mockedFs.access).toHaveBeenCalledWith(testFilePath, expect.any(Number));
    });
    
    it('should return false when file does not exist', async () => {
      mockedFs.access.mockRejectedValue(new Error('File not found'));
      
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
      mockedFs.mkdir.mockResolvedValue(undefined);
      mockedOs.homedir.mockReturnValue('/home/user');
    });
    
    afterAll(() => {
      // Restore original environment
      process.env = originalEnv;
    });
    
    it('should use XDG_CONFIG_HOME when set', async () => {
      process.env.XDG_CONFIG_HOME = '/custom/xdg/config';
      
      const result = await getConfigDir();
      
      expect(result).toBe('/custom/xdg/config/thinktank');
      expect(mockedFs.mkdir).toHaveBeenCalledWith('/custom/xdg/config/thinktank', { recursive: true });
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
        
        const result = await getConfigDir();
        
        // Path.join will use the platform-specific separator, which is / in our test environment
        expect(result).toBe(path.join('C:\\Users\\User\\AppData\\Roaming', 'thinktank'));
        expect(mockedFs.mkdir).toHaveBeenCalledWith(path.join('C:\\Users\\User\\AppData\\Roaming', 'thinktank'), { recursive: true });
      });
      
      it('should fallback to homedir/AppData/Roaming when APPDATA is not set', async () => {
        // Clear APPDATA environment variable
        delete process.env.APPDATA;
        
        const result = await getConfigDir();
        
        // Should construct path from homedir
        const expectedPath = path.join('C:\\Users\\User', 'AppData', 'Roaming', 'thinktank');
        expect(result).toBe(expectedPath);
        expect(mockedFs.mkdir).toHaveBeenCalledWith(expectedPath, { recursive: true });
      });
      
      it('should handle Windows permission errors correctly', async () => {
        // Set APPDATA environment variable
        process.env.APPDATA = 'C:\\Users\\User\\AppData\\Roaming';
        
        // Mock directory creation failure with Windows-specific access denied error
        const error = new Error('Access is denied') as NodeJS.ErrnoException;
        error.code = 'EACCES';
        mockedFs.mkdir.mockRejectedValue(error);
        
        // Should throw a FileReadError with proper message
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        await expect(getConfigDir()).rejects.toThrow('Permission denied creating config directory');
        await expect(getConfigDir()).rejects.toThrow('administrative privileges');
      });
      
      it('should handle Windows EPERM errors correctly', async () => {
        // Set APPDATA environment variable
        process.env.APPDATA = 'C:\\Users\\User\\AppData\\Roaming';
        
        // Mock directory creation failure with Windows-specific permission error
        const error = new Error('Operation not permitted') as NodeJS.ErrnoException;
        error.code = 'EPERM';
        mockedFs.mkdir.mockRejectedValue(error);
        
        // Should throw a FileReadError with proper message
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        await expect(getConfigDir()).rejects.toThrow('Permission denied creating config directory');
      });
      
      it('should handle Windows ENOENT errors correctly', async () => {
        // Set APPDATA environment variable to a non-existent path
        process.env.APPDATA = 'C:\\Invalid\\Path';
        
        // Mock directory creation failure with Windows-specific no such file or directory error
        const error = new Error('No such file or directory') as NodeJS.ErrnoException;
        error.code = 'ENOENT';
        mockedFs.mkdir.mockRejectedValue(error);
        
        // Should throw a FileReadError with proper message
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        await expect(getConfigDir()).rejects.toThrow('Unable to create configuration directory');
        await expect(getConfigDir()).rejects.toThrow('AppData folder may not exist');
      });
      
      it('should handle empty APPDATA environment variable', async () => {
        // Set APPDATA to empty string
        process.env.APPDATA = '';
        
        const result = await getConfigDir();
        
        // Should fallback to homedir path
        const expectedPath = path.join('C:\\Users\\User', 'AppData', 'Roaming', 'thinktank');
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
        const result = await getConfigDir();
        
        expect(result).toBe('/Users/macuser/.config/thinktank');
        expect(mockedFs.mkdir).toHaveBeenCalledWith('/Users/macuser/.config/thinktank', { recursive: true });
      });
      
      it('should use XDG_CONFIG_HOME when set on macOS', async () => {
        // Set XDG_CONFIG_HOME environment variable
        process.env.XDG_CONFIG_HOME = '/Users/macuser/.xdg/config';
        
        const result = await getConfigDir();
        
        // Should use XDG_CONFIG_HOME when explicitly set, even on macOS
        expect(result).toBe('/Users/macuser/.xdg/config/thinktank');
        expect(mockedFs.mkdir).toHaveBeenCalledWith('/Users/macuser/.xdg/config/thinktank', { recursive: true });
      });
      
      it('should throw an error when homedir is not available on macOS', async () => {
        // Mock homedir to return empty string or undefined
        mockedOs.homedir.mockReturnValue('');
        
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        await expect(getConfigDir()).rejects.toThrow('Unable to determine home directory');
      });
      
      it('should handle macOS permission errors correctly', async () => {
        // Mock directory creation failure with macOS permission error
        const error = new Error('Permission denied') as NodeJS.ErrnoException;
        error.code = 'EACCES';
        mockedFs.mkdir.mockRejectedValue(error);
        
        // Should throw a FileReadError with macOS-specific message
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        await expect(getConfigDir()).rejects.toThrow('Permission denied creating config directory');
        await expect(getConfigDir()).rejects.toThrow('System Integrity Protection');
      });
      
      it('should handle macOS ENOENT errors correctly', async () => {
        // Mock directory creation failure with path not found error
        const error = new Error('No such file or directory') as NodeJS.ErrnoException;
        error.code = 'ENOENT';
        mockedFs.mkdir.mockRejectedValue(error);
        
        // Should throw a FileReadError with macOS-specific message
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        await expect(getConfigDir()).rejects.toThrow('Library/Preferences path may not exist');
      });
      
      it('should handle macOS read-only filesystem errors correctly', async () => {
        // Mock directory creation failure with read-only filesystem error
        const error = new Error('Read-only file system') as NodeJS.ErrnoException;
        error.code = 'EROFS';
        mockedFs.mkdir.mockRejectedValue(error);
        
        // Should throw a FileReadError with macOS-specific message
        await expect(getConfigDir()).rejects.toThrow(FileReadError);
        await expect(getConfigDir()).rejects.toThrow('read-only file system');
        await expect(getConfigDir()).rejects.toThrow('System Integrity Protection');
      });
      
      afterEach(() => {
        // Clean up XDG environment variable
        delete process.env.XDG_CONFIG_HOME;
      });
    });
    
    it('should use ~/.config on Linux', async () => {
      // Mock platform as Linux
      Object.defineProperty(process, 'platform', { value: 'linux' });
      
      const result = await getConfigDir();
      
      expect(result).toBe('/home/user/.config/thinktank');
      expect(mockedFs.mkdir).toHaveBeenCalledWith('/home/user/.config/thinktank', { recursive: true });
    });
    
    it('should throw FileReadError when directory creation fails', async () => {
      // Mock directory creation failure
      const error = new Error('Permission denied');
      mockedFs.mkdir.mockRejectedValue(error);
      
      await expect(getConfigDir()).rejects.toThrow(FileReadError);
      await expect(getConfigDir()).rejects.toThrow('Failed to create or access config directory');
    });
  });
  
  describe('getConfigFilePath', () => {
    it('should return the correct config file path', async () => {
      // Mock getConfigDir to return a specific path
      jest.spyOn(fs, 'mkdir').mockResolvedValue(undefined);
      mockedOs.homedir.mockReturnValue('/home/user');
      
      // Mock platform as Linux
      Object.defineProperty(process, 'platform', { value: 'linux' });
      
      const result = await getConfigFilePath();
      
      expect(result).toBe('/home/user/.config/thinktank/config.json');
    });
  });
});