/**
 * Unit tests for file reader module
 */
import { readFileContent, fileExists, FileReadError } from '../fileReader';
import fs from 'fs/promises';
import path from 'path';

// Mock fs.promises module
jest.mock('fs/promises');
const mockedFs = jest.mocked(fs);

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
});