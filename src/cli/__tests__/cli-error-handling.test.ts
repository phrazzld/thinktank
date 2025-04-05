/**
 * Tests for CLI error handling
 * 
 * These tests verify that the CLI properly handles and displays
 * various types of errors with helpful, actionable information.
 */
import { ThinktankError, errorCategories } from '../../core/errors';

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
    // Create an API Key error
    const error = new ThinktankError('Missing API key for provider: openai', {
      category: errorCategories.API,
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
    
    // Show general help for common errors
    if (error.category === errorCategories.FILESYSTEM) {
      output += '\nCorrect usage:\n';
      output += '  > thinktank prompt.txt [group]\n';
      output += '  > thinktank prompt.txt provider:model\n';
    }
  } else if (error instanceof Error) {
    output += `Unexpected error: ${error.message}\n`;
  } else {
    output += 'An unknown error occurred\n';
  }
  
  return output;
}