/**
 * Unit tests for the updated _handleWorkflowError helper function
 */
import { _handleWorkflowError } from '../runThinktankHelpers';
import { 
  ThinktankError, 
  ConfigError, 
  ApiError, 
  FileSystemError, 
  PermissionError,
  NetworkError,
  errorCategories
} from '../../core/errors';
import * as categorization from '../../core/errors/utils/categorization';

// Import spinner helper
import { createMockSpinner } from './oraTestHelper';

// Create a mock spinner
const mockSpinner = createMockSpinner();

// Mock the createContextualError function we'll create
jest.mock('../../core/errors/utils/categorization', () => {
  const actual = jest.requireActual('../../core/errors/utils/categorization');
  return {
    ...actual,
    createContextualError: jest.fn((error, context) => {
      if (error instanceof ThinktankError) {
        return error;
      }
      
      // For testing we'll return specific error types based on the error message
      const message = error instanceof Error ? error.message : String(error);
      
      if (message.includes('File not found')) {
        return new FileSystemError(message, {
          cause: error instanceof Error ? error : undefined,
          filePath: context.input,
          suggestions: [
            `Check that the file exists at the specified path`,
            `Current working directory: ${context.cwd}`
          ]
        });
      } else if (message.includes('Permission')) {
        return new PermissionError(message, {
          cause: error instanceof Error ? error : undefined,
          suggestions: [
            'Check file permissions',
            'Try using a different location'
          ]
        });
      } else if (message.includes('API key')) {
        return new ApiError(message, {
          cause: error instanceof Error ? error : undefined,
          suggestions: [
            'Check that you have set the correct environment variables',
            'Verify your API key is valid'
          ]
        });
      } else if (message.includes('Network')) {
        return new NetworkError(message, {
          cause: error instanceof Error ? error : undefined,
          suggestions: [
            'Check your internet connection',
            'Verify the service is online'
          ]
        });
      } else if (message.includes('Config')) {
        return new ConfigError(message, {
          cause: error instanceof Error ? error : undefined,
          suggestions: [
            'Check your configuration file',
            'Verify the configuration values'
          ]
        });
      }
      
      // Default to ThinktankError for unknown cases
      return new ThinktankError(`Unknown error: ${message}`, {
        cause: error instanceof Error ? error : undefined,
        category: errorCategories.UNKNOWN,
        suggestions: [
          'This is an unexpected error',
          'Please report this issue'
        ]
      });
    })
  };
});

