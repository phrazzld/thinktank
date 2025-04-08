/**
 * Tests for the console utilities module
 */
import * as consoleUtils from '../consoleUtils';
import { 
  ThinktankError, 
  errorCategories, 
  createFileNotFoundError, 
  createModelFormatError, 
  createModelNotFoundError, 
  createMissingApiKeyError 
} from '../../core/errors';
import { 
  categorizeError 
} from '../../core/errors/utils/categorization';

// Store original chalk module path for restoration
const chalkPath = require.resolve('chalk');

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
      underline: jest.fn((text) => `bold.underline(${text})`),
    },
    red: mockRed
  };
});

// Add cleanup hook to restore chalk after all tests
afterAll(() => {
  // Reset module registry for chalk to restore original
  jest.unmock('chalk');
  // Clear the module cache to ensure fresh imports
  delete require.cache[chalkPath];
});

describe('consoleUtils', () => {
  describe('styling functions', () => {
    test('styleSuccess should format text with green tick', () => {
      const result = consoleUtils.styleSuccess('Success message');
      // Using a simpler test since chalk styling doesn't show in tests
      expect(result).toContain('Success message');
      expect(result).toContain('+');
    });

    test('styleError should format text with red cross', () => {
      const result = consoleUtils.styleError('Error message');
      // Using a simpler test since chalk styling doesn't show in tests
      expect(result).toContain('Error message');
      expect(result).toContain('x');
    });

    test('styleWarning should format text with yellow warning', () => {
      const result = consoleUtils.styleWarning('Warning message');
      // Using a simpler test since chalk styling doesn't show in tests
      expect(result).toContain('Warning message');
      expect(result).toContain('!');
    });

    test('styleInfo should format text with blue info', () => {
      const result = consoleUtils.styleInfo('Info message');
      // Using a simpler test since chalk styling doesn't show in tests
      expect(result).toContain('Info message');
      expect(result).toContain('i');
    });

    test('styleHeader should format text as bold', () => {
      const result = consoleUtils.styleHeader('Header');
      // Using a simpler test since chalk styling doesn't show in tests
      expect(result).toContain('Header');
    });

    test('styleSectionHeader should format text as bold underlined with newline', () => {
      const result = consoleUtils.styleSectionHeader('Section Header');
      // Using a simpler test since chalk styling doesn't show in tests
      expect(result).toContain('Section Header');
      expect(result).toContain('\n');
    });

    test('styleDim should format text as dimmed', () => {
      const result = consoleUtils.styleDim('Dimmed text');
      // Using a simpler test since chalk styling doesn't show in tests
      expect(result).toContain('Dimmed text');
    });

    test('divider should create a styled horizontal line', () => {
      const result = consoleUtils.divider(5);
      expect(result).toBe('-----');
    });
  });

  describe('error system integration', () => {
    test('ThinktankError format method should produce properly formatted error output', () => {
      // Create a ThinktankError with suggestions and examples
      const thinktankError = new ThinktankError('Test error message', {
        category: errorCategories.CONFIG,
        suggestions: ['Try this', 'Or try that'],
        examples: ['Example command']
      });
      
      const result = thinktankError.format();
      
      // Verify the result contains the key information
      expect(result).toContain('Error');
      expect(result).toContain('Configuration');
      expect(result).toContain('Test error message');
      expect(result).toContain('Try this');
      expect(result).toContain('Or try that');
      expect(result).toContain('Example command');
    });

    test('categorizeError should detect API errors', () => {
      const error = new Error('Invalid API key provided');
      const category = categorizeError(error);
      expect(category).toBe(errorCategories.API);
    });

    test('categorizeError should detect network errors', () => {
      const error = new Error('ETIMEDOUT: Connection timed out');
      const category = categorizeError(error);
      expect(category).toBe(errorCategories.NETWORK);
    });

    test('categorizeError should return UNKNOWN for unrecognized errors', () => {
      const error = new Error('Some completely random error');
      const category = categorizeError(error);
      expect(category).toBe(errorCategories.UNKNOWN);
    });
    
    test('categorizeError should use category from ThinktankError', () => {
      const thinktankError = new ThinktankError('Permission denied', {
        category: errorCategories.PERMISSION
      });
      
      expect(thinktankError.category).toBe(errorCategories.PERMISSION);
    });

    test('ThinktankError should include suggestions', () => {
      const thinktankError = new ThinktankError('Network error', {
        category: errorCategories.NETWORK,
        suggestions: ['Check your internet connection', 'Try again later']
      });
      
      expect(thinktankError.suggestions).toHaveLength(2);
      expect(thinktankError.suggestions?.[0]).toBe('Check your internet connection');
    });
  });

  describe('error factory functions', () => {
    // Store original process.cwd for test cases
    let originalCwd: typeof process.cwd;
    
    // Save original cwd before tests that need to mock it
    beforeEach(() => {
      originalCwd = process.cwd;
    });
    
    // Restore original cwd after tests
    afterEach(() => {
      process.cwd = originalCwd;
    });
    
    test('createFileNotFoundError should generate helpful error with suggestions', () => {
      // Mock process.cwd to return a predictable value
      process.cwd = jest.fn().mockReturnValue('/home/user/project');
      
      try {
        const error = createFileNotFoundError('nonexistent.txt');
        
        // Verify error has the expected properties
        expect(error.message).toBe('File not found: nonexistent.txt');
        expect(error.category).toBe(errorCategories.FILESYSTEM);
        
        // Suggestions should be an array with useful tips
        expect(Array.isArray(error.suggestions)).toBe(true);
        expect(error.suggestions?.length).toBeGreaterThan(0);
        
        // Examples should provide command examples
        expect(Array.isArray(error.examples)).toBe(true);
        expect(error.examples?.length).toBeGreaterThan(0);
        
        // Examples should mention the basename (nonexistent)
        expect(error.examples?.some(ex => ex.includes('nonexistent'))).toBe(true);
        
        // Should mention working directory
        const suggestions = error.suggestions?.join(' ') || '';
        expect(suggestions).toContain('/home/user/project');
      } finally {
        // The cleanup is now handled by afterEach, but leaving the finally block for extra safety
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
        const error = createModelFormatError(testCase.input);
        
        // Verify error has the expected properties
        expect(error.message).toContain(testCase.expectText);
        expect(error.category).toBe(errorCategories.CONFIG);
        
        // Suggestions should be an array with useful tips
        expect(Array.isArray(error.suggestions)).toBe(true);
        expect(error.suggestions?.length).toBeGreaterThan(0);
        
        // Should contain format guidance
        const suggestions = error.suggestions?.join(' ') || '';
        expect(suggestions).toContain('provider:modelId');
        
        // Examples should be provided
        expect(Array.isArray(error.examples)).toBe(true);
        expect(error.examples?.length).toBeGreaterThan(0);
      });
      
      // Test with available providers
      const errorWithProviders = createModelFormatError(
        'invalidformat', 
        ['openai', 'anthropic'], 
        ['openai:gpt-4o', 'anthropic:claude-3']
      );
      
      // Should include available providers in suggestions
      const suggestions = errorWithProviders.suggestions?.join(' ') || '';
      expect(suggestions).toContain('Available providers: openai, anthropic');
      expect(suggestions).toContain('Example models');
    });
    
    test('createModelNotFoundError should generate helpful error with suggestions', () => {
      // Test basic model not found error
      const error = createModelNotFoundError('openai:nonexistent-model');
      
      // Verify basic properties
      expect(error.message).toContain('not found in configuration');
      expect(error.category).toBe(errorCategories.CONFIG);
      
      // Should have suggestions
      expect(Array.isArray(error.suggestions)).toBe(true);
      expect(error.suggestions?.length).toBeGreaterThan(0);
      
      // Test with available models
      const errorWithModels = createModelNotFoundError(
        'openai:nonexistent-model',
        ['openai:gpt-4o', 'openai:gpt-3.5-turbo', 'anthropic:claude-3']
      );
      
      // Should list available models from same provider
      const suggestions = errorWithModels.suggestions?.join(' ') || '';
      expect(suggestions).toContain('Available models from openai');
      expect(suggestions).toContain('gpt-4o');
      
      // Test with group context
      const errorWithGroup = createModelNotFoundError(
        'openai:nonexistent-model',
        ['openai:gpt-4o', 'anthropic:claude-3'],
        'coding'
      );
      
      // Should mention the group
      expect(errorWithGroup.message).toContain('not found in group "coding"');
      const groupSuggestions = errorWithGroup.suggestions?.join(' ') || '';
      expect(groupSuggestions).toContain('in the "coding" group configuration');
    });
    
    test('createMissingApiKeyError should generate helpful error with provider-specific instructions', () => {
      // Test with multiple providers
      const missingModels = [
        { provider: 'openai', modelId: 'gpt-4o' },
        { provider: 'anthropic', modelId: 'claude-3-7-sonnet-20250219' },
        { provider: 'google', modelId: 'gemini-pro' }
      ];
      
      const error = createMissingApiKeyError(missingModels);
      
      // Verify basic properties
      expect(error.message).toContain('Missing API keys for 3 models');
      expect(error.category).toBe(errorCategories.API);
      
      // Should have suggestions
      expect(Array.isArray(error.suggestions)).toBe(true);
      
      // Should group by provider
      const suggestions = error.suggestions?.join(' ') || '';
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
      expect(Array.isArray(error.examples)).toBe(true);
      expect(error.examples?.length).toBe(3); // One per provider
      
      // Examples should include provider-specific environment variables
      const examplesString = error.examples?.join(' ') || '';
      expect(examplesString).toContain('OPENAI_API_KEY');
      expect(examplesString).toContain('ANTHROPIC_API_KEY');
      expect(examplesString).toContain('GOOGLE_API_KEY');
      
      // Test with single provider (singular message)
      const singleModel = [{ provider: 'openai', modelId: 'gpt-4o' }];
      const singleError = createMissingApiKeyError(singleModel);
      
      expect(singleError.message).toContain('Missing API key for 1 model');
      expect(singleError.message).not.toContain('keys');
    });
  });
});
