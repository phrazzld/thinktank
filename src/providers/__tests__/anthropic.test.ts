/**
 * Unit tests for Anthropic provider
 * 
 * Note: We import the provider first to trigger its auto-registration
 */
import { AnthropicProvider, AnthropicProviderError, anthropicProvider } from '../anthropic';
import { ModelOptions, LLMAvailableModel } from '../../core/types';
import { clearRegistry, getProvider } from '../../core/llmRegistry';
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
      
      await expect(provider.generate('Test', 'claude-3-opus-20240229')).rejects.toThrow(AnthropicProviderError);
      await expect(provider.generate('Test', 'claude-3-opus-20240229')).rejects.toThrow('Anthropic API key is missing');
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
      
      expect(mockCreate).toHaveBeenCalledWith({
        model: 'claude-3-opus-20240229',
        messages: [{ role: 'user' as const, content: 'Test prompt' }],
        max_tokens: 1000, // Default value now from cascading config
      });
    });
    
    it('should map ModelOptions to Anthropic parameters', async () => {
      const options: ModelOptions = {
        temperature: 0.7,
        maxTokens: 500,
        topP: 0.9, // Additional option
      };
      
      await provider.generate('Test prompt', 'claude-3-opus-20240229', options);
      
      expect(mockCreate).toHaveBeenCalledWith({
        model: 'claude-3-opus-20240229',
        messages: [{ role: 'user' as const, content: 'Test prompt' }],
        temperature: 0.7,
        max_tokens: 500,
        topP: 0.9,
      });
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
      
      await expect(provider.generate('Test prompt', 'claude-3-opus-20240229')).rejects.toThrow(AnthropicProviderError);
      await expect(provider.generate('Test prompt', 'claude-3-opus-20240229')).rejects.toThrow('Anthropic API error: API error message');
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
    let mockList: jest.Mock;
    
    beforeEach(() => {
      provider = new AnthropicProvider('test-api-key');
      
      // Set up mock for models.list
      mockList = jest.fn();
      
      // Mock the Anthropic client creation to include models.list
      MockedAnthropic.mockImplementation(() => {
        return {
          models: {
            list: mockList
          }
        } as unknown as Anthropic;
      });
    });
    
    it('should fetch models from the Anthropic API', async () => {
      // Mock models.list response
      mockList.mockResolvedValue({
        data: [
          { 
            type: "model",
            id: 'claude-3-opus-20240229',
            display_name: 'Claude 3 Opus',
            created_at: '2024-02-29T00:00:00Z'
          },
          { 
            type: "model",
            id: 'claude-3-sonnet-20240229',
            display_name: 'Claude 3 Sonnet',
            created_at: '2024-02-29T00:00:00Z'
          },
          { 
            type: "model",
            id: 'claude-3-haiku-20240307',
            display_name: 'Claude 3 Haiku',
            created_at: '2024-03-07T00:00:00Z'
          }
        ],
        has_more: false,
        first_id: 'claude-3-opus-20240229',
        last_id: 'claude-3-haiku-20240307'
      });
      
      const models: LLMAvailableModel[] = await provider.listModels('test-api-key');
      
      // Verify SDK was initialized with correct API key
      expect(MockedAnthropic).toHaveBeenCalledWith({ apiKey: 'test-api-key' });
      
      // Verify models.list was called
      expect(mockList).toHaveBeenCalled();
      
      // Verify models are correctly mapped
      expect(models).toHaveLength(3);
      expect(models[0]).toEqual({
        id: 'claude-3-opus-20240229',
        description: 'Claude 3 Opus'
      });
    });
    
    it('should handle missing display names in model data', async () => {
      // Mock models.list response with a model missing display_name
      mockList.mockResolvedValue({
        data: [
          { 
            type: "model",
            id: 'claude-3-opus-20240229',
            // No display_name
            created_at: '2024-02-29T00:00:00Z'
          },
          { 
            type: "model",
            id: 'claude-3-sonnet-20240229',
            display_name: 'Claude 3 Sonnet',
            created_at: '2024-02-29T00:00:00Z'
          }
        ],
        has_more: false,
        first_id: 'claude-3-opus-20240229',
        last_id: 'claude-3-sonnet-20240229'
      });
      
      const models: LLMAvailableModel[] = await provider.listModels('test-api-key');
      
      expect(models).toHaveLength(2);
      expect(models[0].description).toBeUndefined();
      expect(models[1].description).toBe('Claude 3 Sonnet');
    });
    
    it('should handle API errors gracefully', async () => {
      // Mock API error
      const apiError = new Error('Invalid API key');
      mockList.mockRejectedValue(apiError);
      
      // The method should throw AnthropicProviderError
      await expect(provider.listModels('invalid-key')).rejects.toThrow(AnthropicProviderError);
      await expect(provider.listModels('invalid-key')).rejects.toThrow('Anthropic API error when listing models');
    });
  });
});