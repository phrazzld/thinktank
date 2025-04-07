/**
 * Tests for the createContextualError utility function
 */
import { createContextualError } from '../categorization';
import { 
  ThinktankError, 
  ConfigError, 
  ApiError, 
  FileSystemError, 
  PermissionError,
  NetworkError,
  ValidationError,
  errorCategories
} from '../../index';

describe('createContextualError', () => {
  const mockContext = {
    cwd: '/test/cwd',
    input: 'test-input.txt',
    outputDirectory: '/test/output',
    specificModel: 'openai:gpt-4o',
    modelsList: ['openai:gpt-4o', 'anthropic:claude-3'],
    runName: 'test-run'
  };

  it('should properly categorize file not found errors', () => {
    // Arrange
    const originalError = new Error('ENOENT: file not found');

    // Act
    const result = createContextualError(originalError, mockContext);

    // Assert
    expect(result).toBeInstanceOf(FileSystemError);
    expect(result.message).toContain('File not found');
    expect(result.cause).toBe(originalError);
    expect(result.suggestions?.length).toBeGreaterThan(0);
    expect(result.suggestions?.[0]).toContain('Check that the file exists');
    // The test should include context information in suggestions
    expect(result.suggestions?.some(s => s.includes(mockContext.cwd))).toBeTruthy();
  });

  it('should properly categorize permission errors', () => {
    // Arrange
    const originalError = new Error('EACCES: permission denied');

    // Act
    const result = createContextualError(originalError, mockContext);

    // Assert
    expect(result).toBeInstanceOf(PermissionError);
    expect(result.message).toContain('Permission denied');
    expect(result.cause).toBe(originalError);
    expect(result.suggestions?.length).toBeGreaterThan(0);
  });

  it('should properly categorize API key errors', () => {
    // Arrange
    const originalError = new Error('Invalid API key provided');

    // Act
    const result = createContextualError(originalError, mockContext);

    // Assert
    expect(result).toBeInstanceOf(ApiError);
    expect(result.message).toContain('API key');
    expect(result.cause).toBe(originalError);
    expect(result.suggestions?.length).toBeGreaterThan(0);
    // Should include the model information in suggestions
    expect(result.suggestions?.some(s => s.includes('API key'))).toBeTruthy();
  });

  it('should properly categorize network errors', () => {
    // Arrange
    const originalError = new Error('Network connection failed');

    // Act
    const result = createContextualError(originalError, mockContext);

    // Assert
    expect(result).toBeInstanceOf(NetworkError);
    expect(result.message).toContain('Network');
    expect(result.cause).toBe(originalError);
    expect(result.suggestions?.length).toBeGreaterThan(0);
  });

  it('should properly categorize configuration errors', () => {
    // Arrange
    const originalError = new Error('Invalid configuration value');

    // Act
    const result = createContextualError(originalError, mockContext);

    // Assert
    expect(result).toBeInstanceOf(ConfigError);
    expect(result.message).toContain('Configuration');
    expect(result.cause).toBe(originalError);
    expect(result.suggestions?.length).toBeGreaterThan(0);
  });

  it('should properly categorize validation errors', () => {
    // Arrange
    const originalError = new Error('Invalid input parameter');

    // Act
    const result = createContextualError(originalError, mockContext);

    // Assert
    expect(result).toBeInstanceOf(ValidationError);
    expect(result.message).toContain('Invalid');
    expect(result.cause).toBe(originalError);
    expect(result.suggestions?.length).toBeGreaterThan(0);
  });

  it('should create ThinktankError for unknown errors', () => {
    // Arrange
    const originalError = new Error('Some unknown error');

    // Act
    const result = createContextualError(originalError, mockContext);

    // Assert
    expect(result).toBeInstanceOf(ThinktankError);
    expect(result.category).toBe(errorCategories.UNKNOWN);
    expect(result.cause).toBe(originalError);
    expect(result.suggestions?.length).toBeGreaterThan(0);
  });

  it('should correctly handle NodeJS.ErrnoException errors', () => {
    // Arrange
    const originalError = new Error('File system error');
    Object.defineProperty(originalError, 'code', { value: 'ENOENT' });

    // Act
    const result = createContextualError(originalError, mockContext);

    // Assert
    expect(result).toBeInstanceOf(FileSystemError);
    expect(result.message).toContain('File not found');
    expect(result.cause).toBe(originalError);
    expect(result.suggestions?.length).toBeGreaterThan(0);
  });

  it('should pass through ThinktankError instances', () => {
    // Arrange
    const originalError = new ThinktankError('Already categorized error', {
      category: errorCategories.API,
      suggestions: ['Original suggestion']
    });

    // Act
    const result = createContextualError(originalError, mockContext);

    // Assert
    expect(result).toBe(originalError);
    // The function should still add context to existing errors
    expect(result.suggestions?.some(s => s.includes(mockContext.runName))).toBeTruthy();
  });

  it('should handle non-Error objects', () => {
    // Arrange
    const originalError = 'just a string error';

    // Act
    const result = createContextualError(originalError, mockContext);

    // Assert
    expect(result).toBeInstanceOf(ThinktankError);
    expect(result.message).toContain('just a string error');
    expect(result.suggestions?.length).toBeGreaterThan(0);
  });
});
