/**
 * Tests for the context file reader utility
 */
import path from 'path';
import fs from 'fs/promises';
import { Stats } from 'fs';
import { readContextFile } from '../fileReader';

// Mock fs.promises module
jest.mock('fs/promises');

const mockedFs = jest.mocked(fs);

describe('readContextFile', () => {
  const testFilePath = '/path/to/test/file.txt';
  const testContent = 'This is test content\nwith multiple lines.';
  
  beforeEach(() => {
    jest.clearAllMocks();
    mockedFs.access.mockResolvedValue(undefined);
    mockedFs.readFile.mockResolvedValue(testContent);
    mockedFs.stat.mockResolvedValue({
      isFile: () => true,
      size: 1024, // 1KB file size
    } as Stats);
  });
  
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
    
    const result = await readContextFile(relativePath);
    
    expect(result.path).toBe(relativePath); // Original path is preserved in the result
    expect(mockedFs.access).toHaveBeenCalledWith(absolutePath, expect.any(Number));
    expect(mockedFs.readFile).toHaveBeenCalledWith(absolutePath, 'utf-8');
  });
  
  it('should return error info when file is not found', async () => {
    // Mock file not found error
    const error = new Error('File not found') as NodeJS.ErrnoException;
    error.code = 'ENOENT';
    mockedFs.access.mockRejectedValue(error);
    
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
    const error = new Error('Permission denied') as NodeJS.ErrnoException;
    error.code = 'EACCES';
    mockedFs.access.mockRejectedValue(error);
    
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
    mockedFs.stat.mockResolvedValue({
      isFile: () => false,
      size: 0,
    } as Stats);
    
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
    const error = new Error('Some random error');
    mockedFs.readFile.mockRejectedValue(error);
    
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
    // Mock non-Error object rejection
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
  });
});