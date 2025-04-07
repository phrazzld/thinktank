/**
 * Tests for the context file reader utility using centralized mock setup
 */

// Import modules after mocking has been set up by centralized setup
import fs from 'fs';
import fsPromises from 'fs/promises';
import path from 'path';
import { readContextFile, isBinaryFile } from '../fileReader';
import logger from '../logger';

// Import centralized setup helpers
import { setupBasicFs, createFsError } from '../../../jest/setupFiles/fs';

// Import direct utils (still available when needed)
import { resetVirtualFs } from '../../__tests__/utils/virtualFsUtils';

// Mock logger
jest.mock('../logger');
const mockedLogger = jest.mocked(logger);

describe('readContextFile with Centralized Mocks', () => {
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
      // Setup the virtual filesystem with our test file using the central helper
      setupBasicFs({
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
      setupBasicFs({
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
      setupBasicFs({
        [specialCharPath]: testContent
      });
      
      const result = await readContextFile(specialCharPath);
      
      expect(result.path).toBe(specialCharPath);
      expect(result.content).toBe(testContent);
      expect(result.error).toBeNull();
    });

    it('should handle empty files', async () => {
      // Setup the virtual filesystem with an empty file
      setupBasicFs({
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
      setupBasicFs({}); // Empty filesystem
      
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
      setupBasicFs({
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
      setupBasicFs({
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
  });
  
  describe('Binary File Detection', () => {
    it('should return error for binary files', async () => {
      // Create a file with binary content (contains null bytes)
      const binaryContent = 'Some text with \0 null bytes \0 in it';
      setupBasicFs({
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
      setupBasicFs({
        [testFilePath]: utf8Content
      });
      
      const result = await readContextFile(testFilePath);
      
      expect(result.error).toBeNull();
      expect(result.content).toBe(utf8Content);
    });
  });
});
