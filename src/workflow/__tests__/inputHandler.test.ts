/**
 * Unit tests for the InputHandler module
 */
import { mockFsModules, resetVirtualFs, getVirtualFs, createFsError } from '../../__tests__/utils/virtualFsUtils';

// Setup mocks (must be before importing fs modules)
jest.mock('fs', () => mockFsModules().fs);
jest.mock('fs/promises', () => mockFsModules().fsPromises);

// Import fs modules after mocking
import fs from 'fs/promises';
import { processInput, InputSourceType, InputError } from '../inputHandler';
import { normalizeText } from '../../utils/helpers';

// Mock normalizeText
jest.mock('../../utils/helpers', () => ({
  normalizeText: jest.fn((text: string) => text.trim()),
}));

describe('Input Handler', () => {
  // Reset all mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
    resetVirtualFs();
    
    // Setup test files in virtual filesystem
    const virtualFs = getVirtualFs();
    virtualFs.mkdirSync('test', { recursive: true });
    virtualFs.writeFileSync('test-file.txt', '  File content from test  ');
    virtualFs.writeFileSync('protected-file.txt', 'Protected content');
    
    // Spy on fs methods for assertions
    jest.spyOn(fs, 'readFile');
    
    // Make the protected file read-only for permission error tests
    jest.spyOn(fs, 'access').mockImplementation(async (path, _mode) => {
      const pathStr = path.toString();
      if (pathStr.indexOf('protected-file.txt') >= 0) {
        throw createFsError('EACCES', 'Permission denied', 'access', pathStr);
      }
      if (pathStr.indexOf('nonexistent-file.txt') >= 0) {
        throw createFsError('ENOENT', 'File not found', 'access', pathStr);
      }
      // Use the virtual filesystem to check existence
      const virtualFs = getVirtualFs();
      if (!virtualFs.existsSync(pathStr)) {
        throw createFsError('ENOENT', 'File not found', 'access', pathStr);
      }
      return undefined;
    });
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });
  
  describe('File Input', () => {
    it('should process file input correctly', async () => {
      // Call the function
      const result = await processInput({
        input: 'test-file.txt',
        sourceType: InputSourceType.FILE,
      });
      
      // Verify results
      expect(result.content).toBe('File content from test');
      expect(result.sourceType).toBe(InputSourceType.FILE);
      expect(result.sourcePath).toContain('test-file.txt');
      expect(result.metadata.normalized).toBe(true);
      expect(result.metadata.originalLength).toBe(26);
      expect(result.metadata.finalLength).toBe(22);
      
      // Verify mocks were called correctly
      expect(fs.access).toHaveBeenCalled();
      expect(fs.readFile).toHaveBeenCalledWith(expect.stringContaining('test-file.txt'), 'utf-8');
      expect(normalizeText).toHaveBeenCalledWith('  File content from test  ');
    });
    
    it('should handle file not found errors', async () => {
      // Call the function and expect it to throw
      await expect(processInput({
        input: 'nonexistent-file.txt',
        sourceType: InputSourceType.FILE,
      })).rejects.toThrow(InputError);
      
      // Verify error details
      try {
        await processInput({
          input: 'nonexistent-file.txt',
          sourceType: InputSourceType.FILE,
        });
      } catch (e) {
        const error = e as InputError;
        expect(error.message).toContain('not found');
        expect(error.suggestions).toHaveLength(3);
        expect(error.suggestions?.[0]).toContain('Check that the file exists');
      }
    });
    
    it('should handle permission denied errors', async () => {
      // Call the function and expect it to throw
      await expect(processInput({
        input: 'protected-file.txt',
        sourceType: InputSourceType.FILE,
      })).rejects.toThrow(InputError);
      
      // Verify error details
      try {
        await processInput({
          input: 'protected-file.txt',
          sourceType: InputSourceType.FILE,
        });
      } catch (e) {
        const error = e as InputError;
        expect(error.message).toContain('Permission denied');
        expect(error.suggestions).toHaveLength(2);
        expect(error.suggestions?.[0]).toContain('permissions');
      }
    });
    
    it('should handle file read errors', async () => {
      // Create the file in the virtual filesystem so access check passes
      // Create parent directory first to avoid ENOTDIR error
      const virtualFs = getVirtualFs();
      virtualFs.mkdirSync('test', { recursive: true });
      virtualFs.writeFileSync('test/bad-file.txt', 'Bad file content');
      
      // Mock fs.readFile directly to force an error
      jest.spyOn(fs, 'readFile').mockRejectedValueOnce(
        createFsError('EIO', 'Read error', 'read', 'test/bad-file.txt')
      );
      
      // Call the function and expect it to throw
      await expect(processInput({
        input: 'test/bad-file.txt',
        sourceType: InputSourceType.FILE,
      })).rejects.toThrow(InputError);
      
      // Verify error contains original error message
      try {
        await processInput({
          input: 'test/bad-file.txt',
          sourceType: InputSourceType.FILE,
        });
      } catch (e) {
        const error = e as InputError;
        expect(error.message).toContain('Error');
        expect(error.cause).toBeDefined();
      }
    });
    
    it('should respect the normalize option', async () => {
      // Call the function with normalize = false
      const result = await processInput({
        input: 'test-file.txt',
        sourceType: InputSourceType.FILE,
        normalize: false,
      });
      
      // Verify results
      expect(result.content).toBe('  File content from test  ');
      expect(result.metadata.normalized).toBe(false);
      expect(normalizeText).not.toHaveBeenCalled();
    });
  });
  
  describe('Text Input', () => {
    it('should process direct text input correctly', async () => {
      // Call the function
      const result = await processInput({
        input: '  Direct text input  ',
        sourceType: InputSourceType.TEXT,
      });
      
      // Verify results
      expect(result.content).toBe('Direct text input');
      expect(result.sourceType).toBe(InputSourceType.TEXT);
      expect(result.sourcePath).toBeUndefined();
      expect(result.metadata.normalized).toBe(true);
      expect(result.metadata.originalLength).toBe(21);
      expect(result.metadata.finalLength).toBe(17);
      
      // Verify normalizeText was called
      expect(normalizeText).toHaveBeenCalledWith('  Direct text input  ');
    });
    
    it('should respect the normalize option for text input', async () => {
      // Call the function with normalize = false
      const result = await processInput({
        input: '  Raw text  ',
        sourceType: InputSourceType.TEXT,
        normalize: false,
      });
      
      // Verify results
      expect(result.content).toBe('  Raw text  ');
      expect(result.metadata.normalized).toBe(false);
      expect(normalizeText).not.toHaveBeenCalled();
    });
  });
  
  describe('Source Type Detection', () => {
    it('should detect file input by default', async () => {
      // Call the function without specifying source type
      const result = await processInput({
        input: 'test-file.txt',
      });
      
      // Verify the source type was detected as FILE
      expect(result.sourceType).toBe(InputSourceType.FILE);
    });
    
    it('should detect stdin input when input is "-"', async () => {
      // Mock stdin to avoid hanging the test
      const originalStdin = process.stdin;
      const mockStdin: any = {
        on: jest.fn((event, callback) => {
          if (event === 'end') {
            // Simulate immediate end with empty data
            setTimeout(() => callback(), 0);
          }
          return mockStdin;
        }),
        resume: jest.fn(),
        removeAllListeners: jest.fn(),
        pause: jest.fn(),
      };
      
      // Replace process.stdin with our mock
      Object.defineProperty(process, 'stdin', {
        value: mockStdin,
        writable: true,
      });
      
      try {
        // Call the function with "-" as input
        const resultPromise = processInput({
          input: '-',
          // Use a short timeout for testing
          stdinTimeout: 100,
        });
        
        // Verify it's using stdin
        const result = await resultPromise;
        expect(result.sourceType).toBe(InputSourceType.STDIN);
        expect(mockStdin.resume).toHaveBeenCalled();
      } finally {
        // Restore original stdin
        Object.defineProperty(process, 'stdin', {
          value: originalStdin,
          writable: true,
        });
      }
    });
    
    it('should detect text input when input has newlines', async () => {
      // Call the function with multi-line text
      const result = await processInput({
        input: 'First line\nSecond line',
      });
      
      // Verify the source type was detected as TEXT
      expect(result.sourceType).toBe(InputSourceType.TEXT);
    });
    
    it('should detect text input when input starts with a quote', async () => {
      // Call the function with quoted text
      const result = await processInput({
        input: '"This is quoted text"',
      });
      
      // Verify the source type was detected as TEXT
      expect(result.sourceType).toBe(InputSourceType.TEXT);
    });
  });
  
  describe('Error Handling', () => {
    it('should throw an error for empty input', async () => {
      // Call the function with empty input
      await expect(processInput({
        input: '',
      })).rejects.toThrow('Input is required');
    });
    
    it('should throw an error for undefined input', async () => {
      // Call the function with undefined input
      await expect(processInput({
        // @ts-expect-error - Testing undefined input
        input: undefined,
      })).rejects.toThrow('Input is required');
    });
    
    it('should throw an error for unsupported input source types', async () => {
      // Call the function with an invalid source type
      await expect(processInput({
        input: 'test',
        // @ts-expect-error - Testing invalid source type
        sourceType: 'invalid-type',
      })).rejects.toThrow('Unsupported input source type');
    });
  });
});