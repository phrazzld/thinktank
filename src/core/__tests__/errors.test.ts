/**
 * Tests for the error handling system
 */
import { 
  ThinktankError, 
  ConfigError, 
  ApiError, 
  FileSystemError,
  ValidationError,
  NetworkError,
  PermissionError,
  InputError,
  errorCategories
} from '../errors';

describe('Error categories', () => {
  test('error categories are defined', () => {
    expect(errorCategories).toBeDefined();
    expect(errorCategories.API).toBe('API');
    expect(errorCategories.CONFIG).toBe('Configuration');
    expect(errorCategories.NETWORK).toBe('Network');
    expect(errorCategories.FILESYSTEM).toBe('File System');
    expect(errorCategories.PERMISSION).toBe('Permission');
    expect(errorCategories.VALIDATION).toBe('Validation');
    expect(errorCategories.INPUT).toBe('Input');
    expect(errorCategories.UNKNOWN).toBe('Unknown');
  });
});

describe('ThinktankError', () => {
  test('creates error with correct defaults', () => {
    const error = new ThinktankError('Test error');
    expect(error).toBeInstanceOf(Error);
    expect(error.name).toBe('ThinktankError');
    expect(error.message).toBe('Test error');
    expect(error.category).toBe(errorCategories.UNKNOWN);
    expect(error.suggestions).toBeUndefined();
    expect(error.examples).toBeUndefined();
  });

  test('accepts cause in options', () => {
    const cause = new Error('Original error');
    const error = new ThinktankError('Test error', { cause });
    expect(error.cause).toBe(cause);
  });

  test('accepts category in options', () => {
    const error = new ThinktankError('Test error', { category: errorCategories.API });
    expect(error.category).toBe(errorCategories.API);
  });

  test('accepts suggestions in options', () => {
    const suggestions = ['Try this', 'Try that'];
    const error = new ThinktankError('Test error', { suggestions });
    expect(error.suggestions).toEqual(suggestions);
  });

  test('accepts examples in options', () => {
    const examples = ['Example 1', 'Example 2'];
    const error = new ThinktankError('Test error', { examples });
    expect(error.examples).toEqual(examples);
  });

  test('format method formats error correctly', () => {
    const error = new ThinktankError('Test error', {
      category: errorCategories.API,
      suggestions: ['Try this', 'Try that'],
      examples: ['Example 1', 'Example 2']
    });
    
    const formatted = error.format();
    expect(formatted).toContain('Error (API): Test error');
    expect(formatted).toContain('Try this');
    expect(formatted).toContain('Try that');
    expect(formatted).toContain('Example 1');
    expect(formatted).toContain('Example 2');
  });
});

describe('Specialized error classes', () => {
  test('ConfigError has correct defaults', () => {
    const error = new ConfigError('Config error');
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.name).toBe('ConfigError');
    expect(error.category).toBe(errorCategories.CONFIG);
  });

  test('ApiError has correct defaults and providerId', () => {
    const error = new ApiError('API error', { providerId: 'openai' });
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.name).toBe('ApiError');
    expect(error.category).toBe(errorCategories.API);
    expect(error.providerId).toBe('openai');
    expect(error.message).toBe('[openai] API error');
  });

  test('ApiError works without providerId', () => {
    const error = new ApiError('API error');
    expect(error.providerId).toBeUndefined();
    expect(error.message).toBe('API error');
  });

  test('FileSystemError has correct defaults and filePath', () => {
    const error = new FileSystemError('File error', { filePath: '/path/to/file' });
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.name).toBe('FileSystemError');
    expect(error.category).toBe(errorCategories.FILESYSTEM);
    expect(error.filePath).toBe('/path/to/file');
  });

  test('ValidationError has correct defaults', () => {
    const error = new ValidationError('Validation error');
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.name).toBe('ValidationError');
    expect(error.category).toBe(errorCategories.VALIDATION);
  });

  test('NetworkError has correct defaults', () => {
    const error = new NetworkError('Network error');
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.name).toBe('NetworkError');
    expect(error.category).toBe(errorCategories.NETWORK);
  });

  test('PermissionError has correct defaults', () => {
    const error = new PermissionError('Permission error');
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.name).toBe('PermissionError');
    expect(error.category).toBe(errorCategories.PERMISSION);
  });

  test('InputError has correct defaults', () => {
    const error = new InputError('Input error');
    expect(error).toBeInstanceOf(ThinktankError);
    expect(error.name).toBe('InputError');
    expect(error.category).toBe(errorCategories.INPUT);
  });
});