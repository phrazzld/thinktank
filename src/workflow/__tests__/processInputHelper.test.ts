/**
 * Unit tests for the _processInput helper function
 */
import { _processInput } from '../runThinktankHelpers';
import * as inputHandler from '../inputHandler';
import { FileSystemError, ThinktankError } from '../../core/errors';
import { InputSourceType, InputError } from '../inputHandler';

// Mock dependencies
jest.mock('../inputHandler');

// Import spinner helper
import { createMockSpinner } from './oraTestHelper';

// Create a mock spinner
const mockSpinner = createMockSpinner();

describe('_processInput Helper', () => {
  // Reset all mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset mockSpinner state
    mockSpinner.text = '';
  });

  it('should successfully process input', async () => {
    // Setup mocks
    (inputHandler.processInput as jest.Mock).mockResolvedValue({
      content: 'Test prompt content',
      sourceType: InputSourceType.FILE,
      sourcePath: '/path/to/file.txt',
      metadata: {
        processingTimeMs: 5,
        originalLength: 20,
        finalLength: 20,
        normalized: true
      }
    });

    // Call the function
    const result = await _processInput({
      spinner: mockSpinner,
      input: 'file.txt'
    });

    // Verify the result
    expect(result).toEqual({
      inputResult: {
        content: 'Test prompt content',
        sourceType: InputSourceType.FILE,
        sourcePath: '/path/to/file.txt',
        metadata: {
          processingTimeMs: 5,
          originalLength: 20,
          finalLength: 20,
          normalized: true
        }
      }
    });

    // Verify mocks were called correctly
    expect(inputHandler.processInput).toHaveBeenCalledWith({ input: 'file.txt' });

    // Verify spinner interactions
    expect(mockSpinner.text).toContain('Input processed from');
    expect(mockSpinner.text).toContain('characters');
  });

  it('should process direct text input', async () => {
    // Setup mocks
    (inputHandler.processInput as jest.Mock).mockResolvedValue({
      content: 'Direct text input',
      sourceType: InputSourceType.TEXT,
      metadata: {
        processingTimeMs: 2,
        originalLength: 16,
        finalLength: 16,
        normalized: true
      }
    });

    // Call the function
    const result = await _processInput({
      spinner: mockSpinner,
      input: 'Direct text input'
    });

    // Verify the result
    expect(result.inputResult.content).toBe('Direct text input');
    expect(result.inputResult.sourceType).toBe(InputSourceType.TEXT);

    // Verify spinner interactions
    expect(mockSpinner.text).toContain('Input processed from');
    expect(mockSpinner.text).toContain('text');
  });

  it('should process stdin input', async () => {
    // Setup mocks
    (inputHandler.processInput as jest.Mock).mockResolvedValue({
      content: 'Input from stdin',
      sourceType: InputSourceType.STDIN,
      metadata: {
        processingTimeMs: 3,
        originalLength: 15,
        finalLength: 15,
        normalized: true
      }
    });

    // Call the function
    const result = await _processInput({
      spinner: mockSpinner,
      input: '-'
    });

    // Verify the result
    expect(result.inputResult.content).toBe('Input from stdin');
    expect(result.inputResult.sourceType).toBe(InputSourceType.STDIN);

    // Verify spinner interactions
    expect(mockSpinner.text).toContain('Input processed from');
    expect(mockSpinner.text).toContain('stdin');
  });

  it('should handle file not found errors', async () => {
    // Setup mocks
    const error = new InputError('File not found: nonexistent.txt');
    error.message = 'File not found: nonexistent.txt';
    (inputHandler.processInput as jest.Mock).mockRejectedValue(error);

    // Call the function and expect it to throw
    await expect(_processInput({
      spinner: mockSpinner,
      input: 'nonexistent.txt'
    })).rejects.toThrow(FileSystemError);

    // Verify conversion to FileSystemError with correct message
    try {
      await _processInput({
        spinner: mockSpinner,
        input: 'nonexistent.txt'
      });
    } catch (error) {
      expect(error).toBeInstanceOf(FileSystemError);
      if (error instanceof FileSystemError) {
        expect(error.message).toContain('File not found');
        expect(error.filePath).toBe('nonexistent.txt');
        expect(error.suggestions).toBeDefined();
        expect(error.suggestions!.length).toBeGreaterThan(0);
      }
    }
  });

  it('should handle permission denied errors', async () => {
    // Setup mocks
    const error = new InputError('Permission denied: protected.txt');
    error.message = 'Permission denied: protected.txt';
    (inputHandler.processInput as jest.Mock).mockRejectedValue(error);

    // Call the function and expect it to throw
    await expect(_processInput({
      spinner: mockSpinner,
      input: 'protected.txt'
    })).rejects.toThrow(FileSystemError);

    // Verify FileSystemError with correct message
    try {
      await _processInput({
        spinner: mockSpinner,
        input: 'protected.txt'
      });
    } catch (error) {
      expect(error).toBeInstanceOf(FileSystemError);
      if (error instanceof FileSystemError) {
        expect(error.message).toContain('Permission denied');
        expect(error.suggestions).toBeDefined();
        expect(error.suggestions!.some(s => s.includes('permission'))).toBeTruthy();
      }
    }
  });

  it('should handle empty input errors', async () => {
    // Setup mocks
    const error = new InputError('Input is required');
    error.message = 'Input is required';
    (inputHandler.processInput as jest.Mock).mockRejectedValue(error);

    // Call the function and expect it to throw
    await expect(_processInput({
      spinner: mockSpinner,
      input: ''
    })).rejects.toThrow(FileSystemError);

    // Verify appropriate suggestions
    try {
      await _processInput({
        spinner: mockSpinner,
        input: ''
      });
    } catch (error) {
      expect(error).toBeInstanceOf(FileSystemError);
      if (error instanceof FileSystemError) {
        expect(error.message).toBe('Input is required');
        expect(error.suggestions).toBeDefined();
        expect(error.suggestions!.length).toBeGreaterThanOrEqual(2);
      }
    }
  });

  it('should handle stdin errors', async () => {
    // Setup mocks
    const error = new InputError('Stdin timeout after waiting for input');
    error.message = 'Stdin timeout after waiting for input';
    (inputHandler.processInput as jest.Mock).mockRejectedValue(error);

    // Call the function and expect it to throw
    await expect(_processInput({
      spinner: mockSpinner,
      input: '-'
    })).rejects.toThrow(FileSystemError);

    // Verify appropriate error and suggestions
    try {
      await _processInput({
        spinner: mockSpinner,
        input: '-'
      });
    } catch (error) {
      expect(error).toBeInstanceOf(FileSystemError);
      if (error instanceof FileSystemError) {
        expect(error.message).toContain('Input processing error');
        expect(error.suggestions).toBeDefined();
        expect(error.suggestions!.some(s => s.includes('stdin'))).toBeTruthy();
      }
    }
  });

  it('should handle NodeJS.ErrnoException file system errors', async () => {
    // Setup mocks
    const nodeError = new Error('EACCES: permission denied') as NodeJS.ErrnoException;
    nodeError.code = 'EACCES';
    (inputHandler.processInput as jest.Mock).mockRejectedValue(nodeError);

    // Call the function and expect it to throw
    await expect(_processInput({
      spinner: mockSpinner,
      input: 'protected.txt'
    })).rejects.toThrow(FileSystemError);

    // Verify proper error transformation
    try {
      await _processInput({
        spinner: mockSpinner,
        input: 'protected.txt'
      });
    } catch (error) {
      expect(error).toBeInstanceOf(FileSystemError);
      if (error instanceof FileSystemError) {
        expect(error.message).toContain('Permission denied');
        expect(error.filePath).toBe('protected.txt');
      }
    }
  });

  it('should handle ENOENT NodeJS errors', async () => {
    // Setup mocks
    const nodeError = new Error('ENOENT: no such file or directory') as NodeJS.ErrnoException;
    nodeError.code = 'ENOENT';
    (inputHandler.processInput as jest.Mock).mockRejectedValue(nodeError);

    // Call the function and expect it to throw
    await expect(_processInput({
      spinner: mockSpinner,
      input: 'missing.txt'
    })).rejects.toThrow(FileSystemError);

    // Verify proper error message
    try {
      await _processInput({
        spinner: mockSpinner,
        input: 'missing.txt'
      });
    } catch (error) {
      expect(error).toBeInstanceOf(FileSystemError);
      if (error instanceof FileSystemError) {
        expect(error.message).toContain('File not found');
        expect(error.filePath).toBe('missing.txt');
      }
    }
  });

  it('should handle unknown errors', async () => {
    // Setup mocks
    const unknownError = new Error('Something went terribly wrong');
    (inputHandler.processInput as jest.Mock).mockRejectedValue(unknownError);

    // Call the function and expect it to throw
    await expect(_processInput({
      spinner: mockSpinner,
      input: 'file.txt'
    })).rejects.toThrow(ThinktankError);

    // Verify proper wrapping of unknown errors
    try {
      await _processInput({
        spinner: mockSpinner,
        input: 'file.txt'
      });
    } catch (error) {
      expect(error).toBeInstanceOf(ThinktankError);
      if (error instanceof ThinktankError) {
        expect(error.message).toContain('Error processing input');
        expect(error.cause).toBe(unknownError);
      }
    }
  });
});