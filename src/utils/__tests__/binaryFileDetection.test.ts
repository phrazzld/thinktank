/**
 * Tests for binary file detection functionality
 */
import fs from 'fs/promises';
import { Stats } from 'fs';
import { isBinaryFile, readContextFile } from '../fileReader';

// Mock dependencies
jest.mock('fs/promises');

// Access mocked functions
const mockedFs = jest.mocked(fs);

describe('Binary file detection', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Mock file system access
    mockedFs.access.mockResolvedValue(undefined);
    
    // Mock file stats
    mockedFs.stat.mockResolvedValue({
      isFile: () => true,
      isDirectory: () => false,
      size: 1024
    } as Stats);
  });
  
  describe('isBinaryFile function', () => {
    it('should detect text files correctly', async () => {
      // Text file content with normal ASCII text
      const textContent = 'This is a normal text file with some content.\nIt has multiple lines.\nAnd some punctuation!';
      
      const result = isBinaryFile(textContent);
      expect(result).toBe(false);
    });
    
    it('should detect binary files correctly', async () => {
      // Create a buffer with binary data (random bytes including many non-printable characters)
      const binaryData = Buffer.from([
        0x00, 0xFF, 0x01, 0x7F, 0x80, 0xEE, 0xAA, 0x55, 
        0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
        0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0xA0, 0xB0
      ]).toString();
      
      const result = isBinaryFile(binaryData);
      expect(result).toBe(true);
    });
    
    it('should handle UTF-8 text content as non-binary', async () => {
      // Text with UTF-8 characters
      const utf8Content = 'This contains UTF-8 characters: é à ö ñ ß ç 你好 こんにちは';
      
      const result = isBinaryFile(utf8Content);
      expect(result).toBe(false);
    });
  });
  
  describe('readContextFile integration with binary detection', () => {
    it('should return an error for binary files', async () => {
      // Mock readFile to return binary content
      const binaryData = Buffer.from([
        0x00, 0xFF, 0x01, 0x7F, 0x80, 0xEE, 0xAA, 0x55, 
        0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09
      ]).toString();
      
      mockedFs.readFile.mockResolvedValue(binaryData);
      
      const result = await readContextFile('/path/to/binary.file');
      
      expect(result.content).toBeNull();
      expect(result.error).not.toBeNull();
      expect(result.error?.code).toBe('BINARY_FILE');
      expect(result.error?.message).toContain('Binary file detected');
    });
    
    it('should process text files normally', async () => {
      // Mock readFile to return text content
      const textContent = 'This is a text file with normal content.';
      mockedFs.readFile.mockResolvedValue(textContent);
      
      const result = await readContextFile('/path/to/text.file');
      
      expect(result.error).toBeNull();
      expect(result.content).toBe(textContent);
    });
  });
});