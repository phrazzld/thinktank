/**
 * Tests for the console utilities module
 */
import * as consoleUtils from '../consoleUtils';
import { ThinktankError } from '../../core/errors';

// Mock chalk and figures to test our styling without actually modifying strings
jest.mock('chalk', () => {
  const mockRed: any = jest.fn((text) => `red(${text})`);
  mockRed.bold = jest.fn((text) => `red.bold(${text})`);
  
  return {
    green: jest.fn((text) => `green(${text})`),
    yellow: jest.fn((text) => `yellow(${text})`),
    blue: jest.fn((text) => `blue(${text})`),
    cyan: jest.fn((text) => `cyan(${text})`),
    dim: jest.fn((text) => `dim(${text})`),
    bold: {
      blue: jest.fn((text) => `bold.blue(${text})`),
    },
    red: mockRed
  };
});

jest.mock('figures', () => {
  return {
    tick: '+',
    cross: 'x',
    warning: '!',
    info: 'i',
    pointer: '>',
    line: '-',
    bullet: '*',
  };
});

describe('consoleUtils', () => {
  describe('styling functions', () => {
    test('styleSuccess should format text with green tick', () => {
      const result = consoleUtils.styleSuccess('Success message');
      expect(result).toBe('green(+) Success message');
    });

    test('styleError should format text with red cross', () => {
      const result = consoleUtils.styleError('Error message');
      expect(result).toBe('red(x) Error message');
    });

    test('styleWarning should format text with yellow warning', () => {
      const result = consoleUtils.styleWarning('Warning message');
      expect(result).toBe('yellow(!) Warning message');
    });

    test('styleInfo should format text with blue info', () => {
      const result = consoleUtils.styleInfo('Info message');
      expect(result).toBe('blue(i) Info message');
    });

    test('styleHeader should format text as bold blue', () => {
      const result = consoleUtils.styleHeader('Header');
      expect(result).toBe('bold.blue(Header)');
    });

    test('styleDim should format text as dimmed', () => {
      const result = consoleUtils.styleDim('Dimmed text');
      expect(result).toBe('dim(Dimmed text)');
    });

    test('divider should create a styled horizontal line', () => {
      const result = consoleUtils.divider(5);
      expect(result).toBe('dim(-----)');
    });
  });

  describe('error formatting', () => {
    // ThinktankError is imported at the top level
    
    test('formatError should format error with category and tip', () => {
      const result = consoleUtils.formatError(
        'Something went wrong', 
        consoleUtils.errorCategories.API, 
        'Check your API key'
      );
      expect(result).toContain('red.bold(Error)');
      expect(result).toContain('yellow(API)');
      expect(result).toContain('Something went wrong');
      expect(result).toContain('cyan(i)');
      expect(result).toContain('Tip: Check your API key');
    });

    test('formatError should handle Error objects', () => {
      const error = new Error('Failed to connect');
      const result = consoleUtils.formatError(error);
      expect(result).toContain('Failed to connect');
    });
    
    test('formatError should use format() method when given a ThinktankError', () => {
      // Create a ThinktankError with suggestions and examples
      const thinktankError = new ThinktankError('Test error message', {
        category: consoleUtils.errorCategories.CONFIG,
        suggestions: ['Try this', 'Or try that'],
        examples: ['Example command']
      });
      
      // Spy on the format method to verify it's called
      const formatSpy = jest.spyOn(thinktankError, 'format');
      
      const result = consoleUtils.formatError(thinktankError);
      
      // Verify format() was called
      expect(formatSpy).toHaveBeenCalled();
      
      // Verify the result contains the formatted output
      expect(result).toContain('Test error message');
      
      // Clean up
      formatSpy.mockRestore();
    });

    test('categorizeError should detect API errors', () => {
      const error = new Error('Invalid API key provided');
      const category = consoleUtils.categorizeError(error);
      expect(category).toBe(consoleUtils.errorCategories.API);
    });

    test('categorizeError should detect network errors', () => {
      const error = new Error('ETIMEDOUT: Connection timed out');
      const category = consoleUtils.categorizeError(error);
      expect(category).toBe(consoleUtils.errorCategories.NETWORK);
    });

    test('categorizeError should return UNKNOWN for unrecognized errors', () => {
      const error = new Error('Some completely random error');
      const category = consoleUtils.categorizeError(error);
      expect(category).toBe(consoleUtils.errorCategories.UNKNOWN);
    });
    
    test('categorizeError should use category from ThinktankError', () => {
      const thinktankError = new ThinktankError('Permission denied', {
        category: consoleUtils.errorCategories.PERMISSION
      });
      
      const category = consoleUtils.categorizeError(thinktankError);
      expect(category).toBe(consoleUtils.errorCategories.PERMISSION);
    });

    test('getTroubleshootingTip should return appropriate tips', () => {
      const apiError = new Error('Invalid API key');
      const tip = consoleUtils.getTroubleshootingTip(
        apiError, 
        consoleUtils.errorCategories.API
      );
      expect(tip).toContain('Check your API key');
    });
    
    test('getTroubleshootingTip should use first suggestion from ThinktankError if available', () => {
      const thinktankError = new ThinktankError('Network error', {
        category: consoleUtils.errorCategories.NETWORK,
        suggestions: ['Check your internet connection', 'Try again later']
      });
      
      const tip = consoleUtils.getTroubleshootingTip(
        thinktankError,
        consoleUtils.errorCategories.NETWORK
      );
      
      expect(tip).toBe('Check your internet connection');
    });

    test('formatErrorWithTip should automatically categorize and add tip', () => {
      const error = new Error('API key is invalid');
      const result = consoleUtils.formatErrorWithTip(error);
      expect(result).toContain('API');
      expect(result).toContain('Check your API key');
    });
    
    test('formatErrorWithTip should use format() method when given a ThinktankError', () => {
      // Create a ThinktankError with suggestions and examples
      const thinktankError = new ThinktankError('Invalid configuration', {
        category: consoleUtils.errorCategories.CONFIG,
        suggestions: ['Check your config file']
      });
      
      // Spy on the format method to verify it's called
      const formatSpy = jest.spyOn(thinktankError, 'format');
      
      const result = consoleUtils.formatErrorWithTip(thinktankError);
      
      // Verify format() was called
      expect(formatSpy).toHaveBeenCalled();
      
      // Verify the result contains the formatted output
      expect(result).toContain('Invalid configuration');
      
      // Clean up
      formatSpy.mockRestore();
    });

    test('createFileNotFoundError should generate helpful error with suggestions', () => {
      // Mock process.cwd to return a predictable value
      const originalCwd = process.cwd;
      process.cwd = jest.fn().mockReturnValue('/home/user/project');
      
      try {
        const error = consoleUtils.createFileNotFoundError('nonexistent.txt');
        
        // Verify error has the expected properties
        expect(error.message).toBe('File not found: nonexistent.txt');
        expect((error as any).category).toBe(consoleUtils.errorCategories.FILESYSTEM);
        
        // Suggestions should be an array with useful tips
        expect(Array.isArray((error as any).suggestions)).toBe(true);
        expect((error as any).suggestions.length).toBeGreaterThan(0);
        
        // Examples should provide command examples
        expect(Array.isArray((error as any).examples)).toBe(true);
        expect((error as any).examples.length).toBeGreaterThan(0);
        
        // Examples should mention the basename (nonexistent)
        expect((error as any).examples.some((ex: string) => ex.includes('nonexistent'))).toBe(true);
        
        // Should mention working directory
        const suggestions = (error as any).suggestions.join(' ');
        expect(suggestions).toContain('/home/user/project');
      } finally {
        // Restore original method
        process.cwd = originalCwd;
      }
    });
    
    test('createModelFormatError should generate helpful error with suggestions', () => {
      // Different invalid model format scenarios
      const testCases = [
        { input: 'invalidformat', expectText: 'must be specified as' },
        { input: 'openai:', expectText: 'Missing model ID' },
        { input: ':gpt4', expectText: 'Missing provider' }
      ];
      
      testCases.forEach(testCase => {
        const error = consoleUtils.createModelFormatError(testCase.input);
        
        // Verify error has the expected properties
        expect(error.message).toContain(testCase.expectText);
        expect((error as any).category).toBe(consoleUtils.errorCategories.CONFIG);
        
        // Suggestions should be an array with useful tips
        expect(Array.isArray((error as any).suggestions)).toBe(true);
        expect((error as any).suggestions.length).toBeGreaterThan(0);
        
        // Should contain format guidance
        const suggestions = (error as any).suggestions.join(' ');
        expect(suggestions).toContain('provider:modelId');
        
        // Examples should be provided
        expect(Array.isArray((error as any).examples)).toBe(true);
        expect((error as any).examples.length).toBeGreaterThan(0);
      });
      
      // Test with available providers
      const errorWithProviders = consoleUtils.createModelFormatError(
        'invalidformat', 
        ['openai', 'anthropic'], 
        ['openai:gpt-4o', 'anthropic:claude-3']
      );
      
      // Should include available providers in suggestions
      const suggestions = (errorWithProviders as any).suggestions.join(' ');
      expect(suggestions).toContain('Available providers: openai, anthropic');
      expect(suggestions).toContain('Example models');
    });
    
    test('createModelNotFoundError should generate helpful error with suggestions', () => {
      // Test basic model not found error
      const error = consoleUtils.createModelNotFoundError('openai:nonexistent-model');
      
      // Verify basic properties
      expect(error.message).toContain('not found in configuration');
      expect((error as any).category).toBe(consoleUtils.errorCategories.CONFIG);
      
      // Should have suggestions
      expect(Array.isArray((error as any).suggestions)).toBe(true);
      expect((error as any).suggestions.length).toBeGreaterThan(0);
      
      // Test with available models
      const errorWithModels = consoleUtils.createModelNotFoundError(
        'openai:nonexistent-model',
        ['openai:gpt-4o', 'openai:gpt-3.5-turbo', 'anthropic:claude-3']
      );
      
      // Should list available models from same provider
      const suggestions = (errorWithModels as any).suggestions.join(' ');
      expect(suggestions).toContain('Available models from openai');
      expect(suggestions).toContain('gpt-4o');
      
      // Test with group context
      const errorWithGroup = consoleUtils.createModelNotFoundError(
        'openai:nonexistent-model',
        ['openai:gpt-4o', 'anthropic:claude-3'],
        'coding'
      );
      
      // Should mention the group
      expect(errorWithGroup.message).toContain('not found in group "coding"');
      const groupSuggestions = (errorWithGroup as any).suggestions.join(' ');
      expect(groupSuggestions).toContain('in the "coding" group configuration');
    });
    
    test('createMissingApiKeyError should generate helpful error with provider-specific instructions', () => {
      // Test with multiple providers
      const missingModels = [
        { provider: 'openai', modelId: 'gpt-4o' },
        { provider: 'anthropic', modelId: 'claude-3-7-sonnet-20250219' },
        { provider: 'google', modelId: 'gemini-pro' }
      ];
      
      const error = consoleUtils.createMissingApiKeyError(missingModels);
      
      // Verify basic properties
      expect(error.message).toContain('Missing API keys for 3 models');
      expect((error as any).category).toBe(consoleUtils.errorCategories.API);
      
      // Should have suggestions
      expect(Array.isArray((error as any).suggestions)).toBe(true);
      
      // Should group by provider
      const suggestions = (error as any).suggestions.join(' ');
      expect(suggestions).toContain('Missing API key for openai');
      expect(suggestions).toContain('Missing API key for anthropic');
      expect(suggestions).toContain('Missing API key for google');
      
      // Should contain provider-specific instructions
      expect(suggestions).toContain('platform.openai.com/api-keys');
      expect(suggestions).toContain('console.anthropic.com/keys');
      expect(suggestions).toContain('aistudio.google.com/app/apikey');
      
      // Should have environment variable setup instructions
      expect(suggestions).toContain('export PROVIDER_API_KEY=your_key_here');
      expect(suggestions).toContain('set PROVIDER_API_KEY=your_key_here');
      expect(suggestions).toContain('$env:PROVIDER_API_KEY');
      
      // Should have examples
      expect(Array.isArray((error as any).examples)).toBe(true);
      expect((error as any).examples.length).toBe(3); // One per provider
      
      // Examples should include provider-specific environment variables
      const examplesString = (error as any).examples.join(' ');
      expect(examplesString).toContain('OPENAI_API_KEY');
      expect(examplesString).toContain('ANTHROPIC_API_KEY');
      expect(examplesString).toContain('GOOGLE_API_KEY');
      
      // Test with single provider (singular message)
      const singleModel = [{ provider: 'openai', modelId: 'gpt-4o' }];
      const singleError = consoleUtils.createMissingApiKeyError(singleModel);
      
      expect(singleError.message).toContain('Missing API key for 1 model');
      expect(singleError.message).not.toContain('keys');
    });
  });
});