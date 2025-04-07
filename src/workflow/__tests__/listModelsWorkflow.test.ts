/**
 * Unit tests for the list models workflow
 */
import { listAvailableModels } from '../listModelsWorkflow';
import * as configManager from '../../core/configManager';
import * as llmRegistry from '../../core/llmRegistry';
import * as outputFormatter from '../../utils/outputFormatter';
import { AppConfig, LLMAvailableModel, LLMProvider } from '../../core/types';

// Mock dependencies
jest.mock('../../core/configManager');
jest.mock('../../core/llmRegistry');
jest.mock('../../utils/outputFormatter');

describe('List Models Workflow', () => {
  // Setup common test data
  const mockConfig: AppConfig = {
    models: [
      { provider: 'openai', modelId: 'gpt-4o', enabled: true },
      { provider: 'openai', modelId: 'gpt-3.5-turbo', enabled: false },
      { provider: 'anthropic', modelId: 'claude-3-opus-20240229', enabled: true },
      { provider: 'google', modelId: 'gemini-pro', enabled: true },
      { provider: 'nonexistent', modelId: 'model', enabled: true },
    ]
  };
  
  // Sample available models for each provider
  const openaiModels: LLMAvailableModel[] = [
    { id: 'gpt-4o', description: 'Latest GPT-4' },
    { id: 'gpt-3.5-turbo', description: 'GPT-3.5 Turbo' },
    { id: 'gpt-4-turbo', description: 'GPT-4 Turbo' }
  ];
  
  const anthropicModels: LLMAvailableModel[] = [
    { id: 'claude-3-opus-20240229', description: 'Claude 3 Opus' },
    { id: 'claude-3-sonnet-20240229', description: 'Claude 3 Sonnet' },
    { id: 'claude-3-haiku-20240307', description: 'Claude 3 Haiku' }
  ];
  
  const googleModels: LLMAvailableModel[] = [
    { id: 'gemini-pro', description: 'Gemini Pro' },
    { id: 'gemini-ultra', description: 'Gemini Ultra' }
  ];
  
  // Define mock providers with promise-returning functions
  const mockOpenAIProvider: LLMProvider = {
    providerId: 'openai',
    generate: jest.fn(),
    listModels: jest.fn().mockImplementation(() => Promise.resolve(openaiModels))
  };
  
  const mockAnthropicProvider: LLMProvider = {
    providerId: 'anthropic',
    generate: jest.fn(),
    listModels: jest.fn().mockImplementation(() => Promise.resolve(anthropicModels))
  };
  
  const mockGoogleProvider: LLMProvider = {
    providerId: 'google',
    generate: jest.fn(),
    listModels: jest.fn().mockImplementation(() => Promise.resolve(googleModels))
  };
  
  // Provider without listModels method
  const mockLegacyProvider: LLMProvider = {
    providerId: 'legacy',
    generate: jest.fn()
    // No listModels method
  };
  
  const mockProviders: Record<string, LLMProvider> = {
    'openai': mockOpenAIProvider,
    'anthropic': mockAnthropicProvider,
    'google': mockGoogleProvider,
    'legacy': mockLegacyProvider
  };
  
  beforeEach(() => {
    // Reset mocks
    jest.resetAllMocks();
    
    // Set up mock implementations
    (configManager.loadConfig as jest.Mock).mockResolvedValue(mockConfig);
    (llmRegistry.getProvider as jest.Mock).mockImplementation((providerId: string) => mockProviders[providerId]);
    (configManager.getApiKey as jest.Mock).mockImplementation((config) => {
      if (config.provider === 'openai') return 'openai-api-key';
      if (config.provider === 'anthropic') return 'anthropic-api-key';
      if (config.provider === 'google') return 'google-api-key';
      if (config.provider === 'legacy') return 'legacy-api-key';
      return undefined;
    });
    (outputFormatter.formatModelList as jest.Mock).mockImplementation(() => 'Formatted model list');
  });
  
  it.skip('should fetch models from all configured providers by default', async () => {
    const result = await listAvailableModels({});
    
    // Should load config
    expect(configManager.loadConfig).toHaveBeenCalledWith({ configPath: undefined });
    
    // Should get the providers from the registry
    expect(llmRegistry.getProvider).toHaveBeenCalledWith('openai');
    expect(llmRegistry.getProvider).toHaveBeenCalledWith('anthropic');
    expect(llmRegistry.getProvider).toHaveBeenCalledWith('google');
    expect(llmRegistry.getProvider).toHaveBeenCalledWith('nonexistent');
    
    // Should call listModels on providers that have it
    expect(mockOpenAIProvider.listModels).toHaveBeenCalledWith('openai-api-key');
    expect(mockAnthropicProvider.listModels).toHaveBeenCalledWith('anthropic-api-key');
    expect(mockGoogleProvider.listModels).toHaveBeenCalledWith('google-api-key');
    
    // Should format the results
    expect(outputFormatter.formatModelList).toHaveBeenCalled();
    expect(result).toBe('Formatted model list');
    
    // Verify that the format call includes all provider results
    const formatCallArg = (outputFormatter.formatModelList as jest.Mock).mock.calls[0][0];
    expect(formatCallArg).toHaveProperty('openai');
    expect(formatCallArg).toHaveProperty('anthropic');
    expect(formatCallArg).toHaveProperty('google');
    expect(formatCallArg).toHaveProperty('nonexistent');
    
    // Verify that actual model data was passed to formatter
    expect(formatCallArg.openai).toEqual(openaiModels);
    expect(formatCallArg.anthropic).toEqual(anthropicModels);
    expect(formatCallArg.google).toEqual(googleModels);
    expect(formatCallArg.nonexistent).toHaveProperty('error');
  });
  
  it('should filter providers by the specified provider ID', async () => {
    await listAvailableModels({ provider: 'anthropic' });
    
    // Should only get anthropic provider
    expect(llmRegistry.getProvider).toHaveBeenCalledWith('anthropic');
    expect(llmRegistry.getProvider).not.toHaveBeenCalledWith('openai');
    expect(llmRegistry.getProvider).not.toHaveBeenCalledWith('google');
    
    // Should only call listModels on anthropic
    expect(mockAnthropicProvider.listModels).toHaveBeenCalledWith('anthropic-api-key');
    expect(mockOpenAIProvider.listModels).not.toHaveBeenCalled();
    expect(mockGoogleProvider.listModels).not.toHaveBeenCalled();
    
    // Verify that format call only includes anthropic results
    const formatCallArg = (outputFormatter.formatModelList as jest.Mock).mock.calls[0][0];
    expect(formatCallArg).toHaveProperty('anthropic');
    expect(formatCallArg).not.toHaveProperty('openai');
    expect(formatCallArg).not.toHaveProperty('google');
  });
  
  it('should handle a provider filter that matches no models', async () => {
    await listAvailableModels({ provider: 'nonexistent-provider' });
    
    // Should not call any provider's listModels
    expect(mockOpenAIProvider.listModels).not.toHaveBeenCalled();
    expect(mockAnthropicProvider.listModels).not.toHaveBeenCalled();
    expect(mockGoogleProvider.listModels).not.toHaveBeenCalled();
    
    // Should still format (empty) results
    expect(outputFormatter.formatModelList).toHaveBeenCalled();
    const formatCallArg = (outputFormatter.formatModelList as jest.Mock).mock.calls[0][0];
    expect(Object.keys(formatCallArg).length).toBe(0);
  });
  
  it('should use a custom config file when specified', async () => {
    await listAvailableModels({ config: 'custom-config.json' });
    
    // Should load the custom config
    expect(configManager.loadConfig).toHaveBeenCalledWith({ configPath: 'custom-config.json' });
  });
  
  it('should handle providers without listModels method', async () => {
    // Modify mock config to include legacy provider
    const configWithLegacy = {
      ...mockConfig,
      models: [
        ...mockConfig.models,
        { provider: 'legacy', modelId: 'legacy-model', enabled: true }
      ]
    };
    (configManager.loadConfig as jest.Mock).mockResolvedValue(configWithLegacy);
    
    await listAvailableModels({});
    
    // Should get all providers
    expect(llmRegistry.getProvider).toHaveBeenCalledWith('legacy');
    
    // Should call formatModelList with appropriate errors for providers without listModels
    const formatCallArg = (outputFormatter.formatModelList as jest.Mock).mock.calls[0][0];
    expect(formatCallArg).toHaveProperty('legacy');
    expect(formatCallArg.legacy).toHaveProperty('error');
    expect(formatCallArg.legacy.error).toBe("Provider 'legacy' does not support listing models");
  });
  
  it('should handle missing API keys gracefully', async () => {
    // Mock getApiKey to return undefined for anthropic and google
    (configManager.getApiKey as jest.Mock).mockImplementation((config) => {
      if (config.provider === 'openai') return 'openai-api-key';
      return undefined;
    });
    
    await listAvailableModels({});
    
    // Should not try to call listModels on providers with missing API keys
    expect(mockAnthropicProvider.listModels).not.toHaveBeenCalled();
    expect(mockGoogleProvider.listModels).not.toHaveBeenCalled();
    
    // Should include error in formatted results
    const formatCallArg = (outputFormatter.formatModelList as jest.Mock).mock.calls[0][0];
    expect(formatCallArg).toHaveProperty('anthropic');
    expect(formatCallArg.anthropic).toHaveProperty('error');
    expect(formatCallArg.anthropic.error).toContain('Missing API key');
    
    expect(formatCallArg).toHaveProperty('google');
    expect(formatCallArg.google).toHaveProperty('error');
    expect(formatCallArg.google.error).toContain('Missing API key');
  });
  
  it('should handle API errors from providers', async () => {
    // Reset the mocks to ensure clean state
    jest.clearAllMocks();
    
    // Set up mock implementations
    (configManager.loadConfig as jest.Mock).mockResolvedValue(mockConfig);
    (llmRegistry.getProvider as jest.Mock).mockImplementation((providerId: string) => mockProviders[providerId]);
    (configManager.getApiKey as jest.Mock).mockImplementation((config) => {
      if (config.provider === 'openai') return 'openai-api-key';
      if (config.provider === 'anthropic') return 'anthropic-api-key';
      if (config.provider === 'google') return 'google-api-key';
      if (config.provider === 'legacy') return 'legacy-api-key';
      return undefined;
    });
    
    // Set up specific behavior for this test
    // Make anthropic.listModels throw an error
    mockAnthropicProvider.listModels = jest.fn().mockImplementation(() => Promise.reject(new Error('API error')));
    
    // Make google.listModels return a different error
    mockGoogleProvider.listModels = jest.fn().mockImplementation(() => Promise.reject(new Error('Rate limit exceeded')));
    
    // Make openai.listModels succeed
    mockOpenAIProvider.listModels = jest.fn().mockImplementation(() => Promise.resolve(openaiModels));
    
    await listAvailableModels({});
    
    // Should include error in formatted results
    const formatCallArg = (outputFormatter.formatModelList as jest.Mock).mock.calls[0][0];
    
    expect(formatCallArg).toHaveProperty('anthropic');
    expect(formatCallArg.anthropic).toHaveProperty('error');
    expect(formatCallArg.anthropic.error).toContain('API error');
    
    expect(formatCallArg).toHaveProperty('google');
    expect(formatCallArg.google).toHaveProperty('error');
    expect(formatCallArg.google.error).toContain('Rate limit exceeded');
    
    // OpenAI should still succeed
    expect(formatCallArg.openai).toEqual(openaiModels);
  });
  
  it('should handle unexpected errors during provider processing', async () => {
    // Create a scenario where getProvider throws an error for one provider
    (llmRegistry.getProvider as jest.Mock).mockImplementation((providerId: string) => {
      if (providerId === 'anthropic') throw new Error('Registry error');
      return mockProviders[providerId];
    });
    
    await listAvailableModels({});
    
    // Should still process other providers
    expect(mockOpenAIProvider.listModels).toHaveBeenCalled();
    expect(mockGoogleProvider.listModels).toHaveBeenCalled();
    
    // Should include error in formatted results
    const formatCallArg = (outputFormatter.formatModelList as jest.Mock).mock.calls[0][0];
    expect(formatCallArg).toHaveProperty('anthropic');
    expect(formatCallArg.anthropic).toHaveProperty('error');
    expect(formatCallArg.anthropic.error).toContain('Unexpected error');
  });
  
  it('should handle a completely empty config with no models', async () => {
    // Create empty config
    const emptyConfig: AppConfig = {
      models: []
    };
    (configManager.loadConfig as jest.Mock).mockResolvedValue(emptyConfig);
    
    await listAvailableModels({});
    
    // Should not try to get any providers
    expect(llmRegistry.getProvider).not.toHaveBeenCalled();
    
    // Should still format (empty) results
    expect(outputFormatter.formatModelList).toHaveBeenCalled();
    const formatCallArg = (outputFormatter.formatModelList as jest.Mock).mock.calls[0][0];
    expect(Object.keys(formatCallArg).length).toBe(0);
  });
});
