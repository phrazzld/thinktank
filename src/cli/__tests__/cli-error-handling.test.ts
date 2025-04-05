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
  NetworkError,
  PermissionError,
  errorCategories,
  createFileNotFoundError,
  createModelFormatError,
  createMissingApiKeyError
} from '../../core/errors';
import { colors } from '../../utils/consoleUtils';
import { logger } from '../../utils/logger';

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
 * Simulates CLI error display formatting
 * This is based on the actual logic in CLI's handleError function
 */
function formatErrorForCLI(error: Error): string {
  // Capture calls to logger.error in an array so we can reconstruct the output
  const errorMessages: string[] = [];
  const originalLoggerError = logger.error;
  logger.error = jest.fn().mockImplementation((msg: string) => {
    errorMessages.push(msg);
  });
  
  // Call our mock implementation of handleError
  mockHandleError(error);
  
  // Restore logger.error
  logger.error = originalLoggerError;
  
  // Join all captured error messages into a single string
  return errorMessages.join('\n');
}

/**
 * Handles unknown error types directly (not just Error instances)
 */
function formatErrorForCliDirect(error: unknown): string {
  // Capture calls to logger.error in an array so we can reconstruct the output
  const errorMessages: string[] = [];
  const originalLoggerError = logger.error;
  logger.error = jest.fn().mockImplementation((msg: string) => {
    errorMessages.push(msg);
  });
  
  // Call our mock implementation of handleError
  mockHandleError(error);
  
  // Restore logger.error
  logger.error = originalLoggerError;
  
  // Join all captured error messages into a single string
  return errorMessages.join('\n');
}

/**
 * Simulates the CLI's handleError function
 */
function mockHandleError(error: unknown): void {
  // Handle errors with enhanced formatting
  if (error instanceof ThinktankError) {
    // Use the built-in format method for consistent error display
    logger.error(error.format());
    
    // Display cause if available and not already shown in format()
    if (error.cause && !error.format().includes('Cause:')) {
      logger.error(`${colors.dim('Cause:')} ${error.cause.message}`);
    }
    
    // Provide category-specific guidance
    addMockGuidance(error);
    
  } else if (error instanceof Error) {
    // Convert standard Error to ThinktankError for consistent formatting
    const wrappedError = wrapMockError(error);
    logger.error(wrappedError.format());
    
    // Add contextual help for the wrapped error
    addMockGuidance(wrappedError);
  } else {
    // Handle unknown errors (non-Error objects)
    const genericError = new ThinktankError('An unknown error occurred', {
      category: errorCategories.UNKNOWN,
      suggestions: [
        'This is likely an internal error in thinktank',
        'Check for updates to thinktank as this may be a fixed issue',
        'Report this issue if it persists'
      ]
    });
    logger.error(genericError.format());
  }
}

/**
 * Adds category-specific guidance based on error type
 * This simulates the actual implementation in CLI
 */
