/**
 * Unit tests for Anthropic provider
 * 
 * Note: We import the provider first to trigger its auto-registration
 */
import { AnthropicProvider, AnthropicProviderError, anthropicProvider } from '../anthropic';
import { ModelOptions } from '../../../atoms/types';
import { clearRegistry, getProvider } from '../../../organisms/llmRegistry';
import Anthropic from '@anthropic-ai/sdk';

// Mock Anthropic library
jest.mock('@anthropic-ai/sdk');
const MockedAnthropic = jest.mocked(Anthropic);

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
      expect(MockedAnthropic).toHaveBeenCalledWith({ apiKey: 'test-api-key' });
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
      expect(MockedAnthropic).toHaveBeenCalledWith({ apiKey: 'env-api-key' });
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
        max_tokens: 1024, // Default value
        temperature: undefined,
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
});