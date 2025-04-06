/**
 * Unit tests for Anthropic provider
 * 
 * Note: We import the provider first to trigger its auto-registration
 */
import { AnthropicProvider, anthropicProvider } from '../anthropic';
import { ApiError as AnthropicProviderError } from '../../core/errors';
import { ModelOptions, LLMAvailableModel } from '../../core/types';
import { clearRegistry, getProvider } from '../../core/llmRegistry';
import { ApiError, ThinktankError } from '../../core/errors';
import { createProviderRateLimitError } from '../../core/errors/factories/provider';
import Anthropic from '@anthropic-ai/sdk';

// Mock Anthropic library
jest.mock('@anthropic-ai/sdk');
const MockedAnthropic = jest.mocked(Anthropic);

// No need to mock axios, we'll use the SDK

describe('Anthropic Provider', () => {
  const originalEnv = process.env;
  
  beforeEach(() => {
    // Reset environment
    process.env = { ...originalEnv };
    
    // Reset mocks
    jest.clearAllMocks();
    
    // Clear registry
    clearRegistry();
    
    // Reset mock implementation
    MockedAnthropic.mockClear();
    
    // Re-register the provider because we cleared the registry
    try {
      anthropicProvider.providerId; // Access a property to ensure the module is initialized
    } catch (error) {
      // Ignore errors
    }
  });
  
  afterAll(() => {
    // Restore original environment
    process.env = originalEnv;
  });
  
  describe('initialization', () => {
    it('should register with the registry', () => {
      // Register manually for testing
      clearRegistry();
      new AnthropicProvider(); // This should trigger registration
      
      const provider = getProvider('anthropic');
      expect(provider).toBeDefined();
      expect(provider?.providerId).toBe('anthropic');
    });
    
    it('should use API key from constructor', async () => {
      const provider = new AnthropicProvider('test-api-key');
      
      // Prepare mock for client creation
      const mockCreate = jest.fn().mockResolvedValue({
        content: [{ type: 'text', text: 'Test response' }],
        usage: { input_tokens: 5, output_tokens: 5 },
        model: 'claude-3-opus-20240229',
        id: 'test-id',
        type: 'message',
      });

      MockedAnthropic.mockImplementation(() => {
        return {
          messages: {
            create: mockCreate,
          },
        } as unknown as Anthropic;
      });
      
      // Generate should cause client creation
      await provider.generate('Test prompt', 'claude-3-opus-20240229');
      
      // Verify API key was used
      expect(MockedAnthropic).toHaveBeenCalledWith({ apiKey: 'test-api-key', maxRetries: 0 });
    });
    
    it('should use API key from environment', async () => {
      process.env.ANTHROPIC_API_KEY = 'env-api-key';
      const provider = new AnthropicProvider();
      
      // Prepare mock for client creation
      const mockCreate = jest.fn().mockResolvedValue({
        content: [{ type: 'text', text: 'Test response' }],
        usage: { input_tokens: 5, output_tokens: 5 },
        model: 'claude-3-opus-20240229',
        id: 'test-id',
        type: 'message',
      });
      
      MockedAnthropic.mockImplementation(() => {
        return {
          messages: {
            create: mockCreate,
          },
        } as unknown as Anthropic;
      });
      
      // Generate should cause client creation
      await provider.generate('Test prompt', 'claude-3-opus-20240229');
      
      // Verify environment API key was used
      expect(MockedAnthropic).toHaveBeenCalledWith({ apiKey: 'env-api-key', maxRetries: 0 });
    });
    
    it('should throw error if API key is missing', async () => {
      // Ensure ANTHROPIC_API_KEY is not set
      delete process.env.ANTHROPIC_API_KEY;
      
      const provider = new AnthropicProvider();
      
      // Should still be catchable as AnthropicProviderError for backward compatibility
      await expect(provider.generate('Test', 'claude-3-opus-20240229')).rejects.toThrow(AnthropicProviderError);
      // But should also be an instance of ApiError from the new system
      await expect(provider.generate('Test', 'claude-3-opus-20240229')).rejects.toThrow(ApiError);
      // And should be an instance of the base ThinktankError
      await expect(provider.generate('Test', 'claude-3-opus-20240229')).rejects.toThrow(ThinktankError);
      
      try {
        await provider.generate('Test', 'claude-3-opus-20240229');
      } catch (error) {
        // Verify it has the expected properties from ApiError
        const typedError = error as AnthropicProviderError;
        expect(typedError.message).toContain('Anthropic API key is missing');
        expect(typedError.category).toBe('API');
        expect(typedError.providerId).toBe('anthropic');
        expect(typedError.suggestions).toBeDefined();
        expect(typedError.suggestions?.length).toBeGreaterThan(0);
      }
    });
  });
  
  describe('generate', () => {
    // Create a valid provider with API key
    let provider: AnthropicProvider;
    let mockCreate: jest.Mock;
    
    beforeEach(() => {
      provider = new AnthropicProvider('test-api-key');
      
      // Prepare mock for API response
      mockCreate = jest.fn().mockResolvedValue({
        content: [{ type: 'text', text: 'Test response' }],
        usage: { input_tokens: 5, output_tokens: 5 },
        model: 'claude-3-opus-20240229',
        id: 'test-id',
        type: 'message',
      });
      
      MockedAnthropic.mockImplementation(() => {
        return {
          messages: {
            create: mockCreate,
          },
        } as unknown as Anthropic;
      });
    });
    
    it('should call Anthropic API with the correct parameters', async () => {
      await provider.generate('Test prompt', 'claude-3-opus-20240229');
      
      expect(mockCreate).toHaveBeenCalledWith(expect.objectContaining({
        model: 'claude-3-opus-20240229',
        messages: [{ role: 'user' as const, content: 'Test prompt' }],
        max_tokens: 1000, // Default value now from cascading config
        temperature: 0.7, // Default temperature
      }));
    });
    
    it('should map ModelOptions to Anthropic parameters', async () => {
      const options: ModelOptions = {
        temperature: 0.7,
        maxTokens: 500,
        topP: 0.9, // Additional option
      };
      
      await provider.generate('Test prompt', 'claude-3-opus-20240229', options);
      
      expect(mockCreate).toHaveBeenCalledWith(expect.objectContaining({
        model: 'claude-3-opus-20240229',
        messages: [{ role: 'user' as const, content: 'Test prompt' }],
        temperature: 0.7,
        max_tokens: 500,
        top_p: 0.9, // Note: camelCase keys are converted to snake_case
      }));
    });
    
    it('should force temperature to 1 when thinking is enabled', async () => {
      const options: ModelOptions = {
        temperature: 0.5, // This should be overridden
        maxTokens: 500,
        thinking: {
          type: 'enabled',
          budget_tokens: 16000
        }
      };
      
      await provider.generate('Test prompt', 'claude-3-opus-20240229', options);
      
      // Verify temperature is set to exactly 1 regardless of user's setting
      expect(mockCreate).toHaveBeenCalledWith(expect.objectContaining({
        temperature: 1, // Should be exactly 1 as required by Anthropic API
        thinking: {
          type: 'enabled',
          budget_tokens: 16000
        }
      }));
    });
    
    it('should return a properly formatted LLMResponse', async () => {
      const response = await provider.generate('Test prompt', 'claude-3-opus-20240229');
      
      expect(response).toEqual({
        provider: 'anthropic',
        modelId: 'claude-3-opus-20240229',
        text: 'Test response',
        metadata: {
          usage: { input_tokens: 5, output_tokens: 5 },
          model: 'claude-3-opus-20240229',
          id: 'test-id',
          type: 'message',
        },
      });
    });
    
    it('should handle empty response gracefully', async () => {
      // Mock an API response with no content
      mockCreate.mockResolvedValue({
        content: [], // Empty content
        usage: { input_tokens: 0, output_tokens: 0 },
        model: 'claude-3-opus-20240229',
        id: 'test-id',
        type: 'message',
      });
      
      const response = await provider.generate('Test prompt', 'claude-3-opus-20240229');
      
      expect(response.text).toBe('');
    });
    
    it('should handle API errors correctly', async () => {
      // Mock an API error
      mockCreate.mockRejectedValue(new Error('API error message'));
      
      // Should still be catchable as AnthropicProviderError for backward compatibility
      await expect(provider.generate('Test prompt', 'claude-3-opus-20240229')).rejects.toThrow(AnthropicProviderError);
      // But should also be an instance of ApiError from the new system
      await expect(provider.generate('Test prompt', 'claude-3-opus-20240229')).rejects.toThrow(ApiError);
      await expect(provider.generate('Test prompt', 'claude-3-opus-20240229')).rejects.toThrow('[anthropic] API error message');
      
      try {
        await provider.generate('Test prompt', 'claude-3-opus-20240229');
      } catch (error) {
        // Verify it has the expected properties from ApiError
        const typedError = error as AnthropicProviderError;
        expect(typedError.category).toBe('API');
        expect(typedError.providerId).toBe('anthropic');
        expect(typedError.cause).toBeInstanceOf(Error);
        expect(typedError.cause?.message).toBe('API error message');
      }
    });
    
    it('should reuse the Anthropic client for multiple requests', async () => {
      await provider.generate('Test prompt 1', 'claude-3-opus-20240229');
      await provider.generate('Test prompt 2', 'claude-3-opus-20240229');
      
      // Should only create the client once
      expect(MockedAnthropic).toHaveBeenCalledTimes(1);
      
      // But should make two API calls
      expect(mockCreate).toHaveBeenCalledTimes(2);
    });
  });

  describe('listModels', () => {
    let provider: AnthropicProvider;
    
    beforeEach(() => {
      provider = new AnthropicProvider('test-api-key');
    });
    
    it('should fetch models from the hardcoded list', async () => {
      const models: LLMAvailableModel[] = await provider.listModels();
      
      // Verify models are correctly mapped from hardcoded list
      expect(models).toHaveLength(5); // There are 5 models in the hardcoded list
      expect(models[0]).toEqual(expect.objectContaining({
        id: 'claude-3-opus-20240229',
        name: 'Claude 3 Opus',
        description: expect.stringContaining('Anthropic\'s most powerful model')
      }));
    });
    
    // We no longer need to test handling missing display names since we use a hardcoded list
    it('should include capabilities and pricing information', async () => {
      const models: LLMAvailableModel[] = await provider.listModels();
      
      // Check that the models have all required information
      expect(models.length).toBeGreaterThan(0);
      models.forEach(model => {
        expect(model.id).toBeDefined();
        expect(model.description).toBeDefined();
        // Extended fields are still available in the implementation but not in the interface
        const extendedModel = model as any;
        expect(extendedModel.name).toBeDefined();
        expect(extendedModel.provider).toBe('anthropic');
        expect(extendedModel.capabilities).toBeDefined();
        expect(extendedModel.pricing).toBeDefined();
        expect(extendedModel.contextWindow).toBeGreaterThan(0);
      });
    });
    
    it('should handle API errors gracefully', async () => {
      // Create a new provider instance for this test
      const errorProvider = new AnthropicProvider();
      
      // Mock the entire listModels method
      const originalListModels = errorProvider.listModels;
      errorProvider.listModels = jest.fn().mockRejectedValue(
        new ApiError('Invalid API key', {
          providerId: 'anthropic',
          cause: new Error('Invalid API key'),
          suggestions: ['Check your API key']
        })
      );
      
      // Test error handling
      try {
        await errorProvider.listModels('invalid-key');
        fail('Should have thrown an error');
      } catch (error) {
        // Verify it has the expected properties from ApiError
        const typedError = error as AnthropicProviderError;
        expect(typedError.category).toBe('API');
        expect(typedError.providerId).toBe('anthropic');
        expect(typedError.message).toContain('API key');
        expect(typedError.suggestions).toBeDefined();
      }
      
      // Restore the original method
      errorProvider.listModels = originalListModels;
    });
    
    it('should handle rate limiting errors with specific suggestions', async () => {
      // Create a new provider instance for this test
      const rateProvider = new AnthropicProvider();
      
      // Mock the entire listModels method
      const originalListModels = rateProvider.listModels;
      rateProvider.listModels = jest.fn().mockRejectedValue(
        createProviderRateLimitError('anthropic', 'Anthropic', new Error('Rate limit exceeded'))
      );
      
      try {
        await rateProvider.listModels('valid-key');
        fail('Should have thrown an error');
      } catch (error) {
        // Verify error contains specific rate limit suggestions
        const typedError = error as AnthropicProviderError;
        expect(typedError.suggestions).toBeDefined();
        // Check that at least one suggestion mentions rate limiting or waiting
        expect(typedError.suggestions?.some(suggestion => 
          suggestion.toLowerCase().includes('rate') || 
          suggestion.toLowerCase().includes('wait') ||
          suggestion.toLowerCase().includes('delay')
        )).toBe(true);
      }
      
      // Restore the original method
      rateProvider.listModels = originalListModels;
    });
  });
});