function addMockGuidance(error: ThinktankError): void {
  // File System errors
  if (error.category === errorCategories.FILESYSTEM) {
    logger.error('\nCorrect usage:');
    logger.error(`  ${colors.green('>')} thinktank run prompt.txt [--group=group]`);
    logger.error(`  ${colors.green('>')} thinktank run prompt.txt --models=provider:model`);
  }
  
  // Configuration errors
  else if (error.category === errorCategories.CONFIG) {
    logger.error('\nConfiguration help:');
    logger.error(`  ${colors.green('>')} thinktank config view`);
    logger.error(`  ${colors.green('>')} thinktank config set key value`);
    logger.error(`  ${colors.green('>')} Edit ~/.thinktank/config.json directly`);
  }
  
  // API errors
  else if (error.category === errorCategories.API) {
    // For API errors, check if it's provider-specific
    if (error instanceof ApiError && error.providerId) {
      // Provider-specific guidance
      switch (error.providerId.toLowerCase()) {
        case 'openai':
          logger.error('\nOpenAI API help:');
          logger.error(`  ${colors.green('>')} Get API keys: https://platform.openai.com/api-keys`);
          logger.error(`  ${colors.green('>')} Set with: export OPENAI_API_KEY=your_key_here`);
          break;
          
        case 'anthropic':
          logger.error('\nAnthropic API help:');
          logger.error(`  ${colors.green('>')} Get API keys: https://console.anthropic.com/keys`);
          logger.error(`  ${colors.green('>')} Set with: export ANTHROPIC_API_KEY=your_key_here`);
          break;
          
        case 'google':
          logger.error('\nGoogle AI API help:');
          logger.error(`  ${colors.green('>')} Get API keys: https://aistudio.google.com/app/apikey`);
          logger.error(`  ${colors.green('>')} Set with: export GEMINI_API_KEY=your_key_here`);
          break;
          
        case 'openrouter':
          logger.error('\nOpenRouter API help:');
          logger.error(`  ${colors.green('>')} Get API keys: https://openrouter.ai/keys`);
          logger.error(`  ${colors.green('>')} Set with: export OPENROUTER_API_KEY=your_key_here`);
          break;
          
        default:
          logger.error('\nAPI help:');
          logger.error(`  ${colors.green('>')} Ensure you have the correct API key for ${error.providerId}`);
          logger.error(`  ${colors.green('>')} Set with: export ${error.providerId.toUpperCase()}_API_KEY=your_key_here`);
      }
    } else {
      // Generic API error guidance
      logger.error('\nAPI help:');
      logger.error(`  ${colors.green('>')} Check your API credentials`);
      logger.error(`  ${colors.green('>')} Verify network connectivity to API services`);
      logger.error(`  ${colors.green('>')} Run with --debug flag for more information`);
    }
  }
  
  // Network errors
  else if (error.category === errorCategories.NETWORK) {
    logger.error('\nNetwork troubleshooting:');
    logger.error(`  ${colors.green('>')} Check your internet connection`);
    logger.error(`  ${colors.green('>')} Verify you can access the API endpoints (no firewall blocking)`);
    logger.error(`  ${colors.green('>')} Try again in a few minutes if service might be down`);
  }
  
  // Validation errors (including input errors)
  else if (error.category === errorCategories.VALIDATION || error.category === errorCategories.INPUT) {
    logger.error('\nInput help:');
    logger.error(`  ${colors.green('>')} thinktank run prompt.txt [options]`);
    logger.error(`  ${colors.green('>')} Use --help with any command for detailed usage`);
  }
  
  // For unknown or other errors, offer general debugging help
  else if (error.category === errorCategories.UNKNOWN) {
    logger.error('\nTroubleshooting help:');
    logger.error(`  ${colors.green('>')} Run with --debug flag for more information`);
    logger.error(`  ${colors.green('>')} Check thinktank documentation for guidance`);
    logger.error(`  ${colors.green('>')} Report bugs at: https://github.com/phrazzld/thinktank/issues`);
  }
}

/**
 * Wraps standard Error objects in ThinktankError for consistent formatting
 * Mirrors the implementation in the CLI
 */
function wrapMockError(error: Error): ThinktankError {
  // Try to categorize based on message content
  const message = error.message.toLowerCase();
  
  // Network-related errors
  if (message.includes('network') || 
      message.includes('econnrefused') || 
      message.includes('timeout') ||
      message.includes('socket')) {
    return new NetworkError(`Network error: ${error.message}`, {
      cause: error,
      suggestions: [
        'Check your internet connection',
        'Verify that required services are accessible from your network',
        'The service might be down or experiencing issues'
      ]
    });
  }
  
  // File-related errors
  else if (message.includes('file') || 
           message.includes('directory') || 
           message.includes('enoent') ||
           message.includes('permission denied')) {
    return new FileSystemError(`File system error: ${error.message}`, {
      cause: error,
      suggestions: [
        'Check that the file or directory exists',
        'Verify that you have appropriate permissions',
        'Ensure the path is correct'
      ]
    });
  }
  
  // Configuration errors
  else if (message.includes('config') || 
           message.includes('settings') || 
           message.includes('option')) {
    return new ConfigError(`Configuration error: ${error.message}`, {
      cause: error,
      suggestions: [
        'Check your thinktank configuration file',
        'Try resetting to default configuration with: thinktank config reset',
        'Verify that configuration values are in the correct format'
      ]
    });
  }
  
  // Default to unknown category
  return new ThinktankError(`Unexpected error: ${error.message}`, {
    category: errorCategories.UNKNOWN,
    cause: error,
    suggestions: [
      'Run with --debug flag for more detailed information',
      'Check documentation for this feature',
      'This may be an internal error in thinktank'
    ]
  });
}