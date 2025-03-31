/**
 * Unit tests for OpenAI provider
 * 
 * Note: We import the provider first to trigger its auto-registration
 */
import { OpenAIProvider, OpenAIProviderError, openaiProvider } from '../openai';
import { ModelOptions, LLMAvailableModel } from '../../../atoms/types';
import { clearRegistry, getProvider } from '../../../organisms/llmRegistry';
import OpenAI from 'openai';

// Mock OpenAI library
jest.mock('openai');
const MockedOpenAI = jest.mocked(OpenAI);

describe('OpenAI Provider', () => {
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
    
    // Re-register the provider because we cleared the registry
    try {
      openaiProvider.providerId; // Access a property to ensure the module is initialized
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
      new OpenAIProvider(); // This should trigger registration
      
      const provider = getProvider('openai');
      expect(provider).toBeDefined();
      expect(provider?.providerId).toBe('openai');
    });
    
    it('should use API key from constructor', async () => {
      const provider = new OpenAIProvider('test-api-key');
      
      // Prepare mock for client creation
      const mockCreate = jest.fn().mockResolvedValue({
        choices: [{ message: { content: 'Test response' } }],
        usage: { total_tokens: 10 },
        model: 'gpt-4',
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
      await provider.generate('Test prompt', 'gpt-4');
      
      // Verify API key was used
      expect(MockedOpenAI).toHaveBeenCalledWith({ apiKey: 'test-api-key' });
    });
    
    it('should use API key from environment', async () => {
      process.env.OPENAI_API_KEY = 'env-api-key';
      const provider = new OpenAIProvider();
      
      // Prepare mock for client creation
      const mockCreate = jest.fn().mockResolvedValue({
        choices: [{ message: { content: 'Test response' } }],
        usage: { total_tokens: 10 },
        model: 'gpt-4',
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
      await provider.generate('Test prompt', 'gpt-4');
      
      // Verify environment API key was used
      expect(MockedOpenAI).toHaveBeenCalledWith({ apiKey: 'env-api-key' });
    });
    
    it('should throw error if API key is missing', async () => {
      // Ensure OPENAI_API_KEY is not set
      delete process.env.OPENAI_API_KEY;
      
      const provider = new OpenAIProvider();
      
      await expect(provider.generate('Test', 'gpt-4')).rejects.toThrow(OpenAIProviderError);
      await expect(provider.generate('Test', 'gpt-4')).rejects.toThrow('OpenAI API key is missing');
    });
  });
  
  describe('generate', () => {
    // Create a valid provider with API key
    let provider: OpenAIProvider;
    let mockCreate: jest.Mock;
    
    beforeEach(() => {
      provider = new OpenAIProvider('test-api-key');
      
      // Prepare mock for API response
      mockCreate = jest.fn().mockResolvedValue({
        choices: [{ message: { content: 'Test response' } }],
        usage: { total_tokens: 10 },
        model: 'gpt-4',
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
    
    it('should call OpenAI API with the correct parameters', async () => {
      await provider.generate('Test prompt', 'gpt-4');
      
      expect(mockCreate).toHaveBeenCalledWith({
        model: 'gpt-4',
        messages: [{ role: 'user', content: 'Test prompt' }],
      });
    });
    
    it('should map ModelOptions to OpenAI parameters', async () => {
      const options: ModelOptions = {
        temperature: 0.7,
        maxTokens: 500,
        topP: 0.9, // OpenAI-specific option
      };
      
      await provider.generate('Test prompt', 'gpt-4', options);
      
      expect(mockCreate).toHaveBeenCalledWith({
        model: 'gpt-4',
        messages: [{ role: 'user', content: 'Test prompt' }],
        temperature: 0.7,
        max_tokens: 500,
        topP: 0.9,
      });
    });
    
    it('should return a properly formatted LLMResponse', async () => {
      const response = await provider.generate('Test prompt', 'gpt-4');
      
      expect(response).toEqual({
        provider: 'openai',
        modelId: 'gpt-4',
        text: 'Test response',
        metadata: {
          usage: { total_tokens: 10 },
          model: 'gpt-4',
          id: 'test-id',
          created: 123456789,
        },
      });
    });
    
    it('should handle empty response gracefully', async () => {
      // Mock an API response with no content
      mockCreate.mockResolvedValue({
        choices: [{ message: { content: '' } }],
        usage: { total_tokens: 0 },
        model: 'gpt-4',
        id: 'test-id',
        created: 123456789,
      });
      
      const response = await provider.generate('Test prompt', 'gpt-4');
      
      expect(response.text).toBe('');
    });
    
    it('should handle missing choice gracefully', async () => {
      // Mock an API response with no choices
      mockCreate.mockResolvedValue({
        choices: [],
        usage: { total_tokens: 0 },
        model: 'gpt-4',
        id: 'test-id',
        created: 123456789,
      });
      
      const response = await provider.generate('Test prompt', 'gpt-4');
      
      expect(response.text).toBe('');
    });
    
    it('should handle API errors correctly', async () => {
      // Mock an API error
      mockCreate.mockRejectedValue(new Error('API error message'));
      
      await expect(provider.generate('Test prompt', 'gpt-4')).rejects.toThrow(OpenAIProviderError);
      await expect(provider.generate('Test prompt', 'gpt-4')).rejects.toThrow('OpenAI API error: API error message');
    });
    
    it('should reuse the OpenAI client for multiple requests', async () => {
      await provider.generate('Test prompt 1', 'gpt-4');
      await provider.generate('Test prompt 2', 'gpt-4');
      
      // Should only create the client once
      expect(MockedOpenAI).toHaveBeenCalledTimes(1);
      
      // But should make two API calls
      expect(mockCreate).toHaveBeenCalledTimes(2);
    });
  });

  describe('listModels', () => {
    let provider: OpenAIProvider;
    let mockList: jest.Mock;
    
    beforeEach(() => {
      provider = new OpenAIProvider('test-api-key');
      
      // Set up mock for models.list
      mockList = jest.fn();
      
      // Mock the OpenAI client creation to include models.list
      MockedOpenAI.mockImplementation(() => {
        return {
          models: {
            list: mockList
          }
        } as unknown as OpenAI;
      });
    });
    
    it('should fetch models from the OpenAI API', async () => {
      // Mock models.list response to return an AsyncIterable
      const mockModelsList = [
        { 
          id: 'gpt-4o',
          object: "model",
          created: 1686935002,
          owned_by: "openai"
        },
        { 
          id: 'gpt-4-turbo',
          object: "model",
          created: 1686935002,
          owned_by: "organization-owner",
        }
      ];
      
      // Set up the mock to behave like an async iterator
      mockList.mockReturnValue({
        [Symbol.asyncIterator]: async function* () {
          for (const model of mockModelsList) {
            yield model;
          }
        }
      });
      
      const models: LLMAvailableModel[] = await provider.listModels('test-api-key');
      
      // Verify SDK was initialized with correct API key
      expect(MockedOpenAI).toHaveBeenCalledWith({ apiKey: 'test-api-key' });
      
      // Verify models.list was called
      expect(mockList).toHaveBeenCalled();
      
      // Verify models are correctly mapped
      expect(models).toHaveLength(2);
      expect(models[0]).toEqual({
        id: 'gpt-4o',
        description: 'Owned by: openai'
      });
      expect(models[1]).toEqual({
        id: 'gpt-4-turbo',
        description: 'Owned by: organization-owner'
      });
    });
    
    it('should handle empty model list', async () => {
      // Mock empty list
      mockList.mockReturnValue({
        [Symbol.asyncIterator]: async function* () {
          // Empty array so the generator completes without yielding
          const emptyArray: any[] = [];
          for (const item of emptyArray) {
            yield item;
          }
        }
      });
      
      const models = await provider.listModels('test-api-key');
      
      expect(models).toEqual([]);
    });

    it('should handle API errors gracefully', async () => {
      // Mock API error by making the asyncIterator throw
      mockList.mockReturnValue({
        [Symbol.asyncIterator]: async function* () {
          const errorCondition = true;
          if (errorCondition) {
            throw new Error('Invalid API key');
          }
          yield null; // This line is unreachable but satisfies the linter
        }
      });
      
      // The method should throw OpenAIProviderError
      await expect(provider.listModels('invalid-key')).rejects.toThrow(OpenAIProviderError);
      await expect(provider.listModels('invalid-key')).rejects.toThrow('OpenAI API error when listing models');
    });
  });
});