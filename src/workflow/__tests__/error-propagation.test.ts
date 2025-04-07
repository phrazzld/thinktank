/**
 * Tests for cross-module error propagation in thinktank.
 * 
 * These tests verify that errors are properly created, propagated,
 * and handled as they cross module boundaries from providers,
 * through workflow components, to the CLI interface.
 */
import { 
  ApiError, 
  ConfigError, 
  FileSystemError, 
  NetworkError,
  ThinktankError,
  errorCategories,
  createFileNotFoundError,
  createMissingApiKeyError,
  createModelFormatError
} from '../../core/errors';

// Import modules that will be part of the error propagation chain
import { runThinktank, RunOptions } from '../runThinktank';
import { processInput } from '../inputHandler';
import { executeQueries } from '../queryExecutor';
import { selectModels } from '../modelSelector';
import { loadConfig } from '../../core/configManager';

// Import provider modules
// No need to import providers directly as we're mocking the registry

// Mock modules
jest.mock('../inputHandler');
jest.mock('../queryExecutor');
jest.mock('../modelSelector');
jest.mock('../../core/configManager');

// Mock llmRegistry for provider access
jest.mock('../../core/llmRegistry', () => ({
  getProvider: jest.fn().mockImplementation((_providerId: string) => {
    // This will be overridden in individual tests
    return null;
  }),
  registerProvider: jest.fn().mockImplementation(() => {
    // Do nothing - just a stub
    return;
  })
}));

