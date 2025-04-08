/**
 * Tests for the ConcreteLLMClient implementation
 */
import { ConcreteLLMClient } from '../LLMClient';
import { ConfigManagerInterface } from '../interfaces';
import { LLMResponse, ModelOptions, SystemPrompt, AppConfig, ModelConfig } from '../types';
import * as llmRegistry from '../llmRegistry';
import * as configManager from '../configManager';
import { ApiError, ConfigError } from '../errors';

// Mock the llmRegistry module
jest.mock('../llmRegistry', () => ({
  callProvider: jest.fn(),
}));

// Mock findModel and findModelGroup from configManager
jest.mock('../configManager', () => ({
  findModel: jest.fn(),
  findModelGroup: jest.fn(),
  resolveModelOptions: jest.fn(),
}));

describe('ConcreteLLMClient', () => {
  // Mock ConfigManagerInterface
  const mockConfigManager: ConfigManagerInterface = {
    loadConfig: jest.fn(),
    saveConfig: jest.fn(),
    getActiveConfigPath: jest.fn(),
    getDefaultConfigPath: jest.fn(),
  };

  // Mock AppConfig
  const mockAppConfig: AppConfig = {
    models: [
      {
        provider: 'openai',
        modelId: 'gpt-4o',
        enabled: true,
        options: { temperature: 0.7 }
      },
      {
        provider: 'anthropic',
        modelId: 'claude-3-7-sonnet-20250219',
        enabled: true,
        options: { temperature: 0.5 }
      }
    ]
  };

  // Mock response from callProvider
  const mockLLMResponse: LLMResponse = {
    provider: 'openai',
    modelId: 'gpt-4o',
    text: 'Mock response text',
    metadata: { usage: { total_tokens: 150 } }
  };

  beforeEach(() => {
    // Reset all mocks
    jest.clearAllMocks();
    
    // Set up default mock behavior
    (mockConfigManager.loadConfig as jest.Mock).mockResolvedValue(mockAppConfig);
    (configManager.findModel as jest.Mock).mockImplementation((config: AppConfig, provider: string, modelId: string) => {
      return config.models.find((m: ModelConfig) => m.provider === provider && m.modelId === modelId);
    });
    (configManager.findModelGroup as jest.Mock).mockReturnValue(undefined);
    (llmRegistry.callProvider as jest.Mock).mockResolvedValue(mockLLMResponse);
  });

  describe('constructor', () => {
    it('should create an instance with a valid ConfigManagerInterface', () => {
      const client = new ConcreteLLMClient(mockConfigManager);
      expect(client).toBeInstanceOf(ConcreteLLMClient);
    });

    it('should throw an error if ConfigManagerInterface is not provided', () => {
      expect(() => new ConcreteLLMClient(undefined as unknown as ConfigManagerInterface))
        .toThrow('ConfigManagerInterface instance is required');
    });
  });

  describe('generate', () => {
    it('should parse provider:modelId format correctly and call callProvider', async () => {
      const client = new ConcreteLLMClient(mockConfigManager);
      const prompt = 'Test prompt';
      const providerModelId = 'openai:gpt-4o';
      const options: ModelOptions = { temperature: 0.8 };
      
      const result = await client.generate(prompt, providerModelId, options);
      
      // Verify results
      expect(result).toEqual(mockLLMResponse);
      
      // Verify config loading
      expect(mockConfigManager.loadConfig).toHaveBeenCalledTimes(1);
      
      // Verify model lookup
      expect(configManager.findModel).toHaveBeenCalledWith(
        mockAppConfig,
        'openai',
        'gpt-4o'
      );
      
      // Verify callProvider was called with correct arguments
      expect(llmRegistry.callProvider).toHaveBeenCalledWith(
        'openai',
        'gpt-4o',
        prompt,
        mockAppConfig.models[0],
        undefined,
        options,
        undefined
      );
    });

    it('should throw ConfigError for invalid provider:modelId format', async () => {
      const client = new ConcreteLLMClient(mockConfigManager);
      
      // Test various invalid formats
      await expect(client.generate('Test', 'invalid-format'))
        .rejects.toThrow(ConfigError);
        
      await expect(client.generate('Test', 'provider:'))
        .rejects.toThrow(ConfigError);
        
      await expect(client.generate('Test', ':modelId'))
        .rejects.toThrow(ConfigError);
    });

    it('should throw error when model is not found in configuration', async () => {
      const client = new ConcreteLLMClient(mockConfigManager);
      
      // Mock findModel to return undefined (model not found)
      (configManager.findModel as jest.Mock).mockReturnValue(undefined);
      
      await expect(client.generate('Test', 'openai:nonexistent-model'))
        .rejects.toThrow(ConfigError);
    });

    it('should use system prompt from override, model config, or group', async () => {
      const client = new ConcreteLLMClient(mockConfigManager);
      const modelConfig: ModelConfig = {
        provider: 'openai',
        modelId: 'gpt-4o',
        enabled: true,
        systemPrompt: { text: 'Model-specific system prompt' },
        options: { temperature: 0.7 }
      };
      
      // Set up mock to return model with systemPrompt
      (configManager.findModel as jest.Mock).mockReturnValue(modelConfig);
      
      // Case 1: Use override system prompt
      const overrideSystemPrompt: SystemPrompt = { text: 'Override system prompt' };
      await client.generate('Test', 'openai:gpt-4o', undefined, overrideSystemPrompt);
      
      expect(llmRegistry.callProvider).toHaveBeenCalledWith(
        'openai',
        'gpt-4o',
        'Test',
        modelConfig,
        undefined,
        undefined,
        overrideSystemPrompt
      );
      
      // Reset mocks
      jest.clearAllMocks();
      (mockConfigManager.loadConfig as jest.Mock).mockResolvedValue(mockAppConfig);
      (configManager.findModel as jest.Mock).mockReturnValue(modelConfig);
      
      // Case 2: Use model-specific system prompt
      await client.generate('Test', 'openai:gpt-4o');
      
      expect(llmRegistry.callProvider).toHaveBeenCalledWith(
        'openai',
        'gpt-4o',
        'Test',
        modelConfig,
        undefined,
        undefined,
        modelConfig.systemPrompt
      );
      
      // Reset mocks
      jest.clearAllMocks();
      (mockConfigManager.loadConfig as jest.Mock).mockResolvedValue(mockAppConfig);
      (configManager.findModel as jest.Mock).mockReturnValue({ ...modelConfig, systemPrompt: undefined });
      
      // Case 3: Use group system prompt when available
      const groupInfo = {
        groupName: 'test-group',
        systemPrompt: { text: 'Group system prompt' }
      };
      (configManager.findModelGroup as jest.Mock).mockReturnValue(groupInfo);
      
      await client.generate('Test', 'openai:gpt-4o');
      
      expect(llmRegistry.callProvider).toHaveBeenCalledWith(
        'openai',
        'gpt-4o',
        'Test',
        { ...modelConfig, systemPrompt: undefined },
        undefined,
        undefined,
        groupInfo.systemPrompt
      );
    });

    it('should add group information to response when available', async () => {
      const client = new ConcreteLLMClient(mockConfigManager);
      
      // Set up group info
      const groupInfo = {
        groupName: 'test-group',
        systemPrompt: { text: 'Group system prompt' }
      };
      (configManager.findModelGroup as jest.Mock).mockReturnValue(groupInfo);
      
      // For this test, return a response without groupInfo
      const responseWithoutGroupInfo = { ...mockLLMResponse, groupInfo: undefined };
      (llmRegistry.callProvider as jest.Mock).mockResolvedValue(responseWithoutGroupInfo);
      
      const result = await client.generate('Test', 'openai:gpt-4o');
      
      // Verify group info was added to the response
      expect(result.groupInfo).toEqual({
        name: groupInfo.groupName,
        systemPrompt: groupInfo.systemPrompt
      });
    });

    it('should properly handle errors from callProvider', async () => {
      const client = new ConcreteLLMClient(mockConfigManager);
      
      // Mock callProvider to throw an ApiError
      const apiError = new ApiError('API error from provider');
      (llmRegistry.callProvider as jest.Mock).mockRejectedValue(apiError);
      
      // Verify the error is re-thrown without wrapping
      await expect(client.generate('Test', 'openai:gpt-4o'))
        .rejects.toThrow(apiError);
    });

    it('should wrap unknown errors in ApiError', async () => {
      const client = new ConcreteLLMClient(mockConfigManager);
      
      // Mock callProvider to throw a generic Error
      const genericError = new Error('Unknown error');
      (llmRegistry.callProvider as jest.Mock).mockRejectedValue(genericError);
      
      // Verify the error is wrapped in ApiError
      await expect(client.generate('Test', 'openai:gpt-4o'))
        .rejects.toThrow(ApiError);
        
      // Try to catch and examine the error
      try {
        await client.generate('Test', 'openai:gpt-4o');
      } catch (error) {
        expect(error).toBeInstanceOf(ApiError);
        expect((error as ApiError).message).toContain('LLM request failed');
        expect((error as ApiError).cause).toBe(genericError);
        expect((error as ApiError).providerId).toBe('openai');
        expect((error as ApiError).suggestions?.length).toBeGreaterThan(0);
      }
    });

    it('should handle ConfigManagerInterface errors', async () => {
      const client = new ConcreteLLMClient(mockConfigManager);
      
      // Mock loadConfig to throw an error
      const configError = new ConfigError('Config loading error');
      (mockConfigManager.loadConfig as jest.Mock).mockRejectedValue(configError);
      
      // Verify the error is re-thrown without wrapping
      await expect(client.generate('Test', 'openai:gpt-4o'))
        .rejects.toThrow(configError);
    });
  });
});
