/**
 * Tests for the context file reader utility
 */
import path from 'path';
import { readContextFile, MAX_FILE_SIZE } from '../fileReader';
import logger from '../logger';
import {
  resetMockFs,
  setupMockFs,
  mockAccess,
  mockStat,
  mockReadFile,
  mockedFs,
  createFsError
} from '../../__tests__/utils/mockFsUtils';

// Mock dependencies
jest.mock('fs/promises');
jest.mock('../logger');

const mockedLogger = jest.mocked(logger);

describe('readContextFile', () => {
  const testFilePath = '/path/to/test/file.txt';
  const testContent = 'This is test content\nwith multiple lines.';
  
  beforeEach(() => {
    // Reset mocks and set up default behavior
    resetMockFs();
    setupMockFs();
    
    // Set up default mocks for the test file
    mockAccess(testFilePath, true);
    mockReadFile(testFilePath, testContent);
    mockStat(testFilePath, {
      isFile: () => true,
      size: 1024 // 1KB file size
    });
    
    // Reset logger mocks
    jest.clearAllMocks();
  });
  
  describe('Basic Functionality', () => {
    it('should read file content and return path and content together', async () => {
      const result = await readContextFile(testFilePath);
      
      expect(result).toEqual({
        path: testFilePath,
        content: testContent,
        error: null
      });
      expect(mockedFs.access).toHaveBeenCalledWith(testFilePath, expect.any(Number));
      expect(mockedFs.readFile).toHaveBeenCalledWith(testFilePath, 'utf-8');
    });
    
    it('should handle relative paths by resolving them to absolute paths', async () => {
      const relativePath = 'relative/path.txt';
      const absolutePath = path.resolve(process.cwd(), relativePath);
      
      // Set up mocks for the resolved absolute path
      mockAccess(absolutePath, true);
      mockReadFile(absolutePath, testContent);
      mockStat(absolutePath, {
        isFile: () => true,
        size: 1024
      });
      
      const result = await readContextFile(relativePath);
      
      expect(result.path).toBe(relativePath); // Original path is preserved in the result
      expect(mockedFs.access).toHaveBeenCalledWith(absolutePath, expect.any(Number));
      expect(mockedFs.readFile).toHaveBeenCalledWith(absolutePath, 'utf-8');
    });

    it('should handle paths with special characters', async () => {
      const specialCharPath = '/path/with spaces and #special characters!.txt';
      
      // Set up mocks for the special character path
      mockAccess(specialCharPath, true);
      mockReadFile(specialCharPath, testContent);
      mockStat(specialCharPath, {
        isFile: () => true,
        size: 1024
      });
      
      const result = await readContextFile(specialCharPath);
      
      expect(result.path).toBe(specialCharPath);
      expect(mockedFs.access).toHaveBeenCalledWith(specialCharPath, expect.any(Number));
      expect(result.content).toBe(testContent);
      expect(result.error).toBeNull();
    });

    it('should handle empty files', async () => {
      // Mock an empty file
      mockReadFile(testFilePath, '');
      
      const result = await readContextFile(testFilePath);
      
      expect(result.content).toBe('');
      expect(result.error).toBeNull();
    });
  });
  
  describe('Error Handling', () => {
    it('should return error info when file is not found', async () => {
      // Mock file not found error
      mockAccess(testFilePath, false, { errorCode: 'ENOENT' });
      
      const result = await readContextFile(testFilePath);
      
      expect(result).toEqual({
        path: testFilePath,
        content: null,
        error: {
          code: 'ENOENT',
          message: `File not found: ${testFilePath}`
        }
      });
    });
    
    it('should return error info when permission is denied', async () => {
      // Mock permission denied error
      mockAccess(testFilePath, false, { 
        errorCode: 'EACCES',
        errorMessage: 'Permission denied'
      });
      
      const result = await readContextFile(testFilePath);
      
      expect(result).toEqual({
        path: testFilePath,
        content: null,
        error: {
          code: 'EACCES',
          message: `Permission denied to read file: ${testFilePath}`
        }
      });
    });
    
    it('should return error info when path is not a file', async () => {
      // Mock path being a directory, not a file
      mockStat(testFilePath, {
        isFile: () => false,
        isDirectory: () => true,
        size: 0
      });
      
      const result = await readContextFile(testFilePath);
      
      expect(result).toEqual({
        path: testFilePath,
        content: null,
        error: {
          code: 'NOT_FILE',
          message: `Path is not a file: ${testFilePath}`
        }
      });
    });
    
    it('should return error for other read errors', async () => {
      // Mock generic read error
      const readError = new Error('Some random error');
      mockReadFile(testFilePath, readError);
      
      const result = await readContextFile(testFilePath);
      
      expect(result).toEqual({
        path: testFilePath,
        content: null,
        error: {
          code: 'READ_ERROR',
          message: `Error reading file: ${testFilePath}`
        }
      });
    });
    
    it('should handle unknown errors properly', async () => {
      // For unknown errors, we need to manipulate mockedFs directly
      // since our mock utilities expect Error objects
      mockedFs.access.mockRejectedValue('Not an error object');
      
      const result = await readContextFile(testFilePath);
      
      expect(result).toEqual({
        path: testFilePath,
        content: null,
        error: {
          code: 'UNKNOWN',
          message: `Unknown error reading file: ${testFilePath}`
        }
      });
      
      // Reset access mock to prevent affecting other tests
      resetMockFs();
      setupMockFs();
      mockAccess(testFilePath, true);
    });

    it('should return error for file system errors during stat', async () => {
      // Mock fs.stat to throw an error
      const statError = createFsError('EMFILE', 'Too many open files', 'stat', testFilePath);
      mockStat(testFilePath, statError);
      
      const result = await readContextFile(testFilePath);
      
      expect(result.content).toBeNull();
      expect(result.error).not.toBeNull();
      expect(result.error?.code).toBe('READ_ERROR');
    });
  });
  
  describe('Platform-Specific Behavior', () => {
    it('should handle Windows-style paths', async () => {
      // We need to mock path.isAbsolute for Windows paths
      const isAbsoluteSpy = jest.spyOn(path, 'isAbsolute').mockReturnValue(true);
      
      const windowsPath = 'C:\\Users\\user\\Documents\\file.txt';
      
      // Set up mocks for the Windows path
      mockAccess(windowsPath, true);
      mockReadFile(windowsPath, testContent);
      mockStat(windowsPath, {
        isFile: () => true,
        size: 1024
      });
      
      const result = await readContextFile(windowsPath);
      
      expect(result.path).toBe(windowsPath);
      // Since we mocked isAbsolute to return true, the path should be used as-is
      expect(mockedFs.access).toHaveBeenCalledWith(windowsPath, expect.any(Number));
      expect(result.content).toBe(testContent);
      
      // Restore the original implementation
      isAbsoluteSpy.mockRestore();
    });
    
    it('should handle Unix-style absolute paths', async () => {
      const unixPath = '/Users/user/Documents/file.txt';
      
      // Set up mocks for the Unix path
      mockAccess(unixPath, true);
      mockReadFile(unixPath, testContent);
      mockStat(unixPath, {
        isFile: () => true,
        size: 1024
      });
      
      const result = await readContextFile(unixPath);
      
      expect(result.path).toBe(unixPath);
      expect(mockedFs.access).toHaveBeenCalledWith(unixPath, expect.any(Number));
      expect(result.content).toBe(testContent);
    });
  });
  
  describe('File Size Limit', () => {
    it('should return error for files exceeding the size limit', async () => {
      // Set file size above the limit
      const largeSize = MAX_FILE_SIZE + 1024; // 1KB over limit
      mockStat(testFilePath, {
        isFile: () => true,
        size: largeSize
      });
      
      const result = await readContextFile(testFilePath);
      
      expect(result.content).toBeNull();
      expect(result.error).not.toBeNull();
      expect(result.error?.code).toBe('FILE_TOO_LARGE');
      expect(result.error?.message).toContain('exceeds the maximum allowed size');
      expect(mockedLogger.warn).toHaveBeenCalled();
    });
    
    it('should process files exactly at the size limit', async () => {
      // Set file size exactly at the limit
      mockStat(testFilePath, {
        isFile: () => true,
        size: MAX_FILE_SIZE
      });
      
      const result = await readContextFile(testFilePath);
      
      expect(result.error).toBeNull();
      expect(result.content).toBe(testContent);
    });
  });
  
  describe('Binary File Detection', () => {
    it('should return error for binary files', async () => {
      // Mock readFile to return binary-like content with null bytes (which is a strong binary indicator)
      const binaryContent = 'Some text with \0 null bytes \0 in it';
      mockReadFile(testFilePath, binaryContent);
      
      const result = await readContextFile(testFilePath);
      
      expect(result.content).toBeNull();
      expect(result.error).not.toBeNull();
      expect(result.error?.code).toBe('BINARY_FILE');
      expect(result.error?.message).toContain('Binary file detected');
      expect(mockedLogger.warn).toHaveBeenCalled();
    });
  });
  
  describe('Integration Scenarios', () => {
    it('should handle files with special UTF-8 characters', async () => {
      // File with UTF-8 characters
      const utf8Content = 'Unicode characters: 你好 こんにちは éàçñßø';
      mockReadFile(testFilePath, utf8Content);
      
      const result = await readContextFile(testFilePath);
      
      expect(result.error).toBeNull();
      expect(result.content).toBe(utf8Content);
    });
    
    it('should handle files with mixed line endings', async () => {
      // File with mixed line endings (Windows CRLF and Unix LF)
      const mixedContent = 'Line 1\r\nLine 2\nLine 3\r\nLine 4';
      mockReadFile(testFilePath, mixedContent);
      
      const result = await readContextFile(testFilePath);
      
      expect(result.error).toBeNull();
      expect(result.content).toBe(mixedContent);
    });
    
    it('should handle files with trailing whitespace', async () => {
      // File with trailing whitespace
      const contentWithWhitespace = 'Line with trailing space    \nAnother line\t\t';
      mockReadFile(testFilePath, contentWithWhitespace);
      
      const result = await readContextFile(testFilePath);
      
      expect(result.error).toBeNull();
      expect(result.content).toBe(contentWithWhitespace);
    });
  });
});