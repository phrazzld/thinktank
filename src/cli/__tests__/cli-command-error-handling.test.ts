/**
 * Tests for CLI command error handling
 *
 * These tests focus on validating the error objects themselves rather than the CLI command implementations.
 * Note: We're focusing initially on the run and models commands, with more to be added later.
 */
import {
  ConfigError,
  FileSystemError,
  ValidationError,
  createFileNotFoundError,
  createModelFormatError,
} from '../../core/errors';

// Store logger module path for restoration
const loggerPath = require.resolve('../../utils/logger');

// Mock logger to prevent actual console output during tests
jest.mock('../../utils/logger', () => ({
  logger: {
    info: jest.fn(),
    warn: jest.fn(),
    error: jest.fn(),
    debug: jest.fn(),
  },
}));

describe('CLI Command Error Handling', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  // Restore original logger module after all tests
  afterAll(() => {
    jest.unmock('../../utils/logger');
    delete require.cache[loggerPath];
  });

  describe('Run Command Errors', () => {
    it('formats file not found errors with appropriate suggestions and examples', () => {
      // Create a file not found error directly
      const error = createFileNotFoundError('nonexistent.txt');

      // Verify error is properly formatted
      expect(error).toBeInstanceOf(FileSystemError);
      expect(error.message).toContain('File not found');
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.length).toBeGreaterThan(0);
      expect(error.examples).toBeDefined();
      expect(error.examples?.length).toBeGreaterThan(0);
    });

    it('formats model format errors with provider suggestions', () => {
      // Create a model format error directly
      const error = createModelFormatError('invalidmodel', ['openai', 'anthropic']);

      // Verify error is properly formatted
      expect(error).toBeInstanceOf(ConfigError);
      expect(error.message).toContain('Invalid model format');
      expect(error.suggestions).toBeDefined();
      // Model specifications must use the format "provider:modelId"
      expect(error.suggestions?.some(s => s.includes('provider:modelId'))).toBe(true);
      expect(error.examples).toBeDefined();
      expect(error.examples?.length).toBeGreaterThan(0);
    });
  });

  describe('Models Command Errors', () => {
    it('properly creates validation errors for invalid provider options', () => {
      // Create a validation error directly
      const error = new ValidationError('Provider option must be a string', {
        suggestions: [
          'Specify a valid provider ID (e.g., openai, anthropic, google)',
          'Use --provider=openai format to specify the provider',
        ],
        examples: ['thinktank models --provider=openai', 'thinktank models --provider=anthropic'],
      });

      // Verify error is properly formatted
      expect(error).toBeInstanceOf(ValidationError);
      expect(error.message).toContain('Provider option must be a string');
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.some(s => s.includes('valid provider ID'))).toBe(true);
      expect(error.examples).toBeDefined();
      expect(error.examples?.some(e => e.includes('thinktank models --provider='))).toBe(true);
    });

    it('properly creates config errors for invalid providers', () => {
      // Create a config error for invalid provider
      const error = new ConfigError('Invalid provider: invalid', {
        suggestions: [
          'Check that the provider ID is correct',
          'Available providers include: openai, anthropic, google, openrouter',
          'Run without --provider to see all available models',
        ],
        examples: ['thinktank models', 'thinktank models --provider=openai'],
      });

      // Verify error is properly formatted
      expect(error).toBeInstanceOf(ConfigError);
      expect(error.message).toContain('Invalid provider');
      expect(error.suggestions).toBeDefined();
      expect(error.suggestions?.some(s => s.includes('Available providers'))).toBe(true);
      expect(error.examples).toBeDefined();
      expect(error.examples?.some(e => e.includes('thinktank models'))).toBe(true);
    });
  });
});
