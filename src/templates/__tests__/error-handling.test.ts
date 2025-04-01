/**
 * Unit tests for error handling across thinktank
 * 
 * This file focuses specifically on testing the ThinktankError class
 * and how errors are handled, converted, and displayed throughout the application.
 */
import { ThinktankError } from '../runThinktank';
import * as consoleUtils from '../../atoms/consoleUtils';

describe('ThinktankError', () => {
  test('should initialize correctly with message', () => {
    const error = new ThinktankError('Test error message');
    
    expect(error.message).toBe('Test error message');
    expect(error.name).toBe('ThinktankError');
    expect(error.cause).toBeUndefined();
    expect(error.category).toBeUndefined();
    expect(error.suggestions).toBeUndefined();
    expect(error.examples).toBeUndefined();
  });
  
  test('should initialize with cause error', () => {
    const cause = new Error('Original error');
    const error = new ThinktankError('Wrapped error', cause);
    
    expect(error.message).toBe('Wrapped error');
    expect(error.cause).toBe(cause);
  });
  
  test('should support adding category, suggestions and examples', () => {
    const error = new ThinktankError('Categorized error');
    error.category = consoleUtils.errorCategories.API;
    error.suggestions = ['Fix API keys', 'Check documentation'];
    error.examples = ['export API_KEY=value'];
    
    expect(error.category).toBe(consoleUtils.errorCategories.API);
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
  test('should convert from createFileNotFoundError to ThinktankError', () => {
    // Mock process.cwd
    const originalCwd = process.cwd;
    process.cwd = jest.fn().mockReturnValue('/test/dir');
    
    try {
      // Create base error from helper function
      const baseError = consoleUtils.createFileNotFoundError('missing.txt');
      
      // Convert to ThinktankError
      const thinktankError = new ThinktankError(baseError.message);
      thinktankError.category = (baseError as any).category;
      thinktankError.suggestions = (baseError as any).suggestions;
      thinktankError.examples = (baseError as any).examples;
      
      // Verify conversion preserved properties
      expect(thinktankError.category).toBe(consoleUtils.errorCategories.FILESYSTEM);
      expect(Array.isArray(thinktankError.suggestions)).toBe(true);
      expect(thinktankError.suggestions?.length).toBeGreaterThan(0);
      expect(Array.isArray(thinktankError.examples)).toBe(true);
      expect(thinktankError.examples?.length).toBeGreaterThan(0);
      
      // Check content
      expect(thinktankError.message).toContain('missing.txt');
      expect(thinktankError.suggestions?.some(s => s.includes('/test/dir'))).toBe(true);
    } finally {
      // Restore original process.cwd
      process.cwd = originalCwd;
    }
  });
  
  test('should convert from createModelFormatError to ThinktankError', () => {
    const baseError = consoleUtils.createModelFormatError('invalid');
    
    // Convert to ThinktankError
    const thinktankError = new ThinktankError(baseError.message);
    thinktankError.category = (baseError as any).category;
    thinktankError.suggestions = (baseError as any).suggestions;
    thinktankError.examples = (baseError as any).examples;
    
    // Verify conversion preserved properties
    expect(thinktankError.category).toBe(consoleUtils.errorCategories.CONFIG);
    expect(Array.isArray(thinktankError.suggestions)).toBe(true);
    expect(thinktankError.suggestions?.length).toBeGreaterThan(0);
    expect(Array.isArray(thinktankError.examples)).toBe(true);
    expect(thinktankError.examples?.length).toBeGreaterThan(0);
    
    // Check content
    expect(thinktankError.message).toContain('invalid');
    expect(thinktankError.suggestions?.some(s => s.includes('provider:modelId'))).toBe(true);
  });
  
  test('should convert from createMissingApiKeyError to ThinktankError', () => {
    const missingModels = [
      { provider: 'openai', modelId: 'gpt-4o' },
      { provider: 'anthropic', modelId: 'claude-3' }
    ];
    
    const baseError = consoleUtils.createMissingApiKeyError(missingModels);
    
    // Convert to ThinktankError
    const thinktankError = new ThinktankError(baseError.message);
    thinktankError.category = (baseError as any).category;
    thinktankError.suggestions = (baseError as any).suggestions;
    thinktankError.examples = (baseError as any).examples;
    
    // Verify conversion preserved properties
    expect(thinktankError.category).toBe(consoleUtils.errorCategories.API);
    expect(Array.isArray(thinktankError.suggestions)).toBe(true);
    expect(thinktankError.suggestions?.length).toBeGreaterThan(0);
    expect(Array.isArray(thinktankError.examples)).toBe(true);
    expect(thinktankError.examples?.length).toBeGreaterThan(0);
    
    // Check content
    expect(thinktankError.message).toContain('Missing API key');
    expect(thinktankError.suggestions?.some(s => s.includes('openai'))).toBe(true);
    expect(thinktankError.suggestions?.some(s => s.includes('anthropic'))).toBe(true);
    expect(thinktankError.suggestions?.some(s => s.includes('environment variables'))).toBe(true);
  });
});

describe('Error Propagation', () => {
  // Mock console.error
  const originalConsoleError = console.error;
  
  beforeEach(() => {
    console.error = jest.fn();
  });
  
  afterEach(() => {
    console.error = originalConsoleError;
  });
  
  // Helpers for testing error propagation
  function createMockError(): Error {
    const error = new Error('Base error');
    return error;
  }
  
  function enrichThinktankError(error: Error): ThinktankError {
    const thinktankError = new ThinktankError(`Enhanced: ${error.message}`, error);
    thinktankError.category = consoleUtils.errorCategories.API;
    thinktankError.suggestions = ['Try this', 'Or try that'];
    thinktankError.examples = ['Example command'];
    return thinktankError;
  }
  
  test('should properly propagate error as cause in ThinktankError', () => {
    const baseError = createMockError();
    const enhancedError = enrichThinktankError(baseError);
    
    expect(enhancedError.message).toBe('Enhanced: Base error');
    expect(enhancedError.cause).toBe(baseError);
    expect(enhancedError.category).toBe(consoleUtils.errorCategories.API);
  });
  
  test('should include cause when formatting the error', () => {
    const baseError = createMockError();
    const enhancedError = enrichThinktankError(baseError);
    
    // This is a simplified version of the error display logic from cli.ts
    const category = enhancedError.category ? ` (${enhancedError.category})` : '';
    const mainErrorMessage = `Error${category}: ${enhancedError.message}`;
    const causeMessage = enhancedError.cause ? `Cause: ${enhancedError.cause.message}` : '';
    
    expect(mainErrorMessage).toContain('API');
    expect(mainErrorMessage).toContain('Enhanced: Base error');
    expect(causeMessage).toContain('Base error');
  });
});