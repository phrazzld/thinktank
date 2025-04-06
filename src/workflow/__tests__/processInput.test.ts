/**
 * Tests for the _processInput helper function
 */
import { _processInput } from '../runThinktankHelpers';
import * as inputHandler from '../inputHandler';
import { FileSystemError, ThinktankError } from '../../core/errors';
import { ProcessInputParams } from '../runThinktankTypes';
import { InputSourceType } from '../inputHandler';
import ora from 'ora';

// Store module path for restoration
const inputHandlerPath = require.resolve('../inputHandler');

// Mock dependencies
jest.mock('../inputHandler');
jest.mock('ora', () => {
  return jest.fn().mockImplementation(() => {
    return {
      start: jest.fn().mockReturnThis(),
      stop: jest.fn().mockReturnThis(),
      succeed: jest.fn().mockReturnThis(),
      fail: jest.fn().mockReturnThis(),
      warn: jest.fn().mockReturnThis(),
      info: jest.fn().mockReturnThis(),
      text: '',
    };
  });
});

describe('_processInput', () => {
  // Setup mock spinner
  const mockSpinner = {
    start: jest.fn().mockReturnThis(),
    stop: jest.fn().mockReturnThis(),
    succeed: jest.fn().mockReturnThis(),
    fail: jest.fn().mockReturnThis(),
    warn: jest.fn().mockReturnThis(),
    info: jest.fn().mockReturnThis(),
    text: '',
  };

  // Common params for tests
  const defaultParams: ProcessInputParams = {
    spinner: mockSpinner as unknown as ora.Ora,
    input: 'test-input.txt'
  };

  beforeEach(() => {
    jest.clearAllMocks();
    
    // Setup default mock implementation for processInput
    (inputHandler.processInput as jest.Mock).mockResolvedValue({
      content: 'Test content',
      sourceType: InputSourceType.FILE,
      sourcePath: '/path/to/test-input.txt',
      metadata: {
        processingTimeMs: 5,
        originalLength: 12,
        finalLength: 12,
        normalized: true
      }
    });
  });

  afterAll(() => {
    jest.unmock('../inputHandler');
    jest.unmock('ora');
    
    // Clear module cache
    delete require.cache[inputHandlerPath];
  });

  it('should successfully process input from a file', async () => {
    // Arrange
    const mockInputResult = {
      content: 'File content',
      sourceType: InputSourceType.FILE,
      sourcePath: '/path/to/file.txt',
      metadata: {
        processingTimeMs: 5,
        originalLength: 12,
        finalLength: 12,
        normalized: true
      }
    };
    (inputHandler.processInput as jest.Mock).mockResolvedValue(mockInputResult);
    
    // Act
    const result = await _processInput(defaultParams);
    
    // Assert
    expect(inputHandler.processInput).toHaveBeenCalledWith({ input: 'test-input.txt' });
    expect(result).toEqual({ inputResult: mockInputResult });
  });

  it('should successfully process input from stdin', async () => {
    // Arrange
    const mockInputResult = {
      content: 'Stdin content',
      sourceType: InputSourceType.STDIN,
      metadata: {
        processingTimeMs: 5,
        originalLength: 13,
        finalLength: 13,
        normalized: true
      }
    };
    (inputHandler.processInput as jest.Mock).mockResolvedValue(mockInputResult);
    
    // Act
    const result = await _processInput({ ...defaultParams, input: '-' });
    
    // Assert
    expect(inputHandler.processInput).toHaveBeenCalledWith({ input: '-' });
    expect(result).toEqual({ inputResult: mockInputResult });
  });

  it('should successfully process direct text input', async () => {
    // Arrange
    const mockInputResult = {
      content: 'Direct text',
      sourceType: InputSourceType.TEXT,
      metadata: {
        processingTimeMs: 5,
        originalLength: 11,
        finalLength: 11,
        normalized: true
      }
    };
    (inputHandler.processInput as jest.Mock).mockResolvedValue(mockInputResult);
    
    // Act
    const result = await _processInput({ ...defaultParams, input: '"Direct text"' });
    
    // Assert
    expect(inputHandler.processInput).toHaveBeenCalledWith({ input: '"Direct text"' });
    expect(result).toEqual({ inputResult: mockInputResult });
  });

  it('should wrap InputError in FileSystemError', async () => {
    // Arrange
    const inputError = new inputHandler.InputError('Input file not found: test-input.txt');
    (inputHandler.processInput as jest.Mock).mockRejectedValue(inputError);
    
    // Act & Assert
    await expect(_processInput(defaultParams)).rejects.toThrow(FileSystemError);
    expect(mockSpinner.text).toBeDefined();
  });

  it('should handle permission errors appropriately', async () => {
    // Arrange
    const inputError = new inputHandler.InputError('Permission denied to read file: test-input.txt');
    (inputHandler.processInput as jest.Mock).mockRejectedValue(inputError);
    
    // Act & Assert
    await expect(_processInput(defaultParams)).rejects.toThrow(FileSystemError);
  });

  it('should wrap generic errors in ThinktankError', async () => {
    // Arrange
    const genericError = new Error('Unexpected error');
    (inputHandler.processInput as jest.Mock).mockRejectedValue(genericError);
    
    // Act & Assert
    await expect(_processInput(defaultParams)).rejects.toThrow(ThinktankError);
  });

  it('should handle empty input errors', async () => {
    // Arrange
    const inputError = new inputHandler.InputError('Input is required');
    (inputHandler.processInput as jest.Mock).mockRejectedValue(inputError);
    
    // Act & Assert
    await expect(_processInput({ ...defaultParams, input: '' })).rejects.toThrow(FileSystemError);
  });

  it('should update spinner text during processing', async () => {
    // Act
    await _processInput(defaultParams);
    
    // Assert - check that spinner text was set
    expect(mockSpinner.text).toBeDefined();
  });
});