describe('_handleWorkflowError Helper (Updated)', () => {
  // Reset all mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset mockSpinner state
    mockSpinner.text = '';
    // Reset the mock implementation
    (categorization.createContextualError as jest.Mock).mockClear();
  });

  const options = {
    input: 'test-prompt.txt'
  };

  it('should call createContextualError with correct context', () => {
    // Arrange
    const error = new Error('File not found: test-prompt.txt');
    const state = {
      outputDirectoryPath: '/test/output',
      friendlyRunName: 'test-run'
    };
    
    // Act - this will throw, so wrap in try-catch
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner as any,
        options,
        workflowState: state
      });
    } catch (e) {
      // We expect this to throw, ignore
    }
    
    // Assert
    expect(categorization.createContextualError).toHaveBeenCalledWith(
      error,
      expect.objectContaining({
        cwd: expect.any(String),
        input: 'test-prompt.txt',
        outputDirectory: '/test/output',
        runName: 'test-run'
      })
    );
  });

  it('should handle file not found errors', () => {
    // Arrange
    const error = new Error('File not found: nonexistent.txt');
    
    // Act & Assert
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner as any,
        options: { input: 'nonexistent.txt' },
        workflowState: {}
      });
      fail('Function should throw an error');
    } catch (err) {
      if (!(err instanceof FileSystemError)) {
        fail(`Expected FileSystemError but got ${String(err)}`);
      }
      expect(err).toBeInstanceOf(FileSystemError);
      expect(err.message).toContain('File not found');
      expect(err.filePath).toBe('nonexistent.txt');
      expect(err.suggestions).toBeDefined();
      expect(err.suggestions!.length).toBeGreaterThan(0);
    }
    
    // Check that spinner fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should handle permission errors', () => {
    // Arrange
    const error = new Error('Permission denied when writing file');
    
    // Act & Assert
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner as any,
        options,
        workflowState: {
          outputDirectoryPath: '/test/output'
        }
      });
      fail('Function should throw an error');
    } catch (err) {
      if (!(err instanceof PermissionError)) {
        fail(`Expected PermissionError but got ${String(err)}`);
      }
      expect(err).toBeInstanceOf(PermissionError);
      expect(err.message).toContain('Permission denied');
      expect(err.suggestions).toBeDefined();
      expect(err.suggestions!.length).toBeGreaterThan(0);
    }
    
    // Check that spinner fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should handle API key errors', () => {
    // Arrange
    const error = new Error('Invalid API key provided');
    
    // Act & Assert
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner as any,
        options: {
          ...options,
          specificModel: 'openai:gpt-4o'
        },
        workflowState: {}
      });
      fail('Function should throw an error');
    } catch (err) {
      if (!(err instanceof ApiError)) {
        fail(`Expected ApiError but got ${String(err)}`);
      }
      expect(err).toBeInstanceOf(ApiError);
      expect(err.message).toContain('API key');
      expect(err.suggestions).toBeDefined();
      expect(err.suggestions!.length).toBeGreaterThan(0);
    }
    
    // Check that spinner fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should handle network errors', () => {
    // Arrange
    const error = new Error('Network connection failed');
    
    // Act & Assert
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner as any,
        options,
        workflowState: {}
      });
      fail('Function should throw an error');
    } catch (err) {
      if (!(err instanceof NetworkError)) {
        fail(`Expected NetworkError but got ${String(err)}`);
      }
      expect(err).toBeInstanceOf(NetworkError);
      expect(err.message).toContain('Network');
      expect(err.suggestions).toBeDefined();
      expect(err.suggestions!.length).toBeGreaterThan(0);
    }
    
    // Check that spinner fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should handle configuration errors', () => {
    // Arrange
    const error = new Error('Config error: Invalid model format');
    
    // Act & Assert
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner as any,
        options,
        workflowState: {}
      });
      fail('Function should throw an error');
    } catch (err) {
      if (!(err instanceof ConfigError)) {
        fail(`Expected ConfigError but got ${String(err)}`);
      }
      expect(err).toBeInstanceOf(ConfigError);
      expect(err.message).toContain('Config');
      expect(err.suggestions).toBeDefined();
      expect(err.suggestions!.length).toBeGreaterThan(0);
    }
    
    // Check that spinner fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should handle unknown errors', () => {
    // Arrange
    const error = new Error('Some unexpected error');
    
    // Act & Assert
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner as any,
        options,
        workflowState: {}
      });
      fail('Function should throw an error');
    } catch (err) {
      if (!(err instanceof ThinktankError)) {
        fail(`Expected ThinktankError but got ${String(err)}`);
      }
      expect(err).toBeInstanceOf(ThinktankError);
      expect(err.category).toBe(errorCategories.UNKNOWN);
      expect(err.suggestions).toBeDefined();
      expect(err.suggestions!.length).toBeGreaterThan(0);
    }
    
    // Check that spinner fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should pass through existing ThinktankErrors', () => {
    // Arrange
    const originalError = new ApiError('API Rate limit exceeded', {
      suggestions: ['Wait and try again later']
    });
    
    // Act & Assert
    try {
      _handleWorkflowError({
        error: originalError,
        spinner: mockSpinner as any,
        options,
        workflowState: {
          friendlyRunName: 'test-run'
        }
      });
      fail('Function should throw an error');
    } catch (err) {
      if (!(err instanceof ThinktankError)) {
        fail(`Expected ThinktankError but got ${err}`);
      }
      expect(err).toBe(originalError); // Should be the same instance
      expect(err.suggestions).toContain('Wait and try again later');
    }
    
    // Check that spinner fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should handle non-Error objects', () => {
    // Arrange
    const error = 'just a string error';
    
    // Act & Assert
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner as any,
        options,
        workflowState: {}
      });
      fail('Function should throw an error');
    } catch (err) {
      if (!(err instanceof ThinktankError)) {
        fail(`Expected ThinktankError but got ${err}`);
      }
      expect(err).toBeInstanceOf(ThinktankError);
      expect(err.message).toContain('just a string error');
      expect(err.suggestions).toBeDefined();
      expect(err.suggestions!.length).toBeGreaterThan(0);
    }
    
    // Check that spinner fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });
});