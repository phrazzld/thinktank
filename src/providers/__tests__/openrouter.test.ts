/**
 * Unit tests for OpenRouter provider
 * 
 * Note: We import the provider first to trigger its auto-registration
 */
import { OpenRouterProvider, OpenRouterProviderError, openrouterProvider } from '../openrouter';
import { ModelOptions } from '../../core/types';
import { clearRegistry, getProvider } from '../../core/llmRegistry';
import OpenAI from 'openai';
import axios from 'axios';

// Mock OpenAI library
jest.mock('openai');
const MockedOpenAI = jest.mocked(OpenAI);

// Mock axios
jest.mock('axios');
const mockedAxios = jest.mocked(axios);

describe('OpenRouter Provider', () => {
  const originalEnv = process.env;
  
  beforeEach(() => {
    // Reset environment
    process.env = { ...originalEnv };
    
    // Reset mocks
    jest.clearAllMocks();
    
    // Clear registry
    clearRegistry();
    
    // Reset mock implementation
    MockedOpenAI.mockClear();
    mockedAxios.get.mockClear();
    
    // Re-register the provider because we cleared the registry
    try {
      openrouterProvider.providerId; // Access a property to ensure the module is initialized
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
      new OpenRouterProvider(); // This should trigger registration
      
      const provider = getProvider('openrouter');
      expect(provider).toBeDefined();
      expect(provider?.providerId).toBe('openrouter');
    });
    
    it('should use API key from constructor', async () => {
      const provider = new OpenRouterProvider('test-api-key');
      
      // Prepare mock for client creation
      const mockCreate = jest.fn().mockResolvedValue({
        choices: [{ message: { content: 'Test response' } }],
        usage: { total_tokens: 10 },
        model: 'openai/gpt-4o',
        id: 'test-id',
        created: 123456789,
      });
      
      MockedOpenAI.mockImplementation(() => {
        return {
          chat: {
            completions: {
              create: mockCreate,
            },
          },
        } as unknown as OpenAI;
      });
      
      // Generate should cause client creation
      try {
        await provider.generate('Test prompt', 'openai/gpt-4o');
      } catch (error) {
        // Ignore errors for this test
      }
      
      // Verify API key and OpenRouter-specific config was used
      expect(MockedOpenAI).toHaveBeenCalledWith(expect.objectContaining({
        apiKey: 'test-api-key',
        baseURL: 'https://openrouter.ai/api/v1',
      }));
    });
    
    it('should use API key from environment', async () => {
      process.env.OPENROUTER_API_KEY = 'env-api-key';
      const provider = new OpenRouterProvider();
      
      // Prepare mock for client creation
      const mockCreate = jest.fn().mockResolvedValue({
        choices: [{ message: { content: 'Test response' } }],
        usage: { total_tokens: 10 },
        model: 'openai/gpt-4o',
        id: 'test-id',
        created: 123456789,
      });
      
      MockedOpenAI.mockImplementation(() => {
        return {
          chat: {
            completions: {
              create: mockCreate,
            },
          },
        } as unknown as OpenAI;
      });
      
      // Generate should cause client creation
      try {
        await provider.generate('Test prompt', 'openai/gpt-4o');
      } catch (error) {
        // Ignore errors for this test
      }
      
      // Verify environment API key was used
      expect(MockedOpenAI).toHaveBeenCalledWith(expect.objectContaining({
        apiKey: 'env-api-key',
        baseURL: 'https://openrouter.ai/api/v1',
      }));
    });
    
    it('should throw error if API key is missing', async () => {
      // Ensure OPENROUTER_API_KEY is not set
      delete process.env.OPENROUTER_API_KEY;
      
      const provider = new OpenRouterProvider();
      
      await expect(provider.generate('Test', 'openai/gpt-4o')).rejects.toThrow(OpenRouterProviderError);
      await expect(provider.generate('Test', 'openai/gpt-4o')).rejects.toThrow('OpenRouter API key is missing');
    });
  });
  
  describe('generate', () => {
    // Create a valid provider with API key
    let provider: OpenRouterProvider;
    let mockCreate: jest.Mock;
    
    beforeEach(() => {
      provider = new OpenRouterProvider('test-api-key');
      
      // Prepare mock for API response
      mockCreate = jest.fn().mockResolvedValue({
        choices: [{ message: { content: 'Test response' } }],
        usage: { total_tokens: 10 },
        model: 'openai/gpt-4o',
        id: 'test-id',
        created: 123456789,
      });
      
      MockedOpenAI.mockImplementation(() => {
        return {
          chat: {
            completions: {
              create: mockCreate,
            },
          },
        } as unknown as OpenAI;
      });
    });
    
    it('should call OpenRouter API with the correct parameters', async () => {
      await provider.generate('Test prompt', 'openai/gpt-4o');
      
      expect(mockCreate).toHaveBeenCalledWith({
        model: 'openai/gpt-4o',
        messages: [{ role: 'user', content: 'Test prompt' }],
      });
    });
    
    it('should map ModelOptions to OpenRouter parameters', async () => {
      const options: ModelOptions = {
        temperature: 0.7,
        maxTokens: 500,
        topP: 0.9, // Additional option
      };
      
      await provider.generate('Test prompt', 'openai/gpt-4o', options);
      
      expect(mockCreate).toHaveBeenCalledWith({
        model: 'openai/gpt-4o',
        messages: [{ role: 'user', content: 'Test prompt' }],
        temperature: 0.7,
        max_tokens: 500,
        topP: 0.9,
      });
    });
    
    it('should return a properly formatted LLMResponse', async () => {
      const response = await provider.generate('Test prompt', 'openai/gpt-4o');
      
      expect(response).toEqual({
        provider: 'openrouter',
        modelId: 'openai/gpt-4o',
        text: 'Test response',
        metadata: expect.objectContaining({
          usage: { total_tokens: 10 },
          model: 'openai/gpt-4o',
          id: 'test-id',
          created: 123456789,
        }),
      });
    });
    
    it('should handle empty response gracefully', async () => {
      // Mock an API response with no content
      mockCreate.mockResolvedValue({
        choices: [{ message: { content: '' } }],
        usage: { total_tokens: 0 },
        model: 'openai/gpt-4o',
        id: 'test-id',
        created: 123456789,
      });
      
      const response = await provider.generate('Test prompt', 'openai/gpt-4o');
      
      expect(response.text).toBe('');
    });
    
    it('should handle missing choice gracefully', async () => {
      // Mock an API response with no choices
      mockCreate.mockResolvedValue({
        choices: [],
        usage: { total_tokens: 0 },
        model: 'openai/gpt-4o',
        id: 'test-id',
        created: 123456789,
      });
      
      const response = await provider.generate('Test prompt', 'openai/gpt-4o');
      
      expect(response.text).toBe('');
    });
    
    it('should handle API errors correctly', async () => {
      // Mock an API error
      mockCreate.mockRejectedValue(new Error('API error message'));
      
      await expect(provider.generate('Test prompt', 'openai/gpt-4o')).rejects.toThrow(OpenRouterProviderError);
      await expect(provider.generate('Test prompt', 'openai/gpt-4o')).rejects.toThrow('OpenRouter API error: API error message');
    });
    
    it('should reuse the OpenAI client for multiple requests', async () => {
      await provider.generate('Test prompt 1', 'openai/gpt-4o');
      await provider.generate('Test prompt 2', 'openai/gpt-4o');
      
      // Should only create the client once
      expect(MockedOpenAI).toHaveBeenCalledTimes(1);
      
      // But should make two API calls
      expect(mockCreate).toHaveBeenCalledTimes(2);
    });
  });
  
  describe('listModels', () => {
    // Create a valid provider with API key
    let provider: OpenRouterProvider;
    
    beforeEach(() => {
      provider = new OpenRouterProvider('test-api-key');
      
      // Mock axios response
      mockedAxios.get.mockResolvedValue({
        data: {
          data: [
            {
              id: 'openai/gpt-4o',
              name: 'GPT-4o',
              context_length: 128000,
              pricing: {
                prompt: 5,
                completion: 15
              }
            },
            {
              id: 'anthropic/claude-3-opus-20240229',
              name: 'Claude 3 Opus',
              context_length: 200000,
              pricing: {
                prompt: 15,
                completion: 75
              }
            }
          ]
        }
      });
    });
    
    it('should call OpenRouter API with the correct parameters', async () => {
      await provider.listModels('test-api-key');
      
      // Check URL and headers
      expect(mockedAxios.get).toHaveBeenCalledWith(
        'https://openrouter.ai/api/v1/models',
        expect.objectContaining({
          headers: expect.objectContaining({
            'Authorization': 'Bearer test-api-key',
            'HTTP-Referer': 'https://github.com/phrazzld/thinktank',
            'X-Title': 'thinktank CLI',
          })
        })
      );
    });
    
    it('should return properly formatted LLMAvailableModel array', async () => {
      const models = await provider.listModels('test-api-key');
      
      // Check models array
      expect(models).toHaveLength(2);
      
      // Check first model
      expect(models[0]).toEqual({
        id: 'openai/gpt-4o',
        description: expect.stringContaining('GPT-4o'),
      });
      
      // Check second model
      expect(models[1]).toEqual({
        id: 'anthropic/claude-3-opus-20240229',
        description: expect.stringContaining('Claude 3 Opus'),
      });
    });
    
    it('should handle API errors correctly', async () => {
      // Mock API error response
      const error = new Error('API error message');
      (error as any).isAxiosError = true;
      (error as any).response = {
        status: 401,
        data: {
          error: {
            message: 'Invalid API key'
          }
        }
      };
      
      mockedAxios.get.mockRejectedValue(error);
      
      await expect(provider.listModels('invalid-key')).rejects.toThrow(OpenRouterProviderError);
      await expect(provider.listModels('invalid-key')).rejects.toThrow('Error listing OpenRouter models');
    });
    
    it('should throw error on invalid response format', async () => {
      // Mock an invalid API response
      mockedAxios.get.mockResolvedValue({
        data: {
          // Missing 'data' property
          models: []
        }
      });
      
      await expect(provider.listModels('test-api-key')).rejects.toThrow(OpenRouterProviderError);
      await expect(provider.listModels('test-api-key')).rejects.toThrow('Invalid response format');
    });
  });
});