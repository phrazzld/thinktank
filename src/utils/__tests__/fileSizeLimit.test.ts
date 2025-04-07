/**
 * Tests for file size limit checking functionality
 */
import { createVirtualFs, resetVirtualFs, mockFsModules } from '../../__tests__/utils/virtualFsUtils';

// Setup mocks for fs modules
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Import modules after mocking
import fs from 'fs';
import fsPromises from 'fs/promises';
import { readContextFile, MAX_FILE_SIZE } from '../fileReader';
import logger from '../logger';

// Mock logger
jest.mock('../logger');
const mockedLogger = jest.mocked(logger);

describe('File size limit checks', () => {
  const largeFilePath = '/path/to/large-file.txt';
  const normalFilePath = '/path/to/normal-file.txt';
  const veryLargeFilePath = '/path/to/very-large-file.txt';
  const fileContent = 'This is test content';
  
  beforeEach(() => {
    // Reset virtual filesystem and mocks before each test
    resetVirtualFs();
    jest.clearAllMocks();
    
    // Create virtual files for testing
    createVirtualFs({
      [largeFilePath]: fileContent,
      [normalFilePath]: fileContent,
      [veryLargeFilePath]: fileContent
    });
  });
  
  it('should return an error for files exceeding the size limit', async () => {
    // Mock a file that exceeds the size limit (11MB > 10MB default)
    const largeFileSize = 11 * 1024 * 1024;
    
    // Mock stat to return a large file size
    const statSpy = jest.spyOn(fsPromises, 'stat');
    statSpy.mockResolvedValue({
      isFile: () => true,
      size: largeFileSize,
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
    
    const result = await readContextFile(largeFilePath);
    
    // Verify the error response
    expect(result.content).toBeNull();
    expect(result.error).not.toBeNull();
    expect(result.error?.code).toBe('FILE_TOO_LARGE');
    expect(result.error?.message).toContain('exceeds the maximum allowed size');
    
    // Verify that the warning was logged
    expect(mockedLogger.warn).toHaveBeenCalled();
    
    // Check that readFile was never called (we stopped before reading)
    const readFileSpy = jest.spyOn(fsPromises, 'readFile');
    expect(readFileSpy).not.toHaveBeenCalled();
    
    statSpy.mockRestore();
  });
  
  it('should process files within the size limit normally', async () => {
    // Mock a file that's under the size limit (5MB < 10MB default)
    const acceptableFileSize = 5 * 1024 * 1024;
    
    // Mock stat to return an acceptable file size
    const statSpy = jest.spyOn(fsPromises, 'stat');
    statSpy.mockResolvedValue({
      isFile: () => true,
      size: acceptableFileSize,
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
    
    const result = await readContextFile(normalFilePath);
    
    // Verify successful result
    expect(result.error).toBeNull();
    expect(result.content).toBe(fileContent);
    
    statSpy.mockRestore();
  });
  
  it('should include file size and limit in the error message', async () => {
    // Mock a file that exceeds the size limit
    const largeFileSize = 15 * 1024 * 1024; // 15MB
    
    // Mock stat to return a large file size
    const statSpy = jest.spyOn(fsPromises, 'stat');
    statSpy.mockResolvedValue({
      isFile: () => true,
      size: largeFileSize,
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
    
    const result = await readContextFile(veryLargeFilePath);
    
    // Verify the error message contains size information
    expect(result.error?.message).toContain('15MB');
    expect(result.error?.message).toContain('10MB');
    
    statSpy.mockRestore();
  });
  
  it('should process files exactly at the size limit', async () => {
    // Mock a file that's exactly at the size limit
    const exactSizeLimit = MAX_FILE_SIZE;
    
    // Mock stat to return a file size at the limit
    const statSpy = jest.spyOn(fsPromises, 'stat');
    statSpy.mockResolvedValue({
      isFile: () => true,
      size: exactSizeLimit,
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
    
    const result = await readContextFile(normalFilePath);
    
    // Verify file at limit is processed normally
    expect(result.error).toBeNull();
    expect(result.content).toBe(fileContent);
    
    statSpy.mockRestore();
  });
});