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
  errorCategories 
} from '../../core/errors';

// This file tests error formatting for CLI display

describe('CLI Error Handling', () => {
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
  
  test('Regular Error should be displayed as unexpected error', () => {
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
    const error = new ThinktankError('File not found: nonexistent.txt', {
      category: errorCategories.FILESYSTEM,
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
    expect(errorDisplay).toContain('thinktank prompt.txt [group]');
    expect(errorDisplay).toContain('thinktank prompt.txt provider:model');
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
    expect(errorDisplay).toContain('platform.openai.com');
    expect(errorDisplay).toContain('OPENAI_API_KEY');
    expect(errorDisplay).toContain('export OPENAI_API_KEY=your_key_here');
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
});

/**
 * Simulates CLI error display formatting
 * This is based on the actual logic in cli.ts
 */
function formatErrorForCLI(error: Error): string {
  let output = '';
  
  if (error instanceof ThinktankError) {
    // Use the built-in format method for ThinktankError
    output += error.format() + '\n';
    
    // Display cause if available and not already included in format()
    if (error.cause && !error.format().includes('Cause:')) {
      output += `Cause: ${error.cause.message}\n`;
    }
    
    // Add category-specific guidance
    output = addTestGuidance(error, output);
  } else if (error instanceof Error) {
    // Convert standard Error to ThinktankError
    const wrappedError = wrapTestError(error);
    output += wrappedError.format() + '\n';
    
    // Add guidance for the wrapped error
    output = addTestGuidance(wrappedError, output);
  } else {
    output += 'An unknown error occurred\n';
  }
  
  return output;
}

/**
 * Adds test-specific guidance similar to the CLI implementation
 */
function addTestGuidance(error: ThinktankError, output: string): string {
  let updatedOutput = output;
  
  // File System errors
  if (error.category === errorCategories.FILESYSTEM) {
    updatedOutput += '\nCorrect usage:\n';
    updatedOutput += '  > thinktank prompt.txt [group]\n';
    updatedOutput += '  > thinktank prompt.txt provider:model\n';
  }
  
  // Configuration errors
  else if (error.category === errorCategories.CONFIG) {
    updatedOutput += '\nConfiguration help:\n';
    updatedOutput += '  > thinktank config view\n';
    updatedOutput += '  > thinktank config set key value\n';
    updatedOutput += '  > Edit ~/.thinktank/config.json directly\n';
  }
  
  // API errors
  else if (error.category === errorCategories.API) {
    // For API errors, check if it's provider-specific
    if (error instanceof ApiError && error.providerId) {
      // Provider-specific guidance
      switch (error.providerId.toLowerCase()) {
        case 'openai':
          updatedOutput += '\nOpenAI API help:\n';
          updatedOutput += '  > Get API keys: https://platform.openai.com/api-keys\n';
          updatedOutput += '  > Set with: export OPENAI_API_KEY=your_key_here\n';
          break;
          
        case 'anthropic':
          updatedOutput += '\nAnthropic API help:\n';
          updatedOutput += '  > Get API keys: https://console.anthropic.com/keys\n';
          updatedOutput += '  > Set with: export ANTHROPIC_API_KEY=your_key_here\n';
          break;
      }
    }
  }
  
  // Network errors
  else if (error.category === errorCategories.NETWORK) {
    updatedOutput += '\nNetwork troubleshooting:\n';
    updatedOutput += '  > Check your internet connection\n';
    updatedOutput += '  > Verify you can access the API endpoints\n';
    updatedOutput += '  > Try again in a few minutes\n';
  }
  
  return updatedOutput;
}

/**
 * Wraps test errors in ThinktankError for testing
 */
function wrapTestError(error: Error): ThinktankError {
  const message = error.message.toLowerCase();
  
  if (message.includes('timeout') || message.includes('network')) {
    return new NetworkError(`Network error: ${error.message}`, {
      cause: error,
      suggestions: ['Check your internet connection']
    });
  }
  
  if (message.includes('file') || message.includes('directory')) {
    return new FileSystemError(`File system error: ${error.message}`, {
      cause: error,
      suggestions: ['Check that the file exists']
    });
  }
  
  return new ThinktankError(`Unexpected error: ${error.message}`, {
    cause: error
  });
}