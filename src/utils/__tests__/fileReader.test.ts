/**
 * Unit tests for file reader module
 */
import { readFileContent, fileExists, FileReadError, getConfigDir, getConfigFilePath } from '../fileReader';
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
    
    it('should use AppData on Windows', async () => {
      // Mock platform as windows
      Object.defineProperty(process, 'platform', { value: 'win32' });
      process.env.APPDATA = 'C:\\Users\\User\\AppData\\Roaming';
      
      const result = await getConfigDir();
      
      // Path.join will use the platform-specific separator, which is / in our test environment
      expect(result).toBe(path.join('C:\\Users\\User\\AppData\\Roaming', 'thinktank'));
      expect(mockedFs.mkdir).toHaveBeenCalledWith(path.join('C:\\Users\\User\\AppData\\Roaming', 'thinktank'), { recursive: true });
    });
    
    it('should use Library/Preferences on macOS', async () => {
      // Mock platform as macOS
      Object.defineProperty(process, 'platform', { value: 'darwin' });
      
      const result = await getConfigDir();
      
      expect(result).toBe('/home/user/Library/Preferences/thinktank');
      expect(mockedFs.mkdir).toHaveBeenCalledWith('/home/user/Library/Preferences/thinktank', { recursive: true });
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