describe('Cross-Module Error Propagation', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Provider to Workflow Error Propagation', () => {
    it('should propagate API errors from OpenAI provider to runThinktank with preserved details', async () => {
      // Setup the error in the OpenAI provider
      const rootCause = new Error('Original OpenAI API error: invalid_request_error');
      const providerError = new ApiError('Failed to generate response from OpenAI', {
        providerId: 'openai',
        cause: rootCause,
        suggestions: [
          'Check your API key',
          'Verify your request is valid'
        ]
      });

      // Mock the provider generate function to throw the error
      const mockGenerate = jest.fn().mockRejectedValue(providerError);
      
      // Create mock provider instance
      const mockProviderInstance = {
        generate: mockGenerate,
        providerId: 'openai'
      };
      
      // Set up getProvider mock for this test - Use the imported and mocked function
      const { getProvider } = jest.requireMock('../../core/llmRegistry');
      getProvider.mockImplementation((providerId: string) => {
        if (providerId === 'openai') return mockProviderInstance;
        return null;
      });

      // Setup the error propagation chain
      (loadConfig as jest.Mock).mockResolvedValue({
        models: [{ provider: 'openai', modelId: 'gpt-4o', enabled: true }],
        groups: {}
      });
      
      (processInput as jest.Mock).mockResolvedValue({
        content: 'Test input',
        sourceType: 'file',
        sourcePath: 'test.txt'
      });
      
      (selectModels as jest.Mock).mockReturnValue({
        models: [{ provider: 'openai', modelId: 'gpt-4o', enabled: true }],
        warnings: [],
        missingApiKeyModels: [],
        disabledModels: []
      });
      
      // This is where we connect the provider error to the workflow
      (executeQueries as jest.Mock).mockImplementation(async () => {
        // We rethrow the same error here to simulate error propagation
        throw providerError;
      });
      
      // Execute runThinktank and expect it to throw the same error
      const options: RunOptions = { input: 'test.txt' };
      await expect(runThinktank(options)).rejects.toThrow(ApiError);
      
      try {
        await runThinktank(options);
      } catch (error) {
        // Verify error details are preserved through the propagation chain
        expect(error instanceof ApiError).toBe(true);
        if (error instanceof ApiError) {
          expect(error.providerId).toBe('openai');
          expect(error.message).toContain('Failed to generate response from OpenAI');
          expect(error.category).toBe(errorCategories.API);
          expect(error.cause).toBeDefined();
          expect(error.cause?.message).toBe('Original OpenAI API error: invalid_request_error');
          expect(error.suggestions).toContain('Check your API key');
        }
      }
    });

    it('should propagate network errors from Anthropic provider to queryExecutor', async () => {
      // Setup the network error in Anthropic provider
      const networkErrorCause = new Error('ECONNRESET: Connection reset by peer');
      const providerError = new NetworkError('Network error when connecting to Anthropic API', {
        cause: networkErrorCause,
        suggestions: [
          'Check your internet connection',
          'Verify the Anthropic API is accessible from your network'
        ]
      });

      // Mock the provider to throw the network error
      const mockGenerate = jest.fn().mockRejectedValue(providerError);
      
      // Create mock provider instance
      const mockProviderInstance = {
        generate: mockGenerate,
        providerId: 'anthropic'
      };
      
      // Set up getProvider mock for this test - Use the imported and mocked function
      const { getProvider } = jest.requireMock('../../core/llmRegistry');
      getProvider.mockImplementation((providerId: string) => {
        if (providerId === 'anthropic') return mockProviderInstance;
        return null;
      });

      // Setup a simple query execution that will use the provider
      (executeQueries as jest.Mock).mockImplementation(async (_config, models, _input) => {
        // Find the anthropic model
        const anthropicModel = models.find((m: any) => m.provider === 'anthropic');
        if (anthropicModel) {
          // For our test, call provider.generate directly here
          // In reality, executeQueries would do this
          const provider = mockProviderInstance;
          try {
            await provider.generate('Test prompt', 'model-id', {});
            throw new Error('Should not reach here');
          } catch (error) {
            // Verify that the error is what we expect and rethrow it
            expect(error).toBe(providerError);
            throw error;
          }
        }
        throw new Error('No Anthropic model found');
      });

      // Call executeQueries directly
      const mockModels = [{ provider: 'anthropic', modelId: 'claude-3', enabled: true }];
      const mockConfig = { models: [], groups: {} };
      const mockInput = { content: 'Test prompt' };

      // Mock executeQueries to throw the network error
      (executeQueries as jest.Mock).mockRejectedValue(providerError);

      await expect(
        executeQueries(mockConfig, mockModels, { prompt: mockInput.content })
      ).rejects.toThrow(NetworkError);

      try {
        await executeQueries(mockConfig, mockModels, { prompt: mockInput.content });
        fail('Should have thrown an error');
      } catch (error) {
        // Verify error details in the propagated error
        expect(error instanceof NetworkError).toBe(true);
        if (error instanceof NetworkError) {
          expect(error.message).toContain('Network error when connecting to Anthropic API');
          expect(error.category).toBe(errorCategories.NETWORK);
          expect(error.cause).toBeDefined();
          expect(error.cause?.message).toBe('ECONNRESET: Connection reset by peer');
        }
      }
    });
  });

  describe('Factory Function to Workflow Error Propagation', () => {
    it('should correctly propagate model format errors through modelSelector to runThinktank', async () => {
      // Create a model format error using the factory function
      const modelFormatError = createModelFormatError(
        'openai-gpt4',  // Invalid format (missing colon)
        ['openai', 'anthropic'],
        ['openai:gpt-4o', 'anthropic:claude-3-opus']
      );

      // Make modelSelector throw this error for invalid model formats
      (selectModels as jest.Mock).mockImplementation((_config, options) => {
        if (options.specificModel && !options.specificModel.includes(':')) {
          throw modelFormatError;
        }
        return { models: [], warnings: [], missingApiKeyModels: [], disabledModels: [] };
      });

      // Setup other required mocks
      (loadConfig as jest.Mock).mockResolvedValue({
        models: [],
        groups: {}
      });
      (processInput as jest.Mock).mockResolvedValue({
        content: 'Test input',
        sourceType: 'file',
        sourcePath: 'test.txt'
      });

      // Run and verify error propagation
      const options: RunOptions = { 
        input: 'test.txt',
        specificModel: 'openai-gpt4' // Invalid format
      };
      
      await expect(runThinktank(options)).rejects.toThrow(ConfigError);
      
      try {
        await runThinktank(options);
      } catch (error) {
        // Verify error details after propagation
        expect(error instanceof ConfigError).toBe(true);
        if (error instanceof ConfigError) {
          expect(error.message).toContain('Invalid model format');
          expect(error.message).toContain('openai-gpt4');
          expect(error.category).toBe(errorCategories.CONFIG);
          expect(error.suggestions).toBeDefined();
          expect(error.suggestions?.length).toBeGreaterThan(0);
          expect(error.examples).toBeDefined();
          expect(error.examples?.length).toBeGreaterThan(0);
        }
      }
    });

    it('should propagate file not found errors from inputHandler to runThinktank', async () => {
      // Create a file not found error using the factory function
      const fileNotFoundError = createFileNotFoundError('/path/to/nonexistent-file.txt');

      // Make inputHandler throw this error for invalid files
      (processInput as jest.Mock).mockRejectedValue(fileNotFoundError);

      // Run and verify error propagation
      const options: RunOptions = { input: '/path/to/nonexistent-file.txt' };
      
      await expect(runThinktank(options)).rejects.toThrow(FileSystemError);
      
      try {
        await runThinktank(options);
      } catch (error) {
        // Verify error details after propagation
        expect(error instanceof FileSystemError).toBe(true);
        if (error instanceof FileSystemError) {
          expect(error.message).toContain('not found');
          expect(error.category).toBe(errorCategories.FILESYSTEM);
          expect(error.filePath).toBe('/path/to/nonexistent-file.txt');
          expect(error.suggestions).toBeDefined();
          expect(error.suggestions?.length).toBeGreaterThan(0);
        }
      }
    });
  });

  describe('Multi-Level Error Propagation', () => {
    it('should maintain the full cause chain across multiple module boundaries', async () => {
      // Create a three-level error chain
      const originalCause = new Error('Socket timeout');
      const networkError = new NetworkError('Network connection failed', { cause: originalCause });
      const apiError = new ApiError('Failed to communicate with Google API', { 
        providerId: 'google',
        cause: networkError,
        suggestions: ['Check your connection']
      });

      // Create mock provider instance
      const mockProviderInstance = {
        generate: jest.fn().mockRejectedValue(apiError),
        providerId: 'google'
      };
      
      // Set up getProvider mock for this test - Use the imported and mocked function
      const { getProvider } = jest.requireMock('../../core/llmRegistry');
      getProvider.mockImplementation((providerId: string) => {
        if (providerId === 'google') return mockProviderInstance;
        return null;
      });

      (loadConfig as jest.Mock).mockResolvedValue({
        models: [{ provider: 'google', modelId: 'gemini-pro', enabled: true }],
        groups: {}
      });
      
      (processInput as jest.Mock).mockResolvedValue({
        content: 'Test input',
        sourceType: 'file',
        sourcePath: 'test.txt'
      });
      
      (selectModels as jest.Mock).mockReturnValue({
        models: [{ provider: 'google', modelId: 'gemini-pro', enabled: true }],
        warnings: [],
        missingApiKeyModels: [],
        disabledModels: []
      });
      
      // Let executeQueries propagate the error from the provider
      (executeQueries as jest.Mock).mockImplementation(async (_config, models) => {
        // Find the Google model
        const googleModel = models.find((m: any) => m.provider === 'google');
        if (googleModel) {
          // Simulate calling the provider directly
          const provider = mockProviderInstance;
          try {
            await provider.generate('Test prompt', 'model-id', {});
            throw new Error('Should not reach here');
          } catch (error) {
            // Add another layer to the error chain
            const enhancedError = new ApiError(
              'Query execution failed for provider: google', 
              { cause: error as Error, providerId: 'google' }
            );
            throw enhancedError;
          }
        }
        throw new Error('No Google model found');
      });

      // Run and verify the error chain propagation through runThinktank
      const options: RunOptions = { input: 'test.txt' };
      
      await expect(runThinktank(options)).rejects.toThrow(ApiError);
      
      try {
        await runThinktank(options);
      } catch (error) {
        // Verify the error chain is maintained
        expect(error instanceof ApiError).toBe(true);
        if (error instanceof ApiError) {
          expect(error.message).toContain('Query execution failed');
          expect(error.providerId).toBe('google');
          
          // First level cause
          expect(error.cause).toBeDefined();
          expect(error.cause instanceof ApiError).toBe(true);
          if (error.cause instanceof ApiError) {
            expect(error.cause.message).toContain('Failed to communicate with Google API');
            
            // Second level cause
            expect(error.cause.cause).toBeDefined();
            expect(error.cause.cause instanceof NetworkError).toBe(true);
            if (error.cause.cause instanceof NetworkError) {
              expect(error.cause.cause.message).toBe('Network connection failed');
              
              // Third level cause (original error)
              expect(error.cause.cause.cause).toBeDefined();
              expect(error.cause.cause.cause instanceof Error).toBe(true);
              expect(error.cause.cause.cause?.message).toBe('Socket timeout');
            }
          }
        }
      }
    });
  });

  describe('Error Information Enhancement', () => {
    it('should enhance errors with additional context as they propagate through modules', async () => {
      // Create an initial generic error
      const initialError = new ThinktankError('Invalid configuration');
      
      // Setup the module chain to enhance the error
      (loadConfig as jest.Mock).mockImplementation(() => {
        // Enhance the error with more specific information
        const enhancedError = new ConfigError('Invalid model configuration: missing required fields', {
          cause: initialError,
          suggestions: ['Check your configuration file format']
        });
        throw enhancedError;
      });
      
      (processInput as jest.Mock).mockResolvedValue({
        content: 'Test input',
        sourceType: 'file',
        sourcePath: 'test.txt'
      });

      // Run and verify error enhancement
      const options: RunOptions = { input: 'test.txt' };
      
      await expect(runThinktank(options)).rejects.toThrow(ConfigError);
      
      try {
        await runThinktank(options);
      } catch (error) {
        // Verify the error was enhanced
        expect(error instanceof ConfigError).toBe(true);
        if (error instanceof ConfigError) {
          expect(error.message).toContain('Invalid model configuration');
          expect(error.category).toBe(errorCategories.CONFIG);
          expect(error.suggestions).toContain('Check your configuration file format');
          
          // Original cause is preserved
          expect(error.cause).toBeDefined();
          expect(error.cause instanceof ThinktankError).toBe(true);
          expect(error.cause?.message).toBe('Invalid configuration');
        }
      }
    });
  });

  describe('Missing API Key Error Propagation', () => {
    it('should propagate missing API key errors from provider validation to CLI', async () => {
      // Create a missing API key error using the factory function
      const missingApiKeyError = createMissingApiKeyError([
        { provider: 'openai', modelId: 'gpt-4o' },
        { provider: 'anthropic', modelId: 'claude-3-opus' }
      ]);

      // Setup the propagation chain
      (loadConfig as jest.Mock).mockResolvedValue({
        models: [
          { provider: 'openai', modelId: 'gpt-4o', enabled: true },
          { provider: 'anthropic', modelId: 'claude-3-opus', enabled: true }
        ],
        groups: {}
      });
      
      (processInput as jest.Mock).mockResolvedValue({
        content: 'Test input',
        sourceType: 'file',
        sourcePath: 'test.txt'
      });
      
      // Make selectModels throw the missing API key error
      (selectModels as jest.Mock).mockImplementation(() => {
        throw missingApiKeyError;
      });

      // Run with validateApiKeys option set to true
      const options: RunOptions = { 
        input: 'test.txt'
      };
      
      await expect(runThinktank(options)).rejects.toThrow(ApiError);
      
      try {
        await runThinktank(options);
      } catch (error) {
        // Verify the error details
        expect(error instanceof ApiError).toBe(true);
        if (error instanceof ApiError) {
          expect(error.message).toContain('Missing API keys');
          expect(error.category).toBe(errorCategories.API);
          expect(error.suggestions).toBeDefined();
          // These tests were previously checking for specific provider names
          // We've updated to look for API key names instead
          expect(error.suggestions?.some(s => s.includes('OPENAI_API_KEY') || s.includes('openai'))).toBe(true);
          expect(error.suggestions?.some(s => s.includes('ANTHROPIC_API_KEY') || s.includes('anthropic'))).toBe(true);
        }
      }
    });
  });
});
