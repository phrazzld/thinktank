/**
 * Unit tests for the _handleWorkflowError helper function
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

// Import spinner helper
import { createMockSpinner } from './oraTestHelper';

// Create a mock spinner
const mockSpinner = createMockSpinner();

describe('_handleWorkflowError Helper', () => {
  // Reset all mocks before each test
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset mockSpinner state
    mockSpinner.text = '';
  });

  const options = {
    input: 'test-prompt.txt'
  } as any;

  const workflowState = {
    friendlyRunName: 'clever-meadow',
    outputDirectoryPath: '/path/to/output/dir'
  } as any;

  it('should rethrow ThinktankError instances', () => {
    // Create a ThinktankError
    const error = new ThinktankError('Original error message');

    // Call the function and expect it to throw the original error
    expect(() => _handleWorkflowError({
      error,
      spinner: mockSpinner,
      options,
      workflowState
    })).toThrow(error);

    // Verify spinner.fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should handle file not found errors', () => {
    // Create a generic Error with ENOENT message
    const error = new Error('ENOENT: file not found') as NodeJS.ErrnoException;
    error.code = 'ENOENT';

    // Call the function and expect it to throw FileSystemError
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner,
        options: {
          input: 'nonexistent.txt'
        } as any,
        workflowState
      });
      fail('Should have thrown an error');
    } catch (e) {
      const err = e as FileSystemError;
      expect(err).toBeInstanceOf(FileSystemError);
      expect(err.message).toContain('File not found');
      expect(err.filePath).toBe('nonexistent.txt');
      expect(err.suggestions).toBeDefined();
      expect(err.suggestions!.length).toBeGreaterThan(0);
    }

    // Verify spinner.fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should handle permission errors', () => {
    // Create a generic Error with EACCES message
    const error = new Error('EACCES: permission denied') as NodeJS.ErrnoException;
    error.code = 'EACCES';

    // Call the function and expect it to throw PermissionError
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner,
        options,
        workflowState
      });
      fail('Should have thrown an error');
    } catch (e) {
      const err = e as PermissionError;
      expect(err).toBeInstanceOf(PermissionError);
      expect(err.message).toContain('Permission denied');
      expect(err.suggestions).toBeDefined();
      expect(err.suggestions!.some(s => s.includes('permissions'))).toBeTruthy();
    }

    // Verify spinner.fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should handle model errors', () => {
    // Create a generic Error with model-related message
    const error = new Error('Invalid model format: wrong_format');

    // Call the function and expect it to throw ConfigError
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner,
        options: {
          input: 'test-prompt.txt',
          specificModel: 'wrong_format'
        } as any,
        workflowState
      });
      fail('Should have thrown an error');
    } catch (e) {
      const err = e as ConfigError;
      expect(err).toBeInstanceOf(ConfigError);
      expect(err.message).toContain('Invalid model format');
      expect(err.suggestions).toBeDefined();
      expect(err.suggestions!.some(s => s.includes('provider:modelId'))).toBeTruthy();
      expect(err.examples).toBeDefined();
      expect(err.examples!.length).toBeGreaterThan(0);
    }

    // Verify spinner.fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should handle model not found errors', () => {
    // Create a generic Error with model not found message
    const error = new Error('Model "nonexistent:model" not found');

    // Call the function and expect it to throw ConfigError
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner,
        options: {
          input: 'test-prompt.txt',
          specificModel: 'nonexistent:model'
        } as any,
        workflowState
      });
      fail('Should have thrown an error');
    } catch (e) {
      const err = e as ConfigError;
      expect(err).toBeInstanceOf(ConfigError);
      expect(err.message).toContain('not found');
      expect(err.suggestions).toBeDefined();
      expect(err.suggestions!.some(s => s.includes('model'))).toBeTruthy();
    }

    // Verify spinner.fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should handle API key and authentication errors', () => {
    // Create a generic Error with API key related message
    const error = new Error('Invalid API key provided');

    // Call the function and expect it to throw ApiError
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner,
        options,
        workflowState
      });
      fail('Should have thrown an error');
    } catch (e) {
      const err = e as ApiError;
      expect(err).toBeInstanceOf(ApiError);
      expect(err.message).toContain('API key error');
      expect(err.suggestions).toBeDefined();
      expect(err.suggestions!.some(s => s.includes('API key'))).toBeTruthy();
    }

    // Verify spinner.fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should handle network errors', () => {
    // Create a generic Error with network related message
    const error = new Error('Network connection timed out');

    // Call the function and expect it to throw NetworkError
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner,
        options,
        workflowState
      });
      fail('Should have thrown an error');
    } catch (e) {
      const err = e as NetworkError;
      expect(err).toBeInstanceOf(NetworkError);
      expect(err.message).toContain('Network error');
      expect(err.suggestions).toBeDefined();
      expect(err.suggestions!.some(s => s.includes('connection'))).toBeTruthy();
    }

    // Verify spinner.fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should use categorizeError for generic errors', () => {
    // Create a generic Error without obvious categorization hints
    const error = new Error('Something went wrong');

    // Call the function and expect it to throw ThinktankError
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner,
        options,
        workflowState
      });
      fail('Should have thrown an error');
    } catch (e) {
      const err = e as ThinktankError;
      expect(err).toBeInstanceOf(ThinktankError);
      // The actual category will depend on the implementation of categorizeError
      expect(err.category).toBeDefined();
      expect(err.suggestions).toBeDefined();
    }

    // Verify spinner.fail was called
    expect(mockSpinner.fail).toHaveBeenCalled();
  });

  it('should add workflow context to suggestions', () => {
    // Create a basic error
    const error = new Error('Basic error');

    // Call the function
    try {
      _handleWorkflowError({
        error,
        spinner: mockSpinner,
        options,
        workflowState: {
          friendlyRunName: 'clever-meadow',
          outputDirectoryPath: '/custom/output/path'
        } as any
      });
      fail('Should have thrown an error');
    } catch (e) {
      const err = e as ThinktankError;
      if (err instanceof ThinktankError) {
        expect(err.suggestions).toBeDefined();
        expect(err.suggestions!.some(s => s.includes('clever-meadow'))).toBeTruthy();
        
        if (err.category === errorCategories.FILESYSTEM) {
          expect(err.suggestions!.some(s => s.includes('/custom/output/path'))).toBeTruthy();
        }
      }
    }
  });

  it('should handle non-Error objects', () => {
    // Create a non-Error value
    const nonError = 'Just a string error message';

    // Call the function
    try {
      _handleWorkflowError({
        error: nonError,
        spinner: mockSpinner,
        options,
        workflowState
      });
      fail('Should have thrown an error');
    } catch (e) {
      const err = e as ThinktankError;
      expect(err).toBeInstanceOf(ThinktankError);
      expect(err.message).toContain('Unknown error');
      expect(err.message).toContain('Just a string error message');
      expect(err.category).toBe(errorCategories.UNKNOWN);
      expect(err.suggestions).toBeDefined();
      expect(err.suggestions!.length).toBeGreaterThan(0);
    }
  });

  it('should handle null/undefined errors', () => {
    // Call the function with null
    try {
      _handleWorkflowError({
        error: null as any,
        spinner: mockSpinner,
        options,
        workflowState
      });
      fail('Should have thrown an error');
    } catch (e) {
      const err = e as ThinktankError;
      expect(err).toBeInstanceOf(ThinktankError);
      expect(err.message).toContain('Unknown error');
      expect(err.message).toContain('no error information');
    }

    // Call the function with undefined
    try {
      _handleWorkflowError({
        error: undefined as any,
        spinner: mockSpinner,
        options,
        workflowState
      });
      fail('Should have thrown an error');
    } catch (e) {
      const err = e as ThinktankError;
      expect(err).toBeInstanceOf(ThinktankError);
      expect(err.message).toContain('Unknown error');
      expect(err.message).toContain('no error information');
    }
  });
});