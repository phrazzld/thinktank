/**
 * Tests for file size limit checking functionality
 */
import { readContextFile } from '../fileReader';
import { 
  resetMockFs, 
  setupMockFs, 
  mockAccess, 
  mockStat, 
  mockReadFile,
  mockedFs
} from '../../__tests__/utils/mockFsUtils';

// Mock dependencies
jest.mock('fs/promises');

describe('File size limit checks', () => {
  beforeEach(() => {
    // Reset and setup mocks before each test
    resetMockFs();
    setupMockFs();
    
    // Set up default successful access for test files
    mockAccess('/path/to/large-file.txt', true);
    mockAccess('/path/to/normal-file.txt', true);
    mockAccess('/path/to/very-large-file.txt', true);
  });
  
  it('should return an error for files exceeding the size limit', async () => {
    // Mock a file that exceeds the size limit (11MB > 10MB default)
    const largeFileSize = 11 * 1024 * 1024;
    
    // Mock file stats with large size
    mockStat('/path/to/large-file.txt', {
      isFile: () => true,
      isDirectory: () => false,
      size: largeFileSize
    });
    
    // This shouldn't be called because we should fail before reading
    mockReadFile('/path/to/large-file.txt', 'Large file content');
    
    const result = await readContextFile('/path/to/large-file.txt');
    
    // Verify the error response
    expect(result.content).toBeNull();
    expect(result.error).not.toBeNull();
    expect(result.error?.code).toBe('FILE_TOO_LARGE');
    expect(result.error?.message).toContain('exceeds the maximum allowed size');
    
    // Verify that readFile was never called (we stopped before reading)
    expect(mockedFs.readFile).not.toHaveBeenCalled();
  });
  
  it('should process files within the size limit normally', async () => {
    // Mock a file that's under the size limit (5MB < 10MB default)
    const acceptableFileSize = 5 * 1024 * 1024;
    
    // Mock file stats with acceptable size
    mockStat('/path/to/normal-file.txt', {
      isFile: () => true,
      isDirectory: () => false,
      size: acceptableFileSize
    });
    
    // Mock file content
    const fileContent = 'Normal file content';
    mockReadFile('/path/to/normal-file.txt', fileContent);
    
    const result = await readContextFile('/path/to/normal-file.txt');
    
    // Verify successful result
    expect(result.error).toBeNull();
    expect(result.content).toBe(fileContent);
    
    // Verify that readFile was called (file was read)
    expect(mockedFs.readFile).toHaveBeenCalled();
  });
  
  it('should include file size and limit in the error message', async () => {
    // Mock a file that exceeds the size limit
    const largeFileSize = 15 * 1024 * 1024; // 15MB
    
    // Mock file stats
    mockStat('/path/to/very-large-file.txt', {
      isFile: () => true,
      isDirectory: () => false,
      size: largeFileSize
    });
    
    const result = await readContextFile('/path/to/very-large-file.txt');
    
    // Verify the error message contains size information
    expect(result.error?.message).toContain('15MB');
    expect(result.error?.message).toContain('10MB');
  });
});