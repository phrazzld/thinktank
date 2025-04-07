/**
 * Tests for the context file reader utility
 */
import { createVirtualFs, resetVirtualFs, mockFsModules, createFsError } from '../../__tests__/utils/virtualFsUtils';

// Setup mocks for fs modules
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Import modules after mocking
import fs from 'fs';
import fsPromises from 'fs/promises';
import path from 'path';
import { readContextFile, MAX_FILE_SIZE, isBinaryFile } from '../fileReader';
import logger from '../logger';

// Mock logger
jest.mock('../logger');
const mockedLogger = jest.mocked(logger);

describe('readContextFile', () => {
  const testFilePath = path.join('/', 'path', 'to', 'test', 'file.txt');
  const testContent = 'This is test content\nwith multiple lines.';
  
  beforeEach(() => {
    // Reset virtual filesystem and logger mocks
    resetVirtualFs();
    jest.clearAllMocks();
  });
  
  afterEach(() => {
    jest.restoreAllMocks();
  });
  
  describe('Basic Functionality', () => {
    it('should read file content and return path and content together', async () => {
      // Setup the virtual filesystem with our test file
      createVirtualFs({
        [testFilePath]: testContent
      });
      
      const result = await readContextFile(testFilePath);
      
      expect(result).toEqual({
        path: testFilePath,
        content: testContent,
        error: null
      });
    });
    
    it('should handle relative paths by resolving them to absolute paths', async () => {
      const relativePath = 'relative/path.txt';
      const absolutePath = path.resolve(process.cwd(), relativePath);
      
      // Setup the virtual filesystem with our test file
      createVirtualFs({
        [absolutePath]: testContent
      });
      
      const result = await readContextFile(relativePath);
      
      expect(result.path).toBe(relativePath); // Original path is preserved in the result
      expect(result.content).toBe(testContent);
      expect(result.error).toBeNull();
    });

    it('should handle paths with special characters', async () => {
      const specialCharPath = path.join('/', 'path', 'with spaces and #special characters!.txt');
      
      // Setup the virtual filesystem with our test file
      createVirtualFs({
        [specialCharPath]: testContent
      });
      
      const result = await readContextFile(specialCharPath);
      
      expect(result.path).toBe(specialCharPath);
      expect(result.content).toBe(testContent);
      expect(result.error).toBeNull();
    });

    it('should handle empty files', async () => {
      // Setup the virtual filesystem with an empty file
      createVirtualFs({
        [testFilePath]: ''
      });
      
      const result = await readContextFile(testFilePath);
      
      expect(result.content).toBe('');
      expect(result.error).toBeNull();
    });
  });
  
  describe('Error Handling', () => {
    it('should return error info when file is not found', async () => {
      // Don't create the file - it should generate a not found error
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
      // Setup virtual file
      createVirtualFs({
        [testFilePath]: testContent
      });
      
      // Mock access to throw permission denied error
      const accessSpy = jest.spyOn(fsPromises, 'access');
      accessSpy.mockImplementation(() => {
        throw createFsError('EACCES', 'Permission denied', 'access', testFilePath);
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
      
      accessSpy.mockRestore();
    });
    
    it('should return error info when path is not a file', async () => {
      // Setup a directory instead of a file
      const testDirPath = path.join('/', 'path', 'to', 'test');
      createVirtualFs({
        [testDirPath]: '' // Directory
      });
      
      // Mock stat to simulate a directory
      const statSpy = jest.spyOn(fsPromises, 'stat');
      statSpy.mockResolvedValue({
        isFile: () => false,
        isDirectory: () => true,
        size: 0,
        // Include other required properties
        isBlockDevice: () => false,
        isCharacterDevice: () => false,
        isFIFO: () => false,
        isSocket: () => false,
        isSymbolicLink: () => false,
        dev: 0,
        ino: 0,
        mode: 0,
        nlink: 0,
        uid: 0,
        gid: 0,
        rdev: 0,
        blksize: 0,
        blocks: 0,
        atimeMs: 0,
        mtimeMs: 0,
        ctimeMs: 0,
        birthtimeMs: 0,
        atime: new Date(),
        mtime: new Date(),
        ctime: new Date(),
        birthtime: new Date()
      } as fs.Stats);
      
      const result = await readContextFile(testDirPath);
      
      expect(result).toEqual({
        path: testDirPath,
        content: null,
        error: {
          code: 'NOT_FILE',
          message: `Path is not a file: ${testDirPath}`
        }
      });
      
      statSpy.mockRestore();
    });
    
    it('should return error for other read errors', async () => {
      // Setup virtual file
      createVirtualFs({
        [testFilePath]: testContent
      });
      
      // Mock readFile to throw an error
      const readFileSpy = jest.spyOn(fsPromises, 'readFile');
      readFileSpy.mockImplementation(() => {
        throw new Error('Some random error');
      });
      
      const result = await readContextFile(testFilePath);
      
      expect(result).toEqual({
        path: testFilePath,
        content: null,
        error: {
          code: 'READ_ERROR',
          message: `Error reading file: ${testFilePath}`
        }
      });
      
      readFileSpy.mockRestore();
    });
    
    it('should handle unknown errors properly', async () => {
      // Setup virtual file
      createVirtualFs({
        [testFilePath]: testContent
      });
      
      // Mock access to throw a non-standard error
      const accessSpy = jest.spyOn(fsPromises, 'access');
      accessSpy.mockImplementation(() => {
        throw 'Not an error object'; // Deliberately not an Error instance
      });
      
      const result = await readContextFile(testFilePath);
      
      expect(result).toEqual({
        path: testFilePath,
        content: null,
        error: {
          code: 'UNKNOWN',
          message: `Unknown error reading file: ${testFilePath}`
        }
      });
      
      accessSpy.mockRestore();
    });

    it('should return error for file system errors during stat', async () => {
      // Setup virtual file
      createVirtualFs({
        [testFilePath]: testContent
      });
      
      // Mock stat to throw an error
      const statSpy = jest.spyOn(fsPromises, 'stat');
      statSpy.mockImplementation(() => {
        throw createFsError('EMFILE', 'Too many open files', 'stat', testFilePath);
      });
      
      const result = await readContextFile(testFilePath);
      
      expect(result.content).toBeNull();
      expect(result.error).not.toBeNull();
      expect(result.error?.code).toBe('READ_ERROR');
      
      statSpy.mockRestore();
    });
  });
  
  describe('Platform-Specific Behavior', () => {
    it('should handle Windows-style paths', async () => {
      // We need to mock path.isAbsolute for Windows paths
      const isAbsoluteSpy = jest.spyOn(path, 'isAbsolute').mockReturnValue(true);
      
      const windowsPath = 'C:\\Users\\user\\Documents\\file.txt';
      
      // Setup the virtual filesystem with our test file using the Windows path
      createVirtualFs({
        [windowsPath]: testContent
      });
      
      const result = await readContextFile(windowsPath);
      
      expect(result.path).toBe(windowsPath);
      expect(result.content).toBe(testContent);
      
      // Restore the original implementation
      isAbsoluteSpy.mockRestore();
    });
    
    it('should handle Unix-style absolute paths', async () => {
      const unixPath = path.join('/', 'Users', 'user', 'Documents', 'file.txt');
      
      // Setup the virtual filesystem with our test file
      createVirtualFs({
        [unixPath]: testContent
      });
      
      const result = await readContextFile(unixPath);
      
      expect(result.path).toBe(unixPath);
      expect(result.content).toBe(testContent);
    });
  });
  
  describe('File Size Limit', () => {
    it('should return error for files exceeding the size limit', async () => {
      // Instead of mocking stat, we'll need to create a file and then spy on stat
      // to return a large size
      createVirtualFs({
        [testFilePath]: testContent
      });
      
      // Mock stat to return a large file size
      const statSpy = jest.spyOn(fsPromises, 'stat');
      const largeSize = MAX_FILE_SIZE + 1024; // 1KB over limit
      
      statSpy.mockResolvedValue({
        isFile: () => true,
        size: largeSize,
        isDirectory: () => false,
        isBlockDevice: () => false,
        isCharacterDevice: () => false,
        isFIFO: () => false,
        isSocket: () => false,
        isSymbolicLink: () => false,
        dev: 0,
        ino: 0,
        mode: 0,
        nlink: 0,
        uid: 0,
        gid: 0,
        rdev: 0,
        blksize: 0,
        blocks: 0,
        atimeMs: 0,
        mtimeMs: 0,
        ctimeMs: 0,
        birthtimeMs: 0,
        atime: new Date(),
        mtime: new Date(),
        ctime: new Date(),
        birthtime: new Date()
      } as fs.Stats);
      
      const result = await readContextFile(testFilePath);
      
      expect(result.content).toBeNull();
      expect(result.error).not.toBeNull();
      expect(result.error?.code).toBe('FILE_TOO_LARGE');
      expect(result.error?.message).toContain('exceeds the maximum allowed size');
      expect(mockedLogger.warn).toHaveBeenCalled();
      
      statSpy.mockRestore();
    });
    
    it('should process files exactly at the size limit', async () => {
      // Create the file
      createVirtualFs({
        [testFilePath]: testContent
      });
      
      // Mock stat to return the exact size limit
      const statSpy = jest.spyOn(fsPromises, 'stat');
      statSpy.mockResolvedValue({
        isFile: () => true,
        size: MAX_FILE_SIZE,
        isDirectory: () => false,
        isBlockDevice: () => false,
        isCharacterDevice: () => false,
        isFIFO: () => false,
        isSocket: () => false,
        isSymbolicLink: () => false,
        dev: 0,
        ino: 0,
        mode: 0,
        nlink: 0,
        uid: 0,
        gid: 0,
        rdev: 0,
        blksize: 0,
        blocks: 0,
        atimeMs: 0,
        mtimeMs: 0,
        ctimeMs: 0,
        birthtimeMs: 0,
        atime: new Date(),
        mtime: new Date(),
        ctime: new Date(),
        birthtime: new Date()
      } as fs.Stats);
      
      const result = await readContextFile(testFilePath);
      
      expect(result.error).toBeNull();
      expect(result.content).toBe(testContent);
      
      statSpy.mockRestore();
    });
  });
  
  describe('Binary File Detection', () => {
    it('should return error for binary files', async () => {
      // Create a file with binary content (contains null bytes)
      const binaryContent = 'Some text with \0 null bytes \0 in it';
      createVirtualFs({
        [testFilePath]: binaryContent
      });
      
      const result = await readContextFile(testFilePath);
      
      expect(result.content).toBeNull();
      expect(result.error).not.toBeNull();
      expect(result.error?.code).toBe('BINARY_FILE');
      expect(result.error?.message).toContain('Binary file detected');
      expect(mockedLogger.warn).toHaveBeenCalled();
    });
    
    it('should directly test the isBinaryFile function', () => {
      // Test with binary content
      expect(isBinaryFile('text with \0 null byte')).toBe(true);
      
      // Test with text having many control characters
      let controlChars = '';
      for (let i = 0; i < 500; i++) {
        controlChars += String.fromCharCode(i % 32);
      }
      expect(isBinaryFile(`Some text with ${controlChars} control chars`)).toBe(true);
      
      // Test with normal text content
      expect(isBinaryFile('Normal text with no binary content')).toBe(false);
      
      // Test with empty content
      expect(isBinaryFile('')).toBe(false);
    });
  });
  
  describe('Integration Scenarios', () => {
    it('should handle files with special UTF-8 characters', async () => {
      // Create a file with UTF-8 characters
      const utf8Content = 'Unicode characters: 你好 こんにちは éàçñßø';
      createVirtualFs({
        [testFilePath]: utf8Content
      });
      
      const result = await readContextFile(testFilePath);
      
      expect(result.error).toBeNull();
      expect(result.content).toBe(utf8Content);
    });
    
    it('should handle files with mixed line endings', async () => {
      // Create a file with mixed line endings (Windows CRLF and Unix LF)
      const mixedContent = 'Line 1\r\nLine 2\nLine 3\r\nLine 4';
      createVirtualFs({
        [testFilePath]: mixedContent
      });
      
      const result = await readContextFile(testFilePath);
      
      expect(result.error).toBeNull();
      expect(result.content).toBe(mixedContent);
    });
    
    it('should handle files with trailing whitespace', async () => {
      // Create a file with trailing whitespace
      const contentWithWhitespace = 'Line with trailing space    \nAnother line\t\t';
      createVirtualFs({
        [testFilePath]: contentWithWhitespace
      });
      
      const result = await readContextFile(testFilePath);
      
      expect(result.error).toBeNull();
      expect(result.content).toBe(contentWithWhitespace);
    });
  });
});