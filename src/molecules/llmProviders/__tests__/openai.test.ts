/**
 * Unit tests for OpenAI provider
 * 
 * Note: We import the provider first to trigger its auto-registration
 */
import { OpenAIProvider, OpenAIProviderError, openaiProvider } from '../openai';
import { ModelOptions } from '../../../atoms/types';
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
});