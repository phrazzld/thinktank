/**
 * Unit tests for error handling across thinktank
 * 
 * This file focuses specifically on testing the ThinktankError class
 * and how errors are handled, converted, and displayed throughout the application.
 */
import { 
  ThinktankError,
  createFileNotFoundError,
  createModelFormatError,
  createMissingApiKeyError,
  errorCategories
} from '../../core/errors';

describe('ThinktankError', () => {
  test('should initialize correctly with message', () => {
    const error = new ThinktankError('Test error message');
    
    expect(error.message).toBe('Test error message');
    expect(error.name).toBe('ThinktankError');
    expect(error.cause).toBeUndefined();
    expect(error.category).toBe(errorCategories.UNKNOWN); // Default category
    expect(error.suggestions).toBeUndefined();
    expect(error.examples).toBeUndefined();
  });
  
  test('should initialize with cause error in options', () => {
    const cause = new Error('Original error');
    const error = new ThinktankError('Wrapped error', { cause });
    
    expect(error.message).toBe('Wrapped error');
    expect(error.cause).toBe(cause);
  });
  
  test('should support initializing with category, suggestions and examples', () => {
    const error = new ThinktankError('Categorized error', {
      category: errorCategories.API,
      suggestions: ['Fix API keys', 'Check documentation'],
      examples: ['export API_KEY=value']
    });
    
    expect(error.category).toBe(errorCategories.API);
    expect(error.suggestions).toEqual(['Fix API keys', 'Check documentation']);
    expect(error.examples).toEqual(['export API_KEY=value']);
  });
  
  test('should be correctly identified by instanceof', () => {
    const error = new ThinktankError('Test error');
    
    expect(error instanceof ThinktankError).toBe(true);
    expect(error instanceof Error).toBe(true);
  });
});

describe('Error Conversion', () => {
  test('should use createFileNotFoundError factory function', () => {
    // Need to assert this is a ThinktankError for TypeScript
    // Mock process.cwd
    const originalCwd = process.cwd;
    process.cwd = jest.fn().mockReturnValue('/test/dir');
    
    try {
      // Create error directly from factory function 
      // (which now returns a FileSystemError which extends ThinktankError)
      const fileNotFoundError = createFileNotFoundError('missing.txt');
      
      // Verify error has the correct type and properties
      expect(fileNotFoundError).toBeInstanceOf(ThinktankError);
      
      // Cast to ThinktankError for TypeScript
      const typedError = fileNotFoundError as ThinktankError;
      
      expect(typedError.category).toBe(errorCategories.FILESYSTEM);
      expect(Array.isArray(typedError.suggestions)).toBe(true);
      expect(typedError.suggestions?.length).toBeGreaterThan(0);
      expect(Array.isArray(typedError.examples)).toBe(true);
      expect(typedError.examples?.length).toBeGreaterThan(0);
      
      // Check content
      expect(typedError.message).toContain('missing.txt');
      expect(typedError.suggestions?.some(s => s.includes('/test/dir'))).toBe(true);
    } finally {
      // Restore original process.cwd
      process.cwd = originalCwd;
    }
  });
  
  test('should use createModelFormatError factory function', () => {
    // Create error directly from factory function
    // (which now returns a ConfigError which extends ThinktankError)
    const modelFormatError = createModelFormatError('invalid');
    
    // Verify error has the correct type and properties
    expect(modelFormatError).toBeInstanceOf(ThinktankError);
    
    // Cast to ThinktankError for TypeScript
    const typedError = modelFormatError as ThinktankError;
    
    expect(typedError.category).toBe(errorCategories.CONFIG);
    expect(Array.isArray(typedError.suggestions)).toBe(true);
    expect(typedError.suggestions?.length).toBeGreaterThan(0);
    expect(Array.isArray(typedError.examples)).toBe(true);
    expect(typedError.examples?.length).toBeGreaterThan(0);
    
    // Check content
    expect(typedError.message).toContain('invalid');
    expect(typedError.suggestions?.some(s => s.includes('provider:modelId'))).toBe(true);
  });
  
  test('should use createMissingApiKeyError factory function', () => {
    const missingModels = [
      { provider: 'openai', modelId: 'gpt-4o' },
      { provider: 'anthropic', modelId: 'claude-3' }
    ];
    
    // Create error directly from factory function
    // (which now returns an ApiError which extends ThinktankError)
    const apiKeyError = createMissingApiKeyError(missingModels);
    
    // Verify error has the correct type and properties
    expect(apiKeyError).toBeInstanceOf(ThinktankError);
    
    // Cast to ThinktankError for TypeScript
    const typedError = apiKeyError as ThinktankError;
    
    expect(typedError.category).toBe(errorCategories.API);
    expect(Array.isArray(typedError.suggestions)).toBe(true);
    expect(typedError.suggestions?.length).toBeGreaterThan(0);
    expect(Array.isArray(typedError.examples)).toBe(true);
    expect(typedError.examples?.length).toBeGreaterThan(0);
    
    // Check content
    expect(typedError.message).toContain('Missing API key');
    expect(typedError.suggestions?.some(s => s.includes('openai'))).toBe(true);
    expect(typedError.suggestions?.some(s => s.includes('anthropic'))).toBe(true);
    expect(typedError.suggestions?.some(s => s.includes('environment variables'))).toBe(true);
  });
});

describe('Error Propagation', () => {
  // Mock console.error is restored by jest.restoreAllMocks()
  
  beforeEach(() => {
    jest.spyOn(console, 'error').mockImplementation(() => {});
  });
  
  afterEach(() => {
    jest.restoreAllMocks();
  });
  
  // Helpers for testing error propagation
  function createMockError(): Error {
    const error = new Error('Base error');
    return error;
  }
  
  function enrichThinktankError(error: Error): ThinktankError {
    return new ThinktankError(`Enhanced: ${error.message}`, {
      cause: error,
      category: errorCategories.API,
      suggestions: ['Try this', 'Or try that'],
      examples: ['Example command']
    });
  }
  
  test('should properly propagate error as cause in ThinktankError', () => {
    const baseError = createMockError();
    const enhancedError = enrichThinktankError(baseError);
    
    expect(enhancedError.message).toBe('Enhanced: Base error');
    expect(enhancedError.cause).toBe(baseError);
    expect(enhancedError.category).toBe(errorCategories.API);
  });
  
  test('should include formatted output with format() method', () => {
    const baseError = createMockError();
    const enhancedError = enrichThinktankError(baseError);
    
    // Use the new format() method
    const formattedError = enhancedError.format();
    
    expect(formattedError).toContain('Error');
    expect(formattedError).toContain('API');
    expect(formattedError).toContain('Enhanced: Base error');
    expect(formattedError).toContain('Try this');
    expect(formattedError).toContain('Or try that');
    expect(formattedError).toContain('Example command');
  });
});
