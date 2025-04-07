/**
 * Tests for CLI error handling
 * 
 * These tests verify that the CLI properly handles and displays
 * various types of errors with helpful, actionable information.
 */
import { 
  ThinktankError, 
  ApiError, 
  ConfigError,
  FileSystemError,
  PermissionError,
  errorCategories,
  createFileNotFoundError,
  createModelFormatError,
  createMissingApiKeyError
} from '../../core/errors';
// Import handleError function from our test helpers
// We're using a separate file just for testing to avoid CLI initialization
// This is needed to prevent the CLI from running after tests complete
import { handleError } from './test-helpers/cli-handlers';
import { logger } from '../../utils/logger';

// Store logger module path for restoration
const loggerPath = require.resolve('../../utils/logger');

// Mock the logger to prevent actual console output during tests
jest.mock('../../utils/logger', () => ({
  logger: {
    error: jest.fn(),
    debug: jest.fn(),
    info: jest.fn(),
    warn: jest.fn()
  }
}));

// This file tests error formatting for CLI display
describe('CLI Error Handling', () => {
  // Create a spy for process.exit
  const originalExit = process.exit;
  
  beforeEach(() => {
    jest.clearAllMocks();
    // Mock process.exit
    process.exit = jest.fn() as unknown as (code?: number) => never;
  });
  
  afterEach(() => {
    // Restore original process.exit
    process.exit = originalExit;
  });
  
  // Restore original logger module after all tests
  afterAll(() => {
    jest.unmock('../../utils/logger');
    delete require.cache[loggerPath];
  });
  
  test('ThinktankError with detailed information should display correctly', () => {
    // Create an enhanced ThinktankError
    const error = new ThinktankError('Unable to find model "openai:gpt5"', {
      category: errorCategories.CONFIG,
      suggestions: [
        'Check your model specification',
        'Available models: openai:gpt-4o, openai:gpt-3.5-turbo'
      ],
      examples: [
        'thinktank prompt.txt openai:gpt-4o',
        'thinktank prompt.txt coding'
      ]
    });
    
    // This simulates CLI error handling logic
    const errorDisplay = formatErrorForCLI(error);
    
    // Verify error display contains expected components
    expect(errorDisplay).toContain('Error (Configuration)');
    expect(errorDisplay).toContain('Unable to find model "openai:gpt5"');
    expect(errorDisplay).toContain('Suggestions:');
    expect(errorDisplay).toContain('Check your model specification');
    expect(errorDisplay).toContain('Available models');
    expect(errorDisplay).toContain('Examples:');
    expect(errorDisplay).toContain('thinktank prompt.txt openai:gpt-4o');
  });
  
  test('ThinktankError with cause should display both errors', () => {
    // Create a ThinktankError with cause
    const cause = new Error('API key is invalid or expired');
    const error = new ThinktankError('Failed to authenticate with OpenAI API', {
      cause: cause,
      category: errorCategories.API
    });
    
    // Simulate CLI error handling
    const errorDisplay = formatErrorForCLI(error);
    
    // Verify error display
    expect(errorDisplay).toContain('Error (API)');
    expect(errorDisplay).toContain('Failed to authenticate with OpenAI API');
    expect(errorDisplay).toContain('Cause:');
    expect(errorDisplay).toContain('API key is invalid or expired');
  });
  
  test('Regular Error should be wrapped as unexpected error', () => {
    // Create a regular Error
    const error = new Error('Unknown internal error');
    
    // Simulate CLI error handling
    const errorDisplay = formatErrorForCLI(error);
    
    // Verify error display
    expect(errorDisplay).toContain('Unexpected error:');
    expect(errorDisplay).toContain('Unknown internal error');
  });
  
  test('File System errors should include correct usage examples', () => {
    // Create a File System error
    const error = new FileSystemError('File not found: nonexistent.txt', {
      filePath: '/path/to/nonexistent.txt',
      suggestions: [
        'Check that the file exists',
        'Current directory: /home/user'
      ]
    });
    
    // Simulate CLI error handling
    const errorDisplay = formatErrorForCLI(error);
    
    // Verify error display
    expect(errorDisplay).toContain('Error (File System)');
    expect(errorDisplay).toContain('File not found');
    expect(errorDisplay).toContain('Correct usage:');
    expect(errorDisplay).toContain('thinktank run prompt.txt');
  });
  
  test('API Key errors should show provider-specific guidance', () => {
    // Create an API Key error with ApiError class
    const error = new ApiError('Missing API key for provider', {
      providerId: 'openai',
      suggestions: [
        'Get your API key from platform.openai.com',
        'Set the OPENAI_API_KEY environment variable'
      ],
      examples: [
        'export OPENAI_API_KEY=your_key_here'
      ]
    });
    
    // Simulate CLI error handling
    const errorDisplay = formatErrorForCLI(error);
    
    // Verify error display
    expect(errorDisplay).toContain('Error (API)');
    expect(errorDisplay).toContain('Missing API key');
    expect(errorDisplay).toContain('OpenAI API help');
    expect(errorDisplay).toContain('platform.openai.com/api-keys');
    expect(errorDisplay).toContain('OPENAI_API_KEY');
  });
  
  test('Standard errors should be wrapped with appropriate ThinktankError type', () => {
    // Create a standard Error that looks like a network issue
    const error = new Error('Connection timeout when connecting to API server');
    
    // Simulate CLI error handling with wrapped error
    const errorDisplay = formatErrorForCLI(error);
    
    // Verify error display shows networking guidance
    expect(errorDisplay).toContain('Network error');
    expect(errorDisplay).toContain('Connection timeout');
    expect(errorDisplay).toContain('Network troubleshooting');
    expect(errorDisplay).toContain('Check your internet connection');
  });
  
  test('Configuration errors should show configuration help', () => {
    // Create a config error
    const error = new ConfigError('Invalid configuration: model not found', {
      suggestions: [
        'Check available models',
        'Verify model ID format'
      ]
    });
    
    // Simulate CLI error handling
    const errorDisplay = formatErrorForCLI(error);
    
    // Verify configuration help is shown
    expect(errorDisplay).toContain('Error (Configuration)');
    expect(errorDisplay).toContain('Invalid configuration');
    expect(errorDisplay).toContain('Configuration help');
    expect(errorDisplay).toContain('thinktank config view');
  });
  
  test('Google provider API errors should display appropriate guidance', () => {
    // Create a Google-specific API error
    const error = new ApiError('Google AI API error: Rate limit exceeded', {
      providerId: 'google',
      suggestions: [
        'Wait and try again later',
        'Reduce the frequency of requests'
      ]
    });
    
    // Simulate CLI error handling
    const errorDisplay = formatErrorForCLI(error);
    
    // Verify Google-specific help is shown
    expect(errorDisplay).toContain('Error (API)');
    expect(errorDisplay).toContain('Google AI API error');
    expect(errorDisplay).toContain('Google AI API help');
    expect(errorDisplay).toContain('aistudio.google.com/app/apikey');
    expect(errorDisplay).toContain('GEMINI_API_KEY');
  });
  
  test('OpenRouter provider API errors should display appropriate guidance', () => {
    // Create an OpenRouter-specific API error
    const error = new ApiError('OpenRouter API error: Invalid API key', {
      providerId: 'openrouter',
      suggestions: [
        'Check your API key',
        'Visit openrouter.ai to regenerate your key'
      ]
    });
    
    // Simulate CLI error handling
    const errorDisplay = formatErrorForCLI(error);
    
    // Verify OpenRouter-specific help is shown
    expect(errorDisplay).toContain('Error (API)');
    expect(errorDisplay).toContain('OpenRouter API error');
    expect(errorDisplay).toContain('OpenRouter API help');
    expect(errorDisplay).toContain('openrouter.ai/keys');
    expect(errorDisplay).toContain('OPENROUTER_API_KEY');
  });
  
  test('Error factory functions should create properly formatted errors', () => {
    // Create an error using the factory function
    const fileError = createFileNotFoundError('/path/to/file.txt');
    
    // Simulate CLI error handling
    const errorDisplay = formatErrorForCLI(fileError);
    
    // Verify error display
    expect(errorDisplay).toContain('Error (File System)');
    expect(errorDisplay).toContain('File not found');
    expect(errorDisplay).toContain('/path/to/file.txt');
    expect(errorDisplay).toContain('Check that the file exists');
    expect(errorDisplay).toContain('Correct usage:');
  });
  
  test('Model format error factory function should create helpful errors', () => {
    // Create a model format error
    const modelError = createModelFormatError(
      'openai-gpt4',  // Invalid format (missing colon)
      ['openai', 'anthropic', 'google'],
      ['openai:gpt-4o', 'anthropic:claude-3-opus']
    );
    
    // Simulate CLI error handling
    const errorDisplay = formatErrorForCLI(modelError);
    
    // Verify error display
    expect(errorDisplay).toContain('Error (Configuration)');
    expect(errorDisplay).toContain('Invalid model format');
    expect(errorDisplay).toContain('openai-gpt4');
    expect(errorDisplay).toContain('provider:modelId');
    expect(errorDisplay).toContain('Configuration help');
  });
  
  test('API key error factory function should create helpful errors', () => {
    // Create an API key error
    const missingApiKeyError = createMissingApiKeyError([
      { provider: 'openai', modelId: 'gpt-4o' },
      { provider: 'anthropic', modelId: 'claude-3-opus' }
    ]);
    
    // Simulate CLI error handling
    const errorDisplay = formatErrorForCLI(missingApiKeyError);
    
    // Verify error display
    expect(errorDisplay).toContain('Error (API)');
    expect(errorDisplay).toContain('Missing API keys');
    expect(errorDisplay).toContain('openai:gpt-4o');
    expect(errorDisplay).toContain('anthropic:claude-3-opus');
    expect(errorDisplay).toContain('get your API key');
  });
  
  test('Permission errors should show appropriate guidance', () => {
    const permissionError = new PermissionError('Permission denied when writing to file', {
      suggestions: [
        'Check file permissions',
        'Ensure you have write access to the directory'
      ]
    });
    
    // Simulate CLI error handling
    const errorDisplay = formatErrorForCLI(permissionError);
    
    // Verify permission-specific guidance
    expect(errorDisplay).toContain('Error (Permission)');
    expect(errorDisplay).toContain('Permission denied');
    expect(errorDisplay).toContain('Check file permissions');
  });
  
  test('Unknown error type should show fallback guidance', () => {
    // Create a completely unknown error (not even an Error object)
    const unknownError = "This is not an error object";
    
    // Simulate CLI error handling
    const errorDisplay = formatErrorForCliDirect(unknownError);
    
    // Verify fallback guidance is shown
    expect(errorDisplay).toContain('Error (Unknown)');
    expect(errorDisplay).toContain('An unknown error occurred');
    expect(errorDisplay).toContain('This is likely an internal error');
    expect(errorDisplay).toContain('Report this issue if it persists');
  });
});

/**
 * Test helper function that captures CLI error output from the actual handleError function
 * 
 * This function:
 * 1. Mocks process.exit to prevent tests from terminating
 * 2. Captures all calls to logger.error
 * 3. Calls the actual handleError function with the provided error
 * 4. Returns the captured error messages as a string
 * 
 * @param error - The error to handle
 * @returns The formatted error output as a string
 */
function formatErrorForCLI(error: unknown): string {
  // Capture calls to logger.error in an array
  const errorMessages: string[] = [];
  
  // Store the original implementations to restore later
  const originalLoggerError = logger.error;
  const originalProcessExit = process.exit;
  
  // Mock logger.error to capture messages
  logger.error = jest.fn().mockImplementation((msg: string) => {
    errorMessages.push(msg);
  });
  
  // Mock process.exit to prevent test termination
  process.exit = jest.fn() as unknown as (code?: number) => never;
  
  try {
    // Call the actual handleError function from CLI
    handleError(error);
    
    // Return the captured error messages
    return errorMessages.join('\n');
  } finally {
    // Restore the original implementations
    logger.error = originalLoggerError;
    process.exit = originalProcessExit;
  }
}

// Simple alias for formatErrorForCLI to maintain backward compatibility in tests
const formatErrorForCliDirect = formatErrorForCLI;
